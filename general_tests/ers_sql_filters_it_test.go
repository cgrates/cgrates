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
				Tenant:    "cgrates.org",
				AccountID: "1001",
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
				Tenant:    "cgrates.org",
				AccountID: "1001",
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

	sqlFilterTPFiles = map[string]string{
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_SQL_RatingID,*eq,~*req.cost_details.Charges[0].RatingID,RatingID2,
cgrates.org,FLTR_VARS,*string,~*vars.*readerID,mysql,`,
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

func openTestDB(t *testing.T, cdrs ...*engine.CDR) *gorm.DB {
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

	tx := db.Begin()
	if !tx.Migrator().HasTable("cdrs") {
		if err = tx.Migrator().CreateTable(new(engine.CDRsql)); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
	}
	tx.Commit()

	tx = db.Begin().Table(utils.CDRsTBL)
	for _, cdr := range cdrs {
		cdrSQL, err := cdr.AsCDRsql(&engine.JSONMarshaler{})
		if err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		cdrSQL.CreatedAt = time.Now()
		if err := tx.Save(cdrSQL).Error; err != nil {
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
		"CGRID":       cgrID,
		"Category":    "call",
		"CostDetails": utils.ToJSON(cdr2.CostDetails),
		"Destination": "1002",
		"OriginID":    "oid2",
		"RequestType": "*rated",
		"SetupTime":   ts,
		"Subject":     "1001",
		"Tenant":      "cgrates.org",
		"ToR":         "*voice",
		"Usage":       "10000000000",
	}
}

func assertNoRatingID2(t *testing.T, db *gorm.DB) {
	t.Helper()
	var rows []map[string]any
	if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rows).Error; err != nil {
		t.Fatalf("failed to query table: %v", err)
	}
	for _, row := range rows {
		for col, val := range row {
			if strings.Contains(fmt.Sprint(val), "RatingID2") {
				t.Fatalf("expected CDR with RatingID2 to be deleted, found in column %q", col)
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
	assertNoRatingID2(t, db)
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
	assertNoRatingID2(t, db)
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
	db := openTestDB(t, cdr1, cdr2, cdr3)

	// Create cdrsProcessed table for the move target.
	tx := db.Begin()
	if !tx.Migrator().HasTable("cdrsProcessed") {
		if err := tx.Migrator().CreateTable(new(mockTableName)); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
	}
	tx.Commit()
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
  "ees_conns": ["*internal"],
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
	assertNoRatingID2(t, db)

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
	timeStartFormated := movedRows[0]["answer_time"]
	createdAt := movedRows[0]["created_at"]
	updatedAt := movedRows[0]["updated_at"]
	exp := fmt.Sprintf("map[account:1001 answer_time:%s category:call cgrid:%s cost:1.01 cost_details:{\"CGRID\":\"test1\",\"RunID\":\"*default\",\"StartTime\":\"2017-01-09T16:18:21Z\",\"Usage\":180000000000,\"Cost\":2.3,\"Charges\":[{\"RatingID\":\"RatingID2\",\"Increments\":[{\"Usage\":120000000000,\"Cost\":2,\"AccountingID\":\"a012888\",\"CompressFactor\":1},{\"Usage\":1000000000,\"Cost\":0.005,\"AccountingID\":\"44d6c02\",\"CompressFactor\":60}],\"CompressFactor\":1}],\"AccountSummary\":{\"Tenant\":\"cgrates.org\",\"AccountID\":\"1001\",\"BalanceSummaries\":[{\"UUID\":\"uuid1\",\"ID\":\"\",\"Type\":\"*monetary\",\"Initial\":0,\"Value\":50,\"Disabled\":false}],\"AllowNegative\":false,\"Disabled\":false},\"Rating\":{\"c1a5ab9\":{\"ConnectFee\":0.1,\"RoundingMethod\":\"*up\",\"RoundingDecimals\":5,\"MaxCost\":0,\"MaxCostStrategy\":\"\",\"TimingID\":\"\",\"RatesID\":\"ec1a177\",\"RatingFiltersID\":\"43e77dc\"}},\"Accounting\":{\"44d6c02\":{\"AccountID\":\"cgrates.org:1001\",\"BalanceUUID\":\"uuid1\",\"RatingID\":\"\",\"Units\":120.7,\"ExtraChargeID\":\"\"},\"a012888\":{\"AccountID\":\"cgrates.org:1001\",\"BalanceUUID\":\"uuid1\",\"RatingID\":\"\",\"Units\":120.7,\"ExtraChargeID\":\"\"}},\"RatingFilters\":null,\"Rates\":{\"ec1a177\":[{\"GroupIntervalStart\":0,\"Value\":0.01,\"RateIncrement\":60000000000,\"RateUnit\":1000000000}]},\"Timings\":null} cost_source:cost source created_at:%s deleted_at:<nil> destination:1002 extra_fields:{\"field_extr1\":\"val_extr1\",\"fieldextr2\":\"valextr2\"} extra_info:extraInfo id:2 origin_host:192.168.1.1 origin_id:oid2 request_type:*rated run_id:*default setup_time:%s source:test subject:1001 tenant:cgrates.org tor:*voice updated_at:%s usage:10000000000]", timeStartFormated, cgrID, createdAt, timeStartFormated, updatedAt)
	for _, row := range movedRows {
		if !strings.Contains(fmt.Sprintf("%+v", row), exp) {
			t.Errorf("Expected <%v>, \nReceived <%v>", exp, fmt.Sprintf("%+v", row))
		}
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
        "sqlUpdateIndexedFields": ["id", "cgrid"]
      },
      "flags": ["*log"],
      "fields": [
        {"tag": "SetupTime", "path": "*exp.setup_time", "type": "*constant", "value": "2018-11-27 14:21:26"},
        {"tag": "Account", "path": "*exp.account", "type": "*variable", "value": "~*req.ExtraInfo"},
        {"tag": "ID", "path": "*exp.id", "type": "*variable", "value": "~*req.Id"},
        {"tag": "CGRID", "path": "*exp.cgrid", "type": "*variable", "value": "~*req.CGRID"}
      ]
    }
  ]
},
"ers": {
  "ees_conns": ["*localhost"],
  "readers": [
    {
      "id": "mysql",
      "ees_ids": ["SQLExporter"],
      "flags": ["*dryrun", "*export"],
      "fields": [
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
        {"tag": "ExtraInfo", "path": "*cgreq.ExtraInfo", "type": "*variable", "value": "~*req.extra_info", "mandatory": true},
        {"tag": "ID", "path": "*cgreq.Id", "type": "*variable", "value": "~*req.id", "mandatory": true}
      ]
    }
  ]
}
}`,
		DBCfg:            getDBCfg(t),
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	updatedSetupTime := time.Date(2018, 11, 27, 14, 21, 26, 0, time.Local).UTC()
	hasUpdatedRow := func() bool {
		var rows []map[string]any
		if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rows).Error; err != nil {
			return false
		}
		for _, row := range rows {
			if st, ok := row["setup_time"].(time.Time); ok && st.UTC().Equal(updatedSetupTime) {
				return true
			}
		}
		return false
	}
	waitFor(t, hasUpdatedRow, "expected a cdrs row with updated setup_time", 2*time.Second)

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

	if got := countRows(t, db, utils.CDRsTBL); got != 3 {
		t.Fatalf("expected 3 rows, got %d", got)
	}
	var rows []map[string]any
	if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rows).Error; err != nil {
		t.Fatalf("failed to query table: %v", err)
	}
	var countST int
	for _, row := range rows {
		st, ok := row["setup_time"].(time.Time)
		if !ok || !st.UTC().Equal(updatedSetupTime) {
			continue
		}
		countST++
		if got := utils.IfaceAsString(row["account"]); got != "extraInfo" {
			t.Errorf("updated row account = %q, want extraInfo", got)
		}
	}
	if countST != 1 {
		t.Errorf("expected 1 row with updated setup_time, got %d", countST)
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
        "sqlDBName": "cgrates2",
        "sqlTableName": "cdrs",
        "sqlBatchSize": -1,
        "sqlUpdateIndexedFields": ["id", "cgrid"]
      },
      "flags": ["*log"],
      "fields": [
        {"tag": "SetupTime", "path": "*exp.setup_time", "type": "*constant", "value": "2018-11-27 14:21:26"},
        {"tag": "Account", "path": "*exp.account", "type": "*variable", "value": "~*req.extra_info"},
        {"tag": "ID", "path": "*exp.id", "type": "*variable", "value": "~*req.id"},
        {"tag": "CGRID", "path": "*exp.cgrid", "type": "*variable", "value": "~*req.cgrid"}
      ]
    }
  ]
},
"ers": {
  "ees_conns": ["*localhost"],
  "readers": [
    {
      "id": "mysql",
      "ees_success_ids": ["SQLExporter"]
    }
  ]
}
}`,
		DBCfg:            getDBCfg(t),
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	updatedSetupTime := time.Date(2018, 11, 27, 14, 21, 26, 0, time.Local).UTC()
	hasUpdatedRow := func() bool {
		var rows []map[string]any
		if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rows).Error; err != nil {
			return false
		}
		for _, row := range rows {
			if st, ok := row["setup_time"].(time.Time); ok && st.UTC().Equal(updatedSetupTime) {
				return true
			}
		}
		return false
	}
	waitFor(t, hasUpdatedRow, "expected a cdrs row with updated setup_time", 2*time.Second)

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
	var rows []map[string]any
	if err := db.Raw("SELECT * FROM " + utils.CDRsTBL).Scan(&rows).Error; err != nil {
		t.Fatalf("failed to query table: %v", err)
	}
	var countST int
	for _, row := range rows {
		st, ok := row["setup_time"].(time.Time)
		if !ok || !st.UTC().Equal(updatedSetupTime) {
			continue
		}
		countST++
		if got := utils.IfaceAsString(row["account"]); got != "extraInfo" {
			t.Errorf("updated row account = %q, want extraInfo", got)
		}
	}
	if countST != 1 {
		t.Errorf("expected 1 row with updated setup_time, got %d", countST)
	}
}

func TestERSSQLFiltersErr(t *testing.T) {
	_ = openTestDB(t, cdr1, cdr2, cdr3)

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
		"sqlConnMaxLifetime": "5s"
	}
},
"ers": {
	"enabled": true,
	"sessions_conns": ["*localhost"],
	"readers": [
		{
			"id": "mysql",
			"type": "*sql",
			"run_delay": "1m",
			"source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"opts": {
				"sqlDBName": "cgrates2",
				"sqlTableName": "cdrs",
				"sqlBatchSize": 2,
				"sqlDeleteIndexedFields": ["id"]
			},
			"start_delay": "100ms",
			"processed_path": "*delete",
			"tenant": "cgrates.org",
			"filters": [
				"*gt:~*req.answer_time:-168h",
				"FLTR_SQL_RatingID",
				"*string:~*vars.*readerID:mysql",
				"FLTR_VARS",
				"*notempty:~*vars.*readerID:''"
			],
			"flags": ["*dryrun"],
			"fields": [
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
				{"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.usage", "mandatory": true}
			]
		}
	]
}
}`

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON:       jsonCfg,
		DBCfg:            getDBCfg(t),
		TpFiles:          sqlFilterTPFiles,
		LogBuffer:        buf,
		GracefulShutdown: true,
	}
	ng.Run(t)

	errLine := `[ERROR] <ERs> error: <value of filter <*notempty:~*vars.*readerID:''> is not empty <''>>`
	waitForERsLog(t, buf, errLine, 2*time.Second)
}

