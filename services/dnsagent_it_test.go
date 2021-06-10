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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestDNSAgentReload(t *testing.T) {
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

	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, server, make(chan birpc.ClientConnector, 1),
		shdChan, nil, nil, anz, srvDep)
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, srvDep)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(srv, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "dnsagent_reload")
	if err := cfg.V1ReloadConfig(context.Background(), &config.ReloadArgs{
		Section: config.DNSAgentJSON,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !srv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	err := srv.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}

	err = srv.Reload()
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	err = srv.Reload()
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}

	cfg.DNSAgentCfg().Enabled = false
	cfg.GetReloadChan(config.DNSAgentJSON) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}

}

func TestDNSAgentReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	cfg.DNSAgentCfg().Enabled = true
	cfg.DNSAgentCfg().ListenNet = "test"
	cfg.DNSAgentCfg().Listen = "test"
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, srvDep)
	agentSrv, err := agents.NewDNSAgent(cfg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	runtime.Gosched()
	dnsSrv := srv.(*DNSAgent)
	dnsSrv.dns = agentSrv
	err = dnsSrv.listenAndServe()
	if err == nil || err.Error() != "dns: bad network" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "dns: bad network", err)
	}
}

func TestDNSAgentReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.DNSAgentCfg().Enabled = true
	cfg.DNSAgentCfg().ListenNet = "test"
	cfg.DNSAgentCfg().Listen = "test"
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, srvDep)
	agentSrv, err := agents.NewDNSAgent(cfg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	runtime.Gosched()
	dnsSrv := srv.(*DNSAgent)
	dnsSrv.dns = agentSrv
	err = dnsSrv.Reload()
	if err == nil || err.Error() != "dns: server not started" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "dns: server not started", err)
	}
}

func TestDNSAgentReload4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.DNSAgentCfg().Enabled = true
	cfg.DNSAgentCfg().ListenNet = "tls"
	cfg.TLSCfg().ServerCerificate = "bad_certificate"
	cfg.TLSCfg().ServerKey = "bad_key"
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, srvDep)

	runtime.Gosched()
	dnsSrv := srv.(*DNSAgent)
	dnsSrv.dns = nil
	err := dnsSrv.Start()
	if err == nil || err.Error() != "open bad_certificate: no such file or directory" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "open bad_certificate: no such file or directory", err)
	}
	dnsSrv.dns = nil
}

func TestDNSAgentReload5(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.DNSAgentCfg().Enabled = true

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, srvDep)
	err := srv.Start()
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	srv.(*DNSAgent).oldListen = "127.0.0.1:2093"
	time.Sleep(10 * time.Millisecond)
	runtime.Gosched()
	runtime.Gosched()
	err = srv.Reload()
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func TestDNSAgentReload6(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.DNSAgentCfg().Enabled = true

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, srvDep)
	err := srv.Start()
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	srv.(*DNSAgent).oldListen = "127.0.0.1:2093"
	cfg.DNSAgentCfg().ListenNet = "tls"
	cfg.TLSCfg().ServerCerificate = "bad_certificate"
	cfg.TLSCfg().ServerKey = "bad_key"
	time.Sleep(10 * time.Millisecond)
	runtime.Gosched()
	runtime.Gosched()
	err = srv.Reload()
	if err == nil || err.Error() != "open bad_certificate: no such file or directory" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "open bad_certificate: no such file or directory", err)
	}
}
