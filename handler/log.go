package handler

import (
	"github.com/qnib/qcollect/metric"

	"encoding/json"
	"fmt"
	"time"

	l "github.com/Sirupsen/logrus"
)

func init() {
	RegisterHandler("Log", newLog)
}

// Log type
type Log struct {
	BaseHandler
}

// newLog returns a new Debug handler.
func newLog(
	channel chan metric.Metric,
	initialInterval int,
	initialBufferSize int,
	initialTimeout time.Duration,
	log *l.Entry) Handler {

	inst := new(Log)
	inst.name = "Log"

	inst.interval = initialInterval
	inst.maxBufferSize = initialBufferSize
	inst.log = log
	inst.channel = channel

	return inst
}

// Configure accepts the different configuration options for the Log handler
func (h *Log) Configure(configMap map[string]interface{}) {
	h.configureCommonParams(configMap)
}

// Run runs the handler main loop
func (h *Log) Run() {
	h.run(h.emitMetrics)
}

func (h Log) convertToLog(incomingMetric metric.Metric) (string, error) {
	jsonOut, err := json.Marshal(incomingMetric)
	return string(jsonOut), err
}

func (h *Log) emitMetrics(metrics []metric.Metric) bool {
	h.log.Info("Starting to emit ", len(metrics), " metrics")

	if len(metrics) == 0 {
		h.log.Warn("Skipping send because of an empty payload")
		return false
	}

	for _, m := range metrics {
		if dpString, err := h.convertToLog(m); err != nil {
			h.log.Error(fmt.Sprintf("Cannot convert metric %q to JSON: %s", m, err))
			continue
		} else {
			h.log.Info(dpString)
		}
	}
	return true
}
