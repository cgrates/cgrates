//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"path"
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

func TestDiameterAgentReload1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.ListenCfg().BiJSONListen = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSChan := make(chan birpc.ClientConnector, 1)
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	cacheSChan <- cacheSrv
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	db := NewDataDBService(cfg, nil, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, server, make(chan birpc.ClientConnector, 1),
		nil, anz, srvDep)
	diamSrv := NewDiameterAgent(cfg, filterSChan, shdChan, nil, nil, srvDep)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(diamSrv, sS, db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	if diamSrv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "diamagent_mysql"),
			Section: config.DA_JSN,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !diamSrv.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err = diamSrv.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = diamSrv.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}

	cfg.DiameterAgentCfg().Enabled = false
	cfg.GetReloadChan(config.DA_JSN) <- struct{}{}
	err2 := diamSrv.Reload()
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err2)
	}
	time.Sleep(10 * time.Millisecond)
	if diamSrv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}

func TestDiameterAgentReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.ListenCfg().BiJSONListen = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSChan := make(chan birpc.ClientConnector, 1)
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	cacheSChan <- cacheSrv
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDiameterAgent(cfg, filterSChan, shdChan, nil, nil, srvDep)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	cfg.DiameterAgentCfg().Enabled = false
	srv.(*DiameterAgent).stopChan = make(chan struct{}, 1)
	srv.Shutdown()
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}

func TestDiameterAgentReload3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	caps := engine.NewCaps(cfg.CoreSCfg().Caps, cfg.CoreSCfg().CapsStrategy)
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSChan := make(chan birpc.ClientConnector, 1)
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	cacheSChan <- cacheSrv
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDiameterAgent(cfg, filterSChan, shdChan, nil, caps, srvDep)

	cfg.DiameterAgentCfg().Listeners[0].Network = "bad"
	cfg.DiameterAgentCfg().DictionariesPath = ""

	da := srv.(*DiameterAgent)
	if err = da.start(nil, da.caps); err != nil {
		t.Fatal(err)
	}
	cfg.DiameterAgentCfg().Enabled = false
	err = srv.Reload()
	if err != nil {
		t.Fatal(err)
	}

}
