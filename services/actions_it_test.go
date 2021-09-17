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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestActionSReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheActionProfiles))
	close(chS.GetPrecacheChannel(utils.CacheActionProfilesFilterIndexes))
	chSCh := make(chan *engine.CacheS, 1)
	chSCh <- chS
	css := &CacheService{cacheCh: chSCh}
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(shdWg, nil, cfg.GetReloadChan())
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	actRPC := make(chan birpc.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	actS := NewActionService(cfg, db, css, filterSChan, nil, server, actRPC, anz, srvDep)
	engine.NewConnManager(cfg)
	srvMngr.AddServices(actS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	ctx, cancel := context.WithCancel(context.TODO())
	srvMngr.StartServices(ctx, cancel)
	if actS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo")
	if err := cfg.V1ReloadConfig(context.Background(), &config.ReloadArgs{
		Section: config.ActionSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	select {
	case d := <-actRPC:
		actRPC <- d
	case <-time.After(time.Second):
		t.Fatal("It took to long to reload the cache")
	}
	if !actS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err := actS.Start(ctx, cancel)
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = actS.Reload(ctx, cancel)
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.ActionSCfg().Enabled = false
	cfg.GetReloadChan() <- config.SectionToService[config.ActionSJSON]
	time.Sleep(10 * time.Millisecond)

	if actS.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	cancel()
	time.Sleep(10 * time.Millisecond)

}
