package main

import (
	"encoding/json"
	"flag"
	"github.com/rif/cgrates/timespans"
	"log"
	"os"
)

var (
	storage     = flag.String("storage", "kyoto", "kyoto|redis|mongo")
	kyotofile   = flag.String("kyotofile", "storage.kch", "kyoto storage file (storage.kch)")
	redisserver = flag.String("redisserver", "tcp:127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb     = flag.Int("rdb", 10, "redis database number (10)")
	mongoserver = flag.String("mongoserver", "127.0.0.1:27017", "mongo server address (127.0.0.1:27017)")
	mongodb     = flag.String("mdb", "test", "mongo database name (test)")
	redispass   = flag.String("pass", "", "redis database password")
	apfile   = flag.String("apfile", "ap.json", "Activation Periods containing intervals file")
	destfile   = flag.String("destfile", "dest.json", "Destinations file")
)

func writeToStorage(storage timespans.StorageGetter,callDescriptors []*timespans.CallDescriptor,destinations []*timespans.Destination){
	for _, cd := range callDescriptors {
		storage.SetActivationPeriods(cd.GetKey(), cd.ActivationPeriods)
		log.Printf("Storing %q", cd.GetKey())
	}
	for _, d := range destinations {
		storage.SetDestination(d)
		log.Printf("Storing %q", d.Id)
	}
}

func main() {
	flag.Parse()


	log.Print("Reading from ", *apfile, *destfile)

	// reading activation periods
	fin, err := os.Open(*apfile)

	if err != nil {log.Print("Cannot open activation periods input file", err)}

	dec := json.NewDecoder(fin)

	var callDescriptors []*timespans.CallDescriptor
	if err := dec.Decode(&callDescriptors); err != nil {
		log.Println(err)
		return
	}
	fin.Close()

	// reading destinations
	fin, err = os.Open(*destfile)

	if err != nil {log.Print("Cannot open destinations input file", err)}

	dec = json.NewDecoder(fin)

	var destinations []*timespans.Destination
	if err := dec.Decode(&destinations); err != nil {
		log.Println(err)
		return
	}
	fin.Close()

	switch *storage {
	case "kyoto":
		storage, _ := timespans.NewKyotoStorage(*kyotofile)
		defer storage.Close()
		writeToStorage(storage, callDescriptors, destinations)
	case "mongo":
		storage, _ := timespans.NewMongoStorage("127.0.0.1", "test")
		defer storage.Close()
		writeToStorage(storage, callDescriptors, destinations)

	default:
		storage, _ := timespans.NewRedisStorage(*redisserver, *redisdb)
		defer storage.Close()
		writeToStorage(storage, callDescriptors, destinations)
	}
}
