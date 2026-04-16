//go:build integration

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
package general_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const ersDryRunMySQL = "<ERs> DRY_RUN, reader: <mysql>"

var (
	dbConnString = "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'"
	timeStart    = time.Now().Truncate(time.Second)
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

	sqlFilterTPFiles = map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.opts.*rateSCost.CostIntervals[0].Increments[0].RateID,RateID2
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql`,
	}
)

func getDBCfg(t *testing.T) engine.DBCfg {
	t.Helper()
	switch *utils.DBType {
	case utils.MetaInternal:
		return engine.InternalDBCfg
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	return engine.DBCfg{}
}

func openTestDB(t *testing.T, cdrs ...*utils.CDR) *gorm.DB {
	t.Helper()

	cdb, err := gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates")),
		&gorm.Config{AllowGlobalUpdate: true})
	if err != nil {
		t.Fatal(err)
	}
	if err = cdb.Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`).Error; err != nil {
		t.Fatal(err)
	}
	sqlCDB, err := cdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlCDB.SetConnMaxLifetime(5 * time.Second)

	db, err := gorm.Open(mysql.Open(fmt.Sprintf(dbConnString, "cgrates2")),
		&gorm.Config{AllowGlobalUpdate: true})
	if err != nil {
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
	for qry := range strings.SplitSeq(string(fileContent), ";") {
		qry = strings.TrimSpace(qry)
		if len(qry) == 0 {
			continue
		}
		if _, err := sqlDB.Exec(qry); err != nil {
			t.Fatal(err)
		}
	}

	tx := db.Begin()
	tx = tx.Table(utils.CDRsTBL)
	for _, cdr := range cdrs {
		if err := tx.Save(&utils.CDRSQLTable{
			Tenant:    cdr.Tenant,
			Opts:      cdr.Opts,
			Event:     cdr.Event,
			CreatedAt: time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

	var count int64
	db.Table(utils.CDRsTBL).Count(&count)
	if count != int64(len(cdrs)) {
		t.Fatalf("expected %d rows in cdrs, got %d", len(cdrs), count)
	}

	t.Cleanup(func() {
		_ = db.Migrator().DropTable("cdrs")
		_ = db.Exec(`DROP DATABASE cgrates2;`).Error
		if d, err := db.DB(); err == nil {
			_ = d.Close()
		}
		if d, err := cdb.DB(); err == nil {
			_ = d.Close()
		}
	})
	return db
}

func countRows(t *testing.T, db *gorm.DB, table string) int64 {
	t.Helper()
	var count int64
	if err := db.Table(table).Count(&count).Error; err != nil {
		t.Fatalf("failed to count rows in %q: %v", table, err)
	}
	return count
}

func waitFor(t *testing.T, check func() bool, msg string, timeout time.Duration) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	backoff := utils.FibDuration(time.Millisecond, 0)
	for {
		if check() {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("timed out after %s: %s", timeout, msg)
		case <-time.After(backoff()):
		}
	}
}

func waitForERsLog(t *testing.T, buf *bytes.Buffer, substr string, timeout time.Duration) {
	t.Helper()
	waitFor(t,
		func() bool { return strings.Contains(buf.String(), substr) },
		fmt.Sprintf("log line %q not found", substr),
		timeout,
	)
}

func parseCGREvent(t *testing.T, buf *bytes.Buffer) *utils.CGREvent {
	t.Helper()
	logOutput := buf.String()
	_, after, ok := strings.Cut(logOutput, "CGREvent: ")
	if !ok {
		t.Fatalf("CGREvent not found in log output:\n%s", logOutput)
	}
	var ev utils.CGREvent
	if err := json.NewDecoder(strings.NewReader(after)).Decode(&ev); err != nil {
		t.Fatal(err)
	}
	return &ev
}

func baseExpectedEvent() map[string]any {
	ts := timeStart.Format("2006-01-02T15:04:05Z07:00")
	return map[string]any{
		"Account":     "1001",
		"AnswerTime":  ts,
		"Category":    "call",
		"Destination": "1002",
		"RequestType": "*rated",
		"SetupTime":   ts,
		"Subject":     "1001",
		"Tenant":      "cgrates.org",
		"ToR":         "*voice",
		"Usage":       "10000000000",
	}
}

func assertNoRateID2(t *testing.T, db *gorm.DB) {
	t.Helper()
	var rows []map[string]any
	if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rows).Error; err != nil {
		t.Fatalf("failed to query table: %v", err)
	}
	for _, row := range rows {
		for col, val := range row {
			if strings.Contains(fmt.Sprint(val), "RateID2") {
				t.Fatalf("expected CDR with RateID2 to be deleted, found in column %q", col)
			}
		}
	}
}

func TestERSSQLFilters(t *testing.T) {
	db := openTestDB(t, cdr1, cdr2, cdr3)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_filters"),
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	waitForERsLog(t, buf, ersDryRunMySQL, 2*time.Second)
	if got := strings.Count(buf.String(), ersDryRunMySQL); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(baseExpectedEvent()); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}
	if got := countRows(t, db, utils.CDRsTBL); got != 3 {
		t.Fatalf("expected 3 rows, got %d", got)
	}
}

