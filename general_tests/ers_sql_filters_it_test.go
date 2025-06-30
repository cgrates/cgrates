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
package general_tests

import (
	"bufio"
	"bytes"
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbConnString = "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'"
	timeStart    = time.Now()
	cdr1         = &engine.CDR{ // sample with values not realisticy calculated
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		ExtraInfo:   "extraInfo",
		Partial:     false,
		PreRated:    true,
		CostSource:  "cost source",
		CostDetails: &engine.EventCost{
			CGRID:     "test1",
			RunID:     utils.MetaDefault,
			StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
			Usage:     utils.DurationPointer(3 * time.Minute),
			Cost:      utils.Float64Pointer(2.3),
			Charges: []*engine.ChargingInterval{
				{
					RatingID: "c1a5ab9",
					Increments: []*engine.ChargingIncrement{
						{
							Usage:          2 * time.Minute,
							Cost:           2.0,
							AccountingID:   "a012888",
							CompressFactor: 1,
						},
						{
							Usage:          time.Second,
							Cost:           0.005,
							AccountingID:   "44d6c02",
							CompressFactor: 60,
						},
					},
					CompressFactor: 1,
				},
			},
			AccountSummary: &engine.AccountSummary{
				Tenant: "cgrates.org",
				ID:     "1001",
				BalanceSummaries: []*engine.BalanceSummary{
					{
						UUID:  "uuid1",
						Type:  utils.MetaMonetary,
						Value: 50,
					},
				},
				AllowNegative: false,
				Disabled:      false,
			},
			Rating: engine.Rating{
				"c1a5ab9": &engine.RatingUnit{
					ConnectFee:       0.1,
					RoundingMethod:   "*up",
					RoundingDecimals: 5,
					RatesID:          "ec1a177",
					RatingFiltersID:  "43e77dc",
				},
			},
			Accounting: engine.Accounting{
				"a012888": &engine.BalanceCharge{
					AccountID:   "cgrates.org:1001",
					BalanceUUID: "uuid1",
					Units:       120.7,
				},
				"44d6c02": &engine.BalanceCharge{
					AccountID:   "cgrates.org:1001",
					BalanceUUID: "uuid1",
					Units:       120.7,
				},
			},
			Rates: engine.ChargedRates{
				"ec1a177": engine.RateGroups{
					&engine.RGRate{
						GroupIntervalStart: 0,
						Value:              0.01,
						RateIncrement:      time.Minute,
						RateUnit:           time.Second},
				},
			},
		},
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	cgrID = utils.Sha1("oid2", timeStart.String())
	cdr2  = &engine.CDR{ // sample with values not realisticy calculated
		CGRID:       cgrID,
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "oid2",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   timeStart,
		AnswerTime:  timeStart,
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		ExtraInfo:   "extraInfo",
		Partial:     false,
		PreRated:    true,
		CostSource:  "cost source",
		CostDetails: &engine.EventCost{
			CGRID:     "test1",
			RunID:     utils.MetaDefault,
			StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
			Usage:     utils.DurationPointer(3 * time.Minute),
			Cost:      utils.Float64Pointer(2.3),
			Charges: []*engine.ChargingInterval{
				{
					RatingID: "RatingID2",
					Increments: []*engine.ChargingIncrement{
						{
							Usage:          2 * time.Minute,
							Cost:           2.0,
							AccountingID:   "a012888",
							CompressFactor: 1,
						},
						{
							Usage:          time.Second,
							Cost:           0.005,
							AccountingID:   "44d6c02",
							CompressFactor: 60,
						},
					},
					CompressFactor: 1,
				},
			},
			AccountSummary: &engine.AccountSummary{
				Tenant: "cgrates.org",
				ID:     "1001",
				BalanceSummaries: []*engine.BalanceSummary{
					{
						UUID:  "uuid1",
						Type:  utils.MetaMonetary,
						Value: 50,
					},
				},
				AllowNegative: false,
				Disabled:      false,
			},
			Rating: engine.Rating{
				"c1a5ab9": &engine.RatingUnit{
					ConnectFee:       0.1,
					RoundingMethod:   "*up",
					RoundingDecimals: 5,
					RatesID:          "ec1a177",
					RatingFiltersID:  "43e77dc",
				},
			},
			Accounting: engine.Accounting{
				"a012888": &engine.BalanceCharge{
					AccountID:   "cgrates.org:1001",
					BalanceUUID: "uuid1",
					Units:       120.7,
				},
				"44d6c02": &engine.BalanceCharge{
					AccountID:   "cgrates.org:1001",
					BalanceUUID: "uuid1",
					Units:       120.7,
				},
			},
			Rates: engine.ChargedRates{
				"ec1a177": engine.RateGroups{
					&engine.RGRate{
						GroupIntervalStart: 0,
						Value:              0.01,
						RateIncrement:      time.Minute,
						RateUnit:           time.Second},
				},
			},
		},
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
	}
	cdr3 = &engine.CDR{ // sample with values not realisticy calculated
		CGRID:       utils.Sha1("oid3", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "oid3",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		ExtraInfo:   "extraInfo",
		Partial:     false,
		PreRated:    true,
		CostSource:  "cost source",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
	}
)

func TestERSSQLFilters(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	var db, cdb2 *gorm.DB
	t.Run("InitSQLDB", func(t *testing.T) {
		var err error
		if cdb2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}

		if err = cdb2.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		sqlDB, err := cdb2.DB()
		if err != nil {
			t.Fatal(err)
		}
		sqlDB.SetConnMaxLifetime(5 * time.Second) // connections will stay idle even if you close the database. Set MaxLifetime to 5 seconds so that we dont get too many connection attempts error when ran with other tests togather
	})

	t.Run("PutCDRsInDataBase", func(t *testing.T) {
		var err error
		if db, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}
		tx := db.Begin()
		if !tx.Migrator().HasTable("cdrs") {
			if err = tx.Migrator().CreateTable(new(engine.CDRsql)); err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		tx.Commit()
		tx = db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql, err := cdr1.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql2, err := cdr2.AsCDRsql(&engine.JSONMarshaler{})
		cdrsql3, err := cdr3.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql.CreatedAt = time.Now()
		cdrSql2.CreatedAt = time.Now()
		cdrsql3.CreatedAt = time.Now()
		saved := tx.Save(cdrSql)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrSql2)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrsql3)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		tx.Commit()
		time.Sleep(10 * time.Millisecond)
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected table to have 3 results but got ", result)
		}
	})
	defer t.Run("StopSQL", func(t *testing.T) {
		if err := db.Migrator().DropTable("cdrs"); err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`DROP DATABASE cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		if db2, err := db.DB(); err != nil {
			t.Fatal(err)
		} else if err = db2.Close(); err != nil {
			t.Fatal(err)
		}
		if cdb2, err := cdb2.DB(); err != nil {
			t.Fatal(err)
		} else if err = cdb2.Close(); err != nil {
			t.Fatal(err)
		}
	})

	tpFiles := map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.cost_details.Charges[0].RatingID,RatingID2,
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql,`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_filters"),
		DBCfg:            dbcfg,
		TpFiles:          tpFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)
	time.Sleep(1 * time.Second)

	t.Run("VerifyProcessedFieldsFromLogs", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond) // give enough time to process from sql table
		records := 0
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		timeStartFormated := timeStart.Format("2006-01-02T15:04:05Z07:00")
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"CGRID\":\"%s\",\"Category\":\"call\",\"CostDetails\":\"{\\\"CGRID\\\":\\\"test1\\\",\\\"RunID\\\":\\\"*default\\\",\\\"StartTime\\\":\\\"2017-01-09T16:18:21Z\\\",\\\"Usage\\\":180000000000,\\\"Cost\\\":2.3,\\\"Charges\\\":[{\\\"RatingID\\\":\\\"RatingID2\\\",\\\"Increments\\\":[{\\\"Usage\\\":120000000000,\\\"Cost\\\":2,\\\"AccountingID\\\":\\\"a012888\\\",\\\"CompressFactor\\\":1},{\\\"Usage\\\":1000000000,\\\"Cost\\\":0.005,\\\"AccountingID\\\":\\\"44d6c02\\\",\\\"CompressFactor\\\":60}],\\\"CompressFactor\\\":1}],\\\"AccountSummary\\\":{\\\"Tenant\\\":\\\"cgrates.org\\\",\\\"ID\\\":\\\"1001\\\",\\\"BalanceSummaries\\\":[{\\\"UUID\\\":\\\"uuid1\\\",\\\"ID\\\":\\\"\\\",\\\"Type\\\":\\\"*monetary\\\",\\\"Initial\\\":0,\\\"Value\\\":50,\\\"Disabled\\\":false}],\\\"AllowNegative\\\":false,\\\"Disabled\\\":false},\\\"Rating\\\":{\\\"c1a5ab9\\\":{\\\"ConnectFee\\\":0.1,\\\"RoundingMethod\\\":\\\"*up\\\",\\\"RoundingDecimals\\\":5,\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"TimingID\\\":\\\"\\\",\\\"RatesID\\\":\\\"ec1a177\\\",\\\"RatingFiltersID\\\":\\\"43e77dc\\\"}},\\\"Accounting\\\":{\\\"44d6c02\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"},\\\"a012888\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"}},\\\"RatingFilters\\\":null,\\\"Rates\\\":{\\\"ec1a177\\\":[{\\\"GroupIntervalStart\\\":0,\\\"Value\\\":0.01,\\\"RateIncrement\\\":60000000000,\\\"RateUnit\\\":1000000000}]},\\\"Timings\\\":null}\",\"Destination\":\"1002\",\"OriginID\":\"oid2\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{}}>", timeStartFormated, cgrID, timeStartFormated)
		var ersLogsCount int
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, "<ERs> DRYRUN, reader: <mysql>") {
				continue
			}
			records++
			if !strings.Contains(line, expectedLog) {
				t.Errorf("expected \n<%q>, \nreceived\n<%q>", expectedLog, line)
			}
			if strings.Contains(line, "[INFO] <ERs> DRYRUN") {
				ersLogsCount++
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
		}
		if records != 1 {
			t.Errorf("expected ERs to process 1 records, but it processed %d records", records)
		}
		if ersLogsCount != 1 {
			t.Error("Expected only 1 ERS Dryrun log, received: ", ersLogsCount)
		}
	})

	t.Run("VerifyRowsNotDeleted", func(t *testing.T) {
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected 3 rows in table, got: ", result)
		}
		var rslt []map[string]any
		if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rslt).Error; err != nil {
			t.Fatalf("failed to query table: %v", err)
		}
	})
}

