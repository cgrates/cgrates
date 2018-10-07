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
)

func newMigratorSQL(stor engine.StorDB) (sqlMig *migratorSQL) {
	return &migratorSQL{
		storDB:     &stor,
		sqlStorage: stor.(*engine.SQLStorage),
	}
}

type migratorSQL struct {
	storDB     *engine.StorDB
	sqlStorage *engine.SQLStorage
	rowIter    *sql.Rows
}

func (sqlMig *migratorSQL) StorDB() engine.StorDB {
	return *sqlMig.storDB
}

func (mgSQL *migratorSQL) getV1CDR() (v1Cdr *v1Cdrs, err error) {
	if mgSQL.rowIter == nil {
		mgSQL.rowIter, err = mgSQL.sqlStorage.Db.Query("SELECT * FROM cdrs")
		if err != nil {
			return nil, err
		}
	}
	cdrSql := new(engine.CDRsql)
	mgSQL.rowIter.Scan(&cdrSql)
	v1Cdr, err = NewV1CDRFromCDRSql(cdrSql)

	if mgSQL.rowIter.Next() {
		v1Cdr = nil
		mgSQL.rowIter = nil
		return nil, utils.ErrNoMoreData
	}
	return v1Cdr, nil
}

func (mgSQL *migratorSQL) setV1CDR(v1Cdr *v1Cdrs) (err error) {
	tx := mgSQL.sqlStorage.ExportGormDB().Begin()
	cdrSql := v1Cdr.AsCDRsql()
	cdrSql.CreatedAt = time.Now()
	saved := tx.Save(cdrSql)
	if saved.Error != nil {
		return saved.Error
	}
	tx.Commit()
	return nil
}

func (mgSQL *migratorSQL) renameV1SMCosts() (err error) {
	qry := "RENAME TABLE sm_costs TO sessions_costs;"
	if mgSQL.StorDB().GetStorageType() == utils.POSTGRES {
		qry = "ALTER TABLE sm_costs RENAME TO sessions_costs"
	}
	if _, err := mgSQL.sqlStorage.Db.Exec(qry); err != nil {
		return err
	}
	return
}

func (mgSQL *migratorSQL) createV1SMCosts() (err error) {
	qry := fmt.Sprint("CREATE TABLE sm_costs (  id int(11) NOT NULL AUTO_INCREMENT,  cgrid varchar(40) NOT NULL,  run_id  varchar(64) NOT NULL,  origin_host varchar(64) NOT NULL,  origin_id varchar(128) NOT NULL,  cost_source varchar(64) NOT NULL,  `usage` BIGINT NOT NULL,  cost_details MEDIUMTEXT,  created_at TIMESTAMP NULL,deleted_at TIMESTAMP NULL,  PRIMARY KEY (`id`),UNIQUE KEY costid (cgrid, run_id),KEY origin_idx (origin_host, origin_id),KEY run_origin_idx (run_id, origin_id),KEY deleted_at_idx (deleted_at));")
	if mgSQL.StorDB().GetStorageType() == utils.POSTGRES {
		qry = `
	CREATE TABLE sm_costs (
	  id SERIAL PRIMARY KEY,
	  cgrid VARCHAR(40) NOT NULL,
	  run_id  VARCHAR(64) NOT NULL,
	  origin_host VARCHAR(64) NOT NULL,
	  origin_id VARCHAR(128) NOT NULL,
	  cost_source VARCHAR(64) NOT NULL,
	  usage BIGINT NOT NULL,
	  cost_details jsonb,
	  created_at TIMESTAMP WITH TIME ZONE,
	  deleted_at TIMESTAMP WITH TIME ZONE NULL,
	  UNIQUE (cgrid, run_id)
	);
		`
	}
	if _, err := mgSQL.sqlStorage.Db.Exec("DROP TABLE IF EXISTS sessions_costs;"); err != nil {
		return err
	}
	if _, err := mgSQL.sqlStorage.Db.Exec("DROP TABLE IF EXISTS sm_costs;"); err != nil {
		return err
	}
	if _, err := mgSQL.sqlStorage.Db.Exec(qry); err != nil {
		return err
	}
	return
}

func (mgSQL *migratorSQL) getV2SMCost() (v2Cost *v2SessionsCost, err error) {
	if mgSQL.rowIter == nil {
		mgSQL.rowIter, err = mgSQL.sqlStorage.Db.Query("SELECT * FROM sessions_costs")
		if err != nil {
			return nil, err
		}
	}
	scSql := new(engine.SessionsCostsSQL)
	mgSQL.rowIter.Scan(&scSql)
	v2Cost, err = NewV2SessionsCostFromSessionsCostSql(scSql)

	if mgSQL.rowIter.Next() {
		v2Cost = nil
		mgSQL.rowIter = nil
		return nil, utils.ErrNoMoreData
	}
	return v2Cost, nil
}

func (mgSQL *migratorSQL) setV2SMCost(v2Cost *v2SessionsCost) (err error) {
	tx := mgSQL.sqlStorage.ExportGormDB().Begin()
	smSql := v2Cost.AsSessionsCostSql()
	smSql.CreatedAt = time.Now()
	saved := tx.Save(smSql)
	if saved.Error != nil {
		return saved.Error
	}
	tx.Commit()
	return
}

func (mgSQL *migratorSQL) remV2SMCost(v2Cost *v2SessionsCost) (err error) {
	tx := mgSQL.sqlStorage.ExportGormDB().Begin()
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
