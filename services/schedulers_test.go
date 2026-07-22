// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"reflect"
	"sync"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/scheduler"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestSchedulerSCoverage for cover testing
func TestSchedulerSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheActionPlans))
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	schS := NewSchedulerService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)

	if schS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	schS.schS = &scheduler.Scheduler{}
	if !schS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	serviceName := schS.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.SchedulerS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.SchedulerS, serviceName)
	}
	shouldRun := schS.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	if !reflect.DeepEqual(schS.GetScheduler(), schS.schS) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", schS.schS, schS.GetScheduler())
	}
}
