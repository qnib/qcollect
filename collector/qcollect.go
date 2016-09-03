package collector

import (
	"github.com/qnib/qcollect/metric"

	"runtime"

	l "github.com/Sirupsen/logrus"
)

type memStatRetriever func() *runtime.MemStats

// Qcollect collector type
type Qcollect struct {
	baseCollector
	memStats memStatRetriever
}

func init() {
	RegisterCollector("Qcollect", newQcollect)
}

// newQcollect creates a new Test collector.
func newQcollect(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	f := new(Qcollect)

	f.log = log
	f.channel = channel
	f.interval = initialInterval

	f.name = "Qcollect"
	f.memStats = getMemStats
	return f
}

// Configure this takes a dictionary of values with which the handler can configure itself
func (f *Qcollect) Configure(configMap map[string]interface{}) {
	f.configureCommonParams(configMap)
}

// Collect produces some random test metrics.
func (f Qcollect) Collect() {
	for _, m := range f.getGoMetrics() {
		f.Channel() <- m
	}
}

func (f Qcollect) getGoMetrics() []metric.Metric {
	m := f.memStats()

	ret := []metric.Metric{
		buildCounter("NumGoroutine", uint64(runtime.NumGoroutine())),
		buildSimpleMetric("Alloc", m.Alloc),
		buildCounter("TotalAlloc", m.TotalAlloc),
		buildSimpleMetric("Sys", m.Sys),
		buildCounter("Lookups", m.Lookups),
		buildCounter("Mallocs", m.Mallocs),
		buildCounter("Frees", m.Frees),
		buildSimpleMetric("HeapAlloc", m.HeapAlloc),
		buildSimpleMetric("HeapSys", m.HeapSys),
		buildSimpleMetric("HeapIdle", m.HeapIdle),
		buildSimpleMetric("HeapInuse", m.HeapInuse),
		buildSimpleMetric("HeapReleased", m.HeapReleased),
		buildSimpleMetric("HeapObjects", m.HeapObjects),
		buildSimpleMetric("StackInuse", m.StackInuse),
		buildSimpleMetric("StackSys", m.StackSys),
		buildSimpleMetric("MSpanInuse", m.MSpanInuse),
		buildSimpleMetric("MSpanSys", m.MSpanSys),
		buildSimpleMetric("MCacheInuse", m.MCacheInuse),
		buildSimpleMetric("MCacheSys", m.MCacheSys),
		buildSimpleMetric("BuckHashSys", m.BuckHashSys),
		buildSimpleMetric("GCSys", m.GCSys),
		buildSimpleMetric("OtherSys", m.OtherSys),
		buildSimpleMetric("NextGC", m.NextGC),
		buildSimpleMetric("LastGC", m.LastGC),
		buildCounter("PauseTotalNs", m.PauseTotalNs),
		buildCounter("NumGC", uint64(m.NumGC)),
	}

	return ret
}

// ----------------------------------------------------------------------------
// utility methods
// ----------------------------------------------------------------------------

// See https://golang.org/src/runtime/mstats.go?s=3251:5102#L82
func getMemStats() *runtime.MemStats {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	return stats
}

func buildSimpleMetric(name string, value uint64) (m metric.Metric) {
	m = metric.New(name)
	m.Value = float64(value)
	return m
}

func buildCounter(name string, value uint64) (m metric.Metric) {
	m = buildSimpleMetric(name, value)
	m.MetricType = metric.Counter
	return m
}
