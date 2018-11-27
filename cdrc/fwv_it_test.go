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

//
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

var fwvCfgPath string
var fwvCfg *config.CGRConfig
var fwvRpc *rpc.Client
var fwvCdrcCfg *config.CdrcCfg

var FW_CDR_FILE1 = `HDR0001DDB     ABC                                     Some Connect A.B.                       DDB-Some-10022-20120711-309.CDR         00030920120711100255
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

func TestFwvitInitCfg(t *testing.T) {
	var err error
	fwvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcfwv")
	if fwvCfg, err = config.NewCGRConfigFromFolder(fwvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// Creates cdr files and moves them into processing folder
func TestFwvitCreateCdrFiles(t *testing.T) {
	if fwvCfg == nil {
		t.Fatal("Empty default cdrc configuration")
	}
	for _, cdrcCfg := range fwvCfg.CdrcProfiles["/tmp/cgr_fwv/cdrc/in"] {
		if cdrcCfg.ID == "FWV1" {
			fwvCdrcCfg = cdrcCfg
		}
	}
	for _, cdrcProfiles := range fwvCfg.CdrcProfiles {
		for _, cdrcInst := range cdrcProfiles {
			for _, dir := range []string{cdrcInst.CdrInDir, cdrcInst.CdrOutDir} {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal("Error removing folder: ", dir, err)
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Error creating folder: ", dir, err)
				}
			}
		}
	}
}

func TestFwvitStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fwvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestFwvitRpcConn(t *testing.T) {
	var err error
	fwvRpc, err = jsonrpc.Dial("tcp", fwvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestFwvitInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(fwvCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func TestFwvitResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(fwvCfg); err != nil {
		t.Fatal(err)
	}
}

func TestFwvitProcessFiles(t *testing.T) {
	fileName := "test1.fwv"
	if err := ioutil.WriteFile(path.Join("/tmp", fileName), []byte(FW_CDR_FILE1), 0755); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(path.Join("/tmp", fileName), path.Join(fwvCdrcCfg.CdrInDir, fileName)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Duration(1) * time.Second)
	filesInDir, _ := ioutil.ReadDir(fwvCdrcCfg.CdrInDir)
	if len(filesInDir) != 0 {
		t.Errorf("Files in cdrcInDir: %d", len(filesInDir))
	}
	filesOutDir, _ := ioutil.ReadDir(fwvCdrcCfg.CdrOutDir)
	if len(filesOutDir) != 1 {
		t.Errorf("In CdrcOutDir, expecting 1 files, got: %d", len(filesOutDir))
	}
}

func TestFwvitAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := fwvRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 8 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := fwvRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{OriginIDs: []string{"CDR0000010"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestFwvitKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Begin tests for cdrc fwv with new filters
func TestFwvit2InitCfg(t *testing.T) {
	var err error
	fwvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcfwvwithfilter")
	if fwvCfg, err = config.NewCGRConfigFromFolder(fwvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// Creates cdr files and moves them into processing folder
func TestFwvit2CreateCdrFiles(t *testing.T) {
	if fwvCfg == nil {
		t.Fatal("Empty default cdrc configuration")
	}
	for _, cdrcCfg := range fwvCfg.CdrcProfiles["/tmp/cgr_fwv/cdrc/in"] {
		if cdrcCfg.ID == "FWVWithFilter" {
			fwvCdrcCfg = cdrcCfg
		}
	}
	if err := os.RemoveAll(fwvCdrcCfg.CdrInDir); err != nil {
		t.Fatal("Error removing folder: ", fwvCdrcCfg.CdrInDir, err)
	}
	if err := os.MkdirAll(fwvCdrcCfg.CdrInDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", fwvCdrcCfg.CdrInDir, err)
	}
	if err := os.RemoveAll(fwvCdrcCfg.CdrOutDir); err != nil {
		t.Fatal("Error removing folder: ", fwvCdrcCfg.CdrOutDir, err)
	}
	if err := os.MkdirAll(fwvCdrcCfg.CdrOutDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", fwvCdrcCfg.CdrOutDir, err)
	}
}

func TestFwvit2StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fwvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestFwvit2RpcConn(t *testing.T) {
	var err error
	fwvRpc, err = jsonrpc.Dial("tcp", fwvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestFwvit2InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(fwvCfg); err != nil {
		t.Fatal(err)
	}
}

func TestFwvit2ProcessFiles(t *testing.T) {
	fileName := "test1.fwv"
	if err := ioutil.WriteFile(path.Join("/tmp", fileName), []byte(FW_CDR_FILE1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(path.Join("/tmp", fileName), path.Join(fwvCdrcCfg.CdrInDir, fileName)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Duration(1) * time.Second)
	filesInDir, _ := ioutil.ReadDir(fwvCdrcCfg.CdrInDir)
	if len(filesInDir) != 0 {
		t.Errorf("Files in cdrcInDir: %d", len(filesInDir))
	}
	filesOutDir, _ := ioutil.ReadDir(fwvCdrcCfg.CdrOutDir)
	if len(filesOutDir) != 1 {
		t.Errorf("In CdrcOutDir, expecting 1 files, got: %d", len(filesOutDir))
	}
}

func TestFwvit2AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := fwvRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestFwvit2KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Begin tests for cdrc fwv with new filters
func TestFwvit3InitCfg(t *testing.T) {
	var err error
	fwvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcfwvwithfilter")
	if fwvCfg, err = config.NewCGRConfigFromFolder(fwvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestFwvit3InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(fwvCfg); err != nil {
		t.Fatal(err)
	}
}

// Creates cdr files and moves them into processing folder
func TestFwvit3CreateCdrFiles(t *testing.T) {
	if fwvCfg == nil {
		t.Fatal("Empty default cdrc configuration")
	}
	for _, cdrcCfg := range fwvCfg.CdrcProfiles["/tmp/cgr_fwv/cdrc/in"] {
		if cdrcCfg.ID == "FWVWithFilterID" {
			fwvCdrcCfg = cdrcCfg
		}
	}
	if err := os.RemoveAll(fwvCdrcCfg.CdrInDir); err != nil {
		t.Fatal("Error removing folder: ", fwvCdrcCfg.CdrInDir, err)
	}
	if err := os.MkdirAll(fwvCdrcCfg.CdrInDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", fwvCdrcCfg.CdrInDir, err)
	}
	if err := os.RemoveAll(fwvCdrcCfg.CdrOutDir); err != nil {
		t.Fatal("Error removing folder: ", fwvCdrcCfg.CdrOutDir, err)
	}
	if err := os.MkdirAll(fwvCdrcCfg.CdrOutDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", fwvCdrcCfg.CdrOutDir, err)
	}
}

func TestFwvit3StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fwvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestFwvit3RpcConn(t *testing.T) {
	var err error
	fwvRpc, err = jsonrpc.Dial("tcp", fwvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestFwvit3AddFilters(t *testing.T) {
	filter := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_FWV",
		Rules: []*engine.FilterRule{
			{
				Type:      "*string",
				FieldName: "0-10",
				Values:    []string{"CDR0000010"},
			},
		},
	}
	var result string
	if err := fwvRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func TestFwvit3ProcessFiles(t *testing.T) {
	fileName := "test1.fwv"
	if err := ioutil.WriteFile(path.Join("/tmp", fileName), []byte(FW_CDR_FILE1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(path.Join("/tmp", fileName), path.Join(fwvCdrcCfg.CdrInDir, fileName)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Duration(1) * time.Second)
	filesInDir, _ := ioutil.ReadDir(fwvCdrcCfg.CdrInDir)
	if len(filesInDir) != 0 {
		t.Errorf("Files in cdrcInDir: %d", len(filesInDir))
	}
	filesOutDir, _ := ioutil.ReadDir(fwvCdrcCfg.CdrOutDir)
	if len(filesOutDir) != 1 {
		t.Errorf("In CdrcOutDir, expecting 1 files, got: %d", len(filesOutDir))
	}
}

func TestFwvit3AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := fwvRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestFwvit3KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
