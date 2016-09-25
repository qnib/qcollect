# coding=utf-8

"""
Collect icmp round trip times per hop

#### Dependencies

 * libparistraceroute1 (as paris-traceroute)

"""

import re

import diamond.collector
from subprocess import Popen, PIPE


class TracerouteCollector(diamond.collector.ProcessCollector):

    def get_default_config_help(self):
        config_help = super(TracerouteCollector, self).get_default_config_help()
        config_help.update({
            'bin':          "The path to the tracerouting library.",
            'destport':     "The target port number",
            'hosts':        "Hosts to run the traceroute command on",
            'protocol':     "The protocol to use for the traceroute pings (icmp, udp, tcp)",
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(TracerouteCollector, self).get_default_config()
        config.update({
            'path':     'traceroute',
            'hosts':    { "yelp":"yelp.com" },
            'protocol': 'icmp',
        })
        return config

    def collect(self):

        protocol_args = self._protocol_config()
        if not protocol_args:
            self.log.error(
                "Please specify a protocol for the traceroute,\n"
                + " options (icmp, tcp, udp)"
            )
            return None

        for pseudo_hostname, address in self.config.get('hosts', {}).iteritems():

            metric_name = '.'.join([
                pseudo_hostname,
                'RoundTripTime',
            ])

            if 'bin' not in self.config:
                self.log.error(
                    "Please specify the path of the canonical binary"
                )
                return None

            cmd = [self.config['bin'], '-nq1', '-w1', protocol_args, address]

            try:
                process = Popen(cmd, stdout=PIPE, stderr=PIPE)
                errors = process.stderr.readline()
                if errors:
                    self.log.error(
                        "Error running traceroute process: {0!s}".format(errors)
                    )
                    continue

                while True:
                    line = process.stdout.readline()
                    if not line:
                        break

                    # A hop contains: 
                    # hop, ip, rtt
                    # in that order.
                    hop_data = line.split()
                    if not hop_data or len(hop_data) not in [2, 3]:
                        continue

                    hop_number = ip = None
                    rtt = 0

                    try:
                        [hop_number, ip, rtt_ms] = hop_data
                        rtt = re.match('([0-9\.]+)ms', rtt_ms).group(1)
                    except ValueError as e:
                        [hop_number, ip] = hop_data

                    if hop_number is None or ip is None:
                        continue

                    rtt = float(rtt)
                    self.dimensions = {
                        'hop': hop_number,
                    }

                    if '*' not in ip:
                        self.dimensions['ip'] = ip

                    self.publish(metric_name, rtt)
            except Exception as e:
                self.log.error(
                    "Error running TracerouteCollector: {0!s}".format(e)
                )
                continue

    def _protocol_config(self):

        protocol = self.config['protocol'].lower()
        destport = self.config.get('destport', 80)

        if protocol == 'udp':
            protocol_args = '-U'
        elif protocol == 'tcp':
            protocol_args = '-Tp{0!s}'.format(destport)
        elif protocol == 'icmp':
            protocol_args = '-I'
        else:
            return None

        return protocol_args
