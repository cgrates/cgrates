//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ers

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	jsonCfgPath string
	jsonCfgDIR  string
	jsonCfg     *config.CGRConfig
	jsonRPC     *birpc.Client

	fileContent = `
{
	"Tenant": "cgrates.org",
	"Account": "voiceAccount",
	"AnswerTime": "2018-08-24T16:00:26Z",
	"SetupTime": "2018-08-24T16:00:26Z",
	"Destination": "+4986517174963",
	"OriginHost": "192.168.1.1",
	"OriginID": "testJsonCDR",
	"RequestType": "*pseudoprepaid",
	"Source": "jsonFile",
	"Usage": 120000000000
}`
	jsonTests = []func(t *testing.T){
		testCreateDirs,
		testJSONInitConfig,
		testJSONInitCdrDb,
		testJSONResetDataDb,
		testJSONStartEngine,
		testJSONRpcConn,
		testJSONAddData,
		testJSONHandleFile,
		testJSONVerify,
		testCleanupFiles,
		testJSONKillEngine,
	}
)

func TestJSONReadFile(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		jsonCfgDIR = "ers_internal"
	case utils.MetaMySQL:
		jsonCfgDIR = "ers_mysql"
	case utils.MetaMongo:
		jsonCfgDIR = "ers_mongo"
	case utils.MetaPostgres:
		jsonCfgDIR = "ers_postgres"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, test := range jsonTests {
		t.Run(jsonCfgDIR, test)
	}
}

func testJSONInitConfig(t *testing.T) {
	var err error
	jsonCfgPath = path.Join(*utils.DataDir, "conf", "samples", jsonCfgDIR)
	if jsonCfg, err = config.NewCGRConfigFromPath(jsonCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testJSONInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(jsonCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testJSONResetDataDb(t *testing.T) {
	if err := engine.InitDataDB(jsonCfg); err != nil {
		t.Fatal(err)
	}
}

func testJSONStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(jsonCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testJSONRpcConn(t *testing.T) {
	var err error
	jsonRPC, err = newRPCClient(jsonCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testJSONAddData(t *testing.T) {
	//add a charger
	tpDirPath := t.TempDir()
	filePath := path.Join(tpDirPath, utils.ChargersCsv)
	err := os.WriteFile(filePath,
		[]byte("cgrates.org,Default,,2014-07-14T14:35:00Z,*default,*none,20"),
		0644)
	if err != nil {
		t.Errorf("could not write to file %s: %v",
			filePath, err)
	}
	var reply string
	args := &utils.AttrLoadTpFromFolder{FolderPath: tpDirPath}
	err = jsonRPC.Call(context.Background(),
		utils.APIerSv1LoadTariffPlanFromFolder,
		args, &reply)
	if err != nil {
		t.Errorf("%s call failed for path %s: %v",
			utils.APIerSv1LoadTariffPlanFromFolder, tpDirPath, err)
	}

	attrSetAcnt := struct {
		Tenant  string
		Account string
	}{
		Tenant:  "cgrates.org",
		Account: "voiceAccount",
	}
	if err := jsonRPC.Call(context.Background(), utils.APIerSv2SetAccount, &attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "voiceAccount",
		BalanceType: utils.MetaVoice,
		Value:       600000000000,
		Balance: map[string]any{
			utils.ID:        utils.MetaDefault,
			"RatingSubject": "*zero1m",
			utils.Weight:    10.0,
		},
	}
	if err := jsonRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}

	var acnt *engine.Account
	if err := jsonRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "voiceAccount"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MetaVoice][0].Value != 600000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MetaVoice][0])
	}
}

// The default scenario, out of ers defined in .cfg file
func testJSONHandleFile(t *testing.T) {
	fileName := "file1.json"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(fileContent), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/ErsJSON/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testJSONVerify(t *testing.T) {
	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			OriginIDs: []string{"testJsonCDR"},
		},
	}
	if err := jsonRPC.Call(context.Background(), utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != 2*time.Minute {
			t.Errorf("Unexpected usage for CDR: %d", cdrs[0].Usage)
		} else if cdrs[0].ExtraFields["ReaderID"] != "JSONReader" {
			t.Errorf("Expected readerID <%v>, received <%v>", "JSONReader", cdrs[0].ExtraFields["ReaderID"])
		}
	}

	var acnt *engine.Account
	if err := jsonRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "voiceAccount"}, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 1 || acnt.BalanceMap[utils.MetaVoice][0].Value != 480000000000 {
		t.Errorf("Unexpected balance received: %+v", acnt.BalanceMap[utils.MetaVoice][0])
	}
}

func testJSONKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}

func TestFileJSONServeErrTimeDuration0(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	rdr, err := NewJSONFileER(cfg, cfgIdx, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	rdr.Config().RunDelay = time.Duration(0)
	result := rdr.Serve()
	if !reflect.DeepEqual(result, nil) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
	}
}

func TestFileJSONServeErrTimeDurationNeg1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrS := &engine.FilterS{}
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			Type:          utils.MetaFileJSON,
			RunDelay:      -1,
			ID:            "json_reader",
			SourcePath:    "/var/spool/cgrates/ers/in",
			ProcessedPath: "/var/spool/cgrates/out",
		},
	}
	srv := NewERService(cfg, nil, fltrS, nil)
	stopChan := make(chan struct{}, 1)
	cfgRldChan := make(chan struct{}, 1)
	err := srv.ListenAndServe(stopChan, cfgRldChan)

	expected := "no such file or directory"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}

