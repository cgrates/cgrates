/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package engine

import (
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"path"

	_ "github.com/bmizerany/pq"
	"github.com/jinzhu/gorm"
)

type PostgresStorage struct {
	*SQLStorage
}

func NewPostgresStorage(host, port, name, user, password string, maxConn, maxIdleConn int) (Storage, error) {
	connectString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, name, user, password)
	db, err := gorm.Open("postgres", connectString)
	if err != nil {
		return nil, err
	}
	err = db.DB().Ping()
	if err != nil {
		return nil, err
	}
	db.DB().SetMaxIdleConns(maxIdleConn)
	db.DB().SetMaxOpenConns(maxConn)
	//db.LogMode(true)

	return &PostgresStorage{&SQLStorage{Db: db.DB(), db: db}}, nil
}

func (self *PostgresStorage) Flush() (err error) {
	cfg := config.CgrConfig()
	for _, scriptName := range []string{CREATE_CDRS_TABLES_SQL, CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := self.CreateTablesFromScript(path.Join(cfg.DataFolderPath, "storage", utils.POSTGRES, scriptName)); err != nil {
			return err
		}
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := self.Db.Query(fmt.Sprintf("SELECT 1 FROM %s", tbl)); err != nil {
			return err
		}
	}
	return nil
}

func (self *PostgresStorage) SetTPTiming(tm *utils.ApierTPTiming) error {
	if tm == nil {
		return nil //Nothing to set
	}
	tx := self.db.Begin()
	if err := tx.Where(&TpTiming{Tpid: tm.TPid, Tag: tm.TimingId}).Delete(TpTiming{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(&TpTiming{Tpid: tm.TPid, Tag: tm.TimingId, Years: tm.Years, Months: tm.Months, MonthDays: tm.MonthDays, WeekDays: tm.WeekDays, Time: tm.Time}).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
