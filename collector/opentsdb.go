package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/qnib/qcollect/metric"

	l "github.com/Sirupsen/logrus"
)

const (
	//DefaultOpenTSDBCollectorPort is the TCP port that OpenTSDB clients will push data to
	DefaultOpenTSDBCollectorPort = "4242"
	// MetricRegex provides the default OpenTSDB regex
	MetricRegex = `put (?P<name>[0-9\.\-\_a-zA-Z]+)\s+(?P<dimensions>[0-9\.\-\_\=\,a-zA-Z]+)\s+(?P<time>\d+)\s+(?P<value>[0-9\.]+)`
)

// OpenTSDB collector type
type OpenTSDB struct {
	baseCollector
	port          string
	serverStarted bool
	metricRegex   *regexp.Regexp
	incoming      chan string
}

func init() {
	RegisterCollector("OpenTSDB", newOpenTSDB)
}

// newOpenTSDB creates a new OpenTSDB collector.
func newOpenTSDB(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	d := new(OpenTSDB)

	d.log = log
	d.channel = channel
	d.interval = initialInterval

	d.name = "OpenTSDB"
	d.incoming = make(chan string)
	d.port = DefaultOpenTSDBCollectorPort
	d.serverStarted = false
	d.SetCollectorType("listener")
	return d
}

// Configure the collector
func (c *OpenTSDB) Configure(configMap map[string]interface{}) {
	if port, exists := configMap["port"]; exists {
		c.port = port.(string)
	}
	if regex, exists := configMap["metric-regex"]; exists {
		c.metricRegex = regexp.MustCompile(regex.(string))
	} else {
		c.metricRegex = regexp.MustCompile(MetricRegex)
	}
	c.configureCommonParams(configMap)
}

// Port returns collectors listen port
func (c *OpenTSDB) Port() string {
	return c.port
}

// collectOpenTSDB opens up and reads from the a TCP socket and
// writes what it's read to a local channel.

func (c *OpenTSDB) collectOpenTSDB() {
	addr, err := net.ResolveTCPAddr("tcp", ":"+c.port)

	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		c.log.Fatal("Cannot listen on OpenTSDB socket", err)
	}

	// figure out the port bind for Port()
	c.port = strings.Split(l.Addr().String(), ":")[1]

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			c.log.Fatal(err)
		}
		go c.readOpenTSDBMetrics(conn)
	}
}

// readOpenTSDBMetrics reads from the connection
func (c *OpenTSDB) readOpenTSDBMetrics(conn *net.TCPConn) {
	defer conn.Close()
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second)
	reader := bufio.NewReader(conn)
	c.log.Info("Connection started: ", conn.RemoteAddr())
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			c.log.Warn("Error while reading OpenTSDB metrics", err)
			break
		}
		c.log.Debug("Read: ", line)
		c.incoming <- bytes.NewBuffer(line).String()
	}
	c.log.Info("Connection closed: ", conn.RemoteAddr())
}

// Collect reads metrics collected from OpenTSDB collectors, converts
// them to qcollect's Metric type and publishes them to handlers.
func (c *OpenTSDB) Collect() {
	if !c.serverStarted {
		c.serverStarted = true
		go c.collectOpenTSDB()
	}

	for line := range c.incoming {
		if metric, ok := c.parseMetric(string(line)); ok {
			c.Channel() <- metric
		}
	}
}

func (c *OpenTSDB) parseMetric(line string) (metric.Metric, bool) {
	match := c.metricRegex.FindStringSubmatch(line)
	if match == nil {
		msg := fmt.Sprintf("could not match '%s' against regex", line)
		return metric.New(msg), false
	}
	dims := map[string]string{}
	for _, item := range strings.Split(match[2], ",") {
		i := strings.Split(item, "=")
		dims[i[0]] = i[1]
	}
	i, err := strconv.ParseInt(match[3], 10, 64)
	if err != nil {
		msg := fmt.Sprintf("Not an UNIX epoch '%s'", match[3])
		return metric.New(msg), false
	}
	tm := time.Unix(i, 0)
	v, _ := strconv.Atoi(match[4])
	m := metric.NewExt(match[1], "gauge", float64(v), dims, tm, false)
	return m, true
}
