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

	connString  string
	connType    string
	tableName   string
	dbFilters   []string // filters converted to SQL WHERE conditions from reader config filters
	lazyFilters []string // filters used when processing reader events

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
	var filtersObjList []*engine.Filter // List of filter objects from rdr.Config().Filters, received from DB
	for _, fltr := range rdr.Config().Filters {
		if resultFilter, err := rdr.dm.GetFilter(config.CgrConfig().GeneralCfg().DefaultTenant, fltr, true, false, utils.NonTransactional); err != nil {
			rdr.rdrErr <- err
			return
		} else {
			filtersObjList = append(filtersObjList, resultFilter)
		}
	}
	for _, filterObj := range filtersObjList { // seperate filters used for WHERE clause from other filters, and build query conditions out of them
		var lazyFltrPopulated bool // Track if a lazyFilter is already populated by the previous filterObj.Rules, so we dont store the same lazy filter more than once
		for _, rule := range filterObj.Rules {
			var firstItem string   // Excluding ~*req, hold the first item of an element, left empty if no more than 1 item in element. e.g. "cost_details" out of ~*req.cost_details.Charges[0].RatingID or "" out of ~*req.answer_time
			var restOfItems string // Excluding ~*req, hold the rest of the items past the first one. If only 1 item in all element, holds that item. e.g. "Charges[0].RatingID" out of ~*req.cost_details.Charges[0].RatingID or "answer_time" out of ~*req.answer_time
			switch {
			case strings.HasPrefix(rule.Element, utils.MetaDynReq+utils.NestingSep): // convert filter to WHERE condition only on filters with ~*req.
				elementItems := rule.ElementItems()[1:] // exclude first item: ~*req
				if len(elementItems) > 1 {
					firstItem = elementItems[0]
					restOfItems = strings.Join(elementItems[1:], utils.NestingSep)
				} else {
					restOfItems = elementItems[0]
				}
			default: // If not used in the WHERE condition, put the filter in rdr.lazyFilters
				if !lazyFltrPopulated {
					rdr.lazyFilters = append(rdr.lazyFilters, filterObj.ID)
					lazyFltrPopulated = true
				}
				continue
			}
			conditions := utils.FilterToSQLQuery(rule.Type, firstItem, restOfItems, rule.Values, strings.HasPrefix(rule.Type, utils.MetaNot))
			rdr.dbFilters = append(rdr.dbFilters, strings.Join(conditions, " OR "))
		}
	}
	tm := time.NewTimer(0) // Timer matching rdr.Config().RunDelay, will delay the for loop until timer expires. It doesnt wait for the loop to finish an iteration to start.
	for {
		tx := db.Table(rdr.tableName).Select(utils.Meta) // Select everything from the table
		for _, whereQ := range rdr.dbFilters {
			tx = tx.Where(whereQ) // apply WHERE conditions to the select if any
		}
		rows, err := tx.Rows() // get all rows selected
		if err != nil {
			rdr.rdrErr <- err
			return
		}
		colNames, err := rows.Columns() // get column names from rows selected
		if err != nil {
			rdr.rdrErr <- err
			rows.Close()
			return
		}
		for rows.Next() { // iterate on each row
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
			columns := make([]any, len(colNames))        // create a list of interfaces correlating to the columns selected
			columnPointers := make([]any, len(colNames)) // create a list of interfaces pointing to columns to be gotten from rows.Scan
			for i := range columns {
				columnPointers[i] = &columns[i]
			}
			if err = rows.Scan(columnPointers...); err != nil { // copy row values to their respective column
				rdr.rdrErr <- err
				rows.Close()
				return
			}

			ev := make(map[string]any)         // event to be processed
			for i, colName := range colNames { // populate ev from columns
				ev[colName] = columns[i]
			}
			if rdr.Config().ProcessedPath == utils.MetaDelete {
				sqlWhereVars := make(map[string]any) // map used for conditioning the DELETE query
				if rdr.Config().Opts.SQL.DeleteIndexedFields != nil {
					for _, fieldName := range *rdr.Config().Opts.SQL.DeleteIndexedFields {
						if _, has := ev[fieldName]; has && fieldName != createdAt && fieldName != updatedAt && fieldName != deletedAt { // ignore the sql colums for filter only
							addValidFieldToSQLWHEREVars(sqlWhereVars, fieldName, ev[fieldName])
						}
					}
				}
				if len(sqlWhereVars) == 0 {
					for i, colName := range colNames {
						if colName != createdAt && colName != updatedAt && colName != deletedAt { // ignore the sql colums for filter only
							addValidFieldToSQLWHEREVars(sqlWhereVars, colName, columns[i])
						}
					}
				}
				if err = tx.Delete(nil, sqlWhereVars).Error; err != nil { // to ensure we don't read it again
					utils.Logger.Warning(
						fmt.Sprintf("<%s> deleting message %s error: %s",
							utils.ERs, utils.ToJSON(ev), err.Error()))
					rdr.rdrErr <- err
					rows.Close()
					return
				}
			}

			go func(ev map[string]any) {
				if err := rdr.processMessage(ev); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing message %s error: %s",
							utils.ERs, utils.ToJSON(ev), err.Error()))
				}
				if rdr.Config().ConcurrentReqs != -1 {
					<-rdr.cap
				}
			}(ev)
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

// Helper function to add valid time and non-time values to the sqlWhereVars map
func addValidFieldToSQLWHEREVars(sqlWhereVars map[string]any, fieldName string, value any) {
	switch dateTimeCol := value.(type) {
	case time.Time:
		if dateTimeCol.IsZero() {
			return
		}
		sqlWhereVars[fieldName] = value
	case *time.Time:
		if dateTimeCol == nil || dateTimeCol.IsZero() {
			return
		}
		sqlWhereVars[fieldName] = value
	case nil:
		return
	default:
		sqlWhereVars[fieldName] = utils.IfaceAsString(value)
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
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.lazyFilters,
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
