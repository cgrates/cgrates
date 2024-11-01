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

package apis

import (
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
	coreService := cores.NewCoreService(cfg, caps, nil, make(chan struct{}), nil, nil)
	cS := NewCoreSv1(coreService)
	var reply map[string]any
	if err := cS.Status(context.Background(), &cores.V1StatusParams{}, &reply); err != nil {
		t.Error(err)
	}
}

func TestCoreSSleep(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, nil, make(chan struct{}), nil, nil)
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
	var closed bool
	coreService := cores.NewCoreService(cfg, caps, nil, make(chan struct{}), nil, func() { closed = true })
	cS := NewCoreSv1(coreService)
	arg := &utils.CGREvent{}
	var reply string
	if err := cS.Shutdown(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Errorf("Expected OK, received %+v", reply)
	}
	if !closed {
		t.Error("Did not stop the engine")
	}
}

func TestStartCPUProfiling(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, nil, make(chan struct{}), nil, nil)
	cS := NewCoreSv1(coreService)
	args := &utils.DirectoryArgs{
		DirPath: "dir_path",
		APIOpts: map[string]any{},
		Tenant:  "cgrates.org",
	}

	var reply string
	errExp := "could not create CPU profile: open dir_path/cpu.prof: no such file or directory"
	if err := cS.StartCPUProfiling(context.Background(), args, &reply); err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestStopCPUProfiling(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, nil, make(chan struct{}), nil, nil)
	cS := NewCoreSv1(coreService)
	args := &utils.TenantWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	var reply string
	errExp := "stop CPU profiling: not started yet"
	if err := cS.StopCPUProfiling(context.Background(), args, &reply); err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}

func TestStopMemoryProfiling(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(2, utils.MetaTopUp)
	coreService := cores.NewCoreService(cfg, caps, nil, make(chan struct{}), new(sync.WaitGroup), nil)
	cS := NewCoreSv1(coreService)
	var reply string
	errExp := "stop memory profiling: not started yet"
	if err := cS.StopMemoryProfiling(context.Background(),
		utils.TenantWithAPIOpts{
			Tenant:  "cgrates.org",
			APIOpts: map[string]any{},
		}, &reply); err.Error() != errExp {
		t.Errorf("Expected %v\n but received %v", errExp, err)
	}
}
