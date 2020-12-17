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

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/cores"

	"github.com/cgrates/cgrates/config"
)

//TestResponderCoverage for cover testing
func TestResponderCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	server := cores.NewServer(nil)
	internalChan := make(chan rpcclient.ClientConnector, 1)
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan,
		shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	srv := NewResponderService(cfg, server, internalChan,
		shdChan, anz, srvDep)
	if srv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(srv))
	}
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv.resp = &engine.Responder{}
	if !srv.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err := srv.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	serviceName := srv.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.ResponderS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ResponderS, serviceName)
	}
	getResponder := srv.GetResponder()
	if !reflect.DeepEqual(getResponder, srv.resp) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", srv.resp, getResponder)
	}
	shouldRun := srv.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, shouldRun)
	}
}
