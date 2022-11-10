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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestDataDBCoverage for cover testing
func TestDataDBCoverage(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	//chS := engine.NewCacheS(cfg, nil, nil)
	filterSChan := make(chan *engine.FilterS, 1)
	filterSChan <- nil
	srvDep := map[string]*sync.WaitGroup{utils.DataDB: new(sync.WaitGroup)}
	cM := engine.NewConnManager(cfg, nil)
	db := NewDataDBService(cfg, cM, srvDep)
	if db.IsRunning() {
		t.Errorf("Expected service to be down")
	}
	//populates dataDb with something in order to call the close function
	dataDb := new(engine.RedisStorage)
	db.dm = engine.NewDataManager(dataDb,
		&config.CacheCfg{}, nil)
	if !db.IsRunning() {
		t.Errorf("Expected service to be running")
	}
	oldcfg := &config.DataDbCfg{
		Type: utils.Mongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			RedisMaxConns:           10,
			RedisConnectAttempts:    20,
			RedisSentinel:           "",
			RedisCluster:            false,
			RedisClusterSync:        5 * time.Second,
			RedisClusterOndownDelay: 0,
			RedisConnectTimeout:     0,
			RedisReadTimeout:        0,
			RedisWriteTimeout:       0,
			MongoQueryTimeout:       10 * time.Second,
			RedisTLS:                false,
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
		},
	}
	db.oldDBCfg = oldcfg
	serviceName := db.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.DataDB) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.DataDB, serviceName)
	}
	shouldRun := db.ShouldRun()
	if !reflect.DeepEqual(shouldRun, false) {
		t.Errorf("\nExpecting <false>,\n Received <%+v>", shouldRun)
	}
	getDMChan := db.GetDMChan()
	if !reflect.DeepEqual(getDMChan, db.dbchan) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", db.dbchan, getDMChan)
	}
	db.dm = &engine.DataManager{}
	getDM := db.GetDM()
	if !reflect.DeepEqual(getDM, db.dm) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", db.dm, getDM)
	}

}
