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
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sCostPathIn   string
	sCostPathOut  string
	sCostCfgIn    *config.CGRConfig
	sCostCfgOut   *config.CGRConfig
	sCostMigrator *Migrator
	sCostAction   string
)

var sTestssCostIT = []func(t *testing.T){
	testSessionCostITConnect,
	testSessionCostITRename,
	testSessionCostITFlush,
	testSessionCostITMigrate,
}

func TestSessionCostITMongo(t *testing.T) {
	var err error
	sCostPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	sCostPathOut = path.Join(*dataDir, "conf", "samples", "tutmongojson")
	sCostCfgIn, err = config.NewCGRConfigFromFolder(sCostPathIn)
	if err != nil {
		t.Error(err)
	}
	sCostCfgOut, err = config.NewCGRConfigFromFolder(sCostPathOut)
	if err != nil {
		t.Error(err)
	}
	for _, stest := range sTestssCostIT {
		t.Run("TestSessionSCostITMigrateMongo", stest)
	}
}

func TestSessionCostITMySql(t *testing.T) {
	var err error
	sCostPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	sCostCfgIn, err = config.NewCGRConfigFromFolder(sCostPathIn)
	if err != nil {
		t.Error(err)
	}
	sCostCfgOut, err = config.NewCGRConfigFromFolder(sCostPathIn)
	if err != nil {
		t.Error(err)
	}
	for _, stest := range sTestssCostIT {
		t.Run("TestSessionSCostITMigrateMySql", stest)
	}
}

func testSessionCostITConnect(t *testing.T) {
	storDBIn, err := NewMigratorStorDB(sCostCfgIn.StorDbCfg().StorDBType,
		sCostCfgIn.StorDbCfg().StorDBHost, sCostCfgIn.StorDbCfg().StorDBPort,
		sCostCfgIn.StorDbCfg().StorDBName, sCostCfgIn.StorDbCfg().StorDBUser,
		sCostCfgIn.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		t.Error(err)
	}
	storDBOut, err := NewMigratorStorDB(sCostCfgOut.StorDbCfg().StorDBType,
		sCostCfgOut.StorDbCfg().StorDBHost, sCostCfgOut.StorDbCfg().StorDBPort,
		sCostCfgOut.StorDbCfg().StorDBName, sCostCfgOut.StorDbCfg().StorDBUser,
		sCostCfgOut.StorDbCfg().StorDBPass,
		config.CgrConfig().StorDbCfg().StorDBMaxOpenConns,
		config.CgrConfig().StorDbCfg().StorDBMaxIdleConns,
		config.CgrConfig().StorDbCfg().StorDBConnMaxLifetime,
		config.CgrConfig().StorDbCfg().StorDBCDRSIndexes)
	if err != nil {
		t.Error(err)
	}
	sCostMigrator, err = NewMigrator(nil, nil,
		storDBIn, storDBOut,
		false, false, false)
	if err != nil {
		t.Error(err)
	}
}

func testSessionCostITRename(t *testing.T) {
	var err error
	if err = sCostMigrator.storDBIn.createV1SMCosts(); err != nil {
		t.Error(err)
	}
	currentVersion := engine.Versions{
		utils.SessionSCosts: 1,
	}
	err = sCostMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for SessionsCosts ", err.Error())
	}
	if vrs, err := sCostMigrator.storDBOut.StorDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.SessionSCosts] != 1 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.SessionSCosts])
	}
	err, _ = sCostMigrator.Migrate([]string{utils.MetaSessionsCosts})
	if err != nil {
		t.Error("Error when migrating SessionsCosts ", err.Error())
	}
	if vrs, err := sCostMigrator.storDBOut.StorDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.SessionSCosts] != 2 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.SessionSCosts])
	}

}

func testSessionCostITFlush(t *testing.T) {
	if err := sCostMigrator.storDBOut.StorDB().Flush(
		path.Join(sCostCfgIn.DataFolderPath, "storage", sCostCfgIn.StorDbCfg().StorDBType)); err != nil {
		t.Error(err)
	}
}

func testSessionCostITMigrate(t *testing.T) {
	cc := &engine.CallCost{
		Direction:   utils.OUT,
		Cost:        1.23,
		Destination: "0723045326",
		Timespans: []*engine.TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &engine.RateInterval{
					Rating: &engine.RIRate{
						Rates: engine.RateGroups{
							&engine.Rate{
								GroupIntervalStart: 0,
								Value:              100,
								RateIncrement:      10 * time.Second,
								RateUnit:           time.Second,
							},
						},
					},
				},
			},
		},
		TOR: utils.VOICE,
	}
	v2Cost := &v2SessionsCost{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		RunID:       utils.DEFAULT_RUNID,
		Usage:       time.Duration(10),
		CostSource:  utils.MetaSessionS,
		CostDetails: cc,
	}
	var err error
	if err = sCostMigrator.storDBIn.setV2SMCost(v2Cost); err != nil {
		t.Error(err)
	}
	currentVersion := engine.Versions{
		utils.SessionSCosts: 2,
	}
	err = sCostMigrator.storDBOut.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for SessionsCosts ", err.Error())
	}
	if vrs, err := sCostMigrator.storDBOut.StorDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.SessionSCosts] != 2 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.SessionSCosts])
	}
	err, _ = sCostMigrator.Migrate([]string{utils.MetaSessionsCosts})
	if err != nil {
		t.Error("Error when migrating SessionsCosts ", err.Error())
	}
	if rcvCosts, err := sCostMigrator.storDBOut.StorDB().GetSMCosts("", utils.DEFAULT_RUNID, "", ""); err != nil {
		t.Error(err)
	} else if len(rcvCosts) != 1 {
		t.Errorf("Unexpected number of SessionsCosts returned: %d", len(rcvCosts))
	}
	if vrs, err := sCostMigrator.storDBOut.StorDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.SessionSCosts] != 3 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.SessionSCosts])
	}
}
