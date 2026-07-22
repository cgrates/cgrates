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

// TestCoreSCoverage for cover testing
func TestCoreSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(1, "test_caps")
	server := cores.NewServer(nil)
	internalCoreSChan := make(chan birpc.ClientConnector, 1)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	srv := NewCoreService(cfg, caps, server,
		internalCoreSChan, anz, nil, nil, nil, srvDep)
	if srv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(srv))
	}
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv.cS = &cores.CoreService{}
	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.CoreS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.CoreS, serviceName)
	}
	shouldRun := srv.ShouldRun()
	if !reflect.DeepEqual(shouldRun, true) {
		t.Errorf("\nExpecting <true>,\n Received <%+v>", shouldRun)
	}
	getCoreS := srv.GetCoreS()
	if getCoreS == nil {
		t.Errorf("\nExpecting not <nil>,\n Received <%+v>", getCoreS)
	}
	//populates connChan with something in order to call the shutdown function
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Error(err)
	}
	srv.connChan <- cacheSrv
	srv.stopChan = make(chan struct{})
	getShut := srv.Shutdown()
	if getShut != nil {
		t.Errorf("\nExpecting not <nil>,\n Received <%+v>", getShut)
	}
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}

}

func TestCoreServiceReload(t *testing.T) {
	coreService := &CoreService{}
	err := coreService.Reload()
	if err != nil {
		t.Errorf("Expected Reload to return no error, got %v", err)
	}
}