func TestERSSQLFiltersDeleteIndexedFields(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	var db, cdb2 *gorm.DB
	t.Run("InitSQLDB", func(t *testing.T) {
		var err error
		if cdb2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}

		if err = cdb2.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		sqlDB, err := cdb2.DB()
		if err != nil {
			t.Fatal(err)
		}
		sqlDB.SetConnMaxLifetime(5 * time.Second) // connections will stay idle even if you close the database. Set MaxLifetime to 5 seconds so that we dont get too many connection attempts error when ran with other tests togather

	})

	t.Run("PutCDRsInDataBase", func(t *testing.T) {
		var err error
		if db, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}
		tx := db.Begin()
		if !tx.Migrator().HasTable("cdrs") {
			if err = tx.Migrator().CreateTable(new(engine.CDRsql)); err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		tx.Commit()
		tx = db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql, err := cdr1.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql2, err := cdr2.AsCDRsql(&engine.JSONMarshaler{})
		cdrsql3, err := cdr3.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql.CreatedAt = time.Now()
		cdrSql2.CreatedAt = time.Now()
		cdrsql3.CreatedAt = time.Now()
		saved := tx.Save(cdrSql)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrSql2)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrsql3)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		tx.Commit()
		time.Sleep(10 * time.Millisecond)
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected table to have 3 results but got ", result)
		}
	})
	defer t.Run("StopSQL", func(t *testing.T) {
		if err := db.Migrator().DropTable("cdrs"); err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`DROP DATABASE cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		if db2, err := db.DB(); err != nil {
			t.Fatal(err)
		} else if err = db2.Close(); err != nil {
			t.Fatal(err)
		}
		if cdb2, err := cdb2.DB(); err != nil {
			t.Fatal(err)
		} else if err = cdb2.Close(); err != nil {
			t.Fatal(err)
		}

	})

	tpFiles := map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.cost_details.Charges[0].RatingID,RatingID2,
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql,`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_delete_indexed_fields"),
		DBCfg:            dbcfg,
		TpFiles:          tpFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)
	time.Sleep(1 * time.Second)

	t.Run("VerifyProcessedFieldsFromLogs", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond) // give enough time to process from sql table
		records := 0
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		timeStartFormated := timeStart.Format("2006-01-02T15:04:05Z07:00")
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"CGRID\":\"%s\",\"Category\":\"call\",\"CostDetails\":\"{\\\"CGRID\\\":\\\"test1\\\",\\\"RunID\\\":\\\"*default\\\",\\\"StartTime\\\":\\\"2017-01-09T16:18:21Z\\\",\\\"Usage\\\":180000000000,\\\"Cost\\\":2.3,\\\"Charges\\\":[{\\\"RatingID\\\":\\\"RatingID2\\\",\\\"Increments\\\":[{\\\"Usage\\\":120000000000,\\\"Cost\\\":2,\\\"AccountingID\\\":\\\"a012888\\\",\\\"CompressFactor\\\":1},{\\\"Usage\\\":1000000000,\\\"Cost\\\":0.005,\\\"AccountingID\\\":\\\"44d6c02\\\",\\\"CompressFactor\\\":60}],\\\"CompressFactor\\\":1}],\\\"AccountSummary\\\":{\\\"Tenant\\\":\\\"cgrates.org\\\",\\\"ID\\\":\\\"1001\\\",\\\"BalanceSummaries\\\":[{\\\"UUID\\\":\\\"uuid1\\\",\\\"ID\\\":\\\"\\\",\\\"Type\\\":\\\"*monetary\\\",\\\"Initial\\\":0,\\\"Value\\\":50,\\\"Disabled\\\":false}],\\\"AllowNegative\\\":false,\\\"Disabled\\\":false},\\\"Rating\\\":{\\\"c1a5ab9\\\":{\\\"ConnectFee\\\":0.1,\\\"RoundingMethod\\\":\\\"*up\\\",\\\"RoundingDecimals\\\":5,\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"TimingID\\\":\\\"\\\",\\\"RatesID\\\":\\\"ec1a177\\\",\\\"RatingFiltersID\\\":\\\"43e77dc\\\"}},\\\"Accounting\\\":{\\\"44d6c02\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"},\\\"a012888\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"}},\\\"RatingFilters\\\":null,\\\"Rates\\\":{\\\"ec1a177\\\":[{\\\"GroupIntervalStart\\\":0,\\\"Value\\\":0.01,\\\"RateIncrement\\\":60000000000,\\\"RateUnit\\\":1000000000}]},\\\"Timings\\\":null}\",\"Destination\":\"1002\",\"OriginID\":\"oid2\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{}}>", timeStartFormated, cgrID, timeStartFormated)
		var ersLogsCount int
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, "<ERs> DRYRUN, reader: <mysql>") {
				continue
			}
			records++
			if !strings.Contains(line, expectedLog) {
				t.Errorf("expected \n<%q>, \nreceived\n<%q>", expectedLog, line)
			}
			if strings.Contains(line, "[INFO] <ERs> DRYRUN") {
				ersLogsCount++
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
		}
		if records != 1 {
			t.Errorf("expected ERs to process 1 records, but it processed %d records", records)
		}
		if ersLogsCount != 1 {
			t.Error("Expected only 1 ERS Dryrun log, received: ", ersLogsCount)
		}
	})

	t.Run("VerifyRowsNotDeleted", func(t *testing.T) {
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 2 {
			t.Fatal("Expected 2 rows in table ", result)
		}
		var rslt []map[string]any
		if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rslt).Error; err != nil {
			t.Fatalf("failed to query table: %v", err)
		}

		// Print the entire table as a string
		for _, row := range rslt {
			for col, value := range row {
				if strings.Contains(fmt.Sprintln(value), "RatingID2") {
					t.Fatalf("Expected CDR with RatingID: \"RatingID2\" to be deleted. Received column <%q>, value <%q>", col, value)
				}
			}
		}
	})
}
func TestERSSQLFiltersWithMetaDelete(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	var db, cdb2 *gorm.DB
	t.Run("InitSQLDB", func(t *testing.T) {
		var err error
		if cdb2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}

		if err = cdb2.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		sqlDB, err := cdb2.DB()
		if err != nil {
			t.Fatal(err)
		}
		sqlDB.SetConnMaxLifetime(5 * time.Second) // connections will stay idle even if you close the database. Set MaxLifetime to 5 seconds so that we dont get too many connection attempts error when ran with other tests togather
	})

	t.Run("PutCDRsInDataBase", func(t *testing.T) {
		var err error
		if db, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}
		tx := db.Begin()
		if !tx.Migrator().HasTable("cdrs") {
			if err = tx.Migrator().CreateTable(new(engine.CDRsql)); err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		tx.Commit()
		tx = db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql, err := cdr1.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql2, err := cdr2.AsCDRsql(&engine.JSONMarshaler{})
		cdrsql3, err := cdr3.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql.CreatedAt = time.Now()
		cdrSql2.CreatedAt = time.Now()
		cdrsql3.CreatedAt = time.Now()
		saved := tx.Save(cdrSql)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrSql2)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrsql3)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		tx.Commit()
		time.Sleep(10 * time.Millisecond)
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected table to have 3 results but got ", result)
		}
	})
	defer t.Run("StopSQL", func(t *testing.T) {
		if err := db.Migrator().DropTable("cdrs"); err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`DROP DATABASE cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		if db2, err := db.DB(); err != nil {
			t.Fatal(err)
		} else if err = db2.Close(); err != nil {
			t.Fatal(err)
		}
		if cdb2, err := cdb2.DB(); err != nil {
			t.Fatal(err)
		} else if err = cdb2.Close(); err != nil {
			t.Fatal(err)
		}
	})

	tpFiles := map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.cost_details.Charges[0].RatingID,RatingID2,
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql,`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_meta_delete"),
		DBCfg:            dbcfg,
		TpFiles:          tpFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)
	time.Sleep(1 * time.Second)

	t.Run("VerifyProcessedFieldsFromLogs", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond) // give enough time to process from sql table
		records := 0
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		timeStartFormated := timeStart.Format("2006-01-02T15:04:05Z07:00")
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"CGRID\":\"%s\",\"Category\":\"call\",\"CostDetails\":\"{\\\"CGRID\\\":\\\"test1\\\",\\\"RunID\\\":\\\"*default\\\",\\\"StartTime\\\":\\\"2017-01-09T16:18:21Z\\\",\\\"Usage\\\":180000000000,\\\"Cost\\\":2.3,\\\"Charges\\\":[{\\\"RatingID\\\":\\\"RatingID2\\\",\\\"Increments\\\":[{\\\"Usage\\\":120000000000,\\\"Cost\\\":2,\\\"AccountingID\\\":\\\"a012888\\\",\\\"CompressFactor\\\":1},{\\\"Usage\\\":1000000000,\\\"Cost\\\":0.005,\\\"AccountingID\\\":\\\"44d6c02\\\",\\\"CompressFactor\\\":60}],\\\"CompressFactor\\\":1}],\\\"AccountSummary\\\":{\\\"Tenant\\\":\\\"cgrates.org\\\",\\\"ID\\\":\\\"1001\\\",\\\"BalanceSummaries\\\":[{\\\"UUID\\\":\\\"uuid1\\\",\\\"ID\\\":\\\"\\\",\\\"Type\\\":\\\"*monetary\\\",\\\"Initial\\\":0,\\\"Value\\\":50,\\\"Disabled\\\":false}],\\\"AllowNegative\\\":false,\\\"Disabled\\\":false},\\\"Rating\\\":{\\\"c1a5ab9\\\":{\\\"ConnectFee\\\":0.1,\\\"RoundingMethod\\\":\\\"*up\\\",\\\"RoundingDecimals\\\":5,\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"TimingID\\\":\\\"\\\",\\\"RatesID\\\":\\\"ec1a177\\\",\\\"RatingFiltersID\\\":\\\"43e77dc\\\"}},\\\"Accounting\\\":{\\\"44d6c02\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"},\\\"a012888\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"}},\\\"RatingFilters\\\":null,\\\"Rates\\\":{\\\"ec1a177\\\":[{\\\"GroupIntervalStart\\\":0,\\\"Value\\\":0.01,\\\"RateIncrement\\\":60000000000,\\\"RateUnit\\\":1000000000}]},\\\"Timings\\\":null}\",\"Destination\":\"1002\",\"OriginID\":\"oid2\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{}}>", timeStartFormated, cgrID, timeStartFormated)
		var ersLogsCount int
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, "<ERs> DRYRUN, reader: <mysql>") {
				continue
			}
			records++
			if !strings.Contains(line, expectedLog) {
				t.Errorf("expected \n<%q>, \nreceived\n<%q>", expectedLog, line)
			}
			if strings.Contains(line, "[INFO] <ERs> DRYRUN") {
				ersLogsCount++
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
		}
		if records != 1 {
			t.Errorf("expected ERs to process 1 records, but it processed %d records", records)
		}
		if ersLogsCount != 1 {
			t.Error("Expected only 1 ERS Dryrun log, received: ", ersLogsCount)
		}
	})

	t.Run("VerifyRowsNotDeleted", func(t *testing.T) {
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 2 {
			t.Error("Expected 2 rows in table, got: ", result)
		}
		var rslt []map[string]any
		if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rslt).Error; err != nil {
			t.Fatalf("failed to query table: %v", err)
		}

		// Print the entire table as a string
		for _, row := range rslt {
			for col, value := range row {
				if strings.Contains(fmt.Sprintln(value), "RatingID2") {
					t.Fatalf("Expected CDR with RatingID: \"RatingID2\" to be deleted. Received column <%q>, value <%q>", col, value)
				}
			}
		}
	})
}