func TestERSSQLFiltersDeleteIndexedFields(t *testing.T) {
	db := openTestDB(t, cdr1, cdr2, cdr3)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_filters"),
		ConfigJSON: `{
"ers": {
  "readers": [
    {
      "id": "mysql",
      "processed_path": "*delete",
      "opts": {
        "sqlBatchSize": 2,
        "sqlDeleteIndexedFields": ["id"]
      }
    }
  ]
}
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	waitFor(t,
		func() bool { return countRows(t, db, utils.CDRsTBL) == 2 },
		"expected 2 rows in cdrs after delete",
		2*time.Second,
	)
	if got := strings.Count(buf.String(), ersDryRunMySQL); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(baseExpectedEvent()); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}
	assertNoRateID2(t, db)
}

func TestERSSQLFiltersWithMetaDelete(t *testing.T) {
	db := openTestDB(t, cdr1, cdr2, cdr3)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_filters"),
		ConfigJSON: `{
"ers": {
  "readers": [
    {
      "id": "mysql",
      "processed_path": "*delete",
      "opts": {
        "sqlBatchSize": 1
      }
    }
  ]
}
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	waitFor(t,
		func() bool { return countRows(t, db, utils.CDRsTBL) == 2 },
		"expected 2 rows in cdrs after delete",
		2*time.Second,
	)
	if got := strings.Count(buf.String(), ersDryRunMySQL); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(baseExpectedEvent()); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}
	assertNoRateID2(t, db)
}

