package collector

import (
	"github.com/qnib/qcollect/test_utils"

	"github.com/qnib/qcollect/metric"

	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	l "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiamondConfigureEmptyConfig(t *testing.T) {
	config := make(map[string]interface{})

	d := newDiamond(nil, 12, nil).(*Diamond)
	d.Configure(config)

	assert.Equal(t,
		d.Interval(),
		12,
		"should be the default collection interval",
	)
}

func TestDiamondConfigure(t *testing.T) {
	config := make(map[string]interface{})
	config["interval"] = 9999
	config["port"] = "0"
	d := newDiamond(nil, 12, nil).(*Diamond)
	d.Configure(config)

	assert := assert.New(t)
	assert.Equal(d.Interval(), 9999, "should be the defined interval")
	assert.Equal(d.Port(), "0", "should be the defined port")
}

func TestDiamondCollect(t *testing.T) {
	config := make(map[string]interface{})
	config["port"] = "0"

	testChannel := make(chan metric.Metric)
	testLog := test_utils.BuildLogger()

	d := newDiamond(testChannel, 123, testLog).(*Diamond)
	d.Configure(config)

	// start collecting Diamond metrics
	go d.Collect()

	conn, err := connectToDiamondCollector(d)
	require.Nil(t, err, "should connect")
	require.NotNil(t, conn, "should connect")

	emitTestMetric(conn)

	select {
	case m := <-d.Channel():
		assert.Equal(t, m.Name, "test")
	case <-time.After(1 * time.Second):
		t.Fail()
	}
}

func TestParseJsonToMetric(t *testing.T) {
	rawData := []byte(`
[{
   "name": "foobar",
   "type":  "GAUGE",
   "value": 100.0,
   "dimensions": {
      "host": "windrunner"
   }
}]
        `)
	d := newDiamond(nil, 12, nil).(*Diamond)
	metrics, ok := d.parseMetrics(rawData)
	assert.True(t, ok)
	for _, metric := range metrics {
		assert.Equal(t, "gauge", metric.MetricType)
	}
}

func TestInvalidJsonToMetric(t *testing.T) {
	rawData := []byte(`
[{
   "name": "foobar",
   "type":  "GAUGE",
   "value": "100.0",
   "dimensions": {
      "host": "windrunner"
}]
        `)
	l := defaultLog.WithFields(l.Fields{"collector": "diamond"})
	d := newDiamond(nil, 12, l).(*Diamond)
	_, ok := d.parseMetrics(rawData)
	assert.False(t, ok)
}

func connectToDiamondCollector(d *Diamond) (net.Conn, error) {
	// emit a Diamond metric
	var (
		conn net.Conn
		err  error
	)
	for retry := 0; retry < 3; retry++ {
		if conn, err = net.DialTimeout("tcp", "localhost:"+d.Port(), 2*time.Second); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return conn, err
}

func emitTestMetric(conn net.Conn) {
	var metrics []metric.Metric
	metrics = append(metrics, metric.New("test"))
	b, _ := json.Marshal(metrics)
	fmt.Fprintf(conn, string(b)+"\n")
	fmt.Fprintf(conn, string(b)+"\n")
}
