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

func TestResourceSReload(t *testing.T) {
	// utils.Logger.SetLogLevel(7)
	cfg := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.ThresholdSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)
	close(chS.GetPrecacheChannel(utils.CacheThresholdProfiles))
	close(chS.GetPrecacheChannel(utils.CacheThresholds))
	close(chS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes))
	close(chS.GetPrecacheChannel(utils.CacheResourceProfiles))
	close(chS.GetPrecacheChannel(utils.CacheResources))
	close(chS.GetPrecacheChannel(utils.CacheResourceFilterIndexes))
	chSCh := make(chan *engine.CacheS, 1)
	chSCh <- chS
	css := &CacheService{cacheCh: chSCh}
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(shdWg, nil, cfg.GetReloadChan())
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, srvDep)
	tS := NewThresholdService(cfg, db, css, filterSChan, nil, server, make(chan birpc.ClientConnector, 1), anz, srvDep)
	reS := NewResourceService(cfg, db, css, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)
	engine.NewConnManager(cfg)
	srvMngr.AddServices(tS, reS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep), db)
	ctx, cancel := context.WithCancel(context.TODO())
	srvMngr.StartServices(ctx, cancel)
	if reS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo")
	if err := cfg.V1ReloadConfig(context.Background(), &config.ReloadArgs{
		Section: config.ResourceSJSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	runtime.Gosched()
	runtime.Gosched()
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !reS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err := reS.Start(ctx, cancel)
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err = reS.Reload(ctx, cancel)
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	cfg.ResourceSCfg().Enabled = false
	cfg.GetReloadChan() <- config.SectionToService[config.ResourceSJSON]
	time.Sleep(10 * time.Millisecond)
	if reS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	cancel()
	time.Sleep(10 * time.Millisecond)
}
