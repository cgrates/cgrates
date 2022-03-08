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

func TestFreeSwitchAgentReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	ctx, cancel := context.WithCancel(context.TODO())
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}()
	shdWg := new(sync.WaitGroup)

	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(shdWg, nil, cfg.GetReloadChan())
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1),
		nil, anz, srvDep)
	srv := NewFreeswitchAgent(cfg, nil, srvDep)
	engine.NewConnManager(cfg)
	srvMngr.AddServices(srv, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	srvMngr.StartServices(ctx, cancel)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "tutorial_tests", "fs_evsock", "cgrates", "etc", "cgrates")
	if err := cfg.V1ReloadConfig(context.Background(), &config.ReloadArgs{
		Section: config.FreeSWITCHAgentJSON,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}

	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	// the engine should be stopped as we could not connect to freeswitch

}

func TestFreeSwitchAgentReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewFreeswitchAgent(cfg, nil, srvDep)

	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	srv.(*FreeswitchAgent).fS = &agents.FSsessions{}
	if !srv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	ctx, cancel := context.WithCancel(context.TODO())
	err := srv.Start(ctx, cancel)
	if err == nil || err.Error() != "service already running" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "service already running", err)
	}
	err = srv.Shutdown()
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}
func TestFreeSwitchAgentReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewFreeswitchAgent(cfg, nil, srvDep)

	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	srv.(*FreeswitchAgent).fS = &agents.FSsessions{}
	if !srv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	ctx, cancel := context.WithCancel(context.TODO())
	err := srv.Start(ctx, cancel)
	if err == nil || err.Error() != "service already running" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "service already running", err)
	}
	err = srv.Shutdown()
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func TestFreeSwitchAgentReload4(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewFreeswitchAgent(cfg, nil, srvDep)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	agentCfg := &config.FsAgentCfg{
		Enabled:             true,
		SessionSConns:       nil,
		SubscribePark:       true,
		CreateCdr:           true,
		ExtraFields:         nil,
		LowBalanceAnnFile:   "",
		EmptyBalanceContext: "",
		EmptyBalanceAnnFile: "",
		MaxWaitConnection:   0,
		EventSocketConns: []*config.FsConnCfg{
			{
				Address:    "",
				Password:   "",
				Reconnects: 0,
				Alias:      "",
			},
		},
	}
	srv.(*FreeswitchAgent).fS = agents.NewFSsessions(agentCfg, "", nil)
	err := srv.(*FreeswitchAgent).connect(srv.(*FreeswitchAgent).fS, func() {})
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func TestFreeSwitchAgentReload5(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewFreeswitchAgent(cfg, nil, srvDep)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}

	srv.(*FreeswitchAgent).fS = nil
	ctx, cancel := context.WithCancel(context.TODO())
	err := srv.Start(ctx, cancel)
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}

func TestFreeSwitchAgentReload6(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewFreeswitchAgent(cfg, nil, srvDep)
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	agentCfg := &config.FsAgentCfg{
		Enabled:             true,
		SessionSConns:       nil,
		SubscribePark:       true,
		CreateCdr:           true,
		ExtraFields:         nil,
		LowBalanceAnnFile:   "",
		EmptyBalanceContext: "",
		EmptyBalanceAnnFile: "",
		MaxWaitConnection:   0,
		EventSocketConns: []*config.FsConnCfg{
			{
				Address:    "",
				Password:   "",
				Reconnects: 0,
				Alias:      "",
			},
		},
	}
	srv.(*FreeswitchAgent).fS = agents.NewFSsessions(agentCfg, "", nil)
	ctx, cancel := context.WithCancel(context.TODO())
	err := srv.Reload(ctx, cancel)
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}
