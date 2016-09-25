# coding=utf-8

"""
Collect counters from scribe

#### Dependencies

    * /usr/sbin/scribe_ctrl, distributed with scribe

"""
import os
import subprocess
import string

import diamond.collector
from diamond.collector import str_to_bool


class ScribeCollector(diamond.collector.Collector):

    GAUGES = set([
        'buffer_size',
    ])

    def get_default_config_help(self):
        config_help = super(ScribeCollector, self).get_default_config_help()
        config_help.update({
            'exclude_pattern': 'Exclude items from scribe buffer that match this pattern',
            'buffer_path': 'Path to scribe buffer',
            'scribe_leaf': 'Is this a scribe leaf?',
            'scribe_ctrl_bin': 'Path to scribe_ctrl binary',
            'scribe_port': 'Scribe port',
        })
        return config_help

    def get_default_config(self):
        config = super(ScribeCollector, self).get_default_config()
        config.update({
            'exclude_pattern': None,
            'buffer_path': None,
            'path': 'scribe',
            'scribe_leaf': None,
            'scribe_ctrl_bin': self.find_binary('/usr/sbin/scribe_ctrl'),
            'scribe_port': None,
        })
        return config

    def key_to_metric(self, key):
        """Replace all non-letter characters with underscores"""
        return ''.join(l if l in string.letters else '_' for l in key)

    def get_scribe_ctrl_output(self):
        cmd = [self.config['scribe_ctrl_bin'], 'counters']

        if self.config['scribe_port'] is not None:
            cmd.append(self.config['scribe_port'])

        self.log.debug("Running command %r", cmd)

        try:
            p = subprocess.Popen(cmd, stdout=subprocess.PIPE,
                                 stderr=subprocess.PIPE)
        except OSError:
            self.log.error("Unable to run %r", cmd)
            return ""

        stdout, stderr = p.communicate()

        if p.wait() != 0:
            self.log.warning("Command failed %r", cmd)
            self.log.warning(stderr)

        return stdout

    def get_scribe_stats(self):
        data = {}

        if os.path.exists(self.config['scribe_ctrl_bin']):
            output = self.get_scribe_ctrl_output()


            for line in output.splitlines():
                key, val = line.rsplit(':', 1)
                metric = self.key_to_metric(key)
                data[metric] = int(val)

        if self.config['buffer_path']:
            cmd = ['du', '-sb', '--apparent-size', self.config['buffer_path']]
            if self.config['exclude_pattern']:
                cmd.append(
                    "--exclude={0!s}".format(self.config['exclude_pattern'])
                )
            try:
                p = subprocess.Popen(cmd, stdout=subprocess.PIPE,
                                     stderr=subprocess.PIPE)
                output, errors = p.communicate()
                if errors:
                    self.log.error(
                        "Error running {0!r}, {1!s}".format(cmd, errors)
                    )
                else:
                    data['buffer_size'] = int(output[:output.find('\t')])
            except OSError:
                self.log.error(
                    "Unable to run {0!r}".format(cmd)
                )

        return data.items()

    def collect(self):
        for stat, val in self.get_scribe_stats():
            metric_name = '.'.join(['scribe', stat])
            self.log.debug(
                "Publishing: {0} {1}".format(stat, val)
            )
            if str_to_bool(self.config['scribe_leaf']):
                self.dimensions = { 'node_type': 'leaf' }
            else:
                self.dimensions = { 'node_type': 'aggregator' }
            if stat in self.GAUGES:
                self.publish(metric_name, val)
            else:
                self.publish_cumulative_counter(metric_name, val)
