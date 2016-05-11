package main

import (
	"flag"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

func main() {
	flag.Parse()
	var err error
	var cdrsMasterRpc *rpcclient.RpcClient
	var cdrsMasterCfgPath string
	var cdrsMasterCfg *config.CGRConfig
	cdrsMasterCfgPath = path.Join(*dataDir, "conf", "samples", "cdrsreplicationmaster")
	if cdrsMasterCfg, err = config.NewCGRConfigFromFolder(cdrsMasterCfgPath); err != nil {
		log.Fatal("Got config error: ", err.Error())
	}
	cdrsMasterRpc, err = rpcclient.NewRpcClient("tcp", cdrsMasterCfg.RPCJSONListen, 1, 1, "json", nil)
	if err != nil {
		log.Fatal("Could not connect to rater: ", err.Error())
	}
	cdrs := make([]*engine.CDR, 0)
	for i := 0; i < 10000; i++ {
		cdr := &engine.CDR{OriginID: fmt.Sprintf("httpjsonrpc_%d", i),
			ToR: utils.VOICE, OriginHost: "192.168.1.1", Source: "UNKNOWN", RequestType: utils.META_PSEUDOPREPAID,
			Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
			SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
		cdrs = append(cdrs, cdr)
	}
	var reply string
	for _, cdr := range cdrs {
		if err := cdrsMasterRpc.Call("CdrsV2.ProcessCdr", cdr, &reply); err != nil {
			log.Fatal("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			log.Fatal("Unexpected reply received: ", reply)
		}
	}
}
