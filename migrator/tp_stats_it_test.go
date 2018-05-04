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
	tpStatsPathIn   string
	tpStatsPathOut  string
	tpStatsCfgIn    *config.CGRConfig
	tpStatsCfgOut   *config.CGRConfig
	tpStatsMigrator *Migrator
	tpStats         []*utils.TPStats
)

var sTestsTpStatsIT = []func(t *testing.T){
	testTpStatsITConnect,
	testTpStatsITFlush,
	testTpStatsITPopulate,
	testTpStatsITMove,
	testTpStatsITCheckData,
}

func TestTpStatsMove(t *testing.T) {
	for _, stest := range sTestsTpStatsIT {
		t.Run("TestTpStatsMove", stest)
	}
}

func testTpStatsITConnect(t *testing.T) {
	var err error
	tpStatsPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpStatsCfgIn, err = config.NewCGRConfigFromFolder(tpStatsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpStatsPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpStatsCfgOut, err = config.NewCGRConfigFromFolder(tpStatsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := engine.ConfigureStorDB(tpStatsCfgIn.StorDBType, tpStatsCfgIn.StorDBHost,
		tpStatsCfgIn.StorDBPort, tpStatsCfgIn.StorDBName,
		tpStatsCfgIn.StorDBUser, tpStatsCfgIn.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := engine.ConfigureStorDB(tpStatsCfgOut.StorDBType,
		tpStatsCfgOut.StorDBHost, tpStatsCfgOut.StorDBPort, tpStatsCfgOut.StorDBName,
		tpStatsCfgOut.StorDBUser, tpStatsCfgOut.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpStatsMigrator, err = NewMigrator(nil, nil, tpStatsCfgIn.DataDbType,
		tpStatsCfgIn.DBDataEncoding, storDBIn, storDBOut, tpStatsCfgIn.StorDBType, nil,
		tpStatsCfgIn.DataDbType, tpStatsCfgIn.DBDataEncoding, nil,
		tpStatsCfgIn.StorDBType, false, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpStatsITFlush(t *testing.T) {
	if err := tpStatsMigrator.storDBIn.Flush(
		path.Join(tpStatsCfgIn.DataFolderPath, "storage", tpStatsCfgIn.StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpStatsMigrator.storDBOut.Flush(
		path.Join(tpStatsCfgOut.DataFolderPath, "storage", tpStatsCfgOut.StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpStatsITPopulate(t *testing.T) {
	tpStats = []*utils.TPStats{
		&utils.TPStats{
			Tenant:    "cgrates.org",
			TPid:      "TPS1",
			ID:        "Stat1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			TTL: "1",
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{
					MetricID:   "*sum",
					Parameters: "Param1",
				},
			},
			Blocker:      false,
			Stored:       false,
			Weight:       20,
			MinItems:     1,
			ThresholdIDs: []string{"ThreshValue", "ThreshValueTwo"},
		},
	}
	if err := tpStatsMigrator.storDBIn.SetTPStats(tpStats); err != nil {
		t.Error("Error when setting TpFilter ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpStatsMigrator.storDBOut.SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpFilter ", err.Error())
	}
}

func testTpStatsITMove(t *testing.T) {
	err, _ := tpStatsMigrator.Migrate([]string{utils.MetaTpStats})
	if err != nil {
		t.Error("Error when migrating TpFilter ", err.Error())
	}
}

func testTpStatsITCheckData(t *testing.T) {
	result, err := tpStatsMigrator.storDBOut.GetTPStats(
		tpStats[0].TPid, tpStats[0].ID)
	if err != nil {
		t.Error("Error when getting TpFilter ", err.Error())
	}
	if !reflect.DeepEqual(tpStats[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tpStats[0], result[0])
	}
	result, err = tpStatsMigrator.storDBIn.GetTPStats(
		tpStats[0].TPid, tpStats[0].ID)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}
