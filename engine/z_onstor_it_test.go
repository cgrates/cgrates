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

	"github.com/ericlagergren/decimal"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	rdsITdb   *RedisStorage
	mgoITdb   *MongoStorage
	onStor    *DataManager
	onStorCfg string

	// subtests to be executed for each confDIR
	sTestsOnStorIT = []func(t *testing.T){
		testOnStorITFlush,
		testOnStorITIsDBEmpty,
		// ToDo: test cache flush for a prefix
		// ToDo: testOnStorITLoadAccountingCache
		testOnStorITResource,
		testOnStorITResourceProfile,
		//testOnStorITCRUDHistory,
		testOnStorITCRUDStructVersion,
		testOnStorITStatQueueProfile,
		testOnStorITStatQueue,
		testOnStorITThresholdProfile,
		testOnStorITThreshold,
		testOnStorITFilter,
		testOnStorITRouteProfile,
		testOnStorITAttributeProfile,
		testOnStorITFlush,
		testOnStorITIsDBEmpty,
		testOnStorITTestAttributeSubstituteIface,
		testOnStorITChargerProfile,
		testOnStorITDispatcherProfile,
		testOnStorITRateProfile,
		testOnStorITActionProfile,
		testOnStorITAccount,
		//testOnStorITCacheActionTriggers,
		//testOnStorITCRUDActionTriggers,
	}
)

func TestOnStorIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		onStor = NewDataManager(NewInternalDB(nil, nil, true),
			config.CgrConfig().CacheCfg(), nil)
	case utils.MetaMySQL:
		cfg := config.NewDefaultCGRConfig()
		rdsITdb, err = NewRedisStorage(
			fmt.Sprintf("%s:%s", cfg.DataDbCfg().Host, cfg.DataDbCfg().Port),
			4, cfg.DataDbCfg().User, cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
			utils.RedisMaxConns, "", false, 0, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
		if err != nil {
			t.Fatal("Could not connect to Redis", err.Error())
		}
		onStorCfg = cfg.DataDbCfg().Name
		onStor = NewDataManager(rdsITdb, config.CgrConfig().CacheCfg(), nil)
	case utils.MetaMongo:
		cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "cdrsv2mongo")
		mgoITCfg, err := config.NewCGRConfigFromPath(cdrsMongoCfgPath)
		if err != nil {
			t.Fatal(err)
		}
		if mgoITdb, err = NewMongoStorage(mgoITCfg.StorDbCfg().Host,
			mgoITCfg.StorDbCfg().Port, mgoITCfg.StorDbCfg().Name,
			mgoITCfg.StorDbCfg().User, mgoITCfg.StorDbCfg().Password,
			mgoITCfg.GeneralCfg().DBDataEncoding,
			utils.StorDB, nil, 10*time.Second); err != nil {
			t.Fatal(err)
		}
		onStorCfg = mgoITCfg.StorDbCfg().Name
		onStor = NewDataManager(mgoITdb, config.CgrConfig().CacheCfg(), nil)
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsOnStorIT {
		t.Run(*dbType, stest)
	}
}

func testOnStorITFlush(t *testing.T) {
	if err := onStor.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
}

func testOnStorITIsDBEmpty(t *testing.T) {
	test, err := onStor.DataDB().IsDBEmpty()
	if err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
}

