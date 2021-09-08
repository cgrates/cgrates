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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//TestAttributeSCoverage for cover testing
func TestAttributeSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	attrRPC := make(chan birpc.ClientConnector, 1)
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	chS := NewCacheService(cfg, db, server, make(chan context.ClientConnector, 1), anz, nil, srvDep)
	attrS := NewAttributeService(cfg, db, chS, filterSChan, server, attrRPC, anz, &DispatcherService{srvsReload: map[string]chan struct{}{}}, srvDep)
	if attrS == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(attrS))
	}
	attrS2 := &AttributeService{
		connChan:    make(chan birpc.ClientConnector, 1),
		cfg:         cfg,
		dm:          db,
		cacheS:      chS,
		filterSChan: filterSChan,
		server:      server,
		anz:         anz,
		srvDep:      srvDep,
		dspS:        &DispatcherService{srvsReload: map[string]chan struct{}{}},
	}
	if attrS2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	attrS2.attrS = &engine.AttributeService{}
	if !attrS2.IsRunning() {
		t.Errorf("Expected service to be running")
	}

	shouldRun := attrS2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}
	serviceName := attrS2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.AttributeS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.AttributeS, serviceName)
	}

	attrS2.connChan <- &testMockClients{}
	shutdownErr := attrS2.Shutdown()
	if shutdownErr != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", shutdownErr)
	}
	if attrS2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
