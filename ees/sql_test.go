// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSqlGetMetrics(t *testing.T) {
	em, err := utils.NewExporterMetrics("", "Local")
	if err != nil {
		t.Fatal(err)
	}
	sqlEe := &SQLEe{
		em: em,
	}
	if rcv := sqlEe.GetMetrics(); !reflect.DeepEqual(rcv, sqlEe.em) {
		t.Errorf("Expected %+v but got %+v", utils.ToJSON(rcv), utils.ToJSON(sqlEe.em))
	}
}

func TestNewSQLeUrl(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLTableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLDBName = utils.StringPointer("postgres")
	cgrCfg.EEsCfg().Exporters[0].Opts.PgSSLMode = utils.StringPointer("test")
	sqlEe := &SQLEe{
		cfg:  cgrCfg.EEsCfg().Exporters[0],
		reqs: newConcReq(0),
	}
	errExpect := "db type <> not supported"
	if err := sqlEe.initDialector(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestNewSQLeUrlSQL(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLTableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLDBName = utils.StringPointer("mysql")
	cgrCfg.EEsCfg().Exporters[0].ExportPath = `mysql://cgrates:CGRateS.org@127.0.0.1:3306`
	sqlEe := &SQLEe{
		cfg:  cgrCfg.EEsCfg().Exporters[0],
		reqs: newConcReq(0),
	}
	dialectExpect := mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		"cgrates", "CGRateS.org", "127.0.0.1", "3306", "mysql"))
	if err := sqlEe.initDialector(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sqlEe.dialect, dialectExpect) {
		t.Errorf("Expected %v but received %v", utils.ToJSON(dialectExpect), utils.ToJSON(sqlEe.dialect))
	}
}

func TestNewSQLeUrlPostgres(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLTableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLDBName = utils.StringPointer("postgres")
	cgrCfg.EEsCfg().Exporters[0].ExportPath = `postgres://cgrates:CGRateS.org@127.0.0.1:3306`
	sqlEe := &SQLEe{
		cfg:  cgrCfg.EEsCfg().Exporters[0],
		reqs: newConcReq(0),
	}
	dialectExpect := postgres.Open(fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		"127.0.0.1", "3306", "postgres", "cgrates", "CGRateS.org", utils.SQLDefaultPgSSLMode))
	if err := sqlEe.initDialector(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sqlEe.dialect, dialectExpect) {
		t.Errorf("Expected %v but received %v", utils.ToJSON(dialectExpect), utils.ToJSON(sqlEe.dialect))
	}
}

func TestNewSQLeExportPathError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLTableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQLDBName = utils.StringPointer("postgres")
	cgrCfg.EEsCfg().Exporters[0].ExportPath = ":foo"
	sqlEe := &SQLEe{
		cfg:  cgrCfg.EEsCfg().Exporters[0],
		reqs: newConcReq(0),
	}
	errExpect := `parse ":foo": missing protocol scheme`
	if err := sqlEe.initDialector(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

type mockDialect2 struct {
	gorm.Dialector
}

func (mockDialect2) Initialize(db *gorm.DB) error { return nil }

func TestOpenDBError2(t *testing.T) {
	tmp := logger.Default
	logger.Default = logger.Default.LogMode(logger.Silent)
	mckDialect := new(mockDialect2)
	_, _, err := openDB(mckDialect, &config.EventExporterOpts{})
	errExpect := "invalid db"
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	logger.Default = tmp
}

type mockDialectErr struct {
	gorm.Dialector
}

func (mockDialectErr) Initialize(db *gorm.DB) error {
	return utils.ErrNotFound
}

func TestOpenDBError3(t *testing.T) {
	tmp := logger.Default
	logger.Default = logger.Default.LogMode(logger.Silent)
	mckDialect := new(mockDialectErr)
	_, _, err := openDB(mckDialect, &config.EventExporterOpts{})
	errExpect := "NOT_FOUND"
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	logger.Default = tmp
}
