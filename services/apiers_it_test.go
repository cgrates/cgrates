//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestApiersReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan bool, 1)
	chS := engine.NewCacheS(cfg, nil)
	close(chS.GetPrecacheChannel(utils.CacheThresholdProfiles))
	close(chS.GetPrecacheChannel(utils.CacheThresholds))
	close(chS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes))

	cfg.ThresholdSCfg().Enabled = true
	cfg.SchedulerCfg().Enabled = true
	server := utils.NewServer()
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	db := NewDataDBService(cfg, nil)
	cfg.StorDbCfg().Type = utils.MetaInternal
	stordb := NewStorDBService(cfg)
	schS := NewSchedulerService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil)
	tS := NewThresholdService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1))

	apiSv1 := NewAPIerSv1Service(cfg, db, stordb, filterSChan, server, schS, new(ResponderService),
		make(chan birpc.ClientConnector, 1), nil)

	apiSv2 := NewAPIerSv2Service(apiSv1, cfg, server, make(chan birpc.ClientConnector, 1))
	srvMngr.AddServices(apiSv1, apiSv2, schS, tS,
		NewLoaderService(cfg, db, filterSChan, server, engineShutdown, make(chan birpc.ClientConnector, 1), nil), db, stordb)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if apiSv1.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if apiSv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if stordb.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	if err := cfg.V1ReloadConfigFromPath(&config.ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.ApierS,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
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
	cfg.ApierCfg().Enabled = false
	cfg.GetReloadChan(config.ApierS) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if apiSv1.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if apiSv2.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	engineShutdown <- true
}
