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

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestCdrsReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan bool, 1)
	chS := engine.NewCacheS(cfg, nil)

	close(chS.GetPrecacheChannel(utils.CacheChargerProfiles))
	close(chS.GetPrecacheChannel(utils.CacheChargerFilterIndexes))
	cfg.ChargerSCfg().Enabled = true
	cfg.RalsCfg().RALsEnabled = true
	responderChan := make(chan rpcclient.RpcClientConnection, 1)
	responderChan <- v1.NewResourceSv1(nil)
	server := utils.NewServer()
	srvMngr := servmanager.NewServiceManager(cfg /*dm*/, nil,
		/*cdrStorage*/ nil,
		/*loadStorage*/ nil, filterSChan,
		server, nil, engineShutdown)
	srvMngr.SetCacheS(chS)
	cdrS := NewCDRServer()
	srvMngr.AddService(cdrS, NewResponderService(responderChan), NewChargerService())
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if cdrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(&config.ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.CDRS_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !cdrS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.CdrsCfg().Enabled = false
	cfg.GetReloadChan(config.CDRS_JSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if cdrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	engineShutdown <- true
}