// func TestFileJSONServeTimeDefault(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	cfgIdx := 0
// 	rdr, err := NewJSONFileER(cfg, cfgIdx, nil, nil, nil, nil,nil)
// 	if err != nil {
// 		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
// 	}
// 	rdr.Config().RunDelay = time.Duration(1)
// 	result := rdr.Serve()
// 	if !reflect.DeepEqual(result, nil) {
// 		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
// 	}
// }

// func TestFileJSONServeTimeDefaultChanExit(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	cfgIdx := 0
// 	rdrExit := make(chan struct{}, 1)
// 	rdr, err := NewJSONFileER(cfg, cfgIdx, nil, nil, nil, rdrExit)
// 	if err != nil {
// 		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
// 	}
// 	rdrExit <- struct{}{}
// 	rdr.Config().RunDelay = time.Duration(1)
// 	result := rdr.Serve()
// 	if !reflect.DeepEqual(result, nil) {
// 		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
// 	}
// }

// func TestFileJSONProcessFile(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	cfgIdx := 0
// 	rdr, err := NewJSONFileER(cfg, cfgIdx, nil, nil, nil, nil,nil)
// 	if err != nil {
// 		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
// 	}
// 	expected := "open : no such file or directory"
// 	err2 := rdr.(*JSONFileER).processFile("", "")
// 	if err2 == nil || err2.Error() != expected {
// 		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err2)
// 	}
// }

func TestFileJSONProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileJSONProcessEvent/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.json"))
	if err != nil {
		t.Error(err)
	}
	fcTemp := map[string]any{
		"2":  "tor_test",
		"3":  "originid_test",
		"4":  "requestType_test",
		"6":  "tenant_test",
		"7":  "category_test",
		"8":  "account_test",
		"9":  "subject_test",
		"10": "destination_test",
		"11": "setupTime_test",
		"12": "answerTime_test",
		"13": "usage_test",
	}
	rcv, err := json.Marshal(fcTemp)
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(rcv))
	file.Close()
	eR := &JSONFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		sourceDir: filePath,
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "account_test",
			utils.AnswerTime:   "answerTime_test",
			utils.Category:     "category_test",
			utils.Destination:  "destination_test",
			utils.OriginID:     "originid_test",
			utils.RequestType:  "requestType_test",
			utils.SetupTime:    "setupTime_test",
			utils.Subject:      "subject_test",
			utils.Tenant:       "tenant_test",
			utils.ToR:          "tor_test",
			utils.Usage:        "usage_test",
		},
		APIOpts: map[string]any{},
	}
	// expEvent := &utils.CGREvent{}
	fname := "file1.json"
	if err := eR.processFile(fname); err != nil {
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

func TestFileJSONProcessEventReadError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileJSONProcessEvent/"
	fname := "file2.json"
	eR := &CSVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		sourceDir: filePath,
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	errExpect := "open /tmp/TestFileJSONProcessEvent/file2.json: no such file or directory"
	if err := eR.processFile(fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestFileJSONProcessEventError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileJSONProcessEvent/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.json"))
	if err != nil {
		t.Error(err)
	}
	fcTemp := map[string]any{
		"2":  "tor_test",
		"3":  "originid_test",
		"4":  "requestType_test",
		"6":  "tenant_test",
		"7":  "category_test",
		"8":  "account_test",
		"9":  "subject_test",
		"10": "destination_test",
		"11": "setupTime_test",
		"12": "answerTime_test",
		"13": "usage_test",
	}
	rcv, err := json.Marshal(fcTemp)
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(rcv))
	file.Close()
	eR := &JSONFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		sourceDir: filePath,
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}

	eR.Config().Fields = []*config.FCTemplate{
		{},
	}
	fname := "file1.json"
	errExpect := "unsupported type: <>"
	if err := eR.processFile(fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileJSONProcessEventError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].Fields = []*config.FCTemplate{}
	data, err := engine.NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Error(err)
	}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := engine.NewFilterS(cfg, nil, dm)
	filePath := "/tmp/TestFileJSONProcessEvent/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.json"))
	if err != nil {
		t.Error(err)
	}
	fcTemp := map[string]any{
		"2":  "tor_test",
		"3":  "originid_test",
		"4":  "requestType_test",
		"6":  "tenant_test",
		"7":  "category_test",
		"8":  "account_test",
		"9":  "subject_test",
		"10": "destination_test",
		"11": "setupTime_test",
		"12": "answerTime_test",
		"13": "usage_test",
	}
	rcv, err := json.Marshal(fcTemp)
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(rcv))
	file.Close()
	eR := &JSONFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		sourceDir: filePath,
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	fname := "file1.json"

	//
	eR.Config().Filters = []string{"Filter1"}
	errExpect := "NOT_FOUND:Filter1"
	if err := eR.processFile(fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	//
	eR.Config().Filters = []string{"*exists:~*req..Account:"}
	if err := eR.processFile(fname); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileJSON(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &JSONFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		sourceDir: "/tmp/ErsJSON/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
	}
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
}

func TestFileJSONServeDefault(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &JSONFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		sourceDir: "/tmp/ErsJSON/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
	}
	filePath := "/tmp/ErsJSON/out/"
	err := os.MkdirAll(filePath, 0777)
	if err != nil {
		t.Error(err)
	}
	for i := 1; i < 4; i++ {
		if _, err := os.Create(path.Join(filePath, fmt.Sprintf("file%d.json", i))); err != nil {
			t.Error(err)
		}
	}
	eR.Config().RunDelay = 1 * time.Millisecond
	os.Create(path.Join(filePath, "file1.txt"))
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(eR.rdrExit)
	}()
	eR.serveDefault()
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileJSONExit(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &JSONFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		sourceDir: "/tmp/ErsJSON/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
	}
	eR.Config().RunDelay = 1 * time.Millisecond
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
	eR.rdrExit <- struct{}{}
}
