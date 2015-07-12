/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var flatstoreCfgPath string
var flatstoreCfg *config.CGRConfig
var flatstoreRpc *rpc.Client
var flatstoreCdrcCfg *config.CdrcConfig

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
INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475`

func TestFlatstoreLclInitCfg(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	flatstoreCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcflatstore")
	if flatstoreCfg, err = config.NewCGRConfigFromFolder(flatstoreCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestFlatstoreLclInitCdrDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitStorDb(flatstoreCfg); err != nil {
		t.Fatal(err)
	}
}

// Creates cdr files and moves them into processing folder
func TestFlatstoreLclCreateCdrFiles(t *testing.T) {
	if !*testLocal {
		return
	}
	if flatstoreCfg == nil {
		t.Fatal("Empty default cdrc configuration")
	}
	flatstoreCdrcCfg = flatstoreCfg.CdrcProfiles["/tmp/cgr_flatstore/cdrc/in"]["FLATSTORE"]
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

func TestFlatstoreLclStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := engine.StopStartEngine(flatstoreCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestFlatstoreLclRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	flatstoreRpc, err = jsonrpc.Dial("tcp", flatstoreCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestFlatstoreLclProcessFiles(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := ioutil.WriteFile(path.Join("/tmp", "acc_1.log"), []byte(fullSuccessfull), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := ioutil.WriteFile(path.Join("/tmp", "missed_calls_1.log"), []byte(fullMissed), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := ioutil.WriteFile(path.Join("/tmp", "acc_2.log"), []byte(part1), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := ioutil.WriteFile(path.Join("/tmp", "acc_3.log"), []byte(part2), 0644); err != nil {
		t.Fatal(err.Error)
	}
	//Rename(oldpath, newpath string)
	for _, fileName := range []string{"acc_1.log", "missed_calls_1.log", "acc_2.log", "acc_3.log"} {
		if err := os.Rename(path.Join("/tmp", fileName), path.Join(flatstoreCdrcCfg.CdrInDir, fileName)); err != nil {
			t.Fatal(err)
		}
	}
}

/*

// Creates cdr files and starts the engine
func TestCreateCdr3File(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := os.RemoveAll(cdrcCfg.CdrInDir); err != nil {
		t.Fatal("Error removing folder: ", cdrcCfg.CdrInDir, err)
	}
	if err := os.MkdirAll(cdrcCfg.CdrInDir, 0755); err != nil {
		t.Fatal("Error creating folder: ", cdrcCfg.CdrInDir, err)
	}
	if err := ioutil.WriteFile(path.Join(cdrcCfg.CdrInDir, "file3.csv"), []byte(fileContent3), 0644); err != nil {
		t.Fatal(err.Error)
	}
}

func TestProcessCdr3Dir(t *testing.T) {
	if !*testLocal {
		return
	}
	if cdrcCfg.Cdrs == utils.INTERNAL { // For now we only test over network
		cdrcCfg.Cdrs = "127.0.0.1:2013"
	}
	if err := startEngine(); err != nil {
		t.Fatal(err.Error())
	}
	cdrc, err := NewCdrc(cdrcCfgs, true, nil, make(chan struct{}))
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := cdrc.processCdrDir(); err != nil {
		t.Error(err)
	}
	stopEngine()
}
*/
