/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
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
	dmFrom     = make(map[string]migrator.MigratorDataDB)
	dmTo       = make(map[string]migrator.MigratorDataDB)
	err        error
	dfltCfg    = config.NewDefaultCGRConfig()
	cfgPath    = cgrMigratorFlags.String(utils.CfgPathCgr, utils.EmptyString,
		"Configuration directory path.")
	printConfig = cgrMigratorFlags.Bool(utils.PrintCfgCgr, false, "Print the configuration object in JSON format")
	exec        = cgrMigratorFlags.String(utils.ExecCgr, utils.EmptyString, "fire up automatic migration "+
		"<*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups|*filters|*datadb>")
	version = cgrMigratorFlags.Bool(utils.VersionCgr, false, "prints the application version")

	inDBDataEncoding = cgrMigratorFlags.String(utils.DBDataEncodingCfg, dfltCfg.GeneralCfg().DBDataEncoding,
		"the encoding used to store object Data in strings")
	dbRedisMaxConns = cgrMigratorFlags.Int(utils.RedisMaxConnsCfg, dfltCfg.DbCfg().Opts.RedisMaxConns,
		"The connection pool size")
	dbRedisConnectAttempts = cgrMigratorFlags.Int(utils.RedisConnectAttemptsCfg, dfltCfg.DbCfg().Opts.RedisConnectAttempts,
		"The maximum amount of dial attempts")
	inDataDBRedisSentinel = cgrMigratorFlags.String(utils.RedisSentinelNameCfg, dfltCfg.DbCfg().Opts.RedisSentinel,
		"the name of redis sentinel")
	dbRedisCluster = cgrMigratorFlags.Bool(utils.RedisClusterCfg, false,
		"Is the redis datadb a cluster")
	dbRedisClusterSync = cgrMigratorFlags.Duration(utils.RedisClusterSyncCfg, dfltCfg.DbCfg().Opts.RedisClusterSync,
		"The sync interval for the redis cluster")
	dbRedisClusterDownDelay = cgrMigratorFlags.Duration(utils.RedisClusterOnDownDelayCfg, dfltCfg.DbCfg().Opts.RedisClusterOndownDelay,
		"The delay before executing the commands if the redis cluster is in the CLUSTERDOWN state")
	dbRedisConnectTimeout = cgrMigratorFlags.Duration(utils.RedisConnectTimeoutCfg, dfltCfg.DbCfg().Opts.RedisConnectTimeout,
		"The amount of wait time until timeout for a connection attempt")
	dbRedisReadTimeout = cgrMigratorFlags.Duration(utils.RedisReadTimeoutCfg, dfltCfg.DbCfg().Opts.RedisReadTimeout,
		"The amount of wait time until timeout for reading operations")
	dbRedisWriteTimeout = cgrMigratorFlags.Duration(utils.RedisWriteTimeoutCfg, dfltCfg.DbCfg().Opts.RedisWriteTimeout,
		"The amount of wait time until timeout for writing operations")
	dbRedisPoolPipelineWindow = cgrMigratorFlags.Duration(utils.RedisPoolPipelineWindowCfg, dfltCfg.DbCfg().Opts.RedisPoolPipelineWindow,
		"Duration after which internal pipelines are flushed. Zero disables implicit pipelining.")
	dbRedisPoolPipelineLimit = cgrMigratorFlags.Int(utils.RedisPoolPipelineLimitCfg, dfltCfg.DbCfg().Opts.RedisPoolPipelineLimit,
		"Maximum number of commands that can be pipelined before flushing. Zero means no limit.")
	dbRedisTls               = cgrMigratorFlags.Bool(utils.RedisTLSCfg, false, "Enable TLS when connecting to Redis")
	dbRedisClientCertificate = cgrMigratorFlags.String(utils.RedisClientCertificateCfg, utils.EmptyString, "Path to the client certificate")
	dbRedisClientKey         = cgrMigratorFlags.String(utils.RedisClientKeyCfg, utils.EmptyString, "Path to the client key")
	dbRedisCACertificate     = cgrMigratorFlags.String(utils.RedisCACertificateCfg, utils.EmptyString, "Path to the CA certificate")
	dbQueryTimeout           = cgrMigratorFlags.Duration(utils.MongoQueryTimeoutCfg, dfltCfg.DbCfg().Opts.MongoQueryTimeout,
		"The timeout for queries")
	dbMongoConnScheme = cgrMigratorFlags.String(utils.MongoConnSchemeCfg, dfltCfg.DbCfg().Opts.MongoConnScheme,
		"Scheme for MongoDB connection <mongodb|mongodb+srv>")

	outDataDBRedisSentinel = cgrMigratorFlags.String(utils.OutDBRedisSentinel, utils.MetaDataDB,
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
				mgrCfg.ConfigDBCfg().Password, mgrCfg.GeneralCfg().DBDataEncoding, nil, nil,
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

	if *inDBDataEncoding != dfltCfg.GeneralCfg().DBDataEncoding {
		mgrCfg.GeneralCfg().DBDataEncoding = *inDBDataEncoding
	}
	if *dbRedisMaxConns != dfltCfg.DbCfg().Opts.RedisMaxConns {
		mgrCfg.DbCfg().Opts.RedisMaxConns = *dbRedisMaxConns
	}
	if *dbRedisConnectAttempts != dfltCfg.DbCfg().Opts.RedisConnectAttempts {
		mgrCfg.DbCfg().Opts.RedisConnectAttempts = *dbRedisConnectAttempts
	}
	if *inDataDBRedisSentinel != dfltCfg.DbCfg().Opts.RedisSentinel {
		mgrCfg.DbCfg().Opts.RedisSentinel = *inDataDBRedisSentinel
	}
	if *dbRedisCluster != dfltCfg.DbCfg().Opts.RedisCluster {
		mgrCfg.DbCfg().Opts.RedisCluster = *dbRedisCluster
	}
	if *dbRedisClusterSync != dfltCfg.DbCfg().Opts.RedisClusterSync {
		mgrCfg.DbCfg().Opts.RedisClusterSync = *dbRedisClusterSync
	}
	if *dbRedisClusterDownDelay != dfltCfg.DbCfg().Opts.RedisClusterOndownDelay {
		mgrCfg.DbCfg().Opts.RedisClusterOndownDelay = *dbRedisClusterDownDelay
	}
	if *dbRedisConnectTimeout != dfltCfg.DbCfg().Opts.RedisConnectTimeout {
		mgrCfg.DbCfg().Opts.RedisConnectTimeout = *dbRedisConnectTimeout
	}
	if *dbRedisReadTimeout != dfltCfg.DbCfg().Opts.RedisReadTimeout {
		mgrCfg.DbCfg().Opts.RedisReadTimeout = *dbRedisReadTimeout
	}
	if *dbRedisWriteTimeout != dfltCfg.DbCfg().Opts.RedisWriteTimeout {
		mgrCfg.DbCfg().Opts.RedisWriteTimeout = *dbRedisWriteTimeout
	}
	if *dbRedisPoolPipelineWindow != dfltCfg.DbCfg().Opts.RedisPoolPipelineWindow {
		mgrCfg.DbCfg().Opts.RedisPoolPipelineWindow = *dbRedisPoolPipelineWindow
	}
	if *dbRedisPoolPipelineLimit != dfltCfg.DbCfg().Opts.RedisPoolPipelineLimit {
		mgrCfg.DbCfg().Opts.RedisPoolPipelineLimit = *dbRedisPoolPipelineLimit
	}
	if *dbRedisTls != dfltCfg.DbCfg().Opts.RedisTLS {
		mgrCfg.DbCfg().Opts.RedisTLS = *dbRedisTls
	}
	if *dbRedisClientCertificate != dfltCfg.DbCfg().Opts.RedisClientCertificate {
		mgrCfg.DbCfg().Opts.RedisClientCertificate = *dbRedisClientCertificate
	}
	if *dbRedisClientKey != dfltCfg.DbCfg().Opts.RedisClientKey {
		mgrCfg.DbCfg().Opts.RedisClientKey = *dbRedisClientKey
	}
	if *dbRedisCACertificate != dfltCfg.DbCfg().Opts.RedisCACertificate {
		mgrCfg.DbCfg().Opts.RedisCACertificate = *dbRedisCACertificate
	}
	if *dbQueryTimeout != dfltCfg.DbCfg().Opts.MongoQueryTimeout {
		mgrCfg.DbCfg().Opts.MongoQueryTimeout = *dbQueryTimeout
	}
	if *dbMongoConnScheme != dfltCfg.DbCfg().Opts.MongoConnScheme {
		mgrCfg.DbCfg().Opts.MongoConnScheme = *dbMongoConnScheme
	}

	if *outDataDBRedisSentinel == utils.MetaDataDB {
		if dfltCfg.MigratorCgrCfg().OutDBOpts.RedisSentinel == mgrCfg.MigratorCgrCfg().OutDBOpts.RedisSentinel {
			mgrCfg.MigratorCgrCfg().OutDBOpts.RedisSentinel = dfltCfg.DbCfg().Opts.RedisSentinel
		}
	} else {
		mgrCfg.MigratorCgrCfg().OutDBOpts.RedisSentinel = *outDataDBRedisSentinel
	}

	toDBIDsList := []string{} // collect all DBConns of Items in data_db config
	for _, item := range mgrCfg.DbCfg().Items {
		if !slices.Contains(toDBIDsList, item.DBConn) {
			toDBIDsList = append(toDBIDsList, item.DBConn)
		}
	}

	fromDBIDsList := []string{} // collect all DBConns of MigratorFromItems in migrator config
	for _, item := range mgrCfg.MigratorCgrCfg().FromItems {
		if !slices.Contains(fromDBIDsList, item.DBConn) {
			fromDBIDsList = append(fromDBIDsList, item.DBConn)
		}
	}

	// order and compare the DBConns. If IDs are the same it means the db conns will be the same
	sameDataDB = utils.EqualUnorderedStringSlices(fromDBIDsList, toDBIDsList)

	if dmFrom, err = migrator.NewMigratorDataDBs(fromDBIDsList, mgrCfg.GeneralCfg().DBDataEncoding, mgrCfg); err != nil {
		log.Fatal(err)
	}

	if *printConfig {
		cfgJSON := utils.ToIJSON(mgrCfg.AsMapInterface())
		log.Printf("Configuration loaded from %q:\n%s", *cfgPath, cfgJSON)
	}

	if sameDataDB {
		dmTo = dmFrom
	} else {
		if dmTo, err = migrator.NewMigratorDataDBs(toDBIDsList, mgrCfg.GeneralCfg().DBDataEncoding, mgrCfg); err != nil {
			log.Fatal(err)
		}
	}

	m, err := migrator.NewMigrator(mgrCfg.DbCfg(), dmFrom, dmTo, *dryRun, sameDataDB)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()
	config.SetCgrConfig(mgrCfg)
	if exec != nil && *exec != utils.EmptyString { // Run migrator
		mig := strings.Split(*exec, utils.FieldsSep)
		err, migrstats := m.Migrate(mig)
		if err != nil {
			log.Fatal(err)
		}
		if *verbose {
			log.Printf("Data migrated: %+v", migrstats)
		}
		return
	}

}
