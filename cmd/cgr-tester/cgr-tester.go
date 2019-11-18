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
	"math"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cgrTesterFlags = flag.NewFlagSet("cgr-tester", flag.ContinueOnError)
	cgrConfig, _   = config.NewDefaultCGRConfig()
	tstCfg         = config.CgrConfig()
	cpuprofile     = cgrTesterFlags.String("cpuprofile", "", "write cpu profile to file")
	memprofile     = cgrTesterFlags.String("memprofile", "", "write memory profile to this file")
	runs           = cgrTesterFlags.Int("runs", int(math.Pow10(5)), "stress cycle number")

	cfgPath = cgrTesterFlags.String("config_path", "",
		"Configuration directory path.")

	parallel       = cgrTesterFlags.Int("parallel", 0, "run n requests in parallel")
	datadbType     = cgrTesterFlags.String("datadb_type", cgrConfig.DataDbCfg().DataDbType, "The type of the DataDb database <redis>")
	datadbHost     = cgrTesterFlags.String("datadb_host", cgrConfig.DataDbCfg().DataDbHost, "The DataDb host to connect to.")
	datadbPort     = cgrTesterFlags.String("datadb_port", cgrConfig.DataDbCfg().DataDbPort, "The DataDb port to bind to.")
	datadbName     = cgrTesterFlags.String("datadb_name", cgrConfig.DataDbCfg().DataDbName, "The name/number of the DataDb to connect to.")
	datadbUser     = cgrTesterFlags.String("datadb_user", cgrConfig.DataDbCfg().DataDbUser, "The DataDb user to sign in as.")
	datadbPass     = cgrTesterFlags.String("datadb_pass", cgrConfig.DataDbCfg().DataDbPass, "The DataDb user's password.")
	dbdataEncoding = cgrTesterFlags.String("dbdata_encoding", cgrConfig.GeneralCfg().DBDataEncoding, "The encoding used to store object data in strings.")
	redisSentinel  = cgrTesterFlags.String("redis_sentinel", cgrConfig.DataDbCfg().DataDbSentinelName, "The name of redis sentinel")
	raterAddress   = cgrTesterFlags.String("rater_address", "", "Rater address for remote tests. Empty for internal rater.")
	tor            = cgrTesterFlags.String("tor", utils.VOICE, "The type of record to use in queries.")
	category       = cgrTesterFlags.String("category", "call", "The Record category to test.")
	tenant         = cgrTesterFlags.String("tenant", "cgrates.org", "The type of record to use in queries.")
	subject        = cgrTesterFlags.String("subject", "1001", "The rating subject to use in queries.")
	destination    = cgrTesterFlags.String("destination", "1002", "The destination to use in queries.")
	json           = cgrTesterFlags.Bool("json", false, "Use JSON RPC")
	version        = cgrTesterFlags.Bool("version", false, "Prints the application version.")
	nilDuration    = time.Duration(0)
	usage          = cgrTesterFlags.String("usage", "1m", "The duration to use in call simulation.")
	fPath          = cgrTesterFlags.String("file_path", "", "read requests from file with path")
	reqSep         = cgrTesterFlags.String("req_separator", "\n\n", "separator for requests in file")

	err error
)

