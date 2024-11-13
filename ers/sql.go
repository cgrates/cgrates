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
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}, dm *engine.DataManager) (er EventReader, err error) {
	rdr := &SQLEventReader{
		cgrCfg:        cfg,
		dm:            dm,
		cfgIdx:        cfgIdx,
		fltrS:         fltrS,
		rdrEvents:     rdrEvents,
		partialEvents: partialEvents,
		rdrExit:       rdrExit,
		rdrErr:        rdrErr,
	}
	if concReq := rdr.Config().ConcurrentReqs; concReq != -1 {
		rdr.cap = make(chan struct{}, concReq)
	}
	if err = rdr.setURL(rdr.Config().SourcePath, rdr.Config().Opts); err != nil {
		return nil, err
	}
	return rdr, nil
}

// SQLEventReader implements EventReader interface for sql
type SQLEventReader struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	dm     *engine.DataManager
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
	if rdr.Config().RunDelay <= 0 { // 0 disables the automatic read, maybe done per API
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

// Creates mysql conditions used in WHERE statement out of filters
func valueQry(ruleType, elem, field string, values []string, not bool) (conditions []string) {
	// here are for the filters that their values are empty: *exists, *notexists, *empty, *notempty..
	if len(values) == 0 {
		switch ruleType {
		case utils.MetaExists, utils.MetaNotExists:
			if not {
				if elem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf(" %s IS NOT NULL", field))
					return
				}
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.%s') IS NOT NULL", elem, field))
				return
			}
			if elem == utils.EmptyString {
				conditions = append(conditions, fmt.Sprintf(" %s IS NULL", field))
				return
			}
			conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.%s') IS NULL", elem, field))
		case utils.MetaEmpty, utils.MetaNotEmpty:
			if not {
				if elem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf(" %s != ''", field))
					return
				}
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.%s') != ''", elem, field))
				return
			}
			if elem == utils.EmptyString {
				conditions = append(conditions, fmt.Sprintf(" %s == ''", field))
				return
			}
			conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.%s') == ''", elem, field))
		}
		return
	}
	// here are for the filters that can have more than one value: *string, *prefix, *suffix ..
	for _, value := range values {
		switch value { // in case we have boolean values, it should be queried over 1 or 0
		case "true":
			value = "1"
		case "false":
			value = "0"
		}
		var singleCond string
		switch ruleType {
		case utils.MetaString, utils.MetaNotString, utils.MetaEqual, utils.MetaNotEqual:
			if not {
				if elem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf(" %s != '%s'", field, value))
					continue
				}
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.%s') != '%s'",
					elem, field, value))
				continue
			}
			if elem == utils.EmptyString {
				singleCond = fmt.Sprintf(" %s = '%s'", field, value)
			} else {
				singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.%s') = '%s'", elem, field, value)
			}
		case utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
			if ruleType == utils.MetaGreaterOrEqual {
				if elem == utils.EmptyString {
					singleCond = fmt.Sprintf(" %s >= %s", field, value)
				} else {
					singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.%s') >= %s", elem, field, value)
				}
			} else if ruleType == utils.MetaGreaterThan {
				if elem == utils.EmptyString {
					singleCond = fmt.Sprintf(" %s > %s", field, value)
				} else {
					singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.%s') > %s", elem, field, value)
				}
			} else if ruleType == utils.MetaLessOrEqual {
				if elem == utils.EmptyString {
					singleCond = fmt.Sprintf(" %s <= %s", field, value)
				} else {
					singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.%s') <= %s", elem, field, value)
				}
			} else if ruleType == utils.MetaLessThan {
				if elem == utils.EmptyString {
					singleCond = fmt.Sprintf(" %s < %s", field, value)
				} else {
					singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.%s') < %s", elem, field, value)
				}
			}
		case utils.MetaPrefix, utils.MetaNotPrefix:
			if not {
				if elem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf(" %s NOT LIKE '%s%%'", field, value))
					continue
				}
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.%s') NOT LIKE '%s%%'", elem, field, value))
				continue
			}
			if elem == utils.EmptyString {
				singleCond = fmt.Sprintf(" %s LIKE '%s%%'", field, value)
			} else {
				singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.%s') LIKE '%s%%'", elem, field, value)
			}
		case utils.MetaSuffix, utils.MetaNotSuffix:
			if not {
				if elem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf(" %s NOT LIKE '%%%s'", field, value))
					continue
				}
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.%s') NOT LIKE '%%%s'", elem, field, value))
				continue
			}
			if elem == utils.EmptyString {
				singleCond = fmt.Sprintf(" %s LIKE '%%%s'", field, value)
			} else {
				singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.%s') LIKE '%%%s'", elem, field, value)
			}
		case utils.MetaRegex, utils.MetaNotRegex:
			if not {
				if elem == utils.EmptyString {
					conditions = append(conditions, fmt.Sprintf(" %s NOT REGEXP '%s'", field, value))
					continue
				}
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.%s') NOT REGEXP '%s'", elem, field, value))
				continue
			}
			if elem == utils.EmptyString {
				singleCond = fmt.Sprintf(" %s REGEXP '%s'", field, value)
			} else {
				singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.%s') REGEXP '%s'", elem, field, value)
			}
		}
		conditions = append(conditions, singleCond)
	}
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
	var filters []*engine.Filter
	var whereQueries []string
	var renewedFltrs []string
	for _, fltr := range rdr.Config().Filters {
		if result, err := rdr.dm.GetFilter(config.CgrConfig().GeneralCfg().DefaultTenant, fltr, true, false, utils.NonTransactional); err != nil {
			rdr.rdrErr <- err
			return
		} else {
			filters = append(filters, result)
			if !strings.Contains(fltr, utils.DynamicDataPrefix+utils.MetaReq) {
				renewedFltrs = append(renewedFltrs, fltr)
			}
		}
	}
	rdr.Config().Filters = renewedFltrs // remove filters containing *req
	for _, filter := range filters {
		for _, rule := range filter.Rules {
			var elem, field string
			switch {
			case strings.HasPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep):
				field = strings.TrimPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep)
				parts := strings.SplitN(field, ".", 2)
				if len(parts) == 2 { // Split in 2 pieces if it contains any more dots in the field
					// First part (before the first dot)
					elem = parts[0]
					// Second part (everything after the first dot)
					field = parts[1]
				}
			default:
				continue
			}
			conditions := valueQry(rule.Type, elem, field, rule.Values, strings.HasPrefix(rule.Type, utils.MetaNot))
			whereQueries = append(whereQueries, strings.Join(conditions, " OR "))
		}
	}
	for {
		tx := db.Table(rdr.tableName).Select(utils.Meta)
		for _, whereQ := range whereQueries {
			tx = tx.Where(whereQ)
		}
		rows, err := tx.Rows()
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
				rdr.cap <- struct{}{}
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
			if rdr.Config().ProcessedPath == utils.MetaDelete {
				if err = db.Table(rdr.tableName).Delete(nil, fltr).Error; err != nil { // to ensure we don't read it again
					utils.Logger.Warning(
						fmt.Sprintf("<%s> deleting message %s error: %s",
							utils.ERs, utils.ToJSON(msg), err.Error()))
					rdr.rdrErr <- err
					rows.Close()
					return
				}
			}

			go func(msg map[string]any) {
				if err := rdr.processMessage(msg); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing message %s error: %s",
							utils.ERs, utils.ToJSON(msg), err.Error()))
				}
				if rdr.Config().ConcurrentReqs != -1 {
					<-rdr.cap
				}
			}(msg)
		}
		rows.Close()
		tm.Reset(rdr.Config().RunDelay) // reset the timer to RunDelay
		select {                        // wait for timer or rdrExit
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
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
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
	ssl := utils.SQLDefaultSSLMode
	if sqlOpts := opts.SQL; sqlOpts != nil {
		if sqlOpts.DBName != nil {
			dbname = *sqlOpts.DBName
		}

		if sqlOpts.PgSSLMode != nil {
			ssl = *sqlOpts.PgSSLMode
		}

		rdr.tableName = utils.CDRsTBL
		if sqlOpts.TableName != nil {
			rdr.tableName = *sqlOpts.TableName
		}
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
