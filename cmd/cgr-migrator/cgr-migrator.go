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
	dmIN       migrator.MigratorDataDB
	dmOUT      migrator.MigratorDataDB
	err        error
	dfltCfg    = config.NewDefaultCGRConfig()
	cfgPath    = cgrMigratorFlags.String(utils.CfgPathCgr, utils.EmptyString,
		"Configuration directory path.")
	printConfig = cgrMigratorFlags.Bool(utils.PrintCfgCgr, false, "Print the configuration object in JSON format")
	exec        = cgrMigratorFlags.String(utils.ExecCgr, utils.EmptyString, "fire up automatic migration "+
		"<*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups|*filters|*datadb>")
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
	dbRedisTls               = cgrMigratorFlags.Bool(utils.RedisTLSCfg, false, "Enable TLS when connecting to Redis")
	dbRedisClientCertificate = cgrMigratorFlags.String(utils.RedisClientCertificateCfg, utils.EmptyString, "Path to the client certificate")
	dbRedisClientKey         = cgrMigratorFlags.String(utils.RedisClientKeyCfg, utils.EmptyString, "Path to the client key")
	dbRedisCACertificate     = cgrMigratorFlags.String(utils.RedisCACertificateCfg, utils.EmptyString, "Path to the CA certificate")

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
				mgrCfg.ConfigDBCfg().Opts, nil)
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
	if *printConfig {
		cfgJSON := utils.ToIJSON(mgrCfg.AsMapInterface(mgrCfg.GeneralCfg().RSRSep))
		log.Printf("Configuration loaded from %q:\n%s", *cfgPath, cfgJSON)
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

	m, err := migrator.NewMigrator(dmIN, dmOUT,
		*dryRun, sameDataDB)
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
		if *verbose {
			log.Printf("Data migrated: %+v", migrstats)
		}
		return
	}

}
