package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	l "github.com/Sirupsen/logrus"
	"github.com/qnib/qcollect/config"
	"github.com/qnib/qcollect/metric"
)

const (
	endpoint = "unix:///var/run/docker.sock"
)

// DockerStats collector type.
// previousCPUValues contains the last cpu-usage values per container.
// dockerClient is the client for the Docker remote API.
type DockerStats struct {
	baseCollector
	dockerClient  *client.Client
	statsTimeout  int
	compiledRegex map[string]*Regex
	skipRegex     *regexp.Regexp
	bufferRegex   *regexp.Regexp
	endpoint      string
	perCore       bool
	cpuThrottle   bool
	blkIO         bool
	mu            *sync.Mutex
}

// CntStat is encapsulate a fetch container stats for a gochannel
type CntStat struct {
	Container types.Container
	Stats     types.StatsJSON
	Ok        bool
}

// CPUValues struct contains the last cpu-usage values in order to compute properly the current values.
// (see calculateCPUPercent() for more details)
type CPUValues struct {
	totCPU, systemCPU uint64
}

// Regex struct contains the info used to get the user specific dimensions from the docker env variables
// tag: is the environmental variable you want to get the value from
// regex: is the reg exp used to extract the value from the env var
type Regex struct {
	tag   string
	regex *regexp.Regexp
}

func init() {
	RegisterCollector("DockerStats", newDockerStats)
}

// newDockerStats creates a new DockerStats collector.
func newDockerStats(channel chan metric.Metric, initialInterval int, log *l.Entry) Collector {
	d := new(DockerStats)

	d.log = log
	d.channel = channel
	d.interval = initialInterval
	d.mu = new(sync.Mutex)

	d.name = "DockerStats"
	d.perCore = false
	d.cpuThrottle = false
	d.blkIO = false
	d.compiledRegex = make(map[string]*Regex)
	return d
}

// GetEndpoint Returns endpoint of DockerStats instance
func (d *DockerStats) GetEndpoint() string {
	return d.endpoint
}

// Configure takes a dictionary of values with which the handler can configure itself.
func (d *DockerStats) Configure(configMap map[string]interface{}) {
	if timeout, exists := configMap["dockerStatsTimeout"]; exists {
		d.statsTimeout = min(config.GetAsInt(timeout, d.interval), d.interval)
	} else {
		d.statsTimeout = d.interval
	}
	if dockerEndpoint, exists := configMap["dockerEndPoint"]; exists {
		if str, ok := dockerEndpoint.(string); ok {
			d.endpoint = str
		} else {
			d.log.Warn("Failed to cast dokerEndPoint: ", reflect.TypeOf(dockerEndpoint))
		}
	} else {
		d.endpoint = endpoint
	}
	cli, err := client.NewEnvClient()
	if err != nil {
		d.log.Warn("Could not create docker client...", err)
	} else {
		d.dockerClient = cli
	}
	if generatedDimensions, exists := configMap["generatedDimensions"]; exists {
		for dimension, generator := range generatedDimensions.(map[string]interface{}) {
			for key, regx := range config.GetAsMap(generator) {
				re, err := regexp.Compile(regx)
				if err != nil {
					d.log.Warn("Failed to compile regex: ", regx, err)
				} else {
					d.compiledRegex[dimension] = &Regex{regex: re, tag: key}
				}
			}
		}
	}
	if skipRegex, skipExists := configMap["skipContainerRegex"]; skipExists {
		d.skipRegex = regexp.MustCompile(skipRegex.(string))
	}
	if perCore, pcExists := configMap["per-core"]; pcExists {
		if perCore == "true" {
			d.perCore = true
		}
	}
	if cpuThrottle, ctExists := configMap["cpu-throttle"]; ctExists {
		if cpuThrottle == "true" {
			d.cpuThrottle = true
		}
	}
	if blkIO, blkExists := configMap["block-io"]; blkExists {
		if blkIO == "true" {
			d.blkIO = true
		}
	}
	if bufferRegex, exists := configMap["bufferRegex"]; exists {
		d.bufferRegex = regexp.MustCompile(bufferRegex.(string))
	}
	d.configureCommonParams(configMap)
}

// Collect iterates on all the docker containers alive and, if possible, collects the correspondent
// memory and cpu statistics.
// For each container a gorutine is started to spin up the collection process.
func (d *DockerStats) Collect() {
	if d.dockerClient == nil {
		d.log.Error("Invalid endpoint: ", d.endpoint)
		return
	}
	cfilter := filters.NewArgs()
	//cfilter.Add("id", task.ContainerID)
	containers, err := d.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{Filter: cfilter})
	if err != nil {
		d.log.Error("ListContainers() failed: ", err)
		return
	}
	for _, container := range containers {
		contName := strings.TrimPrefix(container.Names[0], "/")
		if err != nil {
			d.log.Error("InspectContainer() failed: ", err)
			continue
		}

		if d.skipRegex != nil && d.skipRegex.MatchString(contName) {
			d.log.Debug("Skip container: ", contName)
			continue
		}
		go d.getDockerContainerInfo(container)
	}
}

