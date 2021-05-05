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
	"bytes"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/ltcache"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	flatstoreCfgPath string
	flatstoreCfgDIR  string
	flatstoreCfg     *config.CGRConfig
	flatstoreRPC     *rpc.Client

	fullSuccessfull = `INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475
BYE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK|1436454410|||||3401:2069362475
INVITE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0|200|OK|1436454647|*postpaid|1002|1001||1877:893549741
BYE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0|200|OK|1436454651|||||1877:893549741
INVITE|36e39a5|42d996f9|3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:0|200|OK|1436454657|*prepaid|1001|1002||2407:1884881533
BYE|36e39a5|42d996f9|3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:0|200|OK|1436454661|||||2407:1884881533
INVITE|3111f3c9|49ca4c42|a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:0|200|OK|1436454690|*prepaid|1001|1002||3099:1909036290
BYE|3111f3c9|49ca4c42|a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:0|200|OK|1436454692|||||3099:1909036290`

	fullMissed = `INVITE|ef6c6256|da501581|0bfdd176d1b93e7df3de5c6f4873ee04@0:0:0:0:0:0:0:0|487|Request Terminated|1436454643|*prepaid|1001|1002||1224:339382783
INVITE|7905e511||81880da80a94bda81b425b09009e055c@0:0:0:0:0:0:0:0|404|Not Found|1436454668|*prepaid|1001|1002||1980:1216490844
INVITE|324cb497|d4af7023|8deaadf2ae9a17809a391f05af31afb0@0:0:0:0:0:0:0:0|486|Busy here|1436454687|*postpaid|1002|1001||474:130115066`

	part1 = `BYE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4ccb@0:0:0:0:0:0:0:0|200|OK|1436454651|||||1877:893549742`

	part2 = `INVITE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4ccb@0:0:0:0:0:0:0:0|200|OK|1436454647|*postpaid|1002|1003||1877:893549742
INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9p@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475`

	flatstoreTests = []func(t *testing.T){
		testCreateDirs,
		testFlatstoreITInitConfig,
		testFlatstoreITInitCdrDb,
		testFlatstoreITResetDataDb,
		testFlatstoreITStartEngine,
		testFlatstoreITRpcConn,
		testFlatstoreITLoadTPFromFolder,
		testFlatstoreITHandleCdr1File,
		testFlatstoreITAnalyseCDRs,
		testCleanupFiles,
		testFlatstoreITKillEngine,
	}
)

func TestFlatstoreFile(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		flatstoreCfgDIR = "ers_internal"
	case utils.MetaMySQL:
		flatstoreCfgDIR = "ers_mysql"
	case utils.MetaMongo:
		flatstoreCfgDIR = "ers_mongo"
	case utils.MetaPostgres:
		flatstoreCfgDIR = "ers_postgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, test := range flatstoreTests {
		t.Run(flatstoreCfgDIR, test)
	}
}

