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
	"gorm.io/gorm/logger"
)

// NewPostgresStorage returns the posgres storDB
func NewPostgresStorage(host, port, name, user, password, pgSchema,
	sslmode, sslcert, sslkey, sslpassword, sslcertmode, sslrootcert string,
	maxConn, maxIdleConn, logLevel int, connMaxLifetime time.Duration) (*SQLStorage, error) {
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
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{AllowGlobalUpdate: true, Logger: logger.Default.LogMode(logger.LogLevel(logLevel))})
	if err != nil {
		return nil, err
	}
	pgStor := new(PostgresStorage)
	if pgStor.Db, err = db.DB(); err != nil {
		return nil, err
	}
	if err = pgStor.Db.Ping(); err != nil {
		return nil, err
	}
	if pgSchema != "" {
		pgStor.Db.Exec(fmt.Sprintf("set search_path='%s'", pgSchema))
	}
	pgStor.Db.SetMaxIdleConns(maxIdleConn)
	pgStor.Db.SetMaxOpenConns(maxConn)
	pgStor.Db.SetConnMaxLifetime(connMaxLifetime)
	//db.LogMode(true)
	pgStor.db = db
	return &SQLStorage{
		Db:      pgStor.Db,
		db:      pgStor.db,
		StorDB:  pgStor,
		SQLImpl: pgStor,
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
	return utils.MetaPostgres
}
