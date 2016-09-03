package config

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strconv"

	"github.com/Sirupsen/logrus"
)

var log = logrus.WithFields(logrus.Fields{"app": "qcollect", "pkg": "config"})

// Config type holds the global qcollect configuration.
type Config struct {
	Prefix                string                            `json:"prefix"`
	Interval              interface{}                       `json:"interval"`
	CollectorsConfigPath  string                            `json:"collectorsConfigPath"`
	DiamondCollectorsPath string                            `json:"diamondCollectorsPath"`
	DiamondCollectors     []string                          `json:"diamondCollectors"`
	Handlers              map[string]map[string]interface{} `json:"handlers"`
	Collectors            []string                          `json:"collectors"`
	DefaultDimensions     map[string]string                 `json:"defaultDimensions"`
	InternalServerConfig  map[string]interface{}            `json:"internalServer"`
}

// ReadConfig reads a qcollect configuration file
func ReadConfig(configFile string) (c Config, e error) {
	log.Info("Reading configuration file at ", configFile)
	contents, e := ioutil.ReadFile(configFile)
	if e != nil {
		log.Error("Config file error: ", e)
		return c, e
	}
	err := json.Unmarshal(contents, &c)
	if err != nil {
		log.Error("Invalid JSON in config: ", err)
		return c, err
	}
	return c, nil
}

// ReadCollectorConfig reads a qcollect collector configuration file
func ReadCollectorConfig(configFile string) (c map[string]interface{}, e error) {
	log.Info("Reading collector configuration file at ", configFile)
	contents, e := ioutil.ReadFile(configFile)
	if e != nil {
		log.Error("Config file error: ", e)
		return c, e
	}
	err := json.Unmarshal(contents, &c)
	if err != nil {
		log.Error("Invalid JSON in config: ", err)
		return c, err
	}
	return c, nil
}

// GetAsFloat parses a string to a float or returns the float if float is passed in
func GetAsFloat(value interface{}, defaultValue float64) (result float64) {
	result = defaultValue

	switch value.(type) {
	case string:
		fromString, err := strconv.ParseFloat(value.(string), 64)
		if err != nil {
			log.Warn("Failed to convert value", value, "to a float64. Falling back to default", defaultValue)
			result = defaultValue
		} else {
			result = fromString
		}
	case float64:
		result = value.(float64)
	}

	return
}

// GetAsInt parses a string/float to an int or returns the int if int is passed in
func GetAsInt(value interface{}, defaultValue int) (result int) {
	result = defaultValue

	switch value.(type) {
	case string:
		fromString, err := strconv.ParseInt(value.(string), 10, 64)
		if err == nil {
			result = int(fromString)
		} else {
			log.Warn("Failed to convert value", value, "to an int")
		}
	case int:
		result = value.(int)
	case int32:
		result = int(value.(int32))
	case int64:
		result = int(value.(int64))
	case float64:
		result = int(value.(float64))
	}

	return
}

// GetAsMap parses a string to a map[string]string
func GetAsMap(value interface{}) (result map[string]string) {
	result = make(map[string]string)

	switch value.(type) {
	case string:
		err := json.Unmarshal([]byte(value.(string)), &result)
		if err != nil {
			log.Warn("Failed to convert value", value, "to a map")
		}
	case map[string]interface{}:
		temp := value.(map[string]interface{})
		for k, v := range temp {
			if str, ok := v.(string); ok {
				result[k] = str
			} else {
				log.Warn("Expected a string but got", reflect.TypeOf(value), ". Discarding handler level metric: ", k)
			}
		}
	case map[string]string:
		result = value.(map[string]string)
	default:
		log.Warn("Expected a string but got", reflect.TypeOf(value), ". Returning empty map!")
	}

	return
}

// GetAsSlice : Parses a json array string to []string
func GetAsSlice(value interface{}) []string {
	result := []string{}

	switch realValue := value.(type) {
	case string:
		err := json.Unmarshal([]byte(realValue), &result)
		if err != nil {
			log.Warn("Failed to convert string:", realValue, "to a []string")
		}
	case []string:
		result = realValue
	case []interface{}:
		result = make([]string, len(realValue))
		for i, value := range realValue {
			result[i] = value.(string)
		}
	default:
		log.Warn("Expected a string array but got", reflect.TypeOf(realValue), ". Returning empty slice!")
	}

	return result
}
