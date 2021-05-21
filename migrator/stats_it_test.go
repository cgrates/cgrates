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

package migrator

import (
	"log"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	stsPathIn   string
	stsPathOut  string
	stsCfgIn    *config.CGRConfig
	stsCfgOut   *config.CGRConfig
	stsMigrator *Migrator
	stsAction   string
)

var sTestsStsIT = []func(t *testing.T){
	testStsITConnect,
	testStsITFlush,
	testStsITMigrateAndMove,
	testStsITMigrateFromv1,
}

func TestStatsQueueITRedis(t *testing.T) {
	var err error
	stsPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	stsCfgIn, err = config.NewCGRConfigFromPath(stsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	stsPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	stsCfgOut, err = config.NewCGRConfigFromPath(stsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	stsAction = utils.Migrate
	for _, stest := range sTestsStsIT {
		t.Run("TestStatsQueueITMigrateRedis", stest)
	}
	stsMigrator.Close()
}

func TestStatsQueueITMongo(t *testing.T) {
	var err error
	stsPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	stsCfgIn, err = config.NewCGRConfigFromPath(stsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	stsPathOut = path.Join(*dataDir, "conf", "samples", "tutmongo")
	stsCfgOut, err = config.NewCGRConfigFromPath(stsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	stsAction = utils.Migrate
	for _, stest := range sTestsStsIT {
		t.Run("TestStatsQueueITMigrateMongo", stest)
	}
	stsMigrator.Close()
}

func TestStatsQueueITMove(t *testing.T) {
	var err error
	stsPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	stsCfgIn, err = config.NewCGRConfigFromPath(stsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	stsPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	stsCfgOut, err = config.NewCGRConfigFromPath(stsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	stsAction = utils.Move
	for _, stest := range sTestsStsIT {
		t.Run("TestStatsQueueITMove", stest)
	}
	stsMigrator.Close()
}

func testStsITConnect(t *testing.T) {
	dataDBIn, err := NewMigratorDataDB(stsCfgIn.DataDbCfg().Type,
		stsCfgIn.DataDbCfg().Host, stsCfgIn.DataDbCfg().Port,
		stsCfgIn.DataDbCfg().Name, stsCfgIn.DataDbCfg().User,
		stsCfgIn.DataDbCfg().Password, stsCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), stsCfgIn.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(stsCfgOut.DataDbCfg().Type,
		stsCfgOut.DataDbCfg().Host, stsCfgOut.DataDbCfg().Port,
		stsCfgOut.DataDbCfg().Name, stsCfgOut.DataDbCfg().User,
		stsCfgOut.DataDbCfg().Password, stsCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), stsCfgOut.DataDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	if stsPathIn == stsPathOut {
		stsMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, true, false, false)
	} else {
		stsMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
			false, false, false, false)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func testStsITFlush(t *testing.T) {
	stsMigrator.dmOut.DataManager().DataDB().Flush("")
	if err := engine.SetDBVersions(stsMigrator.dmOut.DataManager().DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testStsITMigrateAndMove(t *testing.T) {
	tim := time.Date(2012, time.February, 27, 23, 59, 59, 0, time.UTC)
	var filters []*engine.FilterRule
	v1Sts := &v1Stat{
		Id:              "test",      // Config id, unique per config instance
		QueueLength:     10,          // Number of items in the stats buffer
		TimeWindow:      time.Second, // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
		SaveInterval:    time.Second,
		Metrics:         []string{"ASR", "ACD", "ACC"},
		SetupInterval:   []time.Time{time.Now()},
		ToR:             []string{},
		CdrHost:         []string{},
		CdrSource:       []string{},
		ReqType:         []string{},
		Direction:       []string{},
		Tenant:          []string{},
		Category:        []string{},
		Account:         []string{},
		Subject:         []string{},
		DestinationIds:  []string{},
		UsageInterval:   []time.Duration{time.Second},
		PddInterval:     []time.Duration{time.Second},
		Supplier:        []string{},
		DisconnectCause: []string{},
		MediationRunIds: []string{},
		RatedAccount:    []string{},
		RatedSubject:    []string{},
		CostInterval:    []float64{},
		Triggers: engine.ActionTriggers{
			&engine.ActionTrigger{
				ID: "Test",
				Balance: &engine.BalanceFilter{
					ID:             utils.StringPointer("TESTB"),
					Timings:        []*engine.RITiming{},
					ExpirationDate: utils.TimePointer(tim),
					Type:           utils.StringPointer(utils.MetaMonetary),
				},
				ExpirationDate:    tim,
				LastExecutionTime: tim,
				ActivationDate:    tim,
				ThresholdType:     utils.TriggerMaxBalance,
				ThresholdValue:    2,
				ActionsID:         "TEST_ACTIONS",
				Executed:          true,
			},
		},
	}
	x, _ := engine.NewFilterRule(utils.MetaGreaterOrEqual,
		"SetupInterval", []string{v1Sts.SetupInterval[0].String()})
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(utils.MetaGreaterOrEqual,
		"UsageInterval", []string{v1Sts.UsageInterval[0].String()})
	filters = append(filters, x)
	x, _ = engine.NewFilterRule(utils.MetaGreaterOrEqual,
		"PddInterval", []string{v1Sts.PddInterval[0].String()})
	filters = append(filters, x)

	filter := &engine.Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     v1Sts.Id,
		Rules:  filters}

	sqp := &engine.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "test",
		FilterIDs:   []string{v1Sts.Id},
		QueueLength: 10,
		TTL:         0,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: "*asr",
			},
			{
				MetricID: utils.MetaACD,
			},
			{
				MetricID: "*acc",
			},
		},
		ThresholdIDs: []string{"Test"},
		Blocker:      false,
		Stored:       true,
		Weight:       float64(0),
		MinItems:     0,
	}
	sq := &engine.StatQueue{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        v1Sts.Id,
		SQMetrics: make(map[string]engine.StatMetric),
	}
	for _, metric := range sqp.Metrics {
		if stsMetric, err := engine.NewStatMetric(metric.MetricID, 0, []string{}); err != nil {
			t.Error("Error when creating newstatMETRIc ", err.Error())
		} else {
			if _, has := sq.SQMetrics[metric.MetricID]; !has {
				sq.SQMetrics[metric.MetricID] = stsMetric
			}
		}
	}
	switch stsAction {
	case utils.Migrate:
		err := stsMigrator.dmIN.setV1Stats(v1Sts)
		if err != nil {
			t.Error("Error when setting v1Stat ", err.Error())
		}
		currentVersion := engine.Versions{
			utils.StatS:          1,
			utils.Thresholds:     2,
			utils.Accounts:       2,
			utils.Actions:        2,
			utils.ActionTriggers: 2,
			utils.ActionPlans:    2,
			utils.SharedGroups:   2,
		}
		err = stsMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = stsMigrator.Migrate([]string{utils.MetaStats})
		if err != nil {
			t.Error("Error when migrating Stats ", err.Error())
		}

		result, err := stsMigrator.dmOut.DataManager().DataDB().GetStatQueueProfileDrv("cgrates.org", v1Sts.Id)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(sqp, result) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, result)
		}

		result1, err := stsMigrator.dmOut.DataManager().GetFilter("cgrates.org", v1Sts.Id, false, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(filter.ID, result1.ID) {
			t.Errorf("Expecting: %+v, received: %+v", filter.ID, result1.ID)
		} else if !reflect.DeepEqual(len(filter.Rules), len(result1.Rules)) {
			t.Errorf("Expecting: %+v, received: %+v", len(filter.Rules), len(result1.Rules))
		}

		result2, err := stsMigrator.dmOut.DataManager().GetStatQueue("cgrates.org", sq.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(sq.ID, result2.ID) {
			t.Errorf("Expecting: %+v, received: %+v", sq.ID, result2.ID)
		}

	case utils.Move:
		if err := stsMigrator.dmIN.DataManager().SetStatQueueProfile(sqp, false); err != nil {
			t.Error("Error when setting Stats ", err.Error())
		}
		if err := stsMigrator.dmIN.DataManager().SetStatQueue(sq); err != nil {
			t.Error("Error when setting Stats ", err.Error())
		}
		if err := stsMigrator.dmOut.DataManager().SetFilter(filter, true); err != nil {
			t.Error("Error when setting Filter ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := stsMigrator.dmIN.DataManager().DataDB().SetVersions(currentVersion, false)
		if err != nil {
			t.Error("Error when setting version for stats ", err.Error())
		}
		err, _ = stsMigrator.Migrate([]string{utils.MetaStats})
		if err != nil {
			t.Error("Error when migrating Stats ", err.Error())
		}
		result, err := stsMigrator.dmOut.DataManager().DataDB().GetStatQueueProfileDrv(sqp.Tenant, sqp.ID)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		result1, err := stsMigrator.dmOut.DataManager().GetStatQueue(sq.Tenant, sq.ID, false, false, utils.NonTransactional)
		if err != nil {
			t.Error("Error when getting Stats ", err.Error())
		}
		if !reflect.DeepEqual(sqp, result) {
			t.Errorf("Expecting: %+v, received: %+v", sqp, result)
		}
		if !reflect.DeepEqual(sq.ID, result1.ID) {
			t.Errorf("Expecting: %+v, received: %+v", sq.ID, result1.ID)
		}
	}

}

func testStsITMigrateFromv1(t *testing.T) {
	tim := time.Date(2020, time.July, 29, 17, 59, 59, 0, time.UTC)
	v1Sts := &v1Stat{
		Id:              "test",
		QueueLength:     10,
		TimeWindow:      time.Second,
		SaveInterval:    time.Second,
		Metrics:         []string{"ASR", "ACD", "ACC"},
		SetupInterval:   []time.Time{tim},
		ToR:             []string{},
		CdrHost:         []string{},
		CdrSource:       []string{},
		ReqType:         []string{},
		Direction:       []string{},
		Tenant:          []string{},
		Category:        []string{},
		Account:         []string{},
		Subject:         []string{},
		DestinationIds:  []string{},
		UsageInterval:   []time.Duration{time.Second},
		PddInterval:     []time.Duration{time.Second},
		Supplier:        []string{},
		DisconnectCause: []string{},
		MediationRunIds: []string{},
		RatedAccount:    []string{},
		RatedSubject:    []string{},
		CostInterval:    []float64{},
		Triggers: engine.ActionTriggers{
			&engine.ActionTrigger{
				ID: "Test",
				Balance: &engine.BalanceFilter{
					ID:             utils.StringPointer("TESTB"),
					Timings:        []*engine.RITiming{},
					ExpirationDate: utils.TimePointer(tim),
					Type:           utils.StringPointer(utils.MetaMonetary),
				},
				ExpirationDate:    tim,
				LastExecutionTime: tim,
				ActivationDate:    tim,
				ThresholdType:     utils.TriggerMaxBalance,
				ThresholdValue:    2,
				ActionsID:         "TEST_ACTIONS",
				Executed:          true,
			},
		},
	}

	err := stsMigrator.dmIN.setV1Stats(v1Sts)
	if err != nil {
		t.Error("Error when setting v1Stat ", err.Error())
	}

	if err := stsMigrator.dmIN.DataManager().DataDB().SetVersions(engine.Versions{utils.StatS: 1}, true); err != nil {
		t.Errorf("error: <%s> when updating Stats version into dataDB", err.Error())
	}

	if err := stsMigrator.migrateStats(); err != nil {
		t.Error(err)
	}

	if vrs, err := stsMigrator.dmOut.DataManager().DataDB().GetVersions(utils.StatS); err != nil {
		t.Errorf("error: <%s> when updating Stats version into dataDB", err.Error())
	} else if vrs[utils.StatS] != 4 {
		t.Errorf("Expecting: 4, received: %+v", vrs[utils.StatS])
	}

	//from V1 to V2
	var filter *engine.Filter
	if filter, err = stsMigrator.dmOut.DataManager().GetFilter("cgrates.org", "test", false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(config.CgrConfig().GeneralCfg().DefaultTenant, filter.Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", config.CgrConfig().GeneralCfg().DefaultTenant, filter.Tenant)
	} else if !reflect.DeepEqual(v1Sts.Id, filter.ID) {
		t.Errorf("Expecting: %+v, received: %+v", v1Sts.Id, filter.ID)
	} else if filter.ActivationInterval != nil {
		t.Errorf("Expecting: nil, received: %+v", filter.ActivationInterval)
	}

	for _, itm := range filter.Rules {
		switch itm.Element {
		case "SetupInterval":
			if itm.Values[0] != tim.String() {
				t.Errorf("Expecting: %+v, received: %+v", tim.String(), itm.Values[0])
			} else if itm.Type != "*gte" {
				t.Errorf("Expecting: *gte, received: %+v", itm.Type)
			}
		case "UsageInterval":
			if itm.Type != "*gte" {
				t.Errorf("Expecting: *gte, received: %+v", itm.Type)
			} else if itm.Values[0] != "1s" {
				t.Errorf("Expecting: 1s, received: %+v", itm.Values[0])
			}
		case "PddInterval":
			if itm.Type != "*gte" {
				t.Errorf("Expecting: *gte, received: %+v", itm.Type)
			} else if itm.Values[0] != "1s" {
				t.Errorf("Expecting: 1s, received: %+v", itm.Values[0])
			}
		}
	}
	metrics := []*engine.MetricWithFilters{
		{
			MetricID: "*asr",
		}, {
			MetricID: "*acd",
		}, {
			MetricID: "*acc",
		},
	}
	if statQueueProfile, err := stsMigrator.dmOut.DataManager().GetStatQueueProfile("cgrates.org", "test", false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if statQueueProfile.ThresholdIDs[0] != "Test" {
		t.Errorf("Expecting: 'Test', received: %+v", statQueueProfile.ThresholdIDs[0])
	} else if statQueueProfile.Weight != 0 {
		t.Errorf("Expecting: '0', received: %+v", statQueueProfile.Weight)
	} else if !statQueueProfile.Stored {
		t.Errorf("Expecting: 'true', received: %+v", statQueueProfile.Stored)
	} else if statQueueProfile.Blocker {
		t.Errorf("Expecting: 'false', received: %+v", statQueueProfile.Blocker)
	} else if statQueueProfile.QueueLength != 10 {
		t.Errorf("Expecting: '10', received: %+v", statQueueProfile.QueueLength)
	} else if statQueueProfile.ID != "test" {
		t.Errorf("Expecting: 'test', received: %+v", statQueueProfile.ID)
	} else if statQueueProfile.Tenant != "cgrates.org" {
		t.Errorf("Expecting: 'cgrates.org', received: %+v", statQueueProfile.Tenant)
	} else if statQueueProfile.MinItems != 0 {
		t.Errorf("Expecting: '0', received: %+v", statQueueProfile.MinItems)
	} else if statQueueProfile.TTL != 0 {
		t.Errorf("Expecting: '0', received: %+v", statQueueProfile.TTL)
	} else if !reflect.DeepEqual(statQueueProfile.Metrics, metrics) {
		t.Errorf("Expecting: %+v, received: %+v", metrics, statQueueProfile.Metrics)
	}

	//from V2 to V3
	var statQueue *engine.StatQueue
	if statQueue, err = stsMigrator.dmOut.DataManager().GetStatQueue("cgrates.org", "test", false, false, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if statQueue.ID != "test" {
		t.Errorf("Expecting: 'test', received: %+v", statQueue.ID)
	} else if statQueue.Tenant != "cgrates.org" {
		t.Errorf("Expecting: 'cgrates.org', received: %+v", statQueue.Tenant)
	} else if len(statQueue.SQItems) != 0 {
		t.Errorf("Expecting: '0', received: %+v", len(statQueue.SQItems))
	}
	if _, ok := statQueue.SQMetrics["*acc"]; !ok {
		t.Errorf("Expecting *acc item to be present in SQMetrics")
	}
	if _, ok := statQueue.SQMetrics["*acd"]; !ok {
		t.Errorf("Expecting *acd item to be present in SQMetrics")
	}
	if _, ok := statQueue.SQMetrics["*asr"]; !ok {
		t.Errorf("Expecting *asr item to be present in SQMetrics")
	}
}
