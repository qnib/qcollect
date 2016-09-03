package collector

import (
	"github.com/qnib/qcollect/metric"

	"math/rand"
	"time"

	l "github.com/Sirupsen/logrus"
)

type valueGenerator func() float64

func generateRandomValue() float64 {
	return rand.Float64()
}

// Test collector type
type Test struct {
	baseCollector
	metricName   string
	bufferMetric bool
	generator    valueGenerator
}

func init() {
	RegisterCollector("Test", NewTest)
}

// NewTest creates a new Test collector.
func NewTest(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	t := new(Test)

	t.log = log
	t.channel = channel
	t.interval = initialInterval

	t.name = "Test"
	t.metricName = "TestMetric"
	t.generator = generateRandomValue
	t.bufferMetric = false
	return t
}

// Configure this takes a dictionary of values with which the handler can configure itself
func (t *Test) Configure(configMap map[string]interface{}) {
	if metricName, exists := configMap["metricName"]; exists {
		t.metricName = metricName.(string)
	}
	if bufferMetric, exists := configMap["bufferMetric"]; exists {
		t.bufferMetric = bufferMetric.(bool)
	}
	t.configureCommonParams(configMap)
}

// Collect produces some random test metrics.
func (t Test) Collect() {
	metric := metric.New(t.metricName)
	metric.Value = t.generator()
	metric.Buffered = t.bufferMetric
	metric.AddDimension("testing", "yes")
	time.Sleep(3 * time.Second)
	t.Channel() <- metric
	t.log.Debug(metric)
}
