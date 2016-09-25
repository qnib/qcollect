# coding=utf-8

"""
The Collector class is a base class for all metric collectors.
"""

import json
import logging
import os
import platform
import re
import socket
import subprocess
import sys
import time

from diamond.metric import Metric
from error import DiamondException

QCOLLECT_ADDR = ('', 19191)
QCOLLECT_RETRY_COUNT = 3

# Detect the architecture of the system and set the counters for MAX_VALUES
# appropriately. Otherwise, rolling over counters will cause incorrect or
# negative values.

if platform.architecture()[0] == '64bit':
    MAX_COUNTER = (2 ** 64) - 1
else:
    MAX_COUNTER = (2 ** 32) - 1


def str_to_bool(value):
    """
    Converts string truthy/falsey strings to a bool
    Empty strings are false
    """
    if isinstance(value, basestring):
        value = value.strip().lower()
        if value in ['true', 't', 'yes', 'y']:
            return True
        elif value in ['false', 'f', 'no', 'n', '']:
            return False
        else:
            raise NotImplementedError("Unknown bool %s" % value)

    return value


class CollectorErrorHandler(logging.Handler, object):
    def __init__(self, collector):
        super(CollectorErrorHandler, self).__init__()
        self.collector = collector

    def emit(self, error):
        e_type = sys.exc_info()[0]
        report_error(e_type, self.collector)


def report_error(e, collector):
    e_type = sys.exc_info()[0]
    metric_name = 'qcollect.collector_errors'
    metric_value = 1
    if e_type:
        collector.dimensions = {
            'error_type': str(e_type.__name__)
        }
    if collector.can_publish_metric():
        collector.publish(metric_name, metric_value)

