// +build integration

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

package v1

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdreCfgPath   string
	cdreCfg       *config.CGRConfig
	cdreRPC       *rpc.Client
	cdreDataDir   = "/usr/share/cgrates"
	cdreDelay     int
	cdreConfigDIR string //run tests for specific configuration
)

var sTestsCDRE = []func(t *testing.T){
	testCDReInitCfg,
	testCDReInitDataDb,
	testCDReResetStorDb,
	testCDReStartEngine,
	testCDReRPCConn,
	testCDReAddCDRs,
	testCDReExportCDRs,
	testCDReKillEngine,
}

//Test start here
func TestCDRExportMySql(t *testing.T) {
	cdreConfigDIR = "cdrewithfilter"
	for _, stest := range sTestsCDRE {
		t.Run(cdreConfigDIR, stest)
	}
}

func testCDReInitCfg(t *testing.T) {
	var err error
	cdreCfgPath = path.Join(alsPrfDataDir, "conf", "samples", cdreConfigDIR)
	cdreCfg, err = config.NewCGRConfigFromFolder(cdreCfgPath)
	if err != nil {
		t.Error(err)
	}
	cdreCfg.DataFolderPath = alsPrfDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(cdreCfg)
	cdreDelay = 2000
}

func testCDReInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdreCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testCDReResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(cdreCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testCDReStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdreCfgPath, cdreDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCDReRPCConn(t *testing.T) {
	var err error
	cdreRPC, err = jsonrpc.Dial("tcp", cdreCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testCDReAddCDRs(t *testing.T) {
	storedCdrs := []*engine.CDR{
		{CGRID: "Cdr1",
			OrderID: 123, ToR: utils.VOICE, OriginID: "OriginCDR1", OriginHost: "192.168.1.1", Source: "test",
			RequestType: utils.META_RATED, Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		{CGRID: "Cdr2",
			OrderID: 123, ToR: utils.VOICE, OriginID: "OriginCDR2", OriginHost: "192.168.1.1", Source: "test2",
			RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.DEFAULT_RUNID, Usage: time.Duration(5) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		{CGRID: "Cdr3",
			OrderID: 123, ToR: utils.VOICE, OriginID: "OriginCDR3", OriginHost: "192.168.1.1", Source: "test2",
			RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.DEFAULT_RUNID, Usage: time.Duration(30) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		{CGRID: "Cdr4",
			OrderID: 123, ToR: utils.VOICE, OriginID: "OriginCDR4", OriginHost: "192.168.1.1", Source: "test3",
			RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Time{}, RunID: utils.DEFAULT_RUNID, Usage: time.Duration(0) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	for _, cdr := range storedCdrs {
		var reply string
		if err := cdreRPC.Call("CdrsV1.ProcessCDR", cdr, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	time.Sleep(time.Duration(cdreDelay) * time.Millisecond)
}

func testCDReExportCDRs(t *testing.T) {
	attr := ArgExportCDRs{
		ExportTemplate: utils.StringPointer("TemplateWithFilter"),
		Verbose:        true,
	}
	var rply *RplExportedCDRs
	if err := cdreRPC.Call("ApierV1.ExportCDRs", attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rply.ExportedCGRIDs) != 4 {
		t.Errorf("Unexpected number of CDR exported: %s ", utils.ToJSON(rply))
	}
}

func testCDReKillEngine(t *testing.T) {
	if err := engine.KillEngine(cdreDelay); err != nil {
		t.Error(err)
	}
}
