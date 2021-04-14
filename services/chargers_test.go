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
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//TestChargerSCoverage for cover testing
func TestChargerSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Enabled = true
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	server := cores.NewServer(nil)
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	chrS1 := NewChargerService(cfg, db, chS,
		filterSChan, server, make(chan birpc.ClientConnector, 1),
		nil, anz, srvDep)
	if chrS1.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	chrS := &ChargerService{
		connChan:    make(chan birpc.ClientConnector, 1),
		cfg:         cfg,
		dm:          db,
		cacheS:      chS,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     nil,
		anz:         anz,
		srvDep:      srvDep,
	}
	if chrS.IsRunning() {
		t.Errorf("Expected service to be down")
	}

	chrS.chrS = &engine.ChargerService{}
	if !chrS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := chrS.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.ChargerS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ChargerS, serviceName)
	}
	shouldRun := chrS.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	chrS.connChan = make(chan birpc.ClientConnector, 1)
	chrS.connChan <- &testMockClients{}
	shutErr := chrS.Shutdown()
	if shutErr != nil {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", nil, shutErr)
	}

}
