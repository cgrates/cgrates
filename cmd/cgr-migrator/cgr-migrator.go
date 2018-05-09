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
	"github.com/cgrates/cgrates/migrator"
	"github.com/cgrates/cgrates/utils"
)

const ()

var (
	sameDataDB, sameStorDB bool
	dmIN, dmOUT            migrator.MigratorDataDB
	storDBIn, storDBOut    migrator.MigratorStorDB
	err                    error
	dfltCfg                = config.CgrConfig()
	cfgDir                 = flag.String("config_dir", "",
		"Configuration directory path.")

	migrate = flag.String("migrate", "", "fire up automatic migration "+
		"\n <*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups|*stordb|*datadb>")
	version = flag.Bool("version", false, "prints the application version")

	inDataDBType = flag.String("datadb_type", dfltCfg.DataDbType,
		"the type of the DataDB Database <*redis|*mongo>")
	inDataDBHost = flag.String("datadb_host", dfltCfg.DataDbHost,
		"the DataDB host")
	inDataDBPort = flag.String("datadb_port", dfltCfg.DataDbPort,
		"the DataDB port")
	inDataDBName = flag.String("datadb_name", dfltCfg.DataDbName,
		"the name/number of the DataDB")
	inDataDBUser = flag.String("datadb_user", dfltCfg.DataDbUser,
		"the DataDB user")
	inDataDBPass = flag.String("datadb_passwd", dfltCfg.DataDbPass,
		"the DataDB password")

	outDataDBType = flag.String("out_datadb_type", utils.MetaDataDB,
		"output DataDB type <*redis|*mongo>")
	outDataDBHost = flag.String("out_datadb_host", utils.MetaDataDB,
		"output DataDB host to connect to")
	outDataDBPort = flag.String("out_datadb_port", utils.MetaDataDB,
		"output DataDB port")
	outDataDBName = flag.String("out_datadb_name", utils.MetaDataDB,
		"output DataDB name/number")
	outDataDBUser = flag.String("out_datadb_user", utils.MetaDataDB,
		"output DataDB user")
	outDataDBPass = flag.String("out_datadb_passwd", utils.MetaDataDB,
		"output DataDB password")

	inStorDBType = flag.String("stordb_type", dfltCfg.StorDBType,
		"the type of the StorDB Database <*mysql|*postgres|*mongo>")
	inStorDBHost = flag.String("stordb_host", dfltCfg.StorDBHost,
		"the StorDB host")
	inStorDBPort = flag.String("stordb_port", dfltCfg.StorDBPort,
		"the StorDB port")
	inStorDBName = flag.String("stordb_name", dfltCfg.StorDBName,
		"the name/number of the StorDB")
	inStorDBUser = flag.String("stordb_user", dfltCfg.StorDBUser,
		"the StorDB user")
	inStorDBPass = flag.String("stordb_passwd", dfltCfg.StorDBPass,
		"the StorDB password")

	inDBDataEncoding = flag.String("dbdata_encoding", dfltCfg.DBDataEncoding,
		"the encoding used to store object Data in strings")
	outDBDataEncoding = flag.String("out_dbdata_encoding", "",
		"the encoding used to store object Data in strings in move mode")

	outStorDBType = flag.String("out_stordb_type", utils.MetaStorDB,
		"output StorDB type for move mode <*mysql|*postgres|*mongo>")
	outStorDBHost = flag.String("out_stordb_host", utils.MetaStorDB,
		"output StorDB host")
	outStorDBPort = flag.String("out_stordb_port", utils.MetaStorDB,
		"output StorDB port")
	outStorDBName = flag.String("out_stordb_name", utils.MetaStorDB,
		"output StorDB name/number")
	outStorDBUser = flag.String("out_stordb_user", utils.MetaStorDB,
		"output StorDB user")
	outStorDBPass = flag.String("out_stordb_passwd", utils.MetaStorDB,
		"output StorDB password")

	dryRun = flag.Bool("dry_run", false,
		"parse loaded data for consistency and errors, without storing it")
	verbose = flag.Bool("verbose", false, "enable detailed verbose logging output")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}

	mgrCfg := dfltCfg
	if *cfgDir != "" {
		if mgrCfg, err = config.NewCGRConfigFromFolder(*cfgDir); err != nil {
			log.Fatalf("error loading config file %s", err.Error())
		}
	}

	// in settings
	if *inDataDBType != dfltCfg.DataDbType {
		mgrCfg.DataDbType = strings.TrimPrefix(*inDataDBType, "*")
	}
	if *inDataDBHost != dfltCfg.DataDbHost {
		mgrCfg.DataDbHost = *inDataDBHost
	}
	if *inDataDBPort != dfltCfg.DataDbPort {
		mgrCfg.DataDbPort = *inDataDBPort
	}
	if *inDataDBName != dfltCfg.DataDbName {
		mgrCfg.DataDbName = *inDataDBName
	}
	if *inDataDBUser != dfltCfg.DataDbUser {
		mgrCfg.DataDbUser = *inDataDBUser
	}
	if *inDataDBPass != dfltCfg.DataDbPass {
		mgrCfg.DataDbPass = *inDataDBPass
	}
	if *inStorDBType != dfltCfg.StorDBType {
		mgrCfg.StorDBType = strings.TrimPrefix(*inStorDBType, "*")
	}
	if *inStorDBHost != dfltCfg.StorDBHost {
		mgrCfg.StorDBHost = *inStorDBHost
	}
	if *inStorDBPort != dfltCfg.StorDBPort {
		mgrCfg.StorDBPort = *inStorDBPort
	}
	if *inStorDBName != dfltCfg.StorDBName {
		mgrCfg.StorDBName = *inStorDBName
	}
	if *inStorDBUser != dfltCfg.StorDBUser {
		mgrCfg.StorDBUser = *inStorDBUser
	}
	if *inStorDBPass != "" {
		mgrCfg.StorDBPass = *inStorDBPass
	}
	if *inDBDataEncoding != "" {
		mgrCfg.DBDataEncoding = *inDBDataEncoding
	}

	// out settings
	if *outDataDBType == utils.MetaDataDB {
		*outDataDBType = mgrCfg.DataDbType
	} else {
		*outDataDBType = strings.TrimPrefix(*outDataDBType, "*")
	}
	if *outDataDBHost == utils.MetaDataDB {
		*outDataDBHost = mgrCfg.DataDbHost
	}
	if *outDataDBPort == utils.MetaDataDB {
		*outDataDBPort = mgrCfg.DataDbPort
	}
	if *outDataDBName == utils.MetaDataDB {
		*outDataDBName = mgrCfg.DataDbName
	}
	if *outDataDBUser == utils.MetaDataDB {
		*outDataDBUser = mgrCfg.DataDbUser
	}
	if *outDataDBPass == utils.MetaDataDB {
		*outDataDBPass = mgrCfg.DataDbPass
	}
	if *outStorDBType == utils.MetaStorDB {
		*outStorDBType = mgrCfg.StorDBType
	} else {
		*outStorDBType = strings.TrimPrefix(*outStorDBType, "*")
	}
	if *outStorDBHost == utils.MetaStorDB {
		*outStorDBHost = mgrCfg.StorDBHost
	}
	if *outStorDBPort == utils.MetaStorDB {
		*outStorDBPort = mgrCfg.StorDBPort
	}
	if *outStorDBName == utils.MetaStorDB {
		*outStorDBName = mgrCfg.StorDBName
	}
	if *outStorDBUser == utils.MetaStorDB {
		*outStorDBUser = mgrCfg.StorDBUser
	}
	if *outStorDBPass == utils.MetaStorDB {
		*outStorDBPass = mgrCfg.StorDBPass
	}
	if *outDBDataEncoding == "" {
		*outDBDataEncoding = mgrCfg.DBDataEncoding
	}

	sameDataDB = *outDataDBType == mgrCfg.DataDbType &&
		*outDataDBHost == mgrCfg.DataDbHost &&
		*outDataDBPort == mgrCfg.DataDbPort &&
		*outDataDBName == mgrCfg.DataDbName &&
		*outDBDataEncoding == mgrCfg.DBDataEncoding

	sameStorDB = *outStorDBType == mgrCfg.StorDBType &&
		*outStorDBHost == mgrCfg.StorDBHost &&
		*outStorDBPort == mgrCfg.StorDBPort &&
		*outStorDBName == mgrCfg.StorDBName &&
		*outDBDataEncoding == mgrCfg.DBDataEncoding

	if dmIN, err = migrator.NewMigratorDataDB(mgrCfg.DataDbType,
		mgrCfg.DataDbHost, mgrCfg.DataDbPort,
		mgrCfg.DataDbName, mgrCfg.DataDbUser,
		mgrCfg.DataDbPass, mgrCfg.DBDataEncoding,
		mgrCfg.CacheCfg(), 0); err != nil {
		log.Fatal(err)
	}

	if sameDataDB {
		dmOUT = dmIN
	} else if dmOUT, err = migrator.NewMigratorDataDB(*outDataDBType,
		*outDataDBHost, *outDataDBPort,
		*outDataDBName, *outDataDBUser,
		*outDataDBPass, *outDBDataEncoding,
		mgrCfg.CacheCfg(), 0); err != nil {
		log.Fatal(err)
	}

	if storDBIn, err = migrator.NewMigratorStorDB(*inStorDBType,
		*inStorDBHost, *inStorDBPort,
		*inStorDBName, *inStorDBUser, *inStorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes); err != nil {
		log.Fatal(err)
	}

	if sameStorDB {
		storDBOut = storDBIn
	} else if storDBOut, err = migrator.NewMigratorStorDB(*outStorDBType,
		*outStorDBHost, *outStorDBPort,
		*outStorDBName, *outStorDBUser,
		*outStorDBPass,
		mgrCfg.StorDBMaxOpenConns,
		mgrCfg.StorDBMaxIdleConns,
		mgrCfg.StorDBConnMaxLifetime,
		mgrCfg.StorDBCDRSIndexes); err != nil {
		log.Fatal(err)
	}

	m, err := migrator.NewMigrator(dmIN, dmOUT,
		storDBIn, storDBOut,
		*dryRun, sameDataDB, sameStorDB)
	if err != nil {
		log.Fatal(err)
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
