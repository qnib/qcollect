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

// Collect fetches the endpoint and adds it to the channel
func (p *Prometheus) Collect() {
    mfChan := make(chan *dto.MetricFamily, 1024)
	go p2jm.FetchMetricFamilies(p.endpoint, mfChan)
    var f *p2jm.Family
    for mf := range mfChan {
        f = p2jm.NewFamily(mf)
        m := metric.New(f.Name)
        m.MetricType = f.Type
        if f.Type == "GAUGE" {
            m.Value, _ = strconv.ParseFloat(f.Metrics[0].(p2jm.Metric).Value, 64)
        } else if f.Type == "COUNTER" {
            m.Value, _ = strconv.ParseFloat(f.Metrics[0].(p2jm.Metric).Value, 64)
        } else if f.Type == "SUMMARY" {
            m = metric.New(fmt.Sprintf("%s_sum", f.Name))
            mc := metric.New(fmt.Sprintf("%s_count", f.Name))
            m.MetricType = f.Type
            mc.MetricType = f.Type
            s := f.Metrics[0].(p2jm.Summary)
            for k,v := range s.Labels {
                m.AddDimension(k, v)
                mc.AddDimension(k, v)
            }
            for qk, qv := range s.Quantiles {
                mTemp := metric.New(fmt.Sprintf("%s_q%s", f.Name, qk))
                mTemp.MetricType = "QUANTILE"
                for lk,lv := range s.Labels {
                    mTemp.AddDimension(lk, lv)
                }
                mTemp.Value, _ = strconv.ParseFloat(qv, 64)
                p.Channel() <- mTemp
            }
            m.Value, _ = strconv.ParseFloat(s.Sum, 64)
            mc.Value, _ = strconv.ParseFloat(s.Count, 64)
            p.Channel() <- mc
        /*} else if f.Type == "HISTOGRAM" {
            //create histogram metrics?
            continue
        */
        } else {
            p.log.Debugf("Dunno what to do with '%s'", f.Type)
            continue
        }
        p.Channel() <- m

    }
}
