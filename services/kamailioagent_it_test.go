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

func TestKamailioAgentReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.SessionSCfg().Enabled = true
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan struct{}, 1)
	chS := engine.NewCacheS(cfg, nil)

	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS

	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	db := NewDataDBService(cfg, nil)
	anz := NewAnalyzerService(cfg, server, engineShutdown, make(chan rpcclient.ClientConnector, 1))
	sS := NewSessionService(cfg, db, server, make(chan rpcclient.ClientConnector, 1),
		engineShutdown, nil, nil, anz)
	srv := NewKamailioAgent(cfg, engineShutdown, nil)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(srv, sS,
		NewLoaderService(cfg, db, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz), db)
	if err = srvMngr.StartServices(); err != nil {
		t.Fatal(err)
	}
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfigFromPath(&config.ConfigReloadWithOpts{
		Path:    path.Join("/usr", "share", "cgrates", "tutorial_tests", "kamevapi", "cgrates", "etc", "cgrates"),
		Section: config.KamailioAgentJSN,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.KamAgentCfg().Enabled = false
	cfg.GetReloadChan(config.KamailioAgentJSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	select {
	case <-engineShutdown:
		// if the chanel was closed by the connect error (we did not start the kamailio for this tests)
	default:
		close(engineShutdown)
	}
	srvMngr.ShutdownServices(10 * time.Millisecond)
}