// getDockerContainerInfo gets container statistics for the given container.
// results is a channel to make possible the synchronization between the main process and the gorutines (wait-notify pattern).
func (d *DockerStats) getDockerContainerInfo(container types.Container) {
	statsC := make(chan CntStat, 1)
	done := make(chan bool, 1)

	go func() {
		statsC <- d.getContainerStats(container)
	}()
	select {
	case stats := <-statsC:
		if !stats.Ok {
			d.log.Error("Failed to collect docker container stats: ", container.Names)
			break
		} else {

		}
		done <- true

		metrics := d.extractMetrics(container, stats.Stats)
		d.sendMetrics(metrics)

		break
	case <-time.After(time.Duration(d.statsTimeout) * time.Second):
		d.log.Error("Timed out collecting stats for container ", container.ID)
		done <- true
		break
	}
}

func (d *DockerStats) getContainerStats(container types.Container) CntStat {
	d.log.Debug("Inspect: ", container.ID)
	ok := true
	stat, err := d.dockerClient.ContainerStats(context.Background(), container.ID, false)
	content, err := ioutil.ReadAll(stat.Body)
	if err != nil {
		ok = false
	}
	var s types.StatsJSON
	err = json.Unmarshal(content, &s)
	if err != nil {
		ok = false
	}
	s.CPUStats = d.diffCPUUsage(s.PreCPUStats, s.CPUStats)
	return CntStat{
		Container: container,
		Stats:     s,
		Ok:        ok,
	}
}

func (d *DockerStats) extractMetrics(container types.Container, stats types.StatsJSON) []metric.Metric {
	d.mu.Lock()
	defer d.mu.Unlock()
	metrics := d.buildMetrics(container, stats)
	return metrics
}

func (d DockerStats) diffCPUUsage(pre types.CPUStats, cur types.CPUStats) types.CPUStats {
	var cstat types.CPUStats
	cstat.SystemUsage = cur.SystemUsage - pre.SystemUsage
	cstat.ThrottlingData = types.ThrottlingData{
		// Number of periods with throttling active
		Periods: cur.ThrottlingData.Periods - pre.ThrottlingData.Periods,
		// Number of periods when the container hits its throttling limit.
		ThrottledPeriods: cur.ThrottlingData.ThrottledPeriods - pre.ThrottlingData.ThrottledPeriods,
		// Aggregate time the container was throttled for in nanoseconds.
		ThrottledTime: cur.ThrottlingData.ThrottledTime - pre.ThrottlingData.ThrottledTime,
	}
	cstat.CPUUsage.TotalUsage = cur.CPUUsage.TotalUsage - pre.CPUUsage.TotalUsage
	cstat.CPUUsage.UsageInKernelmode = cur.CPUUsage.UsageInKernelmode - pre.CPUUsage.UsageInKernelmode
	cstat.CPUUsage.UsageInUsermode = cur.CPUUsage.UsageInUsermode - pre.CPUUsage.UsageInUsermode
	pCPU := cur.CPUUsage.PercpuUsage
	for idx, c := range pre.CPUUsage.PercpuUsage {
		pCPU[idx] = (pCPU[idx] - c) / cstat.SystemUsage
	}
	cstat.CPUUsage.PercpuUsage = pCPU
	return cstat
}

