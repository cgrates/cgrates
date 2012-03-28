package main

import (
	"flag"
	"github.com/rif/cgrates/timespans"
	"log"
	"net/rpc/jsonrpc"
	"os"
	"time"
)

var (
	balancer = flag.String("balancer", "127.0.0.1:2001", "balancer address host:port")
	tor      = flag.Int("tor", 0, "Type of record")
	cstmid   = flag.String("cstmid", "vdf", "Customer identificator")
	subject  = flag.String("subject", "rif", "The client who made the call")
	dest     = flag.String("dest", "0256", "Destination prefix")
	ts       = flag.String("ts", "2012-02-09T00:00:00Z", "Time start")
	te       = flag.String("te", "2012-02-09T00:10:00Z", "Time end")
)

func main() {
	flag.Parse()
	client, _ := jsonrpc.Dial("tcp", "localhost:2001")
	defer client.Close()

	var err error
	timestart, err := time.Parse(time.RFC3339, *ts)
	if err != nil {
		log.Fatal("Time start format is invalid: ", err)
	}
	timeend, err := time.Parse(time.RFC3339, *te)
	if err != nil {
		log.Fatal("Time end format is invalid: ", err)
	}

	cd := &timespans.CallDescriptor{TOR: *tor,
		CstmId:            *cstmid,
		Subject:           *subject,
		DestinationPrefix: *dest,
		TimeStart:         timestart,
		TimeEnd:           timeend,
	}
	result := timespans.CallCost{}

	switch flag.Arg(0) {
	case "getcost":
		if err = client.Call("Responder.GetCost", cd, &result); err == nil {
			log.Print(result)
		}
	default:
		log.Print("hello!")
	}
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