func TestERSSQLFiltersMove(t *testing.T) {
	db := openTestDB(t, cdr1, cdr2, cdr3)

	// Create cdrsProcessed table for the move target.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	cdrsProcessedSchema := "DROP TABLE IF EXISTS cdrsProcessed; CREATE TABLE cdrsProcessed ( `id` int(11) NOT NULL AUTO_INCREMENT, `tenant` VARCHAR(40) NOT NULL, `opts` JSON NOT NULL, `event` JSON NOT NULL, `created_at` TIMESTAMP NULL, `updated_at` TIMESTAMP NULL, `deleted_at` TIMESTAMP NULL,  PRIMARY KEY (`id`));ALTER TABLE cdrsProcessed ADD COLUMN cdrid VARCHAR(40) GENERATED ALWAYS AS ( JSON_VALUE(opts, '$.\"*cdrID\"') );CREATE UNIQUE INDEX opts_cdrid_idx ON cdrsProcessed (cdrid);"
	for qry := range strings.SplitSeq(cdrsProcessedSchema, ";") {
		qry = strings.TrimSpace(qry)
		if len(qry) == 0 {
			continue
		}
		if _, err := sqlDB.Exec(qry); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_ = db.Migrator().DropTable("cdrsProcessed")
	})

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_filters"),
		ConfigJSON: `{
"ees": {
  "enabled": true,
  "exporters": [
    {
      "id": "SQLExporter",
      "type": "*sql",
      "export_path": "mysql://cgrates:CGRateS.org@127.0.0.1:3306",
      "attempts": 1,
      "opts": {
        "sqlDBName": "cgrates2",
        "sqlTableName": "cdrsProcessed"
      },
      "flags": ["*log"]
    }
  ]
},
"ers": {
  "conns": {
    "*ees": [{"ConnIDs": ["*internal"]}]
  },
  "readers": [
    {
      "id": "mysql",
      "ees_success_ids": ["SQLExporter"],
      "processed_path": "*delete",
      "opts": {
        "sqlBatchSize": 0,
        "sqlDeleteIndexedFields": ["id"]
      }
    }
  ]
}
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	waitFor(t,
		func() bool {
			return countRows(t, db, utils.CDRsTBL) == 2 &&
				countRows(t, db, "cdrsProcessed") == 1
		},
		"expected 2 rows in cdrs and 1 row in cdrsProcessed after move",
		2*time.Second,
	)
	assertNoRateID2(t, db)

	if got := strings.Count(buf.String(), ersDryRunMySQL); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(baseExpectedEvent()); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}

	var movedRows []map[string]any
	if err := db.Raw("SELECT * FROM cdrsProcessed").Scan(&movedRows).Error; err != nil {
		t.Fatalf("failed to query cdrsProcessed: %v", err)
	}
	if got := movedRows[0]["tenant"]; got != "cgrates.org" {
		t.Errorf("moved row tenant = %v, want cgrates.org", got)
	}
	var evMap map[string]any
	if err := json.Unmarshal([]byte(movedRows[0]["event"].(string)), &evMap); err != nil {
		t.Fatal(err)
	}
	if got := evMap["Account"]; got != "1001" {
		t.Errorf("moved row Account = %v, want 1001", got)
	}
	if got := evMap["OriginID"]; got != "oid2" {
		t.Errorf("moved row OriginID = %v, want oid2", got)
	}
}

func TestERSSQLFiltersUpdate(t *testing.T) {
	db := openTestDB(t, cdr1, cdr2, cdr3)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_filters"),
		ConfigJSON: `{
"ees": {
  "enabled": true,
  "exporters": [
    {
      "id": "SQLExporter",
      "type": "*sql",
      "export_path": "mysql://cgrates:CGRateS.org@127.0.0.1:3306",
      "attempts": 1,
      "opts": {
        "sqlDBName": "cgrates2",
        "sqlTableName": "cdrs",
        "sqlUpdateIndexedFields": ["id"]
      },
      "flags": ["*log"],
      "fields": [
        {"tag": "ID", "path": "*exp.id", "type": "*variable", "value": "~*req.Id"},
        {"tag": "Tenant", "path": "*exp.tenant", "type": "*constant", "value": "updatedTenant"}
      ]
    }
  ]
},
"ers": {
  "conns": {
    "*ees": [{"ConnIDs": ["*localhost"]}]
  },
  "readers": [
    {
      "id": "mysql",
      "flags": ["*dryRun", "*export"],
      "fields": [
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
        {"tag": "ExtraInfo", "path": "*cgreq.ExtraInfo", "type": "*variable", "value": "~*req.event.ExtraInfo", "mandatory": true},
        {"tag": "ID", "path": "*cgreq.Id", "type": "*variable", "value": "~*req.id", "mandatory": true}
      ]
    }
  ]
}
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	waitFor(t,
		func() bool {
			var c int64
			db.Table(utils.CDRsTBL).Where("tenant = ?", "updatedTenant").Count(&c)
			return c == 1
		},
		"expected 1 row with tenant=updatedTenant",
		2*time.Second,
	)
	if got := countRows(t, db, utils.CDRsTBL); got != 3 {
		t.Fatalf("expected 3 rows, got %d", got)
	}

	if got := strings.Count(buf.String(), ersDryRunMySQL); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	expectedEvent := baseExpectedEvent()
	expectedEvent["ExtraInfo"] = "extraInfo"
	expectedEvent["Id"] = "2"
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(expectedEvent); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}
}

