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

type StatusMetrics struct {
	PID             int             `json:"pid"`
	GoVersion       string          `json:"go_version"`
	NodeID          string          `json:"node_id"`
	Version         string          `json:"version"`
	Goroutines      int             `json:"goroutines"`
	Threads         int             `json:"threads"`
	MemStats        GoMemStats      `json:"mem_stats"`
	GCDurationStats GCDurationStats `json:"gc_duration_stats"`
	ProcStats       ProcStats       `json:"proc_stats"`
	CapsStats       *CapsStats      `json:"caps_stats"`
	GoMaxProcs      uint64          `json:"go_maxprocs"`
	GoGCPercent     uint64          `json:"go_gc_percent"`
	GoMemLimit      uint64          `json:"go_mem_limit"`
}

func (sm StatusMetrics) ToMap(debug bool, timezone string) (map[string]any, error) {
	if !debug {
		return sm.ToMapCondensed(timezone)
	}
	m := make(map[string]any)
	m["pid"] = sm.PID
	m["go_version"] = sm.GoVersion
	m["node_id"] = sm.NodeID
	m["version"] = sm.Version
	m["goroutines"] = sm.Goroutines
	m["threads"] = sm.Threads
	m["mem_stats"] = sm.MemStats.ToMap()
	m["gc_duration_stats"] = sm.GCDurationStats.ToMap()
	m["proc_stats"] = sm.ProcStats.ToMap()
	if sm.CapsStats != nil {
		m["caps_stats"] = sm.CapsStats.ToMap()
	}
	m["go_maxprocs"] = sm.GoMaxProcs
	m["go_gc_percent"] = sm.GoGCPercent
	m["go_mem_limit"] = sm.GoMemLimit
	return m, nil
}

func (sm StatusMetrics) ToMapCondensed(timezone string) (map[string]any, error) {
	m := make(map[string]any)
	m[utils.PID] = sm.PID
	m[utils.GoVersion] = sm.GoVersion
	m[utils.NodeID] = sm.NodeID
	m[utils.VersionLower] = sm.Version

	startTime, err := utils.ParseTimeDetectLayout(strconv.Itoa(int(sm.ProcStats.StartTime)), timezone)
	if err != nil {
		return nil, err
	}
	m[utils.RunningSince] = startTime.Format(time.UnixDate)

	m[utils.Goroutines] = sm.Goroutines
	m[utils.OpenFiles] = sm.ProcStats.OpenFDs
	m[utils.ResidentMemory] = utils.SizeFmt(float64(sm.ProcStats.ResidentMemory), "")
	m[utils.ActiveMemory] = utils.SizeFmt(float64(sm.MemStats.HeapAlloc), "")
	m[utils.SystemMemory] = utils.SizeFmt(float64(sm.MemStats.Sys), "")
	m[utils.OSThreadsInUse] = sm.Threads

	durStr := strconv.FormatFloat(sm.ProcStats.CPUTime, 'f', -1, 64)
	dur, err := utils.ParseDurationWithSecs(durStr)
	if err != nil {
		return nil, err
	}
	m[utils.CPUTime] = dur.String()

	if sm.CapsStats != nil {
		m[utils.CAPSAllocated] = sm.CapsStats.Allocated
		if sm.CapsStats.Peak != nil {
			m[utils.CAPSPeak] = *sm.CapsStats.Peak
		}
	}
	return m, nil
}

type GoMemStats struct {
	Alloc        uint64  `json:"alloc"`
	TotalAlloc   uint64  `json:"total_alloc"`
	Sys          uint64  `json:"sys"`
	Mallocs      uint64  `json:"mallocs"`
	Frees        uint64  `json:"frees"`
	HeapAlloc    uint64  `json:"heap_alloc"`
	HeapSys      uint64  `json:"heap_sys"`
	HeapIdle     uint64  `json:"heap_idle"`
	HeapInuse    uint64  `json:"heap_inuse"`
	HeapReleased uint64  `json:"heap_released"`
	HeapObjects  uint64  `json:"heap_objects"`
	StackInuse   uint64  `json:"stack_inuse"`
	StackSys     uint64  `json:"stack_sys"`
	MSpanSys     uint64  `json:"mspan_sys"`
	MSpanInuse   uint64  `json:"mspan_inuse"`
	MCacheInuse  uint64  `json:"mcache_inuse"`
	MCacheSys    uint64  `json:"mcache_sys"`
	BuckHashSys  uint64  `json:"buckhash_sys"`
	GCSys        uint64  `json:"gc_sys"`
	OtherSys     uint64  `json:"other_sys"`
	NextGC       uint64  `json:"next_gc"`
	LastGC       float64 `json:"last_gc"`
}

func (ms GoMemStats) ToMap() map[string]any {
	m := make(map[string]any, 23)
	m["alloc"] = ms.Alloc
	m["total_alloc"] = ms.TotalAlloc
	m["sys"] = ms.Sys
	m["mallocs"] = ms.Mallocs
	m["frees"] = ms.Frees
	m["heap_alloc"] = ms.HeapAlloc
	m["heap_sys"] = ms.HeapSys
	m["heap_idle"] = ms.HeapIdle
	m["heap_inuse"] = ms.HeapInuse
	m["heap_released"] = ms.HeapReleased
	m["heap_objects"] = ms.HeapObjects
	m["stack_inuse"] = ms.StackInuse
	m["stack_sys"] = ms.StackSys
	m["mspan_sys"] = ms.MSpanSys
	m["mspan_inuse"] = ms.MSpanInuse
	m["mcache_inuse"] = ms.MCacheInuse
	m["mcache_sys"] = ms.MCacheSys
	m["buckhash_sys"] = ms.BuckHashSys
	m["gc_sys"] = ms.GCSys
	m["other_sys"] = ms.OtherSys
	m["next_gc"] = ms.NextGC
	m["last_gc"] = ms.LastGC
	return m
}

