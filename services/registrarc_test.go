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
	"github.com/cgrates/cgrates/registrarc"
	"github.com/cgrates/cgrates/utils"
)

// TestDispatcherCoverage for cover testing
func TestDispatcherHCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	rpcInternal := map[string]chan birpc.ClientConnector{}
	cM := engine.NewConnManager(cfg, rpcInternal)
	srv := NewRegistrarCService(cfg, server, cM, anz, srvDep)
	if srv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(srv))
	}
	srv2 := &RegistrarCService{
		cfg:     cfg,
		server:  server,
		connMgr: cM,
		anz:     anz,
		srvDep:  srvDep,
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2.dspS = &registrarc.RegistrarCService{}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}

	serviceName := srv2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.RegistrarC) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.RegistrarC, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	srv2.stopChan = make(chan struct{}, 1)
	srv2.dspS = registrarc.NewRegistrarCService(cfg, cM)
	shutdownSrv := srv2.Shutdown()
	if shutdownSrv != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", shutdownSrv)
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
