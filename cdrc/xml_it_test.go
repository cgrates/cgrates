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
package cdrc

import (
	"io/ioutil"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var xmlCfgPath string
var xmlCfg *config.CGRConfig
var cdrcXmlCfgs []*config.CdrcCfg
var cdrcXmlCfg *config.CdrcCfg
var cdrcXmlRPC *rpc.Client
var xmlPathIn1, xmlPathOut1 string

func TestXmlITInitConfig(t *testing.T) {
	var err error
	xmlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcxml")
	if xmlCfg, err = config.NewCGRConfigFromFolder(xmlCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestXmlITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func TestXmlITResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestXmlITCreateCdrDirs(t *testing.T) {
	for _, cdrcProfiles := range xmlCfg.CdrcProfiles {
		for _, cdrcInst := range cdrcProfiles {
			for _, dir := range []string{cdrcInst.CdrInDir, cdrcInst.CdrOutDir} {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal("Error removing folder: ", dir, err)
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Error creating folder: ", dir, err)
				}
			}
			if cdrcInst.ID == "XMLit1" { // Initialize the folders to check later
				xmlPathIn1 = cdrcInst.CdrInDir
				xmlPathOut1 = cdrcInst.CdrOutDir
			}
		}
	}
}

func TestXmlITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(xmlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestXmlITRpcConn(t *testing.T) {
	var err error
	cdrcXmlRPC, err = jsonrpc.Dial("tcp", xmlCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestXmlITHandleCdr1File(t *testing.T) {
	fileName := "file1.xml"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(cdrXmlBroadsoft), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join(xmlPathIn1, fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestXmlITProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join(xmlPathOut1, "file1.xml")); err != nil {
		t.Error(err)
	} else if cdrXmlBroadsoft != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", cdrXmlBroadsoft, string(outContent1))
	}
}

func TestXmlITAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcXmlRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := cdrcXmlRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{DestinationPrefixes: []string{"+4986517174963"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}

}

func TestXmlITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Begin tests for cdrc xml with new filters
func TestXmlIT2InitConfig(t *testing.T) {
	var err error
	xmlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcxmlwithfilter")
	if xmlCfg, err = config.NewCGRConfigFromFolder(xmlCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestXmlIT2InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestXmlIT2CreateCdrDirs(t *testing.T) {
	for _, cdrcProfiles := range xmlCfg.CdrcProfiles {
		for _, cdrcInst := range cdrcProfiles {
			for _, dir := range []string{cdrcInst.CdrInDir, cdrcInst.CdrOutDir} {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal("Error removing folder: ", dir, err)
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Error creating folder: ", dir, err)
				}
			}
			if cdrcInst.ID == "XMLWithFilter" { // Initialize the folders to check later
				xmlPathIn1 = cdrcInst.CdrInDir
				xmlPathOut1 = cdrcInst.CdrOutDir
			}
		}
	}
}

func TestXmlIT2StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(xmlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestXmlIT2RpcConn(t *testing.T) {
	var err error
	cdrcXmlRPC, err = jsonrpc.Dial("tcp", xmlCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestXmlIT2HandleCdr1File(t *testing.T) {
	fileName := "file1.xml"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(cdrXmlBroadsoft), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join(xmlPathIn1, fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestXmlIT2ProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join(xmlPathOut1, "file1.xml")); err != nil {
		t.Error(err)
	} else if cdrXmlBroadsoft != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", cdrXmlBroadsoft, string(outContent1))
	}
}

func TestXmlIT2AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcXmlRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestXmlIT2KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Begin tests for cdrc xml with new filters
func TestXmlIT3InitConfig(t *testing.T) {
	var err error
	xmlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcxmlwithfilter")
	if xmlCfg, err = config.NewCGRConfigFromFolder(xmlCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestXmlIT3InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestXmlIT3CreateCdrDirs(t *testing.T) {
	for _, cdrcProfiles := range xmlCfg.CdrcProfiles {
		for _, cdrcInst := range cdrcProfiles {
			for _, dir := range []string{cdrcInst.CdrInDir, cdrcInst.CdrOutDir} {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal("Error removing folder: ", dir, err)
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Error creating folder: ", dir, err)
				}
			}
			if cdrcInst.ID == "msw_xml" { // Initialize the folders to check later
				xmlPathIn1 = cdrcInst.CdrInDir
				xmlPathOut1 = cdrcInst.CdrOutDir
			}
		}
	}
}

func TestXmlIT3StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(xmlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestXmlIT3RpcConn(t *testing.T) {
	var err error
	cdrcXmlRPC, err = jsonrpc.Dial("tcp", xmlCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestXmlIT3HandleCdr1File(t *testing.T) {
	fileName := "file1.xml"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(xmlContent), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join(xmlPathIn1, fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestXmlIT3ProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join(xmlPathOut1, "file1.xml")); err != nil {
		t.Error(err)
	} else if xmlContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", xmlContent, string(outContent1))
	}
}

func TestXmlIT3AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcXmlRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestXmlIT3KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Begin tests for cdrc xml with new filters
func TestXmlIT4InitConfig(t *testing.T) {
	var err error
	xmlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcxmlwithfilter")
	if xmlCfg, err = config.NewCGRConfigFromFolder(xmlCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestXmlIT4InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestXmlIT4CreateCdrDirs(t *testing.T) {
	for _, cdrcProfiles := range xmlCfg.CdrcProfiles {
		for _, cdrcInst := range cdrcProfiles {
			for _, dir := range []string{cdrcInst.CdrInDir, cdrcInst.CdrOutDir} {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal("Error removing folder: ", dir, err)
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Error creating folder: ", dir, err)
				}
			}
			if cdrcInst.ID == "msw_xml2" { // Initialize the folders to check later
				xmlPathIn1 = cdrcInst.CdrInDir
				xmlPathOut1 = cdrcInst.CdrOutDir
			}
		}
	}
}

func TestXmlIT4StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(xmlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestXmlIT4RpcConn(t *testing.T) {
	var err error
	cdrcXmlRPC, err = jsonrpc.Dial("tcp", xmlCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestXmlIT4HandleCdr1File(t *testing.T) {
	fileName := "file1.xml"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(xmlContent), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join(xmlPathIn1, fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestXmlIT4ProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join(xmlPathOut1, "file1.xml")); err != nil {
		t.Error(err)
	} else if xmlContent != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", xmlContent, string(outContent1))
	}
}

func TestXmlIT4AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcXmlRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestXmlIT4KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Begin tests for cdrc xml with new filters
func TestXmlIT5InitConfig(t *testing.T) {
	var err error
	xmlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcxmlwithfilter")
	if xmlCfg, err = config.NewCGRConfigFromFolder(xmlCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestXmlIT5InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestXmlIT5CreateCdrDirs(t *testing.T) {
	for _, cdrcProfiles := range xmlCfg.CdrcProfiles {
		for _, cdrcInst := range cdrcProfiles {
			for _, dir := range []string{cdrcInst.CdrInDir, cdrcInst.CdrOutDir} {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal("Error removing folder: ", dir, err)
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Error creating folder: ", dir, err)
				}
			}
			if cdrcInst.ID == "XMLWithFilterID" { // Initialize the folders to check later
				xmlPathIn1 = cdrcInst.CdrInDir
				xmlPathOut1 = cdrcInst.CdrOutDir
			}
		}
	}
}

func TestXmlIT5StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(xmlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestXmlIT5RpcConn(t *testing.T) {
	var err error
	cdrcXmlRPC, err = jsonrpc.Dial("tcp", xmlCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestXmlIT5AddFilters(t *testing.T) {
	filter := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_XML",
		Rules: []*engine.FilterRule{
			{
				Type:      "*string",
				FieldName: "broadWorksCDR.cdrData.basicModule.userNumber",
				Values:    []string{"1002"},
			},
			{
				Type:      "*string",
				FieldName: "broadWorksCDR.cdrData.headerModule.type",
				Values:    []string{"Normal"},
			},
		},
	}
	var result string
	if err := cdrcXmlRPC.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestXmlIT5HandleCdr1File(t *testing.T) {
	fileName := "file1.xml"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(cdrXmlBroadsoft), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join(xmlPathIn1, fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestXmlIT5ProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join(xmlPathOut1, "file1.xml")); err != nil {
		t.Error(err)
	} else if cdrXmlBroadsoft != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", cdrXmlBroadsoft, string(outContent1))
	}
}

func TestXmlIT5AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcXmlRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestXmlIT5KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Begin tests for cdrc xml with index
func TestXmlIT6InitConfig(t *testing.T) {
	var err error
	xmlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcxmlwithfilter")
	if xmlCfg, err = config.NewCGRConfigFromFolder(xmlCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestXmlIT6InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestXmlIT6CreateCdrDirs(t *testing.T) {
	for _, cdrcProfiles := range xmlCfg.CdrcProfiles {
		for _, cdrcInst := range cdrcProfiles {
			for _, dir := range []string{cdrcInst.CdrInDir, cdrcInst.CdrOutDir} {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal("Error removing folder: ", dir, err)
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Error creating folder: ", dir, err)
				}
			}
			if cdrcInst.ID == "XMLWithIndex" { // Initialize the folders to check later
				xmlPathIn1 = cdrcInst.CdrInDir
				xmlPathOut1 = cdrcInst.CdrOutDir
			}
		}
	}
}

func TestXmlIT6StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(xmlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestXmlIT6RpcConn(t *testing.T) {
	var err error
	cdrcXmlRPC, err = jsonrpc.Dial("tcp", xmlCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestXmlIT6HandleCdr1File(t *testing.T) {
	fileName := "file1.xml"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(xmlMultipleIndex), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join(xmlPathIn1, fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestXmlIT6ProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join(xmlPathOut1, "file1.xml")); err != nil {
		t.Error(err)
	} else if xmlMultipleIndex != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", xmlMultipleIndex, string(outContent1))
	}
}

func TestXmlIT6AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcXmlRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestXmlIT6KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
