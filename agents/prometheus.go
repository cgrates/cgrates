/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package agents

import (
	"fmt"
	"net/http"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusAgent handles metrics collection for Prometheus.
// It collects stats from StatQueues and exposes them alongside
// optional Go runtime and process metrics.
type PrometheusAgent struct {
	cfg *config.CGRConfig
	cm  *engine.ConnManager

	handler     http.Handler
	statMetrics *prometheus.GaugeVec
}

// NewPrometheusAgent creates and initializes a PrometheusAgent with
// pre-registered metrics based on the provided configuration.
func NewPrometheusAgent(cfg *config.CGRConfig, cm *engine.ConnManager) *PrometheusAgent {
	reg := prometheus.NewRegistry()

	if len(cfg.PrometheusAgentCfg().CoreSConns) != 0 {
		coreMetricsCollector := newCoreMetricsCollector(cfg, cm)
		reg.MustRegister(coreMetricsCollector)
	}

	statMetrics := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "cgrates",
			Subsystem: "stats",
			Name:      "metrics",
			Help:      "Current values for StatQueue metrics",
		}, []string{"tenant", "queue", "metric"})
	reg.MustRegister(statMetrics)
	if cfg.PrometheusAgentCfg().CollectGoMetrics {
		reg.MustRegister(collectors.NewGoCollector())
	}
	if cfg.PrometheusAgentCfg().CollectProcessMetrics {
		reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	handler := promhttp.InstrumentMetricHandler(
		reg,
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
	)

	return &PrometheusAgent{
		cfg:         cfg,
		cm:          cm,
		handler:     handler,
		statMetrics: statMetrics,
	}
}

// ServeHTTP implements http.Handler interface. It updates all metrics on each
// scrape request before exposing them via the Prometheus HTTP handler.
func (pa *PrometheusAgent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pa.updateStatsMetrics()
	pa.handler.ServeHTTP(w, r)
}

// updateStatsMetrics fetches and updates all StatQueue metrics by calling each
// configured StatS connection.
func (pa *PrometheusAgent) updateStatsMetrics() {
	if len(pa.cfg.PrometheusAgentCfg().StatQueueIDs) == 0 {
		return
	}
	for _, connID := range pa.cfg.PrometheusAgentCfg().StatSConns {
		for _, sqID := range pa.cfg.PrometheusAgentCfg().StatQueueIDs {

			tenantID := utils.NewTenantID(sqID)
			if tenantID.Tenant == "" {
				tenantID.Tenant = pa.cfg.GeneralCfg().DefaultTenant
			}

			var metrics map[string]float64
			err := pa.cm.Call(context.Background(), []string{connID},
				utils.StatSv1GetQueueFloatMetrics,
				&utils.TenantIDWithAPIOpts{
					TenantID: tenantID,
				}, &metrics)
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Err(fmt.Sprintf(
					"<%s> failed to retrieve metrics for StatQueue %q (connID=%q): %v",
					utils.PrometheusAgent, sqID, connID, err))
				continue
			}

			for metricID, val := range metrics {
				pa.statMetrics.WithLabelValues(tenantID.Tenant, tenantID.ID, metricID).Set(val)
			}
		}
	}
}

// coreMetricsCollector collects CoreS metrics. Equivalent to Go/Process collectors.
type coreMetricsCollector struct {
	cfg *config.CGRConfig
	cm  *engine.ConnManager

	// Pre-defined descriptors
	descs map[string]*prometheus.Desc
}

