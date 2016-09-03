package util

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/qnib/qcollect/config"
)

// For dependency injection
var ipGetter = getIps

// example configuration::
//
// {
//     	"heartbeat_path": "/var/run/nerve/heartbeat",
//		"instance_id": "srv1-devc",
//		"services": {
//	 		"<SERVICE_NAME>.<otherstuff>": {
//				"host": "<IPADDR>",
//      	    "port": ###,
//      	}
// 		}
//     "services": {
//
// Most imporantly is the port, host and service name. The service name is assumed to be formatted like this::
//
type nerveConfigData struct {
	Services map[string]map[string]interface{}
}

// ParseNerveConfig is responsible for taking the JSON string coming in into a map of service:port
// it will also filter based on only services runnign on this host.
// To deal with multi-tenancy we actually will return port:service
func ParseNerveConfig(raw *[]byte) (map[int]string, error) {
	results := make(map[int]string)
	ips, err := ipGetter()

	if err != nil {
		return results, err
	}
	parsed := new(nerveConfigData)

	// convert the ips into a map for membership tests
	ipMap := make(map[string]bool)
	for _, val := range ips {
		ipMap[val] = true
	}

	err = json.Unmarshal(*raw, parsed)
	if err != nil {
		return results, err
	}

	for rawServiceName, serviceConfig := range parsed.Services {
		host := strings.TrimSpace(serviceConfig["host"].(string))

		_, exists := ipMap[host]
		if exists {
			name := strings.Split(rawServiceName, ".")[0]
			port := config.GetAsInt(serviceConfig["port"], -1)
			if port != -1 {
				results[port] = name
			}
		}
	}

	return results, nil
}

// getIps is responsible for getting all the ips that are associated with this NIC
func getIps() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{}, err
	}

	results := []string{}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return []string{}, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			results = append(results, ip.String())
		}
	}

	return results, nil
}
