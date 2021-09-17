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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestSIPAgentReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdWg := new(sync.WaitGroup)

	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(shdWg, nil, cfg.GetReloadChan())
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, server, make(chan birpc.ClientConnector, 1),
		nil, anz, srvDep)
	srv := NewSIPAgent(cfg, filterSChan, nil, srvDep)
	engine.NewConnManager(cfg)
	srvMngr.AddServices(srv, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	ctx, cancel := context.WithCancel(context.TODO())
	srvMngr.StartServices(ctx, cancel)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "sipagent_mysql")
	if err := cfg.V1ReloadConfig(context.Background(), &config.ReloadArgs{
		Section: config.SIPAgentJSON,
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
	srvStart := srv.Start(ctx, cancel)
	if srvStart != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, srvStart)
	}
	err := srv.Reload(ctx, cancel)
	if err != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	time.Sleep(10 * time.Millisecond)
	cfg.SIPAgentCfg().Enabled = false
	cfg.GetReloadChan() <- config.SectionToService[config.SIPAgentJSON]
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	cancel()
	time.Sleep(10 * time.Millisecond)
}

func TestSIPAgentReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewSIPAgent(cfg, filterSChan, nil, srvDep)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	cfg.SIPAgentCfg().RequestProcessors = []*config.RequestProcessor{
		{
			RequestFields: []*config.FCTemplate{
				{
					Type: utils.MetaTemplate,
				},
			},
		},
	}
	ctx, cancel := context.WithCancel(context.TODO())
	err := srv.Start(ctx, cancel)
	if err == nil || err.Error() != "no template with id: <>" {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", "no template with id: <>", err)
	}

}

func TestSIPAgentReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewSIPAgent(cfg, filterSChan, nil, srvDep)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	ctx, cancel := context.WithCancel(context.TODO())
	err := srv.Start(ctx, cancel)
	if err != nil {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
	srv.(*SIPAgent).oldListen = "test"
	err = srv.Reload(ctx, cancel)
	if err != nil {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", nil, err)
	}
}
