// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"
	"time"

	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQLStorage struct {
	SQLStorage
}

func NewMySQLStorage(host, port, name, user, password, mrshlerStr string,
	maxConn, maxIdleConn, logLevel int, connMaxLifetime time.Duration, location string, dsnParams map[string]string) (*SQLStorage, error) {
	connectString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=%s&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		user, password, host, port, name, location)

	ms, err := NewMarshaler(mrshlerStr)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(mysql.Open(connectString+AppendToMysqlDSNOpts(dsnParams)), &gorm.Config{AllowGlobalUpdate: true, Logger: logger.Default.LogMode(logger.LogLevel(logLevel))})
	if err != nil {
		return nil, err
	}
	mySQLStorage := new(MySQLStorage)
	if mySQLStorage.Db, err = db.DB(); err != nil {
		return nil, err
	}
	if err := mySQLStorage.Db.Ping(); err != nil {
		return nil, err
	}
	mySQLStorage.Db.SetMaxIdleConns(maxIdleConn)
	mySQLStorage.Db.SetMaxOpenConns(maxConn)
	mySQLStorage.Db.SetConnMaxLifetime(connMaxLifetime)
	//db.LogMode(true)
	mySQLStorage.db = db
	return &SQLStorage{
		Db:      mySQLStorage.Db,
		db:      mySQLStorage.db,
		StorDB:  mySQLStorage,
		SQLImpl: mySQLStorage,
		ms:      ms,
	}, nil
}

func AppendToMysqlDSNOpts(opts map[string]string) string {
	if opts != nil {
		var dsn string
		for key, val := range opts {
			dsn = dsn + "&" + key + "=" + val
		}
		return dsn
	}
	return utils.EmptyString
}

// SetVersions will set a slice of versions, updating existing
func (msqlS *MySQLStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	tx := msqlS.db.Begin()
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

func (msqlS *MySQLStorage) extraFieldsExistsQry(field string) string {
	return fmt.Sprintf(" extra_fields LIKE '%%\"%s\":%%'", field)
}

func (msqlS *MySQLStorage) extraFieldsValueQry(field, value string) string {
	return fmt.Sprintf(" extra_fields LIKE '%%\"%s\":\"%s\"%%'", field, value)
}

func (msqlS *MySQLStorage) notExtraFieldsExistsQry(field string) string {
	return fmt.Sprintf(" extra_fields NOT LIKE '%%\"%s\":%%'", field)
}

func (msqlS *MySQLStorage) notExtraFieldsValueQry(field, value string) string {
	return fmt.Sprintf(" extra_fields NOT LIKE '%%\"%s\":\"%s\"%%'", field, value)
}

func (msqlS *MySQLStorage) GetStorageType() string {
	return utils.MetaMySQL
}
