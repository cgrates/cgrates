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
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	// libs for sql DBs
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	createdAt = "created_at"
	updatedAt = "updated_at"
	deletedAt = "deleted_at"
)

// NewSQLEventReader return a new sql event reader
func NewSQLEventReader(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {

	rdr := &SQLEventReader{
		cgrCfg:    cfg,
		cfgIdx:    cfgIdx,
		fltrS:     fltrS,
		rdrEvents: rdrEvents,
		rdrExit:   rdrExit,
		rdrErr:    rdrErr,
	}
	if concReq := rdr.Config().ConcurrentReqs; concReq != -1 {
		rdr.cap = make(chan struct{}, concReq)
		for i := 0; i < concReq; i++ {
			rdr.cap <- struct{}{}
		}
	}
	if err = rdr.setURL(rdr.Config().SourcePath, rdr.Config().ProcessedPath, rdr.Config().Opts); err != nil {
		return
	}
	er = rdr

	return
}

// SQLEventReader implements EventReader interface for sql
type SQLEventReader struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	connString string
	connType   string
	tableName  string

	expConnString string
	expConnType   string
	expTableName  string

	rdrEvents chan *erEvent // channel to dispatch the events created to
	rdrExit   chan struct{}
	rdrErr    chan error
	cap       chan struct{}
}

// Config returns the curent configuration
func (rdr *SQLEventReader) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *SQLEventReader) openDB(dialect gorm.Dialector) (err error) {
	var db *gorm.DB
	if db, err = gorm.Open(dialect, &gorm.Config{AllowGlobalUpdate: true}); err != nil {
		return
	}
	var sqlDB *sql.DB
	if sqlDB, err = db.DB(); err != nil {
		return
	}
	sqlDB.SetMaxOpenConns(10)
	if rdr.Config().RunDelay == time.Duration(0) { // 0 disables the automatic read, maybe done per API
		return
	}
	go rdr.readLoop(db, sqlDB) // read until the connection is closed
	return
}

// Serve will start the gorutines needed to watch the sql topic
func (rdr *SQLEventReader) Serve() (err error) {
	var dialect gorm.Dialector
	switch rdr.connType {
	case utils.MySQL:
		dialect = mysql.Open(rdr.connString)
	case utils.Postgres:
		dialect = postgres.Open(rdr.connString)
	default:
		return fmt.Errorf("db type <%s> not supported", rdr.connType)
	}
	err = rdr.openDB(dialect)
	return
}

func (rdr *SQLEventReader) readLoop(db *gorm.DB, sqlDB io.Closer) {
	defer sqlDB.Close()
	tm := time.NewTimer(0)
	for {
		rows, err := db.Table(rdr.tableName).Select(utils.Meta).Rows()
		if err != nil {
			rdr.rdrErr <- err
			return
		}
		colNames, err := rows.Columns()
		if err != nil {
			rdr.rdrErr <- err
			rows.Close()
			return
		}
		for rows.Next() {
			select {
			case <-rdr.rdrExit:
				utils.Logger.Info(
					fmt.Sprintf("<%s> stop monitoring sql DB <%s>",
						utils.ERs, rdr.Config().SourcePath))
				rows.Close()
				return
			default:
			}
			if err := rows.Err(); err != nil {
				rdr.rdrErr <- err
				rows.Close()
				return
			}
			if rdr.Config().ConcurrentReqs != -1 {
				<-rdr.cap // do not try to read if the limit is reached
			}
			columns := make([]interface{}, len(colNames))
			columnPointers := make([]interface{}, len(colNames))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}
			if err = rows.Scan(columnPointers...); err != nil {
				rdr.rdrErr <- err
				rows.Close()
				return
			}
			msg := make(map[string]interface{})
			fltr := make(map[string]string)
			for i, colName := range colNames {
				msg[colName] = columns[i]
				if colName != createdAt && colName != updatedAt && colName != deletedAt { // ignore the sql colums for filter only
					switch tm := columns[i].(type) { // also ignore the values that are zero for time
					case time.Time:
						if tm.IsZero() {
							continue
						}
					case *time.Time:
						if tm == nil || tm.IsZero() {
							continue
						}
					case nil:
						continue
					}
					fltr[colName] = utils.IfaceAsString(columns[i])
				}
			}
			if err = db.Table(rdr.tableName).Delete(nil, fltr).Error; err != nil { // to ensure we don't read it again
				utils.Logger.Warning(
					fmt.Sprintf("<%s> deleting message %s error: %s",
						utils.ERs, utils.ToJSON(msg), err.Error()))
				rdr.rdrErr <- err
				rows.Close()
				return
			}

			go func(msg map[string]interface{}) {
				if err := rdr.processMessage(msg); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing message %s error: %s",
							utils.ERs, utils.ToJSON(msg), err.Error()))
				}
				if rdr.Config().ProcessedPath != utils.EmptyString {
					if err = rdr.postCDR(columns); err != nil {
						utils.Logger.Warning(
							fmt.Sprintf("<%s> posting message %s error: %s",
								utils.ERs, utils.ToJSON(msg), err.Error()))
					}
				}
				if rdr.Config().ConcurrentReqs != -1 {
					rdr.cap <- struct{}{}
				}
			}(msg)
		}
		rows.Close()
		if rdr.Config().RunDelay < 0 {
			return
		}
		tm.Reset(rdr.Config().RunDelay)
		select {
		case <-rdr.rdrExit:
			tm.Stop()
			utils.Logger.Info(
				fmt.Sprintf("<%s> stop monitoring sql DB <%s>",
					utils.ERs, rdr.Config().SourcePath))
			return
		case <-tm.C:
		}
	}
}

