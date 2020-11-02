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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestChargerSReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.AttributeSCfg().Enabled = true
	engineShutdown := make(chan bool, 1)
	chS := engine.NewCacheS(cfg, nil)
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	close(chS.GetPrecacheChannel(utils.CacheChargerProfiles))
	close(chS.GetPrecacheChannel(utils.CacheChargerFilterIndexes))
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := utils.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	db := NewDataDBService(cfg, nil)
	anz := NewAnalyzerService(cfg, server, engineShutdown, make(chan rpcclient.ClientConnector, 1))
	attrS := NewAttributeService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz)
	chrS := NewChargerService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(attrS, chrS,
		NewLoaderService(cfg, db, filterSChan, server,
			engineShutdown, make(chan rpcclient.ClientConnector, 1), nil, anz), db)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if chrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&config.ConfigReloadWithOpts{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.ChargerSCfgJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !chrS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.ChargerSCfg().Enabled = false
	cfg.GetReloadChan(config.ChargerSCfgJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if chrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	engineShutdown <- true
}
