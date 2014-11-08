/*
Real-time Charging System for Telecom & ISP environments
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
	"path"
	//"reflect"
	"testing"
	//"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var psqlDb *PostgresStorage

func TestPSQLCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrConfig, _ := config.NewDefaultCGRConfig()
	if d, err := NewPostgresStorage("localhost", "5432", cgrConfig.StorDBName, cgrConfig.StorDBUser, cgrConfig.StorDBPass,
		cgrConfig.StorDBMaxOpenConns, cgrConfig.StorDBMaxIdleConns); err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	} else {
		psqlDb = d.(*PostgresStorage)
	}
	for _, scriptName := range []string{CREATE_CDRS_TABLES_SQL, CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := psqlDb.CreateTablesFromScript(path.Join(*dataDir, "storage", utils.POSTGRES, scriptName)); err != nil {
			t.Error("Error on psqlDb creation: ", err.Error())
			return // No point in going further
		}
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := psqlDb.Db.Query(fmt.Sprintf("SELECT 1 from %s", tbl)); err != nil {
			t.Error(err.Error())
		}
	}
}
