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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/migrator"
	"github.com/cgrates/cgrates/utils"
)

var (
	cgrMigratorFlags = flag.NewFlagSet(utils.CgrMigrator, flag.ContinueOnError)

	sameDataDB bool
	sameStorDB bool
	sameOutDB  bool
	dmIN       migrator.MigratorDataDB
	dmOUT      migrator.MigratorDataDB
	storDBIn   migrator.MigratorStorDB
	storDBOut  migrator.MigratorStorDB
	err        error
	dfltCfg    = config.NewDefaultCGRConfig()
	cfgPath    = cgrMigratorFlags.String(utils.CfgPathCgr, utils.EmptyString,
		"Configuration directory path.")

	exec = cgrMigratorFlags.String(utils.ExecCgr, utils.EmptyString, "fire up automatic migration "+
		"<*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups|*filters|*stordb|*datadb>")
	version = cgrMigratorFlags.Bool(utils.VersionCgr, false, "prints the application version")

	inDataDBType = cgrMigratorFlags.String(utils.DataDBTypeCgr, dfltCfg.DataDbCfg().Type,
		"the type of the DataDB Database <*redis|*mongo>")
	inDataDBHost = cgrMigratorFlags.String(utils.DataDBHostCgr, dfltCfg.DataDbCfg().Host,
		"the DataDB host")
	inDataDBPort = cgrMigratorFlags.String(utils.DataDBPortCgr, dfltCfg.DataDbCfg().Port,
		"the DataDB port")
	inDataDBName = cgrMigratorFlags.String(utils.DataDBNameCgr, dfltCfg.DataDbCfg().Name,
		"the name/number of the DataDB")
	inDataDBUser = cgrMigratorFlags.String(utils.DataDBUserCgr, dfltCfg.DataDbCfg().User,
		"the DataDB user")
	inDataDBPass = cgrMigratorFlags.String(utils.DataDBPasswdCgr, dfltCfg.DataDbCfg().Password,
		"the DataDB password")
	inDBDataEncoding = cgrMigratorFlags.String(utils.DBDataEncodingCfg, dfltCfg.GeneralCfg().DBDataEncoding,
		"the encoding used to store object Data in strings")
	inDataDBRedisSentinel = cgrMigratorFlags.String(utils.RedisSentinelNameCfg, utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisSentinelNameCfg]),
		"the name of redis sentinel")
	dbRedisCluster = cgrMigratorFlags.Bool(utils.RedisClusterCfg, false,
		"Is the redis datadb a cluster")
	dbRedisClusterSync = cgrMigratorFlags.String(utils.RedisClusterSyncCfg, utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClusterSyncCfg]),
		"The sync interval for the redis cluster")
	dbRedisClusterDownDelay = cgrMigratorFlags.String(utils.RedisClusterOnDownDelayCfg, utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClusterOnDownDelayCfg]),
		"The delay before executing the commands if the redis cluster is in the CLUSTERDOWN state")
	dbQueryTimeout = cgrMigratorFlags.String(utils.MongoQueryTimeoutCfg, utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.MongoQueryTimeoutCfg]),
		"The timeout for queries")
	dbRedisTls               = cgrMigratorFlags.Bool(utils.RedisTLS, false, "Enable TLS when connecting to Redis")
	dbRedisClientCertificate = cgrMigratorFlags.String(utils.RedisClientCertificate, utils.EmptyString, "Path to the client certificate")
	dbRedisClientKey         = cgrMigratorFlags.String(utils.RedisClientKey, utils.EmptyString, "Path to the client key")
	dbRedisCACertificate     = cgrMigratorFlags.String(utils.RedisCACertificate, utils.EmptyString, "Path to the CA certificate")

	outDataDBType = cgrMigratorFlags.String(utils.OutDataDBTypeCfg, utils.MetaDataDB,
		"output DataDB type <*redis|*mongo>")
	outDataDBHost = cgrMigratorFlags.String(utils.OutDataDBHostCfg, utils.MetaDataDB,
		"output DataDB host to connect to")
	outDataDBPort = cgrMigratorFlags.String(utils.OutDataDBPortCfg, utils.MetaDataDB,
		"output DataDB port")
	outDataDBName = cgrMigratorFlags.String(utils.OutDataDBNameCfg, utils.MetaDataDB,
		"output DataDB name/number")
	outDataDBUser = cgrMigratorFlags.String(utils.OutDataDBUserCfg, utils.MetaDataDB,
		"output DataDB user")
	outDataDBPass = cgrMigratorFlags.String(utils.OutDataDBPasswordCfg, utils.MetaDataDB,
		"output DataDB password")
	outDBDataEncoding = cgrMigratorFlags.String(utils.OutDataDBEncodingCfg, utils.MetaDataDB,
		"the encoding used to store object Data in strings in move mode")
	outDataDBRedisSentinel = cgrMigratorFlags.String(utils.OutDataDBRedisSentinel, utils.MetaDataDB,
		"the name of redis sentinel")

	inStorDBType = cgrMigratorFlags.String(utils.StorDBTypeCgr, dfltCfg.StorDbCfg().Type,
		"the type of the StorDB Database <*mysql|*postgres|*mongo>")
	inStorDBHost = cgrMigratorFlags.String(utils.StorDBHostCgr, dfltCfg.StorDbCfg().Host,
		"the StorDB host")
	inStorDBPort = cgrMigratorFlags.String(utils.StorDBPortCgr, dfltCfg.StorDbCfg().Port,
		"the StorDB port")
	inStorDBName = cgrMigratorFlags.String(utils.StorDBNameCgr, dfltCfg.StorDbCfg().Name,
		"the name/number of the StorDB")
	inStorDBUser = cgrMigratorFlags.String(utils.StorDBUserCgr, dfltCfg.StorDbCfg().User,
		"the StorDB user")
	inStorDBPass = cgrMigratorFlags.String(utils.StorDBPasswdCgr, dfltCfg.StorDbCfg().Password,
		"the StorDB password")

	outStorDBType = cgrMigratorFlags.String(utils.OutStorDBTypeCfg, utils.MetaStorDB,
		"output StorDB type for move mode <*mysql|*postgres|*mongo>")
	outStorDBHost = cgrMigratorFlags.String(utils.OutStorDBHostCfg, utils.MetaStorDB,
		"output StorDB host")
	outStorDBPort = cgrMigratorFlags.String(utils.OutStorDBPortCfg, utils.MetaStorDB,
		"output StorDB port")
	outStorDBName = cgrMigratorFlags.String(utils.OutStorDBNameCfg, utils.MetaStorDB,
		"output StorDB name/number")
	outStorDBUser = cgrMigratorFlags.String(utils.OutStorDBUserCfg, utils.MetaStorDB,
		"output StorDB user")
	outStorDBPass = cgrMigratorFlags.String(utils.OutStorDBPasswordCfg, utils.MetaStorDB,
		"output StorDB password")

	dryRun = cgrMigratorFlags.Bool(utils.DryRunCfg, false,
		"parse loaded data for consistency and errors, without storing it")
	verbose = cgrMigratorFlags.Bool(utils.VerboseCgr, false, "enable detailed verbose logging output")
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
	if *cfgPath != utils.EmptyString {
		if mgrCfg, err = config.NewCGRConfigFromPath(context.Background(), *cfgPath); err != nil {
			log.Fatalf("error loading config file %s", err.Error())
		}
		if mgrCfg.ConfigDBCfg().Type != utils.MetaInternal {
			d, err := engine.NewDataDBConn(mgrCfg.ConfigDBCfg().Type,
				mgrCfg.ConfigDBCfg().Host, mgrCfg.ConfigDBCfg().Port,
				mgrCfg.ConfigDBCfg().Name, mgrCfg.ConfigDBCfg().User,
				mgrCfg.ConfigDBCfg().Password, mgrCfg.GeneralCfg().DBDataEncoding,
				mgrCfg.ConfigDBCfg().Opts)
			if err != nil { // Cannot configure getter database, show stopper
				utils.Logger.Crit(fmt.Sprintf("Could not configure configDB: %s exiting!", err))
				return
			}
			if err = mgrCfg.LoadFromDB(context.Background(), d); err != nil {
				log.Fatalf("Could not parse config: <%s>", err.Error())
				return
			}
		}
		config.SetCgrConfig(mgrCfg)
	}

	// inDataDB
	if *inDataDBType != dfltCfg.DataDbCfg().Type {
		mgrCfg.DataDbCfg().Type = strings.TrimPrefix(*inDataDBType, utils.MaskChar)
	}
	if *inDataDBHost != dfltCfg.DataDbCfg().Host {
		mgrCfg.DataDbCfg().Host = *inDataDBHost
	}
	if *inDataDBPort != dfltCfg.DataDbCfg().Port {
		mgrCfg.DataDbCfg().Port = *inDataDBPort
	}
	if *inDataDBName != dfltCfg.DataDbCfg().Name {
		mgrCfg.DataDbCfg().Name = *inDataDBName
	}
	if *inDataDBUser != dfltCfg.DataDbCfg().User {
		mgrCfg.DataDbCfg().User = *inDataDBUser
	}
	if *inDataDBPass != dfltCfg.DataDbCfg().Password {
		mgrCfg.DataDbCfg().Password = *inDataDBPass
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
	if *dbRedisClusterSync != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClusterSyncCfg]) {
		mgrCfg.DataDbCfg().Opts[utils.RedisClusterSyncCfg] = *dbRedisClusterSync
	}
	if *dbRedisClusterDownDelay != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClusterOnDownDelayCfg]) {
		mgrCfg.DataDbCfg().Opts[utils.RedisClusterOnDownDelayCfg] = *dbRedisClusterDownDelay
	}
	if *dbQueryTimeout != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.MongoQueryTimeoutCfg]) {
		mgrCfg.DataDbCfg().Opts[utils.MongoQueryTimeoutCfg] = *dbQueryTimeout
	}

	rdsTLS, _ := utils.IfaceAsBool(dfltCfg.DataDbCfg().Opts[utils.RedisTLS])
	if *dbRedisTls != rdsTLS {
		mgrCfg.DataDbCfg().Opts[utils.RedisTLS] = *dbRedisTls
	}
	if *dbRedisClientCertificate != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClientCertificate]) {
		mgrCfg.DataDbCfg().Opts[utils.RedisClientCertificate] = *dbRedisClientCertificate
	}
	if *dbRedisClientKey != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClientKey]) {
		mgrCfg.DataDbCfg().Opts[utils.RedisClientKey] = *dbRedisClientKey
	}
	if *dbRedisCACertificate != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisCACertificate]) {
		mgrCfg.DataDbCfg().Opts[utils.RedisCACertificate] = *dbRedisCACertificate
	}

	// outDataDB
	if *outDataDBType == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBType == mgrCfg.MigratorCgrCfg().OutDataDBType {
			mgrCfg.MigratorCgrCfg().OutDataDBType = mgrCfg.DataDbCfg().Type
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBType = strings.TrimPrefix(*outDataDBType, utils.MaskChar)
	}

	if *outDataDBHost == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBHost == mgrCfg.MigratorCgrCfg().OutDataDBHost {
			mgrCfg.MigratorCgrCfg().OutDataDBHost = mgrCfg.DataDbCfg().Host
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBHost = *outDataDBHost
	}
	if *outDataDBPort == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBPort == mgrCfg.MigratorCgrCfg().OutDataDBPort {
			mgrCfg.MigratorCgrCfg().OutDataDBPort = mgrCfg.DataDbCfg().Port
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBPort = *outDataDBPort
	}
	if *outDataDBName == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBName == mgrCfg.MigratorCgrCfg().OutDataDBName {
			mgrCfg.MigratorCgrCfg().OutDataDBName = mgrCfg.DataDbCfg().Name
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBName = *outDataDBName
	}
	if *outDataDBUser == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBUser == mgrCfg.MigratorCgrCfg().OutDataDBUser {
			mgrCfg.MigratorCgrCfg().OutDataDBUser = mgrCfg.DataDbCfg().User
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBUser = *outDataDBUser
	}
	if *outDataDBPass == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBPassword == mgrCfg.MigratorCgrCfg().OutDataDBPassword {
			mgrCfg.MigratorCgrCfg().OutDataDBPassword = mgrCfg.DataDbCfg().Password
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

	sameDataDB = mgrCfg.MigratorCgrCfg().OutDataDBType == mgrCfg.DataDbCfg().Type &&
		mgrCfg.MigratorCgrCfg().OutDataDBHost == mgrCfg.DataDbCfg().Host &&
		mgrCfg.MigratorCgrCfg().OutDataDBPort == mgrCfg.DataDbCfg().Port &&
		mgrCfg.MigratorCgrCfg().OutDataDBName == mgrCfg.DataDbCfg().Name &&
		mgrCfg.MigratorCgrCfg().OutDataDBEncoding == mgrCfg.GeneralCfg().DBDataEncoding

	if dmIN, err = migrator.NewMigratorDataDB(mgrCfg.DataDbCfg().Type,
		mgrCfg.DataDbCfg().Host, mgrCfg.DataDbCfg().Port,
		mgrCfg.DataDbCfg().Name, mgrCfg.DataDbCfg().User,
		mgrCfg.DataDbCfg().Password, mgrCfg.GeneralCfg().DBDataEncoding,
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
		mgrCfg.StorDbCfg().Type = strings.TrimPrefix(*inStorDBType, utils.MaskChar)
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
		mgrCfg.MigratorCgrCfg().OutStorDBType = strings.TrimPrefix(*outStorDBType, utils.MaskChar)
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
	if exec != nil && *exec != utils.EmptyString { // Run migrator
		migrstats := make(map[string]int)
		mig := strings.Split(*exec, utils.FieldsSep)
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
