//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"path"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestApiersReload(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	shdChan := utils.NewSyncedChan()
	shdWg := new(sync.WaitGroup)
	chS := engine.NewCacheS(cfg, nil, nil)
	close(chS.GetPrecacheChannel(utils.CacheThresholdProfiles))
	close(chS.GetPrecacheChannel(utils.CacheThresholds))
	close(chS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes))
	close(chS.GetPrecacheChannel(utils.CacheActionPlans))

	cfg.ThresholdSCfg().Enabled = true
	cfg.SchedulerCfg().Enabled = true
	server := cores.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, shdChan, shdWg, nil)
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	db := NewDataDBService(cfg, nil, false, srvDep)
	cfg.StorDbCfg().Type = utils.MetaInternal
	stordb := NewStorDBService(cfg, false, srvDep)
	anz := NewAnalyzerService(cfg, server, filterSChan, shdChan, make(chan birpc.ClientConnector, 1), srvDep)
	schS := NewSchedulerService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)
	tS := NewThresholdService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil, anz, srvDep)
	rspd := NewResponderService(cfg, server, make(chan birpc.ClientConnector, 1), shdChan, anz, srvDep, filterSChan)
	apiSv1 := NewAPIerSv1Service(cfg, db, stordb, filterSChan, server, schS, rspd,
		make(chan birpc.ClientConnector, 1), nil, anz, srvDep)

	apiSv2 := NewAPIerSv2Service(apiSv1, cfg, server, make(chan birpc.ClientConnector, 1), anz, srvDep)
	srvMngr.AddServices(apiSv1, apiSv2, schS, tS, db, stordb)
	if err := srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if apiSv1.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if apiSv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if stordb.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfig(context.Background(),
		&config.ReloadArgs{
			Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
			Section: config.ApierS,
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(100 * time.Millisecond) //need to switch to gorutine
	if !apiSv1.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !apiSv2.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !stordb.IsRunning() {
		t.Errorf("Expected service to be running")
	}

	err := apiSv1.Start()
	if err == nil || err != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err)
	}
	err2 := apiSv2.Start()
	if err2 == nil || err2 != utils.ErrServiceAlreadyRunning {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ErrServiceAlreadyRunning, err2)
	}
	err = apiSv1.Reload()
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	err2 = apiSv2.Reload()
	if err2 != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err2)
	}
	expected := &v1.APIerSv1{}
	getAPIerSv1 := apiSv1.GetAPIerSv1()
	if reflect.DeepEqual(expected, getAPIerSv1) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(getAPIerSv1))
	}
	cfg.ApierCfg().Enabled = false
	cfg.GetReloadChan(config.ApierS) <- struct{}{}
	time.Sleep(100 * time.Millisecond)
	if apiSv1.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if apiSv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	shdChan.CloseOnce()
	time.Sleep(10 * time.Millisecond)
}
