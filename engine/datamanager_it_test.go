//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
	dm2              *DataManager
	dataMngConfigDIR string

	sTestsDMit = []func(t *testing.T){
		testDMitDataFlush,
		testDMitCRUDStatQueue,
	}
)

func TestDMitinitDB(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	var dataDB DataDB
	var err error

	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		dataDB, err = NewRedisStorage(
			fmt.Sprintf("%s:%s", cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort),
			4, cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
			utils.REDIS_MAX_CONNS, "")
		if err != nil {
			t.Fatal("Could not connect to Redis", err.Error())
		}
	case utils.MetaMongo:
		cdrsMongoCfgPath := path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
		mgoITCfg, err := config.NewCGRConfigFromPath(cdrsMongoCfgPath)
		if err != nil {
			t.Fatal(err)
		}
		dataDB, err = NewMongoStorage(mgoITCfg.StorDbCfg().Host,
			mgoITCfg.StorDbCfg().Port, mgoITCfg.StorDbCfg().Name,
			mgoITCfg.StorDbCfg().User, mgoITCfg.StorDbCfg().Password,
			mgoITCfg.GeneralCfg().DBDataEncoding, nil, false)
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
	if err := dm2.RemoveStatQueue(sq.Tenant, sq.ID, utils.EmptyString); err != nil {
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

func TestDmRatingProfileCategory(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	rdsITdb, err := NewRedisStorage(
		"127.0.0.1:6379", 10, "", "msgpack", 4, "")
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	dm := NewDataManager(rdsITdb, cfg.CacheCfg(), nil)
	rprfs := []*RatingProfile{
		{
			Id:                    utils.ConcatenatedKey(utils.META_OUT, "cgrates.org", "*any", "1002"),
			RatingPlanActivations: RatingPlanActivations{},
		}, {
			Id:                    utils.ConcatenatedKey(utils.META_OUT, "cgrates.org", "call", "1002"),
			RatingPlanActivations: RatingPlanActivations{},
		},
		{
			Id:                    utils.ConcatenatedKey(utils.META_OUT, "cgrates.org", "*any", "1001"),
			RatingPlanActivations: RatingPlanActivations{},
		}, {
			Id:                    utils.ConcatenatedKey(utils.META_OUT, "cgrates.org", "sms", "1001"),
			RatingPlanActivations: RatingPlanActivations{},
		},
	}

	for _, rprf := range rprfs {
		if err := dm.SetRatingProfile(rprf, utils.NonTransactional); err != nil {
			t.Error(err)
		}

	}
	if err := dm.RemoveRatingProfile(rprfs[2].Id, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetRatingProfile(rprfs[1].Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetRatingProfile(rprfs[0].Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetRatingProfile(rprfs[3].Id, true, utils.NonTransactional); err != nil {
		t.Error(err)
	}
}
