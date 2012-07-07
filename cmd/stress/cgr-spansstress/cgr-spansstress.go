/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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
	"github.com/cgrates/cgrates/timespans"
	"log"
	"os"
	"runtime/pprof"
	"time"
	"runtime"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
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

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := timespans.CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}

	getter, err := timespans.NewRedisStorage("", 10)
	defer getter.Close()

	timespans.SetStorageGetter(getter)

	log.Printf("Runnning %d cycles...", *runs)
	var result *timespans.CallCost
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
