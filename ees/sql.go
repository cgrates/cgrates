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

package ees

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
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
	sync.RWMutex
}

type sqlPosterRequest struct {
	Querry string
	Values []any
}

func (sqlEe *SQLEe) initDialector() (err error) {
	var u *url.URL
	// var err error
	if u, err = url.Parse(strings.TrimPrefix(sqlEe.Cfg().ExportPath, utils.Meta)); err != nil {
		return
	}
	password, _ := u.User.Password()

	dbname := utils.SQLDefaultDBName
	if sqlEe.Cfg().Opts.SQLDBName != nil {
		dbname = *sqlEe.Cfg().Opts.SQLDBName
	}
	ssl := utils.SQLDefaultPgSSLMode
	if sqlEe.Cfg().Opts.PgSSLMode != nil {
		ssl = *sqlEe.Cfg().Opts.PgSSLMode
	}
	// tableName is mandatory in opts
	if sqlEe.Cfg().Opts.SQLTableName != nil {
		sqlEe.tableName = *sqlEe.Cfg().Opts.SQLTableName
	} else {
		return utils.NewErrMandatoryIeMissing(utils.SQLTableNameOpt)
	}

	// var dialect gorm.Dialector
	switch u.Scheme {
	case utils.MySQL:
		sqlEe.dialect = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			u.User.Username(), password, u.Hostname(), u.Port(), dbname) + engine.AppendToMysqlDSNOpts(sqlEe.Cfg().Opts.MYSQLDSNParams))
	case utils.Postgres:
		sqlEe.dialect = postgres.Open(fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", u.Hostname(), u.Port(), dbname, u.User.Username(), password, ssl))
	default:
		return fmt.Errorf("db type <%s> not supported", u.Scheme)
	}
	return
}

func openDB(dialect gorm.Dialector, opts *config.EventExporterOpts) (db *gorm.DB, sqlDB *sql.DB, err error) {
	if db, err = gorm.Open(dialect, &gorm.Config{AllowGlobalUpdate: true}); err != nil {
		return
	}
	if sqlDB, err = db.DB(); err != nil {
		return
	}

	if opts.SQLMaxIdleConns != nil {
		sqlDB.SetMaxIdleConns(*opts.SQLMaxIdleConns)
	}
	if opts.SQLMaxOpenConns != nil {
		sqlDB.SetMaxOpenConns(*opts.SQLMaxOpenConns)
	}
	if opts.SQLConnMaxLifetime != nil {
		sqlDB.SetConnMaxLifetime(*opts.SQLConnMaxLifetime)
	}

	return
}

func (sqlEe *SQLEe) Cfg() *config.EventExporterCfg { return sqlEe.cfg }

func (sqlEe *SQLEe) Connect() (err error) {
	sqlEe.Lock()
	if sqlEe.db == nil || sqlEe.sqldb == nil {
		sqlEe.db, sqlEe.sqldb, err = openDB(sqlEe.dialect, sqlEe.Cfg().Opts)
	}
	sqlEe.Unlock()
	return
}

func (sqlEe *SQLEe) ExportEvent(_ *context.Context, req, _ any) error {
	sqlEe.reqs.get()
	sqlEe.RLock()
	defer func() {
		sqlEe.RUnlock()
		sqlEe.reqs.done()
	}()
	if sqlEe.db == nil {
		return utils.ErrDisconnected
	}
	sReq := req.(*sqlPosterRequest)
	return sqlEe.db.Table(sqlEe.tableName).Exec(sReq.Querry, sReq.Values...).Error
}

func (sqlEe *SQLEe) Close() (err error) {
	sqlEe.Lock()
	if sqlEe.sqldb != nil {
		err = sqlEe.sqldb.Close()
	}
	sqlEe.db = nil
	sqlEe.sqldb = nil
	sqlEe.Unlock()
	return
}

func (sqlEe *SQLEe) GetMetrics() *utils.SafeMapStorage { return sqlEe.dc }

func (sqlEe *SQLEe) ExtraData(ev *utils.CGREvent) any { return nil }

func (sqlEe *SQLEe) PrepareMap(mp *utils.CGREvent) (any, error) { return nil, nil }

func (sqlEe *SQLEe) PrepareOrderMap(mp *utils.OrderedNavigableMap) (any, error) {
	var vals []any
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
