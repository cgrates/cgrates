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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

		testSQLAddData,
		testSQLReader2,

		testSQLStop,
	}
	sqlTests2 = []func(t *testing.T){
		testSQLInitConfig2,
		testSQLInitDBs2,
		testSQLInitCdrDb2,
		testSQLInitDB2,
		testSQLReader3,
		testSQLEmptyTable2,
		testSQLPoster2,

		testSQLAddData2,
		testSQLReader4,

		testSQLStop2,
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
	if sqlCfg, err = config.NewCGRConfigFromJSONStringWithDefaults(`{
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
					"source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",					// read data from this path
					"opts": {
						"sqlDBName":"cgrates2",
						"sqlDBNameProcessed":"cgrates2",
						"sqlTableNameProcessed":"cdrs2",
					},
					"processed_path": "",	// move processed data here
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
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, sqlCfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
}

func testSQLInitCdrDb(t *testing.T) {
	if err := engine.InitStorDB(sqlCfg); err != nil {
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
}
func testSQLInitDB(t *testing.T) {
	cdr.CGRID = utils.UUIDSha1Prefix()
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
	if !tx.Migrator().HasTable("cdrs2") {
		if err = tx.Migrator().CreateTable(new(testModelSql)); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
	}
	tx.Commit()
	tx = db.Begin()
	tx = tx.Table(utils.CDRsTBL)
	cdrSql := cdr.AsCDRsql()
	cdrSql.CreatedAt = time.Now()
	saved := tx.Save(cdrSql)
	if saved.Error != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	tx.Commit()
	time.Sleep(10 * time.Millisecond)
	var result int64
	db.Table(utils.CDRsTBL).Count(&result)
	if result != 1 {
		t.Fatal("Expected table to have only one result ", result)
	}
}

func testSQLAddData(t *testing.T) {
	tx := db.Begin()
	tx = tx.Table(utils.CDRsTBL)
	cdrSql := cdr.AsCDRsql()
	cdrSql.CreatedAt = time.Now()
	saved := tx.Save(cdrSql)
	if saved.Error != nil {
		tx.Rollback()
		t.Fatal(saved.Error)
	}
	tx.Commit()
	time.Sleep(10 * time.Millisecond)
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
			APIOpts: map[string]interface{}{},
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
	var result int64
	db.Table(utils.CDRsTBL).Count(&result)
	if result != 0 {
		t.Fatal("Expected empty table ", result)
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
			APIOpts: map[string]interface{}{},
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
	close(rdrExit)
	if err := db.Migrator().DropTable("cdrs2"); err != nil {
		t.Fatal(err)
	}
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

}

func TestSQLReaderServeBadTypeErr(t *testing.T) {
	rdr := &SQLEventReader{
		connType: "badType",
	}
	expected := "db type <badType> not supported"
	err := rdr.Serve()
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", expected, err)
	}
}

func TestSQL2(t *testing.T) {
	// sqlCfgPath = path.Join(*dataDir, "conf", "samples", "ers_reload", "disabled")
	for _, test := range sqlTests2 {
		t.Run("TestSQL", test)
	}
}

func testSQLInitConfig2(t *testing.T) {
	var err error
	if sqlCfg, err = config.NewCGRConfigFromJSONStringWithDefaults(`{
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
					"source_path": "*mysql://cgrates:CGRateS.org@127.0.0.1:3306",					// read data from this path
					"opts": {
						"sqlDBName":"cgrates2",
						"sqlDBNameProcessed":"cgrates2",
						"sqlTableNameProcessed":"cdrs2",
					},
					"processed_path": "",	// move processed data here
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
	utils.Logger, _ = utils.Newlogger(utils.MetaSysLog, sqlCfg.GeneralCfg().NodeID)
	utils.Logger.SetLogLevel(7)
}

func testSQLInitCdrDb2(t *testing.T) {
	if err := engine.InitStorDB(sqlCfg); err != nil {
		t.Fatal(err)
	}
}

func testSQLInitDBs2(t *testing.T) {
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
}
func testSQLInitDB2(t *testing.T) {
	cdr.CGRID = utils.UUIDSha1Prefix()
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
	if !tx.Migrator().HasTable("cdrs2") {
		if err = tx.Migrator().CreateTable(new(testModelSql)); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
	}
	tx.Commit()
	tx = db.Begin()
	tx = tx.Table(utils.CDRsTBL)
	cdrSql := cdr.AsCDRsql()
	cdrSql.CreatedAt = time.Now()
	saved := tx.Save(cdrSql)
	if saved.Error != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	tx.Commit()
	time.Sleep(10 * time.Millisecond)
	var result int64
	db.Table(utils.CDRsTBL).Count(&result)
	if result != 1 {
		t.Fatal("Expected table to have only one result ", result)
	}
}

func testSQLAddData2(t *testing.T) {
	tx := db.Begin()
	tx = tx.Table(utils.CDRsTBL)
	cdrSql := cdr.AsCDRsql()
	cdrSql.CreatedAt = time.Now()
	saved := tx.Save(cdrSql)
	if saved.Error != nil {
		tx.Rollback()
		t.Fatal(saved.Error)
	}
	tx.Commit()
	time.Sleep(10 * time.Millisecond)
}
func testSQLReader3(t *testing.T) {
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
			APIOpts: map[string]interface{}{},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout")
	}
}

func testSQLEmptyTable2(t *testing.T) {
	time.Sleep(10 * time.Millisecond)
	var result int64
	db.Table(utils.CDRsTBL).Count(&result)
	if result != 0 {
		t.Fatal("Expected empty table ", result)
	}
}

func testSQLReader4(t *testing.T) {
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
			APIOpts: map[string]interface{}{},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Errorf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout")
	}
}

