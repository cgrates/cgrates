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

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	xmlCfgPath string
	xmlCfgDIR  string
	xmlCfg     *config.CGRConfig
	xmlRPC     *rpc.Client

	xmlTests = []func(t *testing.T){
		testCreateDirs,
		testXMLITInitConfig,
		testXMLITInitCdrDb,
		testXMLITResetDataDb,
		testXMLITStartEngine,
		testXMLITRpcConn,
		testXMLITLoadTPFromFolder,
		testXMLITHandleCdr1File,
		testXmlITAnalyseCDRs,
		testCleanupFiles,
		testXMLITKillEngine,
	}
)

func TestXMLReadFile(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		xmlCfgDIR = "ers_internal"
	case utils.MetaMySQL:
		xmlCfgDIR = "ers_mysql"
	case utils.MetaMongo:
		xmlCfgDIR = "ers_mongo"
	case utils.MetaPostgres:
		xmlCfgDIR = "ers_postgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, test := range xmlTests {
		t.Run(xmlCfgDIR, test)
	}
}

func testXMLITInitConfig(t *testing.T) {
	var err error
	xmlCfgPath = path.Join(*dataDir, "conf", "samples", xmlCfgDIR)
	if xmlCfg, err = config.NewCGRConfigFromPath(xmlCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testXMLITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testXMLITResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

func testXMLITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(xmlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testXMLITRpcConn(t *testing.T) {
	var err error
	xmlRPC, err = newRPCClient(xmlCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testXMLITLoadTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	var loadInst utils.LoadInstance
	if err := xmlRPC.Call(utils.APIerSv2LoadTariffPlanFromFolder,
		attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

var cdrXmlBroadsoft = `<?xml version="1.0" encoding="ISO-8859-1"?>
<!DOCTYPE broadWorksCDR>
<broadWorksCDR version="19.0">
  <cdrData>
    <headerModule>
      <recordId>
        <eventCounter>0002183384</eventCounter>
        <systemId>CGRateSaabb</systemId>
        <date>20160419210000.104</date>
        <systemTimeZone>1+020000</systemTimeZone>
      </recordId>
      <type>Start</type>
    </headerModule>
  </cdrData>
  <cdrData>
    <headerModule>
      <recordId>
        <eventCounter>0002183385</eventCounter>
        <systemId>CGRateSaabb</systemId>
        <date>20160419210005.247</date>
        <systemTimeZone>1+020000</systemTimeZone>
      </recordId>
      <serviceProvider>MBC</serviceProvider>
      <type>Normal</type>
    </headerModule>
    <basicModule>
      <userNumber>1001</userNumber>
      <groupNumber>2001</groupNumber>
      <asCallType>Network</asCallType>
      <callingNumber>1001</callingNumber>
      <callingPresentationIndicator>Public</callingPresentationIndicator>
      <calledNumber>+4915117174963</calledNumber>
      <startTime>20160419210005.247</startTime>
      <userTimeZone>1+020000</userTimeZone>
      <localCallId>25160047719:0</localCallId>
      <answerIndicator>Yes</answerIndicator>
      <answerTime>20160419210006.813</answerTime>
      <releaseTime>20160419210020.296</releaseTime>
      <terminationCause>016</terminationCause>
      <chargeIndicator>y</chargeIndicator>
      <releasingParty>local</releasingParty>
      <userId>1001@cgrates.org</userId>
      <clidPermitted>Yes</clidPermitted>
      <namePermitted>Yes</namePermitted>
    </basicModule>
    <centrexModule>
      <group>CGR_GROUP</group>
      <trunkGroupName>CGR_GROUP/CGR_GROUP_TRUNK30</trunkGroupName>
      <trunkGroupInfo>Normal</trunkGroupInfo>
      <locationList>
        <locationInformation>
          <location>1001@cgrates.org</location>
          <locationType>Primary Device</locationType>
        </locationInformation>
      </locationList>
      <locationUsage>31.882</locationUsage>
    </centrexModule>
    <ipModule>
      <route>gw04.cgrates.org</route>
      <networkCallID>74122796919420162305@172.16.1.2</networkCallID>
      <codec>PCMA/8000</codec>
      <accessDeviceAddress>172.16.1.4</accessDeviceAddress>
      <accessCallID>BW2300052501904161738474465@172.16.1.10</accessCallID>
      <codecUsage>31.882</codecUsage>
      <userAgent>OmniPCX Enterprise R11.0.1 k1.520.22.b</userAgent>
    </ipModule>
  </cdrData>
  <cdrData>
    <headerModule>
      <recordId>
        <eventCounter>0002183386</eventCounter>
        <systemId>CGRateSaabb</systemId>
        <date>20160419210006.909</date>
        <systemTimeZone>1+020000</systemTimeZone>
      </recordId>
      <serviceProvider>MBC</serviceProvider>
      <type>Normal</type>
    </headerModule>
    <basicModule>
      <userNumber>1002</userNumber>
      <groupNumber>2001</groupNumber>
      <asCallType>Network</asCallType>
      <callingNumber>+4986517174964</callingNumber>
      <callingPresentationIndicator>Public</callingPresentationIndicator>
      <calledNumber>1001</calledNumber>
      <startTime>20160419210006.909</startTime>
      <userTimeZone>1+020000</userTimeZone>
      <localCallId>27280048121:0</localCallId>
      <answerIndicator>Yes</answerIndicator>
      <answerTime>20160419210007.037</answerTime>
      <releaseTime>20160419210030.322</releaseTime>
      <terminationCause>016</terminationCause>
      <chargeIndicator>y</chargeIndicator>
      <releasingParty>local</releasingParty>
      <userId>314028947650@cgrates.org</userId>
      <clidPermitted>Yes</clidPermitted>
      <namePermitted>Yes</namePermitted>
    </basicModule>
    <centrexModule>
      <group>CGR_GROUP</group>
      <trunkGroupName>CGR_GROUP/CGR_GROUP_TRUNK65</trunkGroupName>
      <trunkGroupInfo>Normal</trunkGroupInfo>
      <locationList>
        <locationInformation>
          <location>31403456100@cgrates.org</location>
          <locationType>Primary Device</locationType>
        </locationInformation>
      </locationList>
      <locationUsage>26.244</locationUsage>
    </centrexModule>
    <ipModule>
      <route>gw01.cgrates.org</route>
      <networkCallID>108352493719420162306@172.31.250.150</networkCallID>
      <codec>PCMA/8000</codec>
      <accessDeviceAddress>172.16.1.4</accessDeviceAddress>
      <accessCallID>2345300069121904161716512907@172.16.1.10</accessCallID>
      <codecUsage>26.244</codecUsage>
      <userAgent>Altitude vBox</userAgent>
    </ipModule>
  </cdrData>
  <cdrData>
    <headerModule>
      <recordId>
        <eventCounter>0002183486</eventCounter>
        <systemId>CGRateSaabb</systemId>
        <date>20160419211500.104</date>
        <systemTimeZone>1+020000</systemTimeZone>
      </recordId>
      <type>End</type>
    </headerModule>
  </cdrData>
</broadWorksCDR>`

// The default scenario, out of ers defined in .cfg file
func testXMLITHandleCdr1File(t *testing.T) {
	fileName := "file1.xml"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := os.WriteFile(tmpFilePath, []byte(cdrXmlBroadsoft), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/xmlErs/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testXmlITAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := xmlRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 6 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := xmlRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{DestinationPrefixes: []string{"+4915117174963"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 3 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testXMLITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func TestNewXMLFileER(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].SourcePath = "/tmp/xmlErs/out/"
	cfg.ERsCfg().Readers[0].ConcurrentReqs = 1
	fltrs := &engine.FilterS{}
	expEr := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/xmlErs/out",
		rdrEvents: nil,
		rdrError:  nil,
		rdrExit:   nil,
		conReqs:   make(chan struct{}, 1),
	}
	var value struct{}
	expEr.conReqs <- value
	eR, err := NewXMLFileER(cfg, 0, nil, nil, nil, fltrs, nil)
	expConReq := make(chan struct{}, 1)
	expConReq <- struct{}{}
	if <-expConReq != <-eR.(*XMLFileER).conReqs {
		t.Errorf("Expected %v but received %v", <-expConReq, <-eR.(*XMLFileER).conReqs)
	}
	expEr.conReqs = nil
	eR.(*XMLFileER).conReqs = nil
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eR.(*XMLFileER), expEr) {
		t.Errorf("Expected %v but received %v", expEr.conReqs, eR.(*XMLFileER).conReqs)
	}
}

func TestFileXMLProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileXMLProcessEvent/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.xml"))
	if err != nil {
		t.Error(err)
	}
	xmlData := `<?xml version="1.0" encoding="ISO-8859-1"?>
  <!DOCTYPE broadWorksCDR>
  <broadWorksCDR version="19.0">
    <cdrData>
      <basicModule>
        <localCallId>
          <localCallId>25160047719:0</localCallId>
        </localCallId>
      </basicModule>
    </cdrData>
  </broadWorksCDR>
  `
	file.Write([]byte(xmlData))
	file.Close()
	eR := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/xmlErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}

	//or set the default Fields of cfg.ERsCfg().Readers[0].Fields
	eR.Config().Fields = []*config.FCTemplate{
		{
			Tag:       "OriginID",
			Type:      utils.MetaConstant,
			Path:      "*cgreq.OriginID",
			Value:     config.NewRSRParsersMustCompile("25160047719:0", utils.InfieldSep),
			Mandatory: true,
		},
	}

	eR.Config().Fields[0].ComputePath()

	eR.conReqs <- struct{}{}
	fileName := "file1.xml"
	if err := eR.processFile(filePath, fileName); err != nil {
		t.Error(err)
	}
	expEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			"OriginID": "25160047719:0",
		},
		APIOpts: make(map[string]interface{}),
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

func TestFileXMLProcessEventError1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileXMLProcessEvent/"
	fname := "file1.xml"
	eR := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/xmlErs/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	errExpect := "open /tmp/TestFileXMLProcessEvent/file1.xml: no such file or directory"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
}

func TestFileXMLProcessEVentError2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].Fields = []*config.FCTemplate{}
	data := engine.NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	filePath := "/tmp/TestFileXMLProcessEvent/"
	fname := "file1.xml"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.xml"))
	if err != nil {
		t.Error(err)
	}
	xmlData := `<?xml version="1.0" encoding="ISO-8859-1"?>
  <!DOCTYPE broadWorksCDR>
  <broadWorksCDR version="19.0">
    <cdrData>
      <basicModule>
        <localCallId>
          <localCallId>25160047719:0</localCallId>
        </localCallId>
      </basicModule>
    </cdrData>
  </broadWorksCDR>
  `
	file.Write([]byte(xmlData))
	file.Close()
	eR := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/xmlErs/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	eR.Config().Tenant = config.RSRParsers{
		{
			Rules: "test",
		},
	}

	//
	eR.Config().Filters = []string{"Filter1"}
	errExpect := "NOT_FOUND:Filter1"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}

	//
	eR.Config().Filters = []string{"*exists:~*req..Account:"}
	errExpect = "rename /tmp/TestFileXMLProcessEvent/file1.xml /var/spool/cgrates/ers/out/file1.xml: no such file or directory"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileXMLProcessEVentError3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	// fltrs := &engine.FilterS{}
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileXMLProcessEvent/"
	fname := "file1.xml"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.xml"))
	if err != nil {
		t.Error(err)
	}
	xmlData := `<?xml version="1.0" encoding="ISO-8859-1"?>
  <!DOCTYPE broadWorksCDR>
  <broadWorksCDR version="19.0">
    <cdrData>
      <basicModule>
        <localCallId>
          <localCallId>25160047719:0</localCallId>
        </localCallId>
      </basicModule>
    </cdrData>
  </broadWorksCDR>
  `
	file.Write([]byte(xmlData))
	file.Close()
	eR := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/xmlErs/out/",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}

	eR.Config().Fields = []*config.FCTemplate{
		{
			Tag:       "OriginID",
			Type:      utils.MetaConstant,
			Path:      "*cgreq.OriginID",
			Value:     nil,
			Mandatory: true,
		},
	}

	eR.Config().Fields[0].ComputePath()
	errExpect := "Empty source value for fieldID: <OriginID>"
	if err := eR.processFile(filePath, fname); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileXMLProcessEventParseError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.ERsCfg().Readers[0].ProcessedPath = ""
	fltrs := &engine.FilterS{}
	filePath := "/tmp/TestFileXMLProcessEvent/"
	if err := os.MkdirAll(filePath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(filePath, "file1.xml"))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
  <XMLField>test/XMLField>`))
	file.Close()
	eR := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/xmlErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}

	fileName := "file1.xml"
	errExpect := "XML syntax error on line 2: unexpected EOF"
	if err := eR.processFile(filePath, fileName); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v but received %v", errExpect, err)
	}
	if err := os.RemoveAll(filePath); err != nil {
		t.Error(err)
	}
}

func TestFileXML(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/xmlErs/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	err := os.MkdirAll(eR.rdrDir, 0777)
	if err != nil {
		t.Error(err)
	}

	eR.Config().Fields = []*config.FCTemplate{
		{
			Tag:       "OriginID",
			Type:      utils.MetaConstant,
			Path:      "*cgreq.OriginID",
			Value:     config.NewRSRParsersMustCompile("25160047719:0", utils.InfieldSep),
			Mandatory: true,
		},
	}

	eR.Config().Fields[0].ComputePath()

	for i := 1; i < 4; i++ {
		if _, err := os.Create(path.Join(eR.rdrDir, fmt.Sprintf("file%d.xml", i))); err != nil {
			t.Error(err)
		}
	}

	eR.Config().RunDelay = time.Duration(-1)
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}

	eR.Config().RunDelay = 1 * time.Millisecond
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
	os.Create(path.Join(eR.rdrDir, "file1.txt"))
	eR.Config().RunDelay = 1 * time.Millisecond
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
}

func TestFileXMLError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/xmlErsError/out",
		rdrEvents: make(chan *erEvent, 1),
		rdrError:  make(chan error, 1),
		rdrExit:   make(chan struct{}),
		conReqs:   make(chan struct{}, 1),
	}
	eR.conReqs <- struct{}{}
	err := os.MkdirAll(eR.rdrDir, 0777)
	if err != nil {
		t.Error(err)
	}

	eR.Config().Fields = []*config.FCTemplate{
		{
			Tag:       "OriginID",
			Type:      utils.MetaConstant,
			Path:      "*cgreq.OriginID",
			Value:     config.NewRSRParsersMustCompile("25160047719:0", utils.InfieldSep),
			Mandatory: true,
		},
	}

	eR.Config().Fields[0].ComputePath()

	for i := 1; i < 4; i++ {
		if _, err := os.Create(path.Join(eR.rdrDir, fmt.Sprintf("file%d.xml", i))); err != nil {
			t.Error(err)
		}
	}
	os.Create(path.Join(eR.rdrDir, "file1.txt"))
	eR.Config().RunDelay = 1 * time.Millisecond
	if err := eR.Serve(); err != nil {
		t.Error(err)
	}
}

func TestFileXMLExit(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltrs := &engine.FilterS{}
	eR := &XMLFileER{
		cgrCfg:    cfg,
		cfgIdx:    0,
		fltrS:     fltrs,
		rdrDir:    "/tmp/xmlErs/out",
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
