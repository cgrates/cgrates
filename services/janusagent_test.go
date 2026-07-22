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

func TestJanusAgentCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	server := cores.NewServer(nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	rpcInternal := map[string]chan birpc.ClientConnector{}
	cM := engine.NewConnManager(cfg, rpcInternal)
	srv := NewJanusAgent(cfg, filterSChan, server, cM, nil, srvDep)
	if srv == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(srv))
	}
	if srv.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	srv2 := &JanusAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		server:      server,
		started:     true,
		connMgr:     cM,
		srvDep:      srvDep,
	}
	if !srv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := srv2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.JanusAgent) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.JanusAgent, serviceName)
	}
}

func TestJanusAgentReload(t *testing.T) {
	ja := &JanusAgent{}
	err := ja.Reload()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
