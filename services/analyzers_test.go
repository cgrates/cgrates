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

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/analyzers"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//TestAnalyzerCoverage for cover testing
func TestAnalyzerCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	connChan := make(chan birpc.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, connChan, srvDep)
	if anz == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(anz))
	}
	anz2 := &AnalyzerService{
		RWMutex:     sync.RWMutex{},
		cfg:         cfg,
		server:      server,
		filterSChan: filterSChan,
		stopChan:    make(chan struct{}, 1),
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
