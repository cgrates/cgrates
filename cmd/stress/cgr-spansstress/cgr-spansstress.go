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
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
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
	cd := engine.CallDescriptor{
		TimeStart:    time.Date(2013, time.December, 13, 22, 30, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, time.December, 13, 22, 31, 0, 0, time.UTC),
		CallDuration: 60 * time.Second,
		Direction:    "*out",
		TOR:          "call",
		Tenant:       "cgrates.org",
		Subject:      "1001",
		Destination:  "+49676016500",
	}
	getter, err := engine.ConfigureDataStorage(utils.REDIS, "127.0.0.1", "6379", "10", "", "", utils.MSGPACK)
	if err != nil {
		log.Fatal("Could not connect to data store: ", err)
	}
	defer getter.Close()

	engine.SetDataStorage(getter)
	if err := getter.PreCache(nil, nil, nil, nil); err != nil {
		log.Printf("Pre-caching error: %v", err)
		return
	}

	log.Printf("Runnning %d cycles...", *runs)
	var result *engine.CallCost
	j := 0
	start := time.Now()
	for i := 0; i < *runs; i++ {
		result, err = cd.Debit()
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
	duration := time.Since(start)
	log.Print(result, j, err)
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	log.Printf("memstats before GC: Kbytes = %d footprint = %d",
		memstats.HeapAlloc/1024, memstats.Sys/1024)
	log.Printf("Elapsed: %v resulted: %v req/s.", duration, float64(*runs)/duration.Seconds())
}
