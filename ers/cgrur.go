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

// NewCgrUr returns a new *cgrur event reader
func NewCgrUr(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}, dm *engine.DataManager) (EventReader, error) {

	rdr := &CgrUR{
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

// Cgrur implements EventReader for the *cgrur type.
type CgrUR struct {
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

func (rdr *CgrUR) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *CgrUR) Serve() error {
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

func (rdr *CgrUR) openDB() (*gorm.DB, *sql.DB, error) {
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

func (rdr *CgrUR) readLoop(db *gorm.DB, sqlDB io.Closer) {
	defer sqlDB.Close()
	if rdr.Config().StartDelay > 0 {
		select {
		case <-time.After(rdr.Config().StartDelay):
		case <-rdr.rdrExit:
			utils.Logger.Info(
				fmt.Sprintf("<%s> stop monitoring cgrur table <%s>",
					utils.ERs, rdr.Config().SourcePath))
			return
		}
	}
	dbFilters, dbFilterArgs, lazyFilters, err := engine.FiltersForQuery(rdr.Config().Filters, rdr.dm, rdr.cgrCfg.GeneralCfg().DefaultTenant)
	if err != nil {
		rdr.rdrErr <- err
		return
	}
	selectWhereQuery := strings.Join(dbFilters, " AND ")
	tm := time.NewTimer(0)
	for {
		var urs []*utils.URSQLTable
		tx := db.Table(rdr.tableName).Model(&utils.URSQLTable{})
		if selectWhereQuery != "" {
			tx = tx.Where(selectWhereQuery, dbFilterArgs...)
		}
		if rdr.Config().Opts.SQLBatchSize != nil && *rdr.Config().Opts.SQLBatchSize > 0 {
			tx = tx.Limit(*rdr.Config().Opts.SQLBatchSize)
		}
		if err := tx.Find(&urs).Error; err != nil {
			rdr.rdrErr <- err
			return
		}
		for _, urSql := range urs {
			select {
			case <-rdr.rdrExit:
				utils.Logger.Info(
					fmt.Sprintf("<%s> stop monitoring cgrur table <%s>",
						utils.ERs, rdr.Config().SourcePath))
				return
			default:
			}
			if rdr.Config().ConcurrentReqs != -1 {
				rdr.cap <- struct{}{}
			}
			if rdr.Config().ProcessedPath == utils.MetaDelete {
				if err := db.Table(rdr.tableName).Delete(&utils.URSQLTable{}, urSql.ID).Error; err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> deleting UR id <%d> error: %s",
							utils.ERs, urSql.ID, err.Error()))
					rdr.rdrErr <- err
					return
				}
			}
			go func(urSql *utils.URSQLTable) {
				if rdrEv, err := rdr.processMessage(urSql, lazyFilters); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing UR id <%d> error: %s",
							utils.ERs, urSql.ID, err.Error()))
				} else { // send event to ERs ListenAndServe
					rdrEvChan := rdr.rdrEvents
					if _, isPartial := rdrEv.cgrEvent.APIOpts[utils.PartialOpt]; isPartial {
						rdrEvChan = rdr.partialEvents
					}
					rdrEvChan <- rdrEv
				}
				if rdr.Config().ConcurrentReqs != -1 {
					<-rdr.cap
				}
			}(urSql)
		}
		tm.Reset(rdr.Config().RunDelay)
		select {
		case <-rdr.rdrExit:
			tm.Stop()
			utils.Logger.Info(
				fmt.Sprintf("<%s> stop monitoring cgr UR table <%s>",
					utils.ERs, rdr.Config().SourcePath))
			return
		case <-tm.C:
		}
	}
}

func (rdr *CgrUR) run(filters []string) error {
	if rdr.connType != utils.MySQL {
		return errors.New("manual cgrur processing supports only mysql")
	}
	db, sqlDB, err := rdr.openDB()
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()
	dbFilters, dbFilterArgs, lazyFilters, err := engine.FiltersForQuery(filters, rdr.dm, rdr.cgrCfg.GeneralCfg().DefaultTenant)
	if err != nil {
		return err
	}
	var urs []*utils.URSQLTable
	tx := db.Table(rdr.tableName).Model(&utils.URSQLTable{})
	if selectWhereQuery := strings.Join(dbFilters, " AND "); selectWhereQuery != "" {
		tx = tx.Where(selectWhereQuery, dbFilterArgs...)
	}
	if rdr.Config().Opts.SQLBatchSize != nil && *rdr.Config().Opts.SQLBatchSize > 0 {
		tx = tx.Limit(*rdr.Config().Opts.SQLBatchSize)
	}
	if err := tx.Find(&urs).Error; err != nil {
		return err
	}
	for _, urSQL := range urs {
		if rdr.Config().ProcessedPath == utils.MetaDelete {
			if err := db.Table(rdr.tableName).Delete(&utils.URSQLTable{}, urSQL.ID).Error; err != nil {
				return err
			}
		}
		if rdrEv, err := rdr.processMessage(urSQL, lazyFilters); err != nil {
			return err
		} else { // send events to ERs ListenAndServe
			rdrEvChan := rdr.rdrEvents
			if _, isPartial := rdrEv.cgrEvent.APIOpts[utils.PartialOpt]; isPartial {
				rdrEvChan = rdr.partialEvents
			}
			rdrEvChan <- rdrEv
		}
	}
	return nil
}