type mockTableName struct {
	ID          int64
	Cgrid       string
	RunID       string
	OriginHost  string
	Source      string
	OriginID    string
	TOR         string
	RequestType string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Destination string
	SetupTime   time.Time
	AnswerTime  *time.Time
	Usage       int64
	ExtraFields string
	CostSource  string
	Cost        float64
	CostDetails string
	ExtraInfo   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func (mtn mockTableName) TableName() string {
	return "cdrsProcessed"
}

func TestERSSQLFiltersMove(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	var db, cdb2 *gorm.DB
	t.Run("InitSQLDB", func(t *testing.T) {
		var err error
		if cdb2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}

		if err = cdb2.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		sqlDB, err := cdb2.DB()
		if err != nil {
			t.Fatal(err)
		}
		sqlDB.SetConnMaxLifetime(5 * time.Second) // connections will stay idle even if you close the database. Set MaxLifetime to 5 seconds so that we dont get too many connection attempts error when ran with other tests togather
	})
	t.Run("PutCDRsInDataBase", func(t *testing.T) {
		var err error
		if db, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}
		tx := db.Begin()
		if !tx.Migrator().HasTable("cdrs") {
			if err = tx.Migrator().CreateTable(new(engine.CDRsql)); err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		if !tx.Migrator().HasTable("cdrsProcessed") {
			if err = tx.Migrator().CreateTable(new(mockTableName)); err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		tx.Commit()
		tx = db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql, err := cdr1.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql2, err := cdr2.AsCDRsql(&engine.JSONMarshaler{})
		cdrsql3, err := cdr3.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql.CreatedAt = time.Now()
		cdrSql2.CreatedAt = time.Now()
		cdrsql3.CreatedAt = time.Now()
		saved := tx.Save(cdrSql)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrSql2)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrsql3)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		tx.Commit()
		time.Sleep(10 * time.Millisecond)
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected table to have 3 results but got ", result)
		}
	})
	defer t.Run("StopSQL", func(t *testing.T) {
		if err := db.Migrator().DropTable("cdrs"); err != nil {
			t.Fatal(err)
		}
		if err := db.Migrator().DropTable("cdrsProcessed"); err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`DROP DATABASE cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		if db2, err := db.DB(); err != nil {
			t.Fatal(err)
		} else if err = db2.Close(); err != nil {
			t.Fatal(err)
		}
		if cdb2, err := cdb2.DB(); err != nil {
			t.Fatal(err)
		} else if err = cdb2.Close(); err != nil {
			t.Fatal(err)
		}
	})

	tpFiles := map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.cost_details.Charges[0].RatingID,RatingID2,
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql,`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_move"),
		DBCfg:            dbcfg,
		TpFiles:          tpFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)
	time.Sleep(1 * time.Second)

	t.Run("VerifyProcessedFieldsFromLogs", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond) // give enough time to process from sql table
		records := 0
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		timeStartFormated := timeStart.Format("2006-01-02T15:04:05Z07:00")
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"CGRID\":\"%s\",\"Category\":\"call\",\"CostDetails\":\"{\\\"CGRID\\\":\\\"test1\\\",\\\"RunID\\\":\\\"*default\\\",\\\"StartTime\\\":\\\"2017-01-09T16:18:21Z\\\",\\\"Usage\\\":180000000000,\\\"Cost\\\":2.3,\\\"Charges\\\":[{\\\"RatingID\\\":\\\"RatingID2\\\",\\\"Increments\\\":[{\\\"Usage\\\":120000000000,\\\"Cost\\\":2,\\\"AccountingID\\\":\\\"a012888\\\",\\\"CompressFactor\\\":1},{\\\"Usage\\\":1000000000,\\\"Cost\\\":0.005,\\\"AccountingID\\\":\\\"44d6c02\\\",\\\"CompressFactor\\\":60}],\\\"CompressFactor\\\":1}],\\\"AccountSummary\\\":{\\\"Tenant\\\":\\\"cgrates.org\\\",\\\"ID\\\":\\\"1001\\\",\\\"BalanceSummaries\\\":[{\\\"UUID\\\":\\\"uuid1\\\",\\\"ID\\\":\\\"\\\",\\\"Type\\\":\\\"*monetary\\\",\\\"Initial\\\":0,\\\"Value\\\":50,\\\"Disabled\\\":false}],\\\"AllowNegative\\\":false,\\\"Disabled\\\":false},\\\"Rating\\\":{\\\"c1a5ab9\\\":{\\\"ConnectFee\\\":0.1,\\\"RoundingMethod\\\":\\\"*up\\\",\\\"RoundingDecimals\\\":5,\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"TimingID\\\":\\\"\\\",\\\"RatesID\\\":\\\"ec1a177\\\",\\\"RatingFiltersID\\\":\\\"43e77dc\\\"}},\\\"Accounting\\\":{\\\"44d6c02\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"},\\\"a012888\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"}},\\\"RatingFilters\\\":null,\\\"Rates\\\":{\\\"ec1a177\\\":[{\\\"GroupIntervalStart\\\":0,\\\"Value\\\":0.01,\\\"RateIncrement\\\":60000000000,\\\"RateUnit\\\":1000000000}]},\\\"Timings\\\":null}\",\"Destination\":\"1002\",\"OriginID\":\"oid2\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{}}>", timeStartFormated, cgrID, timeStartFormated)
		var ersLogsCount int
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, "<ERs> DRYRUN, reader: <mysql>") {
				continue
			}
			records++
			if !strings.Contains(line, expectedLog) {
				t.Errorf("expected \n<%q>, \nreceived\n<%q>", expectedLog, line)
			}
			if strings.Contains(line, "[INFO] <ERs> DRYRUN") {
				ersLogsCount++
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
		}
		if records != 1 {
			t.Errorf("expected ERs to process 1 records, but it processed %d records", records)
		}
		if ersLogsCount != 1 {
			t.Error("Expected only 1 ERS Dryrun log, received: ", ersLogsCount)
		}
	})

	t.Run("VerifyRowsCount", func(t *testing.T) {
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 2 {
			t.Fatal("Expected 2 rows in table, got: ", result)
		}
		var rslt []map[string]any
		if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rslt).Error; err != nil {
			t.Errorf("failed to query table: %v", err)
		}
		// Print the entire table as a string
		for _, row := range rslt {
			for col, value := range row {
				if strings.Contains(fmt.Sprintln(value), "RatingID2") {
					t.Fatalf("Expected CDR with RatingID: \"RatingID2\" to be deleted. Received column <%q>, value <%q>", col, value)
				}
			}
		}

		var result2 int64
		db.Table("cdrsProcessed").Count(&result2)
		if result2 != 1 {
			t.Fatal("Expected 1 rows in table, got: ", result2)
		}
		var rslt2 []map[string]any
		if err := db.Raw("SELECT * FROM " + "cdrsProcessed").Scan(&rslt2).Error; err != nil {
			t.Errorf("failed to query table: %v", err)
		}
		timeStartFormated := rslt2[0]["answer_time"]
		createdAt := rslt2[0]["created_at"]
		updatedAt := rslt2[0]["updated_at"]
		exp := fmt.Sprintf("map[account:1001 answer_time:%s category:call cgrid:%s cost:1.01 cost_details:{\"CGRID\":\"test1\",\"RunID\":\"*default\",\"StartTime\":\"2017-01-09T16:18:21Z\",\"Usage\":180000000000,\"Cost\":2.3,\"Charges\":[{\"RatingID\":\"RatingID2\",\"Increments\":[{\"Usage\":120000000000,\"Cost\":2,\"AccountingID\":\"a012888\",\"CompressFactor\":1},{\"Usage\":1000000000,\"Cost\":0.005,\"AccountingID\":\"44d6c02\",\"CompressFactor\":60}],\"CompressFactor\":1}],\"AccountSummary\":{\"Tenant\":\"cgrates.org\",\"ID\":\"1001\",\"BalanceSummaries\":[{\"UUID\":\"uuid1\",\"ID\":\"\",\"Type\":\"*monetary\",\"Initial\":0,\"Value\":50,\"Disabled\":false}],\"AllowNegative\":false,\"Disabled\":false},\"Rating\":{\"c1a5ab9\":{\"ConnectFee\":0.1,\"RoundingMethod\":\"*up\",\"RoundingDecimals\":5,\"MaxCost\":0,\"MaxCostStrategy\":\"\",\"TimingID\":\"\",\"RatesID\":\"ec1a177\",\"RatingFiltersID\":\"43e77dc\"}},\"Accounting\":{\"44d6c02\":{\"AccountID\":\"cgrates.org:1001\",\"BalanceUUID\":\"uuid1\",\"RatingID\":\"\",\"Units\":120.7,\"ExtraChargeID\":\"\"},\"a012888\":{\"AccountID\":\"cgrates.org:1001\",\"BalanceUUID\":\"uuid1\",\"RatingID\":\"\",\"Units\":120.7,\"ExtraChargeID\":\"\"}},\"RatingFilters\":null,\"Rates\":{\"ec1a177\":[{\"GroupIntervalStart\":0,\"Value\":0.01,\"RateIncrement\":60000000000,\"RateUnit\":1000000000}]},\"Timings\":null} cost_source:cost source created_at:%s deleted_at:<nil> destination:1002 extra_fields:{\"field_extr1\":\"val_extr1\",\"fieldextr2\":\"valextr2\"} extra_info:extraInfo id:2 origin_host:192.168.1.1 origin_id:oid2 request_type:*rated run_id:*default setup_time:%s source:test subject:1001 tenant:cgrates.org tor:*voice updated_at:%s usage:10000000000]", timeStartFormated, cgrID, createdAt, timeStartFormated, updatedAt)
		// Print the entire table as a string
		for _, row := range rslt2 {
			if !strings.Contains(fmt.Sprintf("%+v", row), exp) {
				t.Errorf("Expected <%v>, \nReceived <%v>", exp, fmt.Sprintf("%+v", row))
			}
		}
	})
}

