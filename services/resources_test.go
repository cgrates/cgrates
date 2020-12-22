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

//TestResourceSCoverage for cover testing
func TestResourceSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan rpcclient.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, srvDep)
	reS := NewResourceService(cfg, db, chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), nil, anz, srvDep)

	if reS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	reS2 := ResourceService{
		cfg:         cfg,
		dm:          db,
		cacheS:      chS,
		filterSChan: filterSChan,
		server:      server,
		connChan:    make(chan rpcclient.ClientConnector, 1),
		connMgr:     nil,
		anz:         anz,
		srvDep:      srvDep,
		reS:         &engine.ResourceService{},
	}
	if !reS2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	serviceName := reS2.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.ResourceS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ResourceS, serviceName)
	}
	shouldRun := reS2.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
}
