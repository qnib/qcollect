package collector

import (
	"github.com/qnib/qcollect/test_utils"

	"github.com/qnib/qcollect/metric"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})
	test := NewTest(nil, 123, nil).(*Test)
	test.Configure(config)

	assert.Equal(t,
		test.Interval(),
		123,
		"should be the default collection interval",
	)
}

func TestTestConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999

	// the channel and logger don't matter
	test := NewTest(nil, 12, nil).(*Test)
	test.Configure(config)

	assert.Equal(t,
		test.Interval(),
		9999,
		"should be the defined interval",
	)
}

func TestTestConfigureMetricName(t *testing.T) {
	config := make(map[string]interface{})
	config["metricName"] = "lala"

	testChannel := make(chan metric.Metric)
	testLogger := test_utils.BuildLogger()

	test := NewTest(testChannel, 123, testLogger).(*Test)
	test.Configure(config)

	go test.Collect()

	select {
	case m := <-test.Channel():
		// don't test for the value - only metric name
		assert.Equal(t, m.Name, "lala")
	case <-time.After(4 * time.Second):
		t.Fail()
	}
}

func TestTestCollect(t *testing.T) {
	config := make(map[string]interface{})

	testChannel := make(chan metric.Metric)
	testLogger := test_utils.BuildLogger()

	// conforms to the valueGenerator interface in the collector
	mockGen := func() float64 {
		return 4.0
	}

	test := NewTest(testChannel, 123, testLogger).(*Test)
	test.Configure(config)
	test.generator = mockGen

	go test.Collect()

	select {
	case m := <-test.Channel():
		assert.Equal(t, 4.0, m.Value)
		return
	case <-time.After(4 * time.Second):
		t.Fail()
	}
}
