package main

import (
	"flag"	
	"github.com/simonz05/godis"
	"github.com/fsouza/gokabinet/kc"
	"github.com/rif/cgrates/timespans"	
	"log"
	"os"	
	"encoding/json"		
)

var (	
	storage = flag.String("storage", "kyoto", "kyoto | redis")
	kyotofile = flag.String("kyotofile", "storage.kch", "kyoto storage file (storage.kch)")
	redisserver = flag.String("server", "tcp:127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb = flag.Int("db", 10, "redis database number (10)")
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

	if *storage == "kyoto" {
		db, _ := kc.Open(*kyotofile, kc.WRITE)
		defer db.Close()
		for _, cd := range callDescriptors{
			key := cd.GetKey()
			db.Set(key, cd.EncodeValues())
			log.Printf("Storing %q", key)
		}	
	} else {		
		db := godis.New(*redisserver, *redisdb, *redispass)
		defer db.Quit()
		for _, cd := range callDescriptors{
			key := cd.GetKey()
			db.Set(key, cd.EncodeValues())
			log.Printf("Storing %q", key)
		}		
	}	
}
