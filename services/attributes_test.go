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

// TestAttributeSCoverage for cover testing
func TestAttributeSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	chS := engine.NewCacheS(cfg, nil, nil)
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	attrRPC := make(chan birpc.ClientConnector, 1)
	db := NewDataDBService(cfg, nil, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	attrS := NewAttributeService(cfg, db, chS, filterSChan, server, attrRPC, anz, srvDep)
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
	chS = engine.NewCacheS(cfg, nil, nil)
	srv, err := engine.NewService(chS)
	if err != nil {
		t.Error(err)
	}
	attrS2.connChan <- srv
	shutdownErr := attrS2.Shutdown()
	if shutdownErr != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", shutdownErr)
	}
	if attrS2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}

func TestAttributeServiceReload(t *testing.T) {
	attrService := &AttributeService{}
	err := attrService.Reload()
	if err != nil {
		t.Errorf("Expected Reload to return no error, got %v", err)
	}
}