func testFlatstoreITInitConfig(t *testing.T) {
	var err error
	flatstoreCfgPath = path.Join(*dataDir, "conf", "samples", flatstoreCfgDIR)
	if flatstoreCfg, err = config.NewCGRConfigFromPath(flatstoreCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testFlatstoreITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(flatstoreCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testFlatstoreITResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(flatstoreCfg); err != nil {
		t.Fatal(err)
	}
}

func testFlatstoreITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(flatstoreCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testFlatstoreITRpcConn(t *testing.T) {
	var err error
	flatstoreRPC, err = newRPCClient(flatstoreCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testFlatstoreITLoadTPFromFolder(t *testing.T) {
	//add a default charger
	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var result string
	if err := flatstoreRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

// The default scenario, out of ers defined in .cfg file
func testFlatstoreITHandleCdr1File(t *testing.T) {
	if err := os.WriteFile(path.Join("/tmp", "acc_1.log"), []byte(fullSuccessfull), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.WriteFile(path.Join("/tmp", "missed_calls_1.log"), []byte(fullMissed), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.WriteFile(path.Join("/tmp", "acc_2.log"), []byte(part1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.WriteFile(path.Join("/tmp", "acc_3.log"), []byte(part2), 0644); err != nil {
		t.Fatal(err.Error())
	}
	//Rename(oldpath, newpath string)
	for _, fileName := range []string{"acc_1.log", "missed_calls_1.log", "acc_2.log", "acc_3.log"} {
		if err := os.Rename(path.Join("/tmp", fileName), path.Join("/tmp/flatstoreErs/in", fileName)); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(time.Second)
	// check the files to be processed
	filesInDir, _ := os.ReadDir("/tmp/flatstoreErs/in")
	if len(filesInDir) != 0 {
		fls := make([]string, len(filesInDir))
		for i, fs := range filesInDir {
			fls[i] = fs.Name()
		}
		t.Errorf("Files in ersInDir: %+v", fls)
	}
	filesOutDir, _ := os.ReadDir("/tmp/flatstoreErs/out")
	if len(filesOutDir) != 5 {
		ids := []string{}
		for _, fD := range filesOutDir {
			ids = append(ids, fD.Name())
		}
		t.Errorf("Unexpected number of files in output directory: %+v, %q", len(filesOutDir), ids)
	}
	ePartContent := "INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9p@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475\n"
	if partContent, err := os.ReadFile(path.Join("/tmp/flatstoreErs/out", "acc_3.log.tmp")); err != nil {
		t.Error(err)
	} else if (ePartContent) != (string(partContent)) {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", ePartContent, string(partContent))
	}
}

func testFlatstoreITAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := flatstoreRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 8 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := flatstoreRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{MinUsage: "1"}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 5 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testFlatstoreITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func TestFlatstoreProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	cfg.ERsCfg().Readers[0].Opts[utils.FstFailedCallsPrefixOpt] = "file"
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFlatstoreProcessEvent/"
	fname := "file1.csv"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, fname))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte("INVITE,a,ToR,b,c,d,e,f,g,h,i,j,k,l"))
	file.Close()
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.cache = ltcache.NewCache(-1, 0, false, eR.dumpToFile)
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "g",
			utils.AnswerTime:   "k",
			utils.Category:     "f",
			utils.Destination:  "i",
			utils.OriginID:     "b",
			utils.RequestType:  "c",
			utils.SetupTime:    "j",
			utils.Subject:      "h",
			utils.Tenant:       "e",
			utils.ToR:          "ToR",
			utils.Usage:        "l",
		},
		APIOpts: map[string]interface{}{},
	}
	eR.conReqs <- struct{}{}
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

//Test the case in which the file name does not match a prefix
func TestFlatstoreProcessEvent2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	cfg.ERsCfg().Readers[0].Fields = append(cfg.ERsCfg().Readers[0].Fields, &config.FCTemplate{
		Tag:   "Usage",
		Path:  "*cgreq.Usage",
		Type:  utils.MetaUsageDifference,
		Value: config.NewRSRParsersMustCompile("~*bye.6;~*invite.6", utils.InfieldSep),
	})
	for _, v := range cfg.ERsCfg().Readers[0].Fields {
		v.ComputePath()
	}
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFlatstoreProcessEvent/"
	fname := "file1.csv"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, fname))
	if err != nil {
		t.Error(err)
	}
	//baToR
	file.Write([]byte("INVITE,a,ToR,b,c,d,2013-12-30T15:00:01Z,f,g,h,i,j,k,l"))
	file.Close()
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	record := []string{utils.FstBye, "a", "ToR", "b", "c", "d", "2013-12-30T16:00:01Z", "f", "g", "h", "i", "j", "k", "l"}
	pr := &fstRecord{method: utils.FstBye, values: record, fileName: fname}
	eR.cache = ltcache.NewCache(ltcache.UnlimitedCaching, 0, false, nil)
	eR.cache.Set(utils.ConcatenatedKey("baToR", utils.FstBye), pr, []string{"baToR"})
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "g",
			utils.AnswerTime:   "k",
			utils.Category:     "f",
			utils.Destination:  "i",
			utils.OriginID:     "b",
			utils.RequestType:  "c",
			utils.SetupTime:    "j",
			utils.Subject:      "h",
			utils.Tenant:       "2013-12-30T15:00:01Z",
			utils.ToR:          "ToR",
			utils.Usage:        "1h0m0s",
		},
		APIOpts: map[string]interface{}{},
	}
	eR.conReqs <- struct{}{}
	eR.Config().Opts[utils.FstFailedCallsPrefixOpt] = "x"
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

func TestFlatstoreProcessEvent2CacheNotSet(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFlatstoreProcessEvent/"
	fname := "file1.csv"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, fname))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte("INVITE,a,ToR,b,c,d,2013-12-30T15:00:01Z,f,g,h,i,j,k,l"))
	file.Close()
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}

	eR.cache = ltcache.NewCache(ltcache.UnlimitedCaching, 0, false, eR.dumpToFile)

	eR.conReqs <- struct{}{}
	eR.Config().Opts[utils.FstFailedCallsPrefixOpt] = "x"
	if err := eR.processFile(filePath, fname); err != nil {
		t.Error(err)
	}

	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

