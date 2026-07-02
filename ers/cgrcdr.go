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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewCgrCdr returns a new *cgrcdr event reader
func NewCgrCdr(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}, dm *engine.DataManager) (EventReader, error) {

	rdr := &CgrCDR{
		cgrCfg:        cfg,
		cfgIdx:        cfgIdx,
		fltrS:         fltrS,
		dm:            dm,
		rdrEvents:     rdrEvents,
		partialEvents: partialEvents,
		rdrExit:       rdrExit,
		rdrErr:        rdrErr,
	}
	if concReq := rdr.Config().ConcurrentReqs; concReq != -1 {
		rdr.cap = make(chan struct{}, concReq)
	}
	if err := rdr.setURL(rdr.Config().SourcePath, rdr.Config().Opts); err != nil {
		return nil, err
	}
	return rdr, nil
}

// Cgrcdr implements EventReader for the *cgrcdr type.
type CgrCDR struct {
	cgrCfg *config.CGRConfig
	cfgIdx int
	fltrS  *engine.FilterS
	dm     *engine.DataManager

	connString  string
	connType    string
	tableName   string
	dbFilters   []string
	lazyFilters []string

	rdrEvents     chan *erEvent
	partialEvents chan *erEvent
	rdrExit       chan struct{}
	rdrErr        chan error
	cap           chan struct{}
}

func (rdr *CgrCDR) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *CgrCDR) Serve() (err error) {
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
	if rdr.Config().RunDelay == time.Duration(0) {
		return
	}
	go rdr.readLoop(db, sqlDB)
	return
}

func (rdr *CgrCDR) readLoop(db *gorm.DB, sqlDB io.Closer) {
	defer sqlDB.Close()
	if rdr.Config().StartDelay > 0 {
		select {
		case <-time.After(rdr.Config().StartDelay):
		case <-rdr.rdrExit:
			utils.Logger.Info(
				fmt.Sprintf("<%s> stop monitoring cgrcdr table <%s>",
					utils.ERs, rdr.Config().SourcePath))
			return
		}
	}
	var filtersObjList []*engine.Filter
	for _, fltrID := range rdr.Config().Filters {
		f, err := rdr.dm.GetFilter(context.TODO(), rdr.cgrCfg.GeneralCfg().DefaultTenant,
			fltrID, true, false, utils.NonTransactional)
		if err != nil {
			rdr.rdrErr <- err
			return
		}
		filtersObjList = append(filtersObjList, f)
	}
	for _, filterObj := range filtersObjList {
		if err := engine.CheckFilter(filterObj); err != nil {
			rdr.rdrErr <- err
			return
		}
		var lazyFltrPopulated bool
		for _, rule := range filterObj.Rules {
			if strings.HasPrefix(rule.Element, utils.MetaDynReq+utils.NestingSep) {
				rdr.dbFilters = append(rdr.dbFilters, strings.Join(rule.FilterToSQLQuery(), " OR "))
				continue
			}

			if !lazyFltrPopulated {
				rdr.lazyFilters = append(rdr.lazyFilters, filterObj.ID)
				lazyFltrPopulated = true
			}
		}
	}
	selectWhereQuery := strings.Join(rdr.dbFilters, " AND ")
	tm := time.NewTimer(0)
	for {
		var cdrs []*utils.CDRSQLTable
		tx := db.Table(rdr.tableName).Model(&utils.CDRSQLTable{})
		if selectWhereQuery != "" {
			tx = tx.Where(selectWhereQuery)
		}
		if rdr.Config().Opts.SQLBatchSize != nil && *rdr.Config().Opts.SQLBatchSize > 0 {
			tx = tx.Limit(*rdr.Config().Opts.SQLBatchSize)
		}
		if err := tx.Find(&cdrs).Error; err != nil {
			rdr.rdrErr <- err
			return
		}
		for _, cdrSql := range cdrs {
			select {
			case <-rdr.rdrExit:
				utils.Logger.Info(
					fmt.Sprintf("<%s> stop monitoring cgrcdr table <%s>",
						utils.ERs, rdr.Config().SourcePath))
				return
			default:
			}
			if rdr.Config().ConcurrentReqs != -1 {
				rdr.cap <- struct{}{}
			}
			if rdr.Config().ProcessedPath == utils.MetaDelete {
				if err := db.Table(rdr.tableName).Delete(&utils.CDRSQLTable{}, cdrSql.ID).Error; err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> deleting CDR id <%d> error: %s",
							utils.ERs, cdrSql.ID, err.Error()))
					rdr.rdrErr <- err
					return
				}
			}
			go func(cdrSql *utils.CDRSQLTable) {
				if err := rdr.processMessage(cdrSql); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing CDR id <%d> error: %s",
							utils.ERs, cdrSql.ID, err.Error()))
				}
				if rdr.Config().ConcurrentReqs != -1 {
					<-rdr.cap
				}
			}(cdrSql)
		}
		tm.Reset(rdr.Config().RunDelay)
		select {
		case <-rdr.rdrExit:
			tm.Stop()
			utils.Logger.Info(
				fmt.Sprintf("<%s> stop monitoring cgr CDR table <%s>",
					utils.ERs, rdr.Config().SourcePath))
			return
		case <-tm.C:
		}
	}
}

func (rdr *CgrCDR) processMessage(cdrSql *utils.CDRSQLTable) error {
	cdr := &utils.CDR{
		Tenant:    cdrSql.Tenant,
		Opts:      cdrSql.Opts,
		Event:     cdrSql.Event,
		CreatedAt: cdrSql.CreatedAt,
		UpdatedAt: cdrSql.UpdatedAt,
		DeletedAt: cdrSql.DeletedAt,
	}
	cgrEv := cdr.CGREvent()
	if pass, err := rdr.fltrS.Pass(context.TODO(), cgrEv.Tenant, rdr.lazyFilters,
		cgrEv.AsDataProvider()); err != nil || !pass {
		return err
	}
	rdrEv := rdr.rdrEvents
	if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
		rdrEv = rdr.partialEvents
	}
	var rawEvent map[string]any
	if len(rdr.Config().EEsSuccessIDs) != 0 || len(rdr.Config().EEsFailedIDs) != 0 {
		rawEvent = map[string]any{
			utils.ID:           cdrSql.ID,
			utils.Tenant:       cdrSql.Tenant,
			utils.OptsCfg:      cdrSql.Opts,
			utils.EventLowCase: cdrSql.Event,
		}
	}
	rdrEv <- &erEvent{
		cgrEvent: cgrEv,
		rawEvent: rawEvent,
		rdrCfg:   rdr.Config(),
	}
	return nil
}

func (rdr *CgrCDR) setURL(inURL string, opts *config.EventReaderOpts) error {
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
		rdr.connString = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
			u.Hostname(), u.Port(), dbname, u.User.Username(), password, ssl)
	default:
		return fmt.Errorf("unknown dbType %s", rdr.connType)
	}
	return nil
}
