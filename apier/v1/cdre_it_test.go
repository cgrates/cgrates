//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package v1

import (
	"net/rpc"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdreCfgPath   string
	cdreCfg       *config.CGRConfig
	cdreRPC       *rpc.Client
	cdreDataDir   = "/usr/share/cgrates"
	cdreConfigDIR string //run tests for specific configuration

	sTestsCDRE = []func(t *testing.T){
		testCDReInitCfg,
		testCDReInitDataDb,
		testCDReResetStorDb,
		testCDReStartEngine,
		testCDReRPCConn,
		testCDReAddCDRs,
		testCDReExportCDRs,
		testCDReExportCDRs2,
		testCDReFromFolder,
		testCDReProcessExternalCdr,
		testCDReKillEngine,
	}
)

// Test start here
func TestCDRExport(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		cdreConfigDIR = "cdrewithfilter_internal"
	case utils.MetaMySQL:
		cdreConfigDIR = "cdrewithfilter_mysql"
	case utils.MetaMongo:
		cdreConfigDIR = "cdrewithfilter_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCDRE {
		t.Run(cdreConfigDIR, stest)
	}
}

func TestCDRExportWithAttributeS(t *testing.T) {
	cdreConfigDIR = "cdrewithattributes"
	tests := sTestsCDRE[0:6]
	tests = append(tests, testCDReAddAttributes)
	tests = append(tests, testCDReExportCDRsWithAttributes)
	tests = append(tests, testCDReKillEngine)
	for _, stest := range tests {
		t.Run(cdreConfigDIR, stest)
	}
}

