package collector

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
    "log"
    "os"


    p2jm "github.com/qnib/prom2json/lib"
	dto "github.com/prometheus/client_model/go"
	l "github.com/Sirupsen/logrus"
	"github.com/qnib/qcollect/metric"
)

const (
	pEndpoint = "http:///var/run/docker.sock"
)

// Prometheus collector type.
// Fetches metrics from Prometheus endpoint
type Prometheus struct {
	baseCollector
	endpoint      string
	mu            *sync.Mutex
}


func init() {
	RegisterCollector("DockerStats", newDockerStats)
}

// newPrometheus creates a new DockerStats collector.
func newPrometheus(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
    p := new(Prometheus)

	p.log = log
	p.channel = channel
	p.interval = initialInterval
	p.mu = new(sync.Mutex)

	p.name = "Prometheus"
	return p
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
	//families := p.getPrometheusMetrics()
	/*if err != nil {
		p.log.Error("Error while collecting metrics: ", err)
		return
	}*/
	m := metric.New("test")
	m.Value = 1
	/*
    metric.AddDimension("model", model)

    for mf := range families {
	    p.Channel() <- familyToMetric(mf)
	    p.log.Debug(mf)
    }*/
    p.Channel() <- m
}

func familyToMetric(mf p2jm.Family) (metric.Metric) {
    m := metric.New(mf.Name)
    /*if mf.Type == "GAUGE" {
        m.Value = mf.Metrics[0].Value
    } else if mf.Type == "COUNTER" {
        m.Value = mf.Metrics[0].Value
    }*/
    return m
}

// getPrometheusMetrics reads endpoint and parses it
func (p *Prometheus) getPrometheusMetrics() ([]*p2jm.Family) {
    mfChan := make(chan *dto.MetricFamily, 1024)

	go p2jm.FetchMetricFamilies(p.endpoint, mfChan)

	result := []*p2jm.Family{}
    for mf := range mfChan {
		result = append(result, p2jm.NewFamily(mf))
	}
	json, err := json.Marshal(result)
	if err != nil {
		log.Fatalln("error marshaling JSON:", err)
	}
	if _, err := os.Stdout.Write(json); err != nil {
		log.Fatalln("error writing to stdout:", err)
	}
	fmt.Println()
    return result
}
