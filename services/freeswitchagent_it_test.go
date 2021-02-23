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

/*
func TestFreeSwitchAgentReload(t *testing.T) {
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
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	sS := NewSessionService(cfg, db, server, make(chan rpcclient.ClientConnector, 1),
		shdChan, nil, nil, anz, srvDep)
	srv := NewFreeswitchAgent(cfg, shdChan, nil, srvDep)
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
		Path:    path.Join("/usr", "share", "cgrates", "tutorial_tests", "fs_evsock", "cgrates", "etc", "cgrates"),
		Section: config.FreeSWITCHAgentJSN,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expecting OK ,received %s", reply)
	}

	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	// the engine should be stopped as we could not connect to freeswitch
	agentCfg := &config.FsAgentCfg{
		Enabled:          true,
		CreateCdr:        true,
		SubscribePark:    true,
		EventSocketConns: []*config.FsConnCfg{},
	}

	srv.(*FreeswitchAgent).fS = agents.NewFSsessions(agentCfg, "", nil)
	runtime.Gosched()
	err := srv.Reload()
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	time.Sleep(10 * time.Millisecond)

}
func TestFreeSwitchAgentReload2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	cfg.SessionSCfg().Enabled = true
	cfg.SessionSCfg().ListenBijson = ""
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewFreeswitchAgent(cfg, shdChan, nil, srvDep)

	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	srv.(*FreeswitchAgent).fS = &agents.FSsessions{}
	if !srv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	err := srv.Start()
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
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewFreeswitchAgent(cfg, shdChan, nil, srvDep)

	if srv.IsRunning() {
		t.Fatalf("Expected service to be down")
	}
	srv.(*FreeswitchAgent).fS = &agents.FSsessions{}
	if !srv.IsRunning() {
		t.Fatalf("Expected service to be running")
	}
	err := srv.Start()
	if err == nil || err.Error() != "service already running" {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", "service already running", err)
	}
	err = srv.Shutdown()
	if err != nil {
		t.Fatalf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
}
*/
