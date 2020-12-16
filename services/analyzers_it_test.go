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

	"github.com/cgrates/cgrates/analyzers"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

//TestNewActionService for cover testing
func TestNewAnalyzerCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	shdChan := utils.NewSyncedChan()
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	connChan := make(chan rpcclient.ClientConnector, 1)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, connChan, srvDep)
	if anz == nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", utils.ToJSON(anz))
	}
	anz2 := &AnalyzerService{
		connChan:    connChan,
		cfg:         cfg,
		server:      server,
		filterSChan: filterSChan,
		shdChan:     shdChan,
		srvDep:      srvDep,
	}
	if anz2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	anz2.anz = &analyzers.AnalyzerService{}
	if !anz2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	err := anz2.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
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
}
