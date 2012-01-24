package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	host = flag.String("host", "localhost:8080", "target host:port")
)

func main() {
	flag.Parse()
	fmt.Println(*host)
	log.Print("Bye!")
}
