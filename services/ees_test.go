// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/ees"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestEventExporterSCoverage for cover testing
func TestEventExporterSCoverage(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	chS := engine.NewCacheS(cfg, nil, nil)
	cfg.AttributeSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	srv := NewEventExporterService(cfg, filterSChan, engine.NewConnManager(cfg, nil), server, make(chan birpc.ClientConnector, 1), anz, srvDep)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2 := &EventExporterService{
		cfg:         cfg,
		filterSChan: filterSChan,
		connMgr:     engine.NewConnManager(cfg, nil),
		server:      server,
		intConnChan: make(chan birpc.ClientConnector, 1),
		anz:         anz,
		srvDep:      srvDep,
		eeS:         &ees.EventExporterS{},
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv2.ServiceName()
	if serviceName != utils.EEs {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.EEs, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if shouldRun != false {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	srv2.intConnChan <- cacheSrv
	shutErr := srv2.Shutdown()
	if shutErr != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", shutErr)
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