//Test pairToRecord() error, where both methods of unpaired record object are INVITE
func TestFlatstoreProcessEvent2Error2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFlatstoreProcessEvent/"
	fname := "file1.csv"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, fname))
	if err != nil {
		t.Error(err)
	}
	//Create new logger
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	//baToR
	file.Write([]byte("INVITE,a,ToR,b,c,d,2013-12-30T15:00:01Z,f,g,h,i,j,k,l"))
	file.Close()
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	record := []string{"INVITE", "a", "ToR", "b", "c", "d", "2013-12-30T16:00:01Z", "f", "g", "h", "i", "j", "k", "l"}
	pr := &fstRecord{method: utils.FstInvite, values: record, fileName: fname}
	eR.cache = ltcache.NewCache(ltcache.UnlimitedCaching, 0, false, eR.dumpToFile)
	eR.cache.Set("baToR:INVITE", pr, []string{"baToR"})
	eR.conReqs <- struct{}{}
	eR.Config().Opts[utils.FstFailedCallsPrefixOpt] = "x"
	if err := eR.processFile(filePath, fname); err != nil {
		t.Error(err)
	}
	errExpect := "[WARNING] <ERs> Overwriting the INVITE method for record <baToR>"
	if rcv := buf.String(); !strings.Contains(rcv, errExpect) {
		t.Errorf("\nExpected %v but \nreceived %v", errExpect, rcv)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
	buf.Reset()
}

//Fields in template are empty in order to trigger SetFields() error
func TestFlatstoreProcessEventError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFlatstoreProcessEvent/"
	fname := "file1.csv"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, fname))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`#ToR,OriginID,RequestType,Tenant,Category,Account,Subject,Destination,SetupTime,AnswerTime,Usage
,,*voice,OriginCDR1,*prepaid,,cgrates.org,*call,1001,SUBJECT_TEST_1001,1002,2021-01-07 17:00:02 +0000 UTC,2021-01-07 17:00:04 +0000 UTC,1h2m`))
	file.Close()
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}

	eR.Config().Fields = []*config.FCTemplate{
		{},
	}

	errExpect := `unsupported method: <"">`
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

//Test invalid filters in order to trigger Pass() function error
func TestFlatstoreProcessEventError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].Fields = []*config.FCTemplate{}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := engine.NewFilterS(cfg, nil, dm)
	filePath := "/tmp/TestFlatstoreProcessEvent/"
	fname := "file1.csv"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, fname))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`#ToR,OriginID,RequestType,Tenant,Category,Account,Subject,Destination,SetupTime,AnswerTime,Usage
BYE,,*voice,OriginCDR1,*prepaid,,cgrates.org,*call,1001,SUBJECT_TEST_1001,1002,2021-01-07 17:00:02 +0000 UTC,2021-01-07 17:00:04 +0000 UTC,1h2m`))
	file.Close()
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
		cache:     ltcache.NewCache(-1, 0, false, nil),
	}
	eR.conReqs <- struct{}{}
	eR.cache.Set("OriginCDR1*voice:INVITE", &fstRecord{method: utils.FstInvite, values: []string{}}, []string{"OriginCDR1*voice"})
	//
	eR.Config().Filters = []string{"Filter1"}
	errExpect := "NOT_FOUND:Filter1"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	//
	eR.cache.Set("OriginCDR1*voice:INVITE", &fstRecord{method: utils.FstInvite, values: []string{}}, []string{"OriginCDR1*voice"})

	eR.Config().Filters = []string{"*exists:~*req..Account:"}
	errExpect = "Invalid fieldPath [ Account]"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFlatstoreProcessEventDirError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFlatstoreProcessEvent/"
	fname := "file1.csv"
	eR := &FlatstoreER{
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
	errExpect := "open /tmp/TestFlatstoreProcessEvent/file1.csv: no such file or directory"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestFlatstore(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
}

func TestFlatstoreServeDefault(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	filePath := "/tmp/flatstoreErs/out"
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
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(eR.rdrExit)
	}()
	eR.serveDefault()
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileFlatstoreExit(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &FlatstoreER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/flatstoreErs/out",
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

func TestFlatstoreServeErrTimeDurationNeg1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	rdr, err := NewFlatstoreER(cfg, cfgIdx, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	rdr.Config().RunDelay = time.Duration(-1)
	expected := "no such file or directory"
	err = rdr.Serve()
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, err)
	}
}
