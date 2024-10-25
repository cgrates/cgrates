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
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	dataDB engine.DataDB

	cgrLoaderFlags = flag.NewFlagSet(utils.CgrLoader, flag.ContinueOnError)
	dfltCfg        = config.CgrConfig()
	cfgPath        = cgrLoaderFlags.String(utils.CfgPathCgr, utils.EmptyString,
		"Configuration directory path.")
	printConfig = cgrLoaderFlags.Bool(utils.PrintCfgCgr, false, "Print the configuration object in JSON format")
	dataDBType  = cgrLoaderFlags.String(utils.DataDBTypeCgr, dfltCfg.DataDbCfg().Type,
		"The type of the DataDB database <*redis|*mongo>")
	dataDBHost = cgrLoaderFlags.String(utils.DataDBHostCgr, dfltCfg.DataDbCfg().Host,
		"The DataDb host to connect to.")
	dataDBPort = cgrLoaderFlags.String(utils.DataDBPortCgr, dfltCfg.DataDbCfg().Port,
		"The DataDb port to bind to.")
	dataDBName = cgrLoaderFlags.String(utils.DataDBNameCgr, dfltCfg.DataDbCfg().Name,
		"The name/number of the DataDb to connect to.")
	dataDBUser = cgrLoaderFlags.String(utils.DataDBUserCgr, dfltCfg.DataDbCfg().User,
		"The DataDb user to sign in as.")
	dataDBPasswd = cgrLoaderFlags.String(utils.DataDBPasswdCgr, dfltCfg.DataDbCfg().Password,
		"The DataDb user's password.")
	dbDataEncoding = cgrLoaderFlags.String(utils.DBDataEncodingCfg, dfltCfg.GeneralCfg().DBDataEncoding,
		"The encoding used to store object data in strings")
	dbRedisMaxConns = cgrLoaderFlags.Int(utils.RedisMaxConnsCfg, dfltCfg.DataDbCfg().Opts.RedisMaxConns,
		"The connection pool size")
	dbRedisConnectAttempts = cgrLoaderFlags.Int(utils.RedisConnectAttemptsCfg, dfltCfg.DataDbCfg().Opts.RedisConnectAttempts,
		"The maximum amount of dial attempts")
	dbRedisSentinel = cgrLoaderFlags.String(utils.RedisSentinelNameCfg, dfltCfg.DataDbCfg().Opts.RedisSentinel,
		"The name of redis sentinel")
	dbRedisCluster = cgrLoaderFlags.Bool(utils.RedisClusterCfg, false,
		"Is the redis datadb a cluster")
	dbRedisClusterSync = cgrLoaderFlags.Duration(utils.RedisClusterSyncCfg, dfltCfg.DataDbCfg().Opts.RedisClusterSync,
		"The sync interval for the redis cluster")
	dbRedisClusterDownDelay = cgrLoaderFlags.Duration(utils.RedisClusterOnDownDelayCfg, dfltCfg.DataDbCfg().Opts.RedisClusterOndownDelay,
		"The delay before executing the commands if the redis cluster is in the CLUSTERDOWN state")
	dbRedisConnectTimeout = cgrLoaderFlags.Duration(utils.RedisConnectTimeoutCfg, dfltCfg.DataDbCfg().Opts.RedisConnectTimeout,
		"The amount of wait time until timeout for a connection attempt")
	dbRedisReadTimeout = cgrLoaderFlags.Duration(utils.RedisReadTimeoutCfg, dfltCfg.DataDbCfg().Opts.RedisReadTimeout,
		"The amount of wait time until timeout for reading operations")
	dbRedisWriteTimeout = cgrLoaderFlags.Duration(utils.RedisWriteTimeoutCfg, dfltCfg.DataDbCfg().Opts.RedisWriteTimeout,
		"The amount of wait time until timeout for writing operations")
	dbRedisPoolPipelineWindow = cgrLoaderFlags.Duration(utils.RedisPoolPipelineWindowCfg, dfltCfg.DataDbCfg().Opts.RedisPoolPipelineWindow,
		"Duration after which internal pipelines are flushed. Zero disables implicit pipelining.")
	dbRedisPoolPipelineLimit = cgrLoaderFlags.Int(utils.RedisPoolPipelineLimitCfg, dfltCfg.DataDbCfg().Opts.RedisPoolPipelineLimit,
		"Maximum number of commands that can be pipelined before flushing. Zero means no limit.")
	dbRedisTls               = cgrLoaderFlags.Bool(utils.RedisTLSCfg, false, "Enable TLS when connecting to Redis")
	dbRedisClientCertificate = cgrLoaderFlags.String(utils.RedisClientCertificateCfg, utils.EmptyString, "Path to the client certificate")
	dbRedisClientKey         = cgrLoaderFlags.String(utils.RedisClientKeyCfg, utils.EmptyString, "Path to the client key")
	dbRedisCACertificate     = cgrLoaderFlags.String(utils.RedisCACertificateCfg, utils.EmptyString, "Path to the CA certificate")
	dbQueryTimeout           = cgrLoaderFlags.Duration(utils.MongoQueryTimeoutCfg, dfltCfg.DataDbCfg().Opts.MongoQueryTimeout,
		"The timeout for queries")
	dbMongoConnScheme = cgrLoaderFlags.String(utils.MongoConnSchemeCfg, dfltCfg.DataDbCfg().Opts.MongoConnScheme,
		"Scheme for MongoDB connection <mongodb|mongodb+srv>")

	cachingArg = cgrLoaderFlags.String(utils.CachingArgCgr, utils.EmptyString,
		"Caching strategy used when loading TP")
	cachingDlay = cgrLoaderFlags.Duration(utils.CachingDlayCfg, 0, "Adds delay before cache reload")
	tpid        = cgrLoaderFlags.String(utils.TpIDCfg, dfltCfg.LoaderCgrCfg().TpID,
		"The tariff plan ID from the database")
	dataPath = cgrLoaderFlags.String(utils.PathCfg, dfltCfg.LoaderCgrCfg().DataPath,
		"The path to folder containing the data files")
	version = cgrLoaderFlags.Bool(utils.VersionCgr, false,
		"Prints the application version.")
	verbose = cgrLoaderFlags.Bool(utils.VerboseCgr, false,
		"Enable detailed verbose logging output")
	dryRun = cgrLoaderFlags.Bool(utils.DryRunCfg, false,
		"When true will not save loaded data to dataDb but just parse it for consistency and errors.")
	fieldSep = cgrLoaderFlags.String(utils.FieldSepCgr, ",",
		`Separator for csv file (by default "," is used)`)

	importID       = cgrLoaderFlags.String(utils.ImportIDCgr, utils.EmptyString, "Uniquely identify an import/load, postpended to some automatic fields")
	timezone       = cgrLoaderFlags.String(utils.TimezoneCfg, dfltCfg.GeneralCfg().DefaultTimezone, `Timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>`)
	disableReverse = cgrLoaderFlags.Bool(utils.DisableReverseCgr, false, "Will disable reverse mappings rebuilding")
	remove         = cgrLoaderFlags.Bool(utils.RemoveCgr, false, "Will remove instead of adding data from DB")
	apiKey         = cgrLoaderFlags.String(utils.APIKeyCfg, utils.EmptyString, "Api Key used to comosed ArgDispatcher")
	routeID        = cgrLoaderFlags.String(utils.RouteIDCfg, utils.EmptyString, "RouteID used to comosed ArgDispatcher")
	tenant         = cgrLoaderFlags.String(utils.TenantCfg, dfltCfg.GeneralCfg().DefaultTenant, "If set, will overwrite the default tenant")

	cacheSAddress = cgrLoaderFlags.String(utils.CacheSAddress, dfltCfg.LoaderCgrCfg().CachesConns[0],
		"CacheS component to contact for cache reloads, empty to disable automatic cache reloads")
	schedulerAddress = cgrLoaderFlags.String(utils.SchedulerAddress, dfltCfg.LoaderCgrCfg().ActionSConns[0], "")
	rpcEncoding      = cgrLoaderFlags.String(utils.RpcEncodingCgr, rpcclient.JSONrpc, "RPC encoding used <*gob|*json>")
)

