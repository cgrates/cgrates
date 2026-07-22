//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

	sTestsDMit = []func(t *testing.T){
		testDMitDataFlush,
		testDMitCRUDStatQueue,
	}
)

func TestDMitinitDB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	var dataDB DataDB
	var err error

	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		dataDB, err = NewRedisStorage(
			fmt.Sprintf("%s:%s", cfg.DataDbCfg().Host, cfg.DataDbCfg().Port),
			4, cfg.DataDbCfg().User, cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
			10, 20, "", false, 0, 0, 0, 150*time.Microsecond, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
		if err != nil {
			t.Fatal("Could not connect to Redis", err.Error())
		}
	case utils.MetaMongo:
		cdrsMongoCfgPath := path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
		mgoITCfg, err := config.NewCGRConfigFromPath(cdrsMongoCfgPath)
		if err != nil {
			t.Fatal(err)
		}
		dataDB, err = NewMongoStorage("mongodb", mgoITCfg.StorDbCfg().Host,
			mgoITCfg.StorDbCfg().Port, mgoITCfg.StorDbCfg().Name,
			mgoITCfg.StorDbCfg().User, mgoITCfg.StorDbCfg().Password,
			mgoITCfg.GeneralCfg().DBDataEncoding,
			utils.StorDB, nil, 10*time.Second)
		if err != nil {
			t.Fatal("Could not connect to Mongo", err.Error())
		}
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	dm2 = NewDataManager(dataDB, config.CgrConfig().CacheCfg(), nil)

	for _, stest := range sTestsDMit {
		t.Run(*utils.DBType, stest)
	}
}

func testDMitDataFlush(t *testing.T) {
	if err := dm2.dataDB.Flush(utils.EmptyString); err != nil {
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
					"cgrates.org:ev1": {Stat: 1},
					"cgrates.org:ev2": {Stat: 1},
					"cgrates.org:ev3": {Stat: 0},
				},
			},
		},
	}
	if _, rcvErr := dm2.GetStatQueue(sq.Tenant, sq.ID, true, false, utils.EmptyString); rcvErr != utils.ErrNotFound {
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
	if rcv, err := dm2.GetStatQueue(sq.Tenant, sq.ID, true, true, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("expecting: %v, received: %v", sq, rcv)
	}
	if _, ok := Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != true {
		t.Error("should be in cache")
	}
	if err := dm2.RemoveStatQueue(sq.Tenant, sq.ID); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if _, ok := Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != false {
		t.Error("should not be in cache")
	}
	if _, rcvErr := dm2.GetStatQueue(sq.Tenant, sq.ID, true, false, utils.EmptyString); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}
