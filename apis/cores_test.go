// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCoreSStatus(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, cores.NewProfiler(new(sync.WaitGroup)), make(chan struct{}), nil, nil)
	cS := NewCoreSv1(coreService)
	var reply map[string]any
	if err := cS.Status(context.Background(), &cores.V1StatusParams{}, &reply); err != nil {
		t.Error(err)
	}
}

func TestCoreSSleep(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, cores.NewProfiler(new(sync.WaitGroup)), make(chan struct{}), nil, nil)
	cS := NewCoreSv1(coreService)
	arg := &utils.DurationArgs{
		Duration: 1 * time.Millisecond,
	}
	var reply string
	if err := cS.Sleep(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Errorf("Expected OK, received %+v", reply)
	}
}

func TestCoreSShutdown(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	shutdown := utils.NewSyncedChan()
	coreService := cores.NewCoreService(cfg, caps, cores.NewProfiler(new(sync.WaitGroup)), make(chan struct{}), shutdown, nil)
	cS := NewCoreSv1(coreService)
	arg := &utils.CGREvent{}
	var reply string
	if err := cS.Shutdown(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Errorf("Expected OK, received %+v", reply)
	}
	select {
	case <-shutdown.Done():
	default:
		t.Error("engine did not shut down")
	}
}

func TestStartCPUProfiling(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, cores.NewProfiler(new(sync.WaitGroup)), make(chan struct{}), nil, nil)
	cS := NewCoreSv1(coreService)
	dir := filepath.Join(t.TempDir(), "missing")
	args := &utils.DirectoryArgs{
		DirPath: dir,
		APIOpts: map[string]any{},
		Tenant:  "cgrates.org",
	}

	var reply string
	errExp := "could not create CPU profile: open " + filepath.Join(dir, utils.CpuPathCgr) + ": no such file or directory"
	err := cS.StartCPUProfiling(context.Background(), args, &reply)
	if err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestStopCPUProfiling(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, cores.NewProfiler(new(sync.WaitGroup)), make(chan struct{}), nil, nil)
	cS := NewCoreSv1(coreService)
	args := &utils.TenantWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	var reply string
	errExp := "stop CPU profiling: not started yet"
	if err := cS.StopCPUProfiling(context.Background(), args, &reply); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestStopMemoryProfiling(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, cores.NewProfiler(new(sync.WaitGroup)), make(chan struct{}), nil, nil)
	cS := NewCoreSv1(coreService)
	var reply string
	errExp := "stop memory profiling: not started yet"
	if err := cS.StopMemoryProfiling(context.Background(),
		utils.TenantWithAPIOpts{
			Tenant:  "cgrates.org",
			APIOpts: map[string]any{},
		}, &reply); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}
