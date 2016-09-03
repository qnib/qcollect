package config_test

import (
	"github.com/qnib/qcollect/config"

	"io/ioutil"
	"os"
	"testing"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testBadConfiguration = `{
    "prefix": "test.",
    malformed JSON File {123!!!!
}
`

var testGoodConfiguration = `{
    "prefix": "test.",
    "interval": 10,
    "defaultDimensions": {
        "application": "qcollect",
        "host": "dev33-devc"
    },

    "collectorsConfigPath": "/tmp",
    "diamondCollectorsPath": "src/diamond/collectors",
    "diamondCollectors": ["CPUCollector","PingCollector"],
    "collectors": ["Test"],

    "handlers": {
        "Graphite": {
            "server": "10.40.11.51",
            "port": "2003",
            "timeout": 2
        },
        "SignalFx": {
            "authToken": "secret_token",
            "endpoint": "https://ingest.signalfx.com/v2/datapoint",
            "interval": 10,
            "timeout": 2,
			"collectorBlackList": ["TestCollector1", "TestCollector2"]
        }
    }
}
`

var testCollectorConfiguration = `{
	"metricName": "TestMetric",
	"interval": 10
}
`

var (
	tmpTestGoodFile, tmpTestBadFile, tempTestCollectorConfig string
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.ErrorLevel)
	if f, err := ioutil.TempFile("/tmp", "qcollect"); err == nil {
		f.WriteString(testGoodConfiguration)
		tmpTestGoodFile = f.Name()
		f.Close()
		defer os.Remove(tmpTestGoodFile)
	}
	if f, err := ioutil.TempFile("/tmp", "qcollect"); err == nil {
		f.WriteString(testBadConfiguration)
		tmpTestBadFile = f.Name()
		f.Close()
		defer os.Remove(tmpTestBadFile)
	}
	if f, err := ioutil.TempFile("/tmp", "qcollect"); err == nil {
		f.WriteString(testCollectorConfiguration)
		tempTestCollectorConfig = f.Name()
		f.Close()
		defer os.Remove(tempTestCollectorConfig)
	}
	os.Exit(m.Run())
}

func TestGetInt(t *testing.T) {
	assert := assert.New(t)

	val := config.GetAsInt("10", 123)
	assert.Equal(val, 10)

	val = config.GetAsInt("notanint", 123)
	assert.Equal(val, 123)

	val = config.GetAsInt(12.123, 123)
	assert.Equal(val, 12)

	val = config.GetAsInt(12, 123)
	assert.Equal(val, 12)
}

func TestGetFloat(t *testing.T) {
	assert := assert.New(t)

	val := config.GetAsFloat("10", 123)
	assert.Equal(val, 10.0)

	val = config.GetAsFloat("10.21", 123)
	assert.Equal(val, 10.21)

	val = config.GetAsFloat("notanint", 123)
	assert.Equal(val, 123.0)

	val = config.GetAsFloat(12.123, 123)
	assert.Equal(val, 12.123)
}

func TestGetAsMap(t *testing.T) {
	assert := assert.New(t)

	// Test if string can be converted to map[string]string
	stringToParse := "{\"runtimeenv\" : \"dev\", \"region\":\"uswest1-devc\"}"
	expectedValue := map[string]string{
		"runtimeenv": "dev",
		"region":     "uswest1-devc",
	}
	assert.Equal(config.GetAsMap(stringToParse), expectedValue)

	// Test if map[string]interface{} can be converted to map[string]string
	interfaceMapToParse := make(map[string]interface{})
	interfaceMapToParse["runtimeenv"] = "dev"
	interfaceMapToParse["region"] = "uswest1-devc"
	assert.Equal(config.GetAsMap(interfaceMapToParse), expectedValue)
}

func TestGetAsSlice(t *testing.T) {
	assert := assert.New(t)

	// Test if string array can be converted to []string
	stringToParse := "[\"TestCollector1\", \"TestCollector2\"]"
	expectedValue := []string{"TestCollector1", "TestCollector2"}
	assert.Equal(config.GetAsSlice(stringToParse), expectedValue)

	sliceToParse := []string{"TestCollector1", "TestCollector2"}
	assert.Equal(config.GetAsSlice(sliceToParse), expectedValue)
}

func TestGetAsSliceFromJson(t *testing.T) {
	var data interface{}
	jsonString := []byte(`{"listOfStrings": ["a", "b", "c"]}`)

	err := json.Unmarshal(jsonString, &data)
	assert.Nil(t, err)

	if err == nil {
		temp := data.(map[string]interface{})

		res := config.GetAsSlice(temp["listOfStrings"])
		assert.Equal(t, []string{"a", "b", "c"}, res)
	}
}

func TestParseCollectorConfig(t *testing.T) {
	_, err := config.ReadCollectorConfig(tempTestCollectorConfig)
	assert.Nil(t, err, "should succeed")
}

func TestParseGoodConfig(t *testing.T) {
	_, err := config.ReadConfig(tmpTestGoodFile)
	assert.Nil(t, err, "should succeed")
}

func TestParseBadConfig(t *testing.T) {
	_, err := config.ReadConfig(tmpTestBadFile)
	assert.NotNil(t, err, "should fail")
}
