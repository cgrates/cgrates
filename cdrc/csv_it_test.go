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
var cdrcCfgs []*config.CdrcCfg
var cdrcCfg *config.CdrcCfg
var cdrcRpc *rpc.Client

var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var waitRater = flag.Int("wait_rater", 500, "Number of miliseconds to wait for rater to start and cache")

var fileContent1 = `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,default,*voice,dsafdsaf,*rated,*out,cgrates.org,call,1001,1001,+4986517174963,2013-11-07 08:42:25 +0000 UTC,2013-11-07 08:42:26 +0000 UTC,10s,1.0100,val_extra3,"",val_extra1
dbafe9c8614c785a65aabd116dd3959c3c56f7f7,default,*voice,dsafdsag,*rated,*out,cgrates.org,call,1001,1001,+4986517174964,2013-11-07 09:42:25 +0000 UTC,2013-11-07 09:42:26 +0000 UTC,20s,1.0100,val_extra3,"",val_extra1
`

var fileContent2 = `accid21;*prepaid;itsyscom.com;1001;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*postpaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
#accid1;*pseudoprepaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;12;val_extra3;"";val_extra1
accid23;*rated;cgrates.org;1001;086517174963;2013-02-03 19:54:00;26;val_extra3;"";val_extra1`

func TestCsvITInitConfig(t *testing.T) {
	var err error
	csvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrccsv")
	if csvCfg, err = config.NewCGRConfigFromFolder(csvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestCsvITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func TestCsvITResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func TestCsvITCreateCdrDirs(t *testing.T) {
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
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestCsvITRpcConn(t *testing.T) {
	var err error
	cdrcRpc, err = jsonrpc.Dial("tcp", csvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestCsvITHandleCdr1File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cdrctests/csvit1/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of first .xml config
func TestCsvITHandleCdr2File(t *testing.T) {
	fileName := "file2.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent2), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cdrctests/csvit2/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestCsvITProcessedFiles(t *testing.T) {
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

func TestCsvITAnalyseCDRs(t *testing.T) {
	time.Sleep(500 * time.Millisecond)
	var reply []*engine.ExternalCDR
	if err := cdrcRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 5 { // 1 injected, 1 rated, 1 *raw and it's pair in *default run
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := cdrcRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{DestinationPrefixes: []string{"08651"}}, &reply); err == nil || err.Error() != utils.NotFoundCaps {
		t.Error("Unexpected error: ", err) // Original 08651 was converted
	}
}

func TestCsvITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Begin tests for cdrc csv with new filters
var fileContent1_2 = `accid21;*prepaid;itsyscom.com;1002;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*postpaid;itsyscom.com;1002;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
accid23;*rated;cgrates.org;1001;086517174963;2013-02-03 19:54:00;26;val_extra3;"";val_extra1`

func TestCsvIT2InitConfig(t *testing.T) {
	var err error
	csvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrccsv")
	if csvCfg, err = config.NewCGRConfigFromFolder(csvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestCsvIT2InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func TestCsvIT2CreateCdrDirs(t *testing.T) {
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

func TestCsvIT2StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestCsvIT2RpcConn(t *testing.T) {
	var err error
	cdrcRpc, err = jsonrpc.Dial("tcp", csvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Scenario out of first .xml config
func TestCsvIT2HandleCdr2File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1_2), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/csvwithfilter1/csvit1/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestCsvIT2ProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent2, err := ioutil.ReadFile("/tmp/csvwithfilter1/csvit1/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContent1_2 != string(outContent2) {
		t.Errorf("Expecting: %q, received: %q", fileContent1_2, string(outContent2))
	}
}

func TestCsvIT2AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := cdrcRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{RequestTypes: []string{utils.META_PREPAID}}, &reply); err != nil {
		t.Error("Unexpected error: ", err) // Original 08651 was converted
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestCsvIT2KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
}

// Begin tests for cdrc csv with new filters
var fileContent1_3 = `accid21;*prepaid;itsyscom.com;1002;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*prepaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
accid23;*prepaid;cgrates.org;1002;086517174963;2013-02-03 19:54:00;76;val_extra3;"";val_extra1`

func TestCsvIT3InitConfig(t *testing.T) {
	var err error
	csvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrccsv")
	if csvCfg, err = config.NewCGRConfigFromFolder(csvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestCsvIT3InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func TestCsvIT3CreateCdrDirs(t *testing.T) {
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

func TestCsvIT3StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestCsvIT3RpcConn(t *testing.T) {
	var err error
	cdrcRpc, err = jsonrpc.Dial("tcp", csvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Scenario out of first .xml config
func TestCsvIT3HandleCdr2File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1_3), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/csvwithfilter2/csvit2/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestCsvIT3ProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent2, err := ioutil.ReadFile("/tmp/csvwithfilter2/csvit2/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContent1_3 != string(outContent2) {
		t.Errorf("Expecting: %q, received: %q", fileContent1_3, string(outContent2))
	}
}

func TestCsvIT3AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestCsvIT3KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
}

// Begin tests for cdrc csv with new filters
var fileContent1_4 = `accid21;*prepaid;itsyscom.com;1002;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1
accid22;*postpaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1
accid23;*postpaid;cgrates.org;1002;086517174963;2013-02-03 19:54:00;76;val_extra3;"";val_extra1`

func TestCsvIT4InitConfig(t *testing.T) {
	var err error
	csvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrccsv")
	if csvCfg, err = config.NewCGRConfigFromFolder(csvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestCsvIT4InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func TestCsvIT4CreateCdrDirs(t *testing.T) {
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

func TestCsvIT4StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestCsvIT4RpcConn(t *testing.T) {
	var err error
	cdrcRpc, err = jsonrpc.Dial("tcp", csvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Scenario out of first .xml config
func TestCsvIT4HandleCdr2File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1_4), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/csvwithfilter3/csvit3/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestCsvIT4ProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent4, err := ioutil.ReadFile("/tmp/csvwithfilter3/csvit3/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContent1_4 != string(outContent4) {
		t.Errorf("Expecting: %q, received: %q", fileContent1_4, string(outContent4))
	}
}

func TestCsvIT4AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestCsvIT4KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Begin tests for cdrc csv with new filters
var fileContent1_5 = `accid21;*prepaid;itsyscom.com;1002;086517174963;2013-02-03 19:54:00;62;val_extra3;"";val_extra1;10.10.10.10
accid22;*postpaid;itsyscom.com;1001;+4986517174963;2013-02-03 19:54:00;123;val_extra3;"";val_extra1;11.10.10.10
accid23;*postpaid;cgrates.org;1002;086517174963;2013-02-03 19:54:00;76;val_extra3;"";val_extra1;12.10.10.10
accid24;*postpaid;cgrates.org;1001;+4986517174963;2013-02-03 19:54:00;76;val_extra3;"";val_extra1;12.10.10.10`

func TestCsvIT5InitConfig(t *testing.T) {
	var err error
	csvCfgPath = path.Join(*dataDir, "conf", "samples", "cdrccsv")
	if csvCfg, err = config.NewCGRConfigFromFolder(csvCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestCsvIT5InitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(csvCfg); err != nil {
		t.Fatal(err)
	}
}

func TestCsvIT5CreateCdrDirs(t *testing.T) {
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

func TestCsvIT5StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(csvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestCsvIT5RpcConn(t *testing.T) {
	var err error
	cdrcRpc, err = jsonrpc.Dial("tcp", csvCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestCsvIT5AddFilters(t *testing.T) {
	filter := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_CDRC_ACC",
		Rules: []*engine.FilterRule{
			{
				Type:      "*string",
				FieldName: "3",
				Values:    []string{"1002"},
			},
		},
	}
	var result string
	if err := cdrcRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	filter2 := &engine.Filter{
		Tenant: "itsyscom.com",
		ID:     "FLTR_CDRC_ACC",
		Rules: []*engine.FilterRule{
			{
				Type:      "*string",
				FieldName: "3",
				Values:    []string{"1001"},
			},
		},
	}
	if err := cdrcRpc.Call("ApierV1.SetFilter", filter2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

// Scenario out of first .xml config
func TestCsvIT5HandleCdr2File(t *testing.T) {
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1_5), 0644); err != nil {
		t.Fatal(err.Error())
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/csvwithfilter4/csvit4/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestCsvIT5ProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent4, err := ioutil.ReadFile("/tmp/csvwithfilter4/csvit4/out/file1.csv"); err != nil {
		t.Error(err)
	} else if fileContent1_5 != string(outContent4) {
		t.Errorf("Expecting: %q, received: %q", fileContent1_5, string(outContent4))
	}
}

func TestCsvIT5AnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestCsvIT5KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