type GCDurationStats struct {
	Quantiles []Quantile `json:"quantiles"`
	Sum       float64    `json:"sum"`
	Count     uint64     `json:"count"`
}

func (s GCDurationStats) ToMap() map[string]any {
	m := make(map[string]any, 3)
	m["quantiles"] = s.Quantiles
	m["sum"] = s.Sum
	m["count"] = s.Count
	return m
}

type Quantile struct {
	Quantile float64 `json:"quantile"`
	Value    float64 `json:"value"`
}

type ProcStats struct {
	CPUTime              float64 `json:"cpu_time"`
	MaxFDs               uint64  `json:"max_fds"`
	OpenFDs              int     `json:"open_fds"`
	ResidentMemory       int     `json:"resident_memory"`
	StartTime            float64 `json:"start_time"`
	VirtualMemory        uint    `json:"virtual_memory"`
	MaxVirtualMemory     uint64  `json:"max_virtual_memory"`
	NetworkReceiveTotal  float64 `json:"network_receive_total"`
	NetworkTransmitTotal float64 `json:"network_transmit_total"`
}

func (ps ProcStats) ToMap() map[string]any {
	m := make(map[string]any, 9)
	m["cpu_time"] = ps.CPUTime
	m["max_fds"] = ps.MaxFDs
	m["open_fds"] = ps.OpenFDs
	m["resident_memory"] = ps.ResidentMemory
	m["start_time"] = ps.StartTime
	m["virtual_memory"] = ps.VirtualMemory
	m["max_virtual_memory"] = ps.MaxVirtualMemory
	m["network_receive_total"] = ps.NetworkReceiveTotal
	m["network_transmit_total"] = ps.NetworkTransmitTotal
	return m
}

type CapsStats struct {
	Allocated int  `json:"allocated"`
	Peak      *int `json:"peak"`
}

func (cs *CapsStats) ToMap() map[string]any {
	m := make(map[string]any, 2)
	m["allocated"] = cs.Allocated
	m["peak"] = cs.Peak
	return m
}

func computeAppMetrics() (StatusMetrics, error) {
	vers, err := utils.GetCGRVersion()
	if err != nil {
		return StatusMetrics{}, err
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memStats := GoMemStats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		Mallocs:      m.Mallocs,
		Frees:        m.Frees,
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
		MSpanInuse:   m.MSpanInuse,
		MSpanSys:     m.MSpanSys,
		MCacheInuse:  m.MCacheInuse,
		MCacheSys:    m.MCacheSys,
		BuckHashSys:  m.BuckHashSys,
		GCSys:        m.GCSys,
		OtherSys:     m.OtherSys,
		NextGC:       m.NextGC,
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
		Count:     uint64(stats.NumGC),
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
		procStats.VirtualMemory = stat.VirtualMemory()
		procStats.ResidentMemory = stat.ResidentMemory()
		if startTime, err := stat.StartTime(); err == nil {
			procStats.StartTime = startTime
		} else {
			return StatusMetrics{}, err
		}
	} else {
		return StatusMetrics{}, err
	}
	if fds, err := p.FileDescriptorsLen(); err == nil {
		procStats.OpenFDs = fds
	} else {
		return StatusMetrics{}, err
	}

	if limits, err := p.Limits(); err == nil {
		procStats.MaxFDs = limits.OpenFiles
		procStats.MaxVirtualMemory = limits.AddressSpace
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
		"/sched/gomaxprocs:threads",
		"/gc/gogc:percent",
		"/memory/limit:bytes",
		"/gc/gomemlimit:bytes",
	}
	samples := make([]metrics.Sample, len(metricNames))
	for i, name := range metricNames {
		samples[i].Name = name
	}
	metrics.Read(samples)
	goMaxProcs := getUint64Metric(samples, "/sched/gomaxprocs:threads")
	goGCPercent := getUint64Metric(samples, "/gc/gogc:percent")
	goMemLimit := getUint64Metric(samples, "/gc/gomemlimit:bytes")

	return StatusMetrics{
		PID:             pid,
		GoVersion:       runtime.Version(),
		Version:         vers,
		Goroutines:      runtime.NumGoroutine(),
		Threads:         threads,
		MemStats:        memStats,
		GCDurationStats: gcDur,
		ProcStats:       procStats,
		GoMaxProcs:      goMaxProcs,
		GoGCPercent:     goGCPercent,
		GoMemLimit:      goMemLimit,
	}, nil
}

// getUint64Metric retrieves a uint64 metric by name
func getUint64Metric(samples []metrics.Sample, name string) uint64 {
	for _, sample := range samples {
		if sample.Name == name {
			switch sample.Value.Kind() {
			case metrics.KindUint64:
				return sample.Value.Uint64()
			case metrics.KindFloat64:
				return uint64(sample.Value.Float64())
			case metrics.KindBad:
				panic(fmt.Sprintf("metric %s has bad kind", name))
			default:
				panic(fmt.Sprintf("metric %s has unexpected unsupported kind: %v", name, sample.Value.Kind()))
			}
		}
	}
	return 0
}
