# coding=utf-8

"""
 Collects JMX metrics from the Jolokia Agent. Jolokia is an HTTP bridge that
provides access to JMX MBeans without the need to write Java code. See the
[Reference Guide](http://www.jolokia.org/reference/html/index.html) for more
information.

By default, all MBeans will be queried for metrics. All numerical values will
be published to Graphite; anything else will be ignored. JolokiaCollector will
create a reasonable namespace for each metric based on each MBeans domain and
name. e.g) ```java.lang:name=ParNew,type=GarbageCollector``` would become
```java.lang.name_ParNew.type_GarbageCollector```.

#### Dependencies

 * Jolokia
 * A running JVM with Jolokia installed/configured

#### Example Configuration

If desired, JolokiaCollector can be configured to query specific MBeans by
providing a list of ```mbeans```. If ```mbeans``` is not provided, all MBeans
will be queried for metrics.  Note that the mbean prefix is checked both
with and without rewrites (including fixup re-writes) applied.  This allows
you to specify "java.lang:name=ParNew,type=GarbageCollector" (the raw name from
jolokia) or "java.lang.name_ParNew.type_GarbageCollector" (the fixed name
as used for output)

If the ```regex``` flag is set to True, mbeans will match based on regular
expressions rather than a plain textual match.

The ```rewrite``` section provides a way of renaming the data keys before
it sent out to the handler.  The section consists of pairs of from-to
regular expressions.  If the resultant name is completely blank, the
metric is not published, providing a way to exclude specific metrics within
an mbean.

```
    host = localhost
    port = 8778
    mbeans = "java.lang:name=ParNew,type=GarbageCollector",
     "org.apache.cassandra.metrics:name=WriteTimeouts,type=ClientRequestMetrics"
    [rewrite]
    java = coffee
    "-v\d+\.\d+\.\d+" = "-AllVersions"
    ".*GetS2Activities.*" = ""
```
"""

import diamond.collector
import json
import re
import time
import urllib
import urllib2
import sys

class MBean(object):
    def __init__(self, prefix, bean_key, bean_value):
        self.prefix = prefix
        self.bean_key = bean_key
        self.bean_value = bean_value

    def parse(self, patch_dimensions, patch_metric_name):
        metric_prefix, meta = self.prefix.split(':', 1)
        raw_dims = self.parse_dimension(meta)
        self.metric_name, self.metric_type, self.dimensions = patch_dimensions(self, raw_dims)
        raw_name_list = [metric_prefix]
        if self.metric_type:
            raw_name_list.append(self.metric_type)
        if self.metric_name:
            raw_name_list.append(self.metric_name)

        metric_name_list = patch_metric_name(self, raw_name_list)
        return metric_name_list, self.dimensions

    def parse_dimension(self, meta):
        dimensions = {}
        for k, v in [kv.split('=') for kv in meta.split(',')]:
            dimensions[str(k)] = v
        return dimensions