func testSQLPoster2(t *testing.T) {
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

func testSQLStop2(t *testing.T) {
	close(rdrExit)
	if err := db.Migrator().DropTable("cdrs2"); err != nil {
		t.Fatal(err)
	}
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

}

func TestSQLProcessMessageError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	testSQLEventReader := &SQLEventReader{
		cgrCfg:        cfg,
		cfgIdx:        0,
		fltrS:         &engine.FilterS{},
		connString:    "",
		connType:      "",
		tableName:     "testName",
		expConnString: "",
		expConnType:   utils.Postgres,
		expTableName:  "",
		rdrEvents:     nil,
		rdrExit:       nil,
		rdrErr:        nil,
		cap:           nil,
	}

	msgTest := map[string]interface{}{}
	err := testSQLEventReader.processMessage(msgTest)
	expected := "NOT_FOUND:ToR"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestSQLSetURLError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	testSQLEventReader := &SQLEventReader{
		cgrCfg:        cfg,
		cfgIdx:        0,
		fltrS:         &engine.FilterS{},
		connString:    "",
		connType:      "",
		tableName:     "testName",
		expConnString: "",
		expConnType:   utils.Postgres,
		expTableName:  "",
		rdrEvents:     nil,
		rdrExit:       nil,
		rdrErr:        nil,
		cap:           nil,
	}
	err := testSQLEventReader.setURL("http://user^:passwo^rd@foo.com/", "", nil)
	expected := `parse "http://user^:passwo^rd@foo.com/": net/url: invalid userinfo`
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestSQLSetURLError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	testSQLEventReader := &SQLEventReader{
		cgrCfg:        cfg,
		cfgIdx:        0,
		fltrS:         &engine.FilterS{},
		connString:    "",
		connType:      "",
		tableName:     "testName",
		expConnString: "",
		expConnType:   utils.Postgres,
		expTableName:  "",
		rdrEvents:     nil,
		rdrExit:       nil,
		rdrErr:        nil,
		cap:           nil,
	}
	err := testSQLEventReader.setURL("*mysql://cgrates:CGRateS.org@127.0.0.1:3306", "http://user^:passwo^rd@foo.com/", nil)
	expected := `parse "http://user^:passwo^rd@foo.com/": net/url: invalid userinfo`
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

func TestErsSqlPostCDRS(t *testing.T) {
	tmp := logger.Default
	logger.Default = logger.Default.LogMode(logger.Silent)
	cfg := config.NewDefaultCGRConfig()
	fltr := &engine.FilterS{}
	reader := cfg.ERsCfg().Readers[0].Clone()
	reader.Type = utils.MetaSQL
	reader.ID = "file_reader"
	reader.ConcurrentReqs = -1
	reader.Opts = map[string]interface{}{"db_name": "cgrates2"}
	reader.SourcePath = "*mysql://cgrates:CGRateS.org@127.0.0.1:3306"
	reader.ProcessedPath = ""
	cfg.ERsCfg().Readers = append(cfg.ERsCfg().Readers, reader)
	if len(cfg.ERsCfg().Readers) != 2 {
		t.Errorf("Expecting: <2>, received: <%+v>", len(cfg.ERsCfg().Readers))
	}
	sqlEvReader, err := NewSQLEventReader(cfg, 1, nil, nil, fltr, nil)
	if err != nil {
		t.Errorf("Expecting: <nil>, received: <%+v>", err)
	}
	sqlEvReader.(*SQLEventReader).expConnType = utils.MySQL
	result := sqlEvReader.(*SQLEventReader).postCDR([]interface{}{})
	expected := "Error 1045: Access denied for user ''@'localhost' (using password: NO)"
	if result == nil {
		t.Errorf("\nExpected: <%+v>, \nreceived: <%+v>", expected, result)
	}
	logger.Default = tmp
}
