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
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/cgrates/cgrates/migrator"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	migrate = flag.String("migrate", "", "Fire up automatic migration <*cost_details|*set_versions>")
	version = flag.Bool("version", false, "Prints the application version.")

	dataDBType = flag.String("datadb_type", config.CgrConfig().DataDbType, "The type of the DataDb database <redis>")
	dataDBHost = flag.String("datadb_host", config.CgrConfig().DataDbHost, "The DataDb host to connect to.")
	dataDBPort = flag.String("datadb_port", config.CgrConfig().DataDbPort, "The DataDb port to bind to.")
	dataDBName = flag.String("datadb_name", config.CgrConfig().DataDbName, "The name/number of the DataDb to connect to.")
	dataDBUser = flag.String("datadb_user", config.CgrConfig().DataDbUser, "The DataDb user to sign in as.")
	dataDBPass = flag.String("datadb_passwd", config.CgrConfig().DataDbPass, "The DataDb user's password.")

	storDBType = flag.String("stordb_type", config.CgrConfig().StorDBType, "The type of the storDb database <mysql>")
	storDBHost = flag.String("stordb_host", config.CgrConfig().StorDBHost, "The storDb host to connect to.")
	storDBPort = flag.String("stordb_port", config.CgrConfig().StorDBPort, "The storDb port to bind to.")
	storDBName = flag.String("stordb_name", config.CgrConfig().StorDBName, "The name/number of the storDb to connect to.")
	storDBUser = flag.String("stordb_user", config.CgrConfig().StorDBUser, "The storDb user to sign in as.")
	storDBPass = flag.String("stordb_passwd", config.CgrConfig().StorDBPass, "The storDb user's password.")

	oldDataDBType = flag.String("old_datadb_type", config.CgrConfig().DataDbType, "The type of the DataDb database <redis>")
	oldDataDBHost = flag.String("old_datadb_host", config.CgrConfig().DataDbHost, "The DataDb host to connect to.")
	oldDataDBPort = flag.String("old_datadb_port", config.CgrConfig().DataDbPort, "The DataDb port to bind to.")
	oldDataDBName = flag.String("old_datadb_name", "11", "The name/number of the DataDb to connect to.")
	oldDataDBUser = flag.String("old_datadb_user", config.CgrConfig().DataDbUser, "The DataDb user to sign in as.")
	oldDataDBPass = flag.String("old_datadb_passwd", config.CgrConfig().DataDbPass, "The DataDb user's password.")

	oldStorDBType = flag.String("old_stordb_type", config.CgrConfig().StorDBType, "The type of the storDb database <mysql>")
	oldStorDBHost = flag.String("old_stordb_host", config.CgrConfig().StorDBHost, "The storDb host to connect to.")
	oldStorDBPort = flag.String("old_stordb_port", config.CgrConfig().StorDBPort, "The storDb port to bind to.")
	oldStorDBName = flag.String("old_stordb_name", config.CgrConfig().StorDBName, "The name/number of the storDb to connect to.")
	oldStorDBUser = flag.String("old_stordb_user", config.CgrConfig().StorDBUser, "The storDb user to sign in as.")
	oldStorDBPass = flag.String("old_stordb_passwd", config.CgrConfig().StorDBPass, "The storDb user's password.")

	loadHistorySize = flag.Int("load_history_size", config.CgrConfig().LoadHistorySize, "Limit the number of records in the load history")
	oldLoadHistorySize = flag.Int("old_load_history_size", config.CgrConfig().LoadHistorySize, "Limit the number of records in the load history")

	dbDataEncoding = flag.String("dbdata_encoding", config.CgrConfig().DBDataEncoding, "The encoding used to store object data in strings")
	oldDBDataEncoding = flag.String("old_dbdata_encoding", config.CgrConfig().DBDataEncoding, "The encoding used to store object data in strings")
	//nu salvez doar citesc din oldDb 
	//dryRun          = flag.Bool("dry_run", false, "When true will not save loaded data to dataDb but just parse it for consistency and errors.")
	//verbose         = flag.Bool("verbose", false, "Enable detailed verbose logging output")
	//slice mapstring int  cate acc [0]am citit si [1]cate acc am scris
	//stats           = flag.Bool("stats", false, "Generates statsistics about given data.")

)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}
if migrate != nil && *migrate != "" { // Run migrator
		dataDB, err := engine.ConfigureDataStorage(*dataDBType, *dataDBHost, *dataDBPort, *dataDBName, *dataDBUser, *dataDBPass, *dbDataEncoding, config.CgrConfig().CacheConfig, *loadHistorySize)
		if err != nil {
			log.Fatal(err)
		}
		oldDataDB, err := engine.ConfigureDataStorage(*oldDataDBType, *oldDataDBHost, *oldDataDBPort, *oldDataDBName, *oldDataDBUser, *oldDataDBPass, *oldDBDataEncoding, config.CgrConfig().CacheConfig, *oldLoadHistorySize)
		if err != nil {
			log.Fatal(err)
		}
			storDB, err := engine.ConfigureStorStorage(*storDBType, *storDBHost, *storDBPort, *storDBName, *storDBUser, *storDBPass, *dbDataEncoding,
			config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
		if err != nil {
			log.Fatal(err)
		}
			oldstorDB, err := engine.ConfigureStorStorage(*oldStorDBType, *oldStorDBHost, *oldStorDBPort, *oldStorDBName, *oldStorDBUser, *oldStorDBPass, *oldDBDataEncoding,
			config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
		if err != nil {
			log.Fatal(err)
		}
		m,err := migrator.NewMigrator(dataDB, *dataDBType, *dbDataEncoding, storDB, *storDBType,oldDataDB,*oldDataDBType,*oldDBDataEncoding,oldstorDB,*oldStorDBType)
		 if err != nil {
			log.Fatal(err)
		}
		if err:=m.Migrate(*migrate); err != nil {
			log.Fatal(err)
		}
		log.Print("Done migrating!")
		return
	}
	}
