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
	printConfig = cgrMigratorFlags.Bool(utils.PrintCfgCgr, false, "Print the configuration object in JSON format")
	exec        = cgrMigratorFlags.String(utils.ExecCgr, utils.EmptyString, "fire up automatic migration "+
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
	dbRedisMaxConns = cgrMigratorFlags.Int(utils.RedisMaxConnsCfg, dfltCfg.DataDbCfg().Opts.RedisMaxConns,
		"The connection pool size")
	dbRedisConnectAttempts = cgrMigratorFlags.Int(utils.RedisConnectAttemptsCfg, dfltCfg.DataDbCfg().Opts.RedisConnectAttempts,
		"The maximum amount of dial attempts")
	inDataDBRedisSentinel = cgrMigratorFlags.String(utils.RedisSentinelNameCfg, dfltCfg.DataDbCfg().Opts.RedisSentinel,
		"the name of redis sentinel")
	dbRedisCluster = cgrMigratorFlags.Bool(utils.RedisClusterCfg, false,
		"Is the redis datadb a cluster")
	dbRedisClusterSync = cgrMigratorFlags.Duration(utils.RedisClusterSyncCfg, dfltCfg.DataDbCfg().Opts.RedisClusterSync,
		"The sync interval for the redis cluster")
	dbRedisClusterDownDelay = cgrMigratorFlags.Duration(utils.RedisClusterOnDownDelayCfg, dfltCfg.DataDbCfg().Opts.RedisClusterOndownDelay,
		"The delay before executing the commands if the redis cluster is in the CLUSTERDOWN state")
	dbRedisConnectTimeout = cgrMigratorFlags.Duration(utils.RedisConnectTimeoutCfg, dfltCfg.DataDbCfg().Opts.RedisConnectTimeout,
		"The amount of wait time until timeout for a connection attempt")
	dbRedisReadTimeout = cgrMigratorFlags.Duration(utils.RedisReadTimeoutCfg, dfltCfg.DataDbCfg().Opts.RedisReadTimeout,
		"The amount of wait time until timeout for reading operations")
	dbRedisWriteTimeout = cgrMigratorFlags.Duration(utils.RedisWriteTimeoutCfg, dfltCfg.DataDbCfg().Opts.RedisWriteTimeout,
		"The amount of wait time until timeout for writing operations")
	dbQueryTimeout = cgrMigratorFlags.Duration(utils.MongoQueryTimeoutCfg, dfltCfg.DataDbCfg().Opts.MongoQueryTimeout,
		"The timeout for queries")
	dbMongoConnScheme = cgrMigratorFlags.String(utils.MongoConnSchemeCfg, dfltCfg.DataDbCfg().Opts.MongoConnScheme,
		"Scheme for MongoDB connection <mongodb|mongodb+srv>")
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
		if mgrCfg, err = config.NewCGRConfigFromPath(*cfgPath); err != nil {
			log.Fatalf("error loading config file %s", err.Error())
		}
		config.SetCgrConfig(mgrCfg)
	}

	// inDataDB
	if *inDataDBType != dfltCfg.DataDbCfg().Type {
		mgrCfg.DataDbCfg().Type = *inDataDBType
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
	if *dbRedisMaxConns != dfltCfg.DataDbCfg().Opts.RedisMaxConns {
		mgrCfg.DataDbCfg().Opts.RedisMaxConns = *dbRedisMaxConns
	}
	if *dbRedisConnectAttempts != dfltCfg.DataDbCfg().Opts.RedisConnectAttempts {
		mgrCfg.DataDbCfg().Opts.RedisConnectAttempts = *dbRedisConnectAttempts
	}
	if *inDataDBRedisSentinel != dfltCfg.DataDbCfg().Opts.RedisSentinel {
		mgrCfg.DataDbCfg().Opts.RedisSentinel = *inDataDBRedisSentinel
	}
	if *dbRedisCluster != dfltCfg.DataDbCfg().Opts.RedisCluster {
		mgrCfg.DataDbCfg().Opts.RedisCluster = *dbRedisCluster
	}
	if *dbRedisClusterSync != dfltCfg.DataDbCfg().Opts.RedisClusterSync {
		mgrCfg.DataDbCfg().Opts.RedisClusterSync = *dbRedisClusterSync
	}
	if *dbRedisClusterDownDelay != dfltCfg.DataDbCfg().Opts.RedisClusterOndownDelay {
		mgrCfg.DataDbCfg().Opts.RedisClusterOndownDelay = *dbRedisClusterDownDelay
	}
	if *dbRedisConnectTimeout != dfltCfg.DataDbCfg().Opts.RedisConnectTimeout {
		mgrCfg.DataDbCfg().Opts.RedisConnectTimeout = *dbRedisConnectTimeout
	}
	if *dbRedisReadTimeout != dfltCfg.DataDbCfg().Opts.RedisReadTimeout {
		mgrCfg.DataDbCfg().Opts.RedisReadTimeout = *dbRedisReadTimeout
	}
	if *dbRedisWriteTimeout != dfltCfg.DataDbCfg().Opts.RedisWriteTimeout {
		mgrCfg.DataDbCfg().Opts.RedisWriteTimeout = *dbRedisWriteTimeout
	}
	if *dbQueryTimeout != dfltCfg.DataDbCfg().Opts.MongoQueryTimeout {
		mgrCfg.DataDbCfg().Opts.MongoQueryTimeout = *dbQueryTimeout
	}
	if *dbMongoConnScheme != dfltCfg.DataDbCfg().Opts.MongoConnScheme {
		mgrCfg.DataDbCfg().Opts.MongoConnScheme = *dbMongoConnScheme
	}
	if *dbRedisTls != dfltCfg.DataDbCfg().Opts.RedisTLS {
		mgrCfg.DataDbCfg().Opts.RedisTLS = *dbRedisTls
	}
	if *dbRedisClientCertificate != dfltCfg.DataDbCfg().Opts.RedisClientCertificate {
		mgrCfg.DataDbCfg().Opts.RedisClientCertificate = *dbRedisClientCertificate
	}
	if *dbRedisClientKey != dfltCfg.DataDbCfg().Opts.RedisClientKey {
		mgrCfg.DataDbCfg().Opts.RedisClientKey = *dbRedisClientKey
	}
	if *dbRedisCACertificate != dfltCfg.DataDbCfg().Opts.RedisCACertificate {
		mgrCfg.DataDbCfg().Opts.RedisCACertificate = *dbRedisCACertificate
	}

	// outDataDB
	if *outDataDBType == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDataDBType == mgrCfg.MigratorCgrCfg().OutDataDBType {
			mgrCfg.MigratorCgrCfg().OutDataDBType = mgrCfg.DataDbCfg().Type
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBType = *outDataDBType
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
		if dfltCfg.MigratorCgrCfg().OutDataDBOpts.RedisSentinel == mgrCfg.MigratorCgrCfg().OutDataDBOpts.RedisSentinel {
			mgrCfg.MigratorCgrCfg().OutDataDBOpts.RedisSentinel = dfltCfg.DataDbCfg().Opts.RedisSentinel
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDataDBOpts.RedisSentinel = *outDataDBRedisSentinel
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
		mgrCfg.CacheCfg(), mgrCfg.DataDbCfg().Opts, mgrCfg.DataDbCfg().Items); err != nil {
		log.Fatal(err)
	}

	if sameDataDB {
		dmOUT = dmIN
	} else if dmOUT, err = migrator.NewMigratorDataDB(mgrCfg.MigratorCgrCfg().OutDataDBType,
		mgrCfg.MigratorCgrCfg().OutDataDBHost, mgrCfg.MigratorCgrCfg().OutDataDBPort,
		mgrCfg.MigratorCgrCfg().OutDataDBName, mgrCfg.MigratorCgrCfg().OutDataDBUser,
		mgrCfg.MigratorCgrCfg().OutDataDBPassword, mgrCfg.MigratorCgrCfg().OutDataDBEncoding,
		mgrCfg.CacheCfg(), mgrCfg.MigratorCgrCfg().OutDataDBOpts, mgrCfg.DataDbCfg().Items); err != nil {
		log.Fatal(err)
	}

	// inStorDB
	if *inStorDBType != dfltCfg.StorDbCfg().Type {
		mgrCfg.StorDbCfg().Type = *inStorDBType
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
		mgrCfg.MigratorCgrCfg().OutStorDBType = *outStorDBType
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
		mgrCfg.StorDbCfg().Opts, mgrCfg.StorDbCfg().Items); err != nil {
		log.Fatal(err)
	}

	if sameStorDB {
		storDBOut = storDBIn
	} else if storDBOut, err = migrator.NewMigratorStorDB(mgrCfg.MigratorCgrCfg().OutStorDBType,
		mgrCfg.MigratorCgrCfg().OutStorDBHost, mgrCfg.MigratorCgrCfg().OutStorDBPort,
		mgrCfg.MigratorCgrCfg().OutStorDBName, mgrCfg.MigratorCgrCfg().OutStorDBUser,
		mgrCfg.MigratorCgrCfg().OutStorDBPassword, mgrCfg.GeneralCfg().DBDataEncoding,
		mgrCfg.StorDbCfg().StringIndexedFields, mgrCfg.StorDbCfg().PrefixIndexedFields,
		mgrCfg.MigratorCgrCfg().OutStorDBOpts, mgrCfg.StorDbCfg().Items); err != nil {
		log.Fatal(err)
	}

	sameOutDB = mgrCfg.MigratorCgrCfg().OutStorDBType == mgrCfg.MigratorCgrCfg().OutDataDBType &&
		mgrCfg.MigratorCgrCfg().OutStorDBHost == mgrCfg.MigratorCgrCfg().OutDataDBHost &&
		mgrCfg.MigratorCgrCfg().OutStorDBPort == mgrCfg.MigratorCgrCfg().OutDataDBPort &&
		mgrCfg.MigratorCgrCfg().OutStorDBName == mgrCfg.MigratorCgrCfg().OutDataDBName

	if *printConfig {
		cfgJSON := utils.ToIJSON(mgrCfg.AsMapInterface(mgrCfg.GeneralCfg().RSRSep))
		log.Printf("Configuration loaded from %q:\n%s", *cfgPath, cfgJSON)
	}

	m, err := migrator.NewMigrator(dmIN, dmOUT,
		storDBIn, storDBOut,
		*dryRun, sameDataDB, sameStorDB, sameOutDB)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()
	config.SetCgrConfig(mgrCfg)
	if exec != nil && *exec != utils.EmptyString { // Run migrator
		if migrstats, err := m.Migrate(strings.Split(*exec, utils.FieldsSep)); err != nil {
			log.Fatal(err)
		} else if *verbose {
			log.Printf("Data migrated: %+v", migrstats)
		}
	}

}
