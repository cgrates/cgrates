/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

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
	cdrsMasterRpc, err = rpcclient.NewRpcClient("tcp", cdrsMasterCfg.ListenCfg().RPCJSONListen, false, "", "", "", 1, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), "json", nil, false)
	if err != nil {
		log.Fatal("Could not connect to rater: ", err.Error())
	}
	cdrs := make([]*engine.CDR, 0)
	for i := 0; i < 10000; i++ {
		cdr := &engine.CDR{OriginID: fmt.Sprintf("httpjsonrpc_%d", i),
			ToR: utils.VOICE, OriginHost: "192.168.1.1", Source: "UNKNOWN", RequestType: utils.META_PSEUDOPREPAID,
			Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
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
