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
	"net/rpc"
	"os"
	"path"
	"path/filepath"
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

	ackSuccessfull = `INVITE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK||*prepaid|1001|1002||3401:2069362475|flatstore"1.0
ACK|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK|1436454408|*prepaid|1001|1002||3401:2069362475|flatstore"1.0
BYE|2daec40c|548625ac|dd0c4c617a9919d29a6175cdff223a9e@0:0:0:0:0:0:0:0|200|OK|1436454410|||||3401:2069362475|flatstore"1.0
INVITE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0|200|OK||*postpaid|1002|1001||1877:893549741|flatstore"1.0
ACK|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0|200|OK|1436454647|*postpaid|1002|1001||1877:893549741|flatstore"1.0
BYE|f9d3d5c3|c863a6e3|214d8f52b566e33a9349b184e72a4cca@0:0:0:0:0:0:0:0|200|OK|1436454651|||||1877:893549741|flatstore"1.0
INVITE|36e39a5|42d996f9|3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:0|200|OK||*prepaid|1001|1002||2407:1884881533|flatstore"1.0
ACK|36e39a5|42d996f9|3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:0|200|OK|1436454657|*prepaid|1001|1002||2407:1884881533|flatstore"1.0
BYE|36e39a5|42d996f9|3a63321dd3b325eec688dc2aefb6ac2d@0:0:0:0:0:0:0:0|200|OK|1436454661|||||2407:1884881533|flatstore"1.0
INVITE|3111f3c9|49ca4c42|a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:0|200|OK||*prepaid|1001|1002||3099:1909036290|flatstore"1.0
ACK|3111f3c9|49ca4c42|a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:0|200|OK|1436454690|*prepaid|1001|1002||3099:1909036290|flatstore"1.0
BYE|3111f3c9|49ca4c42|a58ebaae40d08d6757d8424fb09c4c54@0:0:0:0:0:0:0:0|200|OK|1436454692|||||3099:1909036290|flatstore"1.0`

	ackSuccessfull2 = `"1","INVITE","CG123456789-1","123456789","CGqwerty-asd5sdf456s1f53a1sf51s648486h4et86h46","200","OK","2021-06-01 00:00:09","1001","1001@192.168.56.203","192.168.56.203","192.168.56.20","sip:1001@192.168.56.203:5060;transport=UDP","1002","1002","local","1","test"
"2","ACK","CG123456789-1","123456789","CGqwerty-asd5sdf456s1f53a1sf51s648486h4et86h46","200","OK","2021-06-01 00:00:09","1001","1001@192.168.56.203","192.168.56.203","192.168.56.20",,"1002","1002","192.168.57.203","1","test"
"3","INVITE","CG123456789-2","123456799","CGqwerty-56dfghs56hj4fg56j56sdgf516351drfg","200","OK","2021-06-01 00:00:10","1003","1003@192.168.56.204","192.168.56.204","192.168.56.21","sip:1003@192.168.56.204:5060;transport=UDP","1004","1004","local","1","test"
"4","ACK","CG123456789-2","123456799","CGqwerty-56dfghs56hj4fg56j56sdgf516351drfg","200","OK","2021-06-01 00:00:10","1003","1003@192.168.56.204","192.168.56.204","192.168.56.21",,"1004","1004","192.168.57.204","1","test"
"5","BYE","CG123456789-1","123456789","CGqwerty-asd5sdf456s1f53a1sf51s648486h4et86h46","200","OK","2021-06-01 00:00:37","1001","1001@192.168.56.203","192.168.56.203","192.168.56.20",,"1002","1002","192.168.57.203","1","test"
"6","CANCEL","CG123456789-3",,"CGqwerty-dfgdfhdfgjfgdh534dfg563h56dr56deghed6r5","200","OK","2021-06-01 00:00:42","1005","1005@192.168.56.201","192.168.56.201","192.168.56.2",,"1006","1006","192.168.56.301","0","test"
"7","BYE","CG123456789-2","123456799","CGqwerty-56dfghs56hj4fg56j56sdgf516351drfg","200","OK","2021-06-01 00:00:50","1003","1003@192.168.56.204","192.168.56.204","192.168.56.21",,"1004","1004","192.168.57.204","1","test"`

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
		testFlatstoreITHandleCdr2File,
		testFlatstoreITAnalyseCDRs2,
		testFlatstoreITHandleCdr3File,
		testFlatstoreITAnalyseCDRs3,
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
	ids := []string{}
	for _, fD := range filesOutDir {
		ids = append(ids, fD.Name())
	}
	if len(filesOutDir) != 5 {
		t.Errorf("Unexpected number of files in output directory: %+v, %q", len(filesOutDir), ids)
	}
	ePartContent := "flatStore|dd0c4c617a9919d29a6175cdff223a9p@0:0:0:0:0:0:0:02daec40c548625ac|*prepaid|1001|1001|1002|1436454408|1436454408|200 OK|3401:2069362475\n"
	tmpl := path.Join("/tmp/flatstoreErs/out", "f7aed15c98b31fea0e3b02b52fc947879a3c5bbc.*.tmp")
	if match, err := filepath.Glob(tmpl); err != nil {
		t.Error(err)
	} else if len(match) != 1 {
		t.Errorf("Wrong number of files matches the template: %q", match)
		t.Errorf("template: %q", tmpl)
		t.Errorf("files: %q", ids)
	} else if partContent, err := os.ReadFile(match[0]); err != nil {
		t.Error(err)
	} else if ePartContent != string(partContent) {
		t.Errorf("Expecting:\n%q\nReceived:\n%q", ePartContent, string(partContent))
	}
}

func testFlatstoreITAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := flatstoreRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 8 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
		t.Error(utils.ToJSON(reply))
	}
	if err := flatstoreRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{MinUsage: "1"}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 5 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

// The default scenario, out of ers defined in .cfg file
func testFlatstoreITHandleCdr2File(t *testing.T) {
	if err := os.WriteFile(path.Join("/tmp", "acc_1.log"), []byte(ackSuccessfull), 0644); err != nil {
		t.Fatal(err.Error())
	}
	//Rename(oldpath, newpath string)
	if err := os.Rename(path.Join("/tmp", "acc_1.log"), path.Join("/tmp/flatstoreACKErs/in", "acc_1.log")); err != nil {
		t.Fatal(err)
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
	filesOutDir, _ := os.ReadDir("/tmp/flatstoreACKErs/out")
	ids := []string{}
	for _, fD := range filesOutDir {
		ids = append(ids, fD.Name())
	}
	if len(filesOutDir) != 1 {
		t.Errorf("Unexpected number of files in output directory: %+v, %q", len(filesOutDir), ids)
	}
}

func testFlatstoreITAnalyseCDRs2(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := flatstoreRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{OriginHosts: []string{"flatStoreACK"}, MinUsage: "1"}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 4 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testFlatstoreITHandleCdr3File(t *testing.T) {
	if err := os.WriteFile(path.Join("/tmp", "acc_1.log"), []byte(ackSuccessfull2), 0644); err != nil {
		t.Fatal(err.Error())
	}
	//Rename(oldpath, newpath string)
	if err := os.Rename(path.Join("/tmp", "acc_1.log"), path.Join("/tmp/flatstoreMMErs/in", "acc_1.log")); err != nil {
		t.Fatal(err)
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
	filesOutDir, _ := os.ReadDir("/tmp/flatstoreMMErs/out")
	ids := []string{}
	for _, fD := range filesOutDir {
		ids = append(ids, fD.Name())
	}
	if len(filesOutDir) != 1 {
		t.Errorf("Unexpected number of files in output directory: %+v, %q", len(filesOutDir), ids)
	}
}

func testFlatstoreITAnalyseCDRs3(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := flatstoreRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{OriginHosts: []string{"flatstoreMMErs"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 3 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := flatstoreRPC.Call(utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{OriginHosts: []string{"flatstoreMMErs"}, MinUsage: "1"}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testFlatstoreITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
