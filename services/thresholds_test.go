/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package services

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestThresholdSCoverage for cover testing
func TestThresholdSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, false, srvDep)
	tS := NewThresholdService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)
	if tS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	thrs1 := engine.NewThresholdService(&engine.DataManager{}, &config.CGRConfig{}, &engine.FilterS{}, nil)
	tS2 := &ThresholdService{
		cfg:         cfg,
		dm:          db,
		cacheS:      chS,
		filterSChan: filterSChan,
		server:      server,
		thrs:        thrs1,
		connChan:    make(chan birpc.ClientConnector, 1),
		anz:         anz,
		srvDep:      srvDep,
	}
	if !tS2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := tS2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.ThresholdS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ThresholdS, serviceName)
	}
	shouldRun := tS2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
}

func TestThresholdServiceStartBiRPC(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	close(chS.GetPrecacheChannel(utils.CacheThresholdProfiles))
	close(chS.GetPrecacheChannel(utils.CacheThresholds))
	close(chS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes))
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, false, srvDep)
	db.GetDMChan() <- nil
	engine.NewConnManager(cfg, nil)

	tS := NewThresholdService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)

	done := make(chan error, 1)
	go func() { done <- tS.Start() }()
	t.Cleanup(func() {
		if tS.IsRunning() {
			_ = tS.Shutdown()
		}
	})

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start() returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start() blocked for 5s, likely deadlock")
	}
	if !tS.IsRunning() {
		t.Error("expected service to be running after Start()")
	}
}

func TestThresholdServiceRestart(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	close(chS.GetPrecacheChannel(utils.CacheThresholdProfiles))
	close(chS.GetPrecacheChannel(utils.CacheThresholds))
	close(chS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes))
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, false, srvDep)
	db.GetDMChan() <- nil
	engine.NewConnManager(cfg, nil)

	connChan := make(chan birpc.ClientConnector, 1)
	tS := NewThresholdService(cfg, db, chS, filterSChan, server, connChan, nil, anz, srvDep)

	if err := tS.Start(); err != nil {
		t.Fatalf("first Start() error: %v", err)
	}
	t.Cleanup(func() {
		if tS.IsRunning() {
			_ = tS.Shutdown()
		}
	})
	if !tS.IsRunning() {
		t.Fatal("expected service to be running after first Start()")
	}

	if err := tS.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}
	if tS.IsRunning() {
		t.Fatal("expected service to be stopped after Shutdown()")
	}

	if err := tS.Start(); err != nil {
		t.Fatalf("second Start() error: %v", err)
	}
	if !tS.IsRunning() {
		t.Fatal("expected service to be running after second Start()")
	}
}
