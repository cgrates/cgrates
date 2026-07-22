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
	"github.com/cgrates/cgrates/utils"
)

// TestResponderCoverage for cover testing
func TestResponderCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	server := cores.NewServer(nil)
	internalChan := make(chan birpc.ClientConnector, 1)
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan,
		shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	srv := NewResponderService(cfg, server, internalChan,
		shdChan, anz, srvDep, filterSChan)
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
	if err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrServiceAlreadyRunning, err)
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
