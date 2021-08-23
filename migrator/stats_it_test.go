//go:build integration
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
}

func TestStatsQueueITRedis(t *testing.T) {
	var err error
	stsPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	stsCfgIn, err = config.NewCGRConfigFromPath(stsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	stsCfgOut, err = config.NewCGRConfigFromPath(stsPathIn)
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
	stsCfgOut, err = config.NewCGRConfigFromPath(stsPathIn)
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
	dataDBIn, err := NewMigratorDataDB(stsCfgIn.DataDbCfg().DataDbType,
		stsCfgIn.DataDbCfg().DataDbHost, stsCfgIn.DataDbCfg().DataDbPort,
		stsCfgIn.DataDbCfg().DataDbName, stsCfgIn.DataDbCfg().DataDbUser,
		stsCfgIn.DataDbCfg().DataDbPass, stsCfgIn.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "", stsCfgIn.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	dataDBOut, err := NewMigratorDataDB(stsCfgOut.DataDbCfg().DataDbType,
		stsCfgOut.DataDbCfg().DataDbHost, stsCfgOut.DataDbCfg().DataDbPort,
		stsCfgOut.DataDbCfg().DataDbName, stsCfgOut.DataDbCfg().DataDbUser,
		stsCfgOut.DataDbCfg().DataDbPass, stsCfgOut.GeneralCfg().DBDataEncoding,
		config.CgrConfig().CacheCfg(), "", stsCfgOut.DataDbCfg().Items)
	if err != nil {
		log.Fatal(err)
	}
	stsMigrator, err = NewMigrator(dataDBIn, dataDBOut, nil, nil,
		false, false, false, false)
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
		Id:              "test",                         // Config id, unique per config instance
		QueueLength:     10,                             // Number of items in the stats buffer
		TimeWindow:      time.Duration(1) * time.Second, // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
		SaveInterval:    time.Duration(1) * time.Second,
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
		UsageInterval:   []time.Duration{1 * time.Second},
		PddInterval:     []time.Duration{1 * time.Second},
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
					Type:           utils.StringPointer(utils.MONETARY),
				},
				ExpirationDate:    tim,
				LastExecutionTime: tim,
				ActivationDate:    tim,
				ThresholdType:     utils.TRIGGER_MAX_BALANCE,
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
		TTL:         time.Duration(0) * time.Second,
		Metrics: []*engine.MetricWithFilters{
			&engine.MetricWithFilters{
				MetricID: "*asr",
			},
			&engine.MetricWithFilters{
				MetricID: utils.MetaACD,
			},
			&engine.MetricWithFilters{
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

		result1, err := engine.GetFilter(stsMigrator.dmOut.DataManager(), "cgrates.org", v1Sts.Id, false, false, utils.NonTransactional)
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
		if err := stsMigrator.dmOut.DataManager().SetFilter(filter); err != nil {
			t.Error("Error when setting Filter ", err.Error())
		}
		currentVersion := engine.CurrentDataDBVersions()
		err := stsMigrator.dmOut.DataManager().DataDB().SetVersions(currentVersion, false)
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
