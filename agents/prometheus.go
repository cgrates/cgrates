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

// Prometheus metric names
const (
	// Gauge metrics
	promGoGoroutines                 = "go_goroutines"
	promGoThreads                    = "go_threads"
	promProcessOpenFds               = "process_open_fds"
	promProcessMaxFds                = "process_max_fds"
	promProcessResidentMemoryBytes   = "process_resident_memory_bytes"
	promProcessVirtualMemoryBytes    = "process_virtual_memory_bytes"
	promProcessVirtualMemoryMaxBytes = "process_virtual_memory_max_bytes"
	promProcessStartTimeSeconds      = "process_start_time_seconds"
	promGoMemstatsAllocBytes         = "go_memstats_alloc_bytes"
	promGoMemstatsHeapAllocBytes     = "go_memstats_heap_alloc_bytes"
	promGoMemstatsHeapIdleBytes      = "go_memstats_heap_idle_bytes"
	promGoMemstatsHeapInuseBytes     = "go_memstats_heap_inuse_bytes"
	promGoMemstatsHeapObjects        = "go_memstats_heap_objects"
	promGoMemstatsHeapReleasedBytes  = "go_memstats_heap_released_bytes"
	promGoMemstatsHeapSysBytes       = "go_memstats_heap_sys_bytes"
	promGoMemstatsBuckHashSysBytes   = "go_memstats_buck_hash_sys_bytes"
	promGoMemstatsGCSysBytes         = "go_memstats_gc_sys_bytes"
	promGoMemstatsMCacheInuseBytes   = "go_memstats_mcache_inuse_bytes"
	promGoMemstatsMCacheSysBytes     = "go_memstats_mcache_sys_bytes"
	promGoMemstatsMSpanInuseBytes    = "go_memstats_mspan_inuse_bytes"
	promGoMemstatsMSpanSysBytes      = "go_memstats_mspan_sys_bytes"
	promGoMemstatsNextGCBytes        = "go_memstats_next_gc_bytes"
	promGoMemstatsOtherSysBytes      = "go_memstats_other_sys_bytes"
	promGoMemstatsStackInuseBytes    = "go_memstats_stack_inuse_bytes"
	promGoMemstatsStackSysBytes      = "go_memstats_stack_sys_bytes"
	promGoMemstatsSysBytes           = "go_memstats_sys_bytes"
	promGoMemstatsLastGCTimeSeconds  = "go_memstats_last_gc_time_seconds"
	promGoGCGogcPercent              = "go_gc_gogc_percent"
	promGoGCGomemlimitBytes          = "go_gc_gomemlimit_bytes"
	promGoSchedGomaxprocsThreads     = "go_sched_gomaxprocs_threads"
	promGoInfo                       = "go_info"
	promGoGCDurationSeconds          = "go_gc_duration_seconds"

	// Counter metrics
	promProcessCPUSecondsTotal          = "process_cpu_seconds_total"
	promProcessNetworkReceiveByteTotal  = "process_network_receive_bytes_total"
	promProcessNetworkTransmitByteTotal = "process_network_transmit_bytes_total"
	promGoMemstatsAllocBytesTotal       = "go_memstats_alloc_bytes_total"
	promGoMemstatsMallocsTotal          = "go_memstats_mallocs_total"
	promGoMemstatsFreesTotal            = "go_memstats_frees_total"
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
		promGoGoroutines:                 "Number of goroutines that currently exist.",
		promGoThreads:                    "Number of OS threads created.",
		promProcessOpenFds:               "Number of open file descriptors.",
		promProcessMaxFds:                "Maximum number of open file descriptors.",
		promProcessResidentMemoryBytes:   "Resident memory size in bytes.",
		promProcessVirtualMemoryBytes:    "Virtual memory size in bytes.",
		promProcessVirtualMemoryMaxBytes: "Maximum amount of virtual memory available in bytes.",
		promProcessStartTimeSeconds:      "Start time of the process since unix epoch in seconds.",
		promGoMemstatsAllocBytes:         "Number of bytes allocated in heap and currently in use. Equals to /memory/classes/heap/objects:bytes.",
		promGoMemstatsHeapAllocBytes:     "Number of heap bytes allocated and currently in use, same as go_memstats_alloc_bytes. Equals to /memory/classes/heap/objects:bytes.",
		promGoMemstatsHeapIdleBytes:      "Number of heap bytes waiting to be used. Equals to /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.",
		promGoMemstatsHeapInuseBytes:     "Number of heap bytes that are in use. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes",
		promGoMemstatsHeapObjects:        "Number of currently allocated objects. Equals to /gc/heap/objects:objects.",
		promGoMemstatsHeapReleasedBytes:  "Number of heap bytes released to OS. Equals to /memory/classes/heap/released:bytes.",
		promGoMemstatsHeapSysBytes:       "Number of heap bytes obtained from system. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes + /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.",
		promGoMemstatsBuckHashSysBytes:   "Number of bytes used by the profiling bucket hash table. Equals to /memory/classes/profiling/buckets:bytes.",
		promGoMemstatsGCSysBytes:         "Number of bytes used for garbage collection system metadata. Equals to /memory/classes/metadata/other:bytes.",
		promGoMemstatsMCacheInuseBytes:   "Number of bytes in use by mcache structures. Equals to /memory/classes/metadata/mcache/inuse:bytes.",
		promGoMemstatsMCacheSysBytes:     "Number of bytes used for mcache structures obtained from system. Equals to /memory/classes/metadata/mcache/inuse:bytes + /memory/classes/metadata/mcache/free:bytes.",
		promGoMemstatsMSpanInuseBytes:    "Number of bytes in use by mspan structures. Equals to /memory/classes/metadata/mspan/inuse:bytes.",
		promGoMemstatsMSpanSysBytes:      "Number of bytes used for mspan structures obtained from system. Equals to /memory/classes/metadata/mspan/inuse:bytes + /memory/classes/metadata/mspan/free:bytes.",
		promGoMemstatsNextGCBytes:        "Number of heap bytes when next garbage collection will take place. Equals to /gc/heap/goal:bytes.",
		promGoMemstatsOtherSysBytes:      "Number of bytes used for other system allocations. Equals to /memory/classes/other:bytes.",
		promGoMemstatsStackInuseBytes:    "Number of bytes obtained from system for stack allocator in non-CGO environments. Equals to /memory/classes/heap/stacks:bytes.",
		promGoMemstatsStackSysBytes:      "Number of bytes obtained from system for stack allocator. Equals to /memory/classes/heap/stacks:bytes + /memory/classes/os-stacks:bytes.",
		promGoMemstatsSysBytes:           "Number of bytes obtained from system. Equals to /memory/classes/total:byte.",
		promGoMemstatsLastGCTimeSeconds:  "Number of seconds since 1970 of last garbage collection.",
		promGoGCGogcPercent:              "Heap size target percentage configured by the user, otherwise 100. This value is set by the GOGC environment variable, and the runtime/debug.SetGCPercent function. Sourced from /gc/gogc:percent.",
		promGoGCGomemlimitBytes:          "Go runtime memory limit configured by the user, otherwise math.MaxInt64. This value is set by the GOMEMLIMIT environment variable, and the runtime/debug.SetMemoryLimit function. Sourced from /gc/gomemlimit:bytes.",
		promGoSchedGomaxprocsThreads:     "The current runtime.GOMAXPROCS setting, or the number of operating system threads that can execute user-level Go code simultaneously. Sourced from /sched/gomaxprocs:threads.",
	}
	for name, help := range gaugeMetrics {
		c.descs[name] = prometheus.NewDesc(name, help, []string{"node_id"}, nil)
	}

	counterMetrics := map[string]string{
		promProcessCPUSecondsTotal:          "Total user and system CPU time spent in seconds.",
		promProcessNetworkReceiveByteTotal:  "Number of bytes received by the process over the network.",
		promProcessNetworkTransmitByteTotal: "Number of bytes sent by the process over the network.",
		promGoMemstatsAllocBytesTotal:       "Total number of bytes allocated in heap until now, even if released already. Equals to /gc/heap/allocs:bytes.",
		promGoMemstatsMallocsTotal:          "Total number of heap objects allocated, both live and gc-ed. Semantically a counter version for go_memstats_heap_objects gauge. Equals to /gc/heap/allocs:objects + /gc/heap/tiny/allocs:objects.",
		promGoMemstatsFreesTotal:            "Total number of heap objects frees. Equals to /gc/heap/frees:objects + /gc/heap/tiny/allocs:objects.",
	}
	for name, help := range counterMetrics {
		c.descs[name] = prometheus.NewDesc(name, help, []string{"node_id"}, nil)
	}

	c.descs[promGoInfo] = prometheus.NewDesc(
		promGoInfo,
		"Information about the Go environment.",
		[]string{"node_id", "version"},
		nil)

	c.descs[promGoGCDurationSeconds] = prometheus.NewDesc(
		promGoGCDurationSeconds,
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
		nodeID, ok := reply[utils.NodeID].(string)
		if !ok {
			panic("missing node_id in CoreSv1.Status reply")
		}

		if val, ok := reply[utils.MetricRuntimeGoroutines].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs[promGoGoroutines], prometheus.GaugeValue, val, nodeID)
		}
		if val, ok := reply[utils.MetricRuntimeThreads].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs[promGoThreads], prometheus.GaugeValue, val, nodeID)
		}
		if version, ok := reply[utils.GoVersion].(string); ok {
			ch <- prometheus.MustNewConstMetric(c.descs[promGoInfo], prometheus.GaugeValue, 1, nodeID, version)
		}

		if procStats, ok := reply[utils.FieldProcStats].(map[string]any); ok {
			if val, ok := procStats[utils.MetricProcCPUTime].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promProcessCPUSecondsTotal], prometheus.CounterValue, val, nodeID)
			}
			if val, ok := procStats[utils.MetricProcOpenFDs].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promProcessOpenFds], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats[utils.MetricProcMaxFDs].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promProcessMaxFds], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats[utils.MetricProcResidentMemory].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promProcessResidentMemoryBytes], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats[utils.MetricProcVirtualMemory].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promProcessVirtualMemoryBytes], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats[utils.MetricProcMaxVirtualMemory].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promProcessVirtualMemoryMaxBytes], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats[utils.MetricProcStartTime].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promProcessStartTimeSeconds], prometheus.GaugeValue, val, nodeID)
			}
			if val, ok := procStats[utils.MetricProcNetworkReceiveTotal].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promProcessNetworkReceiveByteTotal], prometheus.CounterValue, val, nodeID)
			}
			if val, ok := procStats[utils.MetricProcNetworkTransmitTotal].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promProcessNetworkTransmitByteTotal], prometheus.CounterValue, val, nodeID)
			}
		}

		if memStats, ok := reply[utils.FieldMemStats].(map[string]any); ok {
			var memGaugeMap = map[string]string{
				promGoMemstatsAllocBytes:        utils.MetricMemAlloc,
				promGoMemstatsHeapAllocBytes:    utils.MetricMemHeapAlloc,
				promGoMemstatsHeapIdleBytes:     utils.MetricMemHeapIdle,
				promGoMemstatsHeapInuseBytes:    utils.MetricMemHeapInuse,
				promGoMemstatsHeapObjects:       utils.MetricMemHeapObjects,
				promGoMemstatsHeapReleasedBytes: utils.MetricMemHeapReleased,
				promGoMemstatsHeapSysBytes:      utils.MetricMemHeapSys,
				promGoMemstatsBuckHashSysBytes:  utils.MetricMemBuckHashSys,
				promGoMemstatsGCSysBytes:        utils.MetricMemGCSys,
				promGoMemstatsMCacheInuseBytes:  utils.MetricMemMCacheInuse,
				promGoMemstatsMCacheSysBytes:    utils.MetricMemMCacheSys,
				promGoMemstatsMSpanInuseBytes:   utils.MetricMemMSpanInuse,
				promGoMemstatsMSpanSysBytes:     utils.MetricMemMSpanSys,
				promGoMemstatsNextGCBytes:       utils.MetricMemNextGC,
				promGoMemstatsOtherSysBytes:     utils.MetricMemOtherSys,
				promGoMemstatsStackInuseBytes:   utils.MetricMemStackInuse,
				promGoMemstatsStackSysBytes:     utils.MetricMemStackSys,
				promGoMemstatsSysBytes:          utils.MetricMemSys,
			}
			for metricName, key := range memGaugeMap {
				if val, ok := memStats[key].(float64); ok {
					ch <- prometheus.MustNewConstMetric(c.descs[metricName], prometheus.GaugeValue, val, nodeID)
				}
			}
			var memCounterMap = map[string]string{
				promGoMemstatsAllocBytesTotal: utils.MetricMemTotalAlloc,
				promGoMemstatsMallocsTotal:    utils.MetricMemMallocs,
				promGoMemstatsFreesTotal:      utils.MetricMemFrees,
			}
			for metricName, key := range memCounterMap {
				if val, ok := memStats[key].(float64); ok {
					ch <- prometheus.MustNewConstMetric(c.descs[metricName], prometheus.CounterValue, val, nodeID)
				}
			}
			if val, ok := memStats[utils.MetricMemLastGC].(float64); ok {
				ch <- prometheus.MustNewConstMetric(c.descs[promGoMemstatsLastGCTimeSeconds], prometheus.GaugeValue, val, nodeID)
			}
		}

		if gcStats, ok := reply[utils.FieldGCDurationStats].(map[string]any); ok {
			var count uint64
			var sum float64
			quantileValues := make(map[float64]float64)
			if c, ok := gcStats[utils.MetricGCCount].(float64); ok {
				count = uint64(c)
			}
			if s, ok := gcStats[utils.MetricGCSum].(float64); ok {
				sum = s
			}

			// Handle different types that may be returned based on connection type:
			// - []any: from serialized RPC connections where type information is lost
			// - []cores.Quantile: from direct (*internal) calls where type is preserved
			switch quantiles := gcStats[utils.MetricGCQuantiles].(type) {
			case []any:
				for _, q := range quantiles {
					if qMap, ok := q.(map[string]any); ok {
						if quantile, ok := qMap[utils.MetricGCQuantile].(float64); ok {
							if val, ok := qMap[utils.MetricGCValue].(float64); ok {
								quantileValues[quantile] = val
							}
						}
					}
				}
			case []cores.Quantile:
				for _, q := range quantiles {
					quantileValues[q.Quantile] = q.Value
				}
			}

			ch <- prometheus.MustNewConstSummary(
				c.descs[promGoGCDurationSeconds],
				count,
				sum,
				quantileValues,
				nodeID,
			)
		}

		if val, ok := reply[utils.MetricRuntimeMaxProcs].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs[promGoSchedGomaxprocsThreads], prometheus.GaugeValue, val, nodeID)
		}
		if val, ok := reply[utils.MetricGCPercent].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs[promGoGCGogcPercent], prometheus.GaugeValue, val, nodeID)
		}
		if val, ok := reply[utils.MetricMemLimit].(float64); ok {
			ch <- prometheus.MustNewConstMetric(c.descs[promGoGCGomemlimitBytes], prometheus.GaugeValue, val, nodeID)
		}
	}
}
