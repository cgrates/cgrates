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

package v1

import (
	"net/rpc"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	eeSCfgPath   string
	eeSCfg       *config.CGRConfig
	eeSRPC       *rpc.Client
	eeSConfigDIR string //run tests for specific configuration

	sTestsEEs = []func(t *testing.T){
		testEEsPrepareFolder,
		testEEsInitCfg,
		testEEsInitDataDb,
		testEEsResetStorDb,
		testEEsStartEngine,
		testEEsRPCConn,
		testEEsAddCDRs,
		testEEsExportCDRs,
		testEEsVerifyExports,
		testEEsExportCDRsMultipleExporters,
		testEEsVerifyExportsMultipleExporters,
		testEEsKillEngine,
		testEEsCleanFolder,
	}
)

//Test start here
func TestExportCDRs(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		eeSConfigDIR = "ees_internal"
	case utils.MetaMySQL:
		eeSConfigDIR = "ees_mysql"
	case utils.MetaMongo:
		eeSConfigDIR = "ees_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsEEs {
		t.Run(eeSConfigDIR, stest)
	}
}

func testEEsPrepareFolder(t *testing.T) {
	for _, dir := range []string{"/tmp/testCSV", "/tmp/testCSV2", "/tmp/testCSV3"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testEEsInitCfg(t *testing.T) {
	var err error
	eeSCfgPath = path.Join(*dataDir, "conf", "samples", eeSConfigDIR)
	eeSCfg, err = config.NewCGRConfigFromPath(eeSCfgPath)
	if err != nil {
		t.Fatal(err)
	}
}

func testEEsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(eeSCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testEEsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(eeSCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testEEsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(eeSCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testEEsRPCConn(t *testing.T) {
	var err error
	eeSRPC, err = newRPCClient(eeSCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testEEsAddCDRs(t *testing.T) {
	//add a default charger
	chargerProfile := &ChargerWithOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        utils.MetaRaw,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var result string
	if err := eeSRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	storedCdrs := []*engine.CDR{
		{CGRID: "Cdr1",
			OrderID: 1, ToR: utils.MetaVoice, OriginID: "OriginCDR1", OriginHost: "192.168.1.1", Source: "test",
			RequestType: utils.MetaNone, Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC),
			AnswerTime: time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC), RunID: utils.MetaDefault, Usage: 10 * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		{CGRID: "Cdr2",
			OrderID: 2, ToR: utils.MetaVoice, OriginID: "OriginCDR2", OriginHost: "192.168.1.1", Source: "test2",
			RequestType: utils.MetaNone, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC),
			AnswerTime: time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC), RunID: utils.MetaDefault, Usage: 5 * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		{CGRID: "Cdr3",
			OrderID: 3, ToR: utils.MetaVoice, OriginID: "OriginCDR3", OriginHost: "192.168.1.1", Source: "test2",
			RequestType: utils.MetaNone, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC),
			AnswerTime: time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC), RunID: utils.MetaDefault, Usage: 30 * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		{CGRID: "Cdr4",
			OrderID: 4, ToR: utils.MetaVoice, OriginID: "OriginCDR4", OriginHost: "192.168.1.1", Source: "test3",
			RequestType: utils.MetaNone, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC),
			AnswerTime: time.Date(2018, 10, 4, 15, 3, 10, 0, time.UTC), RunID: utils.MetaDefault, Usage: 0,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	for _, cdr := range storedCdrs {
		var reply string
		if err := eeSRPC.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithOpts{CDR: cdr}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
}

func testEEsExportCDRs(t *testing.T) {
	attr := &utils.ArgExportCDRs{
		ExporterIDs: []string{"CSVExporter"},
		Verbose:     true,
	}
	var rply map[string]interface{}
	if err := eeSRPC.Call(utils.APIerSv1ExportCDRs, &attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if len(rply) != 1 {
		t.Errorf("Expected %+v, received: %+v", 1, len(rply))
	} else {
		val, _ := rply["CSVExporter"]
		for k, v := range val.(map[string]interface{}) {
			switch k {
			case utils.FirstExpOrderID:
				if v != 1.0 {
					t.Errorf("Expected %+v, received: %+v", 1.0, v)
				}
			case utils.LastExpOrderID:
				if v != 4.0 {
					t.Errorf("Expected %+v, received: %+v", 4.0, v)
				}
			case utils.NumberOfEvents:
				if v != 4.0 {
					t.Errorf("Expected %+v, received: %+v", 4.0, v)
				}
			case utils.TotalCost:
				if v != 4.04 {
					t.Errorf("Expected %+v, received: %+v", 4.04, v)
				}

			}
		}
	}
}

func testEEsVerifyExports(t *testing.T) {
	time.Sleep(time.Second + 600*time.Millisecond)
	var files []string
	err := filepath.Walk("/tmp/testCSV/", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, utils.CSVSuffix) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Errorf("Expected %+v, received: %+v", 1, len(files))
	}
	eCnt := "Cdr1,*raw,*voice,OriginCDR1,*none,cgrates.org,call,1001,1001,+4986517174963,2018-10-04T15:03:10Z,2018-10-04T15:03:10Z,10,1.01\n" +
		"Cdr2,*raw,*voice,OriginCDR2,*none,cgrates.org,call,1001,1001,+4986517174963,2018-10-04T15:03:10Z,2018-10-04T15:03:10Z,5,1.01\n" +
		"Cdr3,*raw,*voice,OriginCDR3,*none,cgrates.org,call,1001,1001,+4986517174963,2018-10-04T15:03:10Z,2018-10-04T15:03:10Z,30,1.01\n" +
		"Cdr4,*raw,*voice,OriginCDR4,*none,cgrates.org,call,1001,1001,+4986517174963,2018-10-04T15:03:10Z,2018-10-04T15:03:10Z,0,1.01\n"
	if outContent1, err := os.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if len(eCnt) != len(string(outContent1)) {
		t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt), len(string(outContent1)))
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testEEsExportCDRsMultipleExporters(t *testing.T) {
	attr := &utils.ArgExportCDRs{
		ExporterIDs: []string{"CSVExporter", "CSVExporter2"},
		Verbose:     true,
	}
	var rply map[string]interface{}
	if err := eeSRPC.Call(utils.APIerSv1ExportCDRs, &attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if len(rply) != 2 {
		t.Errorf("Expected %+v, received: %+v", 1, len(rply))
	} else {
		for _, expID := range []string{"CSVExporter", "CSVExporter2"} {
			val, _ := rply[expID]
			for k, v := range val.(map[string]interface{}) {
				switch k {
				case utils.FirstExpOrderID:
					if v != 1.0 {
						t.Errorf("Expected %+v, received: %+v", 1.0, v)
					}
				case utils.LastExpOrderID:
					if v != 4.0 {
						t.Errorf("Expected %+v, received: %+v", 4.0, v)
					}
				case utils.NumberOfEvents:
					if v != 4.0 {
						t.Errorf("Expected %+v, received: %+v", 4.0, v)
					}
				case utils.TotalCost:
					if v != 4.04 {
						t.Errorf("Expected %+v, received: %+v", 4.04, v)
					}

				}
			}
		}
	}
}

func testEEsVerifyExportsMultipleExporters(t *testing.T) {
	time.Sleep(time.Second)
	var files []string
	err := filepath.Walk("/tmp/testCSV2/", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, utils.CSVSuffix) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Errorf("Expected %+v, received: %+v", 1, len(files))
	}
	eCnt := "Cdr1,*raw,*voice,OriginCDR1,*none,cgrates.org,call,1001,1001,+4986517174963,2018-10-04T15:03:10Z,2018-10-04T15:03:10Z,10,1.01\n" +
		"Cdr2,*raw,*voice,OriginCDR2,*none,cgrates.org,call,1001,1001,+4986517174963,2018-10-04T15:03:10Z,2018-10-04T15:03:10Z,5,1.01\n" +
		"Cdr3,*raw,*voice,OriginCDR3,*none,cgrates.org,call,1001,1001,+4986517174963,2018-10-04T15:03:10Z,2018-10-04T15:03:10Z,30,1.01\n" +
		"Cdr4,*raw,*voice,OriginCDR4,*none,cgrates.org,call,1001,1001,+4986517174963,2018-10-04T15:03:10Z,2018-10-04T15:03:10Z,0,1.01\n"
	if outContent1, err := os.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if len(eCnt) != len(string(outContent1)) {
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testEEsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testEEsCleanFolder(t *testing.T) {
	for _, dir := range []string{"/tmp/testCSV", "/tmp/testCSV2", "/tmp/testCSV3"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}
