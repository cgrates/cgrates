//go:build integration
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

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fwvCfgPath string
	fwvCfgDIR  string
	fwvCfg     *config.CGRConfig
	fwvRPC     *rpc.Client

	fwvTests = []func(t *testing.T){
		testCreateDirs,
		testFWVITInitConfig,
		testFWVITInitCdrDb,
		testFWVITResetDataDb,
		testFWVITStartEngine,
		testFWVITRpcConn,
		testFWVITLoadTPFromFolder,
		testFWVITHandleCdr1File,
		testFWVITAnalyseCDRs,
		testCleanupFiles,
		testFWVITKillEngine,
	}
)

func TestFWVReadFile(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		fwvCfgDIR = "ers_internal"
	case utils.MetaMySQL:
		fwvCfgDIR = "ers_mysql"
	case utils.MetaMongo:
		fwvCfgDIR = "ers_mongo"
	case utils.MetaPostgres:
		fwvCfgDIR = "ers_postgres"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, test := range fwvTests {
		t.Run(fwvCfgDIR, test)
	}
}

func testFWVITInitConfig(t *testing.T) {
	var err error
	fwvCfgPath = path.Join(*dataDir, "conf", "samples", fwvCfgDIR)
	if fwvCfg, err = config.NewCGRConfigFromPath(fwvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testFWVITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDB(fwvCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testFWVITResetDataDb(t *testing.T) {
	if err := engine.InitDataDB(fwvCfg); err != nil {
		t.Fatal(err)
	}
}

func testFWVITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fwvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testFWVITRpcConn(t *testing.T) {
	var err error
	fwvRPC, err = newRPCClient(fwvCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testFWVITLoadTPFromFolder(t *testing.T) {
	//add a default charger
	chargerProfile := &apis.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var result string
	if err := fwvRPC.Call(utils.AdminSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

var fwvContent = `HDR0001DDB     ABC                                     Some Connect A.B.                       DDB-Some-10022-20120711-309.CDR         00030920120711100255
CDR0000010  0 20120708181506000123451234         0040123123120                  004                                            000018009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009
CDR0000020  0 20120708190945000123451234         0040123123120                  004                                            000016009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009
CDR0000030  0 20120708191009000123451234         0040123123120                  004                                            000020009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009
CDR0000040  0 20120708231043000123451234         0040123123120                  004                                            000011009980010001ISDN  ABC   10Buiten uw regio                         EHV 00000009190000000009
CDR0000050  0 20120709122216000123451235         004212                         004                                            000217009980010001ISDN  ABC   10Buiten uw regio                         HMR 00000000190000000000
CDR0000060  0 20120709130542000123451236         0012323453                     004                                            000019009980010001ISDN  ABC   35Sterdiensten                            AP  00000000190000000000
CDR0000070  0 20120709140032000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000080  0 20120709140142000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000090  0 20120709150305000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000100  0 20120709150414000123451237         0040012323453100               001                                            000057009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000110  0 20120709150531000123451237         0040012323453100               001                                            000059009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000120  0 20120709150635000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000130  0 20120709151756000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000140  0 20120709154549000123451237         0040012323453100               001                                            000052009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000150  0 20120709154701000123451237         0040012323453100               001                                            000121009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000160  0 20120709154842000123451237         0040012323453100               001                                            000055009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000170  0 20120709154956000123451237         0040012323453100               001                                            000115009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000180  0 20120709155131000123451237         0040012323453100               001                                            000059009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000190  0 20120709155236000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000200  0 20120709160309000123451237         0040012323453100               001                                            000100009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000210  0 20120709160415000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000220  0 20120709161739000123451237         0040012323453100               001                                            000058009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000230  0 20120709170356000123123459         0040123234531                  004                                            000012002760010001ISDN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000240  0 20120709181036000123123450         0012323453                     004                                            000042009980010001ISDN  ABC   05Binnen uw regio                         AP  00000010190000000010
CDR0000250  0 20120709191245000123123458         0040123232350                  004                                            000012002760000001PSTN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000260  0 20120709202324000123123459         0040123234531                  004                                            000011002760010001ISDN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000270  0 20120709211756000123451237         0040012323453100               001                                            000051009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000280  0 20120709211852000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000290  0 20120709212904000123123458         0040123232350                  004                                            000012002760000001PSTN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000300  0 20120709073707000123123459         0040123234531                  004                                            000012002760010001ISDN  276   10Buiten uw regio                         TB  00000009190000000009
CDR0000310  0 20120709085451000123451237         0040012323453100               001                                            000744009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000320  0 20120709091756000123451237         0040012323453100               001                                            000050009980030001ISDN  ABD   20Internationaal                          NLB 00000000190000000000
CDR0000330  0 20120710070434000123123458         0040123232350                  004                                            000012002760000001PSTN  276   10Buiten uw regio                         TB  00000009190000000009
TRL0001DDB     ABC                                     Some Connect A.B.                       DDB-Some-10022-20120711-309.CDR         0003090000003300000030550000000001000000000100Y
`

// The default scenario, out of ers defined in .cfg file
func testFWVITHandleCdr1File(t *testing.T) {
	fileName := "file1.fwv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(fwvContent), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/fwvErs/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testFWVITAnalyseCDRs(t *testing.T) {
	time.Sleep(time.Second)
	var reply []*engine.ExternalCDR
	if err := fwvRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 34 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := fwvRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{OriginIDs: []string{"CDR0000010"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testFWVITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func TestNewFWVFileER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	cfg.ERsCfg().Readers[cfgIdx].SourcePath = "/"
	expected := &FWVFileER{
		cgrCfg: cfg,
		cfgIdx: cfgIdx,
		rdrDir: "",
	}
	cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs = 1
	result, err := NewFWVFileER(cfg, cfgIdx, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	result.(*FWVFileER).conReqs = nil
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestFWVFileConfig(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers = []*config.EventReaderCfg{
		{
			ID:             "file_reader1",
			Type:           utils.MetaFileCSV,
			RunDelay:       -1,
			ConcurrentReqs: 1024,
			SourcePath:     "/tmp/ers/in",
			ProcessedPath:  "/tmp/ers/out",
			Tenant:         nil,
			Timezone:       utils.EmptyString,
			Filters:        []string{},
			Flags:          utils.FlagsWithParams{},
			Opts:           make(map[string]interface{}),
		},
		{
			ID:             "file_reader2",
			Type:           utils.MetaFileCSV,
			RunDelay:       -1,
			ConcurrentReqs: 1024,
			SourcePath:     "/tmp/ers/in",
			ProcessedPath:  "/tmp/ers/out",
			Tenant:         nil,
			Timezone:       utils.EmptyString,
			Filters:        []string{},
			Flags:          utils.FlagsWithParams{},
			Opts:           make(map[string]interface{}),
		},
	}
	expected := cfg.ERsCfg().Readers[0]
	rdr, err := NewFWVFileER(cfg, 0, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	result := rdr.Config()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestFileFWVProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileFWVProcessEvent/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.fwv"))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte("test,test2"))
	file.Close()
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	fname := "file1.fwv"
	errExpect := "unsupported field prefix: <> when set fields"
	eR.Config().Fields = []*config.FCTemplate{
		{
			Value: config.RSRParsers{
				{
					Rules: "~*hdr",
				},
			},
			Type: utils.MetaRemove,
			// Path: utils.MetaVars,
		},
	}
	eR.Config().Fields[0].ComputePath()
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileFWVServeErrTimeDuration0(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	rdr, err := NewFWVFileER(cfg, cfgIdx, nil, nil, nil, nil, nil)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	rdr.Config().RunDelay = time.Duration(0)
	result := rdr.Serve()
	if !reflect.DeepEqual(result, nil) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, result)
	}
}

func TestFileFWVServeErrTimeDurationNeg1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfgIdx := 0
	rdr, err := NewFWVFileER(cfg, cfgIdx, nil, nil, nil, nil, nil)
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

func TestFileFWV(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	filePath := "/tmp/fwvErs/out"
	err := os.MkdirAll(filePath, 0777)
	if err != nil {
		t.Error(err)
	}
	for i := 1; i < 4; i++ {
		if _, err := os.Create(path.Join(filePath, fmt.Sprintf("file%d.fwv", i))); err != nil {
			t.Error(err)
		}
	}
	eR.Config().RunDelay = 1 * time.Millisecond
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
	os.Create(path.Join(filePath, "file1.txt"))
	eR.Config().RunDelay = 1 * time.Millisecond
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
}

func TestFileFWVServeDefault(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	filePath := "/tmp/fwvErs/out"
	err := os.MkdirAll(filePath, 0777)
	if err != nil {
		t.Error(err)
	}
	for i := 1; i < 4; i++ {
		if _, err := os.Create(path.Join(filePath, fmt.Sprintf("file%d.fwv", i))); err != nil {
			t.Error(err)
		}
	}
	os.Create(path.Join(filePath, "file1.txt"))
	eR.Config().RunDelay = 1 * time.Millisecond
	go func() {
		time.Sleep(20 * time.Millisecond)
		close(eR.rdrExit)
	}()
	eR.serveDefault()
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileFWVExit(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
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

func TestFileFWVProcessTrailer(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := engine.NewFilterS(cfg, nil, dm)
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"OriginID": "testOriginID",
		},
		APIOpts: map[string]interface{}{},
	}
	eR.conReqs <- struct{}{}
	filePath := "/tmp/TestFileFWVProcessTrailer/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.txt"))
	if err != nil {
		t.Error(err)
	}
	trailerFields := []*config.FCTemplate{
		{
			Tag:   "OriginId",
			Path:  "*cgreq.OriginID",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("testOriginID", utils.InfieldSep),
		},
	}
	eR.Config().Fields = trailerFields
	eR.Config().Fields[0].ComputePath()
	if err := eR.processTrailer(file, 0, 0, "/tmp/fwvErs/out", trailerFields); err != nil {
		t.Error(err)
	}
	select {
	case data := <-eR.rdrEvents:
		expEvent.ID = data.cgrEvent.ID
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

func TestFileFWVProcessTrailerError1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := engine.NewFilterS(cfg, nil, dm)
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	filePath := "/tmp/TestFileFWVProcessTrailer/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.txt"))
	if err != nil {
		t.Error(err)
	}
	trailerFields := []*config.FCTemplate{
		{},
	}
	errExpect := "unsupported type: <>"
	if err := eR.processTrailer(file, 0, 0, "/tmp/fwvErs/out", trailerFields); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileFWVProcessTrailerError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	eR.Config().Tenant = config.RSRParsers{
		{
			Rules: "cgrates.org",
		},
	}
	filePath := "/tmp/TestFileFWVProcessTrailer/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.txt"))
	if err != nil {
		t.Error(err)
	}

	trailerFields := []*config.FCTemplate{
		{
			Tag:   "OriginId",
			Path:  "*cgreq.OriginID",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("testOriginID", utils.InfieldSep),
		},
	}

	//
	eR.Config().Filters = []string{"Filter1"}
	errExpect := "NOT_FOUND:Filter1"
	if err := eR.processTrailer(file, 0, 0, "/tmp/fwvErs/out", trailerFields); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestFileFWVProcessTrailerError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	trailerFields := []*config.FCTemplate{
		{
			Tag:   "OriginId",
			Path:  "*cgreq.OriginID",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("testOriginID", utils.InfieldSep),
		},
	}
	var file *os.File
	errExp := "invalid argument"
	if err := eR.processTrailer(file, 0, 0, "/tmp/fwvErs/out", trailerFields); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v but received %v", errExp, err)
	}
}

func TestFileFWVCreateHeaderMap(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"OriginID": "testOriginID",
		},
		APIOpts: map[string]interface{}{},
	}
	hdrFields := []*config.FCTemplate{
		{
			Tag:   "OriginId",
			Path:  "*cgreq.OriginID",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("testOriginID", utils.InfieldSep),
		},
	}
	eR.Config().Fields = hdrFields
	eR.Config().Fields[0].ComputePath()
	record := "testRecord"
	if err := eR.createHeaderMap(record, 0, 0, "/tmp/fwvErs/out", hdrFields); err != nil {
		t.Error(err)
	}
	select {
	case data := <-eR.rdrEvents:
		expEvent.ID = data.cgrEvent.ID
		if !reflect.DeepEqual(data.cgrEvent, expEvent) {
			t.Errorf("Expected %v but received %v", utils.ToJSON(expEvent), utils.ToJSON(data.cgrEvent))
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Time limit exceeded")
	}
}

func TestFileFWVCreateHeaderMapError1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	trailerFields := []*config.FCTemplate{
		{},
	}
	record := "testRecord"
	errExpect := "unsupported type: <>"
	if err := eR.createHeaderMap(record, 0, 0, "/tmp/fwvErs/out", trailerFields); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestFileFWVCreateHeaderMapError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	eR := &FWVFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/fwvErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	record := "testRecord"
	trailerFields := []*config.FCTemplate{
		{
			Tag:   "OriginId",
			Path:  "*cgreq.OriginID",
			Type:  utils.MetaConstant,
			Value: config.NewRSRParsersMustCompile("testOriginID", utils.InfieldSep),
		},
	}

	//
	eR.Config().Filters = []string{"Filter1"}
	errExpect := "NOT_FOUND:Filter1"
	if err := eR.createHeaderMap(record, 0, 0, "/tmp/fwvErs/out", trailerFields); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}
