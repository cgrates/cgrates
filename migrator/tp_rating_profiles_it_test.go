//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	tpRatPrfMigrator.Close()
}

func testTpRatPrfITConnect(t *testing.T) {
	var err error
	tpRatPrfPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	tpRatPrfCfgIn, err = config.NewCGRConfigFromPath(tpRatPrfPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpRatPrfPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	tpRatPrfCfgOut, err = config.NewCGRConfigFromPath(tpRatPrfPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpRatPrfCfgIn.StorDbCfg().Type,
		tpRatPrfCfgIn.StorDbCfg().Host, tpRatPrfCfgIn.StorDbCfg().Port,
		tpRatPrfCfgIn.StorDbCfg().Name, tpRatPrfCfgIn.StorDbCfg().User,
		tpRatPrfCfgIn.StorDbCfg().Password, tpRatPrfCfgIn.GeneralCfg().DBDataEncoding,
		tpRatPrfCfgIn.StorDbCfg().StringIndexedFields, tpRatPrfCfgIn.StorDbCfg().PrefixIndexedFields,
		tpRatPrfCfgIn.StorDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpRatPrfCfgOut.StorDbCfg().Type,
		tpRatPrfCfgOut.StorDbCfg().Host, tpRatPrfCfgOut.StorDbCfg().Port,
		tpRatPrfCfgOut.StorDbCfg().Name, tpRatPrfCfgOut.StorDbCfg().User,
		tpRatPrfCfgOut.StorDbCfg().Password, tpRatPrfCfgOut.GeneralCfg().DBDataEncoding,
		tpRatPrfCfgIn.StorDbCfg().StringIndexedFields, tpRatPrfCfgIn.StorDbCfg().PrefixIndexedFields,
		tpRatPrfCfgOut.StorDbCfg().Opts, nil)
	if err != nil {
		log.Fatal(err)
	}
	tpRatPrfMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpRatPrfITFlush(t *testing.T) {
	if err := tpRatPrfMigrator.storDBIn.StorDB().Flush(
		path.Join(tpRatPrfCfgIn.DataFolderPath, "storage", dbPath(tpRatPrfCfgIn.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}

	if err := tpRatPrfMigrator.storDBOut.StorDB().Flush(
		path.Join(tpRatPrfCfgOut.DataFolderPath, "storage", dbPath(tpRatPrfCfgOut.StorDbCfg().Type))); err != nil {
		t.Error(err)
	}
}

func testTpRatPrfITPopulate(t *testing.T) {
	tpRatingProfile = []*utils.TPRatingProfile{
		{
			TPid:     "TPRProf1",
			LoadId:   "RPrf",
			Tenant:   "Tenant1",
			Category: "Category",
			Subject:  "Subject",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2014-07-29T15:00:00Z",
					RatingPlanId:     "PlanOne",
					FallbackSubjects: "FallBack",
				},
				{
					ActivationTime:   "2015-07-29T10:00:00Z",
					RatingPlanId:     "PlanTwo",
					FallbackSubjects: "FallOut",
				},
			},
		},
	}
	if err := tpRatPrfMigrator.storDBIn.StorDB().SetTPRatingProfiles(tpRatingProfile); err != nil {
		t.Error("Error when setting TpRatingProfiles ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpRatPrfMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpRatingProfiles ", err.Error())
	}
}

func testTpRatPrfITMove(t *testing.T) {
	_, err := tpRatPrfMigrator.Migrate([]string{utils.MetaTpRatingProfiles})
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