func TestERSSQLFiltersUpdate(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	var db, cdb2 *gorm.DB
	t.Run("InitSQLDB", func(t *testing.T) {
		var err error
		if cdb2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}

		if err = cdb2.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		sqlDB, err := cdb2.DB()
		if err != nil {
			t.Fatal(err)
		}
		sqlDB.SetConnMaxLifetime(5 * time.Second) // connections will stay idle even if you close the database. Set MaxLifetime to 5 seconds so that we dont get too many connection attempts error when ran with other tests togather
	})

	t.Run("PutCDRsInDataBase", func(t *testing.T) {
		var err error
		if db, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}
		tx := db.Begin()
		if !tx.Migrator().HasTable("cdrs") {
			if err = tx.Migrator().CreateTable(new(engine.CDRsql)); err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		tx.Commit()
		tx = db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql, err := cdr1.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql2, err := cdr2.AsCDRsql(&engine.JSONMarshaler{})
		cdrsql3, err := cdr3.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql.CreatedAt = time.Now()
		cdrSql2.CreatedAt = time.Now()
		cdrsql3.CreatedAt = time.Now()
		saved := tx.Save(cdrSql)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrSql2)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrsql3)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		tx.Commit()
		time.Sleep(10 * time.Millisecond)
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected table to have 3 results but got ", result)
		}
	})
	defer t.Run("StopSQL", func(t *testing.T) {
		if err := db.Migrator().DropTable("cdrs"); err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`DROP DATABASE cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		if db2, err := db.DB(); err != nil {
			t.Fatal(err)
		} else if err = db2.Close(); err != nil {
			t.Fatal(err)
		}
		if cdb2, err := cdb2.DB(); err != nil {
			t.Fatal(err)
		} else if err = cdb2.Close(); err != nil {
			t.Fatal(err)
		}
	})

	tpFiles := map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.cost_details.Charges[0].RatingID,RatingID2,
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql,`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_update"),
		DBCfg:            dbcfg,
		TpFiles:          tpFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)
	time.Sleep(1 * time.Second)

	t.Run("VerifyProcessedFieldsFromLogs", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond) // give enough time to process from sql table
		records := 0
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		timeStartFormated := timeStart.Format("2006-01-02T15:04:05Z07:00")
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"CGRID\":\"%s\",\"Category\":\"call\",\"CostDetails\":\"{\\\"CGRID\\\":\\\"test1\\\",\\\"RunID\\\":\\\"*default\\\",\\\"StartTime\\\":\\\"2017-01-09T16:18:21Z\\\",\\\"Usage\\\":180000000000,\\\"Cost\\\":2.3,\\\"Charges\\\":[{\\\"RatingID\\\":\\\"RatingID2\\\",\\\"Increments\\\":[{\\\"Usage\\\":120000000000,\\\"Cost\\\":2,\\\"AccountingID\\\":\\\"a012888\\\",\\\"CompressFactor\\\":1},{\\\"Usage\\\":1000000000,\\\"Cost\\\":0.005,\\\"AccountingID\\\":\\\"44d6c02\\\",\\\"CompressFactor\\\":60}],\\\"CompressFactor\\\":1}],\\\"AccountSummary\\\":{\\\"Tenant\\\":\\\"cgrates.org\\\",\\\"ID\\\":\\\"1001\\\",\\\"BalanceSummaries\\\":[{\\\"UUID\\\":\\\"uuid1\\\",\\\"ID\\\":\\\"\\\",\\\"Type\\\":\\\"*monetary\\\",\\\"Initial\\\":0,\\\"Value\\\":50,\\\"Disabled\\\":false}],\\\"AllowNegative\\\":false,\\\"Disabled\\\":false},\\\"Rating\\\":{\\\"c1a5ab9\\\":{\\\"ConnectFee\\\":0.1,\\\"RoundingMethod\\\":\\\"*up\\\",\\\"RoundingDecimals\\\":5,\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"TimingID\\\":\\\"\\\",\\\"RatesID\\\":\\\"ec1a177\\\",\\\"RatingFiltersID\\\":\\\"43e77dc\\\"}},\\\"Accounting\\\":{\\\"44d6c02\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"},\\\"a012888\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"}},\\\"RatingFilters\\\":null,\\\"Rates\\\":{\\\"ec1a177\\\":[{\\\"GroupIntervalStart\\\":0,\\\"Value\\\":0.01,\\\"RateIncrement\\\":60000000000,\\\"RateUnit\\\":1000000000}]},\\\"Timings\\\":null}\",\"Destination\":\"1002\",\"ExtraInfo\":\"extraInfo\",\"Id\":\"2\",\"OriginID\":\"oid2\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{}}>", timeStartFormated, cgrID, timeStartFormated)
		var ersLogsCount int
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, "<ERs> DRYRUN, reader: <mysql>") {
				continue
			}
			records++
			if !strings.Contains(line, expectedLog) {
				t.Errorf("expected \n<%q>, \nreceived\n<%q>", expectedLog, line)
			}
			if strings.Contains(line, "[INFO] <ERs> DRYRUN") {
				ersLogsCount++
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
		}
		if records != 1 {
			t.Errorf("expected ERs to process 1 records, but it processed %d records", records)
		}
		if ersLogsCount != 1 {
			t.Error("Expected only 1 ERS Dryrun log, received: ", ersLogsCount)
		}
	})

	t.Run("VerifyRowsCount", func(t *testing.T) {
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected 3 rows in table, got: ", result)
		}
		var rslt []map[string]any
		if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rslt).Error; err != nil {
			t.Errorf("failed to query table: %v", err)
		}
		var countST int
		// Print the entire table as a string
		for _, row := range rslt {
			for col, value := range row {
				if col == "setup_time" {
					if value.(time.Time).UTC().Equal(time.Date(2018, 11, 27, 14, 21, 26, 0, time.Local).UTC()) {
						countST++
						if utils.IfaceAsString(row["account"]) != "extraInfo" {
							t.Errorf("Expected CDR to be updated with empty cost_details. Received value <%q>", utils.IfaceAsString(row["cost_details"]))
						}
					}
				}
			}
		}
		if countST != 1 {
			t.Errorf("Expected CDR with origin_id:oid2 to have updated setup_time: <%+v>. Received <%v> \nCounted <%v> with expected setup_time", time.Date(2018, 11, 27, 14, 21, 26, 0, time.Local).UTC(), utils.ToJSON(rslt), countST)
		}
	})
}

