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
	"flag"
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

var (
	csvCfgPath string
	csvCfg     *config.CGRConfig
	csvRPC     *rpc.Client
	dataDir    = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	waitRater  = flag.Int("wait_rater", 500, "Number of miliseconds to wait for rater to start and cache")

	fileContent1 = `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,default,*voice,dsafdsaf,*rated,*out,cgrates.org,call,1001,1001,+4986517174963,2013-11-07 08:42:25 +0000 UTC,2013-11-07 08:42:26 +0000 UTC,10s,1.0100,val_extra3,"",val_extra1
dbafe9c8614c785a65aabd116dd3959c3c56f7f7,default,*voice,dsafdsag,*rated,*out,cgrates.org,call,1001,1001,+4986517174964,2013-11-07 09:42:25 +0000 UTC,2013-11-07 09:42:26 +0000 UTC,20s,1.0100,val_extra3,"",val_extra1
`

	fileContent2 = `accid21;*prepaid;itsyscom.com;1001;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*postpaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
#accid1;*pseudoprepaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;12;val_extra3;"";val_extra1
accid23;*rated;cgrates.org;1001;086517174963;2013-02-03 19:54:00;26;val_extra3;"";val_extra1`

	csvTests = []func(t *testing.T){
		testCsvITCreateCdrDirs,
		testCsvITInitConfig,
		testCsvITInitCdrDb,
		testCsvITResetDataDb,
		testCsvITStartEngine,
		testCsvITRpcConn,
		testCsvITHandleCdr1File,
		testCsvITHandleCdr2File,
		testCsvITProcessedFiles,
		//testCsvITAnalyseCDRs,
		testCsvITKillEngine,
	}
)

func TestCsvReadFile(t *testing.T) {
	csvCfgPath = path.Join(*dataDir, "conf", "samples", "ers")
	for _, test := range csvTests {
		t.Run("TestCsvReadFile", test)
	}
}
func testCsvITInitConfig(t *testing.T) {
	var err error
	if csvCfg, err = config.NewCGRConfigFromPath(csvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func testCsvITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testCsvITResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func testCsvITCreateCdrDirs(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out", "/tmp/ers2/in", "/tmp/ers2/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testCsvITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCsvITRpcConn(t *testing.T) {
	var err error
	csvRPC, err = jsonrpc.Dial("tcp", csvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func testCsvITHandleCdr1File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/ers/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of first .xml config
func testCsvITHandleCdr2File(t *testing.T) {
	fileName := "file2.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent2), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/ers2/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func testCsvITProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(1 * time.Second))
	if outContent1, err := ioutil.ReadFile("/tmp/ers/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContent1 != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", fileContent1, string(outContent1))
	}
	if outContent2, err := ioutil.ReadFile("/tmp/ers2/out/file2.csv"); err != nil {
		t.Error(err)
	} else if fileContent2 != string(outContent2) {
		t.Errorf("Expecting: %q, received: %q", fileContent1, string(outContent2))
	}
}

func testCsvITAnalyseCDRs(t *testing.T) {
	time.Sleep(500 * time.Millisecond)
	var reply []*engine.ExternalCDR
	if err := csvRPC.Call(utils.ApierV2GetCDRs, utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 5 { // 1 injected, 1 rated, 1 *raw and it's pair in *default run
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := csvRPC.Call(utils.ApierV2GetCDRs, utils.RPCCDRsFilter{DestinationPrefixes: []string{"08651"}},
		&reply); err == nil || err.Error() != utils.NotFoundCaps {
		t.Error("Unexpected error: ", err) // Original 08651 was converted
	}
}

func testCsvITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
