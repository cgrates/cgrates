/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to this file")
	runs       = flag.Int("runs", 10000, "stress cycle number")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}
	t1 := time.Date(2013, time.August, 07, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2013, time.August, 07, 18, 30, 0, 0, time.UTC)
	//cd := engine.CallDescriptor{Direction: "*out", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cd := engine.CallDescriptor{Direction: "*out", TOR: "call", Tenant: "cgrates.org", Subject: "1001", Destination: "+49", TimeStart: t1, TimeEnd: t2}

	getter, err := engine.ConfigureDataStorage(utils.REDIS, "localhost", "6379", "", "", "", utils.MSGPACK)
	//getter, err := engine.NewMongoStorage("localhost", "cgrates")
	if err != nil {
		log.Fatal("Could not connect to data store: ", err)
	}
	defer getter.Close()

	engine.SetDataStorage(getter)

	log.Printf("Runnning %d cycles...", *runs)
	var result *engine.CallCost
	j := 0
	start := time.Now()
	for i := 0; i < *runs; i++ {
		result, err = cd.GetCost()
		j = i
	}
	duration := time.Since(start)
	log.Print(result, j, err)
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	log.Printf("memstats before GC: Kbytes = %d footprint = %d",
		memstats.HeapAlloc/1024, memstats.Sys/1024)
	log.Printf("Elapsed: %v resulted: %v req/s.", duration, float64(*runs)/duration.Seconds())
}
