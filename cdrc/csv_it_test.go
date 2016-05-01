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
)

/*
README:

 Enable local tests by passing '-local' to the go test command
 It is expected that the data folder of CGRateS exists at path /usr/share/cgrates/data or passed via command arguments.
 Prior running the tests, create database and users by running:
  mysql -pyourrootpwd < /usr/share/cgrates/data/storage/mysql/create_db_with_users.sql
 What these tests do:
  * Flush tables in storDb.
  * Start engine with default configuration and give it some time to listen (here caching can slow down).
  *
*/

var csvCfgPath string
var csvCfg *config.CGRConfig
var cdrcCfgs []*config.CdrcConfig
var cdrcCfg *config.CdrcConfig
var cdrcRpc *rpc.Client

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.")    // This flag will be passed here via "go test -local" args
var testIT = flag.Bool("integration", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var waitRater = flag.Int("wait_rater", 300, "Number of miliseconds to wait for rater to start and cache")

var fileContent1 = `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,default,*voice,dsafdsaf,*rated,*out,cgrates.org,call,1001,1001,+4986517174963,2013-11-07 08:42:25 +0000 UTC,2013-11-07 08:42:26 +0000 UTC,10s,1.0100,val_extra3,"",val_extra1
dbafe9c8614c785a65aabd116dd3959c3c56f7f7,default,*voice,dsafdsag,*rated,*out,cgrates.org,call,1001,1001,+4986517174964,2013-11-07 09:42:25 +0000 UTC,2013-11-07 09:42:26 +0000 UTC,20s,1.0100,val_extra3,"",val_extra1
`

var fileContent2 = `accid21;*prepaid;itsyscom.com;1001;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*postpaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
#accid1;*pseudoprepaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;12;val_extra3;"";val_extra1
accid23;*rated;cgrates.org;1001;086517174963;2013-02-03 19:54:00;26;val_extra3;"";val_extra1`

func TestCsvITInitConfig(t *testing.T) {
	if !*testIT {
		return
	}
	var err error
	csvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrccsv")
	if csvCfg, err = config.NewCGRConfigFromFolder(csvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestCsvITInitCdrDb(t *testing.T) {
	if !*testIT {
		return
	}
	if err := engine.InitStorDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

/*
func TestCsvITCreateCdrDirs(t *testing.T) {
	if !*testIT {
		return
	}
	for _, cdrcProfiles := range csvCfg.CdrcProfiles {
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


func TestCsvITStartEngine(t *testing.T) {
	if !*testIT {
		return
	}
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}
*/

// Connect rpc client to rater
func TestCsvITRpcConn(t *testing.T) {
	if !*testIT {
		return
	}
	var err error
	cdrcRpc, err = jsonrpc.Dial("tcp", csvCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestCsvITHandleCdr1File(t *testing.T) {
	if !*testIT {
		return
	}
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cdrctests/csvit1/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of first .xml config
func TestCsvITHandleCdr2File(t *testing.T) {
	if !*testIT {
		return
	}
	fileName := "file2.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent2), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cdrctests/csvit2/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestCsvITProcessedFiles(t *testing.T) {
	if !*testIT {
		return
	}
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent1, err := ioutil.ReadFile("/tmp/cdrctests/csvit1/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContent1 != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", fileContent1, string(outContent1))
	}
	if outContent2, err := ioutil.ReadFile("/tmp/cdrctests/csvit2/out/file2.csv"); err != nil {
		t.Error(err)
	} else if fileContent2 != string(outContent2) {
		t.Errorf("Expecting: %q, received: %q", fileContent1, string(outContent2))
	}
}

func TestCsvITKillEngine(t *testing.T) {
	if !*testIT {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