class Collector(object):
    """
    The Collector class is a base class for all metric collectors.
    """

    def __init__(self, config=None, handlers=[], name=None, configfile=None):
        """
        Create a new instance of the Collector class
        """

        # Initialize Members
        if name is None:
            self.name = self.__class__.__name__
        else:
            self.name = name

        # Initialize Logger
        logger_name = '.'.join(['diamond', self.name])
        self.log = logging.getLogger(logger_name)
        error_handler = CollectorErrorHandler(self)
        error_handler.setLevel(logging.ERROR)
        self.log.addHandler(error_handler)

        self._socket = None
        self._reconnect = False
        self.default_dimensions = None
        self.dimensions = None
        self.handlers = handlers
        self.last_values = {}
        self.payload = []

        self.config = {}
        self.configfile = configfile or {}
        self.load_config(config if config else {})

    def _connect(self):
        qcollect_addr = QCOLLECT_ADDR
        try:
            if 'qcollectPort' in self.config:
                qcollect_addr = ('', int(self.config['qcollectPort']))
        except TypeError:
            raise "Invalid qcollect port %s" % self.config['qcollectPort']

        self.log.debug("Connecting to qcollect at %s", qcollect_addr)
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            sock.connect(qcollect_addr)
        except socket.error, msg:
            self.log.warn("Error connecting to qcollect TCP port: %s", msg)
            sys.exit(1)
        return sock

    def load_config(self, config):
        """
        Process a configfile, or reload if previously given one.
        """
        # Load in the collector's defaults
        if self.get_default_config() is not None:
            self.config.update(self.get_default_config())

        # Inject keys to collector's configuration.
        #
        # "enabled" is requied to be compatible with
        # diamond configuration. There are collectors that
        # check if they are enabled.
        #
        # We use "qcollectPort" in collectors to connect
        # to the running qcollect instance.
        self.config['enabled'] = True
        self.config['qcollectPort'] = config['qcollectPort']
        self.config['interval'] = self.configfile.get('interval', config['interval'])

        self.config.update(config.get('defaultConfig', {}))
        self.config.update(self.configfile)
        self.process_config()

    def can_publish_metric(self):
        return self._socket is not None and self._reconnect is False

    def process_config(self):
        """
        Intended to put any code that should be run after any config reload
        event
        """
        if 'byte_unit' in self.config:
            if isinstance(self.config['byte_unit'], basestring):
                self.config['byte_unit'] = self.config['byte_unit'].split()

        if 'enabled' in self.config:
            self.config['enabled'] = str_to_bool(self.config['enabled'])

        if 'measure_collector_time' in self.config:
            self.config['measure_collector_time'] = str_to_bool(
                self.config['measure_collector_time'])

        # Raise an error if both whitelist and blacklist are specified
        if (self.config.get('metrics_whitelist', None)
                and self.config.get('metrics_blacklist', None)):
            raise DiamondException(
                'Both metrics_whitelist and metrics_blacklist specified ' +
                'in file %s' % self.configfile)

        if self.config.get('metrics_whitelist', None):
            self.config['metrics_whitelist'] = re.compile(
                self.config['metrics_whitelist'])
        elif self.config.get('metrics_blacklist', None):
            self.config['metrics_blacklist'] = re.compile(
                self.config['metrics_blacklist'])

        if 'max_buffer_size' in self.config:
            self.config['max_buffer_size'] = int(self.config['max_buffer_size'])

    def get_default_config_help(self):
        """
        Returns the help text for the configuration options for this collector
        """
        return {
            'enabled': 'Enable collecting these metrics',
            'byte_unit': 'Default numeric output(s)',
            'measure_collector_time': 'Collect the collector run time in ms',
            'metrics_whitelist': 'Regex to match metrics to transmit. ' +
                                 'Mutually exclusive with metrics_blacklist',
            'metrics_blacklist': 'Regex to match metrics to block. ' +
                                 'Mutually exclusive with metrics_whitelist',
        }

    def get_default_config(self):
        """
        Return the default config for the collector
        """
        return {
            # Defaults options for all Collectors

            # Uncomment and set to hardcode a hostname for the collector path
            # Keep in mind, periods are seperators in graphite
            # 'hostname': 'my_custom_hostname',

            # If you perfer to just use a different way of calculating the
            # hostname
            # Uncomment and set this to one of these values:
            # fqdn_short  = Default. Similar to hostname -s
            # fqdn        = hostname output
            # fqdn_rev    = hostname in reverse (com.example.www)
            # uname_short = Similar to uname -n, but only the first part
            # uname_rev   = uname -r in reverse (com.example.www)
            # 'hostname_method': 'fqdn_short',

            # All collectors are disabled by default
            'enabled': False,

            # Path Prefix
            'path_prefix': 'servers',

            # Path Prefix for Virtual Machine metrics
            'instance_prefix': 'instances',

            # Path Suffix
            'path_suffix': '',

            # Default Poll Interval (seconds)
            'interval': 300,

            # Default Event TTL (interval multiplier)
            'ttl_multiplier': 2,

            # Default numeric output
            'byte_unit': 'byte',

            # Collect the collector run time in ms
            'measure_collector_time': False,

            # Whitelist of metrics to let through
            'metrics_whitelist': None,

            # Blacklist of metrics to let through
            'metrics_blacklist': None,

            # Max buffer size before flushing to qcollect core
            'max_buffer_size': 300,
        }

    def get_metric_path(self, name, instance=None):
        """
        Get metric path.
        Instance indicates that this is a metric for a
            virtual machine and should have a different
            root prefix.
        """
        if 'path' in self.config:
            path = self.config['path']
        else:
            path = self.__class__.__name__

        if instance is not None:
            if 'instance_prefix' in self.config:
                prefix = self.config['instance_prefix']
            else:
                prefix = 'instances'
            if path == '.':
                return '.'.join([prefix, instance, name])
            else:
                return '.'.join([prefix, instance, path, name])

        if 'path_prefix' in self.config:
            prefix = self.config['path_prefix']
        else:
            prefix = 'systems'

        if 'path_suffix' in self.config:
            suffix = self.config['path_suffix']
        else:
            suffix = None

        # if there is a suffix, add after the hostname
        if suffix:
            prefix = '.'.join((prefix, suffix))

        if path == '.':
            return '.'.join([prefix, name])
        else:
            return '.'.join([prefix, path, name])

    def collect(self):
        """
        Default collector method
        """
        raise NotImplementedError()

    def publish(self, name, value, raw_value=None, precision=0,
                metric_type='GAUGE', instance=None):
        """
        Publish a metric with the given name
        """
        # Check whitelist/blacklist
        if self.config['metrics_whitelist']:
            if not self.config['metrics_whitelist'].match(name):
                return
        elif self.config['metrics_blacklist']:
            if self.config['metrics_blacklist'].match(name):
                return

        # Get metric Path
        path = self.get_metric_path(name, instance=instance)

        # Get metric TTL
        ttl = float(self.config['interval']) * float(
            self.config['ttl_multiplier'])

        dimensions = None
        if self.dimensions is not None:
            dimensions = self.dimensions
            self.dimensions = None

        # Create Metric
        try:
            metric = Metric(path, value, raw_value=raw_value, timestamp=None,
                            precision=precision,
                            metric_type=metric_type, ttl=ttl, dimensions=dimensions)
        except DiamondException:
            self.log.error(('Error when creating new Metric: path=%r, '
                            'value=%r'), path, value)
            raise

        # Publish Metric
        self.publish_metric(metric)

    def publish_metric(self, metric):
        """
        Publish a Metric object

        We will send a payload that is setup specifically for
        qcollect. We prioritize the raw_value but some collectors
        don't set that - so we'll fall back on the value.
        """
        value = metric.raw_value
        if value is None:
            value = metric.value

        payload = {
            'name': metric.getMetricPath(),
            'value': value,
            'type': metric.metric_type,
            'dimensions': {
                'prefix': metric.getPathPrefix(),
                'collector': metric.getCollectorPath(),
                'collectorCanonicalName': self.name,
            }
        }

        payload['dimensions'].update(
            self.default_dimensions or {}
        )
        payload['dimensions'].update(
            metric.dimensions or {}
        )
        self.payload.append(payload)
        if len(self.payload) >= self.config.get('max_buffer_size', 300):
            self.flush()

    def flush(self):
        try:
            payloadStr = "%s\n" % json.dumps(self.payload)
        finally:
            self.payload = []

        success = False

        for i in range(QCOLLECT_RETRY_COUNT):
            try:
                if not self._socket or self._reconnect is True:
                    self._socket = self._connect()
                    self._reconnect = False
                    self.log.debug("Successfully reconnected")

                self._socket.sendall(payloadStr)
                success = True
                self.log.debug("Attempt %d: Wrote: %s" % (i, payloadStr))
                break
            except socket.error, e:
                self.log.exception("Error sending payload on attempt %d. "
                                   "We will reconnect. Payload: %s", (i, payloadStr))
                self._reconnect = True

        if not success:
            self.log.warn("After %d attempts failed to write payload %s", (QCOLLECT_RETRY_COUNT, payloadStr))

    def publish_cumulative_counter(self, name, value, precision=0, instance=None):
        return self.publish(name, value, precision=precision,
                            metric_type='CUMCOUNTER', instance=instance)

    def publish_gauge(self, name, value, precision=0, instance=None):
        return self.publish(name, value, precision=precision,
                            metric_type='GAUGE', instance=instance)

    def publish_counter(self, name, value, precision=0, max_value=0,
                        time_delta=True, interval=None, allow_negative=False,
                        instance=None):
        raw_value = value
        value = self.derivative(name, value, max_value=max_value,
                                time_delta=time_delta, interval=interval,
                                allow_negative=allow_negative,
                                instance=instance)
        return self.publish(name, value, raw_value=raw_value,
                            precision=precision, metric_type='COUNTER',
                            instance=instance)

    def derivative(self, name, new, max_value=0,
                   time_delta=True, interval=None,
                   allow_negative=False, instance=None):
        """
        Calculate the derivative of the metric.
        """
        # Format Metric Path
        path = self.get_metric_path(name, instance=instance)

        if path in self.last_values:
            old = self.last_values[path]
            # Check for rollover
            if new < old:
                old = old - max_value
            # Get Change in X (value)
            derivative_x = new - old

            # If we pass in a interval, use it rather then the configured one
            if interval is None:
                interval = int(self.config['interval'])

            # Get Change in Y (time)
            if time_delta:
                derivative_y = interval
            else:
                derivative_y = 1

            result = float(derivative_x) / float(derivative_y)
            if result < 0 and not allow_negative:
                result = 0
        else:
            result = 0

        # Store Old Value
        self.last_values[path] = new

        # Return result
        return result

    def _run(self):
        """
        Run the collector unless it's already running
        """
        try:
            start_time = time.time()

            # Collect Data
            self.default_dimensions = None
            self.collect()

            end_time = time.time()
            collector_time = int((end_time - start_time) * 1000)

            self.log.debug('Collection took %s ms', collector_time)

            if 'measure_collector_time' in self.config:
                if self.config['measure_collector_time']:
                    metric_name = 'collector_time_ms'
                    metric_value = collector_time
                    self.publish(metric_name, metric_value)
        except Exception as e:
            report_error(sys.exc_info()[0], self)
            # Now that we reported it let's raise the exception
            # so it can be cought by the scheduler
            raise e
        finally:
            # After collector run, invoke a flush
            # method on each handler.
            self.flush()
            self.default_dimensions = None
            for handler in self.handlers:
                handler._flush()

    def find_binary(self, binary):
        """
        Scan and return the first path to a binary that we can find
        """
        if os.path.exists(binary):
            return binary

        # Extract out the filename if we were given a full path
        binary_name = os.path.basename(binary)

        # Gather $PATH
        search_paths = os.environ['PATH'].split(':')

        # Extra paths to scan...
        default_paths = [
            '/usr/bin',
            '/bin'
            '/usr/local/bin',
            '/usr/sbin',
            '/sbin'
            '/usr/local/sbin',
        ]

        for path in default_paths:
            if path not in search_paths:
                search_paths.append(path)

        for path in search_paths:
            if os.path.isdir(path):
                filename = os.path.join(path, binary_name)
                if os.path.exists(filename):
                    return filename

        return binary


class ProcessCollector(Collector):
    """
    Collector with helpers for handling running commands with/without sudo
    """

    def get_default_config_help(self):
        config_help = super(ProcessCollector, self).get_default_config_help()
        config_help.update({
            'use_sudo':     'Use sudo?',
            'sudo_cmd':     'Path to sudo',
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(ProcessCollector, self).get_default_config()
        config.update({
            'use_sudo':     False,
            'sudo_cmd':     self.find_binary('/usr/bin/sudo'),
        })
        return config

    def run_command(self, args):
        if 'bin' not in self.config:
            raise Exception('config does not have any binary configured')
        if not os.access(self.config['bin'], os.X_OK):
            raise Exception('%s is not executable' % self.config['bin'])
        try:
            command = args
            command.insert(0, self.config['bin'])

            if str_to_bool(self.config['use_sudo']):
                command.insert(0, self.config['sudo_cmd'])

            return subprocess.Popen(command,
                                    stdout=subprocess.PIPE).communicate()
        except OSError:
            self.log.exception("Unable to run %s", command)
            return None
