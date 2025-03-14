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

package ees

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSqlGetMetrics(t *testing.T) {
	em := utils.NewExporterMetrics("", time.Local)
	sqlEe := &SQLEe{
		em: em,
	}
	if rcv := sqlEe.GetMetrics(); !reflect.DeepEqual(rcv, sqlEe.em) {
		t.Errorf("Expected %+v but got %+v", utils.ToJSON(rcv), utils.ToJSON(sqlEe.em))
	}
}

func TestNewSQLeUrl(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.TableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.DBName = utils.StringPointer("postgres")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.PgSSLMode = utils.StringPointer("test")
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
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.TableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.DBName = utils.StringPointer("mysql")
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
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.TableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.DBName = utils.StringPointer("postgres")
	cgrCfg.EEsCfg().Exporters[0].ExportPath = `postgres://cgrates:CGRateS.org@127.0.0.1:3306`
	sqlEe := &SQLEe{
		cfg:  cgrCfg.EEsCfg().Exporters[0],
		reqs: newConcReq(0),
	}
	dialectExpect := postgres.Open(fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		"127.0.0.1", "3306", "postgres", "cgrates", "CGRateS.org", utils.SQLDefaultSSLMode))
	if err := sqlEe.initDialector(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sqlEe.dialect, dialectExpect) {
		t.Errorf("Expected %v but received %v", utils.ToJSON(dialectExpect), utils.ToJSON(sqlEe.dialect))
	}
}

func TestNewSQLeExportPathError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.TableName = utils.StringPointer("expTable")
	cgrCfg.EEsCfg().Exporters[0].Opts.SQL.DBName = utils.StringPointer("postgres")
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
	_, _, err := openDB(mckDialect, &config.SQLOpts{})
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
	_, _, err := openDB(mckDialect, &config.SQLOpts{})
	errExpect := "NOT_FOUND"
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	logger.Default = tmp
}

func TestPrepareMap(t *testing.T) {
	sqlEe := &SQLEe{}
	event := &utils.CGREvent{}
	result, err := sqlEe.PrepareMap(event)
	if err != nil {
		t.Errorf("PrepareMap() returned an error: %v", err)
	}
	exp := &sqlPosterRequest{
		Querry: "INSERT INTO  (``) VALUES ();",
		Values: make([]any, 0),
	}
	if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func TestExportEvent(t *testing.T) {
	sqlEe := &SQLEe{
		db:   nil,
		reqs: &concReq{},
	}
	err := sqlEe.ExportEvent(&sqlPosterRequest{}, "")
	if err != utils.ErrDisconnected {
		t.Errorf("Expected error %v, got %v", utils.ErrDisconnected, err)
	}

}
