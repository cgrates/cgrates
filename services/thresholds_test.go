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

//TestThresholdSCoverage for cover testing
func TestThresholdSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, srvDep)
	chS := NewCacheService(cfg, db, server, make(chan context.ClientConnector, 1), anz, nil, srvDep)
	tS := NewThresholdService(cfg, db, chS, filterSChan, nil, server, make(chan birpc.ClientConnector, 1), anz, srvDep)
	if tS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	thrs1 := engine.NewThresholdService(&engine.DataManager{}, &config.CGRConfig{}, &engine.FilterS{}, nil)
	tS2 := &ThresholdService{
		cfg:         cfg,
		dm:          db,
		cacheS:      chS,
		filterSChan: filterSChan,
		server:      server,
		thrs:        thrs1,
		connChan:    make(chan birpc.ClientConnector, 1),
		anz:         anz,
		srvDep:      srvDep,
	}
	if !tS2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := tS2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.ThresholdS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ThresholdS, serviceName)
	}
	shouldRun := tS2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
}
