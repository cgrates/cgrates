// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/url"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
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

	connString string
	connType   string
	tableName  string

	rdrEvents     chan *erEvent
	partialEvents chan *erEvent
	rdrExit       chan struct{}
	rdrErr        chan error
	cap           chan struct{}
}

func (rdr *CgrCDR) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *CgrCDR) Serve() error {
	db, sqlDB, err := rdr.openDB()
	if err != nil {
		return err
	}
	if rdr.Config().RunDelay == 0 {
		return sqlDB.Close()
	}
	go rdr.readLoop(db, sqlDB)
	return nil
}

func (rdr *CgrCDR) openDB() (*gorm.DB, *sql.DB, error) {
	var dialect gorm.Dialector
	switch rdr.connType {
	case utils.MySQL:
		dialect = mysql.Open(rdr.connString)
	case utils.Postgres:
		dialect = postgres.Open(rdr.connString)
	default:
		return nil, nil, fmt.Errorf("db type <%s> not supported", rdr.connType)
	}
	db, err := gorm.Open(dialect, &gorm.Config{AllowGlobalUpdate: true})
	if err != nil {
		return nil, nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}
	return db, sqlDB, nil
}

func (rdr *CgrCDR) filtersForQuery(filters []string) (dbFilters []string, dbFilterArgs []any,
	lazyFilters []string, err error) {
	for _, filterID := range filters {
		filterObj, err := rdr.dm.GetFilter(context.TODO(), rdr.cgrCfg.GeneralCfg().DefaultTenant,
			filterID, true, false, utils.NonTransactional)
		if err != nil {
			return nil, nil, nil, err
		}
		if err := engine.CheckFilter(filterObj); err != nil {
			return nil, nil, nil, err
		}
		var lazyFilter bool
		for _, rule := range filterObj.Rules {
			if strings.HasPrefix(rule.Element, utils.MetaDynReq+utils.NestingSep) {
				conditions, args := rule.FilterToSQLQuery()
				dbFilters = append(dbFilters, strings.Join(conditions, " OR "))
				dbFilterArgs = append(dbFilterArgs, args...)
				continue
			}
			if !lazyFilter {
				lazyFilters = append(lazyFilters, filterObj.ID)
				lazyFilter = true
			}
		}
	}
	return dbFilters, dbFilterArgs, lazyFilters, nil
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
	dbFilters, dbFilterArgs, lazyFilters, err := rdr.filtersForQuery(rdr.Config().Filters)
	if err != nil {
		rdr.rdrErr <- err
		return
	}
	selectWhereQuery := strings.Join(dbFilters, " AND ")
	tm := time.NewTimer(0)
	for {
		var cdrs []*utils.CDRSQLTable
		tx := db.Table(rdr.tableName).Model(&utils.CDRSQLTable{})
		if selectWhereQuery != "" {
			tx = tx.Where(selectWhereQuery, dbFilterArgs...)
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
				if err := rdr.processMessage(cdrSql, lazyFilters); err != nil {
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

func (rdr *CgrCDR) run(filters []string) error {
	if rdr.connType != utils.MySQL {
		return errors.New("manual cgrcdr processing supports only mysql")
	}
	db, sqlDB, err := rdr.openDB()
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()
	dbFilters, dbFilterArgs, lazyFilters, err := rdr.filtersForQuery(filters)
	if err != nil {
		return err
	}
	var cdrs []*utils.CDRSQLTable
	tx := db.Table(rdr.tableName).Model(&utils.CDRSQLTable{})
	if selectWhereQuery := strings.Join(dbFilters, " AND "); selectWhereQuery != "" {
		tx = tx.Where(selectWhereQuery, dbFilterArgs...)
	}
	if rdr.Config().Opts.SQLBatchSize != nil && *rdr.Config().Opts.SQLBatchSize > 0 {
		tx = tx.Limit(*rdr.Config().Opts.SQLBatchSize)
	}
	if err := tx.Find(&cdrs).Error; err != nil {
		return err
	}
	for _, cdrSQL := range cdrs {
		if rdr.Config().ProcessedPath == utils.MetaDelete {
			if err := db.Table(rdr.tableName).Delete(&utils.CDRSQLTable{}, cdrSQL.ID).Error; err != nil {
				return err
			}
		}
		if err := rdr.processMessage(cdrSQL, lazyFilters); err != nil {
			return err
		}
	}
	return nil
}

func (rdr *CgrCDR) processMessage(cdrSql *utils.CDRSQLTable, lazyFilters []string) error {
	cdr := &utils.CDR{
		Tenant:    cdrSql.Tenant,
		Opts:      cdrSql.Opts,
		Event:     cdrSql.Event,
		CreatedAt: cdrSql.CreatedAt,
		UpdatedAt: cdrSql.UpdatedAt,
		DeletedAt: cdrSql.DeletedAt,
	}
	cgrEv := cdr.CGREvent()
	if pass, err := rdr.fltrS.Pass(context.TODO(), cgrEv.Tenant, lazyFilters,
		cgrEv.AsDataProvider()); err != nil || !pass {
		return err
	}
	if len(rdr.Config().Fields) != 0 {
		reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{
			utils.MetaReaderID: utils.NewLeafNode(rdr.Config().ID),
		}}
		// Copy the options because their original values may be exported later.
		opts := maps.Clone(cgrEv.APIOpts)
		agReq := agents.NewAgentRequest(
			utils.MapStorage(cgrEv.Event), reqVars,
			nil, nil, utils.MapStorage(opts), nil,
			cgrEv.Tenant,
			utils.FirstNonEmpty(rdr.Config().Timezone,
				rdr.cgrCfg.GeneralCfg().DefaultTimezone),
			rdr.cgrCfg, nil, rdr.fltrS, nil)
		// Copy the stored Event first so values not listed in Fields stay unchanged.
		for k, v := range cgrEv.Event {
			if err := agReq.CGRRequest.Set(utils.NewFullPath(k), v); err != nil {
				return err
			}
		}
		if err := agReq.SetFields(rdr.Config().Fields); err != nil {
			return err
		}
		if ev := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant,
			utils.NestingSep, agReq.Opts); ev != nil {
			cgrEv.Event = ev.Event
		} else {
			cgrEv.Event = make(map[string]any)
		}
		cgrEv.APIOpts = agReq.Opts
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
