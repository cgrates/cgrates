// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// NewPostgresStorage returns the posgres storDB
func NewPostgresStorage(host, port, name, user, password, sslmode string, maxConn, maxIdleConn, connMaxLifetime int) (*SQLStorage, error) {
	connectString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", host, port, name, user, password, sslmode)
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
	db.DB().SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
	//db.LogMode(true)
	postgressStorage := new(PostgresStorage)
	postgressStorage.db = db
	postgressStorage.Db = db.DB()
	return &SQLStorage{db.DB(), db, postgressStorage, postgressStorage}, nil
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
	return utils.MetaPostgres
}
