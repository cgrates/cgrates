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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/rpcclient"

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
	server := utils.NewServer(nil)
	srvMngr := servmanager.NewServiceManager(cfg, engineShutdown)
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM)
	anz := NewAnalyzerService(cfg, server, engineShutdown, make(chan rpcclient.ClientConnector, 1))
	srvMngr.AddServices(NewAttributeService(cfg, db,
		chS, filterSChan, server, make(chan rpcclient.ClientConnector, 1), anz),
		NewLoaderService(cfg, db, filterSChan, server, engineShutdown, make(chan rpcclient.ClientConnector, 1), nil, anz), db)
	if err = srvMngr.StartServices(); err != nil {
		t.Error(err)
	}
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	var reply string
	cfg.AttributeSCfg().Enabled = true
	if err := cfg.V1ReloadConfigFromPath(&config.ConfigReloadWithOpts{
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
		DataDbType: utils.MONGO,
		DataDbHost: "127.0.0.1",
		DataDbPort: "27017",
		DataDbName: "10",
		DataDbUser: "cgrates",
		Opts: map[string]interface{}{
			utils.QueryTimeoutCfg:            "10s",
			utils.RedisClusterOnDownDelayCfg: "0",
			utils.RedisClusterSyncCfg:        "5s",
			utils.RedisClusterCfg:            false,
			utils.RedisSentinelNameCfg:       "",
			utils.RedisTLS:                   false,
			utils.RedisClientCertificate:     "",
			utils.RedisClientKey:             "",
			utils.RedisCACertificate:         "",
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpt{
			utils.MetaAccounts: {
				Replicate: false,
				Remote:    false},
			utils.MetaReverseDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaDestinations: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaRatingProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaActions: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaAccountActionPlans: {
				Replicate: false,
				Remote:    false},
			utils.MetaActionTriggers: {
				Replicate: false,
				Remote:    false},
			utils.MetaSharedGroups: {
				Replicate: false,
				Remote:    false},
			utils.MetaTimings: {
				Replicate: false,
				Remote:    false},
			utils.MetaResourceProfile: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueues: {
				Replicate: false,
				Remote:    false},
			utils.MetaResources: {
				Replicate: false,
				Remote:    false},
			utils.MetaStatQueueProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholds: {
				Replicate: false,
				Remote:    false},
			utils.MetaThresholdProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaFilters: {
				Replicate: false,
				Remote:    false},
			utils.MetaRouteProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaAttributeProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherHosts: {
				Replicate: false,
				Remote:    false},
			utils.MetaChargerProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaDispatcherProfiles: {
				Replicate: false,
				Remote:    false},
			utils.MetaLoadIDs: {
				Replicate: false,
				Remote:    false},
			utils.MetaIndexes: {
				Replicate: false,
				Remote:    false},
			utils.MetaRateProfiles: {
				Replicate: false,
				Remote:    false},
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
