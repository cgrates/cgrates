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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	ses3CfgDir  string
	ses3CfgPath string
	ses3Cfg     *config.CGRConfig
	ses3RPC     *rpc.Client

	ses3Tests = []func(t *testing.T){
		testSes3ItLoadConfig,
		testSes3ItResetDataDB,
		testSes3ItResetStorDb,
		testSes3ItStartEngine,
		testSes3ItRPCConn,
		testSes3ItLoadFromFolder,
		testSes3ItProcessEvent,
		testSes3ItThreshold1002After,
		testSes3ItStatMetricsAfter,
		testSes3ItProcessEvent,
		testSes3ItThreshold1002After2,
		testSes3ItStatMetricsAfter2,
		testSes3ItStopCgrEngine,
	}
)

func TestSes3ItSessions(t *testing.T) {
	ses3CfgDir = "sessions"
	for _, stest := range ses3Tests {
		t.Run("TestSes3ItTutMysql", stest)
	}
}

func testSes3ItLoadConfig(t *testing.T) {
	ses3CfgPath = path.Join(*dataDir, "conf", "samples", ses3CfgDir)
	if ses3Cfg, err = config.NewCGRConfigFromPath(ses3CfgPath); err != nil {
		t.Error(err)
	}
}

func testSes3ItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(ses3Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes3ItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(ses3Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes3ItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ses3CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSes3ItRPCConn(t *testing.T) {
	var err error
	ses3RPC, err = jsonrpc.Dial("tcp", ses3Cfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func testSes3ItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := ses3RPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testSes3ItProcessEvent(t *testing.T) {
	initUsage := 5 * time.Minute
	args := sessions.V1ProcessMessageArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		ProcessThresholds: true,
		ProcessStats:      true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessEvent",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It2",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       initUsage,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}
	var rply sessions.V1ProcessMessageReply
	if err := ses3RPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	if *rply.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessEvent",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Destination: "1002",
				"OfficeGroup":     "Marketing",
				utils.OriginID:    "TestSSv1It2",
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   "2018-01-07T17:00:00Z",
				utils.AnswerTime:  "2018-01-07T17:00:10Z",
				utils.Usage:       300000000000.0,
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
}

func testSes3ItThreshold1002After(t *testing.T) {
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1001", Hits: 1}
	if err := ses3RPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Tenant, td.Tenant) {
		t.Errorf("expecting: %+v, received: %+v", eTd.Tenant, td.Tenant)
	} else if !reflect.DeepEqual(eTd.ID, td.ID) {
		t.Errorf("expecting: %+v, received: %+v", eTd.ID, td.ID)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd.Hits, td.Hits)
	}
}

func testSes3ItStatMetricsAfter(t *testing.T) {
	var metrics map[string]string
	statMetrics := map[string]string{
		utils.MetaACD: "5m0s",
		utils.MetaASR: "100%",
		utils.MetaTCD: "5m0s",
	}

	if err := ses3RPC.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}, &metrics); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(statMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", statMetrics, metrics)
	}
}

func testSes3ItThreshold1002After2(t *testing.T) {
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1001", Hits: 2}
	if err := ses3RPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Tenant, td.Tenant) {
		t.Errorf("expecting: %+v, received: %+v", eTd.Tenant, td.Tenant)
	} else if !reflect.DeepEqual(eTd.ID, td.ID) {
		t.Errorf("expecting: %+v, received: %+v", eTd.ID, td.ID)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd.Hits, td.Hits)
	}
}

func testSes3ItStatMetricsAfter2(t *testing.T) {
	var metrics map[string]string
	statMetrics := map[string]string{
		utils.MetaACD: "5m0s",
		utils.MetaASR: "100%",
		utils.MetaTCD: "10m0s",
	}

	if err := ses3RPC.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}, &metrics); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(statMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", statMetrics, metrics)
	}
}

func testSes3ItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
