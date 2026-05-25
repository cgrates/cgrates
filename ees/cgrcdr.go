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
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewCgrCDR(cfg *config.EventExporterCfg,
	em *utils.ExporterMetrics) (cgr *CgrCDR, err error) {
	cgr = &CgrCDR{
		cfg:  cfg,
		em:   em,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	err = cgr.initDialector()
	return
}

type CgrCDR struct {
	cfg   *config.EventExporterCfg
	em    *utils.ExporterMetrics
	db    *gorm.DB
	sqldb *sql.DB
	reqs  *concReq

	dialect   gorm.Dialector
	dbType    string
	tableName string
	sync.RWMutex
}

func (cgr *CgrCDR) initDialector() (err error) {
	var u *url.URL
	if u, err = url.Parse(strings.TrimPrefix(cgr.Cfg().ExportPath, utils.Meta)); err != nil {
		return
	}
	password, _ := u.User.Password()

	dbname := utils.SQLDefaultDBName
	if cgr.Cfg().Opts.SQLDBName != nil {
		dbname = *cgr.Cfg().Opts.SQLDBName
	}
	ssl := utils.SQLDefaultPgSSLMode
	if cgr.Cfg().Opts.PgSSLMode != nil {
		ssl = *cgr.Cfg().Opts.PgSSLMode
	}
	cgr.tableName = utils.CDRsTBL
	if cgr.Cfg().Opts.SQLTableName != nil {
		cgr.tableName = *cgr.Cfg().Opts.SQLTableName
	}

	cgr.dbType = u.Scheme
	switch u.Scheme {
	case utils.MySQL:
		cgr.dialect = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			u.User.Username(), password, u.Hostname(), u.Port(), dbname) + engine.AppendToMysqlDSNOpts(cgr.Cfg().Opts.MYSQLDSNParams))
	case utils.Postgres:
		cgr.dialect = postgres.Open(fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
			u.Hostname(), u.Port(), dbname, u.User.Username(), password, ssl))
	default:
		return fmt.Errorf("db type <%s> not supported", u.Scheme)
	}
	return
}

func (cgr *CgrCDR) Cfg() *config.EventExporterCfg { return cgr.cfg }

func (cgr *CgrCDR) Connect() (err error) {
	cgr.Lock()
	if cgr.db == nil || cgr.sqldb == nil {
		cgr.db, cgr.sqldb, err = openDB(cgr.dialect, cgr.Cfg().Opts)
	}
	cgr.Unlock()
	return
}

func (cgr *CgrCDR) Close() (err error) {
	cgr.Lock()
	if cgr.sqldb != nil {
		err = cgr.sqldb.Close()
	}
	cgr.db = nil
	cgr.sqldb = nil
	cgr.Unlock()
	return
}

func (cgr *CgrCDR) GetMetrics() *utils.ExporterMetrics { return cgr.em }

func (cgr *CgrCDR) ExtraData(ev *utils.CGREvent) any { return ev }

// PrepareMap is a no-op it doesn't use templates
func (cgr *CgrCDR) PrepareMap(*utils.CGREvent) (any, error) { return nil, nil }

func (cgr *CgrCDR) PrepareOrderMap(*utils.OrderedNavigableMap) (any, error) { return nil, nil }

func (cgr *CgrCDR) ExportEvent(_ *context.Context, _, extraData any) error {
	cgrEv, ok := extraData.(*utils.CGREvent)
	if !ok {
		return fmt.Errorf("unexpected extraData type %T", extraData)
	}
	if cgrEv.APIOpts == nil {
		cgrEv.APIOpts = make(map[string]any)
	}
	if _, has := cgrEv.APIOpts[utils.MetaCDRID]; !has {
		cgrEv.APIOpts[utils.MetaCDRID] = utils.GetUniqueCDRID(cgrEv)
	}

	cgr.reqs.get()
	cgr.RLock()
	defer func() {
		cgr.RUnlock()
		cgr.reqs.done()
	}()
	if cgr.db == nil {
		return utils.ErrDisconnected
	}

	cdrTable := &utils.CDRSQLTable{
		Tenant:    cgrEv.Tenant,
		Opts:      cgrEv.APIOpts,
		Event:     cgrEv.Event,
		CreatedAt: time.Now(),
	}
	tx := cgr.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Table(cgr.tableName).Save(cdrTable).Error; err != nil {
		tx.Rollback()
		if !strings.Contains(err.Error(), "1062") &&
			!strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("storing CDR %s failed: %v", utils.ToJSON(cgrEv), err)
		}
		cdrID := utils.IfaceAsString(cgrEv.APIOpts[utils.MetaCDRID])
		updTx := cgr.db.Begin()
		if updTx.Error != nil {
			return updTx.Error
		}
		if uerr := updTx.Table(cgr.tableName).Where(cgr.cdrIDQuery(cdrID)).Updates(
			utils.CDRSQLTable{Opts: cgrEv.APIOpts, Event: cgrEv.Event, UpdatedAt: time.Now()}).Error; uerr != nil {
			updTx.Rollback()
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: <%s> updating CDR %s",
					utils.CDRs, uerr.Error(), utils.ToJSON(cgrEv)))
			return utils.ErrPartiallyExecuted
		}
		updTx.Commit()
		return nil
	}
	tx.Commit()
	return nil
}

func (cgr *CgrCDR) cdrIDQuery(cdrID string) string {
	switch cgr.dbType {
	case utils.Postgres:
		return fmt.Sprintf(" opts ->> '*cdrID' = '%s'", cdrID)
	default:
		return fmt.Sprintf(" JSON_VALUE(opts, '$.\"*cdrID\"') = '%s'", cdrID)
	}
}
