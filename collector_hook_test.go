package main

import (
	"testing"
	"time"

	"github.com/qnib/qcollect/collector"
	"github.com/qnib/qcollect/handler"
	"github.com/qnib/qcollect/metric"
	"github.com/qnib/qcollect/test_utils"

	"github.com/stretchr/testify/assert"
)

func TestCollectorLogsErrors(t *testing.T) {
	testLogger := test_utils.BuildLogger()
	testLogger = testLogger.WithField("collector", "Test")

	channel := make(chan metric.Metric)
	config := make(map[string]interface{})

	testCol := collector.NewTest(channel, 123, testLogger)
	testCol.Configure(config)

	timeout := time.Duration(5 * time.Second)
	h := handler.NewTest(channel, 10, 10, timeout, testLogger)

	hook := NewLogErrorHook([]handler.Handler{h})
	testLogger.Logger.Hooks.Add(hook)

	go testCol.Collect()
	testLogger.Error("testing Error log")

	select {
	case m := <-h.Channel():
		assert.Equal(t, "qcollect.collector_errors", m.Name)
		assert.Equal(t, 1.0, m.Value)
		assert.Equal(t, "Test", m.Dimensions["collector"])
		return
	case <-time.After(1 * time.Second):
		t.Fail()
	}
}
