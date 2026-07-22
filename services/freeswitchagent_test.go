// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestFreeSwitchAgentCoverage for cover testing
func TestFreeSwitchAgentCoverage(t *testing.T) {
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

	srv := NewFreeswitchAgent(cfg, shdChan, nil, nil, srvDep)

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2 := FreeswitchAgent{
		cfg:     cfg,
		shdChan: shdChan,
		fS:      &agents.FSsessions{},
		connMgr: nil,
		srvDep:  srvDep,
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.FreeSWITCHAgent) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.FreeSWITCHAgent, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
}
