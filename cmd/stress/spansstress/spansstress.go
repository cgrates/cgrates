package main

import (
	"flag"
	"github.com/rif/cgrates/timespans"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	runs       = flag.Int("runs", 10000, "stress cycle number")
)

func main() {
	flag.Parse()

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
	cd := timespans.CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}

	var result *timespans.CallCost

	getter, _ := timespans.NewRedisStorage("", 10)
	defer getter.Close()

	cd.StorageGetter = getter

	i := 0
	log.Printf("Runnning %d cycles...", *runs)

	for j := 0; j < *runs; j++ {
		result, _ = cd.GetCost()
	}

	log.Print(result)
	log.Print(i)
}
