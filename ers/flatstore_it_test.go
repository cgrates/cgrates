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
	"io/ioutil"
	"net/rpc"
	"os"
	"path"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"

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
		testFlatstoreITCreateCdrDirs,
		testFlatstoreITInitConfig,
		testFlatstoreITInitCdrDb,
		testFlatstoreITResetDataDb,
		testFlatstoreITStartEngine,
		testFlatstoreITRpcConn,
		testFlatstoreITLoadTPFromFolder,
		testFlatstoreITHandleCdr1File,
		testFlatstoreITAnalyseCDRs,
		testFlatstoreITCleanupFiles,
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

func testFlatstoreITCreateCdrDirs(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out",
		"/tmp/ers2/in", "/tmp/ers2/out", "/tmp/init_session/in", "/tmp/init_session/out",
		"/tmp/terminate_session/in", "/tmp/terminate_session/out", "/tmp/cdrs/in",
		"/tmp/cdrs/out", "/tmp/ers_with_filters/in", "/tmp/ers_with_filters/out",
		"/tmp/xmlErs/in", "/tmp/xmlErs/out", "/tmp/fwvErs/in", "/tmp/fwvErs/out",
		"/tmp/partErs1/in", "/tmp/partErs1/out", "/tmp/partErs2/in", "/tmp/partErs2/out",
		"/tmp/flatstoreErs/in", "/tmp/flatstoreErs/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
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
	chargerProfile := &v1.ChargerWithCache{
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
		if err := os.Rename(path.Join("/tmp", fileName), path.Join("/tmp/flatstoreErs/in", fileName)); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(3 * time.Second)
	// check the files to be processed
	filesInDir, _ := ioutil.ReadDir("/tmp/flatstoreErs/in")
	if len(filesInDir) != 0 {
		t.Errorf("Files in ersInDir: %+v", filesInDir)
	}
	filesOutDir, _ := ioutil.ReadDir("/tmp/flatstoreErs/out")
	if len(filesOutDir) != 6 {
		t.Errorf("Unexpected number of files in output directory: %+v", len(filesOutDir))
	}
	ePartContent := "INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9p@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475\n"
	if partContent, err := ioutil.ReadFile(path.Join("/tmp/flatstoreErs/out", "acc_3.log.tmp")); err != nil {
		t.Error(err)
	} else if (ePartContent) != (string(partContent)) {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", ePartContent, string(partContent))
	}
}

func testFlatstoreITAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := flatstoreRPC.Call(utils.APIerSv2GetCDRs, utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 8 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := flatstoreRPC.Call(utils.APIerSv2GetCDRs, utils.RPCCDRsFilter{MinUsage: "1"}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 5 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testFlatstoreITCleanupFiles(t *testing.T) {
	for _, dir := range []string{"/tmp/ers",
		"/tmp/ers2", "/tmp/init_session", "/tmp/terminate_session",
		"/tmp/cdrs", "/tmp/ers_with_filters", "/tmp/xmlErs", "/tmp/fwvErs",
		"/tmp/partErs1", "/tmp/partErs2", "tmp/flatstoreErs"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func testFlatstoreITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
