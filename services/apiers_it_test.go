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
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestApiersReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan struct{}, 2)
	chS := engine.NewCacheS(cfg, nil)
	close(chS.GetPrecacheChannel(utils.CacheThresholdProfiles))
	close(chS.GetPrecacheChannel(utils.CacheThresholds))
	close(chS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes))
	close(chS.GetPrecacheChannel(utils.CacheActionPlans))

	cfg.ThresholdSCfg().Enabled = true
	cfg.SchedulerCfg().Enabled = true
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	db := NewDataDBService(cfg, nil)
	cfg.StorDbCfg().Type = utils.INTERNAL
	stordb := NewStorDBService(cfg)
	anz := NewAnalyzerService(cfg, server, filterSChan, engineShutdown, make(chan rpcclient.ClientConnector, 1))
	schS := NewSchedulerService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz)
	tS := NewThresholdService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz)
	apiSv1 := NewAPIerSv1Service(cfg, db, stordb, filterSChan, server, schS, new(ResponderService),
		make(chan rpcclient.ClientConnector, 1), nil, anz)

	apiSv2 := NewAPIerSv2Service(apiSv1, cfg, server, make(chan rpcclient.ClientConnector, 1), anz)
	srvMngr.AddServices(apiSv1, apiSv2, schS, tS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz), db, stordb)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if apiSv1.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if apiSv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if stordb.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(&config.ConfigReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.ApierS,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(100 * time.Millisecond) //need to switch to gorutine
	if !apiSv1.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !apiSv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !stordb.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.ApierCfg().Enabled = false
	cfg.GetReloadChan(config.ApierS) <- struct{}{}
	time.Sleep(100 * time.Millisecond)
	if apiSv1.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if apiSv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	close(engineShutdown)
	srvMngr.ShutdownServices(10 * time.Millisecond)
}