func loadConfig() (ldrCfg *config.CGRConfig) {
	ldrCfg = config.CgrConfig()
	if *cfgPath != utils.EmptyString {
		var err error
		if ldrCfg, err = config.NewCGRConfigFromPath(context.Background(), *cfgPath); err != nil {
			log.Fatalf("Error loading config file %s", err)
		}
		if ldrCfg.ConfigDBCfg().Type != utils.MetaInternal {
			d, err := engine.NewDataDBConn(ldrCfg.ConfigDBCfg().Type,
				ldrCfg.ConfigDBCfg().Host, ldrCfg.ConfigDBCfg().Port,
				ldrCfg.ConfigDBCfg().Name, ldrCfg.ConfigDBCfg().User,
				ldrCfg.ConfigDBCfg().Password, ldrCfg.GeneralCfg().DBDataEncoding,
				ldrCfg.ConfigDBCfg().Opts, nil)
			if err != nil { // Cannot configure getter database, show stopper
				utils.Logger.Crit(fmt.Sprintf("Could not configure configDB: %s exiting!", err))
				return
			}
			if err = ldrCfg.LoadFromDB(context.Background(), d); err != nil {
				log.Fatalf("Could not parse config: <%s>", err.Error())
				return
			}
		}
		config.SetCgrConfig(ldrCfg)
	}
	// Data for DataDB
	if *dataDBType != dfltCfg.DataDbCfg().Type {
		ldrCfg.DataDbCfg().Type = *dataDBType
	}

	if *dataDBHost != dfltCfg.DataDbCfg().Host {
		ldrCfg.DataDbCfg().Host = *dataDBHost
	}

	if *dataDBPort != dfltCfg.DataDbCfg().Port {
		ldrCfg.DataDbCfg().Port = *dataDBPort
	}

	if *dataDBName != dfltCfg.DataDbCfg().Name {
		ldrCfg.DataDbCfg().Name = *dataDBName
	}

	if *dataDBUser != dfltCfg.DataDbCfg().User {
		ldrCfg.DataDbCfg().User = *dataDBUser
	}

	if *dataDBPasswd != dfltCfg.DataDbCfg().Password {
		ldrCfg.DataDbCfg().Password = *dataDBPasswd
	}

	if *dbRedisMaxConns != dfltCfg.DataDbCfg().Opts.RedisMaxConns {
		ldrCfg.DataDbCfg().Opts.RedisMaxConns = *dbRedisMaxConns
	}
	if *dbRedisConnectAttempts != dfltCfg.DataDbCfg().Opts.RedisConnectAttempts {
		ldrCfg.DataDbCfg().Opts.RedisConnectAttempts = *dbRedisConnectAttempts
	}
	if *dbRedisSentinel != dfltCfg.DataDbCfg().Opts.RedisSentinel {
		ldrCfg.DataDbCfg().Opts.RedisSentinel = *dbRedisSentinel
	}
	if *dbRedisCluster != dfltCfg.DataDbCfg().Opts.RedisCluster {
		ldrCfg.DataDbCfg().Opts.RedisCluster = *dbRedisCluster
	}
	if *dbRedisClusterSync != dfltCfg.DataDbCfg().Opts.RedisClusterSync {
		ldrCfg.DataDbCfg().Opts.RedisClusterSync = *dbRedisClusterSync
	}
	if *dbRedisClusterDownDelay != dfltCfg.DataDbCfg().Opts.RedisClusterOndownDelay {
		ldrCfg.DataDbCfg().Opts.RedisClusterOndownDelay = *dbRedisClusterDownDelay
	}
	if *dbRedisConnectTimeout != dfltCfg.DataDbCfg().Opts.RedisConnectTimeout {
		ldrCfg.DataDbCfg().Opts.RedisConnectTimeout = *dbRedisConnectTimeout
	}
	if *dbRedisReadTimeout != dfltCfg.DataDbCfg().Opts.RedisReadTimeout {
		ldrCfg.DataDbCfg().Opts.RedisReadTimeout = *dbRedisReadTimeout
	}
	if *dbRedisWriteTimeout != dfltCfg.DataDbCfg().Opts.RedisWriteTimeout {
		ldrCfg.DataDbCfg().Opts.RedisWriteTimeout = *dbRedisWriteTimeout
	}
	if *dbRedisPoolPipelineWindow != dfltCfg.DataDbCfg().Opts.RedisPoolPipelineWindow {
		ldrCfg.DataDbCfg().Opts.RedisPoolPipelineWindow = *dbRedisPoolPipelineWindow
	}
	if *dbRedisPoolPipelineLimit != dfltCfg.DataDbCfg().Opts.RedisPoolPipelineLimit {
		ldrCfg.DataDbCfg().Opts.RedisPoolPipelineLimit = *dbRedisPoolPipelineLimit
	}
	if *dbQueryTimeout != dfltCfg.DataDbCfg().Opts.MongoQueryTimeout {
		ldrCfg.DataDbCfg().Opts.MongoQueryTimeout = *dbQueryTimeout
	}
	if *dbMongoConnScheme != dfltCfg.DataDbCfg().Opts.MongoConnScheme {
		ldrCfg.DataDbCfg().Opts.MongoConnScheme = *dbMongoConnScheme
	}
	if *dbRedisTls != dfltCfg.DataDbCfg().Opts.RedisTLS {
		ldrCfg.DataDbCfg().Opts.RedisTLS = *dbRedisTls
	}
	if *dbRedisClientCertificate != dfltCfg.DataDbCfg().Opts.RedisClientCertificate {
		ldrCfg.DataDbCfg().Opts.RedisClientCertificate = *dbRedisClientCertificate
	}
	if *dbRedisClientKey != dfltCfg.DataDbCfg().Opts.RedisClientKey {
		ldrCfg.DataDbCfg().Opts.RedisClientKey = *dbRedisClientKey
	}
	if *dbRedisCACertificate != dfltCfg.DataDbCfg().Opts.RedisCACertificate {
		ldrCfg.DataDbCfg().Opts.RedisCACertificate = *dbRedisCACertificate
	}

	if *dbDataEncoding != dfltCfg.GeneralCfg().DBDataEncoding {
		ldrCfg.GeneralCfg().DBDataEncoding = *dbDataEncoding
	}

	if *tpid != dfltCfg.LoaderCgrCfg().TpID {
		ldrCfg.LoaderCgrCfg().TpID = *tpid
	}

	if *dataPath != dfltCfg.LoaderCgrCfg().DataPath {
		ldrCfg.LoaderCgrCfg().DataPath = *dataPath
	}

	if rune((*fieldSep)[0]) != dfltCfg.LoaderCgrCfg().FieldSeparator {
		ldrCfg.LoaderCgrCfg().FieldSeparator = rune((*fieldSep)[0])
	}

	if *cacheSAddress != dfltCfg.LoaderCgrCfg().CachesConns[0] {
		if *cacheSAddress == utils.EmptyString {
			ldrCfg.LoaderCgrCfg().CachesConns = []string{}
		} else {
			ldrCfg.LoaderCgrCfg().CachesConns = []string{*cacheSAddress}
			if _, has := ldrCfg.RPCConns()[*cacheSAddress]; !has {
				ldrCfg.RPCConns()[*cacheSAddress] = &config.RPCConn{
					Strategy: rpcclient.PoolFirst,
					Conns: []*config.RemoteHost{{
						Address:   *cacheSAddress,
						Transport: *rpcEncoding,
					}},
				}
			}
		}
	}

	if *schedulerAddress != dfltCfg.LoaderCgrCfg().ActionSConns[0] {
		if *schedulerAddress == utils.EmptyString {
			ldrCfg.LoaderCgrCfg().ActionSConns = []string{}
		} else {
			ldrCfg.LoaderCgrCfg().ActionSConns = []string{*schedulerAddress}
			if _, has := ldrCfg.RPCConns()[*schedulerAddress]; !has {
				ldrCfg.RPCConns()[*schedulerAddress] = &config.RPCConn{
					Strategy: rpcclient.PoolFirst,
					Conns: []*config.RemoteHost{{
						Address:   *schedulerAddress,
						Transport: *rpcEncoding,
					}},
				}
			}
		}
	}

	if *importID == utils.EmptyString {
		*importID = utils.UUIDSha1Prefix()
	}

	if *timezone != dfltCfg.GeneralCfg().DefaultTimezone {
		ldrCfg.GeneralCfg().DefaultTimezone = *timezone
	}

	if *disableReverse != dfltCfg.LoaderCgrCfg().DisableReverse {
		ldrCfg.LoaderCgrCfg().DisableReverse = *disableReverse
	}

	if *cachingArg != utils.EmptyString {
		ldrCfg.GeneralCfg().DefaultCaching = *cachingArg
	}
	if *cachingDlay != 0 {
		ldrCfg.GeneralCfg().CachingDelay = *cachingDlay
	}
	return
}

