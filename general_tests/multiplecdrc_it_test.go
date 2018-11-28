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

package general_tests

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

var cfgPath string
var cfg *config.CGRConfig
var rater *rpc.Client

var testCalls = flag.Bool("calls", false, "Run test calls simulation, not by default.")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var storDbType = flag.String("stordb_type", "mysql", "The type of the storDb database <mysql>")
var waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")

func TestMCDRCLoadConfig(t *testing.T) {
	var err error
	cfgPath = path.Join(*dataDir, "conf", "samples", "multiplecdrc")
	if cfg, err = config.NewCGRConfigFromFolder(cfgPath); err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func TestMCDRCResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestMCDRCEmptyTables(t *testing.T) {
	if err := engine.InitStorDb(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestMCDRCCreateCdrDirs(t *testing.T) {
	for _, cdrcProfiles := range cfg.CdrcProfiles {
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
func TestMCDRCStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestMCDRCRpcConn(t *testing.T) {
	var err error
	rater, err = jsonrpc.Dial("tcp", cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Test here LoadTariffPlanFromFolder
func TestMCDRCApierLoadTariffPlanFromFolder(t *testing.T) {
	reply := ""
	// Simple test that command is executed without errors
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testtp")}
	if err := rater.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// The default scenario, out of cdrc defined in .cfg file
func TestMCDRCHandleCdr1File(t *testing.T) {
	var fileContent1 = `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,default,*voice,dsafdsaf,rated,*out,cgrates.org,call,1001,1001,+4986517174963,2013-11-07 08:42:25 +0000 UTC,2013-11-07 08:42:26 +0000 UTC,10000000000,1.0100,val_extra3,"",val_extra1
dbafe9c8614c785a65aabd116dd3959c3c56f7f7,default,*voice,dsafdsag,rated,*out,cgrates.org,call,1001,1001,+4986517174964,2013-11-07 09:42:25 +0000 UTC,2013-11-07 09:42:26 +0000 UTC,20000000000,1.0100,val_extra3,"",val_extra1
`
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cgrates/cdrc1/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of first .xml config
func TestMCDRCHandleCdr2File(t *testing.T) {
	var fileContent = `616350843,20131022145011,20131022172857,3656,1001,,,data,mo,640113,0.000000,1.222656,1.222660
616199016,20131022154924,20131022154955,3656,1001,086517174963,,voice,mo,31,0.000000,0.000000,0.000000
800873243,20140516063739,20140516063739,9774,1001,+49621621391,,sms,mo,1,0.00000,0.00000,0.00000`
	fileName := "file2.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cgrates/cdrc2/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of second .xml config
func TestMCDRCHandleCdr3File(t *testing.T) {
	var fileContent = `4986517174960;4986517174963;Sample Mobile;08.04.2014  22:14:29;08.04.2014  22:14:29;1;193;Offeak;0,072728833;31619
4986517174960;4986517174964;National;08.04.2014  20:34:55;08.04.2014  20:34:55;1;21;Offeak;0,0079135;311`
	fileName := "file3.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cgrates/cdrc3/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestMCDRCStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
