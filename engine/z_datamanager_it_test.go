//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	dm2      *DataManager
	dm2Cache *CacheS

	sTestsDMit = []func(t *testing.T){
		testDMitDataFlush,
		testDMitCRUDStatQueue,
	}
)

func TestDMitinitDB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := NewGuardianLocker(cfg)
	var dataDB DataDB
	var err error

	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaRedis:
		dataDB, err = NewRedisStorage("127.0.0.1:6379", 4, utils.CGRateSLwr,
			cfg.DbCfg().DBConns[utils.MetaDefault].Password,
			cfg.GeneralCfg().DBDataEncoding,
			cfg.DbCfg().DBConns[utils.MetaDefault].Opts.RedisMaxConns,
			cfg.DbCfg().DBConns[utils.MetaDefault].Opts.RedisConnectAttempts, "", false,
			0, 0, 0, 0, 0, 150*time.Microsecond, 0, false,
			utils.EmptyString, utils.EmptyString,
			utils.EmptyString, 1000, nil, nil)
		if err != nil {
			t.Fatal("Could not connect to Redis", err.Error())
		}
	case utils.MetaMySQL:
		t.SkipNow()
	case utils.MetaPostgres, utils.MetaMongo:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	dbCM := NewDBConnManager(map[string]DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm2 = NewDataManager(dbCM, cfg, nil, locker)
	dm2Cache = NewCacheS(cfg, nil, nil, nil, locker)
	dm2.SetCache(dm2Cache)

	for _, stest := range sTestsDMit {
		t.Run(*utils.DBType, stest)
	}
}

func testDMitDataFlush(t *testing.T) {
	if err := dm2.dbConns.dbs[utils.MetaDefault].Flush(utils.EmptyString); err != nil {
		t.Error(err)
	}
	dm2Cache.Clear(nil)
}

func testDMitCRUDStatQueue(t *testing.T) {
	eTime := utils.TimePointer(time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC))
	sq := &utils.StatQueue{
		Tenant: "cgrates.org",
		ID:     "testDMitCRUDStatQueue",
		SQItems: []utils.SQItem{
			{EventID: "cgrates.org:ev1", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev2", ExpiryTime: eTime},
			{EventID: "cgrates.org:ev3", ExpiryTime: eTime},
		},
		SQMetrics: map[string]utils.StatMetric{
			utils.MetaASR: &utils.StatASR{
				Metric: &utils.Metric{
					Value: utils.NewDecimal(2, 0),
					Count: 3,
					Events: map[string]*utils.DecimalWithCompress{
						"cgrates.org:ev1": {Stat: utils.NewDecimal(1, 0)},
						"cgrates.org:ev2": {Stat: utils.NewDecimal(1, 0)},
						"cgrates.org:ev3": {Stat: utils.NewDecimal(0, 0)},
					},
				},
			},
		},
	}
	if _, rcvErr := dm2.GetStatQueue(context.TODO(), sq.Tenant, sq.ID, true, false, utils.EmptyString); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if _, ok := dm2Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != false {
		t.Error("should not be in cache")
	}
	if err := dm2.SetStatQueue(context.TODO(), sq); err != nil {
		t.Error(err)
	}
	if _, ok := dm2Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != false {
		t.Error("should not be in cache")
	}
	if rcv, err := dm2.GetStatQueue(context.TODO(), sq.Tenant, sq.ID, true, true, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("expecting: %v, received: %v", sq, rcv)
	}
	if _, ok := dm2Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != true {
		t.Error("should be in cache")
	}
	if err := dm2.RemoveStatQueue(context.TODO(), sq.Tenant, sq.ID); err != nil {
		t.Error(err)
	}
	dm2Cache.Clear(nil)
	if _, ok := dm2Cache.Get(utils.CacheStatQueues, sq.TenantID()); ok != false {
		t.Error("should not be in cache")
	}
	if _, rcvErr := dm2.GetStatQueue(context.TODO(), sq.Tenant, sq.ID, true, false, utils.EmptyString); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}
