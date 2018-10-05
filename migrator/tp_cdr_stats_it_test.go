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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpCdrStatPathIn   string
	tpCdrStatPathOut  string
	tpCdrStatCfgIn    *config.CGRConfig
	tpCdrStatCfgOut   *config.CGRConfig
	tpCdrStatMigrator *Migrator
	tpCdrStat         []*utils.TPCdrStats
)

var sTestsTpCdrStatIT = []func(t *testing.T){
	testTpCdrStatITConnect,
	testTpCdrStatITFlush,
	testTpCdrStatITPopulate,
	testTpCdrStatITMove,
	testTpCdrStatITCheckData,
}

func TestTpCdrStatMove(t *testing.T) {
	for _, stest := range sTestsTpCdrStatIT {
		t.Run("testTpCdrStatMove", stest)
	}
}

func testTpCdrStatITConnect(t *testing.T) {
	var err error
	tpCdrStatPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpCdrStatCfgIn, err = config.NewCGRConfigFromFolder(tpCdrStatPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpCdrStatPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpCdrStatCfgOut, err = config.NewCGRConfigFromFolder(tpCdrStatPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpCdrStatCfgIn.StorDbCfg().StorDBType,
		tpCdrStatCfgIn.StorDbCfg().StorDBHost, tpCdrStatCfgIn.StorDbCfg().StorDBPort,
		tpCdrStatCfgIn.StorDbCfg().StorDBName, tpCdrStatCfgIn.StorDbCfg().StorDBUser,
		tpCdrStatCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpCdrStatCfgOut.StorDbCfg().StorDBType,
		tpCdrStatCfgOut.StorDbCfg().StorDBHost, tpCdrStatCfgOut.StorDbCfg().StorDBPort,
		tpCdrStatCfgOut.StorDbCfg().StorDBName, tpCdrStatCfgOut.StorDbCfg().StorDBUser,
		tpCdrStatCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpCdrStatMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpCdrStatITFlush(t *testing.T) {
	if err := tpCdrStatMigrator.storDBIn.StorDB().Flush(
		path.Join(tpCdrStatCfgIn.DataFolderPath, "storage", tpCdrStatCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpCdrStatMigrator.storDBOut.StorDB().Flush(
		path.Join(tpCdrStatCfgOut.DataFolderPath, "storage", tpCdrStatCfgOut.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpCdrStatITPopulate(t *testing.T) {
	tpCdrStat = []*utils.TPCdrStats{
		{
			TPid: "TPCdr",
			ID:   "ID",
			CdrStats: []*utils.TPCdrStat{
				{
					QueueLength:      "10",
					TimeWindow:       "0",
					SaveInterval:     "10s",
					Metrics:          "ASR",
					SetupInterval:    "",
					TORs:             "",
					CdrHosts:         "",
					CdrSources:       "",
					ReqTypes:         "",
					Directions:       "",
					Tenants:          "cgrates.org",
					Categories:       "",
					Accounts:         "",
					Subjects:         "1001",
					DestinationIds:   "1003",
					PddInterval:      "",
					UsageInterval:    "",
					Suppliers:        "suppl1",
					DisconnectCauses: "",
					MediationRunIds:  "*default",
					RatedAccounts:    "",
					RatedSubjects:    "",
					CostInterval:     "",
					ActionTriggers:   "CDRST1_WARN",
				},
				{
					QueueLength:      "10",
					TimeWindow:       "0",
					SaveInterval:     "10s",
					Metrics:          "ACC",
					SetupInterval:    "",
					TORs:             "",
					CdrHosts:         "",
					CdrSources:       "",
					ReqTypes:         "",
					Directions:       "",
					Tenants:          "cgrates.org",
					Categories:       "",
					Accounts:         "",
					Subjects:         "1002",
					DestinationIds:   "1003",
					PddInterval:      "",
					UsageInterval:    "",
					Suppliers:        "suppl1",
					DisconnectCauses: "",
					MediationRunIds:  "*default",
					RatedAccounts:    "",
					RatedSubjects:    "",
					CostInterval:     "",
					ActionTriggers:   "CDRST1_WARN",
				},
			},
		},
	}
	if err := tpCdrStatMigrator.storDBIn.StorDB().SetTPCdrStats(tpCdrStat); err != nil {
		t.Error("Error when setting TpCdrStats ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpCdrStatMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpCdrStats ", err.Error())
	}
}

func testTpCdrStatITMove(t *testing.T) {
	err, _ := tpCdrStatMigrator.Migrate([]string{utils.MetaTpCdrStats})
	if err != nil {
		t.Error("Error when migrating TpCdrStats ", err.Error())
	}
}

func testTpCdrStatITCheckData(t *testing.T) {
	// reverseRatingPlanBindings := []*utils.TPRatingPlanBinding{
	// 	&utils.TPRatingPlanBinding{
	// 		DestinationRatesId: "DR_FREESWITCH_USERS",
	// 		TimingId:           "ALWAYS",
	// 		Weight:             10,
	// 	},
	// 	&utils.TPRatingPlanBinding{
	// 		DestinationRatesId: "RateId",
	// 		TimingId:           "TimingID",
	// 		Weight:             12,
	// 	},
	// }
	result, err := tpCdrStatMigrator.storDBOut.StorDB().GetTPCdrStats(
		tpCdrStat[0].TPid, tpCdrStat[0].ID)
	if err != nil {
		t.Error("Error when getting TpCdrStats ", err.Error())
	}
	if !reflect.DeepEqual(tpCdrStat[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tpCdrStat[0], result[0])
	}
	result, err = tpCdrStatMigrator.storDBIn.StorDB().GetTPCdrStats(
		tpCdrStat[0].TPid, tpCdrStat[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