func testOnStorITResourceProfile(t *testing.T) {
	rL := &ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RL_TEST2",
		Weight:       10,
		FilterIDs:    []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-03T13:43:00Z|2014-07-03T13:44:00Z"},
		Limit:        1,
		ThresholdIDs: []string{"TEST_ACTIONS"},
		UsageTTL:     3 * time.Nanosecond,
	}
	if _, rcvErr := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetResourceProfile(rL, false); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		false, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(rL), utils.ToJSON(rcv))
	}
	expectedR := []string{"rsp_cgrates.org:RL_TEST2"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.ResourceProfilesPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	//update
	rL.ThresholdIDs = []string{"TH1", "TH2"}
	if err := onStor.SetResourceProfile(rL, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rL, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(rL), utils.ToJSON(rcv))
	}

	if err := onStor.RemoveResourceProfile(rL.Tenant, rL.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetResourceProfile(rL.Tenant, rL.ID,
		false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITResource(t *testing.T) {
	res := &Resource{
		Tenant: "cgrates.org",
		ID:     "RL1",
		Usages: map[string]*ResourceUsage{
			"RU1": {
				ID:         "RU1",
				ExpiryTime: time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
				Units:      2,
			},
		},
		TTLIdx: []string{"RU1"},
	}
	if _, rcvErr := onStor.GetResource(res.Tenant, res.ID,
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetResource(res, nil, 0, true); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetResource("cgrates.org", "RL1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(res), utils.ToJSON(rcv))
	}
	expectedT := []string{"res_cgrates.org:RL1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.ResourcesPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	res.TTLIdx = []string{"RU1", "RU2"}
	if err := onStor.SetResource(res, nil, 0, true); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetResource("cgrates.org", "RL1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(res, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(res), utils.ToJSON(rcv))
	}

	if err := onStor.RemoveResource(res.Tenant, res.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetResource(res.Tenant, res.ID,
		false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITCRUDHistory(t *testing.T) {
	time := time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC)
	ist := &utils.LoadInstance{
		LoadID:           "Load",
		RatingLoadID:     "RatingLoad",
		AccountingLoadID: "Account",
		LoadTime:         time,
	}
	if err := onStor.DataDB().AddLoadHistory(ist, 1, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetLoadHistory(1, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ist, rcv[0]) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(ist), utils.ToJSON(rcv[0]))
	}
}

func testOnStorITCRUDStructVersion(t *testing.T) {
	if _, err := onStor.DataDB().GetVersions(utils.Accounts); err != utils.ErrNotFound {
		t.Error(err)
	}
	vrs := Versions{
		utils.Accounts:    3,
		utils.Actions:     2,
		utils.CostDetails: 1,
	}
	if err := onStor.DataDB().SetVersions(vrs, false); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	delete(vrs, utils.Actions)
	if err := onStor.DataDB().SetVersions(vrs, true); err != nil { // overwrite
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	eAcnts := Versions{utils.Accounts: vrs[utils.Accounts]}
	if rcv, err := onStor.DataDB().GetVersions(utils.Accounts); err != nil { //query one element
		t.Error(err)
	} else if !reflect.DeepEqual(eAcnts, rcv) {
		t.Errorf("Expecting: %v, received: %v", eAcnts, rcv)
	}
	if _, err := onStor.DataDB().GetVersions(utils.NotAvailable); err != utils.ErrNotFound { //query non-existent
		t.Error(err)
	}
	eAcnts[utils.Accounts] = 2
	vrs[utils.Accounts] = eAcnts[utils.Accounts]
	if err := onStor.DataDB().SetVersions(eAcnts, false); err != nil { // change one element
		t.Error(err)
	}
	if rcv, err := onStor.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	if err = onStor.DataDB().RemoveVersions(eAcnts); err != nil { // remove one element
		t.Error(err)
	}
	delete(vrs, utils.Accounts)
	if rcv, err := onStor.DataDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	if err = onStor.DataDB().RemoveVersions(nil); err != nil { // remove one element
		t.Error(err)
	}
	if _, err := onStor.DataDB().GetVersions(""); err != utils.ErrNotFound { //query non-existent
		t.Error(err)
	}
}

func testOnStorITStatQueueProfile(t *testing.T) {
	sq := &StatQueueProfile{
		Tenant:       "cgrates.org",
		ID:           "test",
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		QueueLength:  2,
		TTL:          0,
		Stored:       true,
		ThresholdIDs: []string{"Thresh1"},
	}
	if _, rcvErr := onStor.GetStatQueueProfile(sq.Tenant, sq.ID,
		true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetStatQueueProfile(sq, false); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
	}
	expectedR := []string{"sqp_cgrates.org:test"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.StatQueueProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	//update
	sq.ThresholdIDs = []string{"TH1", "TH2"}
	if err := onStor.SetStatQueueProfile(sq, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveStatQueueProfile(sq.Tenant, sq.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetStatQueueProfile(sq.Tenant,
		sq.ID, true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITStatQueue(t *testing.T) {
	eTime := utils.TimePointer(time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC))
	sq := &StatQueue{
		Tenant: "cgrates.org",
		ID:     "Test_StatQueue",
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
	if _, rcvErr := onStor.GetStatQueue(sq.Tenant, sq.ID,
		true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetStatQueue(sq, nil, 0, nil, 0, true); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetStatQueue(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
	}
	expectedT := []string{"stq_cgrates.org:Test_StatQueue"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.StatQueuePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	sq.SQMetrics = map[string]StatMetric{
		utils.MetaASR: &StatASR{
			Answered: 3,
			Count:    3,
			Events: map[string]*StatWithCompress{
				"cgrates.org:ev1": {Stat: 1},
				"cgrates.org:ev2": {Stat: 1},
				"cgrates.org:ev3": {Stat: 1},
			},
		},
	}
	if err := onStor.SetStatQueue(sq, nil, 0, nil, 0, true); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetStatQueue(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sq, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(sq), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveStatQueue(sq.Tenant, sq.ID,
		utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetStatQueue(sq.Tenant,
		sq.ID, false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITThresholdProfile(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter2",
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	th := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "test",
		FilterIDs: []string{"TestFilter2"},
		MaxHits:   12,
		MinSleep:  0,
		Blocker:   true,
		Weight:    1.4,
		ActionIDs: []string{"Action1"},
	}
	if err := onStor.SetFilter(context.TODO(), fp, true); err != nil {
		t.Error(err)
	}
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant, th.ID,
		true, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID,
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	expectedR := []string{"thp_cgrates.org:test"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.ThresholdProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedR, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedR, itm)
	}
	//update
	th.ActionIDs = []string{"Action1", "Action2"}
	if err := onStor.SetThresholdProfile(th, true); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetThresholdProfile(th.Tenant, th.ID,
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(th, rcv) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveThresholdProfile(th.Tenant,
		th.ID, utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetThresholdProfile(th.Tenant,
		th.ID, false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITThreshold(t *testing.T) {
	th := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH1",
		Snooze: time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC),
		Hits:   10,
	}
	if _, rcvErr := onStor.GetThreshold("cgrates.org", "TH1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetThreshold(th, 0, true); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(th, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	expectedT := []string{"thd_cgrates.org:TH1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.ThresholdPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	th.Hits = 20
	if err := onStor.SetThreshold(th, 0, true); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetThreshold("cgrates.org", "TH1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(th, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(th), utils.ToJSON(rcv))
	}
	if err := onStor.RemoveThreshold(th.Tenant, th.ID, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetThreshold(th.Tenant, th.ID,
		false, false, utils.NonTransactional); rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITFilter(t *testing.T) {
	fp := &Filter{
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	if err := fp.Compile(); err != nil {
		t.Fatal(err)
	}
	if _, rcvErr := onStor.GetFilter(context.TODO(), "cgrates.org", "Filter1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetFilter(context.TODO(), fp, true); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetFilter(context.TODO(), "cgrates.org", "Filter1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	//get from database
	if rcv, err := onStor.GetFilter(context.TODO(), "cgrates.org", "Filter1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	expectedT := []string{"ftr_cgrates.org:Filter1", "ftr_cgrates.org:TestFilter2"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.FilterPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(len(expectedT), len(itm)) {
		t.Errorf("Expected : %+v, but received %+v", len(expectedT), len(itm))
	}
	//update
	fp.Rules = []*FilterRule{
		{
			Element: "Account",
			Type:    utils.MetaString,
			Values:  []string{"1001", "1002"},
		},
		{
			Element: "Destination",
			Type:    utils.MetaString,
			Values:  []string{"10", "20"},
		},
	}
	if err := fp.Compile(); err != nil {
		t.Fatal(err)
	}
	if err := onStor.SetFilter(context.TODO(), fp, true); err != nil {
		t.Error(err)
	}

	//get from cache
	if rcv, err := onStor.GetFilter(context.TODO(), "cgrates.org", "Filter1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	//get from database
	if rcv, err := onStor.GetFilter(context.TODO(), "cgrates.org", "Filter1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(fp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", fp, rcv)
	}
	if err := onStor.RemoveFilter(context.TODO(), fp.Tenant, fp.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetFilter(context.TODO(), "cgrates.org", "Filter1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetFilter(context.TODO(), "cgrates.org", "Filter1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITRouteProfile(t *testing.T) {
	splProfile := &RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "SPRF_1",
		FilterIDs:         []string{"*string:~*reg.Accout:1002", "*string:~*reg.Destination:11", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Sorting:           "*lowest_cost",
		SortingParameters: []string{},
		Routes: []*Route{
			{
				ID:              "supplier1",
				FilterIDs:       []string{"FLTR_DST_DE"},
				AccountIDs:      []string{"Account1"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGR1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				RouteParameters: "param1",
			},
		},
		Weight: 20,
	}
	if _, rcvErr := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetRouteProfile(splProfile, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	expectedT := []string{"rpp_cgrates.org:SPRF_1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.RouteProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	splProfile.Routes = []*Route{
		{
			ID:              "supplier1",
			FilterIDs:       []string{"FLTR_DST_DE"},
			AccountIDs:      []string{"Account1"},
			RatingPlanIDs:   []string{"RPL_1"},
			ResourceIDs:     []string{"ResGR1"},
			StatIDs:         []string{"Stat1"},
			Weight:          10,
			RouteParameters: "param1",
		},
		{
			ID:              "supplier2",
			FilterIDs:       []string{"FLTR_DST_DE"},
			AccountIDs:      []string{"Account2"},
			RatingPlanIDs:   []string{"RPL_2"},
			ResourceIDs:     []string{"ResGR2"},
			StatIDs:         []string{"Stat2"},
			Weight:          20,
			RouteParameters: "param2",
		},
	}
	if err := onStor.SetRouteProfile(splProfile, false); err != nil {
		t.Error(err)
	}

	//get from cache
	if rcv, err := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(splProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", splProfile, rcv)
	}
	if err := onStor.RemoveRouteProfile(splProfile.Tenant, splProfile.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetRouteProfile("cgrates.org", "SPRF_1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITAttributeProfile(t *testing.T) {
	attrProfile := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf1",
		FilterIDs: []string{"*string:~*reg.Accout:1002", "*string:~*reg.Destination:11", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Contexts:  []string{"con1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FN1",
				Value: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if _, rcvErr := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetAttributeProfile(context.TODO(), attrProfile, false); err != nil {
		t.Error(err)
	}
	//get from cache
	if rcv, err := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	expectedT := []string{"alp_cgrates.org:AttrPrf1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.AttributeProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	attrProfile.Contexts = []string{"con1", "con2", "con3"}
	if err := onStor.SetAttributeProfile(context.TODO(), attrProfile, false); err != nil {
		t.Error(err)
	}

	//get from cache
	if rcv, err := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	//get from database
	if rcv, err := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	if err := onStor.RemoveAttributeProfile(context.TODO(), attrProfile.Tenant,
		attrProfile.ID, utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check cache if removed
	if _, rcvErr := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	//check database if removed
	if _, rcvErr := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		false, true, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITTestAttributeSubstituteIface(t *testing.T) {
	attrProfile := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf1",
		FilterIDs: []string{"*string:~*reg.Accout:1002", "*string:~*reg.Destination:11", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Contexts:  []string{"con1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FN1",
				Value: config.NewRSRParsersMustCompile("Val1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if _, rcvErr := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetAttributeProfile(context.TODO(), attrProfile, false); err != nil {
		t.Error(err)
	}
	//check database
	if rcv, err := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", attrProfile, rcv)
	}
	attrProfile.Attributes = []*Attribute{
		{
			Path:  utils.MetaReq + utils.NestingSep + "FN1",
			Value: config.NewRSRParsersMustCompile("123.123", utils.InfieldSep),
		},
	}
	if err := onStor.SetAttributeProfile(context.TODO(), attrProfile, false); err != nil {
		t.Error(err)
	}
	//check database
	if rcv, err := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(attrProfile), utils.ToJSON(rcv))
	}
	attrProfile.Attributes = []*Attribute{
		{
			Path:  utils.MetaReq + utils.NestingSep + "FN1",
			Value: config.NewRSRParsersMustCompile("true", utils.InfieldSep),
		},
	}
	if err := onStor.SetAttributeProfile(context.TODO(), attrProfile, false); err != nil {
		t.Error(err)
	}
	//check database
	if rcv, err := onStor.GetAttributeProfile(context.TODO(), "cgrates.org", "AttrPrf1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(attrProfile, rcv)) {
		t.Errorf("Expecting: %v, received: %v", utils.ToJSON(attrProfile), utils.ToJSON(rcv))
	}
}

func testOnStorITChargerProfile(t *testing.T) {
	cpp := &ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "CPP_1",
		FilterIDs:    []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR_1"},
		Weight:       20,
	}
	if _, rcvErr := onStor.GetChargerProfile("cgrates.org", "CPP_1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetChargerProfile(cpp, false); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetChargerProfile("cgrates.org", "CPP_1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(cpp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", cpp, rcv)
	}
	expectedT := []string{"cpp_cgrates.org:CPP_1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.ChargerProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	cpp.FilterIDs = []string{"*string:~*req.Accout:1001", "*prefix:~*req.Destination:10"}
	if err := onStor.SetChargerProfile(cpp, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetChargerProfile("cgrates.org", "CPP_1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(cpp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", cpp, rcv)
	}
	if err := onStor.RemoveChargerProfile(cpp.Tenant, cpp.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetChargerProfile("cgrates.org", "CPP_1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITDispatcherProfile(t *testing.T) {
	dpp := &DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:Account:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
		Strategy:  utils.MetaFirst,
		// Hosts:    []string{"192.168.56.203"},
		Weight: 20,
	}
	if _, rcvErr := onStor.GetDispatcherProfile("cgrates.org", "Dsp1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetDispatcherProfile(dpp, false); err != nil {
		t.Error(err)
	}
	//get from database
	if rcv, err := onStor.GetDispatcherProfile("cgrates.org", "Dsp1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(dpp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", dpp, rcv)
	}
	expectedT := []string{"dpp_cgrates.org:Dsp1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.DispatcherProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	dpp.FilterIDs = []string{"*string:~*req.Accout:1001", "*prefix:~*req.Destination:10"}
	if err := onStor.SetDispatcherProfile(dpp, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetDispatcherProfile("cgrates.org", "Dsp1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(dpp, rcv)) {
		t.Errorf("Expecting: %v, received: %v", dpp, rcv)
	}
	if err := onStor.RemoveDispatcherProfile(dpp.Tenant, dpp.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetDispatcherProfile("cgrates.org", "Dsp1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITRateProfile(t *testing.T) {
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001", "*string:~*req.Subject:1002"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"FIRST_GI": {
				ID:        "FIRST_GI",
				FilterIDs: []string{"*gi:~*req.Usage:0"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Blocker: false,
			},
			"SECOND_GI": {
				ID:        "SECOND_GI",
				FilterIDs: []string{"*gi:~*req.Usage:1m"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blocker: false,
			},
		},
	}
	if _, rcvErr := onStor.GetRateProfile(context.TODO(), "cgrates.org", "RP1",
		true, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
	if err := onStor.SetRateProfile(context.TODO(), rPrf, false); err != nil {
		t.Error(err)
	}
	if err = rPrf.Compile(); err != nil {
		t.Fatal(err)
	}
	//get from database
	if rcv, err := onStor.GetRateProfile(context.TODO(), "cgrates.org", "RP1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rPrf, rcv) {
		t.Errorf("Expecting: %v, received: %v", rPrf, rcv)
	}
	expectedT := []string{"rtp_cgrates.org:RP1"}
	if itm, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.RateProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedT, itm) {
		t.Errorf("Expected : %+v, but received %+v", expectedT, itm)
	}
	//update
	rPrf.FilterIDs = []string{"*string:~*req.Accout:1001", "*prefix:~*req.Destination:10"}
	if err := onStor.SetRateProfile(context.TODO(), rPrf, false); err != nil {
		t.Error(err)
	}

	//get from database
	if rcv, err := onStor.GetRateProfile(context.TODO(), "cgrates.org", "RP1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rPrf, rcv)) {
		t.Errorf("Expecting: %v, received: %v", rPrf, rcv)
	}
	if err := onStor.RemoveRateProfile(context.TODO(), rPrf.Tenant, rPrf.ID,
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
	//check database if removed
	if _, rcvErr := onStor.GetRateProfile(context.TODO(), "cgrates.org", "RP1",
		false, false, utils.NonTransactional); rcvErr != nil && rcvErr != utils.ErrNotFound {
		t.Error(rcvErr)
	}
}

func testOnStorITActionProfile(t *testing.T) {
	actPrf := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_ID1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    20,
		Schedule:  utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: utils.NewStringSet([]string{"acc1", "acc2", "acc3"}),
		},
		Actions: []*APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*APDiktat{{
					Path: "~*balance.TestBalance.Value",
				}},
			},
			{
				ID:        "TOPUP_TEST_VOICE",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*APDiktat{{
					Path: "~*balance.TestVoiceBalance.Value",
				}},
			},
		},
	}

	//empty in database
	if _, err := onStor.GetActionProfile("cgrates.org", "TEST_ID1",
		true, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}

	//get from database
	if err := onStor.SetActionProfile(actPrf, false); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetActionProfile("cgrates.org", "TEST_ID1",
		true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, actPrf) {
		t.Errorf("Expecting: %v, received: %v", actPrf, rcv)
	}

	//craft akeysFromPrefix
	expectedKey := []string{"acp_cgrates.org:TEST_ID1"}
	if rcv, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.ActionProfilePrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedKey, rcv) {
		t.Errorf("Expecting: %v, received: %v", expectedKey, rcv)
	}

	//updateFilters
	actPrf.FilterIDs = []string{"*prefix:~*req.Destination:10"}
	if err := onStor.SetActionProfile(actPrf, false); err != nil {
		t.Error(err)
	} else if rcv, err := onStor.GetActionProfile("cgrates.org", "TEST_ID1",
		false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actPrf, rcv) {
		t.Errorf("Expecting: %v, received: %v", actPrf, rcv)
	}

	//remove from database
	if err := onStor.RemoveActionProfile("cgrates.org", "TEST_ID1",
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	} else if _, err := onStor.GetActionProfile("cgrates.org", "TEST_ID1",
		false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testOnStorITAccount(t *testing.T) {
	acctPrf := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"test_filterId", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|2014-07-15T14:25:00Z"},
		Weights: utils.DynamicWeights{
			{
				Weight: 2,
			},
		},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"FLTR_RES_GR2"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type: utils.MetaVoice,
				Units: &utils.Decimal{
					Big: new(decimal.Big).SetUint64(10),
				},
				Opts: map[string]interface{}{
					"key1": "val1",
				},
			}},
		ThresholdIDs: []string{"test_thrs"},
	}

	//empty in database
	if _, err := onStor.GetAccount("cgrates.org", "RP1"); err != utils.ErrNotFound {
		t.Error(err)
	}

	//get from database
	if err := onStor.SetAccount(acctPrf, false); err != nil {
		t.Error(err)
	}
	if rcv, err := onStor.GetAccount("cgrates.org", "RP1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, acctPrf) {
		t.Errorf("Expecting: %v, received: %v", acctPrf, rcv)
	}

	//craft akeysFromPrefix
	expectedKey := []string{"acn_cgrates.org:RP1"}
	if rcv, err := onStor.DataDB().GetKeysForPrefix(context.TODO(), utils.AccountPrefix); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedKey, rcv) {
		t.Errorf("Expecting: %v, received: %v", expectedKey, rcv)
	}

	//updateFilters
	acctPrf.FilterIDs = []string{"*prefix:~*req.Destination:10"}
	if err := onStor.SetAccount(acctPrf, false); err != nil {
		t.Error(err)
	} else if rcv, err := onStor.GetAccount("cgrates.org", "RP1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(acctPrf, rcv) {
		t.Errorf("Expecting: %v, received: %v", acctPrf, rcv)
	}

	//remove from database
	if err := onStor.RemoveAccount("cgrates.org", "RP1",
		utils.NonTransactional, false); err != nil {
		t.Error(err)
	} else if _, err := onStor.GetAccount("cgrates.org", "RP1"); err != utils.ErrNotFound {
		t.Error(err)
	}
}
