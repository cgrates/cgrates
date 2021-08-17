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
	"github.com/cgrates/cgrates/utils"
)

func NewSQLEe(cfg *config.EventExporterCfg,
	dc *utils.SafeMapStorage) (sqlEe *SQLEe, err error) {
	sqlEe = &SQLEe{
		cfg:  cfg,
		dc:   dc,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	err = sqlEe.initDialector()
	return
}

// SQLEe implements EventExporter interface for SQL
type SQLEe struct {
	cfg   *config.EventExporterCfg
	dc    *utils.SafeMapStorage
	db    *gorm.DB
	sqldb *sql.DB
	reqs  *concReq

	dialect   gorm.Dialector
	tableName string
}

type sqlPosterRequest struct {
	Querry string
	Values []interface{}
}

func (sqlEe *SQLEe) initDialector() (err error) {
	var u *url.URL
	// var err error
	if u, err = url.Parse(strings.TrimPrefix(sqlEe.Cfg().ExportPath, utils.Meta)); err != nil {
		return
	}
	password, _ := u.User.Password()

	dbname := utils.SQLDefaultDBName
	if vals, has := sqlEe.Cfg().Opts[utils.SQLDBNameOpt]; has {
		dbname = utils.IfaceAsString(vals)
	}
	ssl := utils.SQLDefaultSSLMode
	if vals, has := sqlEe.Cfg().Opts[utils.SSLModeCfg]; has {
		ssl = utils.IfaceAsString(vals)
	}
	// tableName is mandatory in opts
	if iface, has := sqlEe.Cfg().Opts[utils.SQLTableNameOpt]; !has {
		return utils.NewErrMandatoryIeMissing(utils.SQLTableNameOpt)
	} else {
		sqlEe.tableName = utils.IfaceAsString(iface)
	}

	// var dialect gorm.Dialector
	switch u.Scheme {
	case utils.MySQL:
		sqlEe.dialect = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			u.User.Username(), password, u.Hostname(), u.Port(), dbname))
	case utils.Postgres:
		sqlEe.dialect = postgres.Open(fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", u.Hostname(), u.Port(), dbname, u.User.Username(), password, ssl))
	default:
		return fmt.Errorf("db type <%s> not supported", u.Scheme)
	}
	return
}

func openDB(dialect gorm.Dialector, opts map[string]interface{}) (db *gorm.DB, sqlDB *sql.DB, err error) {
	if db, err = gorm.Open(dialect, &gorm.Config{AllowGlobalUpdate: true}); err != nil {
		return
	}
	if sqlDB, err = db.DB(); err != nil {
		return
	}

	if iface, has := opts[utils.SQLMaxIdleConnsCfg]; has {
		val, err := utils.IfaceAsTInt64(iface)
		if err != nil {
			return nil, nil, err
		}
		sqlDB.SetMaxIdleConns(int(val))
	}
	if iface, has := opts[utils.SQLMaxOpenConns]; has {
		val, err := utils.IfaceAsTInt64(iface)
		if err != nil {
			return nil, nil, err
		}
		sqlDB.SetMaxOpenConns(int(val))
	}
	if iface, has := opts[utils.SQLMaxConnLifetime]; has {
		val, err := utils.IfaceAsDuration(iface)
		if err != nil {
			return nil, nil, err
		}
		sqlDB.SetConnMaxLifetime(val)
	}

	return
}

func (sqlEe *SQLEe) Cfg() *config.EventExporterCfg { return sqlEe.cfg }

func (sqlEe *SQLEe) Connect() (err error) {
	if sqlEe.db == nil || sqlEe.sqldb == nil {
		sqlEe.db, sqlEe.sqldb, err = openDB(sqlEe.dialect, sqlEe.Cfg().Opts)
	}
	return
}

func (sqlEe *SQLEe) ExportEvent(req interface{}, _ string) error {
	sqlEe.reqs.get()
	defer sqlEe.reqs.done()
	sReq := req.(*sqlPosterRequest)
	return sqlEe.db.Table(sqlEe.tableName).Exec(sReq.Querry, sReq.Values...).Error
}

func (sqlEe *SQLEe) Close() error { return sqlEe.sqldb.Close() }

func (sqlEe *SQLEe) GetMetrics() *utils.SafeMapStorage { return sqlEe.dc }

func (sqlEe *SQLEe) PrepareMap(map[string]interface{}) (interface{}, error) { return nil, nil }

func (sqlEe *SQLEe) PrepareOrderMap(mp *utils.OrderedNavigableMap) (interface{}, error) {
	var vals []interface{}
	var colNames []string
	for el := mp.GetFirstElement(); el != nil; el = el.Next() {
		nmIt, _ := mp.Field(el.Value)
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
		sqlQuery = fmt.Sprintf("INSERT INTO %s VALUES (%s); ",
			sqlEe.tableName,
			strings.Join(sqlValues, ","))
	} else {
		sqlQuery = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s); ",
			sqlEe.tableName,
			strings.Join(colNames, ", "),
			strings.Join(sqlValues, ","))
	}
	return &sqlPosterRequest{
		Querry: sqlQuery,
		Values: vals,
	}, nil
}
