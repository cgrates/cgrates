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
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	cdr1         = &utils.CDR{ // sample with values not realisticy calculated
		Tenant: "cgrates.org",
		Opts: map[string]any{
			utils.MetaCDRID:      utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			utils.OptsCDRsExport: false,
			utils.MetaChargeID:   utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaOriginID:   "dsafdsaf",
			utils.MetaRateSCost: &utils.RateProfileCost{
				Cost: utils.NewDecimalFromFloat64(2.3),
				CostIntervals: []*utils.RateSIntervalCost{
					{
						Increments: []*utils.RateSIncrementCost{
							{
								CompressFactor:    1,
								Usage:             utils.NewDecimalFromUsageIgnoreErr("2m"),
								RateID:            "c1a5ab9",
								RateIntervalIndex: 0,
							},
							{
								CompressFactor:    60,
								Usage:             utils.NewDecimalFromUsageIgnoreErr("1s"),
								RateID:            "c1a5ab9",
								RateIntervalIndex: 1,
							},
						},
						CompressFactor: 1,
					},
				},
				ID:      "DEFAULT_RATE",
				MaxCost: utils.NewDecimalFromFloat64(0),
				MinCost: utils.NewDecimalFromFloat64(0),
				Rates: map[string]*utils.IntervalRate{
					"c1a5ab9": {
						FixedFee:      utils.NewDecimalFromFloat64(0.1),
						Increment:     utils.NewDecimalFromUsageIgnoreErr("1m"),
						IntervalStart: utils.NewDecimalFromFloat64(0),
						RecurrentFee:  utils.NewDecimalFromFloat64(0.01),
						Unit:          utils.NewDecimalFromUsageIgnoreErr("1s"),
					},
				},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  utils.MetaDefault,
			utils.MetaSubsys: utils.MetaChargers,
			utils.MetaUsage:  10 * time.Second,
		},
		Event: map[string]any{
			utils.OrderID:      123,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "dsafdsaf",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "test",
			utils.RequestType:  utils.MetaRated,
			utils.Category:     "call",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			utils.AnswerTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			utils.Usage:        10 * time.Second,
			utils.ExtraInfo:    "extraInfo",
			utils.ExtraFields:  map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
	}
	cdrID = utils.Sha1("oid2", timeStart.String())
	cdr2  = &utils.CDR{ // sample with values not realisticy calculated
		Tenant: "cgrates.org",
		Opts: map[string]any{
			utils.MetaCDRID:      cdrID,
			utils.OptsCDRsExport: false,
			utils.MetaChargeID:   cdrID,
			utils.MetaChargers:   true,
			utils.MetaCost:       1.01,
			utils.MetaOriginID:   "dsafdsaf",
			utils.MetaRateSCost: &utils.RateProfileCost{
				Cost: utils.NewDecimalFromFloat64(2.3),
				CostIntervals: []*utils.RateSIntervalCost{
					{
						Increments: []*utils.RateSIncrementCost{
							{
								CompressFactor:    1,
								Usage:             utils.NewDecimalFromUsageIgnoreErr("2m"),
								RateID:            "RateID2",
								RateIntervalIndex: 0,
							},
							{
								CompressFactor:    60,
								Usage:             utils.NewDecimalFromUsageIgnoreErr("1s"),
								RateID:            "RateID2",
								RateIntervalIndex: 1,
							},
						},
						CompressFactor: 1,
					},
				},
				ID:      "DEFAULT_RATE",
				MaxCost: utils.NewDecimalFromFloat64(0),
				MinCost: utils.NewDecimalFromFloat64(0),
				Rates: map[string]*utils.IntervalRate{
					"RateID2": {
						FixedFee:      utils.NewDecimalFromFloat64(0.1),
						Increment:     utils.NewDecimalFromUsageIgnoreErr("1m"),
						IntervalStart: utils.NewDecimalFromFloat64(0),
						RecurrentFee:  utils.NewDecimalFromFloat64(0.01),
						Unit:          utils.NewDecimalFromUsageIgnoreErr("1s"),
					},
				},
			},
			utils.MetaRates:  true,
			utils.MetaRunID:  utils.MetaDefault,
			utils.MetaSubsys: utils.MetaChargers,
			utils.MetaUsage:  10 * time.Second,
		},
		Event: map[string]any{
			utils.OrderID:      123,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "oid2",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "test",
			utils.RequestType:  utils.MetaRated,
			utils.Category:     "call",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    timeStart,
			utils.AnswerTime:   timeStart,
			utils.Usage:        10 * time.Second,
			utils.ExtraInfo:    "extraInfo",
			utils.ExtraFields:  map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
	}
	cdr3 = &utils.CDR{ // sample with values not realisticy calculated
		Tenant: "cgrates.org",
		Opts: map[string]any{
			utils.MetaCDRID:      utils.Sha1("oid3", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			utils.OptsCDRsExport: false,
			utils.MetaChargeID:   utils.Sha1("oid3", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			utils.MetaCost:       1.01,
			utils.MetaOriginID:   "dsafdsaf",
			utils.MetaRates:      true,
			utils.MetaRunID:      utils.MetaDefault,
			utils.MetaUsage:      10 * time.Second,
		},
		Event: map[string]any{
			utils.OrderID:      123,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "oid3",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "test",
			utils.RequestType:  utils.MetaRated,
			utils.Category:     "call",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			utils.AnswerTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			utils.Usage:        10 * time.Second,
			utils.ExtraInfo:    "extraInfo",
			utils.ExtraFields:  map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
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
		fileContent, err := os.ReadFile("/usr/share/cgrates/storage/mysql/create_cdrs_tables.sql")
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		qries := strings.SplitSeq(string(fileContent), ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
		for qry := range qries {
			qry = strings.TrimSpace(qry) // Avoid empty queries
			if len(qry) == 0 {
				continue
			}
			if _, err := sqlDB.Exec(qry); err != nil {
				t.Fatal(err)
			}
		}
		tx := db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql := &utils.CDRSQLTable{
			Tenant:    cdr1.Tenant,
			Opts:      cdr1.Opts,
			Event:     cdr1.Event,
			CreatedAt: time.Now(),
		}
		cdrSql2 := &utils.CDRSQLTable{
			Tenant:    cdr2.Tenant,
			Opts:      cdr2.Opts,
			Event:     cdr2.Event,
			CreatedAt: time.Now(),
		}
		cdrsql3 := &utils.CDRSQLTable{
			Tenant:    cdr3.Tenant,
			Opts:      cdr3.Opts,
			Event:     cdr3.Event,
			CreatedAt: time.Now(),
		}
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
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID,RateID2
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_filters"),
		DBCfg:            dbcfg,
		Encoding:         *utils.Encoding,
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
		timeStartFormated := timeStart.Format(time.RFC3339Nano)

		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"Category\":\"call\",\"Destination\":\"1002\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{\"*cdrID\":\"%s\",\"*originID\":\"dsafdsaf\",\"*rateSCost\":\"{\\\"Altered\\\":null,\\\"Cost\\\":2.3,\\\"CostIntervals\\\":[{\\\"CompressFactor\\\":1,\\\"Increments\\\":[{\\\"CompressFactor\\\":1,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":0,\\\"Usage\\\":120000000000},{\\\"CompressFactor\\\":60,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":1,\\\"Usage\\\":1000000000}]}],\\\"ID\\\":\\\"DEFAULT_RATE\\\",\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"MinCost\\\":0,\\\"Rates\\\":{\\\"RateID2\\\":{\\\"FixedFee\\\":0.1,\\\"Increment\\\":60000000000,\\\"IntervalStart\\\":0,\\\"RecurrentFee\\\":0.01,\\\"Unit\\\":1000000000}}}\"}}>", timeStartFormated, timeStartFormated, cdrID)
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
		fileContent, err := os.ReadFile("/usr/share/cgrates/storage/mysql/create_cdrs_tables.sql")
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		qries := strings.SplitSeq(string(fileContent), ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
		for qry := range qries {
			qry = strings.TrimSpace(qry) // Avoid empty queries
			if len(qry) == 0 {
				continue
			}
			if _, err := sqlDB.Exec(qry); err != nil {
				t.Fatal(err)
			}
		}
		tx := db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql := &utils.CDRSQLTable{
			Tenant:    cdr1.Tenant,
			Opts:      cdr1.Opts,
			Event:     cdr1.Event,
			CreatedAt: time.Now(),
		}
		cdrSql2 := &utils.CDRSQLTable{
			Tenant:    cdr2.Tenant,
			Opts:      cdr2.Opts,
			Event:     cdr2.Event,
			CreatedAt: time.Now(),
		}
		cdrsql3 := &utils.CDRSQLTable{
			Tenant:    cdr3.Tenant,
			Opts:      cdr3.Opts,
			Event:     cdr3.Event,
			CreatedAt: time.Now(),
		}
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
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID,RateID2
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_delete_indexed_fields"),
		DBCfg:            dbcfg,
		Encoding:         *utils.Encoding,
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
		timeStartFormated := timeStart.Format(time.RFC3339Nano)
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"Category\":\"call\",\"Destination\":\"1002\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{\"*cdrID\":\"%s\",\"*originID\":\"dsafdsaf\",\"*rateSCost\":\"{\\\"Altered\\\":null,\\\"Cost\\\":2.3,\\\"CostIntervals\\\":[{\\\"CompressFactor\\\":1,\\\"Increments\\\":[{\\\"CompressFactor\\\":1,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":0,\\\"Usage\\\":120000000000},{\\\"CompressFactor\\\":60,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":1,\\\"Usage\\\":1000000000}]}],\\\"ID\\\":\\\"DEFAULT_RATE\\\",\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"MinCost\\\":0,\\\"Rates\\\":{\\\"RateID2\\\":{\\\"FixedFee\\\":0.1,\\\"Increment\\\":60000000000,\\\"IntervalStart\\\":0,\\\"RecurrentFee\\\":0.01,\\\"Unit\\\":1000000000}}}\"}}>", timeStartFormated, timeStartFormated, cdrID)
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
				if strings.Contains(fmt.Sprintln(value), "RateID2") {
					t.Fatalf("Expected CDR with RatingID: \"RateID2\" to be deleted. Received column <%q>, value <%q>", col, value)
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
		fileContent, err := os.ReadFile("/usr/share/cgrates/storage/mysql/create_cdrs_tables.sql")
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		qries := strings.SplitSeq(string(fileContent), ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
		for qry := range qries {
			qry = strings.TrimSpace(qry) // Avoid empty queries
			if len(qry) == 0 {
				continue
			}
			if _, err := sqlDB.Exec(qry); err != nil {
				t.Fatal(err)
			}
		}
		tx := db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql := &utils.CDRSQLTable{
			Tenant:    cdr1.Tenant,
			Opts:      cdr1.Opts,
			Event:     cdr1.Event,
			CreatedAt: time.Now(),
		}
		cdrSql2 := &utils.CDRSQLTable{
			Tenant:    cdr2.Tenant,
			Opts:      cdr2.Opts,
			Event:     cdr2.Event,
			CreatedAt: time.Now(),
		}
		cdrsql3 := &utils.CDRSQLTable{
			Tenant:    cdr3.Tenant,
			Opts:      cdr3.Opts,
			Event:     cdr3.Event,
			CreatedAt: time.Now(),
		}
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
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID,RateID2
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_meta_delete"),
		DBCfg:            dbcfg,
		Encoding:         *utils.Encoding,
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
		timeStartFormated := timeStart.Format(time.RFC3339Nano)
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"Category\":\"call\",\"Destination\":\"1002\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{\"*cdrID\":\"%s\",\"*originID\":\"dsafdsaf\",\"*rateSCost\":\"{\\\"Altered\\\":null,\\\"Cost\\\":2.3,\\\"CostIntervals\\\":[{\\\"CompressFactor\\\":1,\\\"Increments\\\":[{\\\"CompressFactor\\\":1,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":0,\\\"Usage\\\":120000000000},{\\\"CompressFactor\\\":60,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":1,\\\"Usage\\\":1000000000}]}],\\\"ID\\\":\\\"DEFAULT_RATE\\\",\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"MinCost\\\":0,\\\"Rates\\\":{\\\"RateID2\\\":{\\\"FixedFee\\\":0.1,\\\"Increment\\\":60000000000,\\\"IntervalStart\\\":0,\\\"RecurrentFee\\\":0.01,\\\"Unit\\\":1000000000}}}\"}}>", timeStartFormated, timeStartFormated, cdrID)
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
				if strings.Contains(fmt.Sprintln(value), "RateID2") {
					t.Fatalf("Expected CDR with RatingID: \"RateID2\" to be deleted. Received column <%q>, value <%q>", col, value)
				}
			}
		}
	})
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
		fileContent, err := os.ReadFile("/usr/share/cgrates/storage/mysql/create_cdrs_tables.sql")
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		qries := strings.SplitSeq(string(fileContent), ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
		for qry := range qries {
			qry = strings.TrimSpace(qry) // Avoid empty queries
			if len(qry) == 0 {
				continue
			}
			if _, err := sqlDB.Exec(qry); err != nil {
				t.Fatal(err)
			}
		}
		mockCDRsTable := "DROP TABLE IF EXISTS cdrsProcessed; CREATE TABLE cdrsProcessed ( `id` int(11) NOT NULL AUTO_INCREMENT, `tenant` VARCHAR(40) NOT NULL, `opts` JSON NOT NULL, `event` JSON NOT NULL, `created_at` TIMESTAMP NULL, `updated_at` TIMESTAMP NULL, `deleted_at` TIMESTAMP NULL,  PRIMARY KEY (`id`));ALTER TABLE cdrsProcessed ADD COLUMN cdrid VARCHAR(40) GENERATED ALWAYS AS ( JSON_VALUE(opts, '$.\"*cdrID\"') );CREATE UNIQUE INDEX opts_cdrid_idx ON cdrsProcessed (cdrid);"
		qries = strings.SplitSeq(mockCDRsTable, ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
		for qry := range qries {
			qry = strings.TrimSpace(qry) // Avoid empty queries
			if len(qry) == 0 {
				continue
			}
			if _, err := sqlDB.Exec(qry); err != nil {
				t.Fatal(err)
			}
		}

		tx := db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql := &utils.CDRSQLTable{
			Tenant:    cdr1.Tenant,
			Opts:      cdr1.Opts,
			Event:     cdr1.Event,
			CreatedAt: time.Now(),
		}
		cdrSql2 := &utils.CDRSQLTable{
			Tenant:    cdr2.Tenant,
			Opts:      cdr2.Opts,
			Event:     cdr2.Event,
			CreatedAt: time.Now(),
		}
		cdrsql3 := &utils.CDRSQLTable{
			Tenant:    cdr3.Tenant,
			Opts:      cdr3.Opts,
			Event:     cdr3.Event,
			CreatedAt: time.Now(),
		}
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
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID,RateID2
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_move"),
		DBCfg:            dbcfg,
		Encoding:         *utils.Encoding,
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
		timeStartFormated := timeStart.Format(time.RFC3339Nano)
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"Category\":\"call\",\"Destination\":\"1002\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{\"*cdrID\":\"%s\",\"*originID\":\"dsafdsaf\",\"*rateSCost\":\"{\\\"Altered\\\":null,\\\"Cost\\\":2.3,\\\"CostIntervals\\\":[{\\\"CompressFactor\\\":1,\\\"Increments\\\":[{\\\"CompressFactor\\\":1,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":0,\\\"Usage\\\":120000000000},{\\\"CompressFactor\\\":60,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":1,\\\"Usage\\\":1000000000}]}],\\\"ID\\\":\\\"DEFAULT_RATE\\\",\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"MinCost\\\":0,\\\"Rates\\\":{\\\"RateID2\\\":{\\\"FixedFee\\\":0.1,\\\"Increment\\\":60000000000,\\\"IntervalStart\\\":0,\\\"RecurrentFee\\\":0.01,\\\"Unit\\\":1000000000}}}\"}}>", timeStartFormated, timeStartFormated, cdrID)
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
				if strings.Contains(fmt.Sprintln(value), "RateID2") {
					t.Fatalf("Expected CDR with RatingID: \"RateID2\" to be deleted. Received column <%q>, value <%q>", col, value)
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
		ev := rslt2[0]["event"]
		var evMap map[string]any
		if err := json.Unmarshal([]byte(ev.(string)), &evMap); err != nil {
			log.Fatal(err)
		}
		timeStartFormated := evMap["AnswerTime"]
		createdAt := rslt2[0]["created_at"]
		updatedAt := rslt2[0]["updated_at"]
		exp := fmt.Sprintf("map[cdrid:%s created_at:%s deleted_at:<nil> event:{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"Category\":\"call\",\"Destination\":\"1002\",\"ExtraFields\":{\"field_extr1\":\"val_extr1\",\"fieldextr2\":\"valextr2\"},\"ExtraInfo\":\"extraInfo\",\"OrderID\":123,\"OriginHost\":\"192.168.1.1\",\"OriginID\":\"oid2\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Source\":\"test\",\"Subject\":\"1001\",\"ToR\":\"*voice\",\"Usage\":10000000000} id:2 opts:{\"*cdrID\":\"%s\",\"*cdrsExport\":false,\"*chargeID\":\"%s\",\"*chargers\":true,\"*cost\":1.01,\"*originID\":\"dsafdsaf\",\"*rateSCost\":{\"Altered\":null,\"Cost\":2.3,\"CostIntervals\":[{\"CompressFactor\":1,\"Increments\":[{\"CompressFactor\":1,\"RateID\":\"RateID2\",\"RateIntervalIndex\":0,\"Usage\":120000000000},{\"CompressFactor\":60,\"RateID\":\"RateID2\",\"RateIntervalIndex\":1,\"Usage\":1000000000}]}],\"ID\":\"DEFAULT_RATE\",\"MaxCost\":0,\"MaxCostStrategy\":\"\",\"MinCost\":0,\"Rates\":{\"RateID2\":{\"FixedFee\":0.1,\"Increment\":60000000000,\"IntervalStart\":0,\"RecurrentFee\":0.01,\"Unit\":1000000000}}},\"*rates\":true,\"*runID\":\"*default\",\"*subsys\":\"*chargers\",\"*usage\":10000000000} tenant:cgrates.org updated_at:%s]", cdrID, createdAt, timeStartFormated, timeStartFormated, cdrID, cdrID, updatedAt)
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
		fileContent, err := os.ReadFile("/usr/share/cgrates/storage/mysql/create_cdrs_tables.sql")
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		qries := strings.SplitSeq(string(fileContent), ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
		for qry := range qries {
			qry = strings.TrimSpace(qry) // Avoid empty queries
			if len(qry) == 0 {
				continue
			}
			if _, err := sqlDB.Exec(qry); err != nil {
				t.Fatal(err)
			}
		}
		tx := db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql := &utils.CDRSQLTable{
			Tenant:    cdr1.Tenant,
			Opts:      cdr1.Opts,
			Event:     cdr1.Event,
			CreatedAt: time.Now(),
		}
		cdrSql2 := &utils.CDRSQLTable{
			Tenant:    cdr2.Tenant,
			Opts:      cdr2.Opts,
			Event:     cdr2.Event,
			CreatedAt: time.Now(),
		}
		cdrsql3 := &utils.CDRSQLTable{
			Tenant:    cdr3.Tenant,
			Opts:      cdr3.Opts,
			Event:     cdr3.Event,
			CreatedAt: time.Now(),
		}
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
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID,RateID2
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_update"),
		DBCfg:            dbcfg,
		Encoding:         *utils.Encoding,
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
		timeStartFormated := timeStart.Format(time.RFC3339Nano)
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"Category\":\"call\",\"Destination\":\"1002\",\"ExtraInfo\":\"extraInfo\",\"Id\":\"2\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{\"*cdrID\":\"%s\",\"*originID\":\"dsafdsaf\",\"*rateSCost\":\"{\\\"Altered\\\":null,\\\"Cost\\\":2.3,\\\"CostIntervals\\\":[{\\\"CompressFactor\\\":1,\\\"Increments\\\":[{\\\"CompressFactor\\\":1,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":0,\\\"Usage\\\":120000000000},{\\\"CompressFactor\\\":60,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":1,\\\"Usage\\\":1000000000}]}],\\\"ID\\\":\\\"DEFAULT_RATE\\\",\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"MinCost\\\":0,\\\"Rates\\\":{\\\"RateID2\\\":{\\\"FixedFee\\\":0.1,\\\"Increment\\\":60000000000,\\\"IntervalStart\\\":0,\\\"RecurrentFee\\\":0.01,\\\"Unit\\\":1000000000}}}\"}}>", timeStartFormated, timeStartFormated, cdrID)
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
				if col == "tenant" {
					if value.(string) == "updatedTenant" {
						countST++
					}
				}
			}
		}
		if countST != 1 {
			t.Errorf("Expected CDR to have updated the tenant from <cgrates.org> to <updatedTenant>  Received <%v> \nCounted <%v> with expected setup_time", utils.ToJSON(rslt), countST)
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
		fileContent, err := os.ReadFile("/usr/share/cgrates/storage/mysql/create_cdrs_tables.sql")
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		qries := strings.SplitSeq(string(fileContent), ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
		for qry := range qries {
			qry = strings.TrimSpace(qry) // Avoid empty queries
			if len(qry) == 0 {
				continue
			}
			if _, err := sqlDB.Exec(qry); err != nil {
				t.Fatal(err)
			}
		}
		tx := db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql := &utils.CDRSQLTable{
			Tenant:    cdr1.Tenant,
			Opts:      cdr1.Opts,
			Event:     cdr1.Event,
			CreatedAt: time.Now(),
		}
		cdrSql2 := &utils.CDRSQLTable{
			Tenant:    cdr2.Tenant,
			Opts:      cdr2.Opts,
			Event:     cdr2.Event,
			CreatedAt: time.Now(),
		}
		cdrsql3 := &utils.CDRSQLTable{
			Tenant:    cdr3.Tenant,
			Opts:      cdr3.Opts,
			Event:     cdr3.Event,
			CreatedAt: time.Now(),
		}
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
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID,RateID2
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql`,
	}

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_raw_update"),
		DBCfg:            dbcfg,
		Encoding:         *utils.Encoding,
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
		timeStartFormated := timeStart.Format(time.RFC3339Nano)
		expectedLog := fmt.Sprintf("\"Event\":{\"Account\":\"1001\",\"AnswerTime\":\"%s\",\"Category\":\"call\",\"Destination\":\"1002\",\"RequestType\":\"*rated\",\"SetupTime\":\"%s\",\"Subject\":\"1001\",\"Tenant\":\"cgrates.org\",\"ToR\":\"*voice\",\"Usage\":\"10000000000\"},\"APIOpts\":{\"*cdrID\":\"%s\",\"*originID\":\"dsafdsaf\",\"*rateSCost\":\"{\\\"Altered\\\":null,\\\"Cost\\\":2.3,\\\"CostIntervals\\\":[{\\\"CompressFactor\\\":1,\\\"Increments\\\":[{\\\"CompressFactor\\\":1,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":0,\\\"Usage\\\":120000000000},{\\\"CompressFactor\\\":60,\\\"RateID\\\":\\\"RateID2\\\",\\\"RateIntervalIndex\\\":1,\\\"Usage\\\":1000000000}]}],\\\"ID\\\":\\\"DEFAULT_RATE\\\",\\\"MaxCost\\\":0,\\\"MaxCostStrategy\\\":\\\"\\\",\\\"MinCost\\\":0,\\\"Rates\\\":{\\\"RateID2\\\":{\\\"FixedFee\\\":0.1,\\\"Increment\\\":60000000000,\\\"IntervalStart\\\":0,\\\"RecurrentFee\\\":0.01,\\\"Unit\\\":1000000000}}}\"}}>", timeStartFormated, timeStartFormated, cdrID)
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
				if col == "tenant" {
					if value.(string) == "updatedTenant" {
						countST++
					}
				}
			}
		}
		if countST != 1 {
			t.Errorf("Expected CDR to have updated the tenant from <cgrates.org> to <updatedTenant>  Received <%v> \nCounted <%v> with expected setup_time", utils.ToJSON(rslt), countST)
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
		fileContent, err := os.ReadFile("/usr/share/cgrates/storage/mysql/create_cdrs_tables.sql")
		if err != nil {
			t.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		qries := strings.SplitSeq(string(fileContent), ";") // Script has normally multiple queries separate by ';' go driver does not understand this so we handle it here
		for qry := range qries {
			qry = strings.TrimSpace(qry) // Avoid empty queries
			if len(qry) == 0 {
				continue
			}
			if _, err := sqlDB.Exec(qry); err != nil {
				t.Fatal(err)
			}
		}
		tx := db.Begin()
		tx = tx.Table(utils.CDRsTBL)
		cdrSql := &utils.CDRSQLTable{
			Tenant:    cdr1.Tenant,
			Opts:      cdr1.Opts,
			Event:     cdr1.Event,
			CreatedAt: time.Now(),
		}
		cdrSql2 := &utils.CDRSQLTable{
			Tenant:    cdr2.Tenant,
			Opts:      cdr2.Opts,
			Event:     cdr2.Event,
			CreatedAt: time.Now(),
		}
		cdrsql3 := &utils.CDRSQLTable{
			Tenant:    cdr3.Tenant,
			Opts:      cdr3.Opts,
			Event:     cdr3.Event,
			CreatedAt: time.Now(),
		}
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
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID,RateID2
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql`,
	}

	jsonCfg := `{
"general": {
	"reply_timeout": "10s",
	"default_timezone": "UTC"
},

"logger": {
    "level": 7
},

"admins": {
	"enabled": true
},

"sessions": {
	"enabled": true,
},

"ers": {
	"enabled": true,
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
					"*gt:~*req.event.AnswerTime:-168h", // dont process cdrs with answer_time older than 7 days ago
					"FLTR_SQL_RatingID", // "*eq:~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID:RateID",
					"*string:~*vars.*readerID:mysql",
					"FLTR_VARS", // "*string:~*vars.*readerID:mysql",
					"*notempty:~*vars.*readerID:''",
			],
			"flags": ["*dryRun"],
			"fields":[
				{"tag": "*cdrID", "path": "*opts.*cdrID", "type": "*variable", "value": "~*req.opts.*cdrID", "mandatory": true},
				{"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.event.ToR", "mandatory": true},
				{"tag": "*originID", "path": "*opts.*originID", "type": "*variable", "value": "~*req.opts.*originID", "mandatory": true},
				{"tag": "RequestType", "path": "*cgreq.RequestType", "type": "*variable", "value": "~*req.event.RequestType", "mandatory": true},
				{"tag": "Tenant", "path": "*cgreq.Tenant", "type": "*variable", "value": "~*req.tenant", "mandatory": true},
				{"tag": "Category", "path": "*cgreq.Category", "type": "*variable", "value": "~*req.event.Category", "mandatory": true},
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.event.Account", "mandatory": true},
				{"tag": "Subject", "path": "*cgreq.Subject", "type": "*variable", "value": "~*req.event.Subject", "mandatory": true},
				{"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.event.Destination", "mandatory": true},
				{"tag": "SetupTime", "path": "*cgreq.SetupTime", "type": "*variable", "value": "~*req.event.SetupTime", "mandatory": true},
				{"tag": "AnswerTime", "path": "*cgreq.AnswerTime", "type": "*variable", "value": "~*req.event.AnswerTime", "mandatory": true},
				{"tag": "RateSCost", "path": "*opts.*rateSCost", "type": "*variable", "value": "~*req.opts.*rateSCost", "mandatory": true},
				{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.event.Usage", "mandatory": true},
			],
		},
	],
},

}`

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON:       jsonCfg,
		DBCfg:            dbcfg,
		Encoding:         *utils.Encoding,
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
