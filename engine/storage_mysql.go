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

package engine

import (
	"fmt"
	"time"

	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLStorage struct {
	SQLStorage
}

func NewMySQLStorage(host, port, name, user, password string,
	maxConn, maxIdleConn, connMaxLifetime int, location string) (*SQLStorage, error) {
	connectString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=%s&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		user, password, host, port, name, location)
	db, err := gorm.Open(mysql.Open(connectString), &gorm.Config{AllowGlobalUpdate: true})
	if err != nil {
		return nil, err
	}

	mySQLStorage := new(MySQLStorage)
	if mySQLStorage.Db, err = db.DB(); err != nil {
		return nil, err
	}
	if mySQLStorage.Db.Ping(); err != nil {
		return nil, err
	}
	mySQLStorage.Db.SetMaxIdleConns(maxIdleConn)
	mySQLStorage.Db.SetMaxOpenConns(maxConn)
	mySQLStorage.Db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
	//db.LogMode(true)
	mySQLStorage.db = db
	return &SQLStorage{
		Db:      mySQLStorage.Db,
		db:      mySQLStorage.db,
		StorDB:  mySQLStorage,
		SQLImpl: mySQLStorage,
	}, nil
}

// SetVersions will set a slice of versions, updating existing
func (self *MySQLStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	tx := self.db.Begin()
	if overwrite {
		tx.Table(utils.TBLVersions).Delete(nil)
	}
	for key, val := range vrs {
		vrModel := &TBLVersion{Item: key, Version: val}
		if err = tx.Save(vrModel).Error; err != nil {
			if err = tx.Model(&TBLVersion{}).Where(
				TBLVersion{Item: vrModel.Item}).Updates(TBLVersion{Version: val}).Error; err != nil {
				tx.Rollback()
				return
			}
		}
	}
	tx.Commit()
	return
}

func (self *MySQLStorage) extraFieldsExistsQry(field string) string {
	return fmt.Sprintf(" extra_fields LIKE '%%\"%s\":%%'", field)
}

func (self *MySQLStorage) extraFieldsValueQry(field, value string) string {
	return fmt.Sprintf(" extra_fields LIKE '%%\"%s\":\"%s\"%%'", field, value)
}

func (self *MySQLStorage) notExtraFieldsExistsQry(field string) string {
	return fmt.Sprintf(" extra_fields NOT LIKE '%%\"%s\":%%'", field)
}

func (self *MySQLStorage) notExtraFieldsValueQry(field, value string) string {
	return fmt.Sprintf(" extra_fields NOT LIKE '%%\"%s\":\"%s\"%%'", field, value)
}

func (self *MySQLStorage) GetStorageType() string {
	return utils.MySQL
}
