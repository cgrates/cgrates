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
	"fmt"
	"net/rpc"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	csvCfgPath string
	csvCfgDIR  string
	csvCfg     *config.CGRConfig
	csvRPC     *rpc.Client

	fileContent1 = `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,default,*voice,dsafdsaf,*rated,*out,cgrates.org,call,1001,1001,+4986517174963,2013-11-07 08:42:25 +0000 UTC,2013-11-07 08:42:26 +0000 UTC,10s,1.0100,val_extra3,"",val_extra1
dbafe9c8614c785a65aabd116dd3959c3c56f7f7,default,*voice,dsafdsag,*rated,*out,cgrates.org,call,1001,1001,+4986517174964,2013-11-07 09:42:25 +0000 UTC,2013-11-07 09:42:26 +0000 UTC,20s,1.0100,val_extra3,"",val_extra1
`

	fileContent2 = `accid21;*prepaid;itsyscom.com;1001;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*postpaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
#accid1;*pseudoprepaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;12;val_extra3;"";val_extra1
accid23;*rated;cgrates.org;1001;086517174963;2013-02-03 19:54:00;26;val_extra3;"";val_extra1`

	fileContent3 = `:Tenant,ToR,OriginID,RequestType,Account,Subject,Destination,SetupTime,AnswerTime,Usage
cgrates.org,*voice,SessionFromCsv,*prepaid,1001,ANY2CNT,1002,2018-01-07 17:00:00 +0000 UTC,2018-01-07 17:00:10 +0000 UTC,5m
`

	fileContentForFilter = `accid21;*prepaid;itsyscom.com;1002;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*postpaid;itsyscom.com;1002;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
accid23;*rated;cgrates.org;1001;086517174963;2013-02-03 19:54:00;26;val_extra3;"";val_extra1`

	csvTests = []func(t *testing.T){
		testCreateDirs,
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
		testCsvITReaderWithFilter,
		testCsvITAnalyzeReaderWithFilter,
		testCleanupFiles,
		testCsvITKillEngine,
	}
)

func TestCsvReadFile(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		csvCfgDIR = "ers_internal"
	case utils.MetaMySQL:
		csvCfgDIR = "ers_mysql"
	case utils.MetaMongo:
		csvCfgDIR = "ers_mongo"
	case utils.MetaPostgres:
		csvCfgDIR = "ers_postgres"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, test := range csvTests {
		t.Run(csvCfgDIR, test)
	}
}

