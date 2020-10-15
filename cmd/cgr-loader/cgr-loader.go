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
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	dataDB engine.DataDB
	storDB engine.LoadStorage

	cgrLoaderFlags = flag.NewFlagSet("cgr-loader", flag.ContinueOnError)
	dfltCfg        = config.CgrConfig()
	cfgPath        = cgrLoaderFlags.String("config_path", utils.EmptyString,
		"Configuration directory path.")

	dataDBType = cgrLoaderFlags.String("datadb_type", dfltCfg.DataDbCfg().DataDbType,
		"The type of the DataDB database <*redis|*mongo>")
	dataDBHost = cgrLoaderFlags.String("datadb_host", dfltCfg.DataDbCfg().DataDbHost,
		"The DataDb host to connect to.")
	dataDBPort = cgrLoaderFlags.String("datadb_port", dfltCfg.DataDbCfg().DataDbPort,
		"The DataDb port to bind to.")
	dataDBName = cgrLoaderFlags.String("datadb_name", dfltCfg.DataDbCfg().DataDbName,
		"The name/number of the DataDb to connect to.")
	dataDBUser = cgrLoaderFlags.String("datadb_user", dfltCfg.DataDbCfg().DataDbUser,
		"The DataDb user to sign in as.")
	dataDBPasswd = cgrLoaderFlags.String("datadb_passwd", dfltCfg.DataDbCfg().DataDbPass,
		"The DataDb user's password.")
	dbDataEncoding = cgrLoaderFlags.String("dbdata_encoding", dfltCfg.GeneralCfg().DBDataEncoding,
		"The encoding used to store object data in strings")
	dbRedisSentinel = cgrLoaderFlags.String("redis_sentinel", utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisSentinelNameCfg]),
		"The name of redis sentinel")
	dbRedisCluster = cgrLoaderFlags.Bool("redis_cluster", false,
		"Is the redis datadb a cluster")
	dbRedisClusterSync = cgrLoaderFlags.String("redis_cluster_sync", utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClusterSyncCfg]),
		"The sync interval for the redis cluster")
	dbRedisClusterDownDelay = cgrLoaderFlags.String("redis_cluster_ondown_delay", utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClusterOnDownDelayCfg]),
		"The delay before executing the commands if the redis cluster is in the CLUSTERDOWN state")
	dbQueryTimeout = cgrLoaderFlags.String("query_timeout", utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.QueryTimeoutCfg]),
		"The timeout for queries")
	dbRedisTls               = cgrLoaderFlags.Bool(utils.RedisTLS, false, "Use a tls connection when connecting to redis")
	dbRedisClientCertificate = cgrLoaderFlags.String(utils.RedisClientCertificate, utils.EmptyString, "Path to the client certificate")
	dbRedisClientKey         = cgrLoaderFlags.String(utils.RedisClientKey, utils.EmptyString, "Path to the client key")
	dbRedisCACertificate     = cgrLoaderFlags.String(utils.RedisCACertificate, utils.EmptyString, "Path to the CA certificate")

	storDBType = cgrLoaderFlags.String("stordb_type", dfltCfg.StorDbCfg().Type,
		"The type of the storDb database <*mysql|*postgres|*mongo>")
	storDBHost = cgrLoaderFlags.String("stordb_host", dfltCfg.StorDbCfg().Host,
		"The storDb host to connect to.")
	storDBPort = cgrLoaderFlags.String("stordb_port", dfltCfg.StorDbCfg().Port,
		"The storDb port to bind to.")
	storDBName = cgrLoaderFlags.String("stordb_name", dfltCfg.StorDbCfg().Name,
		"The name/number of the storDb to connect to.")
	storDBUser = cgrLoaderFlags.String("stordb_user", dfltCfg.StorDbCfg().User,
		"The storDb user to sign in as.")
	storDBPasswd = cgrLoaderFlags.String("stordb_passwd", dfltCfg.StorDbCfg().Password,
		"The storDb user's password.")

	cachingArg = cgrLoaderFlags.String("caching", utils.EmptyString,
		"Caching strategy used when loading TP")
	tpid = cgrLoaderFlags.String("tpid", dfltCfg.LoaderCgrCfg().TpID,
		"The tariff plan ID from the database")
	dataPath = cgrLoaderFlags.String("path", dfltCfg.LoaderCgrCfg().DataPath,
		"The path to folder containing the data files")
	version = cgrLoaderFlags.Bool("version", false,
		"Prints the application version.")
	verbose = cgrLoaderFlags.Bool("verbose", false,
		"Enable detailed verbose logging output")
	dryRun = cgrLoaderFlags.Bool("dry_run", false,
		"When true will not save loaded data to dataDb but just parse it for consistency and errors.")
	fieldSep = cgrLoaderFlags.String("field_sep", ",",
		`Separator for csv file (by default "," is used)`)

	importID       = cgrLoaderFlags.String("import_id", utils.EmptyString, "Uniquely identify an import/load, postpended to some automatic fields")
	timezone       = cgrLoaderFlags.String("timezone", dfltCfg.GeneralCfg().DefaultTimezone, `Timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>`)
	disableReverse = cgrLoaderFlags.Bool("disable_reverse_mappings", false, "Will disable reverse mappings rebuilding")
	flushStorDB    = cgrLoaderFlags.Bool("flush_stordb", false, "Remove tariff plan data for id from the database")
	remove         = cgrLoaderFlags.Bool("remove", false, "Will remove instead of adding data from DB")
	apiKey         = cgrLoaderFlags.String("api_key", utils.EmptyString, "Api Key used to comosed ArgDispatcher")
	routeID        = cgrLoaderFlags.String("route_id", utils.EmptyString, "RouteID used to comosed ArgDispatcher")

	fromStorDB    = cgrLoaderFlags.Bool("from_stordb", false, "Load the tariff plan from storDb to dataDb")
	toStorDB      = cgrLoaderFlags.Bool("to_stordb", false, "Import the tariff plan from files to storDb")
	cacheSAddress = cgrLoaderFlags.String("caches_address", dfltCfg.LoaderCgrCfg().CachesConns[0],
		"CacheS component to contact for cache reloads, empty to disable automatic cache reloads")
	schedulerAddress = cgrLoaderFlags.String("scheduler_address", dfltCfg.LoaderCgrCfg().SchedulerConns[0], "")
	rpcEncoding      = cgrLoaderFlags.String("rpc_encoding", rpcclient.JSONrpc, "RPC encoding used <*gob|*json>")
)

