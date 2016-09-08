#!/usr/bin/python
# coding=utf-8

from mock import patch
from mock import Mock

from test import unittest
import configobj
import logging

from diamond.collector import Collector
from diamond.error import DiamondException


class BaseCollectorTest(unittest.TestCase):

    def tearDown(self):
        log = logging.getLogger("diamond.Collector")
        log.removeHandler(log.handlers[0])
        # Ensure that we aren't printing log messages to stdout in unit tests
        log.propagate = False

    def config_object(self):
        config = configobj.ConfigObj()
        config['server'] = {}
        config['server']['collectors_config_path'] = ''
        config['collectors'] = {}
        config['defaultConfig'] = { "conf": "val" }
        config['interval'] = 10
        config['qcollectPort'] = 0
        return config

    def configfile(self):
        return {
            'interval': 5,
            'alice': 'bob',
        }

    @patch('diamond.collector.Collector.publish_metric', autoSpec=True)
    def test_SetDimensions(self, mock_publish):
        c = Collector(self.config_object(), [])
        dimensions = {
            'dim1': 'alice',
            'dim2': 'chains',
        }
        c.dimensions = dimensions
        c.publish('metric1', 1)

        for call in mock_publish.mock_calls:
            name, args, kwargs = call
            metric = args[0]
            self.assertEquals(metric.dimensions, dimensions)
        self.assertEqual(c.dimensions, None)

    @patch('diamond.collector.Collector.publish_metric', autoSpec=True)
    def test_successful_error_metric(self, mock_publish):
        c = Collector(self.config_object(), [])
        mock_socket = Mock()
        c._socket = mock_socket
        with patch.object(c, 'log'):
            try:
                c.publish('metric', "bar")
            except DiamondException:
                pass
        for call in mock_publish.mock_calls:
            name, args, kwargs = call
            metric = args[0]
            self.assertEqual(metric.path, "servers.Collector.qcollect.collector_errors")

    @patch('diamond.collector.Collector.publish_metric', autoSpec=True)
    def test_failed_error_metric_publish(self, mock_publish):
        c = Collector(self.config_object(), [])
        self.assertFalse(c.can_publish_metric())
        with patch.object(c, 'log'):
            try:
                c.publish('metric', "baz")
            except DiamondException:
                pass
        self.assertEquals(len(mock_publish.mock_calls), 0)

    def test_can_publish_metric(self):
        c = Collector(self.config_object(), [])
        self.assertFalse(c.can_publish_metric())

        c._socket = "socket"
        self.assertTrue(c.can_publish_metric())

    def test_batch_size_flush(self):
        c = Collector(self.config_object(), [])
        mock_socket = Mock()
        c._socket = mock_socket
        c._reconnect = False
        c.config['max_buffer_size'] = 2
        with patch.object(c, 'log'):
            try:
                c.publish('metric1', 1)
                c.publish('metric2', 2)
                c.publish('metric3', 3)
            except DiamondException:
                pass
        self.assertEquals(mock_socket.sendall.call_count, 1)
        self.assertEquals(len(c.payload), 1)

    def test_configure_collector(self):
        c = Collector(self.config_object(), [], configfile=self.configfile())
        self.assertEquals(c.config, {
            'ttl_multiplier':2,
            'path_suffix':'',
            'measure_collector_time':False,
            'metrics_blacklist':None,
            'byte_unit':[
                'byte'
            ],
            'instance_prefix':'instances',
            'conf':'val',
            'qcollectPort':0,
            'interval':5,
            'enabled':True,
            'alice':'bob',
            'metrics_whitelist':None,
            'max_buffer_size':300,
            'path_prefix':'servers'
    })
