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
package v2

import (
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdreCfgPath string
	cdreCfg     *config.CGRConfig
	cdreRpc     *rpc.Client
	cdreConfDIR string

	sTestsCDReIT = []func(t *testing.T){
		testV2CDReInitConfig,
		testV2CDReInitDataDb,
		testV2CDReInitCdrDb,
		testV2CDReStartEngine,
		testV2CDReRpcConn,
		testV2CDReLoadTariffPlanFromFolder,
		testV2CDReProcessCDR,
		testV2CDReGetCdrs,
		testV2CDReExportCdrs,
	}
)

func TestCDReIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		cdreConfDIR = "cdrsv2internal"
	case utils.MetaMySQL:
		cdreConfDIR = "cdrsv2mysql"
	case utils.MetaMongo:
		cdreConfDIR = "cdrsv2mongo"
	case utils.MetaPostgres:
		cdreConfDIR = "cdrsv2psql"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsCDReIT {
		t.Run(cdreConfDIR, stest)
	}
}

func testV2CDReInitConfig(t *testing.T) {
	var err error
	cdreCfgPath = path.Join(*dataDir, "conf", "samples", cdreConfDIR)
	if cdreCfg, err = config.NewCGRConfigFromPath(cdreCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testV2CDReInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdreCfg); err != nil {
		t.Fatal(err)
	}
}

func testV2CDReInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cdreCfg); err != nil {
		t.Fatal(err)
	}
}

func testV2CDReStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdreCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV2CDReRpcConn(t *testing.T) {
	cdreRpc, err = newRPCClient(cdreCfg.ListenCfg())
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV2CDReLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst utils.LoadInstance
	if err := cdreRpc.Call(utils.APIerSv2LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func testV2CDReProcessCDR(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDReProcessCDR1",
				utils.OriginHost:  "192.168.1.1",
				utils.Source:      "testV2CDReProcessCDR",
				utils.RequestType: utils.META_RATED,
				utils.Account:     "testV2CDRsProcessCDR",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       time.Minute,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
	}

	var reply string
	if err := cdreRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testV2CDReGetCdrs(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdreRpc.Call(utils.APIerSv2CountCDRs, req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := cdreRpc.Call(utils.APIerSv2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}}
	if err := cdreRpc.Call(utils.APIerSv2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}}
	if err := cdreRpc.Call(utils.APIerSv2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
}

func testV2CDReExportCdrs(t *testing.T) {
	attrExportCdrsToFile := &AttrExportCdrsToFile{
		Verbose:         false,
		ExportDirectory: utils.StringPointer("/tmp/"),
	}
	reply := &utils.ExportedFileCdrs{}

	if err := cdreRpc.Call(utils.APIerSv2ExportCdrsToFile, attrExportCdrsToFile, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply.ExportedCgrIds != nil {
		t.Error("Expecting ExportedCgrIds to be nil when verbose is false")
	} else if reply.UnexportedCgrIds != nil {
		t.Error("Expecting UnexportedCgrIds to be nil when verbose is false")
	}

	attrExportCdrsToFile = &AttrExportCdrsToFile{
		Verbose:         true,
		ExportDirectory: utils.StringPointer("/tmp/"),
	}
	reply = &utils.ExportedFileCdrs{}

	if err := cdreRpc.Call(utils.APIerSv2ExportCdrsToFile, attrExportCdrsToFile, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply.ExportedCgrIds == nil {
		t.Error("Expecting ExportedCgrIds to be different than nil when verbose is true")
	} else if reply.UnexportedCgrIds == nil {
		t.Error("Expecting UnexportedCgrIds to be different than nil when verbose is true")
	}

}
