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

package engine

import (
	"fmt"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	dm2 *DataManager
)

var sTestsDMit = []func(t *testing.T){
	testDMitDataFlush,
}

func TestDMitRedis(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	dataDB, err := NewRedisStorage(
		fmt.Sprintf("%s:%s", cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort),
		4, cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
		utils.REDIS_MAX_CONNS, nil, "")
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	dm2 = NewDataManager(dataDB)
	for _, stest := range sTestsDMit {
		t.Run("TestDMitRedis", stest)
	}
}

func TestDMitMongo(t *testing.T) {
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "cdrsv2mongo")
	mgoITCfg, err := config.NewCGRConfigFromPath(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := NewMongoStorage(mgoITCfg.StorDbCfg().StorDBHost,
		mgoITCfg.StorDbCfg().StorDBPort, mgoITCfg.StorDbCfg().StorDBName,
		mgoITCfg.StorDbCfg().StorDBUser, mgoITCfg.StorDbCfg().StorDBPass,
		utils.StorDB, nil, mgoITCfg.CacheCfg(), false)
	if err != nil {
		t.Fatal("Could not connect to Mongo", err.Error())
	}
	dm2 = NewDataManager(dataDB)
	for _, stest := range sTestsDMit {
		t.Run("TestDMitMongo", stest)
	}
}

func testDMitDataFlush(t *testing.T) {
	if err := dm2.dataDB.Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
}

func testDMitCRUDStatQueue(t *testing.T) {
	eTime := utils.TimePointer(time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC))
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "testDMitCRUDStatQueue",
		SQItems: []SQItem{
			{EventID: "cgrates.org:ev1", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev2", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev3", ExpiryTime: eTime},
		},
		SQMetrics: map[string]StatMetric{
			utils.MetaASR: &StatASR{
				Answered: 2,
				Count:    3,
				Events: map[string]*StatWithCompress{
					"cgrates.org:ev1": &StatWithCompress{Stat: 1},
					"cgrates.org:ev2": &StatWithCompress{Stat: 1},
					"cgrates.org:ev3": &StatWithCompress{Stat: 0},
				},
			},
		},
	}
	if _, rcvErr := dm2.GetStatQueue(sq.Tenant, sq.ID, true, false, ""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if _, ok := Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != false {
		t.Error("should not be in cache")
	}
	if err := dm2.SetStatQueue(sq); err != nil {
		t.Error(err)
	}
	if _, ok := Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != false {
		t.Error("should not be in cache")
	}
	if rcv, err := dm2.GetStatQueue(sq.Tenant, sq.ID, true, false, ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("expecting: %v, received: %v", sq, rcv)
	}
	if _, ok := Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != true {
		t.Error("should be in cache")
	}
	if err := dm2.RemoveStatQueue(sq.Tenant, sq.ID, ""); err != nil {
		t.Error(err)
	}
	if _, ok := Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != false {
		t.Error("should not be in cache")
	}
	if _, rcvErr := dm2.GetStatQueue(sq.Tenant, sq.ID, true, false, ""); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}
