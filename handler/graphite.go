package handler

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/qnib/qcollect/metric"

	l "github.com/Sirupsen/logrus"
)

func init() {
	RegisterHandler("Graphite", newGraphite)
}

// Graphite type
type Graphite struct {
	BaseHandler
	server     string
	port       string
	prefixKeys bool
}

// newGraphite returns a new Graphite handler.
func newGraphite(
	channel chan metric.Metric,
	initialInterval int,
	initialBufferSize int,
	initialTimeout time.Duration,
	log *l.Entry) Handler {

	inst := new(Graphite)
	inst.name = "Graphite"

	inst.interval = initialInterval
	inst.maxBufferSize = initialBufferSize
	inst.timeout = initialTimeout
	inst.log = log
	inst.channel = channel

	return inst
}

// Server returns the Graphite server's name or IP
func (g Graphite) Server() string {
	return g.server
}

// Port returns the Graphite server's port number
func (g Graphite) Port() string {
	return g.port
}

// Configure accepts the different configuration options for the Graphite handler
func (g *Graphite) Configure(configMap map[string]interface{}) {
	if server, exists := configMap["server"]; exists {
		g.server = server.(string)
	} else {
		g.log.Error("There was no server specified for the Graphite Handler, there won't be any emissions")
	}

	if port, exists := configMap["port"]; exists {
		g.port = fmt.Sprint(port)
	} else {
		g.log.Error("There was no port specified for the Graphite Handler, there won't be any emissions")
	}
	if prefixKeys, exists := configMap["prefixKeys"]; exists {
		g.prefixKeys = prefixKeys.(bool)
	}
	g.configureCommonParams(configMap)
}

// Run runs the handler main loop
func (g *Graphite) Run() {
	g.run(g.emitMetrics)
}

func (g Graphite) convertToGraphite(incomingMetric metric.Metric) (datapoint string) {
	//orders dimensions so datapoint keeps consistent name
	var keys []string
	dimensions := incomingMetric.GetDimensions(g.DefaultDimensions())
	for k := range dimensions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if g.prefixKeys && len(keys) > 0 {
		datapoint = g.Prefix() + strings.Join(keys, "_") + "." + incomingMetric.Name
	} else {
		datapoint = g.Prefix() + incomingMetric.Name
	}
	for _, key := range keys {
		datapoint = fmt.Sprintf("%s.%s.%s", datapoint, key, dimensions[key])
	}
	datapoint = fmt.Sprintf("%s %f %d\n", datapoint, incomingMetric.Value, incomingMetric.GetTime().Unix())
	return datapoint
}

func (g *Graphite) emitMetrics(metrics []metric.Metric) bool {
	g.log.Info("Starting to emit ", len(metrics), " metrics")

	if len(metrics) == 0 {
		g.log.Warn("Skipping send because of an empty payload")
		return false
	}

	addr := fmt.Sprintf("%s:%s", g.server, g.port)
	conn, err := net.DialTimeout("tcp", addr, g.timeout)
	if err != nil {
		g.log.Error("Failed to connect ", addr)
		return false
	}

	for _, m := range metrics {
		fmt.Fprintf(conn, g.convertToGraphite(m))
	}
	return true
}
