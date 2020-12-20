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

	"github.com/cgrates/cgrates/dispatcherh"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

//TestDispatcherCoverage for cover testing
func TestDispatcherHCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	rpcInternal := map[string]chan rpcclient.ClientConnector{}
	cM := engine.NewConnManager(cfg, rpcInternal)
	srv := NewDispatcherHostsService(cfg, server, cM, anz, srvDep)
	if srv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(srv))
	}
	srv2 := &DispatcherHostsService{
		cfg:     cfg,
		server:  server,
		connMgr: cM,
		anz:     anz,
		srvDep:  srvDep,
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2.dspS = &dispatcherh.DispatcherHostsService{}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err := srv2.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	rld := srv2.Reload()
	if rld != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", rld)
	}
	serviceName := srv2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.DispatcherH) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.DispatcherH, serviceName)
	}
	shouldRun := srv2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	srv2.stopChan = make(chan struct{}, 1)
	srv2.dspS = dispatcherh.NewDispatcherHService(cfg, cM)
	shutdownSrv := srv2.Shutdown()
	if shutdownSrv != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", shutdownSrv)
	}
	if srv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
}
