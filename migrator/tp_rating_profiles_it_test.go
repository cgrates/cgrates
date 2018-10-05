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
	tpRatPrfPathIn   string
	tpRatPrfPathOut  string
	tpRatPrfCfgIn    *config.CGRConfig
	tpRatPrfCfgOut   *config.CGRConfig
	tpRatPrfMigrator *Migrator
	tpRatingProfile  []*utils.TPRatingProfile
)

var sTestsTpRatPrfIT = []func(t *testing.T){
	testTpRatPrfITConnect,
	testTpRatPrfITFlush,
	testTpRatPrfITPopulate,
	testTpRatPrfITMove,
	testTpRatPrfITCheckData,
}

func TestTpRatPrfMove(t *testing.T) {
	for _, stest := range sTestsTpRatPrfIT {
		t.Run("testTpRatPrfMove", stest)
	}
}

func testTpRatPrfITConnect(t *testing.T) {
	var err error
	tpRatPrfPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpRatPrfCfgIn, err = config.NewCGRConfigFromFolder(tpRatPrfPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpRatPrfPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpRatPrfCfgOut, err = config.NewCGRConfigFromFolder(tpRatPrfPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpRatPrfCfgIn.StorDbCfg().StorDBType,
		tpRatPrfCfgIn.StorDbCfg().StorDBHost, tpRatPrfCfgIn.StorDbCfg().StorDBPort,
		tpRatPrfCfgIn.StorDbCfg().StorDBName, tpRatPrfCfgIn.StorDbCfg().StorDBUser,
		tpRatPrfCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpRatPrfCfgOut.StorDbCfg().StorDBType,
		tpRatPrfCfgOut.StorDbCfg().StorDBHost, tpRatPrfCfgOut.StorDbCfg().StorDBPort,
		tpRatPrfCfgOut.StorDbCfg().StorDBName, tpRatPrfCfgOut.StorDbCfg().StorDBUser,
		tpRatPrfCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpRatPrfMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpRatPrfITFlush(t *testing.T) {
	if err := tpRatPrfMigrator.storDBIn.StorDB().Flush(
		path.Join(tpRatPrfCfgIn.DataFolderPath, "storage", tpRatPrfCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpRatPrfMigrator.storDBOut.StorDB().Flush(
		path.Join(tpRatPrfCfgOut.DataFolderPath, "storage", tpRatPrfCfgOut.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpRatPrfITPopulate(t *testing.T) {
	tpRatingProfile = []*utils.TPRatingProfile{
		{
			TPid:      "TPRProf1",
			LoadId:    "RPrf",
			Direction: "*out",
			Tenant:    "Tenant1",
			Category:  "Category",
			Subject:   "Subject",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2014-07-29T15:00:00Z",
					RatingPlanId:     "PlanOne",
					FallbackSubjects: "FallBack",
					CdrStatQueueIds:  "RandomId",
				},
				{
					ActivationTime:   "2015-07-29T10:00:00Z",
					RatingPlanId:     "PlanTwo",
					FallbackSubjects: "FallOut",
					CdrStatQueueIds:  "RandomIdTwo",
				},
			},
		},
	}
	if err := tpRatPrfMigrator.storDBIn.StorDB().SetTPRatingProfiles(tpRatingProfile); err != nil {
		t.Error("Error when setting TpRatingProfiles ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpRatPrfMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpRatingProfiles ", err.Error())
	}
}

func testTpRatPrfITMove(t *testing.T) {
	err, _ := tpRatPrfMigrator.Migrate([]string{utils.MetaTpRatingProfiles})
	if err != nil {
		t.Error("Error when migrating TpRatingProfiles ", err.Error())
	}
}

func testTpRatPrfITCheckData(t *testing.T) {
	filter := &utils.TPRatingProfile{TPid: tpRatingProfile[0].TPid, LoadId: tpRatingProfile[0].LoadId}
	result, err := tpRatPrfMigrator.storDBOut.StorDB().GetTPRatingProfiles(filter)
	if err != nil {
		t.Error("Error when getting TpRatingProfiles ", err.Error())
	}
	if !reflect.DeepEqual(tpRatingProfile[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tpRatingProfile[0], result[0])
	}
	result, err = tpRatPrfMigrator.storDBIn.StorDB().GetTPRatingProfiles(filter)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
