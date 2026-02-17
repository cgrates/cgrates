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
	"github.com/cgrates/cgrates/sessions"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestSessionSCoverage for cover testing
func TestSessionSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ChargerSCfg().Enabled = true
	cfg.RalsCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	internalChan := make(chan birpc.ClientConnector, 1)
	internalChan <- nil
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	cacheSChan := make(chan birpc.ClientConnector, 1)
	cacheSChan <- cacheSrv
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	cfg.StorDbCfg().Type = utils.MetaInternal
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	srv := NewSessionService(cfg, db, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)
	engine.NewConnManager(cfg, nil)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2 := SessionService{
		cfg:      cfg,
		dm:       db,
		server:   server,
		connChan: make(chan birpc.ClientConnector, 1),
		connMgr:  nil,
		anz:      anz,
		srvDep:   srvDep,
		sm:       &sessions.SessionS{},
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.SessionS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.SessionS, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
}

func TestReload(t *testing.T) {
	smg := &SessionService{}
	err := smg.Reload()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSessionServiceStartBiRPC(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	// Default config already has BiJSONListen set.
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, false, srvDep)
	engine.NewConnManager(cfg, nil)

	smg := NewSessionService(cfg, db, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)

	done := make(chan error, 1)
	go func() { done <- smg.Start() }()
	t.Cleanup(func() {
		if smg.IsRunning() {
			_ = smg.Shutdown()
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
	if !smg.IsRunning() {
		t.Error("expected service to be running after Start()")
	}
}

func TestSessionServiceRestart(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, false, srvDep)
	engine.NewConnManager(cfg, nil)

	connChan := make(chan birpc.ClientConnector, 1)
	smg := NewSessionService(cfg, db, server, connChan, nil, anz, srvDep)

	if err := smg.Start(); err != nil {
		t.Fatalf("first Start() error: %v", err)
	}
	t.Cleanup(func() {
		if smg.IsRunning() {
			_ = smg.Shutdown()
		}
	})
	if !smg.IsRunning() {
		t.Fatal("expected service to be running after first Start()")
	}

	if err := smg.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}
	if smg.IsRunning() {
		t.Fatal("expected service to be stopped after Shutdown()")
	}

	if err := smg.Start(); err != nil {
		t.Fatalf("second Start() error: %v", err)
	}
	if !smg.IsRunning() {
		t.Fatal("expected service to be running after second Start()")
	}
}
