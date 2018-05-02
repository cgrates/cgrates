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
	dfltCfg    = config.CgrConfig()
	sameDataDB = true
	sameStorDB = true
	storDB     engine.StorDB
	instorDB   engine.StorDB

	dmIN      *engine.DataManager
	dmOUT     *engine.DataManager
	outDataDB migrator.MigratorDataDB
	err       error

	oDBDataEncoding string

	migrate = flag.String("migrate", "", "Fire up automatic migration "+
		"\n <*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups|*stordb|*datadb> ")
	version = flag.Bool("version", false, "Prints the application version.")

	inDataDBType = flag.String("datadb_type", dfltCfg.DataDbType,
		"The type of the DataDb Database <*redis>")
	inDataDBHost = flag.String("datadb_host", dfltCfg.DataDbHost,
		"The DataDb host to connect to.")
	inDataDBPort = flag.String("datadb_port", dfltCfg.DataDbPort,
		"The DataDb port to bind to.")
	inDataDBName = flag.String("datadb_name", dfltCfg.DataDbName,
		"The name/number of the DataDb to connect to.")
	inDataDBUser = flag.String("datadb_user", dfltCfg.DataDbUser,
		"The DataDb user to sign in as.")
	inDataDBPass = flag.String("datadb_passwd", dfltCfg.DataDbPass,
		"The DataDb user's password.")

	inStorDBType = flag.String("stordb_type", dfltCfg.StorDBType,
		"The type of the StorDB Database <*mysql|*postgres>")
	inStorDBHost = flag.String("stordb_host", dfltCfg.StorDBHost,
		"The StorDB host to connect to.")
	inStorDBPort = flag.String("stordb_port", dfltCfg.StorDBPort,
		"The StorDB port to bind to.")
	inStorDBName = flag.String("stordb_name", dfltCfg.StorDBName,
		"The name/number of the StorDB to connect to.")
	inStorDBUser = flag.String("stordb_user", dfltCfg.StorDBUser,
		"The StorDB user to sign in as.")
	inStorDBPass = flag.String("stordb_passwd", dfltCfg.StorDBPass,
		"The StorDB user's password.")

	outDataDBType = flag.String("out_datadb_type", utils.MetaDynamic, "The type of the DataDb Database <*redis|*mongo>")
	outDataDBHost = flag.String("out_datadb_host", utils.MetaDynamic, "The DataDb host to connect to.")
	outDataDBPort = flag.String("out_datadb_port", utils.MetaDynamic, "The DataDb port to bind to.")
	outDataDBName = flag.String("out_datadb_name", utils.MetaDynamic, "The name/number of the DataDb to connect to.")
	outDataDBUser = flag.String("out_datadb_user", utils.MetaDynamic, "The DataDb user to sign in as.")
	outDataDBPass = flag.String("out_datadb_passwd", utils.MetaDynamic, "The DataDb user's password.")

	outStorDBType = flag.String("out_stordb_type", utils.MetaDynamic, "The type of the StorDB Database <*mysql|*postgres|*mongo>")
	outStorDBHost = flag.String("out_stordb_host", utils.MetaDynamic, "The StorDB host to connect to.")
	outStorDBPort = flag.String("out_stordb_port", utils.MetaDynamic, "The StorDB port to bind to.")
	outStorDBName = flag.String("out_stordb_name", utils.MetaDynamic, "The name/number of the StorDB to connect to.")
	outStorDBUser = flag.String("out_stordb_user", utils.MetaDynamic, "The StorDB user to sign in as.")
	outStorDBPass = flag.String("out_stordb_passwd", utils.MetaDynamic, "The StorDB user's password.")

	datadb_versions = flag.Bool("datadb_versions", false, "Print DataDB versions")
	stordb_versions = flag.Bool("stordb_versions", false, "Print StorDB versions")

	inDBDataEncoding = flag.String("dbdata_encoding", dfltCfg.DBDataEncoding,
		"The encoding used to store object Data in strings")

	outDBDataEncoding = flag.String("out_dbdata_encoding", "",
		"The encoding used to store object Data in strings")
	dryRun  = flag.Bool("dry_run", false, "When true will not save loaded Data to DataDb but just parse it for consistency and errors.")
	verbose = flag.Bool("verbose", false, "Enable detailed verbose logging output")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}

	ldrCfg := config.CgrConfig()

	if *inDataDBType != dfltCfg.DataDbType {
		ldrCfg.DataDbType = *inDataDBType
	}

	if *inDataDBHost != dfltCfg.DataDbHost {
		ldrCfg.DataDbHost = *inDataDBHost
	}

	if *inDataDBPort != dfltCfg.DataDbPort {
		ldrCfg.DataDbPort = *inDataDBPort
	}

	if *inDataDBName != dfltCfg.DataDbName {
		ldrCfg.DataDbName = *inDataDBName
	}

	if *inDataDBUser != dfltCfg.DataDbUser {
		ldrCfg.DataDbUser = *inDataDBUser
	}

	if *inDataDBPass != dfltCfg.DataDbPass {
		ldrCfg.DataDbPass = *inDataDBPass
	}

	if *inStorDBType != dfltCfg.StorDBType {
		ldrCfg.StorDBType = *inStorDBType
	}

	if *inStorDBHost != dfltCfg.StorDBHost {
		ldrCfg.StorDBHost = *inStorDBHost
	}

	if *inStorDBPort != dfltCfg.StorDBPort {
		ldrCfg.StorDBPort = *inStorDBPort
	}

	if *inStorDBName != dfltCfg.StorDBName {
		ldrCfg.StorDBName = *inStorDBName
	}

	if *inStorDBUser != dfltCfg.StorDBUser {
		ldrCfg.StorDBUser = *inStorDBUser
	}

	if *inStorDBPass != "" {
		ldrCfg.StorDBPass = *inStorDBPass
	}

	if *inDBDataEncoding != "" {
		ldrCfg.DBDataEncoding = *inDBDataEncoding
	}

	if *outDataDBType == utils.MetaDynamic {
		*outDataDBType = ldrCfg.DataDbType
		*outDataDBHost = ldrCfg.DataDbHost
		*outDataDBPort = ldrCfg.DataDbPort
		*outDataDBName = ldrCfg.DataDbName
		*outDataDBUser = ldrCfg.DataDbUser
		*outDataDBPass = ldrCfg.DataDbPass
	} else {
		*outDataDBType = strings.TrimPrefix(*outDataDBType, "*")
		*outDataDBHost = config.DBDefaults.DBHost(*outDataDBType, *outDataDBHost)
		*outDataDBPort = config.DBDefaults.DBPort(*outDataDBType, *outDataDBPort)
		*outDataDBName = config.DBDefaults.DBName(*outDataDBType, *outDataDBName)
		*outDataDBUser = config.DBDefaults.DBUser(*outDataDBType, *outDataDBUser)
		*outDataDBPass = config.DBDefaults.DBPass(*outDataDBType, *outDataDBPass)
	}

	if *outStorDBType != utils.MetaDynamic {
		*outStorDBType = strings.TrimPrefix(*outStorDBType, "*")
		*outStorDBHost = config.DBDefaults.DBHost(*outStorDBType, *outStorDBHost)
		*outStorDBPort = config.DBDefaults.DBPort(*outStorDBType, *outStorDBPort)
		*outStorDBName = config.DBDefaults.DBName(*outStorDBType, *outStorDBName)
		*outStorDBUser = config.DBDefaults.DBUser(*outStorDBType, *outStorDBUser)
		*outStorDBPass = config.DBDefaults.DBPass(*outStorDBType, *outStorDBPass)
	}

	if dmIN, err = engine.ConfigureDataStorage(ldrCfg.DataDbType, ldrCfg.DataDbHost, ldrCfg.DataDbPort,
		ldrCfg.DataDbName, ldrCfg.DataDbUser, ldrCfg.DataDbPass, ldrCfg.DBDataEncoding,
		config.CgrConfig().CacheCfg(), 0); err != nil {
		log.Fatal(err)
	}
	if instorDB, err = engine.ConfigureStorDB(ldrCfg.StorDBType, ldrCfg.StorDBHost, ldrCfg.StorDBPort,
		ldrCfg.StorDBName, ldrCfg.StorDBUser, ldrCfg.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes); err != nil {
		log.Fatal(err)
	}
	if dmOUT, err = engine.ConfigureDataStorage(*outDataDBType, *outDataDBHost, *outDataDBPort,
		*outDataDBName, *outDataDBUser, *outDataDBPass, ldrCfg.DBDataEncoding,
		config.CgrConfig().CacheCfg(), 0); err != nil {
		log.Fatal(err)
	}
	if outDataDB, err = migrator.ConfigureV1DataStorage(*outDataDBType, *outDataDBHost, *outDataDBPort,
		*outDataDBName, *outDataDBUser, *outDataDBPass, ldrCfg.DBDataEncoding); err != nil {
		log.Fatal(err)
	}

	if *outStorDBType == utils.MetaDynamic {
		storDB = instorDB
	} else {
		storDB, err = engine.ConfigureStorDB(*outStorDBType, *outStorDBHost, *outStorDBPort, *outStorDBName, *outStorDBUser, *outStorDBPass,
			config.CgrConfig().StorDBMaxOpenConns, config.CgrConfig().StorDBMaxIdleConns, config.CgrConfig().StorDBConnMaxLifetime, config.CgrConfig().StorDBCDRSIndexes)
		if err != nil {
			log.Fatal(err)
		}
	}
	if ldrCfg.DataDbName != *outDataDBName || ldrCfg.DataDbType != *outDataDBType || ldrCfg.DataDbHost != *outDataDBHost {
		sameDataDB = false
	}
	if ldrCfg.StorDBName != *outStorDBName || ldrCfg.StorDBType != *outStorDBName || ldrCfg.StorDBHost != *outStorDBHost {
		sameStorDB = false
	}
	m, err := migrator.NewMigrator(dmIN, dmOUT, ldrCfg.DataDbType, ldrCfg.DBDataEncoding, storDB, ldrCfg.StorDBType, outDataDB,
		*outDataDBType, ldrCfg.DBDataEncoding, instorDB, *outStorDBType, *dryRun, sameDataDB, sameStorDB, *datadb_versions, *stordb_versions)
	if err != nil {
		log.Fatal(err)
	}
	if *datadb_versions {
		vrs, _ := dmOUT.DataDB().GetVersions("")
		if len(vrs) != 0 {
			log.Printf("DataDB versions : %+v\n", vrs)
		} else {
			log.Printf("DataDB versions not_found")
		}
	}
	if *stordb_versions {
		vrs, _ := storDB.GetVersions("")
		if len(vrs) != 0 {
			log.Printf("StorDB versions : %+v\n", vrs)
		} else {
			log.Printf("StorDB versions not_found")
		}
	}
	if migrate != nil && *migrate != "" { // Run migrator
		migrstats := make(map[string]int)
		mig := strings.Split(*migrate, ",")
		err, migrstats = m.Migrate(mig)
		if err != nil {
			log.Fatal(err)
		}
		if *verbose != false {
			log.Printf("Data migrated: %+v", migrstats)
		}
		return
	}

}
