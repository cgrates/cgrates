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
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdrPathIn   string
	cdrPathOut  string
	cdrCfgIn    *config.CGRConfig
	cdrCfgOut   *config.CGRConfig
	cdrMigrator *Migrator
	cdrAction   string
)

var sTestsCdrIT = []func(t *testing.T){
	testCdrITConnect,
	testCdrITFlush,
	testCdrITMigrateAndMove,
}

func TestCdrITMongo(t *testing.T) {
	var err error
	cdrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	cdrCfgIn, err = config.NewCGRConfigFromPath(cdrPathIn)
	if err != nil {
		t.Error(err)
	}
	for _, stest := range sTestsCdrIT {
		t.Run("TestCdrITMigrateMongo", stest)
	}
	cdrMigrator.Close()
}

func TestCdrITMySql(t *testing.T) {
	var err error
	cdrPathIn = path.Join(*dataDir, "conf", "samples", "tutmysql")
	cdrCfgIn, err = config.NewCGRConfigFromPath(cdrPathIn)
	if err != nil {
		t.Error(err)
	}
	for _, stest := range sTestsCdrIT {
		t.Run("TestCdrITMigrateMySql", stest)
	}
	cdrMigrator.Close()
}

func testCdrITConnect(t *testing.T) {
	storDBIn, err := NewMigratorStorDB(cdrCfgIn.StorDbCfg().Type,
		cdrCfgIn.StorDbCfg().Host, cdrCfgIn.StorDbCfg().Port,
		cdrCfgIn.StorDbCfg().Name, cdrCfgIn.StorDbCfg().User,
		cdrCfgIn.StorDbCfg().Password, cdrCfgIn.GeneralCfg().DBDataEncoding,
		cdrCfgIn.StorDbCfg().StringIndexedFields, cdrCfgIn.StorDbCfg().PrefixIndexedFields,
		cdrCfgIn.StorDbCfg().Opts)
	if err != nil {
		t.Error(err)
	}
	storDBOut, err := NewMigratorStorDB(cdrCfgIn.StorDbCfg().Type,
		cdrCfgIn.StorDbCfg().Host, cdrCfgIn.StorDbCfg().Port,
		cdrCfgIn.StorDbCfg().Name, cdrCfgIn.StorDbCfg().User,
		cdrCfgIn.StorDbCfg().Password, cdrCfgIn.GeneralCfg().DBDataEncoding,
		cdrCfgIn.StorDbCfg().StringIndexedFields, cdrCfgIn.StorDbCfg().PrefixIndexedFields,
		cdrCfgIn.StorDbCfg().Opts)
	if err != nil {
		t.Error(err)
	}

	cdrMigrator, err = NewMigrator(nil, nil,
		storDBIn, storDBOut,
		false, true, false, false)
	if err != nil {
		t.Error(err)
	}
}

func testCdrITFlush(t *testing.T) {
	if err := cdrMigrator.storDBOut.StorDB().Flush(
		path.Join(cdrCfgIn.DataFolderPath, "storage", cdrCfgIn.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testCdrITMigrateAndMove(t *testing.T) {
	cc := &engine.CallCost{
		Destination: "0723045326",
		Timespans: []*engine.TimeSpan{
			{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval: &engine.RateInterval{
					Rating: &engine.RIRate{
						Rates: engine.RateGroups{
							&engine.RGRate{
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
		ToR: utils.MetaVoice,
	}
	v1Cdr := &v1Cdrs{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UnitTest,
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
		Rated:       true,
		CostDetails: cc,
	}
	var err error
	if err = cdrMigrator.storDBIn.setV1CDR(v1Cdr); err != nil {
		t.Error(err)
	}
	currentVersion := engine.Versions{
		utils.CostDetails: 2,
		utils.CDRs:        1,
	}
	err = cdrMigrator.storDBIn.StorDB().SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for CDRs ", err.Error())
	}
	if vrs, err := cdrMigrator.storDBIn.StorDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.CDRs] != 1 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.CDRs])
	}
	err, _ = cdrMigrator.Migrate([]string{utils.MetaCDRs})
	if err != nil {
		t.Error("Error when migrating CDRs ", err.Error())
	}
	if rcvCDRs, _, err := cdrMigrator.storDBOut.StorDB().GetCDRs(new(utils.CDRsFilter), false); err != nil {
		t.Error(err)
	} else if len(rcvCDRs) != 1 {
		t.Errorf("Unexpected number of CDRs returned: %d", len(rcvCDRs))
	}
	if vrs, err := cdrMigrator.storDBOut.StorDB().GetVersions(""); err != nil {
		t.Error(err)
	} else if vrs[utils.CDRs] != 2 {
		t.Errorf("Unexpected version returned: %d", vrs[utils.CDRs])
	}
	//  else if cdrMigrator.stats[utils.CDRs] != 1 {
	// 	t.Errorf("Expected 1, received: %v", cdrMigrator.stats[utils.CDRs])
	// }
}
