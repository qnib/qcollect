package collector

import (
	"fmt"
	"testing"

	l "github.com/Sirupsen/logrus"
	"github.com/qnib/qcollect/metric"
	"github.com/stretchr/testify/assert"
)

func pContains(t *testing.T, metrics []metric.Metric, other metric.Metric) bool {
	mdone := false
	for _, my := range metrics {
		if my.Name == other.Name {
			fmt.Printf("%s => my.Buffered:%v | other.Buffered:%v\n", my.Name, my.Buffered, other.Buffered)
			assert.Equal(t, my.MetricType, other.MetricType)
			assert.Equal(t, my.Value, other.Value)
			assert.Equal(t, my.Dimensions, other.Dimensions)
			assert.Equal(t, my.GetTime(), other.GetTime())
			assert.Equal(t, my.Buffered, other.Buffered)
			mdone = true
		}
	}
	if !mdone {
		assert.True(t, false, fmt.Sprintf("%s not found in metrics", other.Name))
	}
	return true
}

func TestPrometheusNewPrometheus(t *testing.T) {
	expectedChan := make(chan metric.Metric)
	var expectedLogger = defaultLog.WithFields(l.Fields{"collector": "qcollect"})
	//expectedType := make(map[string]*CPUValues)

	c := newPrometheus(expectedChan, 10, expectedLogger).(*Prometheus)

	assert.Equal(t, c.log, expectedLogger)
	assert.Equal(t, c.channel, expectedChan)
	assert.Equal(t, c.interval, 10)
	assert.Equal(t, c.name, "Prometheus")
	c.Configure(make(map[string]interface{}))
	assert.Equal(t, c.GetEndpoint(), pEndpoint)
}

func TestPrometheusConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})

	c := newPrometheus(nil, 123, nil).(*Prometheus)
	c.Configure(config)

	assert.Equal(t, 123, c.Interval())
    assert.Equal(t, "http://localhost:3376/metrics", c.GetEndpoint())

}

func TestPrometheusConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999
    config["prometheusEndpoint"] = "http://localhost:13376/metrics"

	c := newPrometheus(nil, 123, nil).(*Prometheus)
	c.Configure(config)

	assert.Equal(t, 9999, c.Interval())
    assert.Equal(t, "http://localhost:13376/metrics", c.GetEndpoint())
}

/*func TestTransformMetric(t *testing.T) {

}*/
//func TestPrometheusCollect(t *testing.T) {