// newCoreMetricsCollector creates a new collector with pre-defined descriptors
func newCoreMetricsCollector(cfg *config.CGRConfig, cm *engine.ConnManager) *coreMetricsCollector {
	c := &coreMetricsCollector{
		cfg:   cfg,
		cm:    cm,
		descs: make(map[string]*prometheus.Desc),
	}

	gaugeMetrics := map[string]string{
		"go_goroutines":                    "Number of goroutines that currently exist.",
		"go_threads":                       "Number of OS threads created.",
		"process_open_fds":                 "Number of open file descriptors.",
		"process_max_fds":                  "Maximum number of open file descriptors.",
		"process_resident_memory_bytes":    "Resident memory size in bytes.",
		"process_virtual_memory_bytes":     "Virtual memory size in bytes.",
		"process_virtual_memory_max_bytes": "Maximum amount of virtual memory available in bytes.",
		"process_start_time_seconds":       "Start time of the process since unix epoch in seconds.",
		"go_memstats_alloc_bytes":          "Number of bytes allocated in heap and currently in use. Equals to /memory/classes/heap/objects:bytes.",
		"go_memstats_heap_alloc_bytes":     "Number of heap bytes allocated and currently in use, same as go_memstats_alloc_bytes. Equals to /memory/classes/heap/objects:bytes.",
		"go_memstats_heap_idle_bytes":      "Number of heap bytes waiting to be used. Equals to /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.",
		"go_memstats_heap_inuse_bytes":     "Number of heap bytes that are in use. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes",
		"go_memstats_heap_objects":         "Number of currently allocated objects. Equals to /gc/heap/objects:objects.",
		"go_memstats_heap_released_bytes":  "Number of heap bytes released to OS. Equals to /memory/classes/heap/released:bytes.",
		"go_memstats_heap_sys_bytes":       "Number of heap bytes obtained from system. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes + /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.",
		"go_memstats_buck_hash_sys_bytes":  "Number of bytes used by the profiling bucket hash table. Equals to /memory/classes/profiling/buckets:bytes.",
		"go_memstats_gc_sys_bytes":         "Number of bytes used for garbage collection system metadata. Equals to /memory/classes/metadata/other:bytes.",
		"go_memstats_mcache_inuse_bytes":   "Number of bytes in use by mcache structures. Equals to /memory/classes/metadata/mcache/inuse:bytes.",
		"go_memstats_mcache_sys_bytes":     "Number of bytes used for mcache structures obtained from system. Equals to /memory/classes/metadata/mcache/inuse:bytes + /memory/classes/metadata/mcache/free:bytes.",
		"go_memstats_mspan_inuse_bytes":    "Number of bytes in use by mspan structures. Equals to /memory/classes/metadata/mspan/inuse:bytes.",
		"go_memstats_mspan_sys_bytes":      "Number of bytes used for mspan structures obtained from system. Equals to /memory/classes/metadata/mspan/inuse:bytes + /memory/classes/metadata/mspan/free:bytes.",
		"go_memstats_next_gc_bytes":        "Number of heap bytes when next garbage collection will take place. Equals to /gc/heap/goal:bytes.",
		"go_memstats_other_sys_bytes":      "Number of bytes used for other system allocations. Equals to /memory/classes/other:bytes.",
		"go_memstats_stack_inuse_bytes":    "Number of bytes obtained from system for stack allocator in non-CGO environments. Equals to /memory/classes/heap/stacks:bytes.",
		"go_memstats_stack_sys_bytes":      "Number of bytes obtained from system for stack allocator. Equals to /memory/classes/heap/stacks:bytes + /memory/classes/os-stacks:bytes.",
		"go_memstats_sys_bytes":            "Number of bytes obtained from system. Equals to /memory/classes/total:byte.",
		"go_memstats_last_gc_time_seconds": "Number of seconds since 1970 of last garbage collection.",
		"go_gc_gogc_percent":               "Heap size target percentage configured by the user, otherwise 100. This value is set by the GOGC environment variable, and the runtime/debug.SetGCPercent function. Sourced from /gc/gogc:percent.",
		"go_gc_gomemlimit_bytes":           "Go runtime memory limit configured by the user, otherwise math.MaxInt64. This value is set by the GOMEMLIMIT environment variable, and the runtime/debug.SetMemoryLimit function. Sourced from /gc/gomemlimit:bytes.",
		"go_sched_gomaxprocs_threads":      "The current runtime.GOMAXPROCS setting, or the number of operating system threads that can execute user-level Go code simultaneously. Sourced from /sched/gomaxprocs:threads.",
	}
	for name, help := range gaugeMetrics {
		c.descs[name] = prometheus.NewDesc(name, help, []string{"node_id"}, nil)
	}

	counterMetrics := map[string]string{
		"process_cpu_seconds_total":            "Total user and system CPU time spent in seconds.",
		"process_network_receive_bytes_total":  "Number of bytes received by the process over the network.",
		"process_network_transmit_bytes_total": "Number of bytes sent by the process over the network.",
		"go_memstats_alloc_bytes_total":        "Total number of bytes allocated in heap until now, even if released already. Equals to /gc/heap/allocs:bytes.",
		"go_memstats_mallocs_total":            "Total number of heap objects allocated, both live and gc-ed. Semantically a counter version for go_memstats_heap_objects gauge. Equals to /gc/heap/allocs:objects + /gc/heap/tiny/allocs:objects.",
		"go_memstats_frees_total":              "Total number of heap objects frees. Equals to /gc/heap/frees:objects + /gc/heap/tiny/allocs:objects.",
	}
	for name, help := range counterMetrics {
		c.descs[name] = prometheus.NewDesc(name, help, []string{"node_id"}, nil)
	}

	c.descs["go_info"] = prometheus.NewDesc(
		"go_info",
		"Information about the Go environment.",
		[]string{"node_id", "version"},
		nil)

	c.descs["go_gc_duration_seconds"] = prometheus.NewDesc(
		"go_gc_duration_seconds",
		"A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.",
		[]string{"node_id"},
		nil,
	)
	return c
}

// Describe implements prometheus.Collector.
func (c *coreMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		ch <- desc
	}
}