func loadConfig() (ldrCfg *config.CGRConfig) {
	ldrCfg = config.CgrConfig()
	if *cfgPath != utils.EmptyString {
		var err error
		if ldrCfg, err = config.NewCGRConfigFromPath(*cfgPath); err != nil {
			log.Fatalf("Error loading config file %s", err)
		}
		config.SetCgrConfig(ldrCfg)
	}
	// Data for DataDB
	if *dataDBType != dfltCfg.DataDbCfg().DataDbType {
		ldrCfg.DataDbCfg().DataDbType = strings.TrimPrefix(*dataDBType, utils.Meta)
	}

	if *dataDBHost != dfltCfg.DataDbCfg().DataDbHost {
		ldrCfg.DataDbCfg().DataDbHost = *dataDBHost
	}

	if *dataDBPort != dfltCfg.DataDbCfg().DataDbPort {
		ldrCfg.DataDbCfg().DataDbPort = *dataDBPort
	}

	if *dataDBName != dfltCfg.DataDbCfg().DataDbName {
		ldrCfg.DataDbCfg().DataDbName = *dataDBName
	}

	if *dataDBUser != dfltCfg.DataDbCfg().DataDbUser {
		ldrCfg.DataDbCfg().DataDbUser = *dataDBUser
	}

	if *dataDBPasswd != dfltCfg.DataDbCfg().DataDbPass {
		ldrCfg.DataDbCfg().DataDbPass = *dataDBPasswd
	}

	if *dbRedisSentinel != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisSentinelNameCfg]) {
		ldrCfg.DataDbCfg().Opts[utils.RedisSentinelNameCfg] = *dbRedisSentinel
	}

	rdsCls, _ := utils.IfaceAsBool(dfltCfg.DataDbCfg().Opts[utils.RedisClusterCfg])
	if *dbRedisCluster != rdsCls {
		ldrCfg.DataDbCfg().Opts[utils.RedisClusterCfg] = *dbRedisCluster
	}
	if *dbRedisClusterSync != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClusterSyncCfg]) {
		ldrCfg.DataDbCfg().Opts[utils.RedisClusterSyncCfg] = *dbRedisClusterSync
	}
	if *dbRedisClusterDownDelay != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClusterOnDownDelayCfg]) {
		ldrCfg.DataDbCfg().Opts[utils.RedisClusterOnDownDelayCfg] = *dbRedisClusterDownDelay
	}
	if *dbQueryTimeout != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.QueryTimeoutCfg]) {
		ldrCfg.DataDbCfg().Opts[utils.QueryTimeoutCfg] = *dbQueryTimeout
	}

	rdsTLS, _ := utils.IfaceAsBool(dfltCfg.DataDbCfg().Opts[utils.RedisTLS])
	if *dbRedisTls != rdsTLS {
		ldrCfg.DataDbCfg().Opts[utils.RedisTLS] = *dbRedisTls
	}
	if *dbRedisClientCertificate != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClientCertificate]) {
		ldrCfg.DataDbCfg().Opts[utils.RedisClientCertificate] = *dbRedisClientCertificate
	}
	if *dbRedisClientKey != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisClientKey]) {
		ldrCfg.DataDbCfg().Opts[utils.RedisClientKey] = *dbRedisClientKey
	}
	if *dbRedisCACertificate != utils.IfaceAsString(dfltCfg.DataDbCfg().Opts[utils.RedisCACertificate]) {
		ldrCfg.DataDbCfg().Opts[utils.RedisCACertificate] = *dbRedisCACertificate
	}

	if *dbDataEncoding != dfltCfg.GeneralCfg().DBDataEncoding {
		ldrCfg.GeneralCfg().DBDataEncoding = *dbDataEncoding
	}

	// Data for StorDB
	if *storDBType != dfltCfg.StorDbCfg().Type {
		ldrCfg.StorDbCfg().Type = strings.TrimPrefix(*storDBType, utils.Meta)
	}

	if *storDBHost != dfltCfg.StorDbCfg().Host {
		ldrCfg.StorDbCfg().Host = *storDBHost
	}

	if *storDBPort != dfltCfg.StorDbCfg().Port {
		ldrCfg.StorDbCfg().Port = *storDBPort
	}

	if *storDBName != dfltCfg.StorDbCfg().Name {
		ldrCfg.StorDbCfg().Name = *storDBName
	}

	if *storDBUser != dfltCfg.StorDbCfg().User {
		ldrCfg.StorDbCfg().User = *storDBUser
	}

	if *storDBPasswd != dfltCfg.StorDbCfg().Password {
		ldrCfg.StorDbCfg().Password = *storDBPasswd
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

	if *schedulerAddress != dfltCfg.LoaderCgrCfg().SchedulerConns[0] {
		if *schedulerAddress == utils.EmptyString {
			ldrCfg.LoaderCgrCfg().SchedulerConns = []string{}
		} else {
			ldrCfg.LoaderCgrCfg().SchedulerConns = []string{*schedulerAddress}
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
	return
}

func importData(cfg *config.CGRConfig) (err error) {
	if cfg.LoaderCgrCfg().TpID == utils.EmptyString {
		return errors.New("TPid required")
	}
	if *flushStorDB {
		if err = storDB.RemTpData(utils.EmptyString, cfg.LoaderCgrCfg().TpID, map[string]string{}); err != nil {
			return
		}
	}
	csvImporter := engine.TPCSVImporter{
		TPid:     cfg.LoaderCgrCfg().TpID,
		StorDb:   storDB,
		DirPath:  *dataPath,
		Sep:      cfg.LoaderCgrCfg().FieldSeparator,
		Verbose:  *verbose,
		ImportId: *importID,
	}
	return csvImporter.Run()
}

func getLoader(cfg *config.CGRConfig) (loader engine.LoadReader, err error) {
	if *fromStorDB { // Load Tariff Plan from storDb into dataDb
		loader = storDB
		return
	}
	if gprefix := utils.MetaGoogleAPI + utils.CONCATENATED_KEY_SEP; strings.HasPrefix(*dataPath, gprefix) { // Default load from csv files to dataDb
		return engine.NewGoogleCSVStorage(cfg.LoaderCgrCfg().FieldSeparator, strings.TrimPrefix(*dataPath, gprefix))
	}
	if !utils.IsURL(*dataPath) {
		loader = engine.NewFileCSVStorage(cfg.LoaderCgrCfg().FieldSeparator, *dataPath)
		return
	}
	loader = engine.NewURLCSVStorage(cfg.LoaderCgrCfg().FieldSeparator, *dataPath)
	return
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
	engine.NewConnManager(ldrCfg, nil)

	if !*toStorDB {
		if dataDB, err = engine.NewDataDBConn(ldrCfg.DataDbCfg().DataDbType,
			ldrCfg.DataDbCfg().DataDbHost, ldrCfg.DataDbCfg().DataDbPort,
			ldrCfg.DataDbCfg().DataDbName, ldrCfg.DataDbCfg().DataDbUser,
			ldrCfg.DataDbCfg().DataDbPass, ldrCfg.GeneralCfg().DBDataEncoding,
			ldrCfg.DataDbCfg().Opts); err != nil {
			log.Fatalf("Coud not open dataDB connection: %s", err.Error())
		}
		defer dataDB.Close()
	}

	if *fromStorDB || *toStorDB {
		if storDB, err = engine.NewStorDBConn(ldrCfg.StorDbCfg().Type,
			ldrCfg.StorDbCfg().Host, ldrCfg.StorDbCfg().Port,
			ldrCfg.StorDbCfg().Name, ldrCfg.StorDbCfg().User,
			ldrCfg.StorDbCfg().Password, ldrCfg.GeneralCfg().DBDataEncoding,
			ldrCfg.StorDbCfg().StringIndexedFields, ldrCfg.StorDbCfg().PrefixIndexedFields,
			ldrCfg.StorDbCfg().Opts); err != nil {
			log.Fatalf("Coud not open storDB connection: %s", err.Error())
		}
		defer storDB.Close()
	}

	if !*dryRun && *toStorDB { // Import files from a directory into storDb
		if err = importData(ldrCfg); err != nil {
			log.Fatal(err)
		}
		return
	}
	var loader engine.LoadReader
	if loader, err = getLoader(ldrCfg); err != nil {
		log.Fatal(err)
	}
	var tpReader *engine.TpReader
	if tpReader, err = engine.NewTpReader(dataDB, loader,
		ldrCfg.LoaderCgrCfg().TpID, ldrCfg.GeneralCfg().DefaultTimezone,
		ldrCfg.LoaderCgrCfg().CachesConns,
		ldrCfg.LoaderCgrCfg().SchedulerConns, false); err != nil {
		log.Fatal(err)
	}
	if err = tpReader.LoadAll(); err != nil {
		log.Fatal(err)
	}

	if *dryRun { // We were just asked to parse the data, not saving it
		return
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

	// reload cache
	if err = tpReader.ReloadCache(ldrCfg.GeneralCfg().DefaultCaching, *verbose, map[string]interface{}{
		utils.OptsAPIKey:  *apiKey,
		utils.OptsRouteID: *routeID,
	}); err != nil {
		log.Fatal("Could not reload cache: ", err)
	}

	if len(ldrCfg.LoaderCgrCfg().SchedulerConns) != 0 {
		if err = tpReader.ReloadScheduler(*verbose); err != nil {
			log.Fatal("Could not reload scheduler: ", err)
		}
	}
}
