//go:build integration
// +build integration

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

func TestDNSAgentStartReloadShut(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBiJSON = ""
	cfg.DNSAgentCfg().Enabled = true
	cfg.DNSAgentCfg().Listeners = []config.DnsListener{
		{
			Network: "udp",
			Address: ":2055",
		},
		{
			Network: "tcp",
			Address: ":2056",
		},
	}
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, nil, srvDep)
	shdWg := new(sync.WaitGroup)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, nil, false, srvDep)
	server := cores.NewServer(nil)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, server, make(chan birpc.ClientConnector, 1),
		shdChan, nil, anz, srvDep)
	srvMngr.AddServices(srv, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if err := srv.Shutdown(); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := srv.Start(); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := srv.Reload(); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if err := srv.Shutdown(); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Errorf("service is still running")
	}
}

func TestDNSAgentReloadFirst(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBiJSON = ""
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
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	cacheSChan := make(chan birpc.ClientConnector, 1)
	cacheSChan <- cacheSrv

	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, server, make(chan birpc.ClientConnector, 1),
		shdChan, nil, anz, srvDep)
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, nil, srvDep)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(srv, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "dnsagent_reload"),
			Section: config.DNSAgentJson,
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
	err = srv.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "dnsagent_reload"),
			Section: config.DNSAgentJson,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "dnsagent_reload"),
			Section: config.DNSAgentJson,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond)
	cfg.DNSAgentCfg().Enabled = false
	time.Sleep(10 * time.Millisecond)
	cfg.GetReloadChan(config.DNSAgentJson) <- struct{}{}
	time.Sleep(100 * time.Millisecond)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
}

func TestDNSAgentReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBiJSON = ""
	cfg.DNSAgentCfg().Enabled = true
	cfg.DNSAgentCfg().Listeners[0].Network = "test"
	cfg.DNSAgentCfg().Listeners[0].Address = "test"
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, nil, srvDep)
	agentSrv, err := agents.NewDNSAgent(cfg, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	runtime.Gosched()
	dnsSrv := srv.(*DNSAgent)
	dnsSrv.dns = agentSrv

	err = dnsSrv.listenAndServe(make(chan struct{}))
	if err == nil || err.Error() != "dns: bad network" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "dns: bad network", err)
	}
}

func TestDNSAgentReload4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.DNSAgentCfg().Enabled = true
	cfg.DNSAgentCfg().Listeners[0].Network = "tls"
	cfg.TLSCfg().ServerCerificate = "bad_certificate"
	cfg.TLSCfg().ServerKey = "bad_key"
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, nil, srvDep)

	runtime.Gosched()
	dnsSrv := srv.(*DNSAgent)
	dnsSrv.dns = nil

	err := dnsSrv.Start()
	if err == nil || err.Error() != "load certificate error <open bad_certificate: no such file or directory>" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "load certificate error <open bad_certificate: no such file or directory>", err)
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
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, nil, srvDep)
	err := srv.Start()
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	time.Sleep(10 * time.Millisecond)
	runtime.Gosched()
	runtime.Gosched()

	err = srv.Reload()
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
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
	cfg.DNSAgentCfg().Listeners[0].Address = "127.0.0.1:0"
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, nil, srvDep)
	time.Sleep(10 * time.Millisecond)

	err := srv.Start()
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	time.Sleep(10 * time.Millisecond)
	runtime.Gosched()
	runtime.Gosched()

	cfg.DNSAgentCfg().Listeners[0].Network = "tls"
	cfg.TLSCfg().ServerCerificate = "bad_certificate"
	cfg.TLSCfg().ServerKey = "bad_key"
	err = srv.Reload()
	if err == nil || err.Error() != "load certificate error <open bad_certificate: no such file or directory>" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "load certificate error <open bad_certificate: no such file or directory>", err)
	}

}
