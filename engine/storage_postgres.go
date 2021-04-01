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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgresStorage returns the posgres storDB
func NewPostgresStorage(host, port, name, user, password, sslmode string, maxConn, maxIdleConn, connMaxLifetime int) (*SQLStorage, error) {
	connectString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", host, port, name, user, password, sslmode)
	db, err := gorm.Open(postgres.Open(connectString), &gorm.Config{AllowGlobalUpdate: true})
	if err != nil {
		return nil, err
	}
	postgressStorage := new(PostgresStorage)
	if postgressStorage.DB, err = db.DB(); err != nil {
		return nil, err
	}
	if err = postgressStorage.DB.Ping(); err != nil {
		return nil, err
	}
	postgressStorage.DB.SetMaxIdleConns(maxIdleConn)
	postgressStorage.DB.SetMaxOpenConns(maxConn)
	postgressStorage.DB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
	//db.LogMode(true)
	postgressStorage.db = db
	return &SQLStorage{
		DB:      postgressStorage.DB,
		db:      postgressStorage.db,
		StorDB:  postgressStorage,
		SQLImpl: postgressStorage,
	}, nil
}

type PostgresStorage struct {
	SQLStorage
}

func (poS *PostgresStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	tx := poS.db.Begin()
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

func (poS *PostgresStorage) extraFieldsExistsQry(field string) string {
	return fmt.Sprintf(" extra_fields ?'%s'", field)
}

func (poS *PostgresStorage) extraFieldsValueQry(field, value string) string {
	return fmt.Sprintf(" (extra_fields ->> '%s') = '%s'", field, value)
}

func (poS *PostgresStorage) notExtraFieldsExistsQry(field string) string {
	return fmt.Sprintf(" NOT extra_fields ?'%s'", field)
}

func (poS *PostgresStorage) notExtraFieldsValueQry(field, value string) string {
	return fmt.Sprintf(" NOT (extra_fields ?'%s' AND (extra_fields ->> '%s') = '%s')", field, field, value)
}

func (poS *PostgresStorage) GetStorageType() string {
	return utils.Postgres
}