// Using the raw event to update the rows means you dont need to define in the reader's template fields, the fields which you want to send to EES since all of the fields read in that row will be sent to EES. It also means that the request field names used for exporter template fields will be as read on the row; meaning instead of ~*req.SetupTime, it will be written as ~*req.setup_time
func TestERSSQLFiltersRawUpdate(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	var db, cdb2 *gorm.DB
	t.Run("InitSQLDB", func(t *testing.T) {
		var err error
		if cdb2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
			&gorm.Config{
				AllowGlobalUpdate: true,
				Logger:            logger.Default.LogMode(logger.Silent),
			}); err != nil {
			t.Fatal(err)
		}

		if err = cdb2.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		sqlDB, err := cdb2.DB()
		if err != nil {
			t.Fatal(err)
		}
		sqlDB.SetConnMaxLifetime(5 * time.Second) // connections will stay idle even if you close the database. Set MaxLifetime to 5 seconds so that we dont get too many connection attempts error when ran with other tests togather
	})

	t.Run("PutCDRsInDataBase", func(t *testing.T) {
		var err error
		if db, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")),
			&gorm.Config{
				AllowGlobalUpdate: true,
				Logger:            logger.Default.LogMode(logger.Silent),
			}); err != nil {
			t.Fatal(err)
		}
		tx := db.Begin()
		if !tx.Migrator().HasTable("cdrs") {
			if err = tx.Migrator().CreateTable(new(engine.CDRsql)); err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		tx.Commit()
		tx = db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql, err := cdr1.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql2, err := cdr2.AsCDRsql(&engine.JSONMarshaler{})
		cdrsql3, err := cdr3.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql.CreatedAt = time.Now()
		cdrSql2.CreatedAt = time.Now()
		cdrsql3.CreatedAt = time.Now()
		saved := tx.Save(cdrSql)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrSql2)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrsql3)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		tx.Commit()
		time.Sleep(10 * time.Millisecond)
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected table to have 3 results but got ", result)
		}
	})
	defer t.Run("StopSQL", func(t *testing.T) {
		if err := db.Migrator().DropTable("cdrs"); err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`DROP DATABASE cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		if db2, err := db.DB(); err != nil {
			t.Fatal(err)
		} else if err = db2.Close(); err != nil {
			t.Fatal(err)
		}
		if cdb2, err := cdb2.DB(); err != nil {
			t.Fatal(err)
		} else if err = cdb2.Close(); err != nil {
			t.Fatal(err)
		}
	})

	tpFiles := map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.cost_details.Charges[0].RatingID,RatingID2,
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql,`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_raw_update"),
		DBCfg:            dbcfg,
		TpFiles:          tpFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)
	time.Sleep(1 * time.Second)

	t.Run("VerifyProcessedFieldsFromLogs", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond) // give enough time to process from sql table
		records := 0
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		timeStartFormated := timeStart.Format("2006-01-02T15:04:05Z07:00")
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"CGRID\":\"%s\",\"Category\":\"call\",\"CostDetails\":\"{\\\"CGRID\\\":\\\"test1\\\",\\\"RunID\\\":\\\"*default\\\",\\\"StartTime\\\":\\\"2017-01-09T16:18:21Z\\\",\\\"Usage\\\":180000000000,\\\"Cost\\\":2.3,\\\"Charges\\\":[{\\\"RatingID\\\":\\\"RatingID2\\\",\\\"Increments\\\":[{\\\"Usage\\\":120000000000,\\\"Cost\\\":2,\\\"AccountingID\\\":\\\"a012888\\\",\\\"CompressFactor\\\":1},{\\\"Usage\\\":1000000000,\\\"Cost\\\":0.005,\\\"AccountingID\\\":\\\"44d6c02\\\",\\\"CompressFactor\\\":60}],\\\"CompressFactor\\\":1}],\\\"AccountSummary\\\":{\\\"Tenant\\\":\\\"cgrates.org\\\",\\\"ID\\\":\\\"1001\\\",\\\"BalanceSummaries\\\":[{\\\"UUID\\\":\\\"uuid1\\\",\\\"ID\\\":\\\"\\\",\\\"Type\\\":\\\"*monetary\\\",\\\"Initial\\\":0,\\\"Value\\\":50,\\\"Disabled\\\":false}],\\\"AllowNegative\\\":false,\\\"Disabled\\\":false},\\\"Rating\\\":{\\\"c1a5ab9\\\":{\\\"ConnectFee\\\":0.1,\\\"RoundingMethod\\\":\\\"*up\\\",\\\"RoundingDecimals\\\":5,\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"TimingID\\\":\\\"\\\",\\\"RatesID\\\":\\\"ec1a177\\\",\\\"RatingFiltersID\\\":\\\"43e77dc\\\"}},\\\"Accounting\\\":{\\\"44d6c02\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"},\\\"a012888\\\":{\\\"AccountID\\\":\\\"cgrates.org:1001\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"}},\\\"RatingFilters\\\":null,\\\"Rates\\\":{\\\"ec1a177\\\":[{\\\"GroupIntervalStart\\\":0,\\\"Value\\\":0.01,\\\"RateIncrement\\\":60000000000,\\\"RateUnit\\\":1000000000}]},\\\"Timings\\\":null}\",\"Destination\":\"1002\",\"OriginID\":\"oid2\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{}}>", timeStartFormated, cgrID, timeStartFormated)
		var ersLogsCount int
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, "<ERs> DRYRUN, reader: <mysql>") {
				continue
			}
			records++
			if !strings.Contains(line, expectedLog) {
				t.Errorf("expected \n<%q>, \nreceived\n<%q>", expectedLog, line)
			}
			if strings.Contains(line, "[INFO] <ERs> DRYRUN") {
				ersLogsCount++
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
		}
		if records != 1 {
			t.Errorf("expected ERs to process 1 records, but it processed %d records", records)
		}
		if ersLogsCount != 1 {
			t.Error("Expected only 1 ERS Dryrun log, received: ", ersLogsCount)
		}
	})

	t.Run("VerifyRowsCount", func(t *testing.T) {
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected 3 rows in table, got: ", result)
		}
		var rslt []map[string]any
		if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rslt).Error; err != nil {
			t.Errorf("failed to query table: %v", err)
		}
		var countST int
		// Print the entire table as a string
		for _, row := range rslt {
			for col, value := range row {
				if col == "setup_time" {
					if value.(time.Time).UTC().Equal((time.Date(2018, 11, 27, 14, 21, 26, 0, time.Local).UTC())) {
						countST++
						if utils.IfaceAsString(row["account"]) != "extraInfo" {
							t.Errorf("Expected CDR to be updated with empty cost_details. Received value <%q>", utils.IfaceAsString(row["cost_details"]))
						}
					}
				}
			}
		}
		if countST != 1 {
			t.Errorf("Expected CDR with origin_id:oid2 to have updated setup_time: <%v>. Received <%v> \nCounted <%v> with expected setup_time", (time.Date(2018, 11, 27, 14, 21, 26, 0, time.Local).UTC()), utils.ToJSON(rslt), countST)
		}
	})
}

