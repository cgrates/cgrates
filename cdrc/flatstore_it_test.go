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

var flatstoreCfgPath string
var flatstoreCfg *config.CGRConfig
var flatstoreRpc *rpc.Client
var flatstoreCdrcCfg *config.CdrcCfg

var fullSuccessfull = `INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475
BYE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK|1436454410|||||3401:2069362475
INVITE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0|200|OK|1436454647|*postpaid|1002|1001||1877:893549741
BYE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0|200|OK|1436454651|||||1877:893549741
INVITE|36e39a5|42d996f9|3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:0|200|OK|1436454657|*prepaid|1001|1002||2407:1884881533
BYE|36e39a5|42d996f9|3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:0|200|OK|1436454661|||||2407:1884881533
INVITE|3111f3c9|49ca4c42|a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:0|200|OK|1436454690|*prepaid|1001|1002||3099:1909036290
BYE|3111f3c9|49ca4c42|a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:0|200|OK|1436454692|||||3099:1909036290
`

var fullMissed = `INVITE|ef6c6256|da501581|0bfdd176d1b93e7df3de5c6f4873ee04@0:0:0:0:0:0:0:0|487|Request Terminated|1436454643|*prepaid|1001|1002||1224:339382783
INVITE|7905e511||81880da80a94bda81b425b09009e055c@0:0:0:0:0:0:0:0|404|Not Found|1436454668|*prepaid|1001|1002||1980:1216490844
INVITE|324cb497|d4af7023|8deaadf2ae9a17809a391f05af31afb0@0:0:0:0:0:0:0:0|486|Busy here|1436454687|*postpaid|1002|1001||474:130115066`

var part1 = `BYE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4ccb@0:0:0:0:0:0:0:0|200|OK|1436454651|||||1877:893549742
`

var part2 = `INVITE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4ccb@0:0:0:0:0:0:0:0|200|OK|1436454647|*postpaid|1002|1003||1877:893549742
INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9p@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475`

func TestFlatstoreitInitCfg(t *testing.T) {
	var err error
	flatstoreCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcflatstore")
	if flatstoreCfg, err = config.NewCGRConfigFromFolder(flatstoreCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestFlatstoreitInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(flatstoreCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func TestFlatstoreitResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(flatstoreCfg); err != nil {
		t.Fatal(err)
	}
}

// Creates cdr files and moves them into processing folder
func TestFlatstoreitCreateCdrFiles(t *testing.T) {
	if flatstoreCfg == nil {
		t.Fatal("Empty default cdrc configuration")
	}
	for _, cdrcCfg := range flatstoreCfg.CdrcProfiles["/tmp/cgr_flatstore/cdrc/in"] {
		if cdrcCfg.ID == "FLATSTORE" {
			flatstoreCdrcCfg = cdrcCfg
		}
	}
	if err := os.RemoveAll(flatstoreCdrcCfg.CdrInDir); err != nil {
		t.Fatal("Error removing folder: ", flatstoreCdrcCfg.CdrInDir, err)
	}
	if err := os.MkdirAll(flatstoreCdrcCfg.CdrInDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", flatstoreCdrcCfg.CdrInDir, err)
	}
	if err := os.RemoveAll(flatstoreCdrcCfg.CdrOutDir); err != nil {
		t.Fatal("Error removing folder: ", flatstoreCdrcCfg.CdrOutDir, err)
	}
	if err := os.MkdirAll(flatstoreCdrcCfg.CdrOutDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", flatstoreCdrcCfg.CdrOutDir, err)
	}
}

func TestFlatstoreitStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(flatstoreCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestFlatstoreitRpcConn(t *testing.T) {
	var err error
	flatstoreRpc, err = jsonrpc.Dial("tcp", flatstoreCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestFlatstoreitProcessFiles(t *testing.T) {
	if err := ioutil.WriteFile(path.Join("/tmp", "acc_1.log"), []byte(fullSuccessfull), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := ioutil.WriteFile(path.Join("/tmp", "missed_calls_1.log"), []byte(fullMissed), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := ioutil.WriteFile(path.Join("/tmp", "acc_2.log"), []byte(part1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := ioutil.WriteFile(path.Join("/tmp", "acc_3.log"), []byte(part2), 0644); err != nil {
		t.Fatal(err.Error())
	}
	//Rename(oldpath, newpath string)
	for _, fileName := range []string{"acc_1.log", "missed_calls_1.log", "acc_2.log", "acc_3.log"} {
		if err := os.Rename(path.Join("/tmp", fileName), path.Join(flatstoreCdrcCfg.CdrInDir, fileName)); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(time.Duration(3) * time.Second) // Give time for processing to happen and the .unparired file to be written
	filesInDir, _ := ioutil.ReadDir(flatstoreCdrcCfg.CdrInDir)
	if len(filesInDir) != 0 {
		t.Errorf("Files in cdrcInDir: %+v", filesInDir)
	}
	filesOutDir, _ := ioutil.ReadDir(flatstoreCdrcCfg.CdrOutDir)
	if len(filesOutDir) != 5 {
		f := []string{}
		for _, s := range filesOutDir {
			f = append(f, s.Name())
			t.Errorf("File %s:", s.Name())
			if partContent, err := ioutil.ReadFile(path.Join(flatstoreCdrcCfg.CdrOutDir, s.Name())); err != nil {
				t.Error(err)
			} else {
				t.Errorf("%s", partContent)
			}
			t.Errorf("==============================================================================")
		}
		t.Errorf("In CdrcOutDir, expecting 5 files, got: %d, for %s", len(filesOutDir), utils.ToJSON(f))
		return
	}
	ePartContent := "INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9p@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475\n"
	if partContent, err := ioutil.ReadFile(path.Join(flatstoreCdrcCfg.CdrOutDir, "acc_3.log.unpaired")); err != nil {
		t.Error(err)
	} else if (ePartContent) != (string(partContent)) {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", ePartContent, string(partContent))
	}
}

func TestFlatstoreitAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := flatstoreRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 13 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := flatstoreRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{MinUsage: "1"}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 7 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestFlatstoreitKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
