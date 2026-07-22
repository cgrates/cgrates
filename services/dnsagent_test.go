// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestDNSAgentCoverage for cover testing
func TestDNSAgentCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.ListenCfg().BiJSONListen = ""
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Fatal(err)
	}
	cacheSChan := make(chan birpc.ClientConnector, 1)
	cacheSChan <- cacheSrv
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDNSAgent(cfg, filterSChan, shdChan, nil, nil, srvDep)
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	dns, _ := agents.NewDNSAgent(cfg, &engine.FilterS{}, nil, nil)
	srv2 := DNSAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		stopChan:    make(chan struct{}),
		connMgr:     nil,
		srvDep:      srvDep,
		dns:         dns,
	}

	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv2.ServiceName()
	if serviceName != utils.DNSAgent {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.DNSAgent, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if shouldRun != false {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}
	srv2.Shutdown()
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