// buildMetrics creates the actual metrics for the given container.
func (d DockerStats) buildMetrics(container types.Container, stat types.StatsJSON) []metric.Metric {
	mTime := stat.Read
	d.log.Debug("Build Metrics for: ", container.Names[0])
	ret := []metric.Metric{
		d.buildDockerMetric("cpu.system", metric.Gauge, float64(stat.CPUStats.SystemUsage/10000000), mTime),
		d.buildDockerMetric("cpu.usage", metric.Gauge, float64(stat.CPUStats.CPUUsage.TotalUsage/10000000), mTime),
		d.buildDockerMetric("memory.usage", metric.Gauge, float64(stat.MemoryStats.Usage), mTime),
		d.buildDockerMetric("memory.limit", metric.Gauge, float64(stat.MemoryStats.Limit), mTime),
		d.buildDockerMetric("pid.current", metric.Gauge, float64(stat.PidsStats.Current), mTime),
		d.buildDockerMetric("pid.limit", metric.Gauge, float64(stat.PidsStats.Limit), mTime),
	}

	if d.blkIO {
		d.log.Debug("Building BlkIO metrics...")
		metName := "IoMerged"
		for _, bs := range stat.BlkioStats.IoMergedRecursive {
			ret = append(ret, d.buildDockerMetric(fmt.Sprintf("blkio.%s.%s.value", metName, bs.Op), metric.Gauge, float64(bs.Value), mTime))
		}
		metName = "IoQueued"
		for _, bs := range stat.BlkioStats.IoQueuedRecursive {
			ret = append(ret, d.buildDockerMetric(fmt.Sprintf("blkio.%s.%s.value", metName, bs.Op), metric.Gauge, float64(bs.Value), mTime))
		}
		metName = "IoServiceBytes"
		for _, bs := range stat.BlkioStats.IoServiceBytesRecursive {
			ret = append(ret, d.buildDockerMetric(fmt.Sprintf("blkio.%s.%s.value", metName, bs.Op), metric.Gauge, float64(bs.Value), mTime))
		}
		metName = "IoServiceTime"
		for _, bs := range stat.BlkioStats.IoServiceTimeRecursive {
			ret = append(ret, d.buildDockerMetric(fmt.Sprintf("blkio.%s.%s.value", metName, bs.Op), metric.Gauge, float64(bs.Value), mTime))
		}
		metName = "IoServiced"
		for _, bs := range stat.BlkioStats.IoServicedRecursive {
			ret = append(ret, d.buildDockerMetric(fmt.Sprintf("blkio.%s.%s.value", metName, bs.Op), metric.Gauge, float64(bs.Value), mTime))
		}
		metName = "IoTime"
		for _, bs := range stat.BlkioStats.IoTimeRecursive {
			ret = append(ret, d.buildDockerMetric(fmt.Sprintf("blkio.%s.%s.value", metName, bs.Op), metric.Gauge, float64(bs.Value), mTime))
		}
		metName = "IoWaitTime"
		for _, bs := range stat.BlkioStats.IoWaitTimeRecursive {
			ret = append(ret, d.buildDockerMetric(fmt.Sprintf("blkio.%s.%s.value", metName, bs.Op), metric.Gauge, float64(bs.Value), mTime))
		}
		metName = "Sectors"
		for _, bs := range stat.BlkioStats.SectorsRecursive {
			ret = append(ret, d.buildDockerMetric(fmt.Sprintf("blkio.%s.%s.value", metName, bs.Op), metric.Gauge, float64(bs.Value), mTime))
		}
	}

	if d.cpuThrottle {
		d.log.Debug("Building cpuThrottle metrics...")
		addCPU := []metric.Metric{
			d.buildDockerMetric("cpu.throttling.Periods", metric.Gauge, float64(stat.CPUStats.ThrottlingData.Periods), mTime),
			d.buildDockerMetric("cpu.throttling.ThrottledPeriods", metric.Gauge, float64(stat.CPUStats.ThrottlingData.ThrottledPeriods), mTime),
			d.buildDockerMetric("cpu.throttling.ThrottledTime", metric.Gauge, float64(stat.CPUStats.ThrottlingData.ThrottledTime/10000000), mTime),
		}
		ret = append(ret, addCPU...)
	}

	if d.perCore {
		d.log.Debug("Building perCore metrics...")
		for idx, c := range stat.CPUStats.CPUUsage.PercpuUsage {
			ret = append(ret, d.buildDockerMetric(fmt.Sprintf("cpu.ns.core%d", idx), metric.Gauge, float64(c/10000000), mTime))
		}
	}

	for netiface := range stat.Networks {
		// legacy format
		txb := d.buildDockerMetric("TxBytes", metric.CumulativeCounter, float64(stat.Networks[netiface].TxBytes), mTime)
		txb.AddDimension("iface", netiface)
		ret = append(ret, txb)
		rxb := d.buildDockerMetric("RxBytes", metric.CumulativeCounter, float64(stat.Networks[netiface].RxBytes), mTime)
		rxb.AddDimension("iface", netiface)
		ret = append(ret, rxb)
	}
	additionalDimensions := map[string]string{
		"container_id":   container.ID,
		"container_name": strings.TrimPrefix(container.Names[0], "/"),
	}
	for k, v := range container.Labels {
		additionalDimensions[k] = v
	}
	metric.AddToAll(&ret, additionalDimensions)
	return ret
}

// sendMetrics writes all the metrics received to the collector channel.
func (d DockerStats) sendMetrics(metrics []metric.Metric) {
	for _, m := range metrics {
		d.Channel() <- m
	}
}

func (d DockerStats) buildDockerMetric(name string, metricType string, value float64, mTime time.Time) (m metric.Metric) {
	m = metric.New(name)
	m.MetricType = metricType
	m.Value = value
	m.SetTime(mTime)
	if d.bufferRegex != nil && d.bufferRegex.MatchString(name) {
		m.Buffered = true
	}
	return m
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
