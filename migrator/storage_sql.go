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

package migrator

import (
	"database/sql"
	"fmt"

	"github.com/cgrates/cgrates/engine"
	_ "github.com/go-sql-driver/mysql"
)

type migratorSQL struct {
	sqlStorage *engine.SQLStorage
	rowIter    *sql.Rows
}

func (sqlMig *migratorSQL) close() {
	sqlMig.sqlStorage.Close()
}

func (mgSQL *migratorSQL) renameV1SMCosts() (err error) {
	qry := "RENAME TABLE sm_costs TO session_costs;"
	if _, err := mgSQL.sqlStorage.DB.Exec(qry); err != nil {
		return err
	}
	return
}

func (mgSQL *migratorSQL) createV1SMCosts() (err error) {
	qry := fmt.Sprint("CREATE TABLE sm_costs (  id int(11) NOT NULL AUTO_INCREMENT,    run_id  varchar(64) NOT NULL,  origin_host varchar(64) NOT NULL,  origin_id varchar(128) NOT NULL,  cost_source varchar(64) NOT NULL,  `usage` BIGINT NOT NULL,  cost_details MEDIUMTEXT,  created_at TIMESTAMP NULL,deleted_at TIMESTAMP NULL,  PRIMARY KEY (`id`),UNIQUE KEY costid ( run_id),KEY origin_idx (origin_host, origin_id),KEY run_origin_idx (run_id, origin_id),KEY deleted_at_idx (deleted_at));")

	if _, err := mgSQL.sqlStorage.DB.Exec("DROP TABLE IF EXISTS session_costs;"); err != nil {
		return err
	}
	if _, err := mgSQL.sqlStorage.DB.Exec("DROP TABLE IF EXISTS sm_costs;"); err != nil {
		return err
	}
	if _, err := mgSQL.sqlStorage.DB.Exec(qry); err != nil {
		return err
	}
	return
}
