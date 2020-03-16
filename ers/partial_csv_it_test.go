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
	"io/ioutil"
	"net/rpc"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	partCfgPath string
	partCfgDIR  string
	partCfg     *config.CGRConfig
	partRPC     *rpc.Client

	partTests = []func(t *testing.T){
		testPartITCreateCdrDirs,
		testPartITInitConfig,
		testPartITInitCdrDb,
		testPartITResetDataDb,
		testPartITStartEngine,
		testPartITRpcConn,
		testPartITLoadTPFromFolder,
		testPartITHandleCdr1File,
		testPartITHandleCdr2File,
		testPartITHandleCdr3File,
		testPartITVerifyFiles,
		testPartITAnalyseCDRs,
		testPartITCleanupFiles,
		testPartITKillEngine,
	}

	partCsvFileContent1 = `4986517174963,004986517174964,DE-National,04.07.2016 18:58:55,04.07.2016 18:58:55,1,65,Peak,0.014560,498651,partial
4986517174964,004986517174963,DE-National,04.07.2016 20:58:55,04.07.2016 20:58:55,0,74,Offpeak,0.003360,498651,complete
`

	partCsvFileContent2 = `4986517174963,004986517174964,DE-National,04.07.2016 19:00:00,04.07.2016 18:58:55,0,15,Offpeak,0.003360,498651,partial`
	partCsvFileContent3 = `4986517174964,004986517174960,DE-National,04.07.2016 19:05:55,04.07.2016 19:05:55,0,23,Offpeak,0.003360,498651,partial`

	eCacheDumpFile1 = `4986517174963_004986517174964_04.07.2016 18:58:55,1467658800,*rated,086517174963,+4986517174964,2016-07-04T18:58:55Z,2016-07-04T18:58:55Z,15s,-1.00000
4986517174963_004986517174964_04.07.2016 18:58:55,1467658735,*rated,086517174963,+4986517174964,2016-07-04T18:58:55Z,2016-07-04T18:58:55Z,1m5s,-1.00000
`
)

func TestPartReadFile(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		partCfgDIR = "ers_internal"
	case utils.MetaMySQL:
		partCfgDIR = "ers_mysql"
	case utils.MetaMongo:
		partCfgDIR = "ers_mongo"
	case utils.MetaPostgres:
		partCfgDIR = "ers_postgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, test := range partTests {
		t.Run(partCfgDIR, test)
	}
}

func testPartITCreateCdrDirs(t *testing.T) {
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

func testPartITInitConfig(t *testing.T) {
	var err error
	partCfgPath = path.Join(*dataDir, "conf", "samples", partCfgDIR)
	if partCfg, err = config.NewCGRConfigFromPath(partCfgPath); err != nil {
		fmt.Println(err)
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testPartITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(partCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testPartITResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(partCfg); err != nil {
		t.Fatal(err)
	}
}

func testPartITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(partCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testPartITRpcConn(t *testing.T) {
	var err error
	partRPC, err = newRPCClient(partCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testPartITLoadTPFromFolder(t *testing.T) {
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
	if err := partRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

// The default scenario, out of cdrc defined in .cfg file
func testPartITHandleCdr1File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(partCsvFileContent1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/partErs1/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// The default scenario, out of cdrc defined in .cfg file
func testPartITHandleCdr2File(t *testing.T) {
	fileName := "file2.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(partCsvFileContent2), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/partErs1/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// The default scenario, out of cdrc defined in .cfg file
func testPartITHandleCdr3File(t *testing.T) {
	fileName := "file3.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(partCsvFileContent3), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/partErs2/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
	time.Sleep(3 * time.Second)
}

func testPartITVerifyFiles(t *testing.T) {
	filesInDir, _ := ioutil.ReadDir("/tmp/partErs1/out/")
	if len(filesInDir) == 0 {
		t.Fatalf("No files found in folder: <%s>", "/tmp/partErs1/out")
	}
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		if strings.HasPrefix(file.Name(), "4986517174963_004986517174964") {
			fileName = file.Name()
			break
		}
	}
	if contentCacheDump, err := ioutil.ReadFile(path.Join("/tmp/partErs1/out", fileName)); err != nil {
		t.Error(err)
	} else if len(eCacheDumpFile1) != len(string(contentCacheDump)) {
		t.Errorf("Expecting: %q, \n received: %q", eCacheDumpFile1, string(contentCacheDump))
	}
}

func testPartITAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := partRPC.Call(utils.APIerSv2GetCDRs, utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := partRPC.Call(utils.APIerSv2GetCDRs, utils.RPCCDRsFilter{DestinationPrefixes: []string{"+4986517174963"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := partRPC.Call(utils.APIerSv2GetCDRs, utils.RPCCDRsFilter{DestinationPrefixes: []string{"+4986517174960"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testPartITCleanupFiles(t *testing.T) {
	for _, dir := range []string{"/tmp/ers",
		"/tmp/ers2", "/tmp/init_session", "/tmp/terminate_session",
		"/tmp/cdrs", "/tmp/ers_with_filters", "/tmp/xmlErs", "/tmp/fwvErs",
		"/tmp/partErs1", "/tmp/partErs2", "tmp/flatstoreErs"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func testPartITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
