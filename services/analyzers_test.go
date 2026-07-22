// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/analyzers"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestAnalyzerCoverage for cover testing
func TestAnalyzerCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	connChan := make(chan birpc.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, connChan, srvDep)
	if anz == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(anz))
	}
	anz2 := &AnalyzerService{
		RWMutex:     sync.RWMutex{},
		cfg:         cfg,
		server:      server,
		filterSChan: filterSChan,
		stopChan:    make(chan struct{}, 1),
		shdChan:     shdChan,
		connChan:    connChan,
		srvDep:      srvDep,
	}
	if anz2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var rpcClientCnctr birpc.ClientConnector
	getIntrnCdc := anz2.GetInternalCodec(rpcClientCnctr, utils.EmptyString)
	if !reflect.DeepEqual(getIntrnCdc, rpcClientCnctr) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(rpcClientCnctr), utils.ToJSON(getIntrnCdc))
	}

	anz2.anz, _ = analyzers.NewAnalyzerService(cfg)
	if !anz2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := anz2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.AnalyzerS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.AnalyzerS, serviceName)
	}
	shouldRun := anz2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	getAnalyzerS := anz2.GetAnalyzerS()
	if !reflect.DeepEqual(anz2.anz, getAnalyzerS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(anz2.anz), utils.ToJSON(getAnalyzerS))
	}
	var rpcClientCnctr2 birpc.ClientConnector
	getIntrnCdc2 := anz2.GetInternalCodec(rpcClientCnctr2, utils.EmptyString)
	expected2 := anz2.anz.NewAnalyzerConnector(nil, utils.MetaInternal, utils.EmptyString, utils.EmptyString)
	if !reflect.DeepEqual(getIntrnCdc2, expected2) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expected2), utils.ToJSON(getIntrnCdc2))
	}

}
func TestAnalyzerServiceReload(t *testing.T) {
	analyzerService := &AnalyzerService{}
	err := analyzerService.Reload()
	if err != nil {
		t.Errorf("Expected Reload to return no error, got %v", err)
	}
}
