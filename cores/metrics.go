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

package cores

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/prometheus/procfs"
)

// Runtime metrics paths
const (
	metricPathGoMaxProcs  = "/sched/gomaxprocs:threads"
	metricPathGoGCPercent = "/gc/gogc:percent"
	metricPathGoMemLimit  = "/gc/gomemlimit:bytes"
)

// StatusMetrics contains runtime metrics, including process information,
// memory usage, garbage collection stats, and other operational metrics. It's
// used for monitoring and diagnostics via APIs and Prometheus.
//
// NOTE: All numeric values use float64 for consistent representation across
// different connection types (direct *internal calls vs serialized RPC
// connections) while aligning with Prometheus' float64-based metric system.
type StatusMetrics struct {
	PID        float64 `json:"pid"`
	GoVersion  string  `json:"go_version"`
	NodeID     string  `json:"node_id"`
	Version    string  `json:"version"`
	Goroutines float64 `json:"goroutines"`
	Threads    float64 `json:"threads"`

	MemStats        GoMemStats      `json:"mem_stats"`
	GCDurationStats GCDurationStats `json:"gc_duration_stats"`
	ProcStats       ProcStats       `json:"proc_stats"`
	CapsStats       *CapsStats      `json:"caps_stats"`

	MaxProcs  float64 `json:"maxprocs"`
	GCPercent float64 `json:"gc_percent"`
	MemLimit  float64 `json:"mem_limit"`
}

// toMap converts the StatusMetrics to a map[string]any with all fields.
// When debug is false, it calls toMapCondensed to return a simplified view.
func (sm StatusMetrics) toMap(debug bool, timezone string) (map[string]any, error) {
	if !debug {
		return sm.toMapCondensed(timezone)
	}
	m := map[string]any{
		utils.PID:                     sm.PID,
		utils.GoVersion:               sm.GoVersion,
		utils.NodeID:                  sm.NodeID,
		utils.FieldVersion:            sm.Version,
		utils.MetricRuntimeGoroutines: sm.Goroutines,
		utils.MetricRuntimeThreads:    sm.Threads,
		utils.FieldMemStats:           sm.MemStats.toMap(),
		utils.FieldGCDurationStats:    sm.GCDurationStats.toMap(),
		utils.FieldProcStats:          sm.ProcStats.toMap(),
		utils.MetricRuntimeMaxProcs:   sm.MaxProcs,
		utils.MetricGCPercent:         sm.GCPercent,
		utils.MetricMemLimit:          sm.MemLimit,
		utils.FieldCapsStats:          sm.CapsStats.toMap(),
	}
	return m, nil
}

// toMapCondensed provides a simplified map view of StatusMetrics with formatted
// human-readable values.
func (sm StatusMetrics) toMapCondensed(timezone string) (map[string]any, error) {
	m := map[string]any{
		utils.PID:                      sm.PID,
		utils.GoVersion:                sm.GoVersion,
		utils.NodeID:                   sm.NodeID,
		utils.FieldVersion:             sm.Version,
		utils.MetricRuntimeGoroutines:  sm.Goroutines,
		utils.OpenFiles:                sm.ProcStats.OpenFDs,
		utils.MetricProcResidentMemory: utils.SizeFmt(sm.ProcStats.ResidentMemory, ""),
		utils.ActiveMemory:             utils.SizeFmt(sm.MemStats.HeapAlloc, ""),
		utils.SystemMemory:             utils.SizeFmt(sm.MemStats.Sys, ""),
		utils.OSThreadsInUse:           sm.Threads,
	}

	startTime, err := utils.ParseTimeDetectLayout(strconv.Itoa(int(sm.ProcStats.StartTime)), timezone)
	if err != nil {
		return nil, err
	}
	m[utils.RunningSince] = startTime.Format(time.UnixDate)

	durStr := strconv.FormatFloat(sm.ProcStats.CPUTime, 'f', -1, 64)
	dur, err := utils.ParseDurationWithSecs(durStr)
	if err != nil {
		return nil, err
	}
	m[utils.MetricProcCPUTime] = dur.String()

	if sm.CapsStats != nil {
		m[utils.MetricCapsAllocated] = sm.CapsStats.Allocated
		if sm.CapsStats.Peak != nil {
			m[utils.MetricCapsPeak] = *sm.CapsStats.Peak
		}
	}
	return m, nil
}

type GoMemStats struct {
	Alloc        float64 `json:"alloc"`
	TotalAlloc   float64 `json:"total_alloc"`
	Sys          float64 `json:"sys"`
	Mallocs      float64 `json:"mallocs"`
	Frees        float64 `json:"frees"`
	HeapAlloc    float64 `json:"heap_alloc"`
	HeapSys      float64 `json:"heap_sys"`
	HeapIdle     float64 `json:"heap_idle"`
	HeapInuse    float64 `json:"heap_inuse"`
	HeapReleased float64 `json:"heap_released"`
	HeapObjects  float64 `json:"heap_objects"`
	StackInuse   float64 `json:"stack_inuse"`
	StackSys     float64 `json:"stack_sys"`
	MSpanSys     float64 `json:"mspan_sys"`
	MSpanInuse   float64 `json:"mspan_inuse"`
	MCacheInuse  float64 `json:"mcache_inuse"`
	MCacheSys    float64 `json:"mcache_sys"`
	BuckHashSys  float64 `json:"buckhash_sys"`
	GCSys        float64 `json:"gc_sys"`
	OtherSys     float64 `json:"other_sys"`
	NextGC       float64 `json:"next_gc"`
	LastGC       float64 `json:"last_gc"`
}

