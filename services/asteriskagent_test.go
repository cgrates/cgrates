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
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

//TestAsteriskAgentCoverage for cover testing
func TestAsteriskAgentCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.SessionSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	cacheSChan := make(chan rpcclient.ClientConnector, 1)
	cacheSChan <- chS
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	srv := NewAsteriskAgent(cfg, shdChan, nil, srvDep)
	if srv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(srv))
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
	err := srv2.Start()
	if err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
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
	err2 := srv2.Reload()
	if err2 != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err2)
	}
	err2 = srv2.Shutdown()
	if err2 != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err2)
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
