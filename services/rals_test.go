/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package services

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestRalsCoverage for cover testing
func TestRalsCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cfg.ThresholdSCfg().Enabled = true
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cfg.StorDbCfg().Type = utils.MetaInternal
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	ralS := NewRalService(cfg, chS, server,
		make(chan birpc.ClientConnector, 1),
		make(chan birpc.ClientConnector, 1),
		shdChan, nil, anz, srvDep, filterSChan)
	if ralS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	ralS2 := RalService{
		responder: &ResponderService{
			cfg:      cfg,
			server:   server,
			shdChan:  shdChan,
			resp:     &engine.Responder{},
			connChan: make(chan birpc.ClientConnector, 1),
			anz:      anz,
			srvDep:   srvDep,
		},
		cfg:      cfg,
		cacheS:   chS,
		server:   server,
		rals:     &v1.RALsV1{},
		connChan: make(chan birpc.ClientConnector, 1),
	}
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	ralS2.responder.connChan <- cacheSrv
	if !ralS2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := ralS2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.RALService) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.RALService, serviceName)
	}
	shouldRun := ralS2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	if !reflect.DeepEqual(ralS2.GetResponder(), ralS2.responder) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", ralS2.responder, ralS2.GetResponder())
	}
	ralS2.connChan <- cacheSrv
	ralS2.Shutdown()
	if ralS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