func (rdr *CgrUR) processMessage(urSql *utils.URSQLTable, lazyFilters []string) (*erEvent, error) {
	ur := &utils.UR{
		Tenant:    urSql.Tenant,
		Opts:      urSql.Opts,
		Event:     urSql.Event,
		CreatedAt: urSql.CreatedAt,
		UpdatedAt: urSql.UpdatedAt,
		DeletedAt: urSql.DeletedAt,
	}
	cgrEv := ur.CGREvent()
	if pass, err := rdr.fltrS.Pass(context.TODO(), cgrEv.Tenant, lazyFilters,
		cgrEv.AsDataProvider()); err != nil || !pass {
		return nil, err
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
				return nil, err
			}
		}
		if err := agReq.SetFields(rdr.Config().Fields); err != nil {
			return nil, err
		}
		if ev := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant,
			utils.NestingSep, agReq.Opts); ev != nil {
			cgrEv.Event = ev.Event
		} else {
			cgrEv.Event = make(map[string]any)
		}
		cgrEv.APIOpts = agReq.Opts
	}
	var rawEvent map[string]any
	if len(rdr.Config().EEsSuccessIDs) != 0 || len(rdr.Config().EEsFailedIDs) != 0 {
		rawEvent = map[string]any{
			utils.ID:           urSql.ID,
			utils.Tenant:       urSql.Tenant,
			utils.OptsCfg:      urSql.Opts,
			utils.EventLowCase: urSql.Event,
		}
	}
	return &erEvent{
		cgrEvent: cgrEv,
		rawEvent: rawEvent,
		rdrCfg:   rdr.Config(),
	}, nil
}

func (rdr *CgrUR) setURL(inURL string, opts *config.EventReaderOpts) error {
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

	rdr.tableName = utils.URsTBL
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

// GetReaderEvents will gather all *cgrur reader events found using the filter
func (rdr *CgrUR) GetReaderEvents(filters []string) ([]*erEvent, error) {
	if rdr.connType != utils.MySQL {
		return nil, errors.New("manual cgrur processing supports only mysql")
	}
	db, sqlDB, err := rdr.openDB()
	if err != nil {
		return nil, err
	}
	defer func() { _ = sqlDB.Close() }()
	dbFilters, dbFilterArgs, lazyFilters, err := engine.FiltersForQuery(filters, rdr.dm, rdr.cgrCfg.GeneralCfg().DefaultTenant)
	if err != nil {
		return nil, err
	}
	var urs []*utils.URSQLTable
	tx := db.Table(rdr.tableName).Model(&utils.URSQLTable{})
	if selectWhereQuery := strings.Join(dbFilters, " AND "); selectWhereQuery != "" {
		tx = tx.Where(selectWhereQuery, dbFilterArgs...)
	}
	if rdr.Config().Opts.SQLBatchSize != nil && *rdr.Config().Opts.SQLBatchSize > 0 {
		tx = tx.Limit(*rdr.Config().Opts.SQLBatchSize)
	}
	if err := tx.Find(&urs).Error; err != nil {
		return nil, err
	}
	rdrEvents := make([]*erEvent, 0) // collect reader events
	for _, urSQL := range urs {
		if rdr.Config().ProcessedPath == utils.MetaDelete {
			if err := db.Table(rdr.tableName).Delete(&utils.URSQLTable{}, urSQL.ID).Error; err != nil {
				return nil, err
			}
		}

		urSQL.Opts[utils.MetaUR] = true // make sure we get reply from sessions for *ur when processed
		if rdrEv, err := rdr.processMessage(urSQL, lazyFilters); err != nil {
			return nil, err
		} else {
			rdrEvents = append(rdrEvents, rdrEv)
		}
	}
	return rdrEvents, nil
}
