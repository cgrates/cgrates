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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/dispatchers"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//TestDispatcherSCoverage for cover testing
func TestDispatcherSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	chS := NewCacheService(cfg, db, server, make(chan context.ClientConnector, 1), anz, nil, srvDep)
	srv := NewDispatcherService(cfg, db, chS, filterSChan, server,
		make(chan birpc.ClientConnector, 1), nil, anz, srvDep)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2 := DispatcherService{
		RWMutex:     sync.RWMutex{},
		cfg:         cfg,
		dm:          db,
		cacheS:      chS,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     nil,
		connChan:    make(chan birpc.ClientConnector, 1),
		anz:         anz,
		srvDep:      srvDep,
	}
	srv2.dspS = &dispatchers.DispatcherService{}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}

	serviceName := srv2.ServiceName()
	if serviceName != utils.DispatcherS {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.DispatcherS, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if shouldRun != false {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}

	srv2.connChan <- &testMockClients{}
	shutErr := srv2.Shutdown()
	if shutErr != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", shutErr)
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
