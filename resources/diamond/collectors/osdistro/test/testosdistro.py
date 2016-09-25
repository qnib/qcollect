#!/usr/bin/python
# coding=utf-8
################################################################################

from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from mock import Mock
from mock import patch

from diamond.collector import Collector
from osdistro import OSDistroCollector

################################################################################


class TestOSDistroCollector(CollectorTestCase):
    def setUp(self):
        config = get_collector_config('OSDistroCollector', {
        })

        self.collector = OSDistroCollector(config, None)

    def test_import(self):
        self.assertTrue(OSDistroCollector)

    @patch.object(Collector, 'publish')
    def test_should_work_with_real_data(self, publish_mock):

        with patch('osdistro.Popen') as process_mock:
            with patch.object(process_mock.return_value, 'communicate') as comm_mock:
                comm_mock.return_value = [self.getFixture('ubuntu').getvalue(), '']
                self.collector.collect()

        self.assertPublishedMany(publish_mock, {
            'OSDistribution': 1
        })

    @patch.object(Collector, 'publish_metric')
    def test_sent_dimensions(self, publish_metric_mock):

        with patch('osdistro.Popen') as process_mock:
            with patch.object(process_mock.return_value, 'communicate') as comm_mock:
                comm_mock.return_value = [self.getFixture('ubuntu').getvalue(), '']
                self.collector.collect()

        for call in publish_metric_mock.mock_calls:
            name, args, kwargs = call
            metric = args[0]
            self.assertEquals(metric.dimensions, {
                'os_distro': self.getFixture('ubuntu').getvalue().strip()
            })

    @patch.object(Collector, 'publish')
    def test_should_fail_gracefully(self, publish_mock):

        with patch('osdistro.Popen') as process_mock:
            with patch.object(process_mock.return_value, 'communicate') as comm_mock:
                with patch.object(self.collector.log, 'error') as error_logger:
                    comm_mock.return_value = [None, 'Failed to find os_release']
                    self.collector.collect()

        error_logger.assert_called_once_with('Could not run lsb_release: Failed to find os_release')
        self.assertPublishedMany(publish_mock, {})



################################################################################
if __name__ == "__main__":
    unittest.main()
