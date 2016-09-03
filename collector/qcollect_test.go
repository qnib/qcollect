package collector

import (
	"github.com/qnib/qcollect/test_utils"

	"github.com/qnib/qcollect/metric"

	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQcollectConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})

	f := newQcollect(nil, 123, nil)
	f.Configure(config)

	assert.Equal(t,
		f.Interval(),
		123,
		"should be the default collection interval",
	)
}

func TestQcollectConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999

	f := newQcollect(nil, 123, nil)
	f.Configure(config)

	assert.Equal(t,
		f.Interval(),
		9999,
		"should be the defined interval",
	)
}

func TestQcollectCollect(t *testing.T) {
	config := make(map[string]interface{})

	testChannel := make(chan metric.Metric)
	testLog := test_utils.BuildLogger()

	f := newQcollect(testChannel, 123, testLog)
	f.Configure(config)

	go f.Collect()

	select {
	case <-f.Channel():
		return
	case <-time.After(2 * time.Second):
		t.Fail()
	}
}
