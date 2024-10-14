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
package general_tests

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	brodcastCfgPath         string
	brodcastInternalCfgPath string
	brodcastInternalCfgDIR  string
	brodcastCfg             *config.CGRConfig
	brodcastInternalCfg     *config.CGRConfig
	brodcastRPC             *birpc.Client
	brodcastInternalRPC     *birpc.Client

	sTestBrodcastIt = []func(t *testing.T){
		testbrodcastItLoadConfig,
		testbrodcastItResetDataDB,
		testbrodcastItResetStorDb,
		testbrodcastItStartEngine,
		testbrodcastItRPCConn,
		testbrodcastItLoadFromFolder,

		testbrodcastItProccessEvent,
		testbrodcastItGetCDRs,

		testbrodcastItStopCgrEngine,
	}
)

func TestBrodcastRPC(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		brodcastInternalCfgDIR = "tutinternal"
	case utils.MetaMySQL:
		brodcastInternalCfgDIR = "tutmysql"
	case utils.MetaMongo:
		brodcastInternalCfgDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestBrodcastIt {
		t.Run(brodcastInternalCfgDIR, stest)
	}
}

// test for 0 balance with session terminate with 1s usage
func testbrodcastItLoadConfig(t *testing.T) {
	var err error
	brodcastCfgPath = path.Join(*utils.DataDir, "conf", "samples", "internal_broadcast_replication")
	if brodcastCfg, err = config.NewCGRConfigFromPath(brodcastCfgPath); err != nil {
		t.Error(err)
	}
	brodcastInternalCfgPath = path.Join(*utils.DataDir, "conf", "samples", brodcastInternalCfgDIR)
	if brodcastInternalCfg, err = config.NewCGRConfigFromPath(brodcastInternalCfgPath); err != nil {
		t.Error(err)
	}
}

func testbrodcastItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(brodcastInternalCfg); err != nil {
		t.Fatal(err)
	}
}

func testbrodcastItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(brodcastInternalCfg); err != nil {
		t.Fatal(err)
	}
}

func testbrodcastItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(brodcastCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(brodcastInternalCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testbrodcastItRPCConn(t *testing.T) {
	brodcastRPC = engine.NewRPCClient(t, brodcastCfg.ListenCfg())
	brodcastInternalRPC = engine.NewRPCClient(t, brodcastInternalCfg.ListenCfg())
}

func testbrodcastItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := brodcastRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	if err := brodcastInternalRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testbrodcastItProccessEvent(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestSSv1ItProcessCDR",
		Event: map[string]any{
			utils.Tenant:       "cgrates.org",
			utils.Category:     utils.Call,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestSSv1It1Brodcast",
			utils.RequestType:  utils.MetaPostpaid,
			utils.AccountField: "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        10 * time.Minute,
		},
	}

	var rply string
	if err := brodcastRPC.Call(context.Background(), utils.SessionSv1ProcessCDR, args, &rply); err != nil {
		t.Fatal(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	time.Sleep(50 * time.Millisecond)
}
func testbrodcastItGetCDRs(t *testing.T) {
	eCDR := &engine.CDR{
		CGRID:       "ad6cb338dea6eaf2e81507623fbd6b00f60c374f",
		RunID:       "*default",
		OrderID:     0,
		OriginHost:  "",
		Source:      "*sessions",
		OriginID:    "TestSSv1It1Brodcast",
		ToR:         "*voice",
		RequestType: "*postpaid",
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
		AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
		Usage:       600000000000,
		ExtraFields: make(map[string]string),
		ExtraInfo:   "NOT_CONNECTED: RALs",
		Partial:     false,
		PreRated:    false,
		CostSource:  "",
		Cost:        -1,
	}
	var cdrs []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}}}
	if err := brodcastRPC.Call(context.Background(), utils.CDRsV1GetCDRs, &args, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Fatal("Unexpected number of CDRs returned: ", len(cdrs))
	}
	cdrs[0].OrderID = 0 // reset the OrderID
	if !reflect.DeepEqual(eCDR, cdrs[0]) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(eCDR), utils.ToJSON(cdrs[0]))
	}

	if err := brodcastInternalRPC.Call(context.Background(), utils.CDRsV1GetCDRs, &args, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
	cdrs[0].OrderID = 0                           // reset the OrderID
	cdrs[0].SetupTime = cdrs[0].SetupTime.UTC()   // uniform time
	cdrs[0].AnswerTime = cdrs[0].AnswerTime.UTC() // uniform time
	if !reflect.DeepEqual(eCDR, cdrs[0]) {
		t.Errorf("Expected: %s \n,received: %s", utils.ToJSON(eCDR), utils.ToJSON(cdrs[0]))
	}
}

func testbrodcastItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
