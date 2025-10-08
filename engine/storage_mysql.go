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

func NewMySQLStorage(host, port, name, user, password string,
	maxConn, maxIdleConn, logLevel int, connMaxLifetime time.Duration, location string, dsnParams map[string]string) (*SQLStorage, error) {
	connectString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=%s&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
		user, password, host, port, name, location)
	db, err := gorm.Open(mysql.Open(connectString+AppendToMysqlDSNOpts(dsnParams)), &gorm.Config{AllowGlobalUpdate: true, Logger: logger.Default.LogMode(logger.LogLevel(logLevel))})
	if err != nil {
		return nil, err
	}
	mySQLStorage := new(MySQLStorage)
	if mySQLStorage.DB, err = db.DB(); err != nil {
		return nil, err
	}
	if err := mySQLStorage.DB.Ping(); err != nil {
		return nil, err
	}
	mySQLStorage.DB.SetMaxIdleConns(maxIdleConn)
	mySQLStorage.DB.SetMaxOpenConns(maxConn)
	mySQLStorage.DB.SetConnMaxLifetime(connMaxLifetime)
	//db.LogMode(true)
	mySQLStorage.db = db
	return &SQLStorage{
		DB:      mySQLStorage.DB,
		db:      mySQLStorage.db,
		StorDB:  mySQLStorage,
		SQLImpl: mySQLStorage,
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

// cdrIDQuery will query the CDR by its unique cdrID
func (msqlS *MySQLStorage) cdrIDQuery(cdrID string) string {
	return fmt.Sprintf(" JSON_VALUE(opts, '$.\"*cdrID\"') = '%s'", cdrID)
}

// existField will query for every element on json type if the field exists
func (msqlS *MySQLStorage) existField(elem, field string) string {
	return fmt.Sprintf("!JSON_EXISTS(%s, '$.\"%s\"')", elem, field)
}

func (msqlS *MySQLStorage) GetStorageType() string {
	return utils.MetaMySQL
}

func (msqlS *MySQLStorage) valueQry(ruleType, elem, field string, values []string, not bool) (conditions []string) {
	// here are for the filters that their values are empty: *exists, *notexists, *empty, *notempty..
	if len(values) == 0 {
		switch ruleType {
		case utils.MetaExists, utils.MetaNotExists:
			if not {
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') IS NULL", elem, field))
				return
			}
			conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') IS NOT NULL", elem, field))
		case utils.MetaEmpty, utils.MetaNotEmpty:
			if not {
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') != ''", elem, field))
				return
			}
			conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') == ''", elem, field))
		}
		return
	}
	// here are for the filters that can have more than one value: *string, *prefix, *suffix ..
	for _, value := range values {
		value := verifyBool(value) // in case we have boolean values, it should be queried over 1 or 0
		var singleCond string
		switch ruleType {
		case utils.MetaString, utils.MetaNotString, utils.MetaEqual, utils.MetaNotEqual:
			if not {
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') != '%s'",
					elem, field, value))
				continue
			}
			singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') = '%s'", elem, field, value)
		case utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
			switch ruleType {
			case utils.MetaGreaterOrEqual:
				singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') >= %s", elem, field, value)
			case utils.MetaGreaterThan:
				singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') > %s", elem, field, value)
			case utils.MetaLessOrEqual:
				singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') <= %s", elem, field, value)
			case utils.MetaLessThan:
				singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') < %s", elem, field, value)
			}
		case utils.MetaPrefix, utils.MetaNotPrefix:
			if not {
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') NOT LIKE '%s%%'", elem, field, value))
				continue
			}
			singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') LIKE '%s%%'", elem, field, value)
		case utils.MetaSuffix, utils.MetaNotSuffix:
			if not {
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') NOT LIKE '%%%s'", elem, field, value))
				continue
			}
			singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') LIKE '%%%s'", elem, field, value)
		case utils.MetaRegex, utils.MetaNotRegex:
			if not {
				conditions = append(conditions, fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') NOT REGEXP '%s'", elem, field, value))
				continue
			}
			singleCond = fmt.Sprintf(" JSON_VALUE(%s, '$.\"%s\"') REGEXP '%s'", elem, field, value)
		}
		conditions = append(conditions, singleCond)
	}
	return
}

// verifyBool will check the value for booleans in roder to query properly
func verifyBool(value string) string {
	switch value {
	case "true":
		return "1"
	case "false":
		return "0"
	default:
		return value
	}
}
