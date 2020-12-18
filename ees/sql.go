/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewSQLEe(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (sqlEe *SQLEe, err error) {
	sqlEe = &SQLEe{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS, dc: dc}

	var u *url.URL
	if u, err = url.Parse(strings.TrimPrefix(cgrCfg.EEsCfg().Exporters[cfgIdx].ExportPath, utils.Meta)); err != nil {
		return
	}
	password, _ := u.User.Password()

	dbname := utils.SQLDefaultDBName
	if vals, has := cgrCfg.EEsCfg().Exporters[cfgIdx].Opts[utils.SQLDBName]; has {
		dbname = utils.IfaceAsString(vals)
	}
	ssl := utils.SQLDefaultSSLMode
	if vals, has := cgrCfg.EEsCfg().Exporters[cfgIdx].Opts[utils.SQLSSLMode]; has {
		ssl = utils.IfaceAsString(vals)
	}
	// tableName is mandatory in opts
	if iface, has := cgrCfg.EEsCfg().Exporters[cfgIdx].Opts[utils.SQLTableName]; !has {
		return nil, utils.NewErrMandatoryIeMissing(utils.SQLTableName)
	} else {
		sqlEe.tableName = utils.IfaceAsString(iface)
	}

	var connString string
	switch u.Scheme {
	case utils.MYSQL:
		connString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			u.User.Username(), password, u.Hostname(), u.Port(), dbname)
	case utils.POSTGRES:
		connString = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", u.Hostname(), u.Port(), dbname, u.User.Username(), password, ssl)
	default:
		return nil, fmt.Errorf("unknown db_type %s", u.Scheme)
	}

	db, err := gorm.Open(u.Scheme, connString)
	if err != nil {
		return nil, err
	}
	if err = db.DB().Ping(); err != nil {
		return nil, err
	}

	if iface, has := cgrCfg.EEsCfg().Exporters[cfgIdx].Opts[utils.SQLMaxIdleConns]; has {
		val, err := utils.IfaceAsTInt64(iface)
		if err != nil {
			return nil, err
		}
		db.DB().SetMaxIdleConns(int(val))
	}
	if iface, has := cgrCfg.EEsCfg().Exporters[cfgIdx].Opts[utils.SQLMaxOpenConns]; has {
		val, err := utils.IfaceAsTInt64(iface)
		if err != nil {
			return nil, err
		}
		db.DB().SetMaxOpenConns(int(val))
	}
	if iface, has := cgrCfg.EEsCfg().Exporters[cfgIdx].Opts[utils.SQLMaxConnLifetime]; has {
		val, err := utils.IfaceAsDuration(iface)
		if err != nil {
			return nil, err
		}
		db.DB().SetConnMaxLifetime(val)
	}

	sqlEe.db = db
	return
}

// SQLEe implements EventExporter interface for SQL
type SQLEe struct {
	id      string
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	db      *gorm.DB

	tableName string

	sync.RWMutex
	dc utils.MapStorage
}

// ID returns the identificator of this exporter
func (sqlEe *SQLEe) ID() string {
	return sqlEe.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (sqlEe *SQLEe) OnEvicted(_ string, _ interface{}) {
	return
}

// ExportEvent implements EventExporter
func (sqlEe *SQLEe) ExportEvent(cgrEv *utils.CGREventWithOpts) (err error) {
	sqlEe.Lock()
	defer func() {
		if err != nil {
			sqlEe.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		} else {
			sqlEe.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
		}
		sqlEe.Unlock()
	}()
	sqlEe.dc[utils.NumberOfEvents] = sqlEe.dc[utils.NumberOfEvents].(int64) + 1

	var vals []interface{}
	var colNames []string
	req := utils.MapStorage(cgrEv.Event)
	eeReq := NewEventExporterRequest(req, sqlEe.dc, cgrEv.Opts,
		sqlEe.cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].Tenant,
		sqlEe.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(sqlEe.cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].Timezone,
			sqlEe.cgrCfg.GeneralCfg().DefaultTimezone),
		sqlEe.filterS)
	if err = eeReq.SetFields(sqlEe.cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].ContentFields()); err != nil {
		return
	}

	for el := eeReq.cnt.GetFirstElement(); el != nil; el = el.Next() {
		var iface interface{}
		if iface, err = eeReq.cnt.FieldAsInterface(el.Value.Slice()); err != nil {
			return
		}
		pathWithoutIndex := utils.GetPathWithoutIndex(el.Value.String())
		if pathWithoutIndex != utils.MetaRow {
			colNames = append(colNames, pathWithoutIndex)
		}
		vals = append(vals, iface)
	}

	sqlValues := make([]string, len(vals))
	for i := range vals {
		sqlValues[i] = "?"
	}

	var sqlQuery string
	if len(colNames) != len(vals) {
		sqlQuery = fmt.Sprintf("INSERT INTO %s VALUES (%s); ", sqlEe.tableName, strings.Join(sqlValues, ","))
	} else {
		colNamesStr := "(" + strings.Join(colNames, ", ") + ")"
		sqlQuery = fmt.Sprintf("INSERT INTO %s %s VALUES (%s); ", sqlEe.tableName, colNamesStr, strings.Join(sqlValues, ","))
	}

	sqlEe.db.Table(sqlEe.tableName).Exec(sqlQuery, vals...)
	updateEEMetrics(sqlEe.dc, cgrEv.Event, utils.FirstNonEmpty(sqlEe.cgrCfg.EEsCfg().Exporters[sqlEe.cfgIdx].Timezone,
		sqlEe.cgrCfg.GeneralCfg().DefaultTimezone))

	return
}

func (sqlEe *SQLEe) GetMetrics() utils.MapStorage {
	return sqlEe.dc.Clone()
}
