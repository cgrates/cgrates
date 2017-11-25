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
	sameDBname      = true
	outDataDB       migrator.MigratorDataDB
	storDB          engine.Storage
	oDBDataEncoding string
	migrate         = flag.String("migrate", "", "Fire up automatic migration "+
		"\n <*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups|*stordb|*datadb> ")
	version = flag.Bool("version", false, "Prints the application version.")

	outDataDBType = flag.String("out_datadb_type", "", "The type of the DataDb Database <redis>")
	outDataDBHost = flag.String("out_datadb_host", config.CgrConfig().DataDbHost, "The DataDb host to connect to.")
	outDataDBPort = flag.String("out_datadb_port", config.CgrConfig().DataDbPort, "The DataDb port to bind to.")
	outDataDBName = flag.String("out_datadb_name", config.CgrConfig().DataDbName, "The name/number of the DataDb to connect to.")
	outDataDBUser = flag.String("out_datadb_user", config.CgrConfig().DataDbUser, "The DataDb user to sign in as.")
	outDataDBPass = flag.String("out_datadb_passwd", config.CgrConfig().DataDbPass, "The DataDb user's password.")

	outStorDBType = flag.String("out_stordb_type", "", "The type of the StorDB Database <mysql|postgres>")
	outStorDBHost = flag.String("out_stordb_host", config.CgrConfig().StorDBHost, "The StorDB host to connect to.")
	outStorDBPort = flag.String("out_stordb_port", config.CgrConfig().StorDBPort, "The StorDB port to bind to.")
	outStorDBName = flag.String("out_stordb_name", config.CgrConfig().StorDBName, "The name/number of the StorDB to connect to.")
	outStorDBUser = flag.String("out_stordb_user", config.CgrConfig().StorDBUser, "The StorDB user to sign in as.")
	outStorDBPass = flag.String("out_stordb_passwd", config.CgrConfig().StorDBPass, "The StorDB user's password.")

	inDataDBType = flag.String("datadb_type", config.CgrConfig().DataDbType, "The type of the DataDb Database <redis>")
	inDataDBHost = flag.String("datadb_host", config.CgrConfig().DataDbHost, "The DataDb host to connect to.")
	inDataDBPort = flag.String("datadb_port", config.CgrConfig().DataDbPort, "The DataDb port to bind to.")
	inDataDBName = flag.String("datadb_name", config.CgrConfig().DataDbName, "The name/number of the DataDb to connect to.")
	inDataDBUser = flag.String("datadb_user", config.CgrConfig().DataDbUser, "The DataDb user to sign in as.")
	inDataDBPass = flag.String("datadb_passwd", config.CgrConfig().DataDbPass, "The DataDb user's password.")

	inStorDBType = flag.String("stordb_type", config.CgrConfig().StorDBType, "The type of the StorDB Database <mysql|postgres>")
	inStorDBHost = flag.String("stordb_host", config.CgrConfig().StorDBHost, "The StorDB host to connect to.")
	inStorDBPort = flag.String("stordb_port", config.CgrConfig().StorDBPort, "The StorDB port to bind to.")
	inStorDBName = flag.String("stordb_name", config.CgrConfig().StorDBName, "The name/number of the StorDB to connect to.")
	inStorDBUser = flag.String("stordb_user", config.CgrConfig().StorDBUser, "The StorDB user to sign in as.")
	inStorDBPass = flag.String("stordb_passwd", config.CgrConfig().StorDBPass, "The StorDB user's password.")

	loadHistorySize = flag.Int("load_history_size", config.CgrConfig().LoadHistorySize, "Limit the number of records in the load history")

	datadb_versions = flag.Bool("datadb_versions", false, "Print DataDB versions")
	stordb_versions = flag.Bool("stordb_versions", false, "Print StorDB versions")

	dbDataEncoding   = flag.String("dbData_encoding", config.CgrConfig().DBDataEncoding, "The encoding used to store object Data in strings")
	inDBDataEncoding = flag.String("in_dbData_encoding", "", "The encoding used to store object Data in strings")
	dryRun           = flag.Bool("dry_run", false, "When true will not save loaded Data to DataDb but just parse it for consistency and errors.")
	verbose          = flag.Bool("verbose", false, "Enable detailed verbose logging output")
	stats            = flag.Bool("stats", false, "Generates statsistics about migrated Data.")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}
	if *verbose {
		log.Print("Initializing DataDB:", *inDataDBType)
		log.Print("Initializing storDB:", *inStorDBType)
	}
	var dmIN *engine.DataManager
	dmIN, _ = engine.ConfigureDataStorage(*inDataDBType, *inDataDBHost, *inDataDBPort,
		*inDataDBName, *inDataDBUser, *inDataDBPass, *dbDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	instorDB, err := engine.ConfigureStorStorage(*inStorDBType, *inStorDBHost, *inStorDBPort, *inStorDBName, *inStorDBUser, *inStorDBPass, *inDBDataEncoding,
		config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	if *outDataDBType == "" {
		*outDataDBType = *inDataDBType
		*outDataDBHost = *inDataDBHost
		*outDataDBPort = *inDataDBPort
		*outDataDBName = *inDataDBName
		*outDataDBUser = *inDataDBUser
		*outDataDBPass = *inDataDBPass
	}
	if *verbose {
		log.Print("Initializing outDataDB:", *outDataDBType)
	}
	var dmOUT *engine.DataManager
	dmOUT, _ = engine.ConfigureDataStorage(*outDataDBType, *outDataDBHost, *outDataDBPort,
		*outDataDBName, *outDataDBUser, *outDataDBPass, *dbDataEncoding, config.CgrConfig().CacheCfg(), *loadHistorySize)
	outDataDB, err := migrator.ConfigureV1DataStorage(*outDataDBType, *outDataDBHost, *outDataDBPort, *outDataDBName, *outDataDBUser, *outDataDBPass, *dbDataEncoding)
	if err != nil {
		log.Fatal(err)
	}
	storDB = instorDB

	if *verbose {
		if *outStorDBType != "" {
			log.Print("Initializing outStorDB:", *outStorDBType)
		} else {
			log.Print("Initializing outStorDB:", *inStorDBType)
		}
	}
	if *outStorDBType != "" {
		storDB, err = engine.ConfigureStorStorage(*outStorDBType, *outStorDBHost, *outStorDBPort, *outStorDBName, *outStorDBUser, *outStorDBPass, *dbDataEncoding,
			config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
		if err != nil {
			log.Fatal(err)
		}
	}
	if *verbose {
		log.Print("Migrating: ", *migrate)
	}
	if inDataDBName != outDataDBName || inStorDBName != outStorDBName {
		sameDBname = false
	}
	m, err := migrator.NewMigrator(dmIN, dmOUT, *inDataDBType, *dbDataEncoding, storDB, *inStorDBType, outDataDB,
		*outDataDBType, *inDBDataEncoding, instorDB, *outStorDBType, *dryRun, sameDBname, *datadb_versions, *stordb_versions)
	if err != nil {
		log.Fatal(err)
	}
	if *datadb_versions {
		vrs, _ := dmOUT.DataDB().GetVersions(utils.TBLVersions)
		if len(vrs) != 0 {
			log.Print("DataDB versions :", vrs)
		} else {
			log.Print("DataDB versions not_found")
		}
	}
	if *stordb_versions {
		vrs, _ := storDB.GetVersions(utils.TBLVersions)
		if len(vrs) != 0 {
			log.Print("StorDB versions :", vrs)
		} else {
			log.Print("StorDB versions not_found")
		}
	}
	if migrate != nil && *migrate != "" { // Run migrator
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
