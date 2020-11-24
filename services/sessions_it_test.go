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

func TestSessionSReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.ChargerSCfg().Enabled = true
	cfg.RalsCfg().Enabled = true
	cfg.CdrsCfg().Enabled = true
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan struct{}, 1)
	chS := engine.NewCacheS(cfg, nil)

	close(chS.GetPrecacheChannel(utils.CacheChargerProfiles))
	close(chS.GetPrecacheChannel(utils.CacheChargerFilterIndexes))

	close(chS.GetPrecacheChannel(utils.CacheDestinations))
	close(chS.GetPrecacheChannel(utils.CacheReverseDestinations))
	close(chS.GetPrecacheChannel(utils.CacheRatingPlans))
	close(chS.GetPrecacheChannel(utils.CacheRatingProfiles))
	close(chS.GetPrecacheChannel(utils.CacheActions))
	close(chS.GetPrecacheChannel(utils.CacheActionPlans))
	close(chS.GetPrecacheChannel(utils.CacheAccountActionPlans))
	close(chS.GetPrecacheChannel(utils.CacheActionTriggers))
	close(chS.GetPrecacheChannel(utils.CacheSharedGroups))
	close(chS.GetPrecacheChannel(utils.CacheTimings))

	internalChan := make(chan rpcclient.ClientConnector, 1)
	internalChan <- nil
	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS

	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	db := NewDataDBService(cfg, nil)
	cfg.StorDbCfg().Type = utils.INTERNAL
	anz := NewAnalyzerService(cfg, server, filterSChan, engineShutdown, make(chan rpcclient.ClientConnector, 1))
	stordb := NewStorDBService(cfg)
	chrS := NewChargerService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz)
	schS := NewSchedulerService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz)
	ralS := NewRalService(cfg, chS, server,
		make(chan rpcclient.ClientConnector, 1), make(chan rpcclient.ClientConnector, 1),
		engineShutdown, nil, anz)
	cdrS := NewCDRServer(cfg, db, stordb, filterSChan, server,
		make(chan rpcclient.ClientConnector, 1),
		nil, anz)
	srv := NewSessionService(cfg, db, server, make(chan rpcclient.ClientConnector, 1), engineShutdown, nil, nil, anz)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(srv, chrS, schS, ralS, cdrS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz), db, stordb)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if stordb.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(&config.ReloadArgs{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongonew"),
		Section: config.SessionSJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.SessionSCfg().Enabled = false
	cfg.GetReloadChan(config.SessionSJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	close(engineShutdown)
	srvMngr.ShutdownServices(10 * time.Millisecond)
}
