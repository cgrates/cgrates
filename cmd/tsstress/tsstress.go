package main

import (
	"time"
	"log"
	"flag"
	"os"
	"runtime/pprof"
	"github.com/rif/cgrates/timespans"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

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

	i:= 0
	result := &timespans.CallCost{}

	getter, _ := timespans.NewKyotoStorage("storage.kch")
	defer getter.Close()
	
	for ; i < 1e5; i++ {
		
		result, _ = cd.GetCost(getter)
	}
	log.Print(result)
	log.Print(i)
}
