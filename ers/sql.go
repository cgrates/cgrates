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

package ers

import (
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
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
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (EventReader, error) {

	rdr := &SQLEventReader{
		cgrCfg:        cfg,
		cfgIdx:        cfgIdx,
		fltrS:         fltrS,
		rdrEvents:     rdrEvents,
		partialEvents: partialEvents,
		rdrExit:       rdrExit,
		rdrErr:        rdrErr,
	}
	if concReq := rdr.Config().ConcurrentReqs; concReq != -1 {
		rdr.cap = make(chan struct{}, concReq)
		for i := 0; i < concReq; i++ {
			rdr.cap <- struct{}{}
		}
	}
	if err := rdr.setURL(rdr.Config().SourcePath, rdr.Config().Opts); err != nil {
		return nil, err
	}
	return rdr, nil
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

	rdrEvents     chan *erEvent // channel to dispatch the events created to
	partialEvents chan *erEvent // channel to dispatch the partial events created to
	rdrExit       chan struct{}
	rdrErr        chan error
	cap           chan struct{}
}

// Config returns the curent configuration
func (rdr *SQLEventReader) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
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
	var db *gorm.DB
	if db, err = gorm.Open(dialect, &gorm.Config{AllowGlobalUpdate: true}); err != nil {
		return
	}
	var sqlDB *sql.DB
	if sqlDB, err = db.DB(); err != nil {
		return
	}
	sqlDB.SetMaxOpenConns(10)
	if err = sqlDB.Ping(); err != nil {
		return
	}
	if rdr.Config().RunDelay == time.Duration(0) { // 0 disables the automatic read, maybe done per API
		return
	}
	go rdr.readLoop(db, sqlDB) // read until the connection is closed
	return
}

func (rdr *SQLEventReader) readLoop(db *gorm.DB, sqlDB io.Closer) {
	defer sqlDB.Close()
	if rdr.Config().StartDelay > 0 {
		select {
		case <-time.After(rdr.Config().StartDelay):
		case <-rdr.rdrExit:
			utils.Logger.Info(
				fmt.Sprintf("<%s> stop monitoring sql DB <%s>",
					utils.ERs, rdr.Config().SourcePath))
			return
		}
	}
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
			columns := make([]any, len(colNames))
			columnPointers := make([]any, len(colNames))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}
			if err = rows.Scan(columnPointers...); err != nil {
				rdr.rdrErr <- err
				rows.Close()
				return
			}
			msg := make(map[string]any)
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

			go func(msg map[string]any) {
				if err := rdr.processMessage(msg); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing message %s error: %s",
							utils.ERs, utils.ToJSON(msg), err.Error()))
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

func (rdr *SQLEventReader) processMessage(msg map[string]any) (err error) {
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaReaderID: utils.NewLeafNode(rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx].ID)}}
	agReq := agents.NewAgentRequest(
		utils.MapStorage(msg), reqVars,
		nil, nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, nil) // create an AgentRequest
	var pass bool
	if pass, err = rdr.fltrS.Pass(context.TODO(), agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	if err = agReq.SetFields(rdr.Config().Fields); err != nil {
		return
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	rdrEv := rdr.rdrEvents
	if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
		rdrEv = rdr.partialEvents
	}
	rdrEv <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
	return
}

func (rdr *SQLEventReader) setURL(inURL string, opts *config.EventReaderOpts) error {
	inURL = strings.TrimPrefix(inURL, utils.Meta)
	u, err := url.Parse(inURL)
	if err != nil {
		return err
	}
	password, _ := u.User.Password()
	rdr.connType = u.Scheme

	dbname := utils.SQLDefaultDBName
	if opts.SQLDBName != nil {
		dbname = *opts.SQLDBName
	}
	ssl := utils.SQLDefaultPgSSLMode
	if opts.PgSSLMode != nil {
		ssl = *opts.PgSSLMode
	}

	rdr.tableName = utils.CDRsTBL
	if opts.SQLTableName != nil {
		rdr.tableName = *opts.SQLTableName
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
	return nil
}
