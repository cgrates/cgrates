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

package migrator

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type sqlStorage struct {
	Db      *sql.DB
	db      *gorm.DB
	rowIter *sql.Rows
}

func newSqlStorage(host, port, name, user, password string) (*sqlStorage, error) {
	connectString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES,NO_AUTO_CREATE_USER'", user, password, host, port, name)
	db, err := gorm.Open("mysql", connectString)
	if err != nil {
		return nil, err
	}
	if err = db.DB().Ping(); err != nil {
		return nil, err
	}
	return &sqlStorage{Db: db.DB(), db: db}, nil
}

func (sqlStorage *sqlStorage) getV1CDR() (v1Cdr *v1Cdrs, err error) {
	if sqlStorage.rowIter == nil {
		sqlStorage.rowIter, err = sqlStorage.Db.Query("SELECT * FROM cdrs")
		if err != nil {
			return nil, err
		}
	}
	cdrSql := new(engine.CDRsql)
	sqlStorage.rowIter.Scan(&cdrSql)
	v1Cdr, err = NewV1CDRFromCDRSql(cdrSql)

	if sqlStorage.rowIter.Next() {
		v1Cdr = nil
		sqlStorage.rowIter = nil
		return nil, utils.ErrNoMoreData
	}
	return v1Cdr, nil
}

func (sqlStorage *sqlStorage) setV1CDR(v1Cdr *v1Cdrs) (err error) {
	tx := sqlStorage.db.Begin()
	cdrSql := v1Cdr.AsCDRsql()
	cdrSql.CreatedAt = time.Now()
	saved := tx.Save(cdrSql)
	if saved.Error != nil {
		return saved.Error
	}
	tx.Commit()
	return nil
}

func (sqlStorage *sqlStorage) getSMCost() (v2Cost *v2SessionsCost, err error) {
	if sqlStorage.rowIter == nil {
		sqlStorage.rowIter, err = sqlStorage.Db.Query("SELECT * FROM sessions_costs")
		if err != nil {
			return nil, err
		}
	}
	scSql := new(engine.SessionsCostsSQL)
	sqlStorage.rowIter.Scan(&scSql)
	v2Cost, err = NewV2SessionsCostFromSessionsCostSql(scSql)

	if sqlStorage.rowIter.Next() {
		v2Cost = nil
		sqlStorage.rowIter = nil
		return nil, utils.ErrNoMoreData
	}
	return v2Cost, nil
}

func (sqlStorage *sqlStorage) setSMCost(v2Cost *v2SessionsCost) (err error) {
	tx := sqlStorage.db.Begin()
	smSql := v2Cost.AsSessionsCostSql()
	smSql.CreatedAt = time.Now()
	saved := tx.Save(smSql)
	if saved.Error != nil {
		return saved.Error
	}
	tx.Commit()
	return
}

func (sqlStorage *sqlStorage) remSMCost(v2Cost *v2SessionsCost) (err error) {
	tx := sqlStorage.db.Begin()
	var rmParam *engine.SessionsCostsSQL
	if v2Cost != nil {
		rmParam = &engine.SessionsCostsSQL{Cgrid: v2Cost.CGRID,
			RunID: v2Cost.RunID}
	}
	if err := tx.Where(rmParam).Delete(engine.SessionsCostsSQL{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil

}
