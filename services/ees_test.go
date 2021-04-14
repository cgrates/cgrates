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
	"github.com/cgrates/cgrates/ees"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//TestEventExporterSCoverage for cover testing
func TestEventExporterSCoverage(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	srv := NewEventExporterService(cfg, filterSChan, engine.NewConnManager(cfg, nil), server, make(chan birpc.ClientConnector, 1), anz, srvDep)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2 := &EventExporterService{
		cfg:         cfg,
		filterSChan: filterSChan,
		connMgr:     engine.NewConnManager(cfg, nil),
		server:      server,
		intConnChan: make(chan birpc.ClientConnector, 1),
		anz:         anz,
		srvDep:      srvDep,
		rldChan:     make(chan struct{}, 1),
		eeS:         &ees.EventExporterS{},
		stopChan:    make(chan struct{}, 1),
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv2.ServiceName()
	if serviceName != utils.EventExporterS {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.EventExporterS, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if shouldRun != false {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}
	srv2.intConnChan <- &testMockClients{}
	shutErr := srv2.Shutdown()
	if shutErr != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", shutErr)
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
