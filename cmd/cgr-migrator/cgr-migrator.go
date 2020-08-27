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
	"os"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/migrator"
	"github.com/cgrates/cgrates/utils"
)

var (
	cgrMigratorFlags = flag.NewFlagSet("cgr-migrator", flag.ContinueOnError)

	sameDataDB bool
	sameStorDB bool
	sameOutDB  bool
	dmIN       migrator.MigratorDataDB
	dmOUT      migrator.MigratorDataDB
	storDBIn   migrator.MigratorStorDB
	storDBOut  migrator.MigratorStorDB
	err        error
	dfltCfg, _ = config.NewDefaultCGRConfig()
	cfgPath    = cgrMigratorFlags.String("config_path", "",
		"Configuration directory path.")

	exec = cgrMigratorFlags.String("exec", "", "fire up automatic migration "+
		"<*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups|*filters|*stordb|*datadb>")
	version = cgrMigratorFlags.Bool("version", false, "prints the application version")

	inDataDBType = cgrMigratorFlags.String("datadb_type", dfltCfg.DataDbCfg().DataDbType,
		"the type of the DataDB Database <*redis|*mongo>")
	inDataDBHost = cgrMigratorFlags.String("datadb_host", dfltCfg.DataDbCfg().DataDbHost,
		"the DataDB host")
	inDataDBPort = cgrMigratorFlags.String("datadb_port", dfltCfg.DataDbCfg().DataDbPort,
		"the DataDB port")
	inDataDBName = cgrMigratorFlags.String("datadb_name", dfltCfg.DataDbCfg().DataDbName,
		"the name/number of the DataDB")
	inDataDBUser = cgrMigratorFlags.String("datadb_user", dfltCfg.DataDbCfg().DataDbUser,
		"the DataDB user")
	inDataDBPass = cgrMigratorFlags.String("datadb_passwd", dfltCfg.DataDbCfg().DataDbPass,
		"the DataDB password")
	inDBDataEncoding = cgrMigratorFlags.String("dbdata_encoding", dfltCfg.GeneralCfg().DBDataEncoding,
		"the encoding used to store object Data in strings")
	inDataDBRedisSentinel = cgrMigratorFlags.String("redis_sentinel", utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisSentinelNameCfg]),
		"the name of redis sentinel")
	dbRedisCluster = cgrMigratorFlags.Bool("redis_cluster", false,
		"Is the redis datadb a cluster")
	dbRedisClusterSync = cgrMigratorFlags.String("cluster_sync", utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.ClusterSyncCfg]),
		"The sync interval for the redis cluster")
	dbRedisClusterDownDelay = cgrMigratorFlags.String("cluster_ondown_delay", utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.ClusterOnDownDelayCfg]),
		"The delay before executing the commands if the redis cluster is in the CLUSTERDOWN state")
	dbQueryTimeout = cgrMigratorFlags.String("query_timeout", utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.QueryTimeoutCfg]),
		"The timeout for queries")

	outDataDBType = cgrMigratorFlags.String("out_datadb_type", utils.MetaDataDB,
		"output DataDB type <*redis|*mongo>")
	outDataDBHost = cgrMigratorFlags.String("out_datadb_host", utils.MetaDataDB,
		"output DataDB host to connect to")
	outDataDBPort = cgrMigratorFlags.String("out_datadb_port", utils.MetaDataDB,
		"output DataDB port")
	outDataDBName = cgrMigratorFlags.String("out_datadb_name", utils.MetaDataDB,
		"output DataDB name/number")
	outDataDBUser = cgrMigratorFlags.String("out_datadb_user", utils.MetaDataDB,
		"output DataDB user")
	outDataDBPass = cgrMigratorFlags.String("out_datadb_passwd", utils.MetaDataDB,
		"output DataDB password")
	outDBDataEncoding = cgrMigratorFlags.String("out_dbdata_encoding", utils.MetaDataDB,
		"the encoding used to store object Data in strings in move mode")
	outDataDBRedisSentinel = cgrMigratorFlags.String("out_redis_sentinel", utils.MetaDataDB,
		"the name of redis sentinel")

	inStorDBType = cgrMigratorFlags.String("stordb_type", dfltCfg.StorDbCfg().Type,
		"the type of the StorDB Database <*mysql|*postgres|*mongo>")
	inStorDBHost = cgrMigratorFlags.String("stordb_host", dfltCfg.StorDbCfg().Host,
		"the StorDB host")
	inStorDBPort = cgrMigratorFlags.String("stordb_port", dfltCfg.StorDbCfg().Port,
		"the StorDB port")
	inStorDBName = cgrMigratorFlags.String("stordb_name", dfltCfg.StorDbCfg().Name,
		"the name/number of the StorDB")
	inStorDBUser = cgrMigratorFlags.String("stordb_user", dfltCfg.StorDbCfg().User,
		"the StorDB user")
	inStorDBPass = cgrMigratorFlags.String("stordb_passwd", dfltCfg.StorDbCfg().Password,
		"the StorDB password")

	outStorDBType = cgrMigratorFlags.String("out_stordb_type", utils.MetaStorDB,
		"output StorDB type for move mode <*mysql|*postgres|*mongo>")
	outStorDBHost = cgrMigratorFlags.String("out_stordb_host", utils.MetaStorDB,
		"output StorDB host")
	outStorDBPort = cgrMigratorFlags.String("out_stordb_port", utils.MetaStorDB,
		"output StorDB port")
	outStorDBName = cgrMigratorFlags.String("out_stordb_name", utils.MetaStorDB,
		"output StorDB name/number")
	outStorDBUser = cgrMigratorFlags.String("out_stordb_user", utils.MetaStorDB,
		"output StorDB user")
	outStorDBPass = cgrMigratorFlags.String("out_stordb_passwd", utils.MetaStorDB,
		"output StorDB password")

	dryRun = cgrMigratorFlags.Bool("dry_run", false,
		"parse loaded data for consistency and errors, without storing it")
	verbose = cgrMigratorFlags.Bool("verbose", false, "enable detailed verbose logging output")
)

