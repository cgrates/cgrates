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

// TestResourceSCoverage for cover testing
func TestResourceSCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ThresholdSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	chS := engine.NewCacheS(cfg, nil, nil)
	server := cores.NewServer(nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	db := NewDataDBService(cfg, nil, false, srvDep)
	reS := NewResourceService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)

	if reS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	reS2 := ResourceService{
		cfg:         cfg,
		dm:          db,
		cacheS:      chS,
		filterSChan: filterSChan,
		server:      server,
		connChan:    make(chan birpc.ClientConnector, 1),
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
