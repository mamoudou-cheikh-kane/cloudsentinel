// Package metrics provides Prometheus collectors for system-level metrics.
package metrics

import (
	"log/slog"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemCollector gathers system-level metrics (CPU, RAM, disk).
// It implements prometheus.Collector so it can be registered with a Registry.
type SystemCollector struct {
	nodeName string

	cpuUsage    *prometheus.Desc
	memTotal    *prometheus.Desc
	memUsed     *prometheus.Desc
	memFree     *prometheus.Desc
	memPercent  *prometheus.Desc
	diskTotal   *prometheus.Desc
	diskUsed    *prometheus.Desc
	diskPercent *prometheus.Desc
	goroutines  *prometheus.Desc
}

// NewSystemCollector creates a new SystemCollector for the given node.
func NewSystemCollector(nodeName string) *SystemCollector {
	labels := []string{"node"}
	return &SystemCollector{
		nodeName: nodeName,
		cpuUsage: prometheus.NewDesc(
			"cloudsentinel_cpu_usage_percent",
			"Overall CPU usage percentage (0-100).",
			labels, nil,
		),
		memTotal: prometheus.NewDesc(
			"cloudsentinel_memory_total_bytes",
			"Total memory in bytes.",
			labels, nil,
		),
		memUsed: prometheus.NewDesc(
			"cloudsentinel_memory_used_bytes",
			"Used memory in bytes.",
			labels, nil,
		),
		memFree: prometheus.NewDesc(
			"cloudsentinel_memory_free_bytes",
			"Free memory in bytes.",
			labels, nil,
		),
		memPercent: prometheus.NewDesc(
			"cloudsentinel_memory_usage_percent",
			"Memory usage percentage (0-100).",
			labels, nil,
		),
		diskTotal: prometheus.NewDesc(
			"cloudsentinel_disk_total_bytes",
			"Total disk space in bytes for root filesystem.",
			labels, nil,
		),
		diskUsed: prometheus.NewDesc(
			"cloudsentinel_disk_used_bytes",
			"Used disk space in bytes for root filesystem.",
			labels, nil,
		),
		diskPercent: prometheus.NewDesc(
			"cloudsentinel_disk_usage_percent",
			"Disk usage percentage for root filesystem (0-100).",
			labels, nil,
		),
		goroutines: prometheus.NewDesc(
			"cloudsentinel_agent_goroutines",
			"Number of goroutines currently running in the agent.",
			labels, nil,
		),
	}
}

// Describe sends the descriptors of each metric to the provided channel.
func (c *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuUsage
	ch <- c.memTotal
	ch <- c.memUsed
	ch <- c.memFree
	ch <- c.memPercent
	ch <- c.diskTotal
	ch <- c.diskUsed
	ch <- c.diskPercent
	ch <- c.goroutines
}

// Collect is called by the Prometheus registry on each scrape.
// It gathers fresh values for all metrics.
func (c *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	// CPU usage (averaged over 100ms).
	if cpuPercents, err := cpu.Percent(0, false); err == nil && len(cpuPercents) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.cpuUsage, prometheus.GaugeValue, cpuPercents[0], c.nodeName,
		)
	} else if err != nil {
		slog.Warn("cpu collection failed", "error", err)
	}

	// Memory stats.
	if vm, err := mem.VirtualMemory(); err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.memTotal, prometheus.GaugeValue, float64(vm.Total), c.nodeName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.memUsed, prometheus.GaugeValue, float64(vm.Used), c.nodeName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.memFree, prometheus.GaugeValue, float64(vm.Free), c.nodeName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.memPercent, prometheus.GaugeValue, vm.UsedPercent, c.nodeName,
		)
	} else {
		slog.Warn("memory collection failed", "error", err)
	}

	// Disk stats for the root filesystem.
	if du, err := disk.Usage("/"); err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.diskTotal, prometheus.GaugeValue, float64(du.Total), c.nodeName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.diskUsed, prometheus.GaugeValue, float64(du.Used), c.nodeName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.diskPercent, prometheus.GaugeValue, du.UsedPercent, c.nodeName,
		)
	} else {
		slog.Warn("disk collection failed", "error", err)
	}

	// Number of goroutines in the agent itself.
	ch <- prometheus.MustNewConstMetric(
		c.goroutines, prometheus.GaugeValue, float64(runtime.NumGoroutine()), c.nodeName,
	)
}
