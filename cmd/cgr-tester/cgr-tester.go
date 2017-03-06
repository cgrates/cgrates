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
	cgrConfig, _    = config.NewDefaultCGRConfig()
	cpuprofile      = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile      = flag.String("memprofile", "", "write memory profile to this file")
	runs            = flag.Int("runs", 10000, "stress cycle number")
	parallel        = flag.Int("parallel", 0, "run n requests in parallel")
	datadb_type     = flag.String("datadb_type", cgrConfig.DataDbType, "The type of the DataDb database <redis>")
	datadb_host     = flag.String("datadb_host", cgrConfig.DataDbHost, "The DataDb host to connect to.")
	datadb_port     = flag.String("datadb_port", cgrConfig.DataDbPort, "The DataDb port to bind to.")
	datadb_name     = flag.String("datadb_name", cgrConfig.DataDbName, "The name/number of the DataDb to connect to.")
	datadb_user     = flag.String("datadb_user", cgrConfig.DataDbUser, "The DataDb user to sign in as.")
	datadb_pass     = flag.String("datadb_pass", cgrConfig.DataDbPass, "The DataDb user's password.")
	dbdata_encoding = flag.String("dbdata_encoding", cgrConfig.DBDataEncoding, "The encoding used to store object data in strings.")
	raterAddress    = flag.String("rater_address", "", "Rater address for remote tests. Empty for internal rater.")
	tor             = flag.String("tor", utils.VOICE, "The type of record to use in queries.")
	category        = flag.String("category", "call", "The Record category to test.")
	tenant          = flag.String("tenant", "cgrates.org", "The type of record to use in queries.")
	subject         = flag.String("subject", "1001", "The rating subject to use in queries.")
	destination     = flag.String("destination", "1002", "The destination to use in queries.")
	json            = flag.Bool("json", false, "Use JSON RPC")
	loadHistorySize = flag.Int("load_history_size", cgrConfig.LoadHistorySize, "Limit the number of records in the load history")
	version         = flag.Bool("version", false, "Prints the application version.")
	nilDuration     = time.Duration(0)
)

func durInternalRater(cd *engine.CallDescriptor) (time.Duration, error) {
	dataDb, err := engine.ConfigureDataStorage(*datadb_type, *datadb_host, *datadb_port, *datadb_name, *datadb_user, *datadb_pass, *dbdata_encoding, cgrConfig.CacheConfig, *loadHistorySize)
	if err != nil {
		return nilDuration, fmt.Errorf("Could not connect to data database: %s", err.Error())
	}
	defer dataDb.Close()
	engine.SetDataStorage(dataDb)
	if err := dataDb.LoadRatingCache(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
		return nilDuration, fmt.Errorf("Cache rating error: %s", err.Error())
	}
	if err := dataDb.LoadAccountingCache(nil, nil, nil); err != nil {
		return nilDuration, fmt.Errorf("Cache accounting error: %s", err.Error())
	}
	log.Printf("Runnning %d cycles...", *runs)
	var result *engine.CallCost
	j := 0
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
		j = i
	}
	log.Print(result, j, err)
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	log.Printf("memstats before GC: Kbytes = %d footprint = %d",
		memstats.HeapAlloc/1024, memstats.Sys/1024)
	return time.Since(start), nil
}

func durRemoteRater(cd *engine.CallDescriptor) (time.Duration, error) {
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
				client.Call("Responder.GetCost", cd, &result)
				<-sem
				finish <- 1
				// divCall = client.Go("Responder.GetCost", cd, &result, nil)
			}()
		}
		for i := 0; i < *runs; i++ {
			<-finish
		}
		// <-divCall.Done
	} else {
		for j := 0; j < *runs; j++ {
			client.Call("Responder.GetCost", cd, &result)
		}
	}
	log.Println(result)
	return time.Since(start), nil
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println(utils.GetCGRVersion())
		return
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	cd := &engine.CallDescriptor{
		TimeStart:     time.Date(2014, time.December, 11, 55, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, time.December, 11, 55, 31, 0, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		Direction:     "*out",
		TOR:           *tor,
		Category:      *category,
		Tenant:        *tenant,
		Subject:       *subject,
		Destination:   *destination,
	}
	var duration time.Duration
	var err error
	if len(*raterAddress) == 0 {
		duration, err = durInternalRater(cd)
	} else {
		duration, err = durRemoteRater(cd)
	}
	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Printf("Elapsed: %d resulted: %f req/s.", duration, float64(*runs)/duration.Seconds())
	}
}
