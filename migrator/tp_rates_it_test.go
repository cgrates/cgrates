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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpRatesPathIn   string
	tpRatesPathOut  string
	tpRatesCfgIn    *config.CGRConfig
	tpRatesCfgOut   *config.CGRConfig
	tpRatesMigrator *Migrator
	tpRates         []*utils.TPRateRALs
)

var sTestsTpRatesIT = []func(t *testing.T){
	testTpRatesITConnect,
	testTpRatesITFlush,
	testTpRatesITPopulate,
	testTpRatesITMove,
	testTpRatesITCheckData,
}

func TestTpRatesMove(t *testing.T) {
	for _, stest := range sTestsTpRatesIT {
		t.Run("testTpRatesMove", stest)
	}
	tpRatesMigrator.Close()
}

func testTpRatesITConnect(t *testing.T) {
	var err error
	tpRatesPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpRatesCfgIn, err = config.NewCGRConfigFromPath(tpRatesPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpRatesPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpRatesCfgOut, err = config.NewCGRConfigFromPath(tpRatesPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := NewMigratorStorDB(tpRatesCfgIn.StorDbCfg().Type,
		tpRatesCfgIn.StorDbCfg().Host, tpRatesCfgIn.StorDbCfg().Port,
		tpRatesCfgIn.StorDbCfg().Name, tpRatesCfgIn.StorDbCfg().User,
		tpRatesCfgIn.StorDbCfg().Password, tpRatesCfgIn.GeneralCfg().DBDataEncoding,
		tpRatesCfgIn.StorDbCfg().StringIndexedFields, tpRatesCfgIn.StorDbCfg().PrefixIndexedFields,
		tpRatesCfgIn.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := NewMigratorStorDB(tpRatesCfgOut.StorDbCfg().Type,
		tpRatesCfgOut.StorDbCfg().Host, tpRatesCfgOut.StorDbCfg().Port,
		tpRatesCfgOut.StorDbCfg().Name, tpRatesCfgOut.StorDbCfg().User,
		tpRatesCfgOut.StorDbCfg().Password, tpRatesCfgOut.GeneralCfg().DBDataEncoding,
		tpRatesCfgIn.StorDbCfg().StringIndexedFields, tpRatesCfgIn.StorDbCfg().PrefixIndexedFields,
		tpRatesCfgOut.StorDbCfg().Opts)
	if err != nil {
		log.Fatal(err)
	}
	tpRatesMigrator, err = NewMigrator(nil, nil, storDBIn, storDBOut, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpRatesITFlush(t *testing.T) {
	if err := tpRatesMigrator.storDBIn.StorDB().Flush(
		path.Join(tpRatesCfgIn.DataFolderPath, "storage", tpRatesCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}

	if err := tpRatesMigrator.storDBOut.StorDB().Flush(
		path.Join(tpRatesCfgOut.DataFolderPath, "storage", tpRatesCfgOut.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testTpRatesITPopulate(t *testing.T) {
	tpRates = []*utils.TPRateRALs{
		{
			TPid: "TPidTpRate",
			ID:   "RT_FS_USERS",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee:         12,
					Rate:               3,
					RateUnit:           "6s",
					RateIncrement:      "6s",
					GroupIntervalStart: "0s",
				},
				{
					ConnectFee:         12,
					Rate:               3,
					RateUnit:           "4s",
					RateIncrement:      "6s",
					GroupIntervalStart: "1s",
				},
			},
		},
	}
	if err := tpRatesMigrator.storDBIn.StorDB().SetTPRates(tpRates); err != nil {
		t.Error("Error when setting TpRate ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpRatesMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpRate ", err.Error())
	}
}

func testTpRatesITMove(t *testing.T) {
	err, _ := tpRatesMigrator.Migrate([]string{utils.MetaTpRates})
	if err != nil {
		t.Error("Error when migrating TpRate ", err.Error())
	}
}

func testTpRatesITCheckData(t *testing.T) {
	result, err := tpRatesMigrator.storDBOut.StorDB().GetTPRates(
		tpRates[0].TPid, tpRates[0].ID)
	if err != nil {
		t.Error("Error when getting TpRate ", err.Error())
	}
	if err := tpRates[0].RateSlots[0].SetDurations(); err != nil {
		t.Error(err)
	}
	if err := tpRates[0].RateSlots[1].SetDurations(); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(tpRates[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tpRates[0], result[0])
	}
	result, err = tpRatesMigrator.storDBIn.StorDB().GetTPRates(
		tpRates[0].TPid, tpRates[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
