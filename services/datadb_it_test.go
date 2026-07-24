//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func TestDataDBReload(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
	utils.Newlogger(utils.MetaSysLog, cfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)

	engineShutdown := make(chan bool, 1)
	chS := engine.NewCacheS(cfg, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	close(chS.GetPrecacheChannel(utils.CacheAttributeProfiles))
	close(chS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes))
	server := utils.NewServer()
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM)
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan birpc.ClientConnector, 1)),
		NewLoaderService(cfg, db, filterSChan, server, engineShutdown, make(chan birpc.ClientConnector, 1), nil), db)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.AttributeSCfg().Enabled = true
	if err := cfg.V1ReloadConfigFromPath(&config.ConfigReloadWithArgDispatcher{
		Path:    path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo"),
		Section: config.DATADB_JSN,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting OK ,received %s", reply)
	}
	time.Sleep(10 * time.Millisecond) //need to switch to gorutine
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	oldcfg := &config.DataDbCfg{
		DataDbType:   utils.MetaMongo,
		DataDbHost:   "127.0.0.1",
		DataDbPort:   "27017",
		DataDbName:   "10",
		DataDbUser:   "cgrates",
		QueryTimeout: 10 * time.Second,
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaReverseDestinations: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaDestinations: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaRatingPlans: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaRatingProfiles: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaActions: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaActionPlans: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaAccountActionPlans: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaActionTriggers: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaSharedGroups: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaTimings: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaResourceProfile: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaStatQueues: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaResources: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaStatQueueProfiles: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaThresholds: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaThresholdProfiles: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaFilters: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaSupplierProfiles: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaAttributeProfiles: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaDispatcherHosts: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaChargerProfiles: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaDispatcherProfiles: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaFilterIndexes: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
			utils.MetaLoadIDs: {
				Limit:     -1,
				Replicate: false,
				Remote:    false,
				TTL:       time.Duration(0)},
		},
	}
	if !reflect.DeepEqual(oldcfg, db.oldDBCfg) {
		t.Errorf("Expected %s \n received:%s", utils.ToJSON(oldcfg), utils.ToJSON(db.oldDBCfg))
	}
	cfg.AttributeSCfg().Enabled = false
	cfg.GetReloadChan(config.DATADB_JSN) <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	engineShutdown <- true
}
