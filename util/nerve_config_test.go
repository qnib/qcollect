package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestNerveConfig() []byte {
	raw := `
	{
	    "heartbeat_path": "/var/run/nerve/heartbeat",
	    "instance_id": "srv1-devc",
	    "services": {
	        "example_service.main.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "open_timeout": 6,
	                    "port": 6666,
	                    "rise": 1,
	                    "timeout": 6,
	                    "type": "http",
	                    "uri": "/http/example_service.main/13752/status"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 13752,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.5.5:22181",
	                "10.40.5.6:22181",
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_service.main"
	        },
	        "example_service.mesosstage_main.norcal-devc.superregion:norcal-devc.13752.new": {
	            "check_interval": 7,
	            "checks": [
	                {
	                    "fall": 2,
	                    "headers": {},
	                    "host": "127.0.0.1",
	                    "open_timeout": 6,
	                    "port": 6666,
	                    "rise": 1,
	                    "timeout": 6,
	                    "type": "http",
	                    "uri": "/http/example_service.mesosstage_main/13752/status"
	                }
	            ],
	            "host": "10.56.5.21",
	            "port": 22222,
	            "weight": 24,
	            "zk_hosts": [
	                "10.40.5.5:22181",
	                "10.40.5.6:22181",
	                "10.40.1.17:22181"
	            ],
	            "zk_path": "/nerve/superregion:norcal-devc/example_service.mesosstage_main"
	        }
	    }
	}
	`
	return []byte(raw)
}

func TestNerveConfigParsing(t *testing.T) {
	expected := map[int]string{
		22222: "example_service",
		13752: "example_service",
	}

	cfgString := getTestNerveConfig()
	ipGetter = func() ([]string, error) { return []string{"10.56.5.21"}, nil }
	results, err := ParseNerveConfig(&cfgString)
	assert.Nil(t, err)
	assert.Equal(t, expected, results)
}

func TestNerveFilterOnIP(t *testing.T) {
	cfgString := getTestNerveConfig()
	ipGetter = func() ([]string, error) { return []string{"10.56.2.3"}, nil }
	results, err := ParseNerveConfig(&cfgString)
	assert.Nil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}

func TestHandleBadNerveConfig(t *testing.T) {
	// b/c there is valid json coming in it won't error, just have an empty response
	cfgString := []byte("{}")
	ipGetter = func() ([]string, error) { return []string{"10.56.2.3"}, nil }
	results, err := ParseNerveConfig(&cfgString)
	assert.Nil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}

func TestHandlePoorlyFormedJson(t *testing.T) {
	cfgString := []byte("notjson")
	ipGetter = func() ([]string, error) { return []string{"10.56.2.3"}, nil }
	results, err := ParseNerveConfig(&cfgString)
	assert.NotNil(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}
