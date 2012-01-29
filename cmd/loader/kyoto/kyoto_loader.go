package main

import (
	"fmt"
	"github.com/fsouza/gokabinet/kc"
	"flag"
)

var (
	fileName = flag.String("fileName", "storage.kch", "kyoto storage file")
)

func main() {
	flag.Parse()
	db, _ := kc.Open(*fileName, kc.WRITE)
	defer db.Close()
	
   	db.Set("test", "12223")
   	fmt.Println("Done!")
}

