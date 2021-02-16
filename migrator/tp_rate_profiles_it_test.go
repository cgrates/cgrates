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
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpRatePrfPathIn   string
	tpRatePrfPathOut  string
	tpRatePrfCfgIn    *config.CGRConfig
	tpRatePrfCfgOut   *config.CGRConfig
	tpRatePrfMigrator *Migrator
	tpRateProfiles    []*utils.TPRateProfile
)

var sTestsTPRatePrfIT = []func(t *testing.T){
	testTPRateProfileConnect,
	testTPRateProfileFlush,
	testTPRateProfilePopulate,
	testTpRateProfileMove,
	testTpRateProfileCheckData,
}

func TestTPRateProfileIT(t *testing.T) {
	for _, tests := range sTestsTPRatePrfIT {
		t.Run("TestTPRatePrfIT", tests)
	}
	tpRatePrfMigrator.Close()
}

func testTPRateProfileConnect(t *testing.T) {
	var err error
	tpRatePrfPathIn := path.Join(*dataDir, "conf", "samples", "tutmongo")
	if tpRatePrfCfgIn, err = config.NewCGRConfigFromPath(tpRatePrfPathIn); err != nil {
		t.Error(err)
	}
	tpRatePrfPathOut := path.Join(*dataDir, "conf", "samples", "tutmysql")
	if tpRatePrfCfgOut, err = config.NewCGRConfigFromPath(tpRatePrfPathOut); err != nil {
		t.Error(err)
	}
	storDBIn, err := NewMigratorStorDB(tpRatePrfCfgIn.StorDbCfg().Type,
		tpRatePrfCfgIn.StorDbCfg().Host, tpRatePrfCfgIn.StorDbCfg().Port,
		tpRatePrfCfgIn.StorDbCfg().Name, tpRatePrfCfgIn.StorDbCfg().User,
		tpRatePrfCfgIn.StorDbCfg().Password, tpRatePrfCfgIn.GeneralCfg().DBDataEncoding,
		tpRatePrfCfgIn.StorDbCfg().StringIndexedFields, tpRatePrfCfgIn.StorDbCfg().PrefixIndexedFields,
		tpRatePrfCfgIn.StorDbCfg().Opts)
	if err != nil {
		t.Error(err)
	}
	storDBOut, err := NewMigratorStorDB(tpRatePrfCfgOut.StorDbCfg().Type,
		tpRatePrfCfgOut.StorDbCfg().Host, tpRatePrfCfgOut.StorDbCfg().Port,
		tpRatePrfCfgOut.StorDbCfg().Name, tpRatePrfCfgOut.StorDbCfg().User,
		tpRatePrfCfgOut.StorDbCfg().Password, tpRatePrfCfgOut.GeneralCfg().DBDataEncoding,
		tpRatePrfCfgOut.StorDbCfg().StringIndexedFields, tpRatePrfCfgOut.StorDbCfg().PrefixIndexedFields,
		tpRatePrfCfgOut.StorDbCfg().Opts)
	if err != nil {
		t.Error(err)
	}
	tpRatePrfMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut,
		false, false, false, false)
	if err != nil {
		t.Error(err)
	}
}

func testTPRateProfileFlush(t *testing.T) {
	if err := tpRatePrfMigrator.storDBIn.StorDB().Flush(
		path.Join(tpRatePrfCfgIn.DataFolderPath, "storage", tpRatePrfCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
	if err := tpRatePrfMigrator.storDBOut.StorDB().Flush(
		path.Join(tpRatePrfCfgOut.DataFolderPath, "storage", tpRatePrfCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTPRateProfilePopulate(t *testing.T) {
	tpRateProfiles = []*utils.TPRateProfile{
		{
			TPid:            "id_RP1",
			Tenant:          "cgrates.org",
			ID:              "RP1",
			FilterIDs:       []string{"*string:~*req.Subject:1001"},
			Weights:         ";0",
			MinCost:         0.1,
			MaxCost:         0.6,
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.TPRate{
				"FIRST_GI": {
					ID:        "FIRST_GI",
					FilterIDs: []string{"*gi:~*req.Usage:0"},
					Weights:   ";0",
					IntervalRates: []*utils.TPIntervalRate{
						{
							RecurrentFee: 0.12,
							Unit:         "1m",
							Increment:    "1m",
						},
					},
					Blocker: false,
				},
				"SECOND_GI": {
					ID:        "SECOND_GI",
					FilterIDs: []string{"*gi:~*req.Usage:1m"},
					Weights:   ";10",
					IntervalRates: []*utils.TPIntervalRate{
						{
							RecurrentFee: 0.06,
							Unit:         "1m",
							Increment:    "1s",
						},
					},
					Blocker: false,
				},
			},
		},
	}
	//empty in database
	if _, err := tpRatePrfMigrator.storDBIn.StorDB().GetTPRateProfiles(tpRateProfiles[0].TPid,
		utils.EmptyString, tpRateProfiles[0].ID); err != utils.ErrNotFound {
		t.Error(err)
	}

	if err := tpRatePrfMigrator.storDBIn.StorDB().SetTPRateProfiles(tpRateProfiles); err != nil {
		t.Error("Error when setting TpRateProfile ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpRatePrfMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpRateProfile ", err.Error())
	}
}

func testTpRateProfileMove(t *testing.T) {
	err, _ := tpRatePrfMigrator.Migrate([]string{utils.MetaTpRateProfiles})
	if err != nil {
		t.Error("Error when migrating TpRateProfiles", err.Error())
	}
}

func testTpRateProfileCheckData(t *testing.T) {
	rcv, err := tpRatePrfMigrator.storDBOut.StorDB().GetTPRateProfiles(tpRateProfiles[0].TPid,
		utils.EmptyString, tpRateProfiles[0].ID)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv[0], tpRateProfiles[0]) {
		t.Errorf("Expected %+v, received %+v", tpRateProfiles[0], rcv[0])
	}

	_, err = tpRatePrfMigrator.storDBIn.StorDB().GetTPRateProfiles(tpRateProfiles[0].TPid,
		utils.EmptyString, tpRateProfiles[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}

}
