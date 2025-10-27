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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresStorage struct {
	SQLStorage
}

// NewPostgresStorage returns the posgres DB
func NewPostgresStorage(host, port, name, user, password, mrshlerStr string,
	sslmode, sslcert, sslkey, sslpassword, sslcertmode, sslrootcert string,
	maxConn, maxIdleConn, sqlLogLevel int, connMaxLifetime time.Duration) (*SQLStorage, error) {
	var ms utils.Marshaler
	var err error
	if ms, err = utils.NewMarshaler(mrshlerStr); err != nil {
		return nil, err
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, name, user, password, sslmode)
	if sslcert != "" {
		connStr = connStr + " sslcert=" + sslcert
	}
	if sslkey != "" {
		connStr = connStr + " sslkey=" + sslkey
	}
	if sslpassword != "" {
		connStr = connStr + " sslpassword=" + sslpassword
	}
	if sslcertmode != "" {
		connStr = connStr + " sslcertmode=" + sslcertmode
	}
	if sslrootcert != "" {
		connStr = connStr + " sslrootcert=" + sslrootcert
	}
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{AllowGlobalUpdate: true, Logger: logger.Default.LogMode(logger.LogLevel(sqlLogLevel))})
	if err != nil {
		return nil, err
	}
	pgStor := new(PostgresStorage)
	if pgStor.DB, err = db.DB(); err != nil {
		return nil, err
	}
	if err = pgStor.DB.Ping(); err != nil {
		return nil, err
	}
	pgStor.DB.SetMaxIdleConns(maxIdleConn)
	pgStor.DB.SetMaxOpenConns(maxConn)
	pgStor.DB.SetConnMaxLifetime(connMaxLifetime)
	pgStor.ms = ms
	//db.LogMode(true)
	pgStor.db = db
	return &SQLStorage{
		DB:      pgStor.DB,
		db:      pgStor.db,
		ms:      ms,
		DataDB:  pgStor,
		SQLImpl: pgStor,
	}, nil
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

func (poS *PostgresStorage) GetAccountDrv(_ *context.Context, tenant, id string) (ap *utils.Account, err error) {
	var result []*AccountBytesMdl
	if err = poS.db.Model(&AccountBytesMdl{}).Where(&AccountBytesMdl{ID: utils.ConcatenatedKey(tenant, id)}).Find(&result).Error; err != nil {
		return
	}
	if len(result) == 0 {
		return nil, utils.ErrNotFound
	}
	ap = &utils.Account{}
	if err = poS.ms.Unmarshal(result[0].Account, ap); err != nil {
		return
	}
	return
}

func (poS *PostgresStorage) SetAccountDrv(_ *context.Context, ap *utils.Account) (err error) {
	var acc []byte
	if acc, err = poS.ms.Marshal(ap); err != nil {
		return
	}
	tx := poS.db.Begin()
	mdl := &AccountBytesMdl{
		ID:      utils.ConcatenatedKey(ap.Tenant, ap.ID),
		Account: acc,
	}
	if err = tx.Model(&AccountBytesMdl{}).Where(
		AccountBytesMdl{ID: mdl.ID}).Delete(
		AccountBytesMdl{Account: mdl.Account}).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Save(mdl).Error; err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return
}

func (poS *PostgresStorage) RemoveAccountDrv(_ *context.Context, tenant, id string) (err error) {
	tx := poS.db.Begin()
	if err = tx.Model(&AccountBytesMdl{}).Where(&AccountBytesMdl{ID: utils.ConcatenatedKey(tenant, id)}).Delete(&AccountBytesMdl{}).Error; err != nil {
		tx.Rollback()
		return
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

// cdrIDQuery will query the CDR by its unique cdrID
func (poS *PostgresStorage) cdrIDQuery(cdrID string) string {
	return fmt.Sprintf(" opts ->> '*cdrID' = '%s'", cdrID)
}

// existField will query for every element on json type if the field exists
func (poS *PostgresStorage) existField(elem, field string) string {
	return fmt.Sprintf("NOT(%s ? '%s')", elem, field)
}

func (poS *PostgresStorage) GetStorageType() string {
	return utils.MetaPostgres
}

func (poS *PostgresStorage) valueQry(ruleType, elem, field string, values []string, not bool) (conditions []string) {
	// here are for the filters that their values are empty: *exists, *notexists, *empty, *notempty..
	if len(values) == 0 {
		switch ruleType {
		case utils.MetaExists, utils.MetaNotExists:
			if not {
				conditions = append(conditions, fmt.Sprintf("NOT(%s ? '%s')", elem, field))
				return
			}
			conditions = append(conditions, fmt.Sprintf("%s ? '%s'", elem, field))
		case utils.MetaEmpty, utils.MetaNotEmpty:
			if not {
				conditions = append(conditions, fmt.Sprintf(" NOT (%s ->> '%s') = ''", elem, field))
				return
			}
			conditions = append(conditions, fmt.Sprintf(" (%s ->> '%s') = ''", elem, field))
		}
		return
	}
	// here are for the filters that can have more than one value: *string, *prefix, *suffix ..
	for _, value := range values {
		var singleCond string
		switch ruleType {
		case utils.MetaString, utils.MetaNotString, utils.MetaEqual, utils.MetaNotEqual:
			if not {
				conditions = append(conditions, fmt.Sprintf(" NOT (%s ?'%s' AND (%s ->> '%s') = '%s')", elem, field, elem, field, value))
				continue
			}
			singleCond = fmt.Sprintf(" (%s ->> '%s') = '%s'", elem, field, value)
		case utils.MetaLessThan, utils.MetaLessOrEqual, utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
			switch ruleType {
			case utils.MetaGreaterOrEqual:
				singleCond = fmt.Sprintf(" (%s ->> '%s')::numeric >= '%s'", elem, field, value)
			case utils.MetaGreaterThan:
				singleCond = fmt.Sprintf(" (%s ->> '%s')::numeric > '%s'", elem, field, value)
			case utils.MetaLessOrEqual:
				singleCond = fmt.Sprintf(" (%s ->> '%s')::numeric <= '%s'", elem, field, value)
			case utils.MetaLessThan:
				singleCond = fmt.Sprintf(" (%s ->> '%s')::numeric < '%s'", elem, field, value)
			}
		case utils.MetaPrefix, utils.MetaNotPrefix:
			if not {
				conditions = append(conditions, fmt.Sprintf(" NOT ((%s ->> '%s') ILIKE '%s%%')", elem, field, value))
				continue
			}
			singleCond = fmt.Sprintf(" (%s ->> '%s') ILIKE '%s%%'", elem, field, value)
		case utils.MetaSuffix, utils.MetaNotSuffix:
			if not {
				conditions = append(conditions, fmt.Sprintf(" NOT ((%s ->> '%s') ILIKE '%%%s')", elem, field, value))
				continue
			}
			singleCond = fmt.Sprintf(" (%s ->> '%s') ILIKE '%%%s'", elem, field, value)
		case utils.MetaRegex, utils.MetaNotRegex:
			if not {
				conditions = append(conditions, fmt.Sprintf(" (%s ->> '%s') !~ '%s'", elem, field, value))
				continue
			}
			singleCond = fmt.Sprintf(" (%s ->> '%s') ~ '%s'", elem, field, value)
		}
		conditions = append(conditions, singleCond)
	}
	return
}
