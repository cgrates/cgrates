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
	"github.com/cgrates/cgrates/rates"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//TestRateSCoverage for cover testing
func TestRateSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, make(chan birpc.ClientConnector, 1), srvDep)
	chS := NewCacheService(cfg, db, server, make(chan context.ClientConnector, 1), anz, nil, srvDep)
	rS := NewRateService(cfg, chS, filterSChan, db, server, make(chan birpc.ClientConnector, 1), anz, srvDep)

	if rS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	rS2 := RateService{
		cfg:         cfg,
		filterSChan: filterSChan,
		dmS:         db,
		cacheS:      chS,
		server:      server,
		stopChan:    make(chan struct{}),
		intConnChan: make(chan birpc.ClientConnector, 1),
		anz:         anz,
		srvDep:      srvDep,
		rateS:       &rates.RateS{},
	}
	if !rS2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := rS2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.RateS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.RateS, serviceName)
	}
	shouldRun := rS2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	rS2.intConnChan <- &testMockClients{}
	rS2.Shutdown()
	if rS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