func TestERSSQLFiltersRawUpdate(t *testing.T) {
	db := openTestDB(t, cdr1, cdr2, cdr3)

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "ers_mysql_filters"),
		ConfigJSON: `{
"ees": {
  "enabled": true,
  "exporters": [
    {
      "id": "SQLExporter",
      "type": "*sql",
      "export_path": "mysql://cgrates:CGRateS.org@127.0.0.1:3306",
      "attempts": 1,
      "opts": {
        "sqlConnMaxLifetime": "5s",
        "sqlDBName": "cgrates2",
        "sqlTableName": "cdrs",
        "sqlBatchSize": -1,
        "sqlUpdateIndexedFields": ["id"]
      },
      "flags": ["*log"],
      "fields": [
        {"tag": "ID", "path": "*exp.id", "type": "*variable", "value": "~*req.id"},
        {"tag": "Tenant", "path": "*exp.tenant", "type": "*constant", "value": "updatedTenant"}
      ]
    }
  ]
},
"ers": {
  "conns": {
    "*ees": [{"ConnIDs": ["*localhost"]}]
  },
  "readers": [
    {
      "id": "mysql",
      "ees_success_ids": ["SQLExporter"]
    }
  ]
}
}`,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	waitFor(t,
		func() bool {
			var c int64
			db.Table(utils.CDRsTBL).Where("tenant = ?", "updatedTenant").Count(&c)
			return c == 1
		},
		"expected 1 row with tenant=updatedTenant",
		2*time.Second,
	)
	if got := countRows(t, db, utils.CDRsTBL); got != 3 {
		t.Fatalf("expected 3 rows, got %d", got)
	}

	if got := strings.Count(buf.String(), ersDryRunMySQL); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got, want := utils.ToJSON(ev.Event), utils.ToJSON(baseExpectedEvent()); got != want {
		t.Errorf("got event\n%s\nwant\n%s", got, want)
	}
}

func TestERSSQLFiltersErr(t *testing.T) {
	_ = openTestDB(t, cdr1, cdr2, cdr3)

	jsonCfg := `{
"general": {
  "reply_timeout": "10s",
  "default_timezone": "UTC"
},
"admins": {
  "enabled": true
},
"sessions": {
  "enabled": true
},
"ers": {
  "enabled": true,
  "readers": [
    {
      "id": "mysql",
      "type": "*sql",
      "run_delay": "1m",
      "start_delay": "100ms",
      "source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
      "opts": {
        "sqlDBName": "cgrates2",
        "sqlTableName": "cdrs",
        "sqlBatchSize": 2,
        "sqlDeleteIndexedFields": ["id"]
      },
      "processed_path": "*delete",
      "tenant": "cgrates.org",
      "filters": [
        "*gt:~*req.event.AnswerTime:-168h",
        "FLTR_SQL_RatingID",
        "*string:~*vars.*readerID:mysql",
        "FLTR_VARS",
        "*notempty:~*vars.*readerID:''"
      ],
      "flags": ["*dryRun"],
      "fields": [
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
        {"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.event.Usage", "mandatory": true}
      ]
    }
  ]
}
}`

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON:       jsonCfg,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	errLine := `[ERROR] <ERs> error: <value of filter <*notempty:~*vars.*readerID:''> is not empty <''>>`
	waitForERsLog(t, buf, errLine, 2*time.Second)
}

// TestERSSQLFilterUnquote checks that JSON_UNQUOTE wraps only JSON_VALUE,
// not the whole comparison (which produces invalid SQL on MySQL 8).
func TestERSSQLFilterUnquote(t *testing.T) {
	_ = openTestDB(t, cdr2)

	jsonCfg := `{
"general": {
  "reply_timeout": "10s",
  "default_timezone": "UTC"
},
"admins": {
  "enabled": true
},
"sessions": {
  "enabled": true
},
"ers": {
  "enabled": true,
  "readers": [
    {
      "id": "mysql",
      "type": "*sql",
      "run_delay": "1m",
      "start_delay": "100ms",
      "source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
      "opts": {
        "sqlDBName": "cgrates2",
        "sqlTableName": "cdrs",
        "sqlBatchSize": 10
      },
      "tenant": "cgrates.org",
      "filters": [
        "FLTR_SQL_RatingID",
        "FLTR_VARS"
      ],
      "flags": ["*dryRun"],
      "fields": [
        {"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.event.ToR", "mandatory": true},
        {"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.event.Account", "mandatory": true},
        {"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.event.Destination", "mandatory": true}
      ]
    }
  ]
}
}`

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON:       jsonCfg,
		DBCfg:            getDBCfg(t),
		Encoding:         *utils.Encoding,
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	waitForERsLog(t, buf, ersDryRunMySQL, 2*time.Second)
	if got := strings.Count(buf.String(), ersDryRunMySQL); got != 1 {
		t.Fatalf("expected 1 DRY_RUN record, got %d", got)
	}
	ev := parseCGREvent(t, buf)
	if got := ev.Event["Account"]; got != "1001" {
		t.Errorf("expected Account=1001, got %v", got)
	}
}