func getLoader(cfg *config.CGRConfig) (engine.LoadReader, error) {
	if gprefix := utils.MetaGoogleAPI + utils.ConcatenatedKeySep; strings.HasPrefix(*dataPath, gprefix) { // Default load from csv files to dataDb
		return engine.NewGoogleCSVStorage(cfg.LoaderCgrCfg().FieldSeparator, strings.TrimPrefix(*dataPath, gprefix))
	}
	if !utils.IsURL(*dataPath) {
		return engine.NewFileCSVStorage(cfg.LoaderCgrCfg().FieldSeparator, *dataPath)
	}
	return engine.NewURLCSVStorage(cfg.LoaderCgrCfg().FieldSeparator, *dataPath), nil
}

func main() {
	var err error
	if err = cgrLoaderFlags.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
	if *version {
		var version string
		if version, err = utils.GetCGRVersion(); err != nil {
			log.Fatal(err)
		}
		fmt.Println(version)
		return
	}

	ldrCfg := loadConfig()
	// we initialize connManager here with nil for InternalChannels
	engine.NewConnManager(ldrCfg)

	if dataDB, err = engine.NewDataDBConn(ldrCfg.DataDbCfg().Type,
		ldrCfg.DataDbCfg().Host, ldrCfg.DataDbCfg().Port,
		ldrCfg.DataDbCfg().Name, ldrCfg.DataDbCfg().User,
		ldrCfg.DataDbCfg().Password, ldrCfg.GeneralCfg().DBDataEncoding,
		ldrCfg.DataDbCfg().Opts, ldrCfg.DataDbCfg().Items); err != nil {
		log.Fatalf("Coud not open dataDB connection: %s", err.Error())
	}
	defer dataDB.Close()

	var loader engine.LoadReader
	if loader, err = getLoader(ldrCfg); err != nil {
		log.Fatal(err)
	}
	var tpReader *engine.TpReader
	if tpReader, err = engine.NewTpReader(dataDB, loader,
		ldrCfg.LoaderCgrCfg().TpID, ldrCfg.GeneralCfg().DefaultTimezone,
		ldrCfg.LoaderCgrCfg().CachesConns,
		ldrCfg.LoaderCgrCfg().ActionSConns, false); err != nil {
		log.Fatal(err)
	}
	if err = tpReader.LoadAll(); err != nil {
		log.Fatal(err)
	}

	if *dryRun { // We were just asked to parse the data, not saving it
		return
	}
	if *printConfig {
		cfgJSON := utils.ToIJSON(ldrCfg.AsMapInterface(ldrCfg.GeneralCfg().RSRSep))
		log.Printf("Configuration loaded from %q:\n%s", *cfgPath, cfgJSON)
	}
	if *remove {
		if err = tpReader.RemoveFromDatabase(*verbose, *disableReverse); err != nil {
			log.Fatal("Could not delete from database: ", err)
		}
	} else {
		// write maps to database
		if err = tpReader.WriteToDatabase(*verbose, *disableReverse); err != nil {
			log.Fatal("Could not write to database: ", err)
		}
	}

	// delay if needed before cache reload
	if *verbose && ldrCfg.GeneralCfg().CachingDelay != 0 {
		log.Printf("Delaying cache reload for %v", ldrCfg.GeneralCfg().CachingDelay)
		time.Sleep(ldrCfg.GeneralCfg().CachingDelay)
	}

	// reload cache
	if err = tpReader.ReloadCache(context.Background(), ldrCfg.GeneralCfg().DefaultCaching, *verbose, map[string]any{
		utils.OptsAPIKey:  *apiKey,
		utils.OptsRouteID: *routeID,
	}, *tenant); err != nil {
		log.Fatal("Could not reload cache: ", err)
	}

	if len(ldrCfg.LoaderCgrCfg().ActionSConns) != 0 {
		if err = tpReader.ReloadScheduler(*verbose); err != nil {
			log.Fatal("Could not reload scheduler: ", err)
		}
	}
}
