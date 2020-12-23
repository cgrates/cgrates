/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package services

import (
	"sync"
	"testing"

	"github.com/cgrates/cgrates/agents"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

//TestDiameterAgentCoverage for cover testing
func TestDiameterAgentCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewDiameterAgent(cfg, filterSChan, shdChan, nil, srvDep)

	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	rpcInternal := map[string]chan rpcclient.ClientConnector{}
	cM := engine.NewConnManager(cfg, rpcInternal)
	filterS := engine.NewFilterS(cfg, cM, nil)
	da, _ := agents.NewDiameterAgent(cfg, filterS, cM)
	srv2 := &DiameterAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		connMgr:     nil,
		srvDep:      srvDep,
		stopChan:    make(chan struct{}, 1),
		da:          da,
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv2.ServiceName()
	if serviceName != utils.DiameterAgent {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.DiameterAgent, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if shouldRun != false {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}
	shutErr := srv2.Shutdown()
	if shutErr != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", shutErr)
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}

}