func testCDReInitCfg(t *testing.T) {
	var err error
	cdreCfgPath = path.Join(alsPrfDataDir, "conf", "samples", cdreConfigDIR)
	cdreCfg, err = config.NewCGRConfigFromPath(cdreCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	cdreCfg.DataFolderPath = alsPrfDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(cdreCfg)
}

func testCDReInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdreCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testCDReResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(cdreCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testCDReStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdreCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCDReRPCConn(t *testing.T) {
	var err error
	cdreRPC, err = newRPCClient(cdreCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testCDReAddCDRs(t *testing.T) {
	storedCdrs := []*engine.CDR{
		{CGRID: "Cdr1",
			OrderID: 123, ToR: utils.VOICE, OriginID: "OriginCDR1", OriginHost: "192.168.1.1", Source: "test",
			RequestType: utils.META_RATED, Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.MetaDefault, Usage: time.Duration(10) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		{CGRID: "Cdr2",
			OrderID: 123, ToR: utils.VOICE, OriginID: "OriginCDR2", OriginHost: "192.168.1.1", Source: "test2",
			RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.MetaDefault, Usage: time.Duration(5) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		{CGRID: "Cdr3",
			OrderID: 123, ToR: utils.VOICE, OriginID: "OriginCDR3", OriginHost: "192.168.1.1", Source: "test2",
			RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.MetaDefault, Usage: time.Duration(30) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		{CGRID: "Cdr4",
			OrderID: 123, ToR: utils.VOICE, OriginID: "OriginCDR4", OriginHost: "192.168.1.1", Source: "test3",
			RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Time{}, RunID: utils.MetaDefault, Usage: time.Duration(0) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	for _, cdr := range storedCdrs {
		var reply string
		if err := cdreRPC.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithArgDispatcher{CDR: cdr}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	time.Sleep(100 * time.Millisecond)
}

func testCDReExportCDRs(t *testing.T) {
	attr := ArgExportCDRs{
		ExportArgs: map[string]any{
			utils.ExportTemplate: "TemplateWithFilter",
		},
		Verbose: true,
	}
	var rply *RplExportedCDRs
	if err := cdreRPC.Call(utils.APIerSv1ExportCDRs, attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rply.ExportedCGRIDs) != 2 {
		t.Errorf("Unexpected number of CDR exported: %s ", utils.ToJSON(rply))
	}
}

func testCDReExportCDRs2(t *testing.T) {

	storedCdrs := []*engine.CDR{
		{CGRID: "Cdr5",
			OrderID: 1234, ToR: utils.VOICE, OriginID: "OriginCDR1", OriginHost: "192.168.1.1", Source: "test",
			RequestType: utils.META_RATED, Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.MetaDefault, Usage: time.Duration(10) * time.Second,
			ExtraFields: map[string]string{"DisconnectCause": "ORIGINATOR_CANCEL"}, Cost: 13.7,
		},
		{CGRID: "Cdr6",
			OrderID: 124343, ToR: utils.VOICE, OriginID: "OriginCDR2", OriginHost: "192.168.1.1", Source: "test2",
			RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.MetaDefault, Usage: time.Duration(5) * time.Second,
			ExtraFields: map[string]string{"DisconnectCause": "UNALLOCATED_NUMBER"}, Cost: 2.21,
		},
		{CGRID: "Cdr7",
			OrderID: 124343, ToR: utils.VOICE, OriginID: "OriginCDR2", OriginHost: "192.168.1.1", Source: "test2",
			RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
			Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.MetaDefault, Usage: time.Duration(5) * time.Second,
			ExtraFields: map[string]string{"DisconnectCause": "USER_BUSY"}, Cost: 2.21,
		},
	}
	for _, cdr := range storedCdrs {
		var reply string
		if err := cdreRPC.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithArgDispatcher{CDR: cdr}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	attr := ArgExportCDRs{
		ExportArgs: map[string]any{
			utils.ExportTemplate: "TemplateWithFilter",
			utils.FilterIDs:      []string{"*string:~*req.DisconnectCause:ORIGINATOR_CANCEL;USER_BUSY"},
			utils.ExportPath:     "/tmp",
			utils.ExportFileName: "TestFilteredCsv.csv",
		},
		Verbose: true,
	}
	var rply *RplExportedCDRs
	expExportedCdrs := []string{"Cdr5", "Cdr7"}
	if err := cdreRPC.Call(utils.APIerSv1ExportCDRs, attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else {
		sort.Strings(rply.ExportedCGRIDs)
		if !reflect.DeepEqual(rply.ExportedCGRIDs, expExportedCdrs) {
			t.Errorf("Expected CDRs %+v,Received %+v ", expExportedCdrs, rply.ExportedCGRIDs)
		}
	}

}

func testCDReFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := cdreRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

// Test CDR from external sources
func testCDReProcessExternalCdr(t *testing.T) {
	cdr := &engine.ExternalCDR{
		ToR:         utils.VOICE,
		OriginID:    "testextcdr1",
		OriginHost:  "127.0.0.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1003",
		Subject:     "1003",
		Destination: "1001",
		SetupTime:   "2014-08-04T13:00:00Z",
		AnswerTime:  "2014-08-04T13:00:07Z",
		Usage:       "1s",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var reply string
	if err := cdreRPC.Call(utils.CDRsV1ProcessExternalCDR, engine.ExternalCDRWithArgDispatcher{ExternalCDR: cdr}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(50 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{OriginIDs: []string{"testextcdr1"}}
	if err := cdreRPC.Call(utils.APIerSv2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Errorf("Unexpected number of CDRs returned: %v, cdrs=%s ", len(cdrs), utils.ToJSON(cdrs))
		return
	} else {
		for _, c := range cdrs {
			if c.RunID == utils.MetaRaw && c.Cost != -1 {
				t.Errorf("Expected for %s cdr cost to be -1, received: %v", utils.MetaRaw, c.Cost)
			}
			if c.RunID == utils.MetaDefault && c.Cost != 0.3 {
				t.Errorf("Expected for *default cdr cost to be 0.3, received: %v", c.Cost)
			}
			if c.RunID == utils.MetaDefault {
				acdr, err := engine.NewCDRFromExternalCDR(c, "")
				if err != nil {
					t.Error(err)
					return
				}
				if acdr.CostDetails == nil {
					t.Errorf("CostDetails should not be nil")
					return
				}
				if acdr.CostDetails.Usage == nil {
					t.Errorf("CostDetails for processed cdr has usage nil")
				}
				if acdr.CostDetails.Cost == nil {
					t.Errorf("CostDetails for processed cdr has cost nil")
				}
			}
			if c.Usage != "1s" {
				t.Errorf("Expected 1s,received %s", c.Usage)
			}
			if c.Source != utils.UNIT_TEST {
				t.Errorf("Expected %s,received %s", utils.UNIT_TEST, c.Source)
			}
			if c.ToR != utils.VOICE {
				t.Errorf("Expected %s,received %s", utils.VOICE, c.ToR)
			}
			if c.RequestType != utils.META_RATED {
				t.Errorf("Expected %s,received %s", utils.META_RATED, c.RequestType)
			}
			if c.Category != "call" {
				t.Errorf("Expected call,received %s", c.Category)
			}
			if c.Account != "1003" {
				t.Errorf("Expected 1003,received %s", c.Account)
			}
			if c.Subject != "1003" {
				t.Errorf("Expected 1003,received %s", c.Subject)
			}
			if c.Destination != "1001" {
				t.Errorf("Expected 1001,received %s", c.Destination)
			}
			if !reflect.DeepEqual(c.ExtraFields, cdr.ExtraFields) {
				t.Errorf("Expected %s,received %s", utils.ToJSON(c.ExtraFields), utils.ToJSON(cdr.ExtraFields))
			}
		}
	}
}

func testCDReAddAttributes(t *testing.T) {
	alsPrf := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_CDRE",
			Contexts:  []string{"*cdre"},
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.NewRSRParsersMustCompile("ATTR_SUBJECT", true, utils.INFIELD_SEP),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Category,
					Value: config.NewRSRParsersMustCompile("ATTR_CATEGORY", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var result string
	if err := cdreRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := cdreRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_CDRE"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
}

func testCDReExportCDRsWithAttributes(t *testing.T) {
	attr := ArgExportCDRs{
		ExportArgs: map[string]any{
			utils.ExportTemplate: "TemplateWithAttributeS",
		},
		Verbose: true,
	}
	var rply *RplExportedCDRs
	if err := cdreRPC.Call(utils.APIerSv1ExportCDRs, attr, &rply); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(rply.ExportedCGRIDs) != 2 {
		t.Errorf("Unexpected number of CDR exported: %s ", utils.ToJSON(rply))
	}
	fileContent1 := `Cdr3,*default,test2,OriginCDR3,cgrates.org,ATTR_CATEGORY,1001,ATTR_SUBJECT,+4986517174963,30s,-1.0000
Cdr2,*default,test2,OriginCDR2,cgrates.org,ATTR_CATEGORY,1001,ATTR_SUBJECT,+4986517174963,5s,-1.0000
`
	fileContent2 := `Cdr2,*default,test2,OriginCDR2,cgrates.org,ATTR_CATEGORY,1001,ATTR_SUBJECT,+4986517174963,5s,-1.0000
Cdr3,*default,test2,OriginCDR3,cgrates.org,ATTR_CATEGORY,1001,ATTR_SUBJECT,+4986517174963,30s,-1.0000
`
	if outContent1, err := os.ReadFile(rply.ExportedPath); err != nil {
		t.Error(err)
	} else if fileContent1 != string(outContent1) && fileContent2 != string(outContent1) {
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", fileContent1, string(outContent1))
	}
}

func testCDReKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
