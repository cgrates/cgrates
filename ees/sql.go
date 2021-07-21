/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewSQLEe(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc *utils.SafeMapStorage) (sqlEe *SQLEe, err error) {
	sqlEe = &SQLEe{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS, dc: dc}

	dialect, err := sqlEe.NewSQLEeUrl(cgrCfg)
	if err != nil {
		return
	}
	sqlEe.db, sqlEe.sqldb, err = openDB(cgrCfg, cfgIdx, dialect)
	return
}

// SQLEe implements EventExporter interface for SQL
type SQLEe struct {
	id      string
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	db      *gorm.DB
	sqldb   *sql.DB

	tableName string

	dc *utils.SafeMapStorage
}

func (sqlEe *SQLEe) NewSQLEeUrl(cgrCfg *config.CGRConfig) (dialect gorm.Dialector, err error) {
	var u *url.URL
	// var err error
	if u, err = url.Parse(strings.TrimPrefix(cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].ExportPath, utils.Meta)); err != nil {
		return
	}
	password, _ := u.User.Password()

	dbname := utils.SQLDefaultDBName
	if vals, has := cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].Opts[utils.SQLDBNameOpt]; has {
		dbname = utils.IfaceAsString(vals)
	}
	ssl := utils.SQLDefaultSSLMode
	if vals, has := cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].Opts[utils.SSLModeCfg]; has {
		ssl = utils.IfaceAsString(vals)
	}
	// tableName is mandatory in opts
	if iface, has := cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].Opts[utils.SQLTableNameOpt]; !has {
		return nil, utils.NewErrMandatoryIeMissing(utils.SQLTableNameOpt)
	} else {
		sqlEe.tableName = utils.IfaceAsString(iface)
	}

	// var dialect gorm.Dialector
	switch u.Scheme {
	case utils.MySQL:
		dialect = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			u.User.Username(), password, u.Hostname(), u.Port(), dbname))
	case utils.Postgres:
		dialect = postgres.Open(fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", u.Hostname(), u.Port(), dbname, u.User.Username(), password, ssl))
	default:
		return nil, fmt.Errorf("db type <%s> not supported", u.Scheme)
	}
	return
}

func openDB(cgrCfg *config.CGRConfig, cfgIdx int, dialect gorm.Dialector) (db *gorm.DB, sqlDB *sql.DB, err error) {

	if db, err = gorm.Open(dialect, &gorm.Config{AllowGlobalUpdate: true}); err != nil {
		return
	}
	if sqlDB, err = db.DB(); err != nil {
		return
	}

	if iface, has := cgrCfg.EEsCfg().Exporters[cfgIdx].Opts[utils.SQLMaxIdleConnsCfg]; has {
		val, err := utils.IfaceAsTInt64(iface)
		if err != nil {
			return nil, nil, err
		}
		sqlDB.SetMaxIdleConns(int(val))
	}
	if iface, has := cgrCfg.EEsCfg().Exporters[cfgIdx].Opts[utils.SQLMaxOpenConns]; has {
		val, err := utils.IfaceAsTInt64(iface)
		if err != nil {
			return nil, nil, err
		}
		sqlDB.SetMaxOpenConns(int(val))
	}
	if iface, has := cgrCfg.EEsCfg().Exporters[cfgIdx].Opts[utils.SQLMaxConnLifetime]; has {
		val, err := utils.IfaceAsDuration(iface)
		if err != nil {
			return nil, nil, err
		}
		sqlDB.SetConnMaxLifetime(val)
	}

	return
}

// ID returns the identificator of this exporter
func (sqlEe *SQLEe) ID() string {
	return sqlEe.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (sqlEe *SQLEe) OnEvicted(_ string, _ interface{}) {
	sqlEe.sqldb.Close()
}

// ExportEvent implements EventExporter
func (sqlEe *SQLEe) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	defer func() {
		updateEEMetrics(sqlEe.dc, cgrEv.ID, cgrEv.Event, err != nil, utils.FirstNonEmpty(sqlEe.cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].Timezone,
			sqlEe.cgrCfg.GeneralCfg().DefaultTimezone))
	}()
	sqlEe.dc.Lock()
	sqlEe.dc.MapStorage[utils.NumberOfEvents] = sqlEe.dc.MapStorage[utils.NumberOfEvents].(int64) + 1
	sqlEe.dc.Unlock()

	var vals []interface{}
	var colNames []string
	oNm := map[string]*utils.OrderedNavigableMap{
		utils.MetaExp: utils.NewOrderedNavigableMap(),
	}
	eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
		utils.MetaReq:  utils.MapStorage(cgrEv.Event),
		utils.MetaDC:   sqlEe.dc,
		utils.MetaOpts: utils.MapStorage(cgrEv.APIOpts),
		utils.MetaCfg:  sqlEe.cgrCfg.GetDataProvider(),
	}, utils.FirstNonEmpty(cgrEv.Tenant, sqlEe.cgrCfg.GeneralCfg().DefaultTenant),
		sqlEe.filterS, oNm)
	if err = eeReq.SetFields(sqlEe.cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].ContentFields()); err != nil {
		return
	}

	for el := eeReq.ExpData[utils.MetaExp].GetFirstElement(); el != nil; el = el.Next() {
		nmIt, _ := eeReq.ExpData[utils.MetaExp].Field(el.Value)
		pathWithoutIndex := strings.Join(el.Value[:len(el.Value)-1], utils.NestingSep) // remove the index path.index
		if pathWithoutIndex != utils.MetaRow {
			colNames = append(colNames, pathWithoutIndex)
		}
		vals = append(vals, nmIt.Data)
	}

	sqlValues := make([]string, len(vals))
	for i := range vals {
		sqlValues[i] = "?"
	}

	var sqlQuery string
	if len(colNames) != len(vals) {
		sqlQuery = fmt.Sprintf("INSERT INTO %s VALUES (%s); ", sqlEe.tableName, strings.Join(sqlValues, ","))
	} else {
		colNamesStr := "(" + strings.Join(colNames, ", ") + ")"
		sqlQuery = fmt.Sprintf("INSERT INTO %s %s VALUES (%s); ", sqlEe.tableName, colNamesStr, strings.Join(sqlValues, ","))
	}

	sqlEe.db.Table(sqlEe.tableName).Exec(sqlQuery, vals...)

	return
}

func (sqlEe *SQLEe) GetMetrics() *utils.SafeMapStorage {
	return sqlEe.dc.Clone()
}
