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
	inputfile   = flag.String("inputfile", "input.json", "redis database password")
)

func main() {
	flag.Parse()

	log.Printf("Reading from %q", *inputfile)

	fin, err := os.Open(*inputfile)
	defer fin.Close()

	if err != nil {
		log.Print("Cannot open input file", err)
		return
	}

	dec := json.NewDecoder(fin)

	var callDescriptors []*timespans.CallDescriptor
	if err := dec.Decode(&callDescriptors); err != nil {
		log.Println(err)
		return
	}

	switch *storage {
	case "kyoto":
		storage, _ := timespans.NewKyotoStorage(*kyotofile)
		defer storage.Close()
		for _, cd := range callDescriptors {
			storage.SetActivationPeriods(cd.GetKey(), cd.ActivationPeriods)
			log.Printf("Storing %q", cd.GetKey())
		}
	case "mongo":
		storage, _ := timespans.NewMongoStorage("127.0.0.1", "test")
		defer storage.Close()
		for _, cd := range callDescriptors {
			storage.SetActivationPeriods(cd.GetKey(), cd.ActivationPeriods)
			log.Printf("Storing %q", cd.GetKey())
		}

	default:
		storage, _ := timespans.NewRedisStorage(*redisserver, *redisdb)
		defer storage.Close()
		for _, cd := range callDescriptors {
			storage.SetActivationPeriods(cd.GetKey(), cd.ActivationPeriods)
			log.Printf("Storing %q", cd.GetKey())
		}
	}
}