func durInternalRater(cd *engine.CallDescriptorWithArgDispatcher) (time.Duration, error) {
	dbConn, err := engine.NewDataDBConn(tstCfg.DataDbCfg().DataDbType,
		tstCfg.DataDbCfg().DataDbHost, tstCfg.DataDbCfg().DataDbPort,
		tstCfg.DataDbCfg().DataDbName, tstCfg.DataDbCfg().DataDbUser,
		tstCfg.DataDbCfg().DataDbPass, tstCfg.GeneralCfg().DBDataEncoding,
		tstCfg.DataDbCfg().DataDbSentinelName)
	if err != nil {
		return nilDuration, fmt.Errorf("Could not connect to data database: %s", err.Error())
	}
	dm := engine.NewDataManager(dbConn, cgrConfig.CacheCfg(), nil, nil) // for the momentn we use here "" for sentinelName
	defer dm.DataDB().Close()
	engine.SetDataStorage(dm)
	if err := dm.LoadDataDBCache(nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		return nilDuration, fmt.Errorf("Cache rating error: %s", err.Error())
	}
	log.Printf("Runnning %d cycles...", *runs)
	var result *engine.CallCost
	start := time.Now()
	for i := 0; i < *runs; i++ {
		result, err = cd.GetCost()
		if *memprofile != "" {
			runtime.MemProfileRate = 1
			runtime.GC()
			f, err := os.Create(*memprofile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.WriteHeapProfile(f)
			f.Close()
			break
		}
	}
	log.Printf("Result:%s\n", utils.ToJSON(result))

	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	log.Printf("memstats before GC: Kbytes = %d footprint = %d",
		memstats.HeapAlloc/1024, memstats.Sys/1024)
	return time.Since(start), nil
}

func durRemoteRater(cd *engine.CallDescriptorWithArgDispatcher) (time.Duration, error) {
	result := engine.CallCost{}
	var client *rpc.Client
	var err error
	if *json {
		client, err = jsonrpc.Dial("tcp", *raterAddress)
	} else {
		client, err = rpc.Dial("tcp", *raterAddress)
	}

	if err != nil {
		return nilDuration, fmt.Errorf("Could not connect to engine: %s", err.Error())
	}
	defer client.Close()
	start := time.Now()
	if *parallel > 0 {
		// var divCall *rpc.Call
		var sem = make(chan int, *parallel)
		var finish = make(chan int)
		for i := 0; i < *runs; i++ {
			go func() {
				sem <- 1
				client.Call(utils.ResponderGetCost, cd, &result)
				<-sem
				finish <- 1
				// divCall = client.Go(utils.ResponderGetCost, cd, &result, nil)
			}()
		}
		for i := 0; i < *runs; i++ {
			<-finish
		}
		// <-divCall.Done
	} else {
		for j := 0; j < *runs; j++ {
			client.Call(utils.ResponderGetCost, cd, &result)
		}
	}
	log.Printf("Result:%s\n", utils.ToJSON(result))
	return time.Since(start), nil
}

func main() {
	if err := cgrTesterFlags.Parse(os.Args[1:]); err != nil {
		return
	}
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}

	if *cfgPath != "" {
		if tstCfg, err = config.NewCGRConfigFromPath(*cfgPath); err != nil {
			log.Fatalf("error loading config file %s", err.Error())
		}
	}

	if *datadbType != cgrConfig.DataDbCfg().DataDbType {
		tstCfg.DataDbCfg().DataDbType = *datadbType
	}
	if *datadbHost != cgrConfig.DataDbCfg().DataDbHost {
		tstCfg.DataDbCfg().DataDbHost = *datadbHost
	}
	if *datadbPort != cgrConfig.DataDbCfg().DataDbPort {
		tstCfg.DataDbCfg().DataDbPort = *datadbPort
	}
	if *datadbName != cgrConfig.DataDbCfg().DataDbName {
		tstCfg.DataDbCfg().DataDbName = *datadbName
	}
	if *datadbUser != cgrConfig.DataDbCfg().DataDbUser {
		tstCfg.DataDbCfg().DataDbUser = *datadbUser
	}
	if *datadbPass != cgrConfig.DataDbCfg().DataDbPass {
		tstCfg.DataDbCfg().DataDbPass = *datadbPass
	}
	if *dbdataEncoding != "" {
		tstCfg.GeneralCfg().DBDataEncoding = *dbdataEncoding
	}
	if *redisSentinel != "" {
		tstCfg.DataDbCfg().DataDbSentinelName = *redisSentinel
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *fPath != "" {
		frt, err := NewFileReaderTester(*fPath, *raterAddress,
			*parallel, *runs, []byte(*reqSep))
		if err != nil {
			log.Fatal(err)
		}
		if err := frt.Test(); err != nil {
			log.Fatal(err)
		}
		return
	}

	var timeparsed time.Duration
	var err error
	tstart := time.Now()
	timeparsed, err = utils.ParseDurationWithNanosecs(*usage)
	tend := tstart.Add(timeparsed)
	cd := &engine.CallDescriptorWithArgDispatcher{
		CallDescriptor: &engine.CallDescriptor{
			TimeStart:     tstart,
			TimeEnd:       tend,
			DurationIndex: 60 * time.Second,
			TOR:           *tor,
			Category:      *category,
			Tenant:        *tenant,
			Subject:       *subject,
			Destination:   *destination,
		},
	}
	var duration time.Duration
	if len(*raterAddress) == 0 {
		duration, err = durInternalRater(cd)
	} else {
		duration, err = durRemoteRater(cd)
	}
	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Printf("Elapsed: %s resulted: %f req/s.", duration, float64(*runs)/duration.Seconds())
	}
}