func TestERSSQLFiltersErr(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	var db, cdb2 *gorm.DB
	t.Run("InitSQLDB", func(t *testing.T) {
		var err error
		if cdb2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}

		if err = cdb2.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		sqlDB, err := cdb2.DB()
		if err != nil {
			t.Fatal(err)
		}
		sqlDB.SetConnMaxLifetime(5 * time.Second) // connections will stay idle even if you close the database. Set MaxLifetime to 5 seconds so that we dont get too many connection attempts error when ran with other tests togather
	})

	t.Run("PutCDRsInDataBase", func(t *testing.T) {
		var err error
		if db, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")),
			&gorm.Config{
				AllowGlobalUpdate: true,
			}); err != nil {
			t.Fatal(err)
		}
		tx := db.Begin()
		if !tx.Migrator().HasTable("cdrs") {
			if err = tx.Migrator().CreateTable(new(engine.CDRsql)); err != nil {
				tx.Rollback()
				t.Fatal(err)
			}
		}
		tx.Commit()
		tx = db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql, err := cdr1.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql2, err := cdr2.AsCDRsql(&engine.JSONMarshaler{})
		cdrsql3, err := cdr3.AsCDRsql(&engine.JSONMarshaler{})
		cdrSql.CreatedAt = time.Now()
		cdrSql2.CreatedAt = time.Now()
		cdrsql3.CreatedAt = time.Now()
		saved := tx.Save(cdrSql)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrSql2)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		saved = tx.Save(cdrsql3)
		if saved.Error != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		tx.Commit()
		time.Sleep(10 * time.Millisecond)
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 3 {
			t.Error("Expected table to have 3 results but got ", result)
		}
	})
	defer t.Run("StopSQL", func(t *testing.T) {
		if err := db.Migrator().DropTable("cdrs"); err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`DROP DATABASE cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
		if db2, err := db.DB(); err != nil {
			t.Fatal(err)
		} else if err = db2.Close(); err != nil {
			t.Fatal(err)
		}
		if cdb2, err := cdb2.DB(); err != nil {
			t.Fatal(err)
		} else if err = cdb2.Close(); err != nil {
			t.Fatal(err)
		}
	})

	tpFiles := map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.cost_details.Charges[0].RatingID,RatingID2,
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql,`,
	}

	jsonCfg := `{
"general": {
	"log_level": 7
},

"apiers": {
	"enabled": true
},
"filters": {			
	"apiers_conns": ["*localhost"]
},
"stor_db": {
	"opts": {
		"sqlConnMaxLifetime": "5s", // needed while running all integration tests
	},
},
"ers": {									
	"enabled": true,						
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "mysql",										
			"type": "*sql",							
			"run_delay": "1m",									
			"source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",					
			"opts": {
				"sqlDBName":"cgrates2",
				"sqlTableName":"cdrs", 
				"sqlBatchSize": 2,
				"sqlDeleteIndexedFields": ["id"],
			},
			"start_delay": "500ms", // wait for db to be populated before starting reader 
			"processed_path": "*delete",
			"tenant": "cgrates.org",							
			"filters": [
					"*gt:~*req.answer_time:-168h", // dont process cdrs with answer_time older than 7 days ago
					"FLTR_SQL_RatingID", // "*eq:~*req.cost_details.Charges[0].RatingID:RatingID2",
					"*string:~*vars.*readerID:mysql",
					"FLTR_VARS", // "*string:~*vars.*readerID:mysql",
					"*notempty:~*vars.*readerID:''", // invalid filter
			],
			"flags": ["*dryrun"],
			"fields":[
				{"tag": "CGRID", "path": "*cgreq.CGRID", "type": "*variable", "value": "~*req.cgrid", "mandatory": true},
				{"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.tor", "mandatory": true},
				{"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.origin_id", "mandatory": true},
				{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*variable", "value": "~*req.request_type", "mandatory": true},
				{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*variable", "value": "~*req.tenant", "mandatory": true},
				{"tag": "Category", "path": "*cgreq.Category", "type": "*variable", "value": "~*req.category", "mandatory": true},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account", "mandatory": true},
				{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.subject", "mandatory": true},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.destination", "mandatory": true},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable", "value": "~*req.setup_time", "mandatory": true},
				{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable", "value": "~*req.answer_time", "mandatory": true},
				{"tag": "CostDetails", "path": "*cgreq.CostDetails", "type": "*variable", "value": "~*req.cost_details", "mandatory": true},
				{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage", "mandatory": true},
			],
		},
	],
},

}`

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON:       jsonCfg,
		DBCfg:            dbcfg,
		TpFiles:          tpFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)
	time.Sleep(1 * time.Second)

	t.Run("VerifyProcessedFieldsFromLogs", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond) // give enough time to process from sql table
		foundErr := false
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		expectedLog := `[ERROR] <ERs> error: <value of filter <*notempty:~*vars.*readerID:''> is not empty <''>>`
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "[ERROR] <ERs> error: <value of filter <*notempty:~*vars.*readerID:''> is not empty <''>>") {
				foundErr = true
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
		}
		if !foundErr {
			t.Errorf("expected error log <%s> \nreceived <%s>", expectedLog, buf)
		}
	})

}
