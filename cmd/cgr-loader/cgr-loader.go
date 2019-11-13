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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	cgrLoaderFlags = flag.NewFlagSet("cgr-loader", flag.ContinueOnError)
	dfltCfg        = config.CgrConfig()
	cfgPath        = cgrLoaderFlags.String("config_path", "",
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
	dbRedisSentinel = cgrLoaderFlags.String("redis_sentinel", dfltCfg.DataDbCfg().DataDbSentinelName,
		"The name of redis sentinel")

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

	cachingArg = cgrLoaderFlags.String("caching", "",
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
	recursive = cgrLoaderFlags.Bool("recursive", false, "Loads data from folder recursive.")

	fromStorDB    = cgrLoaderFlags.Bool("from_stordb", false, "Load the tariff plan from storDb to dataDb")
	toStorDB      = cgrLoaderFlags.Bool("to_stordb", false, "Import the tariff plan from files to storDb")
	rpcEncoding   = cgrLoaderFlags.String("rpc_encoding", utils.MetaJSONrpc, "RPC encoding used <*gob|*json>")
	cacheSAddress = cgrLoaderFlags.String("caches_address", dfltCfg.LoaderCgrCfg().CachesConns[0].Address,
		"CacheS component to contact for cache reloads, empty to disable automatic cache reloads")
	schedulerAddress = cgrLoaderFlags.String("scheduler_address", dfltCfg.LoaderCgrCfg().SchedulerConns[0].Address, "")

	importID       = cgrLoaderFlags.String("import_id", "", "Uniquely identify an import/load, postpended to some automatic fields")
	timezone       = cgrLoaderFlags.String("timezone", "", `Timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>`)
	disableReverse = cgrLoaderFlags.Bool("disable_reverse_mappings", false, "Will disable reverse mappings rebuilding")
	flushStorDB    = cgrLoaderFlags.Bool("flush_stordb", false, "Remove tariff plan data for id from the database")
	remove         = cgrLoaderFlags.Bool("remove", false, "Will remove instead of adding data from DB")
	apiKey         = cgrLoaderFlags.String("api_key", "", "Api Key used to comosed ArgDispatcher")
	routeID        = cgrLoaderFlags.String("route_id", "", "RouteID used to comosed ArgDispatcher")

	err        error
	dm         *engine.DataManager
	storDb     engine.LoadStorage
	cacheS     rpcclient.RpcClientConnection
	schedulerS rpcclient.RpcClientConnection
	loader     engine.LoadReader
)

func main() {
	if err := cgrLoaderFlags.Parse(os.Args[1:]); err != nil {
		return
	}
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}

	ldrCfg := config.CgrConfig()
	if *cfgPath != "" {
		if ldrCfg, err = config.NewCGRConfigFromPath(*cfgPath); err != nil {
			log.Fatalf("Error loading config file %s", err.Error())
		}
		config.SetCgrConfig(ldrCfg)
	}

	// Data for DataDB
	if *dataDBType != dfltCfg.DataDbCfg().DataDbType {
		ldrCfg.DataDbCfg().DataDbType = strings.TrimPrefix(*dataDBType, "*")
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

	if *dbRedisSentinel != dfltCfg.DataDbCfg().DataDbSentinelName {
		ldrCfg.DataDbCfg().DataDbSentinelName = *dbRedisSentinel
	}

	if *dbDataEncoding != dfltCfg.GeneralCfg().DBDataEncoding {
		ldrCfg.GeneralCfg().DBDataEncoding = *dbDataEncoding
	}

	// Data for StorDB
	if *storDBType != dfltCfg.StorDbCfg().Type {
		ldrCfg.StorDbCfg().Type = strings.TrimPrefix(*storDBType, "*")
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

	if *tpid != dfltCfg.LoaderCgrCfg().DataPath {
		ldrCfg.LoaderCgrCfg().TpID = *tpid
	}

	if *dataPath != dfltCfg.LoaderCgrCfg().DataPath {
		ldrCfg.LoaderCgrCfg().DataPath = *dataPath
	}

	if rune((*fieldSep)[0]) != dfltCfg.LoaderCgrCfg().FieldSeparator {
		ldrCfg.LoaderCgrCfg().FieldSeparator = rune((*fieldSep)[0])
	}

	if *cacheSAddress != dfltCfg.LoaderCgrCfg().CachesConns[0].Address {
		ldrCfg.LoaderCgrCfg().CachesConns = make([]*config.RemoteHost, 0)
		if *cacheSAddress != "" {
			ldrCfg.LoaderCgrCfg().CachesConns = append(ldrCfg.LoaderCgrCfg().CachesConns,
				&config.RemoteHost{
					Address:   *cacheSAddress,
					Transport: *rpcEncoding,
				})
		}
	}

	if *schedulerAddress != dfltCfg.LoaderCgrCfg().SchedulerConns[0].Address {
		ldrCfg.LoaderCgrCfg().SchedulerConns = make([]*config.RemoteHost, 0)
		if *schedulerAddress != "" {
			ldrCfg.LoaderCgrCfg().SchedulerConns = append(ldrCfg.LoaderCgrCfg().SchedulerConns,
				&config.RemoteHost{Address: *schedulerAddress})
		}
	}

	if len(ldrCfg.LoaderCgrCfg().CachesConns) != 0 &&
		*rpcEncoding != dfltCfg.LoaderCgrCfg().CachesConns[0].Transport {
		ldrCfg.LoaderCgrCfg().CachesConns[0].Transport = *rpcEncoding
	}

	if *importID == "" {
		*importID = utils.UUIDSha1Prefix()
	}

	if *timezone != dfltCfg.GeneralCfg().DefaultTimezone {
		ldrCfg.GeneralCfg().DefaultTimezone = *timezone
	}

	if *disableReverse != dfltCfg.LoaderCgrCfg().DisableReverse {
		ldrCfg.LoaderCgrCfg().DisableReverse = *disableReverse
	}

	if !*toStorDB {
		d, err := engine.NewDataDBConn(ldrCfg.DataDbCfg().DataDbType,
			ldrCfg.DataDbCfg().DataDbHost, ldrCfg.DataDbCfg().DataDbPort,
			ldrCfg.DataDbCfg().DataDbName, ldrCfg.DataDbCfg().DataDbUser,
			ldrCfg.DataDbCfg().DataDbPass, ldrCfg.GeneralCfg().DBDataEncoding,
			ldrCfg.DataDbCfg().DataDbSentinelName)
		if err != nil {
			log.Fatalf("Coud not open dataDB connection: %s", err.Error())
		}
		var rmtDBConns []*engine.DataManager
		var rplConns *rpcclient.RpcClientPool
		if len(ldrCfg.DataDbCfg().RmtDataDBCfgs) != 0 {
			rmtDBConns = make([]*engine.DataManager, len(ldrCfg.DataDbCfg().RmtDataDBCfgs))
			for i, dbCfg := range ldrCfg.DataDbCfg().RmtDataDBCfgs {
				dbConn, err := engine.NewDataDBConn(dbCfg.DataDbType,
					dbCfg.DataDbHost, dbCfg.DataDbPort,
					dbCfg.DataDbName, dbCfg.DataDbUser,
					dbCfg.DataDbPass, ldrCfg.GeneralCfg().DBDataEncoding,
					dbCfg.DataDbSentinelName)
				if err != nil {
					log.Fatalf("Coud not open dataDB connection: %s", err.Error())
				}
				rmtDBConns[i] = engine.NewDataManager(dbConn, nil, nil, nil)
			}
		}
		if len(ldrCfg.DataDbCfg().RplConns) != 0 {
			var err error
			rplConns, err = engine.NewRPCPool(rpcclient.POOL_BROADCAST, ldrCfg.TlsCfg().ClientKey,
				ldrCfg.TlsCfg().ClientCerificate, ldrCfg.TlsCfg().CaCertificate,
				ldrCfg.GeneralCfg().ConnectAttempts, ldrCfg.GeneralCfg().Reconnects,
				ldrCfg.GeneralCfg().ConnectTimeout, ldrCfg.GeneralCfg().ReplyTimeout,
				ldrCfg.DataDbCfg().RplConns, nil, false)
			if err != nil {
				log.Fatalf("Coud not confignure dataDB replication connections: %s", err.Error())
			}
		}
		dm = engine.NewDataManager(d, config.CgrConfig().CacheCfg(), rmtDBConns, rplConns)
		defer dm.DataDB().Close()
	}

	if *fromStorDB || *toStorDB {
		if storDb, err = engine.NewStorDBConn(ldrCfg.StorDbCfg().Type,
			ldrCfg.StorDbCfg().Host, ldrCfg.StorDbCfg().Port,
			ldrCfg.StorDbCfg().Name, ldrCfg.StorDbCfg().User,
			ldrCfg.StorDbCfg().Password, ldrCfg.StorDbCfg().SSLMode,
			ldrCfg.StorDbCfg().MaxOpenConns, ldrCfg.StorDbCfg().MaxIdleConns,
			ldrCfg.StorDbCfg().ConnMaxLifetime, ldrCfg.StorDbCfg().StringIndexedFields,
			ldrCfg.StorDbCfg().PrefixIndexedFields); err != nil {
			log.Fatalf("Coud not open storDB connection: %s", err.Error())
		}
		defer storDb.Close()
	}

	if !*dryRun {
		//tpid_remove
		if *toStorDB { // Import files from a directory into storDb
			if ldrCfg.LoaderCgrCfg().TpID == "" {
				log.Fatal("TPid required.")
			}
			if *flushStorDB {
				if err = storDb.RemTpData("", ldrCfg.LoaderCgrCfg().TpID, map[string]string{}); err != nil {
					log.Fatal(err)
				}
			}
			csvImporter := engine.TPCSVImporter{
				TPid:     ldrCfg.LoaderCgrCfg().TpID,
				StorDb:   storDb,
				DirPath:  *dataPath,
				Sep:      ldrCfg.LoaderCgrCfg().FieldSeparator,
				Verbose:  *verbose,
				ImportId: *importID,
			}
			if errImport := csvImporter.Run(); errImport != nil {
				log.Fatal(errImport)
			}
			return
		}
	}

	if *fromStorDB { // Load Tariff Plan from storDb into dataDb
		loader = storDb
	} else if gprefix := utils.MetaGoogleAPI + utils.CONCATENATED_KEY_SEP; strings.HasPrefix(*dataPath, gprefix) { // Default load from csv files to dataDb
		loader, err = engine.NewGoogleCSVStorage(ldrCfg.LoaderCgrCfg().FieldSeparator, strings.TrimPrefix(*dataPath, gprefix), *cfgPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		loader = engine.NewFileCSVStorage(ldrCfg.LoaderCgrCfg().FieldSeparator, *dataPath, *recursive)
	}

	if len(ldrCfg.LoaderCgrCfg().CachesConns) != 0 { // Init connection to CacheS so we can reload it's data
		if cacheS, err = rpcclient.NewRpcClient("tcp",
			ldrCfg.LoaderCgrCfg().CachesConns[0].Address,
			ldrCfg.LoaderCgrCfg().CachesConns[0].TLS, ldrCfg.TlsCfg().ClientKey,
			ldrCfg.TlsCfg().ClientCerificate, ldrCfg.TlsCfg().CaCertificate, 3, 3,
			time.Duration(1*time.Second), time.Duration(5*time.Minute),
			strings.TrimPrefix(ldrCfg.LoaderCgrCfg().CachesConns[0].Transport, utils.Meta),
			nil, false); err != nil {
			log.Fatalf("Could not connect to CacheS: %s", err.Error())
			return
		}
	} else {
		log.Print("WARNING: automatic cache reloading is disabled!")
	}

	if len(ldrCfg.LoaderCgrCfg().SchedulerConns) != 0 { // Init connection to Scheduler so we can reload it's data
		if schedulerS, err = rpcclient.NewRpcClient("tcp",
			ldrCfg.LoaderCgrCfg().SchedulerConns[0].Address,
			ldrCfg.LoaderCgrCfg().SchedulerConns[0].TLS, ldrCfg.TlsCfg().ClientKey,
			ldrCfg.TlsCfg().ClientCerificate, ldrCfg.TlsCfg().CaCertificate, 3, 3,
			time.Duration(1*time.Second), time.Duration(5*time.Minute),
			strings.TrimPrefix(ldrCfg.LoaderCgrCfg().SchedulerConns[0].Transport, utils.Meta),
			nil, false); err != nil {
			log.Fatalf("Could not connect to Scheduler: %s", err.Error())
			return
		}
	}

	tpReader, err := engine.NewTpReader(dm.DataDB(), loader, ldrCfg.LoaderCgrCfg().TpID,
		ldrCfg.GeneralCfg().DefaultTimezone, cacheS, schedulerS)
	if err != nil {
		log.Fatal(err)
	}
	if err = tpReader.LoadAll(); err != nil {
		log.Fatal(err)
	}

	if *dryRun { // We were just asked to parse the data, not saving it
		return
	}

	if !*remove {
		// write maps to database
		if err := tpReader.WriteToDatabase(*verbose, *disableReverse); err != nil {
			log.Fatal("Could not write to database: ", err)
		}
		caching := config.CgrConfig().GeneralCfg().DefaultCaching
		if cachingArg != nil && *cachingArg != utils.EmptyString {
			caching = *cachingArg
		}
		// reload cache
		if err := tpReader.ReloadCache(caching, *verbose, &utils.ArgDispatcher{
			APIKey:  apiKey,
			RouteID: routeID,
		}); err != nil {
			log.Fatal("Could not reload cache: ", err)
		}
	} else {
		if err := tpReader.RemoveFromDatabase(*verbose, *disableReverse); err != nil {
			log.Fatal("Could not delete from database: ", err)
		}
	}
}
