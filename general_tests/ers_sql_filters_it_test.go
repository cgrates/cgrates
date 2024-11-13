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
	db           *gorm.DB
	dbConnString = "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'"
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
				ID:     "testV1CDRsRefundOutOfSessionCost",
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
					AccountID:   "cgrates.org:testV1CDRsRefundOutOfSessionCost",
					BalanceUUID: "uuid1",
					Units:       120.7,
				},
				"44d6c02": &engine.BalanceCharge{
					AccountID:   "cgrates.org:testV1CDRsRefundOutOfSessionCost",
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
	timeStart = time.Now()
	cgrID     = utils.Sha1("dsafdsaf", timeStart.String())
	cdr2      = &engine.CDR{ // sample with values not realisticy calculated
		CGRID:       cgrID,
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
				ID:     "testV1CDRsRefundOutOfSessionCost",
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
					AccountID:   "cgrates.org:testV1CDRsRefundOutOfSessionCost",
					BalanceUUID: "uuid1",
					Units:       120.7,
				},
				"44d6c02": &engine.BalanceCharge{
					AccountID:   "cgrates.org:testV1CDRsRefundOutOfSessionCost",
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

	t.Run("InitSQLDB", func(t *testing.T) {
		var err error
		var db2 *gorm.DB
		if db2, err = gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
			&gorm.Config{
				AllowGlobalUpdate: true,
				Logger:            logger.Default.LogMode(logger.Silent),
			}); err != nil {
			t.Fatal(err)
		}

		if err = db2.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
			t.Fatal(err)
		}
	})

	type testModelSql struct {
		ID          int64
		Cgrid       string
		RunID       string
		OriginHost  string
		Source      string
		OriginID    string
		ToR         string
		RequestType string
		Tenant      string
		Category    string
		Account     string
		Subject     string
		Destination string
		SetupTime   time.Time
		AnswerTime  time.Time
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
		cdrSql := cdr1.AsCDRsql()
		cdrSql2 := cdr2.AsCDRsql()
		cdrSql.CreatedAt = time.Now()
		cdrSql2.CreatedAt = time.Now()
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
		tx.Commit()
		time.Sleep(10 * time.Millisecond)
		var result int64
		db.Table(utils.CDRsTBL).Count(&result)
		if result != 2 {
			t.Error("Expected table to have only one result ", result)
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
	})

	content := `{

	"general": {
		"log_level": 7
	},
	
	"apiers": {
		"enabled": true
	},
	
	"ers": {									
		"enabled": true,						
		"sessions_conns":["*localhost"],
		"apiers_conns": ["*localhost"],
		"readers": [
			{
				"id": "mysql",										
				"type": "*sql",							
				"run_delay": "1m",									
				"source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",					
				"opts": {
					"sqlDBName":"cgrates2",
				},
				"processed_path": "",
				"tenant": "cgrates.org",							
				"filters": [
				"*gt:~*req.answer_time:NOW() - INTERVAL 7 DAY", // dont process cdrs with answer_time older than 7 days ago (continue if answer_time > now-7days)
				"*eq:~*req.cost_details.Charges[0].RatingID:RatingID2",
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
		ConfigJSON: content,
		DBCfg:      dbcfg,
		LogBuffer:  buf,
	}
	ng.Run(t)

	t.Run("VerifyProcessedFieldsFromLogs", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond) // give enough time to process from sql table
		records := 0
		scanner := bufio.NewScanner(strings.NewReader(buf.String()))
		timeStartFormated := timeStart.Format("2006-01-02T15:04:05-07:00")
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"CGRID\":\"%s\",\"Category\":\"call\",\"CostDetails\":\"{\\\"CGRID\\\":\\\"test1\\\",\\\"RunID\\\":\\\"*default\\\",\\\"StartTime\\\":\\\"2017-01-09T16:18:21Z\\\",\\\"Usage\\\":180000000000,\\\"Cost\\\":2.3,\\\"Charges\\\":[{\\\"RatingID\\\":\\\"RatingID2\\\",\\\"Increments\\\":[{\\\"Usage\\\":120000000000,\\\"Cost\\\":2,\\\"AccountingID\\\":\\\"a012888\\\",\\\"CompressFactor\\\":1},{\\\"Usage\\\":1000000000,\\\"Cost\\\":0.005,\\\"AccountingID\\\":\\\"44d6c02\\\",\\\"CompressFactor\\\":60}],\\\"CompressFactor\\\":1}],\\\"AccountSummary\\\":{\\\"Tenant\\\":\\\"cgrates.org\\\",\\\"ID\\\":\\\"testV1CDRsRefundOutOfSessionCost\\\",\\\"BalanceSummaries\\\":[{\\\"UUID\\\":\\\"uuid1\\\",\\\"ID\\\":\\\"\\\",\\\"Type\\\":\\\"*monetary\\\",\\\"Initial\\\":0,\\\"Value\\\":50,\\\"Disabled\\\":false}],\\\"AllowNegative\\\":false,\\\"Disabled\\\":false},\\\"Rating\\\":{\\\"c1a5ab9\\\":{\\\"ConnectFee\\\":0.1,\\\"RoundingMethod\\\":\\\"*up\\\",\\\"RoundingDecimals\\\":5,\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"TimingID\\\":\\\"\\\",\\\"RatesID\\\":\\\"ec1a177\\\",\\\"RatingFiltersID\\\":\\\"43e77dc\\\"}},\\\"Accounting\\\":{\\\"44d6c02\\\":{\\\"AccountID\\\":\\\"cgrates.org:testV1CDRsRefundOutOfSessionCost\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"},\\\"a012888\\\":{\\\"AccountID\\\":\\\"cgrates.org:testV1CDRsRefundOutOfSessionCost\\\",\\\"BalanceUUID\\\":\\\"uuid1\\\",\\\"RatingID\\\":\\\"\\\",\\\"Units\\\":120.7,\\\"ExtraChargeID\\\":\\\"\\\"}},\\\"RatingFilters\\\":null,\\\"Rates\\\":{\\\"ec1a177\\\":[{\\\"GroupIntervalStart\\\":0,\\\"Value\\\":0.01,\\\"RateIncrement\\\":60000000000,\\\"RateUnit\\\":1000000000}]},\\\"Timings\\\":null}\",\"Destination\":\"1002\",\"OriginID\":\"dsafdsaf\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{}}>", timeStartFormated, cgrID, timeStartFormated)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, "<ERs> DRYRUN, reader: <mysql>") {
				continue
			}
			records++
			if !strings.Contains(line, expectedLog) {
				t.Errorf("expected \n<%s>, \nreceived\n<%s>", expectedLog, line)
			}
		}
		if err := scanner.Err(); err != nil {
			t.Errorf("error reading input: %v", err)
		}
		if records != 1 {
			t.Errorf("expected ERs to process 1 records, but it processed %d records", records)
		}
	})
}
