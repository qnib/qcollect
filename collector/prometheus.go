package collector

import (
	//"encoding/json"
	"fmt"
	"reflect"
	"sync"
    "strconv"
    //"log"
    //"os"

    p2jm "github.com/qnib/prom2json/lib"
	dto "github.com/prometheus/client_model/go"
	l "github.com/Sirupsen/logrus"
	"github.com/qnib/qcollect/metric"
)

const (
	pEndpoint = "http://localhost:3376/metrics"
)

// Prometheus collector type.
// Fetches metrics from Prometheus endpoint
type Prometheus struct {
	baseCollector
	endpoint      string
	mu            *sync.Mutex
}


func init() {
	RegisterCollector("Prometheus", newPrometheus)
}

// newPrometheus creates a new Prometheus collector.
func newPrometheus(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
    p := new(Prometheus)

	p.log = log
	p.channel = channel
	p.interval = initialInterval
	p.mu = new(sync.Mutex)

	p.name = "Prometheus"
	return p
}

// GetEndpoint Returns endpoint of instance
func (p *Prometheus) GetEndpoint() string {
	return p.endpoint
}

// Configure takes a dictionary of values with which the handler can configure itself.
func (p *Prometheus) Configure(configMap map[string]interface{}) {
    if pEP, exists := configMap["prometheusEndpoint"]; exists {
		if str, ok := pEP.(string); ok {
			p.endpoint = str
		} else {
			p.log.Warn("Failed to cast dokerEndPoint: ", reflect.TypeOf(pEP))
		}
	} else {
		p.endpoint = pEndpoint
	}
	p.configureCommonParams(configMap)
}

func convertMetric(mf *p2jm.Family) (float64) {
    m, _ := strconv.ParseFloat(mf.Metrics[0].(p2jm.Metric).Value, 64)
    return m
}

// TransformMetric consumes MatricFamily and returns slice of metrics
func (p *Prometheus) TransformMetric(mf *p2jm.Family) ([]metric.Metric) {
	var metrics []metric.Metric
	m := metric.New(mf.Name)
	m.MetricType = mf.Type
	if m.MetricType == "GAUGE" {
		m.Value = convertMetric(mf)
	} else if m.MetricType == "COUNTER" {
        m.Value = convertMetric(mf)
	} else if m.MetricType == "SUMMARY" {
		m = metric.New(fmt.Sprintf("%s_sum", m.Name))
		mc := metric.New(fmt.Sprintf("%s_count", m.Name))
		m.MetricType = mf.Type
		mc.MetricType = mf.Type
		s := mf.Metrics[0].(p2jm.Summary)
		for k,v := range s.Labels {
			m.AddDimension(k, v)
			mc.AddDimension(k, v)
		}
		for qk, qv := range s.Quantiles {
			mTemp := metric.New(fmt.Sprintf("%s_q%s", mf.Name, qk))
			mTemp.MetricType = "QUANTILE"
			for lk,lv := range s.Labels {
				mTemp.AddDimension(lk, lv)
			}
			mTemp.Value, _ = strconv.ParseFloat(qv, 64)
			metrics = append(metrics, mTemp)
		}
		m.Value, _ = strconv.ParseFloat(s.Sum, 64)
		mc.Value, _ = strconv.ParseFloat(s.Count, 64)
		metrics = append(metrics, mc)
	/*} else if f.Type == "HISTOGRAM" {
		//create histogram metrics?
		continue
	*/
	} else {
		p.log.Debugf("Dunno what to do with '%s'", mf.Type)
	}
	return metrics
}

// Collect fetches the endpoint and adds it to the channel
func (p *Prometheus) Collect() {
    mfChan := make(chan *dto.MetricFamily, 1024)
	go p2jm.FetchMetricFamilies(p.endpoint, mfChan)

	var f *p2jm.Family
    for mf := range mfChan {
        f = p2jm.NewFamily(mf)
        ms := p.TransformMetric(f)
		for _, m := range ms {
			p.Channel() <- m
		}
    }
}
