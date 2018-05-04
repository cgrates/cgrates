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
	testCdrITFlush,
	testCdrITMigrateAndMove,
}

func TestCdrITMongoConnection(t *testing.T) {
	var err error
	cdrPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	cdrCfgIn, err = config.NewCGRConfigFromFolder(cdrPathIn)
	if err != nil {
		t.Error(err)
	}
	storDBIn, err := engine.ConfigureStorDB(cdrCfgIn.StorDBType, cdrCfgIn.StorDBHost,
		cdrCfgIn.StorDBPort, cdrCfgIn.StorDBName,
		cdrCfgIn.StorDBUser, cdrCfgIn.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		t.Error(err)
	}
	storDBOut, err := engine.ConfigureStorDB(cdrCfgIn.StorDBType,
		cdrCfgIn.StorDBHost, cdrCfgIn.StorDBPort, cdrCfgIn.StorDBName,
		cdrCfgIn.StorDBUser, cdrCfgIn.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		t.Error(err)
	}
	oldStorDB, err := ConfigureV1StorDB(cdrCfgIn.StorDBType,
		cdrCfgIn.StorDBHost, cdrCfgIn.StorDBPort, cdrCfgIn.StorDBName,
		cdrCfgIn.StorDBUser, cdrCfgIn.StorDBPass)
	if err != nil {
		log.Fatal(err)
	}

	cdrMigrator, err = NewMigrator(nil, nil, cdrCfgIn.DataDbType,
		cdrCfgIn.DBDataEncoding, storDBIn, storDBOut, cdrCfgIn.StorDBType, nil,
		cdrCfgIn.DataDbType, cdrCfgIn.DBDataEncoding, oldStorDB, cdrCfgIn.StorDBType,
		false, false, false, false, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCdrITMongo(t *testing.T) {
	for _, stest := range sTestsCdrIT {
		t.Run("TestCdrITMigrateMongo", stest)
	}
}

func testCdrITFlush(t *testing.T) {
	if err := cdrMigrator.storDBOut.Flush(
		path.Join(cdrCfgIn.DataFolderPath, "storage", cdrCfgIn.StorDBType)); err != nil {
		t.Error(err)
	}
}

func testCdrITMigrateAndMove(t *testing.T) {
	cc := &engine.CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*engine.TimeSpan{
			&engine.TimeSpan{
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
	v1Cdr := &v1Cdrs{
		CGRID:   utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		Source: utils.UNIT_TEST, RequestType: utils.META_RATED, Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.DEFAULT_RUNID, Usage: time.Duration(10),
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01, Rated: true,
		CostDetails: cc,
	}
	var err error
	if err = cdrMigrator.oldStorDB.setV1CDR(v1Cdr); err != nil {
		t.Error(err)
	}
	currentVersion := engine.Versions{
		utils.COST_DETAILS:  2,
		utils.SessionSCosts: 3,
		utils.CDRs:          1,
	}
	err = cdrMigrator.storDBOut.SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for CDRs ", err.Error())
	}
	if rcvCDRs, _, err := cdrMigrator.storDBOut.GetCDRs(new(utils.CDRsFilter), false); err != utils.ErrNotFound {
		t.Error(err)
	}
	err, _ = cdrMigrator.Migrate([]string{utils.MetaCDRs})
	if err != nil {
		t.Error("Error when migrating CDRs ", err.Error())
	}
	if rcvCDRs, _, err := cdrMigrator.storDBOut.GetCDRs(new(utils.CDRsFilter), false); err != nil {
		t.Error(err)
	} else if len(rcvCDRs) != 1 {
		t.Errorf("Unexpected number of CDRs returned: %d", len(rcvCDRs))
	}
}
