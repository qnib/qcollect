package handler

import (
	"github.com/qnib/qcollect/metric"

	"fmt"
	"testing"
	"time"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// waitForSplitSecond waits until the current ms timer is less then 500ms away
// from fliping the full second to avoid wrong UNIX timestamps while converting to graphite metrics
func waitForSplitSecond() {
	curNs := time.Now().Nanosecond()
	for curNs > 500000000 {
		time.Sleep(100 * time.Millisecond)
		curNs = time.Now().Nanosecond()
	}
}

func getTestGraphiteHandler(interval, buffsize, timeoutsec int) *Graphite {
	testChannel := make(chan metric.Metric)
	testLog := l.WithField("testing", "graphite_handler")
	timeout := time.Duration(timeoutsec) * time.Second

	return newGraphite(testChannel, interval, buffsize, timeout, testLog).(*Graphite)
}

func TestGraphiteConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	g := getTestGraphiteHandler(12, 13, 14)
	g.Configure(config)

	assert.Equal(t, 12, g.Interval())
}

func TestGraphiteConfigure(t *testing.T) {
	config := map[string]interface{}{
		"interval":        "10",
		"timeout":         "10",
		"max_buffer_size": "100",
		"server":          "test_server",
		"port":            "10101",
	}

	g := getTestGraphiteHandler(12, 13, 14)
	g.Configure(config)

	assert.Equal(t, 10, g.Interval())
	assert.Equal(t, 100, g.MaxBufferSize())
	assert.Equal(t, "test_server", g.Server())
	assert.Equal(t, "10101", g.Port())
}

func TestGraphiteConfigureIntPort(t *testing.T) {
	config := map[string]interface{}{
		"interval":        "10",
		"timeout":         "10",
		"max_buffer_size": "100",
		"server":          "test_server",
		"port":            10101,
	}

	g := getTestGraphiteHandler(12, 13, 14)
	g.Configure(config)

	assert.Equal(t, 10, g.Interval())
	assert.Equal(t, 100, g.MaxBufferSize())
	assert.Equal(t, "test_server", g.Server())
	assert.Equal(t, "10101", g.Port())
}

// TestConvertToGraphite tests the plain handler convertion
func TestConvertToGraphite(t *testing.T) {
	config := map[string]interface{}{
		"interval":        "10",
		"timeout":         "10",
		"max_buffer_size": "100",
		"server":          "test_server",
		"port":            10101,
	}

	g := getTestGraphiteHandler(12, 13, 14)
	g.Configure(config)

	waitForSplitSecond()
	now := time.Now().Unix()
	m := metric.New("TestMetric")

	dpString := g.convertToGraphite(m)

	assert.Equal(t, fmt.Sprintf("TestMetric 0.000000 %d\n", now), dpString)
}

// TestConvertToGraphiteDims tests the handler convertion with dimensions
func TestConvertToGraphiteDims(t *testing.T) {
	config := map[string]interface{}{
		"interval":        "10",
		"timeout":         "10",
		"max_buffer_size": "100",
		"server":          "test_server",
		"port":            10101,
	}

	g := getTestGraphiteHandler(12, 13, 14)
	g.Configure(config)

	waitForSplitSecond()
	now := time.Now().Unix()
	m := metric.New("TestMetric")

	dims := map[string]string{
		"container_id":   "test-id",
		"container_name": "test-container",
	}
	m.Dimensions = dims

	dpString := g.convertToGraphite(m)

	assert.Equal(t, fmt.Sprintf("TestMetric.container_id.test-id.container_name.test-container 0.000000 %d\n", now), dpString)
}

// TestConvertToGraphitePrefixKey tests the  handler convertion prefixed with all keys
func TestConvertToGraphitePrefixKey(t *testing.T) {
	config := map[string]interface{}{
		"interval":        "10",
		"timeout":         "10",
		"max_buffer_size": "100",
		"server":          "test_server",
		"prefixKeys":      true,
		"port":            10101,
	}

	g := getTestGraphiteHandler(12, 13, 14)
	g.Configure(config)

	waitForSplitSecond()
	now := time.Now().Unix()
	m := metric.New("TestMetric")

	dims := map[string]string{
		"container_id":   "test-id",
		"container_name": "test-container",
	}
	m.Dimensions = dims

	dpString := g.convertToGraphite(m)

	assert.Equal(t, fmt.Sprintf("container_id_container_name.TestMetric.container_id.test-id.container_name.test-container 0.000000 %d\n", now), dpString)
}