class JolokiaCollector(diamond.collector.Collector):
    LIST_URL = "/list?ifModifiedSince=%s&maxDepth=%s"
    READ_URL = "/?ignoreErrors=true&includeStackTrace=false&maxCollectionSize=%s&p=read/%s"
    LIST_QUERY_URL = "/list/%s?maxDepth=%s"

    """
    These domains contain MBeans that are for management purposes,
    or otherwise do not contain useful metrics
    """
    IGNORE_DOMAINS = ['JMImplementation', 'jmx4perl', 'jolokia',
                      'com.sun.management', 'java.util.logging']

    def get_default_config_help(self):
        config_help = super(JolokiaCollector,
                            self).get_default_config_help()
        config_help.update({
            'mbeans':  "Pipe delimited list of MBeans for which to collect"
                       " stats. If not provided, all stats will"
                       " be collected.",
            'regex': "Contols if mbeans option matches with regex,"
                       " False by default.",
            'host': 'Hostname',
            'port': 'Port',
            'mbean_blacklist': 'A list of blacklisted mbeans',
            'rewrite': "This sub-section of the config contains pairs of"
                       " from-to regex rewrites.",
            'url_path': 'Path to jolokia.  typically "jmx" or "jolokia"',
            'listing_max_depth': 'max depth of domain listings tree, 0=deepest, 1=keys only, 2=weird',
            'read_limit': 'Request size to read from jolokia, defaults to 1000, 0 = no limit'
        })
        return config_help

    def get_default_config(self):
        config = super(JolokiaCollector, self).get_default_config()
        config.update({
            'mbeans': [],
            'regex': False,
            'rewrite': [],
            'url_path': 'jolokia',
            'host': 'localhost',
            'mbean_blacklist': [],
            'port': 8778,
            'listing_max_depth': 1,
            'read_limit': 1000,
        })
        self.domain_keys = []
        self.last_list_request = 0
        return config

    def __init__(self, *args, **kwargs):
        super(JolokiaCollector, self).__init__(*args, **kwargs)
        self.mbeans = []
        self.rewrite = {}
        if isinstance(self.config['mbeans'], basestring):
            for mbean in self.config['mbeans'].split('|'):
                self.mbeans.append(mbean.strip())
        elif isinstance(self.config['mbeans'], list):
            self.mbeans = self.config['mbeans']
        if isinstance(self.config['rewrite'], dict):
            self.rewrite = self.config['rewrite']

    def check_mbean(self, mbean):
        if not self.mbeans:
            return True
        mbeanfix = self.clean_up(mbean)
        if self.config['regex'] is not None:
            for chkbean in self.mbeans:
                if re.match(chkbean, mbean) is not None or \
                   re.match(chkbean, mbeanfix) is not None:
                    return True
        else:
            if mbean in self.mbeans or mbeanfix in self.mbeans:
                return True

    def read_metric_path(self, full_path):
        obj = self.read_request(full_path, True)
        mbeans = obj['value'] if obj['status'] == 200 else {}
        self.collect_bean(full_path, mbeans)

    def read_except_blacklist(self, prefix, blacklist):
        listing = self.list_request(prefix)
        try:
            domains = listing['value'] if listing['status'] == 200 else {}
            domain_keys = domains.keys()
            for path in domain_keys:
                full_path = prefix + ":" + path
                if self.check_mbean_blacklist(full_path, blacklist):
                    self.read_metric_path(full_path)
        except KeyError:
            self.log.error("Unable to retrieve mbean listing")

    def check_mbean_blacklist(self, mbean, blacklist):
        for line in blacklist:
            if mbean.find(line) != -1:
                return False
        return True

    def check_domain_for_blacklist(self, domain, blacklist):
        for line in blacklist:
            if line.find(domain) != -1:
                return True
        return False

    def collect(self):
        listing = self.list_request()
        try:
            domains = listing['value'] if listing['status'] == 200 else {}
            if listing['status'] == 200:
                self.domain_keys = domains.keys()
                self.last_list_request = listing.get('timestamp', int(time.time()))
            for domain in self.domain_keys:
                if domain not in self.IGNORE_DOMAINS:
                    self.publish_metric_from_domain(domain)
        except KeyError:
            # The reponse was totally empty, or not an expected format
            self.log.error('Unable to retrieve MBean listing.')

    def publish_metric_from_domain(self, domain):
        if self.check_domain_for_blacklist(domain, self.config["mbean_blacklist"]):
            self.read_except_blacklist(domain, self.config["mbean_blacklist"])
            return
        obj = self.read_request(domain)
        mbeans = obj['value'] if obj['status'] == 200 else {}
        for k, v in mbeans.iteritems():
            if self.check_mbean(k):
                self.collect_bean(k, v)


    def read_json(self, request):
        json_str = request.read()
        return json.loads(json_str)

    def list_request(self, bean_path=None):
        try:
            if bean_path:
                url_path = self.LIST_QUERY_URL % (bean_path,
                                                  self.config['listing_max_depth'])
            else:
                url_path = self.LIST_URL % (self.last_list_request,
                                            self.config['listing_max_depth'])
            url = "http://%s:%s/%s%s" % (self.config['host'],
                                         self.config['port'],
                                         self.config['url_path'],
                                         url_path)
            response = urllib2.urlopen(url)
            return self.read_json(response)
        except (urllib2.HTTPError, ValueError):
            self.log.error('Unable to read JSON response.')
            return {}

    def read_request(self, url_path, read_bean=False):
        try:
            if read_bean:
                url_path = self.READ_URL % (self.config['read_limit'],
                                            self.escape_domain(url_path))
            else:
                url_path = self.READ_URL % (self.config['read_limit'],
                                            self.escape_domain(url_path)) + ":*"
            url = "http://%s:%s/%s%s" % (self.config['host'],
                                         self.config['port'],
                                         self.config['url_path'],
                                         url_path)
            response = urllib2.urlopen(url)
            return self.read_json(response)
        except (urllib2.HTTPError, ValueError):
            self.log.error('Unable to read JSON response.')
            return {}

    # escape the JMX domain per https://jolokia.org/reference/html/protocol.html
    # the Jolokia documentation suggests that, when using the p query parameter,
    # simply urlencoding should be sufficient, but in practice, the '!' appears
    # necessary (and not harmful)
    def escape_domain(self, domain):
        domain = re.sub('!', '!!', domain)
        domain = re.sub('/', '!/', domain)
        domain = re.sub('"', '!"', domain)
        domain = urllib.quote(domain)
        return domain

    def clean_up(self, text):
        text = re.sub('["\'(){}<>\[\]]', '', text)
        text = re.sub('[:,.]+', '.', text)
        text = re.sub('[^a-zA-Z0-9_.+-]+', '_', text)
        for (oldstr, newstr) in self.rewrite.items():
            text = re.sub(oldstr, newstr, text)
        return text

    def collect_bean(self, prefix, obj):
        for k, v in obj.iteritems():
            if type(v) in [int, float, long]:
                key = "%s.%s" % (prefix, k)
                key = self.clean_up(key)
                if key != "":
                    self.publish(key, v)
            elif type(v) in [dict]:
                self.collect_bean("%s.%s" % (prefix, k), v)
            elif type(v) in [list]:
                self.interpret_bean_with_list("%s.%s" % (prefix, k), v)

    def patch_dimensions(self, bean, dims):
        raise NotImplementedError()

    def patch_metric_name(self, bean, metric_name_list):
        raise NotImplementedError()

    def parse_dimension_bean(self, prefix, key, value):
        mbean = MBean(prefix, key, value)
        try:
            metric_name_list, self.dimensions = mbean.parse(self.patch_dimensions, self.patch_metric_name)
            metric_name = '.'.join(metric_name_list)
            metric_name = self.clean_up(metric_name)
            if metric_name == "":
                self.dimensions = {}
                return
            if key.lower() == 'count':
                self.publish_cumulative_counter(metric_name, value)
            else:
                self.publish(metric_name, value)
        except:
            exctype, value = sys.exc_info()[:2]
            self.log.error(str(value))


    # There's no unambiguous way to interpret list values, so
    # this hook lets subclasses handle them.
    def interpret_bean_with_list(self, prefix, values):
        pass
