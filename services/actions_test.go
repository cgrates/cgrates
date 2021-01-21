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

	"github.com/cgrates/cgrates/actions"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

//TestActionSCoverage for cover testing
func TestActionSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	actRPC := make(chan rpcclient.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	actS := NewActionService(cfg, db,
		chS, filterSChan, server, actRPC,
		anz, srvDep)
	if actS == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(actS))
	}
	//initialises an empty chan to call the reload function
	testChan := make(chan struct{})
	//populates actRPC with something in order to call the close function
	actRPC <- chS
	actS2 := &ActionService{
		cfg:         cfg,
		dm:          db,
		cacheS:      chS,
		filterSChan: filterSChan,
		server:      server,
		rldChan:     testChan,
		stopChan:    make(chan struct{}, 1),
		connChan:    actRPC,
		anz:         anz,
		srvDep:      srvDep,
	}
	if actS2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	actS2.acts = actions.NewActionS(cfg, &engine.FilterS{}, &engine.DataManager{})
	if !actS2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := actS2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.ActionS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ActionS, serviceName)
	}
	shouldRun := actS2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
}
