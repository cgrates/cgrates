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

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestStatSReload(t *testing.T) {
	// utils.Logger.SetLogLevel(7)
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.ThresholdSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan struct{}, 1)
	chS := engine.NewCacheS(cfg, nil)
	close(chS.GetPrecacheChannel(utils.CacheThresholdProfiles))
	close(chS.GetPrecacheChannel(utils.CacheThresholds))
	close(chS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes))
	close(chS.GetPrecacheChannel(utils.CacheStatQueueProfiles))
	close(chS.GetPrecacheChannel(utils.CacheStatQueues))
	close(chS.GetPrecacheChannel(utils.CacheStatFilterIndexes))
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	anz := NewAnalyzerService(cfg, server, filterSChan, engineShutdown, make(chan rpcclient.ClientConnector, 1))
	db := NewDataDBService(cfg, nil)
	tS := NewThresholdService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz)
	sS := NewStatService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(tS, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz), db)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if sS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err = cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.STATS_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !sS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.StatSCfg().Enabled = false
	cfg.GetReloadChan(config.STATS_JSON) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if sS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	close(engineShutdown)
	srvMngr.ShutdownServices(10 * time.Millisecond)
}
