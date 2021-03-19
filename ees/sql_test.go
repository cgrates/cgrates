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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
)

func TestSqlID(t *testing.T) {
	sqlEe := &SQLEe{
		id: "3",
	}
	if rcv := sqlEe.ID(); !reflect.DeepEqual(rcv, "3") {
		t.Errorf("Expected %+v but got %+v", "3", rcv)
	}
}

func TestSqlGetMetrics(t *testing.T) {
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	sqlEe := &SQLEe{
		dc: dc,
	}

	if rcv := sqlEe.GetMetrics(); !reflect.DeepEqual(rcv, sqlEe.dc) {
		t.Errorf("Expected %+v but got %+v", utils.ToJSON(rcv), utils.ToJSON(sqlEe.dc))
	}
}

func TestNewSQLeUrl(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.SQLTableName] = "expTable"
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.SQLDBName] = "postgres"
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.SQLSSLMode] = "test"
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	sqlEe := &SQLEe{id: cgrCfg.EEsCfg().Exporters[0].ID,
		cgrCfg: cgrCfg, cfgIdx: 0, filterS: filterS, dc: dc}
	_, err = sqlEe.NewSQLEeUrl(cgrCfg)
	errExpect := "db type <> not supported"
	if err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestNewSQLeUrlSQL(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.SQLTableName] = "expTable"
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.SQLDBName] = "mysql"
	cgrCfg.EEsCfg().Exporters[0].ExportPath = `mysql://cgrates:CGRateS.org@127.0.0.1:3306`
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	sqlEe := &SQLEe{id: cgrCfg.EEsCfg().Exporters[0].ID,
		cgrCfg: cgrCfg, cfgIdx: 0, filterS: filterS, dc: dc}
	dialectExpect := mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		"cgrates", "CGRateS.org", "127.0.0.1", "3306", "mysql"))
	if dialect, err := sqlEe.NewSQLEeUrl(cgrCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dialect, dialectExpect) {
		t.Errorf("Expected %v but received %v", utils.ToJSON(dialectExpect), utils.ToJSON(dialect))
	}
}

func TestNewSQLeUrlPostgres(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.SQLTableName] = "expTable"
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.SQLDBName] = "postgres"
	cgrCfg.EEsCfg().Exporters[0].ExportPath = `postgres://cgrates:CGRateS.org@127.0.0.1:3306`
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	sqlEe := &SQLEe{id: cgrCfg.EEsCfg().Exporters[0].ID,
		cgrCfg: cgrCfg, cfgIdx: 0, filterS: filterS, dc: dc}
	dialectExpect := postgres.Open(fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		"127.0.0.1", "3306", "postgres", "cgrates", "CGRateS.org", utils.SQLDefaultSSLMode))
	if dialect, err := sqlEe.NewSQLEeUrl(cgrCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dialect, dialectExpect) {
		t.Errorf("Expected %v but received %v", utils.ToJSON(dialectExpect), utils.ToJSON(dialect))
	}
}

func TestNewSQLeExportPathError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.SQLTableName] = "expTable"
	cgrCfg.EEsCfg().Exporters[0].Opts[utils.SQLDBName] = "postgres"
	cgrCfg.EEsCfg().Exporters[0].ExportPath = ":foo"
	newIDb := engine.NewInternalDB(nil, nil, true)
	newDM := engine.NewDataManager(newIDb, cgrCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cgrCfg, nil, newDM)
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	sqlEe := &SQLEe{id: cgrCfg.EEsCfg().Exporters[0].ID,
		cgrCfg: cgrCfg, cfgIdx: 0, filterS: filterS, dc: dc}
	errExpect := `parse ":foo": missing protocol scheme`
	if _, err := sqlEe.NewSQLEeUrl(cgrCfg); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}
