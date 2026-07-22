// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

// TestSupplierSCoverage for cover testing
func TestSupplierSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.StatSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	supS := NewRouteService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)

	if supS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	supS2 := &RouteService{
		cfg:         cfg,
		dm:          db,
		cacheS:      chS,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     nil,
		routeS:      &engine.RouteService{},
		connChan:    make(chan birpc.ClientConnector, 1),
		anz:         anz,
		srvDep:      srvDep,
	}
	if !supS2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := supS2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.RouteS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.RouteS, serviceName)
	}
	shouldRun := supS2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	supS2.connChan <- cacheSrv
	supS2.Shutdown()
	if supS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}

func TestRoutesReload(t *testing.T) {
	routeS := &RouteService{}
	err := routeS.Reload()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
