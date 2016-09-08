# coding=utf-8

"""
Collect stats from postfix-stats. postfix-stats is a simple threaded stats
aggregator for Postfix. When running as a syslog destination, it can be used to
get realtime cumulative stats.

#### Dependencies

 * socket
 * json (or simplejson)
 * [postfix-stats](https://github.com/disqus/postfix-stats)

"""

import socket
import sys

try:
    import json
except ImportError:
    import simplejson as json

import diamond.collector

from diamond.collector import str_to_bool


class PostfixCollector(diamond.collector.Collector):

    def get_default_config_help(self):
        config_help = super(PostfixCollector,
                            self).get_default_config_help()
        config_help.update({
            'host':             'Hostname to connect to',
            'port':             'Port to connect to',
            'include_clients':  'Include client connection stats',
            'relay_mode':       'Running postfix in relay mode?'
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(PostfixCollector, self).get_default_config()
        config.update({
            'path':             'postfix',
            'host':             'localhost',
            'port':             7777,
            'include_clients':  True,
            'relay_mode':       False,
        })
        return config

    def get_json(self):
        json_string = ''

        address = (self.config['host'], int(self.config['port']))

        s = None
        try:
            try:
                s = socket.create_connection(address, timeout=1)

                s.sendall('stats\n')

                while 1:
                    data = s.recv(4096)
                    if not data:
                        break
                    json_string += data
            except socket.error:
                self.log.exception("Error talking to postfix-stats")
                return '{}'
        finally:
            if s:
                s.close()

        return json_string or '{}'

    def get_data(self):
        json_string = self.get_json()

        try:
            data = json.loads(json_string)
        except (ValueError, TypeError):
            self.log.exception("Error parsing json from postfix-stats")
            return None

        return data

    def collect(self):
        data = self.get_data()

        if not data:
            return

        if str_to_bool(self.config['include_clients']):

            metric_name = 'postfix.incoming'
            if not str_to_bool(self.config['relay_mode']):
                for client, value in data.get('clients', {}).iteritems():
                    self.dimensions = {
                        'client': str(client),
                    }

                    self.publish_cumulative_counter(metric_name, value)
            else:

                for component, clients in data.get('relay_clients', {}).iteritems():
                    for client, value in clients.iteritems():
                        self.dimensions = {
                            'client': str(client),
                            'queue':str(component),
                        }
                        self.publish_cumulative_counter(metric_name, value)

        for action in (u'in', u'recv', u'send'):
            if action not in data:
                continue

            metric_name = '.'.join(['postfix', str(action)])
            for sect, components in data[action].iteritems():
                if not str_to_bool(self.config['relay_mode']):
                    if sect == 'relay_status':
                        continue

                    for status, value in components.iteritems():
                        self.dimensions = {
                            'status': str(status),
                        }
                        self.publish_cumulative_counter(metric_name, value)
                else:
                    if sect != 'relay_status':
                        continue

                    for component, stats in components.iteritems():
                        for status, value in stats.iteritems():
                            self.dimensions = {
                                'status': str(status),
                                'queue':str(component),
                            }

                            self.publish_cumulative_counter(metric_name, value)

        if u'local' in data:
            metric_name = 'postfix.local'
            for key, value in data[u'local'].iteritems():
                self.dimensions = {'address': str(key)}

                self.publish_cumulative_counter(metric_name, value)