func testCsvITInitConfig(t *testing.T) {
	var err error
	csvCfgPath = path.Join(*dataDir, "conf", "samples", csvCfgDIR)
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

func testCsvITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCsvITRpcConn(t *testing.T) {
	var err error
	csvRPC, err = newRPCClient(csvCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testCsvITLoadTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	var loadInst utils.LoadInstance
	if err := csvRPC.Call(utils.APIerSv2LoadTariffPlanFromFolder,
		attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

// The default scenario, out of ers defined in .cfg file
func testCsvITHandleCdr1File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(fileContent1), 0644); err != nil {
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
	if err := os.WriteFile(tmpFilePath, []byte(fileContent2), 0644); err != nil {
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
	if err := csvRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	aSessions := make([]*sessions.ExternalSession, 0)
	if err := csvRPC.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{},
		&aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	fileName := "file3.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(fileContent3), 0644); err != nil {
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
	} else if len(aSessions) != 3 {
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
	if err := csvRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err == nil ||
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
	time.Sleep(100 * time.Millisecond)

	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs:    []string{"CustomerCharges"},
			OriginIDs: []string{"SessionFromCsv"},
		},
	}
	if err := csvRPC.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.099 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args.RPCCDRsFilter = &utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"},
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
	if err := csvRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testCsvITProcessFilteredCDR(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(fileContentForFilter), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/ers_with_filters/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testCsvITAnalyzeFilteredCDR(t *testing.T) {
	time.Sleep(100 * time.Millisecond)

	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			NotRunIDs: []string{"CustomerCharges", "SupplierCharges"},
			Sources:   []string{"ers_csv"},
		},
	}
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
	time.Sleep(500 * time.Millisecond)
	if outContent1, err := os.ReadFile("/tmp/ers/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContent1 != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", fileContent1, string(outContent1))
	}
	if outContent2, err := os.ReadFile("/tmp/ers2/out/file2.csv"); err != nil {
		t.Error(err)
	} else if fileContent2 != string(outContent2) {
		t.Errorf("Expecting: %q, received: %q", fileContent2, string(outContent2))
	}
	if outContent3, err := os.ReadFile("/tmp/cdrs/out/file3.csv"); err != nil {
		t.Error(err)
	} else if fileContent3 != string(outContent3) {
		t.Errorf("Expecting: %q, received: %q", fileContent3, string(outContent3))
	}
	if outContent4, err := os.ReadFile("/tmp/ers_with_filters/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContentForFilter != string(outContent4) {
		t.Errorf("Expecting: %q, received: %q", fileContentForFilter, string(outContent4))
	}
}

func testCsvITReaderWithFilter(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(fileContent1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/readerWithTemplate/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testCsvITAnalyzeReaderWithFilter(t *testing.T) {
	time.Sleep(100 * time.Millisecond)

	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			NotRunIDs: []string{"CustomerCharges", "SupplierCharges"},
			Sources:   []string{"ers_template_combined"},
		},
	}
	if err := csvRPC.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", utils.ToJSON(cdrs))
	} else if cdrs[0].Account != "1001" || cdrs[1].Account != "1001" {
		t.Errorf("Expecting: 1001, received: <%s> , <%s>", cdrs[0].Account, cdrs[1].Account)
	} else if cdrs[0].Tenant != "cgrates.org" || cdrs[1].Tenant != "cgrates.org" {
		t.Errorf("Expecting: itsyscom.com, received: <%s> , <%s>", cdrs[0].Tenant, cdrs[1].Tenant)
	} else if cdrs[0].ExtraFields["ExtraInfo1"] != "ExtraInfo1" || cdrs[1].ExtraFields["ExtraInfo1"] != "ExtraInfo1" {
		t.Errorf("Expecting: itsyscom.com, received: <%s> , <%s>", cdrs[1].ExtraFields["ExtraInfo1"], cdrs[1].ExtraFields["ExtraInfo1"])
	}
}

func testCsvITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func TestFileCSVProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	cfg.ERsCfg().Readers[0].HeaderDefineChar = ":"
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileCSVProcessEvent/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.csv"))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`:Test,,*voice,OriginCDR1,*prepaid,,cgrates.org,*call,1001,SUBJECT_TEST_1001,1002,2021-01-07 17:00:02 +0000 UTC,2021-01-07 17:00:04 +0000 UTC,1h2m
	:Test2,,*voice,OriginCDR1,*prepaid,,cgrates.org,*call,1001,SUBJECT_TEST_1001,1002,2021-01-07 17:00:02 +0000 UTC,2021-01-07 17:00:04 +0000 UTC,1h2m`))
	file.Close()
	eR := &CSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/ers/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   "2021-01-07 17:00:04 +0000 UTC",
			utils.Category:     "*call",
			utils.Destination:  "1002",
			utils.OriginID:     "OriginCDR1",
			utils.RequestType:  "*prepaid",
			utils.SetupTime:    "2021-01-07 17:00:02 +0000 UTC",
			utils.Subject:      "SUBJECT_TEST_1001",
			utils.Tenant:       "cgrates.org",
			utils.ToR:          "*voice",
			utils.Usage:        "1h2m",
		},
		APIOpts: map[string]interface{}{},
	}
	eR.conReqs <- struct{}{}

	eR.Config().Fields = []*config.FCTemplate{
		{
			Tag:       "ToR",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.ToR",
			Value:     config.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "OriginID",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.OriginID",
			Value:     config.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "RequestType",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.RequestType",
			Value:     config.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "Tenant",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.Tenant",
			Value:     config.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "Category",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.Category",
			Value:     config.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "Account",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.Account",
			Value:     config.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "Subject",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.Subject",
			Value:     config.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "Destination",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.Destination",
			Value:     config.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "SetupTime",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.SetupTime",
			Value:     config.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "AnswerTime",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.AnswerTime",
			Value:     config.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
			Mandatory: true,
		},
		{
			Tag:       "Usage",
			Type:      utils.MetaVariable,
			Path:      "*cgreq.Usage",
			Value:     config.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
			Mandatory: true,
		},
	}

	for idx := range eR.Config().Fields {
		eR.Config().Fields[idx].ComputePath()
	}

	fname := "file1.csv"
	if err := eR.processFile(filePath, fname); err != nil {
		t.Error(err)
	}
	select {
	case data := <-eR.rdrEvents:
		expEvent.ID = data.cgrEvent.ID
		expEvent.Time = data.cgrEvent.Time
		if !reflect.DeepEqual(data.cgrEvent, expEvent) {
			t.Errorf("Expected %v but received %v", utils.ToJSON(expEvent), utils.ToJSON(data.cgrEvent))
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Time limit exceeded")
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileCSVProcessEventError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileCSVProcessEvent/"
	fname := "file1.csv"
	eR := &CSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/ers/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	errExpect := "open /tmp/TestFileCSVProcessEvent/file1.csv: no such file or directory"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestFileCSVProcessEventError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileCSVProcessEvent/"
	fname := "file1.csv"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.csv"))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`#ToR,OriginID,RequestType,Tenant,Category,Account,Subject,Destination,SetupTime,AnswerTime,Usage
	,,*voice,OriginCDR1,*prepaid,,cgrates.org,*call,1001,SUBJECT_TEST_1001,1002,2021-01-07 17:00:02 +0000 UTC,2021-01-07 17:00:04 +0000 UTC,1h2m`))
	file.Close()
	eR := &CSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/ers/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}

	eR.Config().Fields = []*config.FCTemplate{
		{},
	}

	errExpect := "unsupported type: <>"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileCSVProcessEventError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].Fields = []*config.FCTemplate{}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := engine.NewFilterS(cfg, nil, dm)
	filePath := "/tmp/TestFileCSVProcessEvent/"
	fname := "file1.csv"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.csv"))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`#ToR,OriginID,RequestType,Tenant,Category,Account,Subject,Destination,SetupTime,AnswerTime,Usage
	,,*voice,OriginCDR1,*prepaid,,cgrates.org,*call,1001,SUBJECT_TEST_1001,1002,2021-01-07 17:00:02 +0000 UTC,2021-01-07 17:00:04 +0000 UTC,1h2m`))
	file.Close()
	eR := &CSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/ers/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}

	//
	eR.Config().Filters = []string{"Filter1"}
	errExpect := "NOT_FOUND:Filter1"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	//
	eR.Config().Filters = []string{"*exists:~*req..Account:"}
	errExpect = "Invalid fieldPath [ Account]"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileCSVDirErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &CSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/ers/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	eR.Config().RunDelay = -1
	errExpect := "no such file or directory"
	if err := eR.Serve(); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}
func TestFileCSV(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &CSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/ers/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	filePath := "/tmp/ers/out/"
	err := os.MkdirAll(filePath, 0777)
	if err != nil {
		t.Error(err)
	}
	for i := 1; i < 4; i++ {
		if _, err := os.Create(path.Join(filePath, fmt.Sprintf("file%d.csv", i))); err != nil {
			t.Error(err)
		}
	}
	eR.Config().RunDelay = 1 * time.Millisecond
	os.Create(path.Join(filePath, "file1.txt"))
	eR.Config().RunDelay = 1 * time.Millisecond
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
}

func TestFileCSVExit(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &CSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/ers/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	eR.Config().RunDelay = 1 * time.Millisecond
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
	eR.rdrExit <- struct{}{}
}
