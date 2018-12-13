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

package general_tests

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/apier/v1"
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
	cdreConfigDIR string
)

var sTestsCDRE = []func(t *testing.T){
	testCDREInitCfg,
	testCDREInitDataDb,
	testCDREResetStorDb,
	testCDREStartEngine,
	testCDRERpcConn,
	testCDREGetCdrs,
	testCDREExportNotFound,
	testCDREProcessCdr,
	testCDREExport,
	testCDREStopEngine,
}

func TestCDREITMySql(t *testing.T) {
	cdreConfigDIR = "tutmysql"
	for _, stest := range sTestsCDRE {
		t.Run(cdreConfigDIR, stest)
	}
}

func TestCDREITMongo(t *testing.T) {
	cdreConfigDIR = "tutmongonew"
	for _, stest := range sTestsCDRE {
		t.Run(cdreConfigDIR, stest)
	}
}

func testCDREInitCfg(t *testing.T) {
	var err error
	cdreCfgPath = path.Join(cdreDataDir, "conf", "samples", cdreConfigDIR)
	cdreCfg, err = config.NewCGRConfigFromFolder(cdreCfgPath)
	if err != nil {
		t.Error(err)
	}
	cdreCfg.DataFolderPath = cdreDataDir
	config.SetCgrConfig(cdreCfg)
	cdreDelay = 1000
}

func testCDREInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdreCfg); err != nil {
		t.Fatal(err)
	}
}

func testCDREResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(cdreCfg); err != nil {
		t.Fatal(err)
	}
}

func testCDREStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdreCfgPath, cdreDelay); err != nil {
		t.Fatal(err)
	}
}

func testCDRERpcConn(t *testing.T) {
	var err error
	cdreRPC, err = jsonrpc.Dial("tcp", cdreCfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func testCDREGetCdrs(t *testing.T) {
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{}
	if err := cdreRPC.Call("ApierV1.GetCdrs", req, &reply); err.Error() != utils.ErrNotFound.Error() {
		t.Error("Unexpected error: ", err.Error())
	}
}

func testCDREExportNotFound(t *testing.T) {
	var replyExport v1.RplExportedCDRs
	exportArgs := v1.ArgExportCDRs{
		ExportPath:     utils.StringPointer("/tmp"),
		ExportFileName: utils.StringPointer("TestTutITExportCDR.csv"),
		ExportTemplate: utils.StringPointer("TestTutITExportCDR"),
		RPCCDRsFilter:  utils.RPCCDRsFilter{},
	}
	if err := cdreRPC.Call("ApierV1.ExportCDRs", exportArgs,
		&replyExport); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testCDREProcessCdr(t *testing.T) {
	cdr := &engine.CDR{ToR: utils.VOICE, OriginID: "testCDREProcessCdr", OriginHost: "192.168.1.1",
		Source: "TestTutITExportCDR", RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001",
		Subject: "1001", Destination: "1003",
		SetupTime:   time.Date(2016, 11, 30, 17, 5, 24, 0, time.UTC),
		AnswerTime:  time.Date(2016, 11, 30, 17, 6, 4, 0, time.UTC),
		Usage:       time.Duration(98) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	cdr.ComputeCGRID()
	var reply string
	if err := cdreRPC.Call(utils.CdrsV2ProcessCDR, cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testCDREExport(t *testing.T) {
	// time.Sleep(100 * time.Millisecond)
	var replyExport v1.RplExportedCDRs
	exportArgs := v1.ArgExportCDRs{
		ExportPath:     utils.StringPointer("/tmp"),
		ExportFileName: utils.StringPointer("TestTutITExportCDR.csv"),
		ExportTemplate: utils.StringPointer("TestTutITExportCDR"),
		RPCCDRsFilter:  utils.RPCCDRsFilter{},
	}
	if err := cdreRPC.Call("ApierV1.ExportCDRs", exportArgs, &replyExport); err != nil {
		t.Error(err)
	} else if replyExport.TotalRecords != 1 {
		t.Errorf("Unexpected total records: %+v", replyExport.TotalRecords)
	}
	// expFilePath := path.Join(*exportArgs.ExportPath, *exportArgs.ExportFileName)
	// if err := os.Remove(expFilePath); err != nil {
	// 	t.Error(err)
	// }
}

func testCDREStopEngine(t *testing.T) {
	if err := engine.KillEngine(cdreDelay); err != nil {
		t.Error(err)
	}
}
