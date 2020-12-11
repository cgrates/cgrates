/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgresStorage returns the posgres storDB
func NewPostgresStorage(host, port, name, user, password, sslmode string, maxConn, maxIdleConn, connMaxLifetime int) (*SQLStorage, error) {
	connectString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", host, port, name, user, password, sslmode)
	db, err := gorm.Open(postgres.Open(connectString), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	postgressStorage := new(PostgresStorage)
	if postgressStorage.Db, err = db.DB(); err != nil {
		return nil, err
	}
	if err = postgressStorage.Db.Ping(); err != nil {
		return nil, err
	}
	postgressStorage.Db.SetMaxIdleConns(maxIdleConn)
	postgressStorage.Db.SetMaxOpenConns(maxConn)
	postgressStorage.Db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
	//db.LogMode(true)
	postgressStorage.db = db
	return &SQLStorage{
		Db:      postgressStorage.Db,
		db:      postgressStorage.db,
		StorDB:  postgressStorage,
		SQLImpl: postgressStorage,
	}, nil
}

type PostgresStorage struct {
	SQLStorage
}

func (self *PostgresStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	tx := self.db.Begin()
	if overwrite {
		tx.Table(utils.TBLVersions).Delete(nil)
	}
	for key, val := range vrs {
		vrModel := &TBLVersion{Item: key, Version: val}
		if !overwrite {
			if err = tx.Model(&TBLVersion{}).Where(
				TBLVersion{Item: vrModel.Item}).Delete(TBLVersion{Version: val}).Error; err != nil {
				tx.Rollback()
				return
			}
		}
		if err = tx.Save(vrModel).Error; err != nil {
			tx.Rollback()
			return
		}
	}
	tx.Commit()
	return
}

func (self *PostgresStorage) extraFieldsExistsQry(field string) string {
	return fmt.Sprintf(" extra_fields ?'%s'", field)
}

func (self *PostgresStorage) extraFieldsValueQry(field, value string) string {
	return fmt.Sprintf(" (extra_fields ->> '%s') = '%s'", field, value)
}

func (self *PostgresStorage) notExtraFieldsExistsQry(field string) string {
	return fmt.Sprintf(" NOT extra_fields ?'%s'", field)
}

func (self *PostgresStorage) notExtraFieldsValueQry(field, value string) string {
	return fmt.Sprintf(" NOT (extra_fields ?'%s' AND (extra_fields ->> '%s') = '%s')", field, field, value)
}

func (self *PostgresStorage) GetStorageType() string {
	return utils.Postgres
}
