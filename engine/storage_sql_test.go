/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cgrates/config"
	"testing"
	"path"
	"fmt"
	"flag"
)

/*
README:

 Enable these tests by passing '-local' to the go test command

 Only database supported for now is MySQL. Postgres could be easily extended if needed.

 It is expected that the data folder of CGRateS exists at path /usr/share/cgrates/data.

 Prior running the tests, create database and users by running:
  mysql -pyourrootpwd < /usr/share/cgrates/data/storage/mysql/create_db_with_users.sql
*/

const (
	CREATE_CDRS_TABLES_SQL = "create_cdrs_tables.sql"
	CREATE_COSTDETAILS_TABLES_SQL = "create_costdetails_tables.sql"
	CREATE_MEDIATOR_TABLES_SQL = "create_mediator_tables.sql"
	CREATE_TARIFFPLAN_TABLES_SQL = "create_tariffplan_tables.sql"
)

var db *MySQLStorage
var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var scriptsPath = flag.String("scripts_path", "/usr/share/cgrates/data/storage/mysql", "Overwrite default scripts path here")

 
func TestCreateTables(t *testing.T) {
	if !*testLocal { 
		return
	}
	cgrConfig, _ := config.NewDefaultCGRConfig()
	if d, err := NewMySQLStorage(cgrConfig.StorDBHost, cgrConfig.StorDBPort, cgrConfig.StorDBName, cgrConfig.StorDBUser, cgrConfig.StorDBPass); err != nil {
		t.Error("Error on opening database connection: ",err)
		return
	} else {
		db = d.(*MySQLStorage)
	}
	for _, scriptName := range []string{CREATE_CDRS_TABLES_SQL, CREATE_COSTDETAILS_TABLES_SQL, CREATE_MEDIATOR_TABLES_SQL, CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := db.CreateTablesFromScript(path.Join(*scriptsPath, scriptName)); err != nil {
			t.Error("Error on db creation: ", err.Error())
			return // No point in going further
		}
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if  _, err := db.Db.Query(fmt.Sprintf("SELECT 1 from %s", tbl)); err != nil {
			t.Error(err.Error())
		}
	}
}
	


