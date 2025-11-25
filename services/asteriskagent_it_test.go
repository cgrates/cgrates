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
	"github.com/cgrates/rpcclient"
)

func TestAsteriskAgentReload(t *testing.T) {
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
		t.Error(err)
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
		shdChan, cm, anz, srvDep)
	astService := NewAsteriskAgent(cfg, shdChan, cm, srvDep)
	srvMngr.AddServices(astService, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	if astService.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "tutorial_tests", "asterisk_ari", "cgrates", "etc", "cgrates"),
			Section: config.AsteriskAgentJSN,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !astService.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	srvReload := astService.Reload()
	if srvReload != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", srvReload)
	}
	err = astService.Start()
	if err != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	cfg.AsteriskAgentCfg().Enabled = false
	cfg.GetReloadChan(config.AsteriskAgentJSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if astService.IsRunning() {
		t.Fatalf("Expected service to be down")
	}

}

func TestAsteriskAgentReload2(t *testing.T) {
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
		t.Error(err)
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
		shdChan, cm, anz, srvDep)
	astSrv := NewAsteriskAgent(cfg, shdChan, cm, srvDep)
	srvMngr.AddServices(astSrv, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	if err := srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	if astSrv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "tutorial_tests", "asterisk_ari", "cgrates", "etc", "cgrates"),
			Section: config.AsteriskAgentJSN,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !astSrv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	srvReload := astSrv.Reload()
	if srvReload != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", srvReload)
	}
	err = astSrv.Start()
	if err != utils.ErrServiceAlreadyRunning {
		t.Fatalf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	cfg.AsteriskAgentCfg().AsteriskConns = []*config.AsteriskConnCfg{
		{
			Alias:           "127.0.0.1:8448",
			Address:         "127.0.0.1:8088",
			User:            "cgrates",
			Password:        "CGRateS.org",
			ConnectAttempts: 0,
			Reconnects:      0,
		}}
	srvReload = astSrv.Reload()
	if srvReload != nil {
		t.Fatalf("\nExpecting <nil>,\n Received <%+v>", srvReload)
	}
	cfg.AsteriskAgentCfg().Enabled = false
	cfg.GetReloadChan(config.AsteriskAgentJSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if astSrv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
}
