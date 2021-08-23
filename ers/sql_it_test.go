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

package ers

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	sqlCfgPath string
	sqlCfg     *config.CGRConfig
	sqlTests   = []func(t *testing.T){
		testSQLInitConfig,
		testSQLInitDBs,
		testSQLInitCdrDb,
		testSQLInitDB,
		testSQLReader,
		testSQLEmptyTable,
		testSQLPoster,

		testSQLInitDB,
		testSQLReader2,

		testSQLStop,
	}
	cdr = &engine.CDR{
		CGRID: "CGRID",
		RunID: "RunID",
	}
	db           *gorm.DB
	dbConnString = "cgrates:CGRateS.org@tcp(127.0.0.1:3306)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'"
)

func TestSQL(t *testing.T) {
	// sqlCfgPath = path.Join(*dataDir, "conf", "samples", "ers_reload", "disabled")
	for _, test := range sqlTests {
		t.Run("TestSQL", test)
	}
}

func testSQLInitConfig(t *testing.T) {
	var err error
	if sqlCfg, err = config.NewCGRConfigFromJsonStringWithDefaults(`{
		"stor_db": {
			"db_password": "CGRateS.org",
		},
		"ers": {									// EventReaderService
			"enabled": true,						// starts the EventReader service: <true|false>
			"readers": [
				{
					"id": "mysql",										// identifier of the EventReader profile
					"type": "*sql",							// reader type <*file_csv>
					"run_delay": "1",									// sleep interval in seconds between consecutive runs, -1 to use automation via inotify or 0 to disable running all together
					"concurrent_requests": 1024,						// maximum simultaneous requests/files to process, 0 for unlimited
					"source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306?db_name=cgrates2",					// read data from this path
					"processed_path": "db_name=cgrates2&table_name=cdrs2",	// move processed data here
					"tenant": "cgrates.org",							// tenant used by import
					"filters": [],										// limit parsing based on the filters
					"flags": [],										// flags to influence the event processing
					"fields":[									// import fields template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
						{"tag": "CGRID", "type": "*composed", "value": "~*req.cgrid", "path": "*cgreq.CGRID"},
					],
				},
			],
		},
		}`); err != nil {
		t.Fatal(err)
	}
	utils.Newlogger(utils.MetaSysLog, sqlCfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
}

func testSQLInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(sqlCfg); err != nil {
		t.Fatal(err)
	}
}

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

func (_ *testModelSql) TableName() string {
	return "cdrs2"
}

func testSQLInitDBs(t *testing.T) {
	var err error
	if db, err = gorm.Open("mysql", fmt.Sprintf(dbConnString, "cgrates")); err != nil {
		t.Fatal(err)
	}

	if _, err = db.DB().Exec(`CREATE DATABASE IF NOT EXISTS cgrates2;`); err != nil {
		t.Fatal(err)
	}
}
func testSQLInitDB(t *testing.T) {
	cdr.CGRID = utils.UUIDSha1Prefix()
	var err error
	db, err = gorm.Open("mysql", fmt.Sprintf(dbConnString, "cgrates2"))
	if err != nil {
		t.Fatal(err)
	}
	if !db.HasTable("cdrs") {
		db = db.CreateTable(new(engine.CDRsql))
	}
	if !db.HasTable("cdrs2") {
		db = db.CreateTable(new(testModelSql))
	}
	db = db.Table(utils.CDRsTBL)
	tx := db.Begin()
	cdrSql := cdr.AsCDRsql()
	cdrSql.CreatedAt = time.Now()
	saved := tx.Save(cdrSql)
	if saved.Error != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	tx.Commit()
}

func testSQLReader(t *testing.T) {
	rdrEvents = make(chan *erEvent, 1)
	rdrErr = make(chan error, 1)
	rdrExit = make(chan struct{}, 1)
	sqlER, err := NewEventReader(sqlCfg, 1, rdrEvents, rdrErr, new(engine.FilterS), rdrExit)
	if err != nil {
		t.Fatal(err)
	}
	sqlER.Serve()

	select {
	case err = <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
		if ev.rdrCfg.ID != "mysql" {
			t.Errorf("Expected 'mysql' received `%s`", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Time:   ev.cgrEvent.Time,
			Event: map[string]interface{}{
				"CGRID": cdr.CGRID,
			},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout")
	}
}

func testSQLEmptyTable(t *testing.T) {
	time.Sleep(10 * time.Millisecond)
	rows, err := db.Table(utils.CDRsTBL).Select("*").Rows()
	if err != nil {
		t.Fatal(err)
	}
	colNames, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		columns := make([]interface{}, len(colNames))
		columnPointers := make([]interface{}, len(colNames))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err = rows.Scan(columnPointers...); err != nil {
			t.Fatal(err)
		}
		msg := make(map[string]interface{})
		for i, colName := range colNames {
			msg[colName] = columns[i]
		}
		t.Fatal("Expected empty table ", utils.ToJSON(msg))
	}
}

func testSQLReader2(t *testing.T) {
	select {
	case err := <-rdrErr:
		t.Error(err)
	case ev := <-rdrEvents:
		if ev.rdrCfg.ID != "mysql" {
			t.Errorf("Expected 'mysql' received `%s`", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Time:   ev.cgrEvent.Time,
			Event: map[string]interface{}{
				"CGRID": cdr.CGRID,
			},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout")
	}
}

func testSQLPoster(t *testing.T) {
	rows, err := db.Table("cdrs2").Select("*").Rows()
	if err != nil {
		t.Fatal(err)
	}
	colNames, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		columns := make([]interface{}, len(colNames))
		columnPointers := make([]interface{}, len(colNames))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err = rows.Scan(columnPointers...); err != nil {
			t.Fatal(err)
		}
		msg := make(map[string]interface{})
		for i, colName := range colNames {
			msg[colName] = columns[i]
		}
		db.Table("cdrs2").Delete(msg)
		if cgrid := utils.IfaceAsString(msg["cgrid"]); cgrid != cdr.CGRID {
			t.Errorf("Expected: %s ,receieved: %s", cgrid, cdr.CGRID)
		}
	}
}

func testSQLStop(t *testing.T) {
	rdrExit <- struct{}{}
	db = db.DropTable("cdrs2")
	if err := db.Close(); err != nil {
		t.Error(err)
	}
}
