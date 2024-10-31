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
	cM := engine.NewConnManager(cfg)
	db := NewDataDBService(cfg, cM, false, srvDep)
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
		Type: utils.MetaMongo,
		Host: "127.0.0.1",
		Port: "27017",
		Name: "10",
		User: "cgrates",
		Opts: &config.DataDBOpts{
			MongoQueryTimeout: 10 * time.Second,
			RedisClusterSync:  5 * time.Second,
		},
		RmtConns: []string{},
		RplConns: []string{},
		Items: map[string]*config.ItemOpts{
			utils.MetaAccounts:           {},
			utils.MetaActions:            {},
			utils.MetaCronExp:            {},
			utils.MetaResourceProfile:    {},
			utils.MetaStatQueues:         {},
			utils.MetaResources:          {},
			utils.MetaStatQueueProfiles:  {},
			utils.MetaThresholds:         {},
			utils.MetaThresholdProfiles:  {},
			utils.MetaFilters:            {},
			utils.MetaRouteProfiles:      {},
			utils.MetaAttributeProfiles:  {},
			utils.MetaDispatcherHosts:    {},
			utils.MetaChargerProfiles:    {},
			utils.MetaDispatcherProfiles: {},
			utils.MetaLoadIDs:            {},
			utils.MetaRateProfiles:       {},
			utils.MetaActionProfiles:     {},
		},
	}
	db.oldDBCfg = oldcfg
	serviceName := db.ServiceName()
	if !reflect.DeepEqual(serviceName, utils.DataDB) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.DataDB, serviceName)
	}
	if shouldRun := db.ShouldRun(); !shouldRun {
		t.Errorf("\nExpecting <true>,\n Received <%+v>", shouldRun)
	}
}