// Collect implements prometheus.Collector.
func (c *coreMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, connID := range c.cfg.PrometheusAgentCfg().CoreSConns {
		var reply map[string]any
		if err := c.cm.Call(context.Background(), []string{connID},
			utils.CoreSv1Status,
			&cores.V1StatusParams{
				Debug: true,
			},
			&reply); err != nil {
			utils.Logger.Err(fmt.Sprintf(
				"<%s> failed to retrieve metrics (connID=%q): %v",
				utils.PrometheusAgent, connID, err))
			continue
		}
		nodeID, ok := reply["node_id"].(string)
		if !ok {
			panic("missing node_id in CoreSv1.Status reply")
		}

		if val, ok := reply["goroutines"].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs["go_goroutines"], prometheus.GaugeValue, val, nodeID)
		}
		if val, ok := reply["threads"].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs["go_threads"], prometheus.GaugeValue, val, nodeID)
		}
		if version, ok := reply["go_version"].(string); ok {
			ch <- prometheus.MustNewConstMetric(c.descs["go_info"], prometheus.GaugeValue, 1, nodeID, version)
		}

		if procStats, ok := reply["proc_stats"].(map[string]any); ok {
			if val, ok := procStats["cpu_time"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["process_cpu_seconds_total"], prometheus.CounterValue, val, nodeID)
			}
			if val, ok := procStats["open_fds"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["process_open_fds"], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats["max_fds"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["process_max_fds"], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats["resident_memory"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["process_resident_memory_bytes"], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats["virtual_memory"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["process_virtual_memory_bytes"], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats["max_virtual_memory"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["process_virtual_memory_max_bytes"], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats["start_time"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["process_start_time_seconds"], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats["network_receive_total"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["process_network_receive_bytes_total"], prometheus.CounterValue, val, nodeID)
			}
			if val, ok := procStats["network_transmit_total"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["process_network_transmit_bytes_total"], prometheus.CounterValue, val, nodeID)
			}
		}

		if memStats, ok := reply["mem_stats"].(map[string]any); ok {
			memGaugeMap := map[string]string{
				"go_memstats_alloc_bytes":         "alloc",
				"go_memstats_heap_alloc_bytes":    "heap_alloc",
				"go_memstats_heap_idle_bytes":     "heap_idle",
				"go_memstats_heap_inuse_bytes":    "heap_inuse",
				"go_memstats_heap_objects":        "heap_objects",
				"go_memstats_heap_released_bytes": "heap_released",
				"go_memstats_heap_sys_bytes":      "heap_sys",
				"go_memstats_buck_hash_sys_bytes": "buckhash_sys",
				"go_memstats_gc_sys_bytes":        "gc_sys",
				"go_memstats_mcache_inuse_bytes":  "mcache_inuse",
				"go_memstats_mcache_sys_bytes":    "mcache_sys",
				"go_memstats_mspan_inuse_bytes":   "mspan_inuse",
				"go_memstats_mspan_sys_bytes":     "mspan_sys",
				"go_memstats_next_gc_bytes":       "next_gc",
				"go_memstats_other_sys_bytes":     "other_sys",
				"go_memstats_stack_inuse_bytes":   "stack_inuse",
				"go_memstats_stack_sys_bytes":     "stack_sys",
				"go_memstats_sys_bytes":           "sys",
			}
			for metricName, key := range memGaugeMap {
				if val, ok := memStats[key].(float64); ok {
					ch <- prometheus.MustNewConstMetric(c.descs[metricName], prometheus.GaugeValue, val, nodeID)
				}
			}
			memCounterMap := map[string]string{
				"go_memstats_alloc_bytes_total": "total_alloc",
				"go_memstats_mallocs_total":     "mallocs",
				"go_memstats_frees_total":       "frees",
			}
			for metricName, key := range memCounterMap {
				if val, ok := memStats[key].(float64); ok {
					ch <- prometheus.MustNewConstMetric(c.descs[metricName], prometheus.CounterValue, val, nodeID)
				}
			}
			if val, ok := memStats["last_gc"].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs["go_memstats_last_gc_time_seconds"], prometheus.GaugeValue, val, nodeID)
			}
		}

		if gcStats, ok := reply["gc_duration_stats"].(map[string]any); ok {
			var count uint64
			var sum float64
			quantileValues := make(map[float64]float64)
			if c, ok := gcStats["count"].(float64); ok {
				count = uint64(c)
			}
			if s, ok := gcStats["sum"].(float64); ok {
				sum = s
			}
			if quantiles, ok := gcStats["quantiles"].([]any); ok {
				for _, q := range quantiles {
					if qMap, ok := q.(map[string]any); ok {
						if quantile, ok := qMap["quantile"].(float64); ok {
							if val, ok := qMap["value"].(float64); ok {
								quantileValues[quantile] = val
							}
						}
					}
				}
			}
			ch <- prometheus.MustNewConstSummary(
				c.descs["go_gc_duration_seconds"],
				count,
				sum,
				quantileValues,
				nodeID,
			)
		}

		if val, ok := reply["go_maxprocs"].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs["go_sched_gomaxprocs_threads"], prometheus.GaugeValue, val, nodeID)
		}
		if val, ok := reply["go_gc_percent"].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs["go_gc_gogc_percent"], prometheus.GaugeValue, val, nodeID)
		}
		if val, ok := reply["go_mem_limit"].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs["go_gc_gomemlimit_bytes"], prometheus.GaugeValue, val, nodeID)
		}
	}
}
