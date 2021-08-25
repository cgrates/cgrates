//go:build integration
// +build integration

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
package services

import (
	"path"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestRadiusAgentReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	defer func() {
		shdChan.CloseOnce()
		time.Sleep(10 * time.Millisecond)
	}()
	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)

	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS

	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, server, make(chan rpcclient.ClientConnector, 1),
		shdChan, nil, anz, srvDep)
	srv := NewRadiusAgent(cfg, filterSChan, shdChan, nil, srvDep)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(srv, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "radagent_mysql"),
		Section: config.RA_JSN,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	runtime.Gosched()
	if !srv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	runtime.Gosched()
	err := srv.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = srv.Reload()
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.RadiusAgentCfg().Enabled = false
	cfg.GetReloadChan(config.RA_JSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
}

func TestRadiusAgentReload2(t *testing.T) {
	time.Sleep(10 * time.Millisecond)
	cfg := config.NewDefaultCGRConfig()

	cfg.SessionSCfg().Enabled = true
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	defer func() {
		shdChan.CloseOnce()
		time.Sleep(10 * time.Millisecond)
	}()
	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)

	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS

	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, server, make(chan rpcclient.ClientConnector, 1),
		shdChan, nil, anz, srvDep)
	srv := NewRadiusAgent(cfg, filterSChan, shdChan, nil, srvDep)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(srv, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "radagent_mysql"),
		Section: config.RA_JSN,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	runtime.Gosched()
	runtime.Gosched()
	if !srv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	runtime.Gosched()
	runtime.Gosched()
	err := srv.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = srv.Reload()
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	castSrv, canCastSrv := srv.(*RadiusAgent)
	if !canCastSrv {
		t.Fatalf("cannot cast")
	}

	err = srv.Reload()
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	castSrv.lnet = "test_string"
	err = srv.Reload()
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.RadiusAgentCfg().Enabled = false
	cfg.GetReloadChan(config.RA_JSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}

}

func TestRadiusAgentReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RadiusAgentCfg().ClientDictionaries = map[string]string{
		"test": "test",
	}
	cfg.SessionSCfg().Enabled = true
	cfg.RadiusAgentCfg().Enabled = true
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	defer func() {
		shdChan.CloseOnce()
		time.Sleep(10 * time.Millisecond)
	}()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewRadiusAgent(cfg, filterSChan, shdChan, nil, srvDep)
	err := srv.Start()
	if err == nil || err.Error() != "stat test: no such file or directory" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "stat test: no such file or directory", err)
	}
}

func TestRadiusAgentReload4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.RadiusAgentCfg().Enabled = true
	cfg.RadiusAgentCfg().ListenNet = "test"
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewRadiusAgent(cfg, filterSChan, shdChan, nil, srvDep)
	r, err := agents.NewRadiusAgent(cfg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	runtime.Gosched()
	rad := srv.(*RadiusAgent)
	rad.stopChan = make(chan struct{})
	err = rad.listenAndServe(r)
	if err == nil || err.Error() != "unsupported network: <test>" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "unsupported network: <test>", err)
	}
}
