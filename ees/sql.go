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
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewSQLEe(cfg *config.EventExporterCfg,
	em *utils.ExporterMetrics) (sqlEe *SQLEe, err error) {
	sqlEe = &SQLEe{
		cfg:  cfg,
		em:   em,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	err = sqlEe.initDialector()
	return
}

// SQLEe implements EventExporter interface for SQL
type SQLEe struct {
	cfg   *config.EventExporterCfg
	em    *utils.ExporterMetrics
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
	if sqlEe.Cfg().Opts.SQL.DBName != nil {
		dbname = *sqlEe.Cfg().Opts.SQL.DBName
	}
	ssl := utils.SQLDefaultSSLMode
	if sqlEe.Cfg().Opts.SQL.PgSSLMode != nil {
		ssl = *sqlEe.Cfg().Opts.SQL.PgSSLMode
	}
	// tableName is mandatory in opts
	if sqlEe.Cfg().Opts.SQL.TableName != nil {
		sqlEe.tableName = *sqlEe.Cfg().Opts.SQL.TableName
	} else {
		return utils.NewErrMandatoryIeMissing(utils.SQLTableNameOpt)
	}

	// var dialect gorm.Dialector
	switch u.Scheme {
	case utils.MySQL:
		sqlEe.dialect = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			u.User.Username(), password, u.Hostname(), u.Port(), dbname) + engine.AppendToMysqlDSNOpts(sqlEe.Cfg().Opts.SQL.MYSQLDSNParams))
	case utils.Postgres:
		sqlEe.dialect = postgres.Open(fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", u.Hostname(), u.Port(), dbname, u.User.Username(), password, ssl))
	default:
		return fmt.Errorf("db type <%s> not supported", u.Scheme)
	}
	return
}

func openDB(dialect gorm.Dialector, opts *config.SQLOpts) (db *gorm.DB, sqlDB *sql.DB, err error) {
	if db, err = gorm.Open(dialect, &gorm.Config{AllowGlobalUpdate: true}); err != nil {
		return
	}
	if sqlDB, err = db.DB(); err != nil {
		return
	}

	if opts.MaxIdleConns != nil {
		sqlDB.SetMaxIdleConns(*opts.MaxIdleConns)
	}
	if opts.MaxOpenConns != nil {
		sqlDB.SetMaxOpenConns(*opts.MaxOpenConns)
	}
	if opts.ConnMaxLifetime != nil {
		sqlDB.SetConnMaxLifetime(*opts.ConnMaxLifetime)
	}

	return
}

func (sqlEe *SQLEe) Cfg() *config.EventExporterCfg { return sqlEe.cfg }

func (sqlEe *SQLEe) Connect() (err error) {
	sqlEe.Lock()
	if sqlEe.db == nil || sqlEe.sqldb == nil {
		sqlEe.db, sqlEe.sqldb, err = openDB(sqlEe.dialect, sqlEe.Cfg().Opts.SQL)
	}
	sqlEe.Unlock()
	return
}

func (sqlEe *SQLEe) ExportEvent(req any, _ string) error {
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

func (sqlEe *SQLEe) GetMetrics() *utils.ExporterMetrics { return sqlEe.em }

// Create the sqlPosterRequest used to instert the map into the table
func (sqlEe *SQLEe) PrepareMap(cgrEv *utils.CGREvent) (any, error) {
	colNames := make([]string, 0, len(cgrEv.Event)) // slice with all column names to be insterted
	vals := make([]any, 0, len(cgrEv.Event))        // slice with all values to be insterted
	for colName, value := range cgrEv.Event {
		colNames = append(colNames, colName)
		vals = append(vals, value)
	}
	sqlValues := make([]string, len(vals)) // values to be inserted as "?" for the query
	for i := range vals {
		sqlValues[i] = "?"
	}
	sqlQuery := fmt.Sprintf("INSERT INTO %s (`%s`) VALUES (%s);",
		sqlEe.tableName,
		strings.Join(colNames, "`, `"), // back ticks added to include special characters
		strings.Join(sqlValues, ","),
	)
	return &sqlPosterRequest{
		Querry: sqlQuery,
		Values: vals,
	}, nil
}

func (sqlEe *SQLEe) PrepareOrderMap(mp *utils.OrderedNavigableMap) (any, error) {
	var vals []any
	var colNames []string
	var whereVars []string // key-value parts of WHERE clause used on UPDATE
	var whereVals []any    // will hold the values replacing "?" used on WHERE part of UPDATE query
	for el := mp.GetFirstElement(); el != nil; el = el.Next() {
		nmIt, _ := mp.Field(el.Value)
		pathWithoutIndex := strings.Join(el.Value[:len(el.Value)-1], utils.NestingSep) // remove the index path.index
		if pathWithoutIndex != utils.MetaRow {
			colNames = append(colNames, pathWithoutIndex)
		}
		vals = append(vals, nmIt.Data)
		if sqlEe.cfg.Opts.SQL.UpdateIndexedFields != nil {
			for _, updateFields := range *sqlEe.cfg.Opts.SQL.UpdateIndexedFields {
				if pathWithoutIndex == updateFields {
					whereVars = append(whereVars, fmt.Sprintf("%s = ?", updateFields))
					whereVals = append(whereVals, nmIt.Data)
				}
			}
		}
	}
	sqlValues := make([]string, len(vals)+len(whereVals))
	for i := range vals {
		sqlValues[i] = "?"
	}
	var sqlQuery string
	if sqlEe.cfg.Opts.SQL.UpdateIndexedFields != nil {
		if len(whereVars) == 0 {
			return nil, fmt.Errorf("%w: no usable sqlUpdateIndexedFields found <%v>", utils.ErrNotFound, *sqlEe.cfg.Opts.SQL.UpdateIndexedFields)
		}
		setClauses := []string{} // used in SET part of UPDATE query
		for _, col := range colNames {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		}
		sqlQuery = fmt.Sprintf("UPDATE %s SET %s WHERE %s;",
			sqlEe.tableName,
			strings.Join(setClauses, ", "),
			strings.Join(whereVars, " AND "))
		for _, val := range whereVals {
			vals = append(vals, val)
		}
	} else {
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
	}
	return &sqlPosterRequest{
		Querry: sqlQuery,
		Values: vals,
	}, nil
}
