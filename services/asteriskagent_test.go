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

// TestAsteriskAgentCoverage for cover testing
func TestAsteriskAgentCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.ListenCfg().BiJSONListen = ""
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Error(err)
	}
	cacheSChan := make(chan birpc.ClientConnector, 1)
	cacheSChan <- cacheSrv
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	astSrv := NewAsteriskAgent(cfg, shdChan, nil, nil, srvDep)
	if astSrv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(astSrv))
	}
	srv2 := &AsteriskAgent{
		RWMutex:  sync.RWMutex{},
		cfg:      cfg,
		shdChan:  shdChan,
		stopChan: nil,
		smas:     nil,
		connMgr:  nil,
		srvDep:   srvDep,
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2.smas = []*agents.AsteriskAgent{}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	astxAgent := srv2.ServiceName()
	if !reflect.DeepEqual(astxAgent, utils.AsteriskAgent) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.AsteriskAgent, astxAgent)
	}
	shouldRun := srv2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}

	srv2.stopChan = make(chan struct{}, 1)
	//no error for now
	err2 := srv2.Shutdown()
	if err2 != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err2)
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}

func TestAsteriskReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	cfg.ListenCfg().BiJSONListen = ""
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSrv, err := engine.NewService(chS)
	if err != nil {
		t.Error(err)
	}
	cacheSChan := make(chan birpc.ClientConnector, 1)
	cacheSChan <- cacheSrv
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	astSrv := NewAsteriskAgent(cfg, shdChan, nil, nil, srvDep)
	if astSrv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(astSrv))
	}
	srv2 := &AsteriskAgent{
		RWMutex:  sync.RWMutex{},
		cfg:      cfg,
		shdChan:  shdChan,
		stopChan: nil,
		smas:     nil,
		connMgr:  nil,
		srvDep:   srvDep,
	}
	srv2.stopChan = make(chan struct{}, 1)
	err3 := srv2.Reload()
	if err3 != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err3)
	}
}
