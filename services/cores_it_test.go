// +build integration

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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

//TestNewActionService for cover testing
func TestCoreSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	caps := engine.NewCaps(1, "test_caps")
	server := cores.NewServer(nil)
	internalCoreSChan := make(chan rpcclient.ClientConnector, 1)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srv := NewCoreService(cfg, caps, server,
		internalCoreSChan, anz, srvDep)
	if srv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(srv))
	}
	srv2 := NewCoreService(cfg, caps, server,
		internalCoreSChan, anz, srvDep)
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2.cS = &cores.CoreService{}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err := srv2.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	serviceName := srv2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.CoreS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.CoreS, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, true) {
		t.Errorf("\nExpecting <true>,\n Received <%+v>", shouldRun)
	}
	rld := srv2.Reload()
	if rld != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", rld)
	}
	getCoreS := srv2.GetCoreS()
	if getCoreS == nil {
		t.Errorf("\nExpecting not <nil>,\n Received <%+v>", getCoreS)
	}

}
