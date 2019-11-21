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
package ers

import (
	"flag"
	"io/ioutil"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	csvCfgPath string
	csvCfg     *config.CGRConfig
	csvRPC     *rpc.Client
	dataDir    = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	waitRater  = flag.Int("wait_rater", 500, "Number of miliseconds to wait for rater to start and cache")

	fileContent1 = `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,default,*voice,dsafdsaf,*rated,*out,cgrates.org,call,1001,1001,+4986517174963,2013-11-07 08:42:25 +0000 UTC,2013-11-07 08:42:26 +0000 UTC,10s,1.0100,val_extra3,"",val_extra1
dbafe9c8614c785a65aabd116dd3959c3c56f7f7,default,*voice,dsafdsag,*rated,*out,cgrates.org,call,1001,1001,+4986517174964,2013-11-07 09:42:25 +0000 UTC,2013-11-07 09:42:26 +0000 UTC,20s,1.0100,val_extra3,"",val_extra1
`

	fileContent2 = `accid21;*prepaid;itsyscom.com;1001;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*postpaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
#accid1;*pseudoprepaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;12;val_extra3;"";val_extra1
accid23;*rated;cgrates.org;1001;086517174963;2013-02-03 19:54:00;26;val_extra3;"";val_extra1`

	fileContent3 = `cgrates.org,*voice,SessionFromCsv,*prepaid,1001,ANY2CNT,1002,2018-01-07 17:00:00 +0000 UTC,2018-01-07 17:00:10 +0000 UTC,5m
`

	fileContentForFilter = `accid21;*prepaid;itsyscom.com;1002;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*postpaid;itsyscom.com;1002;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
accid23;*rated;cgrates.org;1001;086517174963;2013-02-03 19:54:00;26;val_extra3;"";val_extra1`

	csvTests = []func(t *testing.T){
		testCsvITCreateCdrDirs,
		testCsvITInitConfig,
		testCsvITInitCdrDb,
		testCsvITResetDataDb,
		testCsvITStartEngine,
		testCsvITRpcConn,
		testCsvITLoadTPFromFolder,
		testCsvITHandleCdr1File,
		testCsvITHandleCdr2File,
		testCsvITHandleSessionFile,
		testCsvITCheckSession,
		testCsvITTerminateSession,
		testCsvITProcessCDR,
		testCsvITAnalyseCDRs,
		testCsvITProcessFilteredCDR,
		testCsvITAnalyzeFilteredCDR,
		testCsvITProcessedFiles,
		testCsvITCleanupFiles,
		testCsvITKillEngine,
	}
)

func TestCsvReadFile(t *testing.T) {
	csvCfgPath = path.Join(*dataDir, "conf", "samples", "ers")
	for _, test := range csvTests {
		t.Run("TestCsvReadFile", test)
	}
}

func testCsvITInitConfig(t *testing.T) {
	var err error
	if csvCfg, err = config.NewCGRConfigFromPath(csvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testCsvITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testCsvITResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func testCsvITCreateCdrDirs(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out",
		"/tmp/ers2/in", "/tmp/ers2/out", "/tmp/init_session/in", "/tmp/init_session/out",
		"/tmp/terminate_session/in", "/tmp/terminate_session/out", "/tmp/cdrs/in",
		"/tmp/cdrs/out", "/tmp/ers_with_filters/in", "/tmp/ers_with_filters/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testCsvITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCsvITRpcConn(t *testing.T) {
	var err error
	csvRPC, err = jsonrpc.Dial("tcp", csvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testCsvITLoadTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	var loadInst utils.LoadInstance
	if err := csvRPC.Call(utils.ApierV2LoadTariffPlanFromFolder,
		attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

// The default scenario, out of cdrc defined in .cfg file
func testCsvITHandleCdr1File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/ers/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of first .xml config
func testCsvITHandleCdr2File(t *testing.T) {
	fileName := "file2.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent2), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/ers2/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of first .xml config
func testCsvITHandleSessionFile(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 10.0
	if err := csvRPC.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

	aSessions := make([]*sessions.ExternalSession, 0)
	if err := csvRPC.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{},
		&aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	fileName := "file3.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent3), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/init_session/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testCsvITCheckSession(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := csvRPC.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	}
}

func testCsvITTerminateSession(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	// move the file from init_session/out to terminate_session/in so the terminate session reader
	// can handle it
	if err := os.Rename("/tmp/init_session/out/file3.csv",
		"/tmp/terminate_session/in/file3.csv"); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
	time.Sleep(100 * time.Millisecond)
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := csvRPC.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testCsvITProcessCDR(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	// move the file from init_session/out to terminate_session/in so the terminate session reader
	// can handle it
	if err := os.Rename("/tmp/terminate_session/out/file3.csv",
		"/tmp/cdrs/in/file3.csv"); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testCsvITAnalyseCDRs(t *testing.T) {
	time.Sleep(500 * time.Millisecond)

	var cdrs []*engine.CDR
	args := utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"},
		OriginIDs: []string{"SessionFromCsv"}}
	if err := csvRPC.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.099 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"},
		OriginIDs: []string{"SessionFromCsv"}}
	if err := csvRPC.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.051 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.85
	if err := csvRPC.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func testCsvITProcessFilteredCDR(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContentForFilter), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/ers_with_filters/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testCsvITAnalyzeFilteredCDR(t *testing.T) {
	time.Sleep(500 * time.Millisecond)

	var cdrs []*engine.CDR
	args := utils.RPCCDRsFilter{NotRunIDs: []string{"CustomerCharges", "SupplierCharges"},
		Sources: []string{"ers_csv"}}
	if err := csvRPC.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", utils.ToJSON(cdrs))
	} else if cdrs[0].Account != "1002" || cdrs[1].Account != "1002" {
		t.Errorf("Expecting: 1002, received: <%s> , <%s>", cdrs[0].Account, cdrs[1].Account)
	} else if cdrs[0].Tenant != "itsyscom.com" || cdrs[1].Tenant != "itsyscom.com" {
		t.Errorf("Expecting: itsyscom.com, received: <%s> , <%s>", cdrs[0].Tenant, cdrs[1].Tenant)
	}
}

func testCsvITProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(1 * time.Second))
	if outContent1, err := ioutil.ReadFile("/tmp/ers/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContent1 != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", fileContent1, string(outContent1))
	}
	if outContent2, err := ioutil.ReadFile("/tmp/ers2/out/file2.csv"); err != nil {
		t.Error(err)
	} else if fileContent2 != string(outContent2) {
		t.Errorf("Expecting: %q, received: %q", fileContent2, string(outContent2))
	}
	if outContent3, err := ioutil.ReadFile("/tmp/cdrs/out/file3.csv"); err != nil {
		t.Error(err)
	} else if fileContent3 != string(outContent3) {
		t.Errorf("Expecting: %q, received: %q", fileContent3, string(outContent3))
	}
	if outContent4, err := ioutil.ReadFile("/tmp/ers_with_filters/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContentForFilter != string(outContent4) {
		t.Errorf("Expecting: %q, received: %q", fileContentForFilter, string(outContent4))
	}
}

func testCsvITCleanupFiles(t *testing.T) {
	for _, dir := range []string{"/tmp/ers",
		"/tmp/ers2", "/tmp/init_session",
		"/tmp/terminate_session", "/tmp/cdrs",
		"/tmp/ers_with_filters"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func testCsvITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