func TestERSSQLFilterUnquote(t *testing.T) {
	cdr := &engine.CDR{
		CGRID:       utils.Sha1("cdr1", timeStart.String()),
		RunID:       utils.MetaDefault,
		ToR:         utils.MetaVoice,
		OriginID:    "cdr1",
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
		Usage:       10 * time.Second,
		ExtraFields: map[string]string{"*rateID": "RateID2"},
	}
	_ = openTestDB(t, cdr)

	jsonCfg := `{
"general": {
	"reply_timeout": "10s",
	"default_timezone": "UTC"
},
"apiers": {
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
			"source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"opts": {
				"sqlDBName": "cgrates2",
				"sqlTableName": "cdrs",
				"sqlBatchSize": 10
			},
			"start_delay": "100ms",
			"tenant": "cgrates.org",
			"filters": [
				"*eq:~*req.extra_fields.*rateID:RateID2",
				"*string:~*vars.*readerID:mysql"
			],
			"flags": ["*dryrun"],
			"fields": [
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account", "mandatory": true}
			]
		}
	]
}
}`

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON:       jsonCfg,
		DBCfg:            getDBCfg(t),
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

// MariaDB 10.11.14 (MDEV-37428): JSON_VALUE returns NULL for empty
// strings, so JSON_EXTRACT is needed.
func TestERSSQLFilterMetaEmpty(t *testing.T) {
	emptyCDR := &engine.CDR{
		CGRID:       utils.Sha1("empty", timeStart.String()),
		RunID:       utils.MetaDefault,
		ToR:         utils.MetaVoice,
		OriginID:    "empty",
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
		Usage:       10 * time.Second,
		ExtraFields: map[string]string{"EmptyField": ""},
	}
	filledCDR := &engine.CDR{
		CGRID:       utils.Sha1("filled", timeStart.String()),
		RunID:       utils.MetaDefault,
		ToR:         utils.MetaVoice,
		OriginID:    "filled",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1002",
		Subject:     "1002",
		Destination: "1003",
		SetupTime:   timeStart,
		AnswerTime:  timeStart,
		Usage:       10 * time.Second,
		ExtraFields: map[string]string{"EmptyField": "not_empty"},
	}
	_ = openTestDB(t, emptyCDR, filledCDR)

	jsonCfg := `{
"general": {
	"reply_timeout": "10s",
	"default_timezone": "UTC"
},
"apiers": {
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
			"source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",
			"opts": {
				"sqlDBName": "cgrates2",
				"sqlTableName": "cdrs",
				"sqlBatchSize": 10
			},
			"start_delay": "100ms",
			"tenant": "cgrates.org",
			"filters": [
				"*empty:~*req.extra_fields.EmptyField:",
				"*exists:~*req.extra_fields.EmptyField:",
				"*string:~*vars.*readerID:mysql"
			],
			"flags": ["*dryrun"],
			"fields": [
				{"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.account", "mandatory": true}
			]
		}
	]
}
}`

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON:       jsonCfg,
		DBCfg:            getDBCfg(t),
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
