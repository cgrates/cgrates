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

func TestResourceSReload(t *testing.T) {
	// utils.Logger.SetLogLevel(7)
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.ThresholdSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan bool, 1)
	chS := engine.NewCacheS(cfg, nil)
	close(chS.GetPrecacheChannel(utils.CacheThresholdProfiles))
	close(chS.GetPrecacheChannel(utils.CacheThresholds))
	close(chS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes))
	close(chS.GetPrecacheChannel(utils.CacheResourceProfiles))
	close(chS.GetPrecacheChannel(utils.CacheResources))
	close(chS.GetPrecacheChannel(utils.CacheResourceFilterIndexes))
	server := utils.NewServer()
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	db := NewDataDBService(cfg, nil)
	tS := NewThresholdService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1))
	reS := NewResourceService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(tS, reS,
		NewLoaderService(cfg, db, filterSChan, server, engineShutdown, make(chan birpc.ClientConnector, 1), nil), db)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if reS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	var reply string
	if err = cfg.V1ReloadConfigFromPath(&config.ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.RESOURCES_JSON,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !reS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.ResourceSCfg().Enabled = false
	cfg.GetReloadChan(config.RESOURCES_JSON) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if reS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	engineShutdown <- true
}
