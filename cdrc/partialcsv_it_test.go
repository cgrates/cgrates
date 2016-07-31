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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var partpartcsvCfgPath string
var partcsvCfg *config.CGRConfig
var partcsvRPC *rpc.Client
var partcsvCDRCDirIn, partcsvCDRCDirOut string

var partCsvFileContent1 = `4986517174963,004986517174964,DE-National,04.07.2016 18:58:55,04.07.2016 18:58:55,1,65,Peak,0.014560,498651,partial
4986517174964,004986517174963,DE-National,04.07.2016 20:58:55,04.07.2016 20:58:55,0,74,Offpeak,0.003360,498651,complete
`

var partCsvFileContent2 = `4986517174963,004986517174964,DE-National,04.07.2016 19:00:00,04.07.2016 18:58:55,0,15,Offpeak,0.003360,498651,partial`

var eCacheDumpFile1 = `4986517174963_004986517174964_04.07.2016 18:58:55,1467651600,*rated,086517174963,+4986517174964,2016-07-04T18:58:55+02:00,2016-07-04T18:58:55+02:00,15,-1.00000
4986517174963_004986517174964_04.07.2016 18:58:55,1467651535,*rated,086517174963,+4986517174964,2016-07-04T18:58:55+02:00,2016-07-04T18:58:55+02:00,65,-1.00000
`

func TestPartcsvITInitConfig(t *testing.T) {
	if !*testIT {
		return
	}
	var err error
	partpartcsvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrc_partcsv")
	if partcsvCfg, err = config.NewCGRConfigFromFolder(partpartcsvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestPartcsvITInitCdrDb(t *testing.T) {
	if !*testIT {
		return
	}
	if err := engine.InitStorDb(partcsvCfg); err != nil {
		t.Fatal(err)
	}
}

func TestPartcsvITCreateCdrDirs(t *testing.T) {
	if !*testIT {
		return
	}
	for _, cdrcProfiles := range partcsvCfg.CdrcProfiles {
		for i, cdrcInst := range cdrcProfiles {
			if i == 0 {
				partcsvCDRCDirIn, partcsvCDRCDirOut = cdrcInst.CdrInDir, cdrcInst.CdrOutDir
			}
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

func TestPartcsvITStartEngine(t *testing.T) {
	if !*testIT {
		return
	}
	if _, err := engine.StopStartEngine(partpartcsvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestPartcsvITRpcConn(t *testing.T) {
	if !*testIT {
		return
	}
	var err error
	partcsvRPC, err = jsonrpc.Dial("tcp", partcsvCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestPartcsvITHandleCdr1File(t *testing.T) {
	if !*testIT {
		return
	}
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(partCsvFileContent1), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := os.Rename(tmpFilePath, path.Join(partcsvCDRCDirIn, fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of first .xml config
func TestPartcsvITHandleCdr2File(t *testing.T) {
	if !*testIT {
		return
	}
	fileName := "file2.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(partCsvFileContent2), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := os.Rename(tmpFilePath, path.Join(partcsvCDRCDirIn, fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestPartcsvITProcessedFiles(t *testing.T) {
	if !*testIT {
		return
	}
	time.Sleep(time.Duration(2 * time.Second))
	if outContent1, err := ioutil.ReadFile(path.Join(partcsvCDRCDirOut, "file1.csv")); err != nil {
		t.Error(err)
	} else if partCsvFileContent1 != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", partCsvFileContent1, string(outContent1))
	}
	if outContent2, err := ioutil.ReadFile(path.Join(partcsvCDRCDirOut, "file2.csv")); err != nil {
		t.Error(err)
	} else if partCsvFileContent2 != string(outContent2) {
		t.Errorf("Expecting: %q, received: %q", partCsvFileContent2, string(outContent2))
	}
	filesInDir, _ := ioutil.ReadDir(partcsvCDRCDirOut)
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		if strings.HasPrefix(file.Name(), "4986517174963_004986517174964") {
			fileName = file.Name()
			break
		}
	}
	if contentCacheDump, err := ioutil.ReadFile(path.Join(partcsvCDRCDirOut, fileName)); err != nil {
		t.Error(err)
	} else if eCacheDumpFile1 != string(contentCacheDump) {
		t.Errorf("Expecting: %q, received: %q", eCacheDumpFile1, string(contentCacheDump))
	}
}

func TestPartcsvITAnalyseCDRs(t *testing.T) {
	if !*testIT {
		return
	}
	var reply []*engine.ExternalCDR
	if err := partcsvRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := partcsvRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{DestinationPrefixes: []string{"+4986517174963"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}

}

func TestPartcsvITKillEngine(t *testing.T) {
	if !*testIT {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
