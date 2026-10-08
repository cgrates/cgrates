// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package cores

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewCoreService(t *testing.T) {
	cfgDflt := config.NewDefaultCGRConfig()
	cfgDflt.CoreSCfg().CapsStatsInterval = time.Second
	stopchan := make(chan struct{}, 1)
	defer close(stopchan)
	caps := engine.NewCaps(1, utils.MetaBusy)
	sts := engine.NewCapsStats(cfgDflt.CoreSCfg().CapsStatsInterval, caps, stopchan)
	shutdown := utils.NewSyncedChan()
	profiler := NewProfiler(new(sync.WaitGroup))
	expected := &CoreS{
		cfg:       cfgDflt,
		CapsStats: sts,
		caps:      caps,
		shutdown:  shutdown,
		profiler:  profiler,
	}
	rcv := NewCoreService(cfgDflt, caps, profiler, stopchan, shutdown, nil)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
	rcv.ShutdownEngine()
	select {
	case <-shutdown.Done():
	default:
		t.Error("engine did not shut down")
	}
}

func TestProfilerMemoryRestart(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		dir := t.TempDir()
		p := NewProfiler(new(sync.WaitGroup))
		t.Cleanup(func() { _ = p.StopMemoryProfiling() })
		if err := os.WriteFile(filepath.Join(dir, "mem_1.prof"), nil, 0600); err != nil {
			t.Fatal(err)
		}
		params := MemoryProfilingParams{
			DirPath: dir, Interval: 100 * time.Millisecond, MaxFiles: 1, UseTimestamp: true,
		}
		for range 2 {
			if err := p.StartMemoryProfiling(params); err != nil {
				t.Fatal(err)
			}
			synctest.Wait()
			time.Sleep(2 * params.Interval)
			synctest.Wait()
			if err := p.StopMemoryProfiling(); err != nil {
				t.Fatal(err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 2 {
				t.Fatalf("profiles=%d, want periodic and final", len(entries))
			}
		}
	})
}

func TestProfilerShutdown(t *testing.T) {
	var wg sync.WaitGroup
	p := NewProfiler(&wg)
	cpuPath := filepath.Join(t.TempDir(), utils.CpuPathCgr)
	if err := p.StartCPUProfiling(cpuPath); err != nil {
		t.Fatal(err)
	}
	fileCPU := p.fileCPU
	wg.Add(1)
	defer wg.Done()
	t.Cleanup(func() { _ = p.Shutdown(context.Background()) })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := p.Shutdown(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("shutdown error: got %v, want %v", err, context.Canceled)
	}
	if _, err := fileCPU.Stat(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("CPU file Stat after shutdown: got %v, want %v", err, os.ErrClosed)
	}
	if fi, err := os.Stat(cpuPath); err != nil {
		t.Fatalf("CPU profile missing: %v", err)
	} else if fi.Size() == 0 {
		t.Fatal("CPU profile is empty")
	}
}
