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

func TestSupplierSReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
	cfg.StatSCfg().Enabled = true
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	engineShutdown := make(chan bool, 1)
	chS := engine.NewCacheS(cfg, nil)
	close(chS.GetPrecacheChannel(utils.CacheSupplierProfiles))
	close(chS.GetPrecacheChannel(utils.CacheSupplierFilterIndexes))
	close(chS.GetPrecacheChannel(utils.CacheStatQueueProfiles))
	close(chS.GetPrecacheChannel(utils.CacheStatQueues))
	close(chS.GetPrecacheChannel(utils.CacheStatFilterIndexes))
	server := utils.NewServer()
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	db := NewDataDBService(cfg, nil)
	sts := NewStatService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil)
	supS := NewSupplierService(cfg, db, chS, filterSChan, server, make(chan birpc.ClientConnector, 1), nil)
	engine.NewConnManager(cfg, nil)
	srvMngr.AddServices(supS, sts,
		NewLoaderService(cfg, db, filterSChan, server, engineShutdown, make(chan birpc.ClientConnector, 1), nil), db)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if supS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	var reply string
	if err := cfg.V1ReloadConfigFromPath(&config.ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongonew"),
		Section: config.SupplierSJson,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !supS.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	cfg.SupplierSCfg().Enabled = false
	cfg.GetReloadChan(config.SupplierSJson) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if supS.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	engineShutdown <- true
}