func (ms GoMemStats) toMap() map[string]any {
	return map[string]any{
		utils.MetricMemAlloc:        ms.Alloc,
		utils.MetricMemTotalAlloc:   ms.TotalAlloc,
		utils.MetricMemSys:          ms.Sys,
		utils.MetricMemMallocs:      ms.Mallocs,
		utils.MetricMemFrees:        ms.Frees,
		utils.MetricMemHeapAlloc:    ms.HeapAlloc,
		utils.MetricMemHeapSys:      ms.HeapSys,
		utils.MetricMemHeapIdle:     ms.HeapIdle,
		utils.MetricMemHeapInuse:    ms.HeapInuse,
		utils.MetricMemHeapReleased: ms.HeapReleased,
		utils.MetricMemHeapObjects:  ms.HeapObjects,
		utils.MetricMemStackInuse:   ms.StackInuse,
		utils.MetricMemStackSys:     ms.StackSys,
		utils.MetricMemMSpanSys:     ms.MSpanSys,
		utils.MetricMemMSpanInuse:   ms.MSpanInuse,
		utils.MetricMemMCacheInuse:  ms.MCacheInuse,
		utils.MetricMemMCacheSys:    ms.MCacheSys,
		utils.MetricMemBuckHashSys:  ms.BuckHashSys,
		utils.MetricMemGCSys:        ms.GCSys,
		utils.MetricMemOtherSys:     ms.OtherSys,
		utils.MetricMemNextGC:       ms.NextGC,
		utils.MetricMemLastGC:       ms.LastGC,
	}
}

type GCDurationStats struct {
	Quantiles []Quantile `json:"quantiles"`
	Sum       float64    `json:"sum"`
	Count     float64    `json:"count"`
}

func (s GCDurationStats) toMap() map[string]any {
	return map[string]any{
		utils.MetricGCQuantiles: s.Quantiles,
		utils.MetricGCSum:       s.Sum,
		utils.MetricGCCount:     s.Count,
	}
}

type Quantile struct {
	Quantile float64 `json:"quantile"`
	Value    float64 `json:"value"`
}

type ProcStats struct {
	CPUTime              float64 `json:"cpu_time"`
	MaxFDs               float64 `json:"max_fds"`
	OpenFDs              float64 `json:"open_fds"`
	ResidentMemory       float64 `json:"resident_memory"`
	StartTime            float64 `json:"start_time"`
	VirtualMemory        float64 `json:"virtual_memory"`
	MaxVirtualMemory     float64 `json:"max_virtual_memory"`
	NetworkReceiveTotal  float64 `json:"network_receive_total"`
	NetworkTransmitTotal float64 `json:"network_transmit_total"`
}

func (ps ProcStats) toMap() map[string]any {
	return map[string]any{
		utils.MetricProcCPUTime:              ps.CPUTime,
		utils.MetricProcMaxFDs:               ps.MaxFDs,
		utils.MetricProcOpenFDs:              ps.OpenFDs,
		utils.MetricProcResidentMemory:       ps.ResidentMemory,
		utils.MetricProcStartTime:            ps.StartTime,
		utils.MetricProcVirtualMemory:        ps.VirtualMemory,
		utils.MetricProcMaxVirtualMemory:     ps.MaxVirtualMemory,
		utils.MetricProcNetworkReceiveTotal:  ps.NetworkReceiveTotal,
		utils.MetricProcNetworkTransmitTotal: ps.NetworkTransmitTotal,
	}
}

type CapsStats struct {
	Allocated int  `json:"allocated"`
	Peak      *int `json:"peak"`
}

func (cs *CapsStats) toMap() map[string]any {
	if cs == nil {
		return nil
	}
	return map[string]any{
		utils.MetricCapsAllocated: cs.Allocated,
		utils.MetricCapsPeak:      cs.Peak,
	}
}