func (rdr *SQLEventReader) processMessage(msg map[string]interface{}) (err error) {
	agReq := agents.NewAgentRequest(
		utils.MapStorage(msg), nil,
		nil, nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, nil, nil) // create an AgentRequest
	var pass bool
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	if err = agReq.SetFields(rdr.Config().Fields); err != nil {
		return
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	rdr.rdrEvents <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
	return
}

func (rdr *SQLEventReader) setURL(inURL, outURL string, opts map[string]interface{}) (err error) {
	inURL = strings.TrimPrefix(inURL, utils.Meta)
	var u *url.URL
	if u, err = url.Parse(inURL); err != nil {
		return
	}
	password, _ := u.User.Password()
	rdr.connType = u.Scheme

	dbname := utils.SQLDefaultDBName
	if vals, has := opts[utils.SQLDBNameOpt]; has {
		dbname = utils.IfaceAsString(vals)
	}
	ssl := utils.SQLDefaultSSLMode
	if vals, has := opts[utils.SSLModeCfg]; has {
		ssl = utils.IfaceAsString(vals)
	}

	rdr.tableName = utils.CDRsTBL
	if vals, has := opts[utils.SQLTableNameOpt]; has {
		rdr.tableName = utils.IfaceAsString(vals)
	}
	switch rdr.connType {
	case utils.MySQL:
		rdr.connString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			u.User.Username(), password, u.Hostname(), u.Port(), dbname)
	case utils.Postgres:
		rdr.connString = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", u.Hostname(), u.Port(), dbname, u.User.Username(), password, ssl)
	default:
		return fmt.Errorf("unknown db_type %s", rdr.connType)
	}

	// outURL
	processedOpt := getProcessOptions(opts)
	if len(processedOpt) == 0 &&
		len(outURL) == 0 {
		return
	}
	var outUser, outPassword, outDBname, outSSL, outHost, outPort string
	if len(outURL) == 0 {
		rdr.expConnType = rdr.connType
		outUser = u.User.Username()
		outPassword = password
		outHost = u.Hostname()
		outPort = u.Port()
	} else {
		outURL = strings.TrimPrefix(outURL, utils.Meta)
		var oURL *url.URL
		if oURL, err = url.Parse(outURL); err != nil {
			return
		}
		rdr.expConnType = oURL.Scheme
		outPassword, _ = oURL.User.Password()
		outUser = oURL.User.Username()
		outHost = oURL.Hostname()
		outPort = oURL.Port()
	}

	outDBname = utils.SQLDefaultDBName
	if vals, has := processedOpt[utils.SQLDBNameOpt]; has {
		outDBname = utils.IfaceAsString(vals)
	}
	outSSL = utils.SQLDefaultSSLMode
	if vals, has := processedOpt[utils.SSLModeCfg]; has {
		outSSL = utils.IfaceAsString(vals)
	}
	rdr.expTableName = utils.CDRsTBL
	if vals, has := processedOpt[utils.SQLTableNameOpt]; has {
		rdr.expTableName = utils.IfaceAsString(vals)
	}

	switch rdr.expConnType {
	case utils.MySQL:
		rdr.expConnString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			outUser, outPassword, outHost, outPort, outDBname)
	case utils.Postgres:
		rdr.expConnString = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
			outHost, outPort, outDBname, outUser, outPassword, outSSL)
	default:
		return fmt.Errorf("unknown db_type %s", rdr.expConnType)
	}
	return
}

func (rdr *SQLEventReader) postCDR(in []interface{}) (err error) {
	sqlValues := make([]string, len(in))
	for i := range in {
		sqlValues[i] = "?"
	}
	sqlStatement := fmt.Sprintf("INSERT INTO %s VALUES (%s); ", rdr.expTableName, strings.Join(sqlValues, ","))
	var dialect gorm.Dialector
	switch rdr.expConnType {
	case utils.MySQL:
		dialect = mysql.Open(rdr.expConnString)
	case utils.Postgres:
		dialect = postgres.Open(rdr.expConnString)
	default:
		return fmt.Errorf("db type <%s> not supported", rdr.expConnType)
	}
	var db *gorm.DB
	if db, err = gorm.Open(dialect, &gorm.Config{AllowGlobalUpdate: true}); err != nil {
		return
	}
	var sqlDB *sql.DB
	if sqlDB, err = db.DB(); err != nil {
		return
	}
	defer sqlDB.Close()
	// sqlDB.SetMaxOpenConns(10)
	if err = sqlDB.Ping(); err != nil {
		return
	}
	tx := db.Begin()
	if err = tx.Exec(sqlStatement, in...).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "1062") || strings.Contains(err.Error(), "duplicate key") { // returns 1062/pq when key is duplicated
			return utils.ErrExists
		}
		return
	}
	tx.Commit()
	return
}
