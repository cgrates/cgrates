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
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestCdrsCoverage for cover testing
func TestCdrsCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	chS := engine.NewCacheS(cfg, nil, nil)
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	cfg.ChargerSCfg().Enabled = true
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	cfg.StorDbCfg().Type = utils.MetaInternal
	stordb := NewStorDBService(cfg, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	cdrsRPC := make(chan birpc.ClientConnector, 1)
	cdrS := NewCDRServer(cfg, db, stordb, filterSChan, server,
		cdrsRPC, nil, anz, srvDep)
	if cdrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	//populates cdrS2  with something in order to call the close funct
	cdrS2 := &CDRServer{
		RWMutex:     sync.RWMutex{},
		cfg:         cfg,
		dm:          db,
		storDB:      stordb,
		filterSChan: filterSChan,
		server:      server,
		connChan:    make(chan birpc.ClientConnector, 1),
		connMgr:     nil,
		stopChan:    make(chan struct{}, 1),
		anz:         anz,
		srvDep:      srvDep,
		cdrS:        &engine.CDRServer{},
	}
	srv, err := engine.NewService(chS)
	if err != nil {
		t.Error(err)
	}
	cdrS2.connChan <- srv
	cdrS2.stopChan <- struct{}{}
	if !cdrS2.IsRunning() {
		t.Errorf("Expected service to be running")
	}

	serviceName := cdrS2.ServiceName()
	if serviceName != utils.CDRServer {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.APIerSv1, serviceName)
	}
	shouldRun := cdrS.ShouldRun()
	if shouldRun != false {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}

	shutdownApi1 := cdrS2.Shutdown()
	if shutdownApi1 != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, shutdownApi1)
	}

	if cdrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}

func TestCDRServerReload(t *testing.T) {
	cdrService := &CDRServer{}
	err := cdrService.Reload()
	if err != nil {
		t.Errorf("Expected Reload to return no error, got %v", err)
	}
}