// computeAppMetrics gathers runtime metrics including memory usage, goroutines,
// GC stats, and process information for monitoring and diagnostics.
func computeAppMetrics() (StatusMetrics, error) {
	vers, err := utils.GetCGRVersion()
	if err != nil {
		return StatusMetrics{}, err
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memStats := GoMemStats{
		Alloc:        float64(m.Alloc),
		TotalAlloc:   float64(m.TotalAlloc),
		Sys:          float64(m.Sys),
		Mallocs:      float64(m.Mallocs),
		Frees:        float64(m.Frees),
		HeapAlloc:    float64(m.HeapAlloc),
		HeapSys:      float64(m.HeapSys),
		HeapIdle:     float64(m.HeapIdle),
		HeapInuse:    float64(m.HeapInuse),
		HeapReleased: float64(m.HeapReleased),
		HeapObjects:  float64(m.HeapObjects),
		StackInuse:   float64(m.StackInuse),
		StackSys:     float64(m.StackSys),
		MSpanInuse:   float64(m.MSpanInuse),
		MSpanSys:     float64(m.MSpanSys),
		MCacheInuse:  float64(m.MCacheInuse),
		MCacheSys:    float64(m.MCacheSys),
		BuckHashSys:  float64(m.BuckHashSys),
		GCSys:        float64(m.GCSys),
		OtherSys:     float64(m.OtherSys),
		NextGC:       float64(m.NextGC),
	}

	threads, _ := runtime.ThreadCreateProfile(nil)

	var stats debug.GCStats
	stats.PauseQuantiles = make([]time.Duration, 5)
	debug.ReadGCStats(&stats)
	quantiles := make([]Quantile, 0, 5)

	// Add the first quantile separately
	quantiles = append(quantiles, Quantile{
		Quantile: 0.0,
		Value:    stats.PauseQuantiles[0].Seconds(),
	})

	for idx, pq := range stats.PauseQuantiles[1:] {
		q := Quantile{
			Quantile: float64(idx+1) / float64(len(stats.PauseQuantiles)-1),
			Value:    pq.Seconds(),
		}
		quantiles = append(quantiles, q)
	}
	gcDur := GCDurationStats{
		Quantiles: quantiles,
		Count:     float64(stats.NumGC),
		Sum:       stats.PauseTotal.Seconds(),
	}
	memStats.LastGC = float64(stats.LastGC.UnixNano()) / 1e9

	// Process metrics
	pid := os.Getpid()
	p, err := procfs.NewProc(pid)
	if err != nil {
		return StatusMetrics{}, err
	}

	procStats := ProcStats{}
	if stat, err := p.Stat(); err == nil {
		procStats.CPUTime = stat.CPUTime()
		procStats.VirtualMemory = float64(stat.VirtualMemory())
		procStats.ResidentMemory = float64(stat.ResidentMemory())
		if startTime, err := stat.StartTime(); err == nil {
			procStats.StartTime = startTime
		} else {
			return StatusMetrics{}, err
		}
	} else {
		return StatusMetrics{}, err
	}
	if fds, err := p.FileDescriptorsLen(); err == nil {
		procStats.OpenFDs = float64(fds)
	} else {
		return StatusMetrics{}, err
	}

	if limits, err := p.Limits(); err == nil {
		procStats.MaxFDs = float64(limits.OpenFiles)
		procStats.MaxVirtualMemory = float64(limits.AddressSpace)
	} else {
		return StatusMetrics{}, err
	}

	if netstat, err := p.Netstat(); err == nil {
		var inOctets, outOctets float64
		if netstat.IpExt.InOctets != nil {
			inOctets = *netstat.IpExt.InOctets
		}
		if netstat.IpExt.OutOctets != nil {
			outOctets = *netstat.IpExt.OutOctets
		}
		procStats.NetworkReceiveTotal = inOctets
		procStats.NetworkTransmitTotal = outOctets
	} else {
		return StatusMetrics{}, err
	}

	metricNames := []string{
		metricPathGoMaxProcs,
		metricPathGoGCPercent,
		metricPathGoMemLimit,
	}
	samples := make([]metrics.Sample, len(metricNames))
	for i, name := range metricNames {
		samples[i].Name = name
	}
	metrics.Read(samples)
	goMaxProcs := getFloat64Metric(samples, metricPathGoMaxProcs)
	goGCPercent := getFloat64Metric(samples, metricPathGoGCPercent)
	goMemLimit := getFloat64Metric(samples, metricPathGoMemLimit)

	return StatusMetrics{
		PID:             float64(pid),
		GoVersion:       runtime.Version(),
		Version:         vers,
		Goroutines:      float64(runtime.NumGoroutine()),
		Threads:         float64(threads),
		MemStats:        memStats,
		GCDurationStats: gcDur,
		ProcStats:       procStats,
		MaxProcs:        float64(goMaxProcs),
		GCPercent:       float64(goGCPercent),
		MemLimit:        float64(goMemLimit),
	}, nil
}

// getFloat64Metric retrieves a float64 metric by name.
func getFloat64Metric(samples []metrics.Sample, name string) float64 {
	for _, sample := range samples {
		if sample.Name == name {
			switch sample.Value.Kind() {
			case metrics.KindUint64:
				return float64(sample.Value.Uint64())
			case metrics.KindFloat64:
				return sample.Value.Float64()
			case metrics.KindBad:
				panic(fmt.Sprintf("metric %s has bad kind", name))
			default:
				panic(fmt.Sprintf("metric %s has unexpected unsupported kind: %v", name, sample.Value.Kind()))
			}
		}
	}
	return 0
}
