package main

import (
	"sync"

	"github.com/qnib/qcollect/collector"
	"github.com/qnib/qcollect/config"
	"github.com/qnib/qcollect/handler"
	"github.com/qnib/qcollect/metric"

	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testFakeConfiguration = `{
    "prefix": "test.",
    "interval": 10,
    "defaultDimensions": {
    },

	"collectorsConfigPath": "/tmp",
    "diamondCollectorsPath": "misc/diamond/collectors",
    "diamondCollectors": [],

    "collectors": ["FakeCollector","Test"],

    "handlers": {
    }
}
`

var testCollectorConfiguration = `{
	"metricName": "TestMetric",
	"interval": 10
}
`

var (
	tmpTestFakeFile, tempTestCollectorConfig string
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.ErrorLevel)
	if f, err := ioutil.TempFile("/tmp", "qcollect"); err == nil {
		f.WriteString(testFakeConfiguration)
		tmpTestFakeFile = f.Name()
		f.Close()
		defer os.Remove(tmpTestFakeFile)
	}
	if f, err := ioutil.TempFile("/tmp", "qcollect"); err == nil {
		f.WriteString(testCollectorConfiguration)
		tempTestCollectorConfig = f.Name() + ".conf"
		f.Close()
		defer os.Remove(tempTestCollectorConfig)
	}
	os.Exit(m.Run())
}

func TestStartCollectorsEmptyConfig(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	collectors := startCollectors(config.Config{})

	assert.NotEqual(t, len(collectors), 1, "should create a Collector")
}

func TestStartCollectorUnknownCollector(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	c := make(map[string]interface{})
	collector := startCollector("unknown collector", config.Config{}, c)

	assert.Nil(t, collector, "should NOT create a Collector")
}

func TestStartCollectorsMixedConfig(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	conf, _ := config.ReadConfig(tmpTestFakeFile)
	collectors := startCollectors(conf)

	for _, c := range collectors {
		assert.Equal(t, c.Name(), "Test", "Only create valid collectors")
	}
}

func TestStartCollectorTooLong(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	c := make(map[string]interface{})
	c["interval"] = 1
	collector := startCollector("Test", config.Config{}, c)

	select {
	case m := <-collector.Channel():
		assert.Equal(t, 1.0, m.Value)
		assert.Equal(t, "qcollect.collection_time_exceeded", m.Name)
		assert.Equal(t, "1", m.Dimensions["interval"])
		return
	case <-time.After(5 * time.Second):
		t.Fail()
	}
}

func TestReadFromCollector(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	c := make(map[string]interface{})
	c["interval"] = 1
	collector := collector.New("Test")
	collector.SetInterval(1)
	collector.Configure(c)

	var wg sync.WaitGroup
	wg.Add(2)
	collectorStatChannel := make(chan metric.CollectorEmission)
	go func() {
		defer wg.Done()
		collector.Channel() <- metric.New("hello")
		time.Sleep(time.Duration(2) * time.Second)
		m2 := metric.New("world")
		m2.AddDimension("collectorCanonicalName", "Foobar")
		collector.Channel() <- m2
		time.Sleep(time.Duration(2) * time.Second)
		m2.AddDimension("collectorCanonicalName", "Foobar")
		collector.Channel() <- m2
		close(collector.Channel())
	}()
	collectorMetrics := map[string]uint64{}
	go func() {
		defer wg.Done()
		for collectorMetric := range collectorStatChannel {
			collectorMetrics[collectorMetric.Name] = collectorMetric.EmissionCount
		}
	}()
	readFromCollector(collector, []handler.Handler{}, collectorStatChannel)
	wg.Wait()
	assert.Equal(t, uint64(1), collectorMetrics["Test"])
	assert.Equal(t, uint64(2), collectorMetrics["Foobar"])
}
