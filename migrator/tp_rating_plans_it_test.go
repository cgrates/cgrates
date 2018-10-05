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
	tpRatPlnPathIn   string
	tpRatPlnPathOut  string
	tpRatPlnCfgIn    *config.CGRConfig
	tpRatPlnCfgOut   *config.CGRConfig
	tpRatPlnMigrator *Migrator
	tpRatingPlan     []*utils.TPRatingPlan
)

var sTestsTpRatPlnIT = []func(t *testing.T){
	testTpRatPlnITConnect,
	testTpRatPlnITFlush,
	testTpRatPlnITPopulate,
	testTpRatPlnITMove,
	testTpRatPlnITCheckData,
}

func TestTpRatPlnMove(t *testing.T) {
	for _, stest := range sTestsTpRatPlnIT {
		t.Run("testTpRatPlnMove", stest)
	}
}

func testTpRatPlnITConnect(t *testing.T) {
	var err error
	tpRatPlnPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpRatPlnCfgIn, err = config.NewCGRConfigFromFolder(tpRatPlnPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpRatPlnPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpRatPlnCfgOut, err = config.NewCGRConfigFromFolder(tpRatPlnPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpRatPlnCfgIn.StorDbCfg().StorDBType,
		tpRatPlnCfgIn.StorDbCfg().StorDBHost, tpRatPlnCfgIn.StorDbCfg().StorDBPort,
		tpRatPlnCfgIn.StorDbCfg().StorDBName, tpRatPlnCfgIn.StorDbCfg().StorDBUser,
		tpRatPlnCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpRatPlnCfgOut.StorDbCfg().StorDBType,
		tpRatPlnCfgOut.StorDbCfg().StorDBHost, tpRatPlnCfgOut.StorDbCfg().StorDBPort,
		tpRatPlnCfgOut.StorDbCfg().StorDBName, tpRatPlnCfgOut.StorDbCfg().StorDBUser,
		tpRatPlnCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpRatPlnMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpRatPlnITFlush(t *testing.T) {
	if err := tpRatPlnMigrator.storDBIn.StorDB().Flush(
		path.Join(tpRatPlnCfgIn.DataFolderPath, "storage", tpRatPlnCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpRatPlnMigrator.storDBOut.StorDB().Flush(
		path.Join(tpRatPlnCfgOut.DataFolderPath, "storage", tpRatPlnCfgOut.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpRatPlnITPopulate(t *testing.T) {
	tpRatingPlan = []*utils.TPRatingPlan{
		{
			TPid: "TPRP1",
			ID:   "IDPlan2",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "RateId",
					TimingId:           "TimingID",
					Weight:             12,
				},
				{
					DestinationRatesId: "DR_FREESWITCH_USERS",
					TimingId:           "ALWAYS",
					Weight:             10,
				},
			},
		},
	}
	if err := tpRatPlnMigrator.storDBIn.StorDB().SetTPRatingPlans(tpRatingPlan); err != nil {
		t.Error("Error when setting TpRatingPlans ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpRatPlnMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpRatingPlans ", err.Error())
	}
}

func testTpRatPlnITMove(t *testing.T) {
	err, _ := tpRatPlnMigrator.Migrate([]string{utils.MetaTpRatingPlans})
	if err != nil {
		t.Error("Error when migrating TpRatingPlans ", err.Error())
	}
}

func testTpRatPlnITCheckData(t *testing.T) {
	reverseRatingPlanBindings := []*utils.TPRatingPlanBinding{
		{
			DestinationRatesId: "DR_FREESWITCH_USERS",
			TimingId:           "ALWAYS",
			Weight:             10,
		},
		{
			DestinationRatesId: "RateId",
			TimingId:           "TimingID",
			Weight:             12,
		},
	}
	result, err := tpRatPlnMigrator.storDBOut.StorDB().GetTPRatingPlans(
		tpRatingPlan[0].TPid, tpRatingPlan[0].ID, nil)
	if err != nil {
		t.Error("Error when getting TpRatingPlans ", err.Error())
	}
	if !reflect.DeepEqual(tpRatingPlan[0].TPid, result[0].TPid) {
		t.Errorf("Expecting: %+v, received: %+v", tpRatingPlan[0].TPid, result[0].TPid)
	} else if !reflect.DeepEqual(tpRatingPlan[0].ID, result[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", tpRatingPlan[0].ID, result[0].ID)
	} else if !reflect.DeepEqual(tpRatingPlan[0].RatingPlanBindings, result[0].RatingPlanBindings) &&
		!reflect.DeepEqual(result[0].RatingPlanBindings, reverseRatingPlanBindings) {
		t.Errorf("Expecting: %+v, received: %+v", reverseRatingPlanBindings, result[0].RatingPlanBindings)
	}
	result, err = tpRatPlnMigrator.storDBIn.StorDB().GetTPRatingPlans(
		tpRatingPlan[0].TPid, tpRatingPlan[0].ID, nil)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
