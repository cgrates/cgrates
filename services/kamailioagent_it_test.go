//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestKamailioAgentReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.SessionSCfg().Enabled = true
	cfg.ListenCfg().BiJSONListen = ""
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
	internalSessionSChan := make(chan birpc.ClientConnector, 1)
	cm := engine.NewConnManager(cfg, map[string]chan context.ClientConnector{
		utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS): internalSessionSChan,
	})
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, cm)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}

	db := NewDataDBService(cfg, cm, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, server, make(chan birpc.ClientConnector, 1),
		cm, anz, srvDep)
	srv := NewKamailioAgent(cfg, shdChan, cm, nil, srvDep)
	srvMngr.AddServices(srv, sS, db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}

	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "tutorial_tests", "kamevapi", "cgrates", "etc", "cgrates"),
			Section: config.KamailioAgentJSN,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}

	runtime.Gosched()
	time.Sleep(10 * time.Millisecond)
	kaCfg := &config.KamAgentCfg{
		Enabled:       true,
		SessionSConns: []string{utils.ConcatenatedKey("*birpc_internal", utils.MetaSessionS)},
		CreateCdr:     true,
		RouteProfile:  false,
		EvapiConns:    []*config.KamConnCfg{{Address: "127.0.0.1:8448", Reconnects: 10, Alias: "randomAlias"}},
		Timezone:      "Local",
	}

	srv.(*KamailioAgent).kam, err = agents.NewKamailioAgent(kaCfg, cm, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = srv.Reload()
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	// the engine should be stopped as we could not connect to kamailio
}

func TestKamailioAgentReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.ListenCfg().BiJSONListen = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewKamailioAgent(cfg, shdChan, nil, nil, srvDep)
	srvKam := &agents.KamailioAgent{}
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	srv.(*KamailioAgent).kam = srvKam
	if !srv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	err := srv.Start()
	if err == nil || err.Error() != "service already running" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "service already running", err)
	}
}

func TestKamailioAgentReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.ListenCfg().BiJSONListen = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewKamailioAgent(cfg, shdChan, nil, nil, srvDep)
	srvKam := &agents.KamailioAgent{}
	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	srv.(*KamailioAgent).kam = srvKam
	if !srv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	err := srv.Start()
	if err == nil || err.Error() != "service already running" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "service already running", err)
	}
}
