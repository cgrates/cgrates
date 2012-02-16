package main

import (
	"flag"
	"github.com/simonz05/godis"
	"github.com/fsouza/gokabinet/kc"
    "launchpad.net/mgo"
	"github.com/rif/cgrates/timespans"
	"log"
	"os"
	"encoding/json"
)

var (
	storage = flag.String("storage", "kyoto", "kyoto|redis|mongo")
	kyotofile = flag.String("kyotofile", "storage.kch", "kyoto storage file (storage.kch)")
	redisserver = flag.String("redisserver", "tcp:127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	mongoserver = flag.String("mongoserver", "127.0.0.1:27017", "mongo server address (127.0.0.1:27017)")
	redisdb = flag.Int("rdb", 10, "redis database number (10)")
	mongodb = flag.String("mdb", "test", "mongo database name (test)")
	redispass = flag.String("pass", "", "redis database password")
	inputfile = flag.String("inputfile", "input.json", "redis database password")
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
		db, _ := kc.Open(*kyotofile, kc.WRITE)
		defer db.Close()
		for _, cd := range callDescriptors{
			key := cd.GetKey()
			db.Set(key, cd.EncodeValues())
			log.Printf("Storing %q", key)
		}
	case "mongo":
		 session, err := mgo.Dial(*mongoserver)
        if err != nil {
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Strong, true)

        c := session.DB(*mongodb).C("ap")
        for _, cd := range callDescriptors{
			key := cd.GetKey()
			c.Insert(&timespans.KeyValue{key, cd.EncodeValues()})
			log.Printf("Storing %q", key)
		}

	default:
		db := godis.New(*redisserver, *redisdb, *redispass)
		defer db.Quit()
		for _, cd := range callDescriptors{
			key := cd.GetKey()
			db.Set(key, cd.EncodeValues())
			log.Printf("Storing %q", key)
		}
	}
}
