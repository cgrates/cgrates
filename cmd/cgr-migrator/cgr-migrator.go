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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/migrator"
	"github.com/cgrates/cgrates/utils"
)

var (
	oldDataDB       migrator.V1DataDB
	oldstorDB       engine.Storage
	oStorDBType     string
	odataDBType     string
	oDBDataEncoding string
	migrate         = flag.String("migrate", "", "Fire up automatic migration *to use multiple values use ',' as separator \n <*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups> ")
	version         = flag.Bool("version", false, "Prints the application version.")

	dataDBType = flag.String("datadb_type", config.CgrConfig().DataDbType, "The type of the DataDb database <redis>")
	dataDBHost = flag.String("datadb_host", config.CgrConfig().DataDbHost, "The DataDb host to connect to.")
	dataDBPort = flag.String("datadb_port", config.CgrConfig().DataDbPort, "The DataDb port to bind to.")
	dataDBName = flag.String("datadb_name", config.CgrConfig().DataDbName, "The name/number of the DataDb to connect to.")
	dataDBUser = flag.String("datadb_user", config.CgrConfig().DataDbUser, "The DataDb user to sign in as.")
	dataDBPass = flag.String("datadb_passwd", config.CgrConfig().DataDbPass, "The DataDb user's password.")

	storDBType = flag.String("stordb_type", config.CgrConfig().StorDBType, "The type of the storDb database <mysql|postgres>")
	storDBHost = flag.String("stordb_host", config.CgrConfig().StorDBHost, "The storDb host to connect to.")
	storDBPort = flag.String("stordb_port", config.CgrConfig().StorDBPort, "The storDb port to bind to.")
	storDBName = flag.String("stordb_name", config.CgrConfig().StorDBName, "The name/number of the storDb to connect to.")
	storDBUser = flag.String("stordb_user", config.CgrConfig().StorDBUser, "The storDb user to sign in as.")
	storDBPass = flag.String("stordb_passwd", config.CgrConfig().StorDBPass, "The storDb user's password.")

	oldDataDBType = flag.String("old_datadb_type", "", "The type of the DataDb database <redis>")
	oldDataDBHost = flag.String("old_datadb_host", config.CgrConfig().DataDbHost, "The DataDb host to connect to.")
	oldDataDBPort = flag.String("old_datadb_port", config.CgrConfig().DataDbPort, "The DataDb port to bind to.")
	oldDataDBName = flag.String("old_datadb_name", config.CgrConfig().DataDbName, "The name/number of the DataDb to connect to.")
	oldDataDBUser = flag.String("old_datadb_user", config.CgrConfig().DataDbUser, "The DataDb user to sign in as.")
	oldDataDBPass = flag.String("old_datadb_passwd", config.CgrConfig().DataDbPass, "The DataDb user's password.")

	oldStorDBType = flag.String("old_stordb_type", "", "The type of the storDb database <mysql|postgres>")
	oldStorDBHost = flag.String("old_stordb_host", config.CgrConfig().StorDBHost, "The storDb host to connect to.")
	oldStorDBPort = flag.String("old_stordb_port", config.CgrConfig().StorDBPort, "The storDb port to bind to.")
	oldStorDBName = flag.String("old_stordb_name", config.CgrConfig().StorDBName, "The name/number of the storDb to connect to.")
	oldStorDBUser = flag.String("old_stordb_user", config.CgrConfig().StorDBUser, "The storDb user to sign in as.")
	oldStorDBPass = flag.String("old_stordb_passwd", config.CgrConfig().StorDBPass, "The storDb user's password.")

	loadHistorySize    = flag.Int("load_history_size", config.CgrConfig().LoadHistorySize, "Limit the number of records in the load history")
	oldLoadHistorySize = flag.Int("old_load_history_size", 0, "Limit the number of records in the load history")

	dbDataEncoding    = flag.String("dbdata_encoding", config.CgrConfig().DBDataEncoding, "The encoding used to store object data in strings")
	oldDBDataEncoding = flag.String("old_dbdata_encoding", "", "The encoding used to store object data in strings")
	dryRun            = flag.Bool("dry_run", false, "When true will not save loaded data to dataDb but just parse it for consistency and errors.")
	verbose           = flag.Bool("verbose", false, "Enable detailed verbose logging output")
	stats             = flag.Bool("stats", false, "Generates statsistics about migrated data.")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}
	if migrate != nil && *migrate != "" { // Run migrator
		if *verbose {
			log.Print("Initializing dataDB:", *dataDBType)
			log.Print("Initializing storDB:", *storDBType)
		}
		dataDB, err := engine.ConfigureDataStorage(*dataDBType, *dataDBHost, *dataDBPort, *dataDBName, *dataDBUser, *dataDBPass, *dbDataEncoding, config.CgrConfig().CacheConfig, *loadHistorySize)
		if err != nil {
			log.Fatal(err)
		}
		storDB, err := engine.ConfigureStorStorage(*storDBType, *storDBHost, *storDBPort, *storDBName, *storDBUser, *storDBPass, *dbDataEncoding,
			config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
		if err != nil {
			log.Fatal(err)
		}
		if *oldDataDBType == "" {
			*oldDataDBType = *dataDBType
			*oldDataDBHost = *dataDBHost
			*oldDataDBPort = *dataDBPort
			*oldDataDBName = *dataDBName
			*oldDataDBUser = *dataDBUser
			*oldDataDBPass = *dataDBPass
		}
		if *verbose {
			log.Print("Initializing oldDataDB:", *oldDataDBType)
		}
		oldDataDB, err := migrator.ConfigureV1DataStorage(*oldDataDBType, *oldDataDBHost, *dataDBPort, *dataDBName, *dataDBUser, *dataDBPass, *dbDataEncoding)
		if err != nil {
			log.Fatal(err)
		}
		oldstorDB = storDB

		if *verbose {
			if *oldStorDBType != "" {
				log.Print("Initializing oldstorDB:", *oldStorDBType)
			} else {
				log.Print("Initializing oldstorDB:", *storDBType)
			}
		}
		if *oldStorDBType != "" {
			oldstorDB, err = engine.ConfigureStorStorage(oStorDBType, *oldStorDBHost, *oldStorDBPort, *oldStorDBName, *oldStorDBUser, *oldStorDBPass, *oldDBDataEncoding,
				config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
			if err != nil {
				log.Fatal(err)
			}
		}
		if *verbose {
			log.Print("Migrating: ", *migrate)
		}
		m, err := migrator.NewMigrator(dataDB, *dataDBType, *dbDataEncoding, storDB, *storDBType, oldDataDB, *oldDataDBType, *oldDBDataEncoding, oldstorDB, *oldStorDBType, *dryRun)
		if err != nil {
			log.Fatal(err)
		}
		migrstats := make(map[string]int)
		mig := strings.Split(*migrate, ",")
		log.Print("migrating", mig)
		err, migrstats = m.Migrate(mig)
		if err != nil {
			log.Fatal(err)
		}
		if *stats != false {
			for k, v := range migrstats {
				log.Print(" ", k, " : ", v)
			}
		}
		if *verbose {
			log.Print("Done migrating!")
		}

		return
	}

}