func main() {
	if err := cgrMigratorFlags.Parse(os.Args[1:]); err != nil {
		return
	}
	if *version {
		if rcv, err := utils.GetCGRVersion(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(rcv)
		}
		return
	}

	mgrCfg := dfltCfg
	if *cfgPath != "" {
		if mgrCfg, err = config.NewCGRConfigFromPath(*cfgPath); err != nil {
			log.Fatalf("error loading config file %s", err.Error())
		}
		config.SetCgrConfig(mgrCfg)
	}

	// inDataDB
	if *inDataDBType != dfltCfg.DataDbCfg().DataDbType {
		mgrCfg.DataDbCfg().DataDbType = strings.TrimPrefix(*inDataDBType, "*")
	}
	if *inDataDBHost != dfltCfg.DataDbCfg().DataDbHost {
		mgrCfg.DataDbCfg().DataDbHost = *inDataDBHost
	}
	if *inDataDBPort != dfltCfg.DataDbCfg().DataDbPort {
		mgrCfg.DataDbCfg().DataDbPort = *inDataDBPort
	}
	if *inDataDBName != dfltCfg.DataDbCfg().DataDbName {
		mgrCfg.DataDbCfg().DataDbName = *inDataDBName
	}
	if *inDataDBUser != dfltCfg.DataDbCfg().DataDbUser {
		mgrCfg.DataDbCfg().DataDbUser = *inDataDBUser
	}
	if *inDataDBPass != dfltCfg.DataDbCfg().DataDbPass {
		mgrCfg.DataDbCfg().DataDbPass = *inDataDBPass
	}
	if *inDBDataEncoding != dfltCfg.GeneralCfg().DBDataEncoding {
		mgrCfg.GeneralCfg().DBDataEncoding = *inDBDataEncoding
	}
	if *inDataDBRedisSentinel != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisSentinelNameCfg]) {
		mgrCfg.DataDbCfg().Opts[utils.RedisSentinelNameCfg] = *inDataDBRedisSentinel
	}
	rdsCls, _ := utils.IfaceAsBool(dfltCfg.DataDbCfg().Opts[utils.RedisClusterCfg])
	if *dbRedisCluster != rdsCls {
		mgrCfg.DataDbCfg().Opts[utils.RedisClusterCfg] = *dbRedisCluster
	}
	if *dbRedisClusterSync != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.ClusterSyncCfg]) {
		mgrCfg.DataDbCfg().Opts[utils.ClusterSyncCfg] = *dbRedisClusterSync
	}
	if *dbRedisClusterDownDelay != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.ClusterOnDownDelayCfg]) {
		mgrCfg.DataDbCfg().Opts[utils.ClusterOnDownDelayCfg] = *dbRedisClusterDownDelay
	}
	if *dbQueryTimeout != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.QueryTimeoutCfg]) {
		mgrCfg.DataDbCfg().Opts[utils.QueryTimeoutCfg] = *dbQueryTimeout
	}

	// outDataDB
	if *outDataDBType == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBType == mgrCfg.MigratorCgrCfg().OutDataDBType {
			mgrCfg.MigratorCgrCfg().OutDataDBType = mgrCfg.DataDbCfg().DataDbType
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBType = strings.TrimPrefix(*outDataDBType, "*")
	}

	if *outDataDBHost == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBHost == mgrCfg.MigratorCgrCfg().OutDataDBHost {
			mgrCfg.MigratorCgrCfg().OutDataDBHost = mgrCfg.DataDbCfg().DataDbHost
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBHost = *outDataDBHost
	}
	if *outDataDBPort == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBPort == mgrCfg.MigratorCgrCfg().OutDataDBPort {
			mgrCfg.MigratorCgrCfg().OutDataDBPort = mgrCfg.DataDbCfg().DataDbPort
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBPort = *outDataDBPort
	}
	if *outDataDBName == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBName == mgrCfg.MigratorCgrCfg().OutDataDBName {
			mgrCfg.MigratorCgrCfg().OutDataDBName = mgrCfg.DataDbCfg().DataDbName
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBName = *outDataDBName
	}
	if *outDataDBUser == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBUser == mgrCfg.MigratorCgrCfg().OutDataDBUser {
			mgrCfg.MigratorCgrCfg().OutDataDBUser = mgrCfg.DataDbCfg().DataDbUser
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBUser = *outDataDBUser
	}
	if *outDataDBPass == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBPassword == mgrCfg.MigratorCgrCfg().OutDataDBPassword {
			mgrCfg.MigratorCgrCfg().OutDataDBPassword = mgrCfg.DataDbCfg().DataDbPass
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBPassword = *outDataDBPass
	}
	if *outDBDataEncoding == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBEncoding == mgrCfg.MigratorCgrCfg().OutDataDBEncoding {
			mgrCfg.MigratorCgrCfg().OutDataDBEncoding = mgrCfg.GeneralCfg().DBDataEncoding
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBEncoding = *outDBDataEncoding
	}
	if *outDataDBRedisSentinel == utils.MetaDataDB {
		if utils.IfaceAsString(dfltCfg.MigratorCgrCfg().OutDataDBOpts[utils.RedisSentinelNameCfg]) == utils.IfaceAsString(mgrCfg.MigratorCgrCfg().OutDataDBOpts[utils.RedisSentinelNameCfg]) {
			mgrCfg.MigratorCgrCfg().OutDataDBOpts[utils.RedisSentinelNameCfg] = dfltCfg.DataDbCfg().Opts[utils.RedisSentinelNameCfg]
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBOpts[utils.RedisSentinelNameCfg] = *outDataDBRedisSentinel
	}

	sameDataDB = mgrCfg.MigratorCgrCfg().OutDataDBType == mgrCfg.DataDbCfg().DataDbType &&
		mgrCfg.MigratorCgrCfg().OutDataDBHost == mgrCfg.DataDbCfg().DataDbHost &&
		mgrCfg.MigratorCgrCfg().OutDataDBPort == mgrCfg.DataDbCfg().DataDbPort &&
		mgrCfg.MigratorCgrCfg().OutDataDBName == mgrCfg.DataDbCfg().DataDbName &&
		mgrCfg.MigratorCgrCfg().OutDataDBEncoding == mgrCfg.GeneralCfg().DBDataEncoding

	if dmIN, err = migrator.NewMigratorDataDB(mgrCfg.DataDbCfg().DataDbType,
		mgrCfg.DataDbCfg().DataDbHost, mgrCfg.DataDbCfg().DataDbPort,
		mgrCfg.DataDbCfg().DataDbName, mgrCfg.DataDbCfg().DataDbUser,
		mgrCfg.DataDbCfg().DataDbPass, mgrCfg.GeneralCfg().DBDataEncoding,
		mgrCfg.CacheCfg(), mgrCfg.DataDbCfg().Opts); err != nil {
		log.Fatal(err)
	}

	if sameDataDB {
		dmOUT = dmIN
	} else if dmOUT, err = migrator.NewMigratorDataDB(mgrCfg.MigratorCgrCfg().OutDataDBType,
		mgrCfg.MigratorCgrCfg().OutDataDBHost, mgrCfg.MigratorCgrCfg().OutDataDBPort,
		mgrCfg.MigratorCgrCfg().OutDataDBName, mgrCfg.MigratorCgrCfg().OutDataDBUser,
		mgrCfg.MigratorCgrCfg().OutDataDBPassword, mgrCfg.MigratorCgrCfg().OutDataDBEncoding,
		mgrCfg.CacheCfg(), mgrCfg.MigratorCgrCfg().OutDataDBOpts); err != nil {
		log.Fatal(err)
	}

	// inStorDB
	if *inStorDBType != dfltCfg.StorDbCfg().Type {
		mgrCfg.StorDbCfg().Type = strings.TrimPrefix(*inStorDBType, "*")
	}
	if *inStorDBHost != dfltCfg.StorDbCfg().Host {
		mgrCfg.StorDbCfg().Host = *inStorDBHost
	}
	if *inStorDBPort != dfltCfg.StorDbCfg().Port {
		mgrCfg.StorDbCfg().Port = *inStorDBPort
	}
	if *inStorDBName != dfltCfg.StorDbCfg().Name {
		mgrCfg.StorDbCfg().Name = *inStorDBName
	}
	if *inStorDBUser != dfltCfg.StorDbCfg().User {
		mgrCfg.StorDbCfg().User = *inStorDBUser
	}
	if *inStorDBPass != dfltCfg.StorDbCfg().Password {
		mgrCfg.StorDbCfg().Password = *inStorDBPass
	}

	// outStorDB
	if *outStorDBType == utils.MetaStorDB {
		if dfltCfg.MigratorCgrCfg().OutStorDBType == mgrCfg.MigratorCgrCfg().OutStorDBType {
			mgrCfg.MigratorCgrCfg().OutStorDBType = mgrCfg.StorDbCfg().Type
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutStorDBType = strings.TrimPrefix(*outStorDBType, "*")
	}
	if *outStorDBHost == utils.MetaStorDB {
		if dfltCfg.MigratorCgrCfg().OutStorDBHost == mgrCfg.MigratorCgrCfg().OutStorDBHost {
			mgrCfg.MigratorCgrCfg().OutStorDBHost = mgrCfg.StorDbCfg().Host
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutStorDBHost = *outStorDBHost
	}
	if *outStorDBPort == utils.MetaStorDB {
		if dfltCfg.MigratorCgrCfg().OutStorDBPort == mgrCfg.MigratorCgrCfg().OutStorDBPort {
			mgrCfg.MigratorCgrCfg().OutStorDBPort = mgrCfg.StorDbCfg().Port
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutStorDBPort = *outStorDBPort
	}
	if *outStorDBName == utils.MetaStorDB {
		if dfltCfg.MigratorCgrCfg().OutStorDBName == mgrCfg.MigratorCgrCfg().OutStorDBName {
			mgrCfg.MigratorCgrCfg().OutStorDBName = mgrCfg.StorDbCfg().Name
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutStorDBName = *outStorDBName
	}
	if *outStorDBUser == utils.MetaStorDB {
		if dfltCfg.MigratorCgrCfg().OutStorDBUser == mgrCfg.MigratorCgrCfg().OutStorDBUser {
			mgrCfg.MigratorCgrCfg().OutStorDBUser = mgrCfg.StorDbCfg().User
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutStorDBUser = *outStorDBUser
	}
	if *outStorDBPass == utils.MetaStorDB {
		if dfltCfg.MigratorCgrCfg().OutStorDBPassword == mgrCfg.MigratorCgrCfg().OutStorDBPassword {
			mgrCfg.MigratorCgrCfg().OutStorDBPassword = mgrCfg.StorDbCfg().Password
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutStorDBPassword = *outStorDBPass
	}

	sameStorDB = mgrCfg.MigratorCgrCfg().OutStorDBType == mgrCfg.StorDbCfg().Type &&
		mgrCfg.MigratorCgrCfg().OutStorDBHost == mgrCfg.StorDbCfg().Host &&
		mgrCfg.MigratorCgrCfg().OutStorDBPort == mgrCfg.StorDbCfg().Port &&
		mgrCfg.MigratorCgrCfg().OutStorDBName == mgrCfg.StorDbCfg().Name

	if storDBIn, err = migrator.NewMigratorStorDB(mgrCfg.StorDbCfg().Type,
		mgrCfg.StorDbCfg().Host, mgrCfg.StorDbCfg().Port,
		mgrCfg.StorDbCfg().Name, mgrCfg.StorDbCfg().User,
		mgrCfg.StorDbCfg().Password, mgrCfg.GeneralCfg().DBDataEncoding,
		mgrCfg.StorDbCfg().StringIndexedFields, mgrCfg.StorDbCfg().PrefixIndexedFields,
		mgrCfg.StorDbCfg().Opts); err != nil {
		log.Fatal(err)
	}

	if sameStorDB {
		storDBOut = storDBIn
	} else if storDBOut, err = migrator.NewMigratorStorDB(mgrCfg.MigratorCgrCfg().OutStorDBType,
		mgrCfg.MigratorCgrCfg().OutStorDBHost, mgrCfg.MigratorCgrCfg().OutStorDBPort,
		mgrCfg.MigratorCgrCfg().OutStorDBName, mgrCfg.MigratorCgrCfg().OutStorDBUser,
		mgrCfg.MigratorCgrCfg().OutStorDBPassword, mgrCfg.GeneralCfg().DBDataEncoding,
		mgrCfg.StorDbCfg().StringIndexedFields, mgrCfg.StorDbCfg().PrefixIndexedFields,
		mgrCfg.MigratorCgrCfg().OutStorDBOpts); err != nil {
		log.Fatal(err)
	}

	sameOutDB = mgrCfg.MigratorCgrCfg().OutStorDBType == mgrCfg.MigratorCgrCfg().OutDataDBType &&
		mgrCfg.MigratorCgrCfg().OutStorDBHost == mgrCfg.MigratorCgrCfg().OutDataDBHost &&
		mgrCfg.MigratorCgrCfg().OutStorDBPort == mgrCfg.MigratorCgrCfg().OutDataDBPort &&
		mgrCfg.MigratorCgrCfg().OutStorDBName == mgrCfg.MigratorCgrCfg().OutDataDBName

	m, err := migrator.NewMigrator(dmIN, dmOUT,
		storDBIn, storDBOut,
		*dryRun, sameDataDB, sameStorDB, sameOutDB)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()
	config.SetCgrConfig(mgrCfg)
	if exec != nil && *exec != "" { // Run migrator
		migrstats := make(map[string]int)
		mig := strings.Split(*exec, ",")
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
