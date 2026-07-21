//go:build flaky

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
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
		testSqlEeResetDBs,
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
	Cgrid      string
	AnswerTime time.Time
	Usage      int64
	Cost       float64
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
	if sqlEeCfg, err = config.NewCGRConfigFromPath(context.Background(), sqlEeCfgPath); err != nil {
		t.Error(err)
	}
}

func testSqlEeResetDBs(t *testing.T) {
	if err := engine.InitDB(sqlEeCfg); err != nil {
		t.Fatal(err)
	}
}

func testSqlEeStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sqlEeCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSqlEeRPCConn(t *testing.T) {
	sqlEeRpc = engine.NewRPCClient(t, sqlEeCfg.ListenCfg(), *utils.Encoding)
}

func testSqlEeExportEventFull(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterFull"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{

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
				utils.Cost:         1.01,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				utils.RunID:        utils.MetaDefault,
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
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"SQLExporterPartial"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Event: map[string]any{

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
				utils.Cost:         123,
				"ExtraFields": map[string]string{"extra1": "val_extra1",
					"extra2": "val_extra2", "extra3": "val_extra3"},
			},
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("asd", time.Unix(1383813745, 0).UTC().String()),
				utils.RunID:        utils.MetaDefault,
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
	_, _, err := openDB(dialect, &config.EventExporterOpts{
		SQLMaxIdleConns: utils.IntPointer(2),
	})
	if err != nil {
		t.Error(err)
	}
}

func TestOpenDB2(t *testing.T) {
	dialect := mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		"cgrates", "CGRateS.org", "127.0.0.1", "3306", "cgrates"))
	_, _, err := openDB(dialect, &config.EventExporterOpts{
		SQLMaxOpenConns: utils.IntPointer(2),
	})
	if err != nil {
		t.Error(err)
	}
}

func TestOpenDB3(t *testing.T) {
	dialect := mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		"cgrates", "CGRateS.org", "127.0.0.1", "3306", "cgrates"))
	_, _, err := openDB(dialect, &config.EventExporterOpts{
		SQLConnMaxLifetime: utils.DurationPointer(2),
	})
	if err != nil {
		t.Error(err)
	}
}

func TestSQLExportEvent1(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLTableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLDBName = utils.StringPointer("cgrates")
	cgrCfg.EEsCfg().Exporters[0].ExportPath = `mysql://cgrates:CGRateS.org@127.0.0.1:3306`
	sqlEe, err := NewSQLEe(cgrCfg.EEsCfg().Exporters[0], nil)
	if err != nil {
		t.Error(err)
	}
	if err := sqlEe.Connect(); err != nil {
		t.Fatal(err)
	}
	if err := sqlEe.ExportEvent(context.Background(), &sqlPosterRequest{Querry: "INSERT INTO cdrs VALUES (); ", Values: []any{}}, ""); err != nil {
		t.Error(err)
	}
	sqlEe.Close()
}
