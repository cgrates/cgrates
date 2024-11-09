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

// import (
// 	"path"
// 	"runtime"
// 	"sync"
// 	"testing"
// 	"time"
//
// 	"github.com/cgrates/birpc"
// 	"github.com/cgrates/birpc/context"
// 	"github.com/cgrates/cgrates/agents"
// 	"github.com/cgrates/cgrates/commonlisteners"
// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/servmanager"
// 	"github.com/cgrates/cgrates/utils"
// )
//
// func TestDNSAgentStartReloadShut(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.SessionSCfg().Enabled = true
// 	cfg.SessionSCfg().ListenBijson = ""
// 	cfg.DNSAgentCfg().Enabled = true
// 	cfg.DNSAgentCfg().Listeners = []config.Listener{
// 		{
// 			Network: "udp",
// 			Address: ":2055",
// 		},
// 		{
// 			Network: "tcp",
// 			Address: ":2056",
// 		},
// 	}
// 	utils.Logger, _ = utils.NewSysLogger(cfg.GeneralCfg().NodeID, 7)
// 	filterSChan := make(chan *engine.FilterS, 1)
// 	filterSChan <- nil
// 	ctx, cancel := context.WithCancel(context.TODO())
// 	defer func() {
// 		cancel()
// 		time.Sleep(10 * time.Millisecond)
// 	}()
// 	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
// 	srv := NewDNSAgent(cfg, filterSChan, nil, srvDep)
// 	shdWg := new(sync.WaitGroup)
// 	srvMngr := servmanager.NewServiceManager(shdWg, nil, cfg)
// 	engine.NewConnManager(cfg)
// 	db := NewDataDBService(cfg, nil, false, srvDep)
// 	cls := commonlisteners.NewCommonListenerS(nil)
// 	anz := NewAnalyzerService(cfg, cls, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
// 	sS := NewSessionService(cfg, db, filterSChan, cls, make(chan birpc.ClientConnector, 1),
// 		nil, anz, srvDep)
// 	srvMngr.AddServices(srv, sS,
// 		NewLoaderService(cfg, db, filterSChan, cls, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
// 	runtime.Gosched()
// 	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
// 	if err := srv.Shutdown(); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(10 * time.Millisecond)
// 	if err := srv.Start(ctx, cancel); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(10 * time.Millisecond)
// 	if err := srv.Reload(ctx, cancel); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(10 * time.Millisecond)
// 	if err := srv.Shutdown(); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(10 * time.Millisecond)
// 	if srv.IsRunning() {
// 		t.Errorf("service is still running")
// 	}
// }
//
// func TestDNSAgentReloadFirst(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.SessionSCfg().Enabled = true
// 	cfg.SessionSCfg().ListenBijson = ""
// 	utils.Logger, _ = utils.NewLogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID, 7)
// 	filterSChan := make(chan *engine.FilterS, 1)
// 	filterSChan <- nil
// 	ctx, cancel := context.WithCancel(context.TODO())
// 	defer func() {
// 		cancel()
// 		time.Sleep(10 * time.Millisecond)
// 	}()
// 	shdWg := new(sync.WaitGroup)
//
// 	cls := commonlisteners.NewCommonListenerS(nil)
// 	srvMngr := servmanager.NewServiceManager(shdWg, nil, cfg)
// 	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
// 	db := NewDataDBService(cfg, nil, false, srvDep)
// 	anz := NewAnalyzerService(cfg, cls, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
// 	sS := NewSessionService(cfg, db, filterSChan, cls, make(chan birpc.ClientConnector, 1),
// 		nil, anz, srvDep)
// 	srv := NewDNSAgent(cfg, filterSChan, nil, srvDep)
// 	engine.NewConnManager(cfg)
// 	srvMngr.AddServices(srv, sS,
// 		NewLoaderService(cfg, db, filterSChan, cls, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
// 	srvMngr.StartServices(ctx, cancel)
// 	time.Sleep(100 * time.Millisecond)
// 	if srv.IsRunning() {
// 		t.Fatalf("Expected service to be down")
// 	}
// 	var reply string
// 	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "dnsagent_reload")
// 	if err := cfg.V1ReloadConfig(context.Background(), &config.ReloadArgs{
// 		Section: config.DNSAgentJSON,
// 	}, &reply); err != nil {
// 		t.Fatal(err)
// 	} else if reply != utils.OK {
// 		t.Fatalf("Expecting OK ,received %s", reply)
// 	}
// 	runtime.Gosched()
// 	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
// 	if !srv.IsRunning() {
// 		t.Fatalf("Expected service to be running")
// 	}
// 	err := srv.Start(ctx, cancel)
// 	if err == nil || err != utils.ErrServiceAlreadyRunning {
// 		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
// 	}
//
// 	err = srv.Reload(ctx, cancel)
// 	if err != nil {
// 		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
// 	}
// 	err = srv.Reload(ctx, cancel)
// 	if err != nil {
// 		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
// 	}
// 	time.Sleep(10 * time.Millisecond)
// 	cfg.DNSAgentCfg().Enabled = false
// 	time.Sleep(10 * time.Millisecond)
// 	cfg.GetReloadChan() <- config.SectionToService[config.DNSAgentJSON]
// 	time.Sleep(100 * time.Millisecond)
// 	if srv.IsRunning() {
// 		t.Fatalf("Expected service to be down")
// 	}
//
// }
//
// func TestDNSAgentReload2(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.SessionSCfg().Enabled = true
// 	cfg.SessionSCfg().ListenBijson = ""
// 	cfg.DNSAgentCfg().Enabled = true
// 	cfg.DNSAgentCfg().Listeners[0].Network = "test"
// 	cfg.DNSAgentCfg().Listeners[0].Address = "test"
// 	filterSChan := make(chan *engine.FilterS, 1)
// 	filterSChan <- nil
// 	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
// 	srv := NewDNSAgent(cfg, filterSChan, nil, srvDep)
// 	agentSrv, err := agents.NewDNSAgent(cfg, nil, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	runtime.Gosched()
// 	dnsSrv := srv.(*DNSAgent)
// 	dnsSrv.dns = agentSrv
// 	err = dnsSrv.listenAndServe(make(chan struct{}), func() {})
// 	if err == nil || err.Error() != "dns: bad network" {
// 		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "dns: bad network", err)
// 	}
// }
//
// func TestDNSAgentReload4(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.SessionSCfg().Enabled = true
// 	cfg.DNSAgentCfg().Enabled = true
// 	cfg.DNSAgentCfg().Listeners[0].Network = "tls"
// 	cfg.TLSCfg().ServerCerificate = "bad_certificate"
// 	cfg.TLSCfg().ServerKey = "bad_key"
// 	filterSChan := make(chan *engine.FilterS, 1)
// 	filterSChan <- nil
// 	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
// 	srv := NewDNSAgent(cfg, filterSChan, nil, srvDep)
//
// 	runtime.Gosched()
// 	dnsSrv := srv.(*DNSAgent)
// 	dnsSrv.dns = nil
// 	ctx, cancel := context.WithCancel(context.TODO())
// 	err := dnsSrv.Start(ctx, cancel)
// 	if err == nil || err.Error() != "load certificate error <open bad_certificate: no such file or directory>" {
// 		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "load certificate error <open bad_certificate: no such file or directory>", err)
// 	}
// 	dnsSrv.dns = nil
// }
//
// func TestDNSAgentReload5(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.SessionSCfg().Enabled = true
// 	cfg.DNSAgentCfg().Enabled = true
//
// 	filterSChan := make(chan *engine.FilterS, 1)
// 	filterSChan <- nil
// 	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
// 	srv := NewDNSAgent(cfg, filterSChan, nil, srvDep)
// 	ctx, cancel := context.WithCancel(context.TODO())
// 	err := srv.Start(ctx, cancel)
// 	if err != nil {
// 		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
// 	}
// 	time.Sleep(10 * time.Millisecond)
// 	runtime.Gosched()
// 	runtime.Gosched()
// 	err = srv.Reload(ctx, cancel)
// 	if err != nil {
// 		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
// 	}
// }
//
// func TestDNSAgentReload6(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	cfg.SessionSCfg().Enabled = true
// 	cfg.DNSAgentCfg().Enabled = true
//
// 	filterSChan := make(chan *engine.FilterS, 1)
// 	filterSChan <- nil
// 	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
// 	srv := NewDNSAgent(cfg, filterSChan, nil, srvDep)
// 	cfg.DNSAgentCfg().Listeners[0].Address = "127.0.0.1:0"
// 	ctx, cancel := context.WithCancel(context.TODO())
// 	err := srv.Start(ctx, cancel)
// 	if err != nil {
// 		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
// 	}
// 	cfg.DNSAgentCfg().Listeners[0].Network = "tls"
// 	cfg.TLSCfg().ServerCerificate = "bad_certificate"
// 	cfg.TLSCfg().ServerKey = "bad_key"
// 	time.Sleep(10 * time.Millisecond)
// 	runtime.Gosched()
// 	runtime.Gosched()
// 	err = srv.Reload(ctx, cancel)
// 	if err == nil || err.Error() != "load certificate error <open bad_certificate: no such file or directory>" {
// 		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "load certificate error <open bad_certificate: no such file or directory>", err)
// 	}
// }
