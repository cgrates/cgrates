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
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/dispatchers"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestDispatcherSCoverage for cover testing
func TestDispatcherSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Enabled = true
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
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
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	srv2.connChan <- cacheSrv
	shutErr := srv2.Shutdown()
	if shutErr != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", shutErr)
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}

func TestDispatcherServiceReload(t *testing.T) {
	dspService := &DispatcherService{}
	err := dspService.Reload()
	if err != nil {
		t.Errorf("Reload() returned an error: %v", err)
	}
}

func TestNewDispatcherServiceMap(t *testing.T) {
	dspService := &dispatchers.DispatcherService{}
	srvMap, err := newDispatcherServiceMap(dspService)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
	if srvMap == nil {
		t.Fatal("Expected non-nil map, but got nil")
	}
	expectedLength := 22
	if len(srvMap) != expectedLength {
		t.Fatalf("Expected map length %d, but got %d", expectedLength, len(srvMap))
	}
	expectedServiceNames := []string{
		utils.AttributeSv1,
		utils.CacheSv1,
		utils.CDRsV1,
		utils.CDRsV2,
		utils.ChargerSv1,
		utils.ConfigSv1,
		utils.CoreSv1,
		utils.DispatcherSv1,
		utils.EeSv1,
		utils.ErSv1,
		utils.GuardianSv1,
		utils.RALsV1,
		utils.ReplicatorSv1,
		utils.ResourceSv1,
		utils.ThresholdSv1,
		utils.Responder,
		utils.RouteSv1,
		utils.SchedulerSv1,
		utils.SessionSv1,
		utils.StatSv1,
		utils.RankingSv1,
		utils.TrendSv1,
	}
	for _, name := range expectedServiceNames {
		if _, ok := srvMap[name]; !ok {
			t.Errorf("Expected service %s not found in the map", name)
		}
	}
}
