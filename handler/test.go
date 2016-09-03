package handler

import (
	"github.com/qnib/qcollect/metric"

	"time"

	l "github.com/Sirupsen/logrus"
)

// Test type
type Test struct {
	BaseHandler
}

// NewTest returns a new Test handler.
func NewTest(
	channel chan metric.Metric,
	initialInterval int,
	initialBufferSize int,
	initialTimeout time.Duration,
	log *l.Entry) Handler {

	inst := new(Test)
	inst.name = "Test"

	inst.interval = initialInterval
	inst.maxBufferSize = initialBufferSize
	inst.log = log
	inst.channel = channel

	return inst
}

// Configure accepts the different configuration options for the Test handler
func (h *Test) Configure(configMap map[string]interface{}) {
	h.configureCommonParams(configMap)
}

// Run runs the handler main loop
func (h *Test) Run() {
	h.run(h.emitMetrics)
}

func (h *Test) emitMetrics(metrics []metric.Metric) bool {
	h.log.Info("Starting to emit ", len(metrics), " metrics")

	if len(metrics) == 0 {
		h.log.Warn("Skipping send because of an empty payload")
		return false
	}

	return true
}
