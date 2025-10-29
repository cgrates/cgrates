//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	sqlEeConfigDir string
	sqlEeCfgPath   string
	sqlEeCfg       *config.CGRConfig
	sqlEeRpc       *birpc.Client
	db2            *gorm.DB
	dbConnString   = "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'"

	sTestsSqlEe = []func(t *testing.T){
		testCreateDirectory,
		testSqlEeCreateTable,
		testSqlEeLoadConfig,
		testSqlEeResetDataDB,
		testSqlEeStartEngine,
		testSqlEeRPCConn,
		testSqlEeExportEventFull,
		testSqlEeVerifyExportedEvent,
		testSqlEeExportEventPartial,
		testSqlEeVerifyExportedEvent2,
		testStopCgrEngine,
		testCleanDirectory,
	}
)

func TestSqlEeExport(t *testing.T) {
	sqlEeConfigDir = "ees"
	for _, stest := range sTestsSqlEe {
		t.Run(sqlEeConfigDir, stest)
	}
}

// create a struct serve as model for *sql exporter
type testModelSql struct {
	Cgrid       string
	AnswerTime  time.Time
	Usage       int64
	Cost        float64
	CostDetails string
}

func (*testModelSql) TableName() string {
	return "expTable"
}

func testSqlEeCreateTable(t *testing.T) {
	var err error

	if db2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")), &gorm.Config{
		AllowGlobalUpdate: true,
		Logger:            logger.Default.LogMode(logger.Silent),
	}); err != nil {
		return
	}
	if err = db2.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
		t.Fatal(err)
	}

	if db2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")), &gorm.Config{
		AllowGlobalUpdate: true,
		Logger:            logger.Default.LogMode(logger.Silent),
	}); err != nil {
		return
	}
	tx := db2.Begin()
	if tx.Migrator().HasTable("expTable") {
		if err = tx.Migrator().DropTable(new(testModelSql)); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
	}
	if err = tx.Migrator().CreateTable(new(testModelSql)); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	tx.Commit()
}

func testSqlEeLoadConfig(t *testing.T) {
	var err error
	sqlEeCfgPath = path.Join(*utils.DataDir, "conf", "samples", sqlEeConfigDir)
	if sqlEeCfg, err = config.NewCGRConfigFromPath(sqlEeCfgPath); err != nil {
		t.Error(err)
	}
}

func testSqlEeResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(sqlEeCfg); err != nil {
		t.Fatal(err)
	}
}

func testSqlEeStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sqlEeCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSqlEeRPCConn(t *testing.T) {
	var err error
	sqlEeRpc, err = newRPCClient(sqlEeCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testSqlEeExportEventFull(t *testing.T) {
	cd := &engine.EventCost{
		Cost:  utils.Float64Pointer(0.264933),
		CGRID: "d8534def2b7067f4f5ad4f7ec7bbcc94bb46111a",
		Rates: engine.ChargedRates{
			"3db483c": engine.RateGroups{
				{
					Value:              0.1574,
					RateUnit:           60000000000,
					RateIncrement:      30000000000,
					GroupIntervalStart: 0,
				},
				{
					Value:              0.1574,
					RateUnit:           60000000000,
					RateIncrement:      1000000000,
					GroupIntervalStart: 30000000000,
				},
			},
		},
		RunID: "*default",
		Usage: utils.DurationPointer(101 * time.Second),
		Rating: engine.Rating{
			"7f3d423": &engine.RatingUnit{
				MaxCost:          40,
				RatesID:          "3db483c",
				TimingID:         "128e970",
				ConnectFee:       0,
				RoundingMethod:   "*up",
				MaxCostStrategy:  "*disconnect",
				RatingFiltersID:  "f8e95f2",
				RoundingDecimals: 4,
			},
		},
		Charges: []*engine.ChargingInterval{
			{
				RatingID: "7f3d423",
				Increments: []*engine.ChargingIncrement{
					{
						Cost:           0.0787,
						Usage:          30000000000,
						AccountingID:   "fee8a3a",
						CompressFactor: 1,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "7f3d423",
				Increments: []*engine.ChargingIncrement{
					{
						Cost:           0.002623,
						Usage:          1000000000,
						AccountingID:   "3463957",
						CompressFactor: 71,
					},
				},
				CompressFactor: 1,
			},
		},
		Timings: engine.ChargedTimings{
			"128e970": &engine.ChargedTiming{
				StartTime: "00:00:00",
			},
		},
		StartTime: time.Date(2019, 12, 06, 11, 57, 32, 0, time.UTC),
		Accounting: engine.Accounting{
			"3463957": &engine.BalanceCharge{
				Units:         0.002623,
				RatingID:      "",
				AccountID:     "cgrates.org:1001",
				BalanceUUID:   "154419f2-45e0-4629-a203-06034ccb493f",
				ExtraChargeID: "",
			},
			"fee8a3a": &engine.BalanceCharge{
				Units:         0.0787,
				RatingID:      "",
				AccountID:     "cgrates.org:1001",
				BalanceUUID:   "154419f2-45e0-4629-a203-06034ccb493f",
				ExtraChargeID: "",
			},
		},
		RatingFilters: engine.RatingFilters{
			"f8e95f2": engine.RatingMatchedFilters{
				"Subject":           "*out:cgrates.org:mo_call_UK_Mobile_O2_GBRCN:*any",
				"RatingPlanID":      "RP_MO_CALL_44800",
				"DestinationID":     "DST_44800",
				"DestinationPrefix": "44800",
			},
		},
		AccountSummary: &engine.AccountSummary{
			ID:            "234189200129930",
			Tenant:        "cgrates.org",
			Disabled:      false,
			AllowNegative: false,
			BalanceSummaries: engine.BalanceSummaries{
				&engine.BalanceSummary{
					ID:       "MOBILE_DATA",
					Type:     "*data",
					UUID:     "08a05723-5849-41b9-b6a9-8ee362539280",
					Value:    3221225472,
					Disabled: false,
				},
				&engine.BalanceSummary{
					ID:       "MOBILE_SMS",
					Type:     "*sms",
					UUID:     "06a87f20-3774-4eeb-826e-a79c5f175fd3",
					Value:    247,
					Disabled: false,
				},
				&engine.BalanceSummary{
					ID:       "MOBILE_VOICE",
					Type:     "*voice",
					UUID:     "4ad16621-6e22-4e35-958e-5e1ff93ad7b7",
					Value:    14270000000000,
					Disabled: false,
				},
				&engine.BalanceSummary{
					ID:       "MONETARY_POSTPAID",
					Type:     "*monetary",
					UUID:     "154419f2-45e0-4629-a203-06034ccb493f",
					Value:    50,
					Disabled: false,
				},
			},
		},
	}
	eventVoice := &engine.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]any{
				utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         1.01,
				utils.CostDetails:  utils.ToJSON(cd),
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := sqlEeRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
}

func testSqlEeExportEventPartial(t *testing.T) {
	eventVoice := &engine.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterPartial"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]any{
				utils.CGRID:        utils.Sha1("asd", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         123,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := sqlEeRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
}

func testSqlEeVerifyExportedEvent(t *testing.T) {
	var result int64
	db2.Table("expTable").Count(&result)
	if result != 1 {
		t.Fatal("Expected table to have only one result ", result)
	}
}

func testSqlEeVerifyExportedEvent2(t *testing.T) {
	var result int64
	db2.Table("expTable").Count(&result)
	if result != 2 {
		t.Fatal("Expected table to have only one result ", result)
	}
}

func TestOpenDB1(t *testing.T) {
	dialect := mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		"cgrates", "CGRateS.org", "127.0.0.1", "3306", "cgrates"))
	_, _, err := openDB(dialect, &config.SQLOpts{
		MaxIdleConns: utils.IntPointer(2),
	})
	if err != nil {
		t.Error(err)
	}
}

func TestOpenDB2(t *testing.T) {
	dialect := mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		"cgrates", "CGRateS.org", "127.0.0.1", "3306", "cgrates"))
	_, _, err := openDB(dialect, &config.SQLOpts{
		MaxOpenConns: utils.IntPointer(2),
	})
	if err != nil {
		t.Error(err)
	}
}

func TestOpenDB3(t *testing.T) {
	dialect := mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		"cgrates", "CGRateS.org", "127.0.0.1", "3306", "cgrates"))
	_, _, err := openDB(dialect, &config.SQLOpts{
		ConnMaxLifetime: utils.DurationPointer(2),
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSQLExportEvent1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.TableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.DBName = utils.StringPointer("cgrates")
	cgrCfg.EEsCfg().Exporters[0].ExportPath = `mysql://cgrates:CGRateS.org@127.0.0.1:3306`
	sqlEe, err := NewSQLEe(cgrCfg.EEsCfg().Exporters[0], nil)
	if err != nil {
		t.Error(err)
	}
	if err := sqlEe.Connect(); err != nil {
		t.Fatal(err)
	}
	if err := sqlEe.ExportEvent(&sqlPosterRequest{Querry: "INSERT INTO cdrs VALUES (); ", Values: []any{}}, ""); err != nil {
		t.Error(err)
	}
	sqlEe.Close()
}
