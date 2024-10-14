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
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	pecdrsCfgPath string
	pecdrsConfDIR string
	pecdrsCfg     *config.CGRConfig
	pecdrsRpc     *birpc.Client

	sTestsCDRsIT_ProcessEvent = []func(t *testing.T){
		testV1CDRsInitConfig,
		testV1CDRsInitDataDb,
		testV1CDRsInitCdrDb,
		testV1CDRsStartEngine,
		testV1CDRsRpcConn,
		testV1CDRsLoadTariffPlanFromFolder,
		testV1CDRsProcessEventExport,
		testV1CDRsProcessEventAttrS,
		testV1CDRsProcessEventChrgS,
		testV1CDRsProcessEventRalS,
		testV1CDRsProcessEventSts,
		testV1CDRsProcessEventStore,
		testV1CDRsProcessEventThreshold,
		testV1CDRsProcessEventExportCheck,

		testV1CDRsV2ProcessEventRalS,

		testV1CDRsKillEngine,
	}
)

func TestCDRsITPE(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		pecdrsConfDIR = "cdrsv1processevent"
	case utils.MetaMySQL:
		pecdrsConfDIR = "cdrsv1processeventmysql"
	case utils.MetaMongo:
		pecdrsConfDIR = "cdrsv1processeventmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCDRsIT_ProcessEvent {
		t.Run(pecdrsConfDIR, stest)
	}
}

func testV1CDRsInitConfig(t *testing.T) {
	var err error
	pecdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", pecdrsConfDIR)
	if pecdrsCfg, err = config.NewCGRConfigFromPath(pecdrsCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testV1CDRsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(pecdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1CDRsInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(pecdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1CDRsStartEngine(t *testing.T) {
	// before starting the engine, create the directories needed for failed posts or
	// clear their contents if they exist already
	if err := os.RemoveAll(pecdrsCfg.GeneralCfg().FailedPostsDir); err != nil {
		t.Fatal("Error removing folder: ", pecdrsCfg.GeneralCfg().FailedPostsDir, err)
	}
	if err := os.MkdirAll(pecdrsCfg.GeneralCfg().FailedPostsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StopStartEngine(pecdrsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1CDRsRpcConn(t *testing.T) {
	pecdrsRpc = engine.NewRPCClient(t, pecdrsCfg.ListenCfg())
}

func testV1CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst string
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*utils.DataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
}

func testV1CDRsProcessEventAttrS(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "test1_processEvent"}
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      acntAttrs.Tenant,
		Account:     acntAttrs.Account,
		BalanceType: utils.MetaVoice,
		Value:       120000000000,
		Balance: map[string]any{
			utils.ID:     "BALANCE1",
			utils.Weight: 20,
		},
	}
	var reply string
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv1SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: %s", reply)
	}
	expectedVoice := 120000000000.0
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expectedVoice {
		t.Errorf("Expecting: %v, received: %v", expectedVoice, rply)
	}
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaAttributes, utils.MetaStore, "*chargers:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test1",
			Event: map[string]any{
				utils.RunID:        "testv1",
				utils.OriginID:     "test1_processEvent",
				utils.OriginHost:   "OriginHost1",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        2 * time.Minute,
			},
		},
	}
	var cdrs []*engine.CDR
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var result string
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAt *engine.AttributeProfile
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv1GetAttributeProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}}, &replyAt); err != nil {
		t.Fatal(err)
	}
	replyAt.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAt) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(replyAt))
	}
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	// check if the CDR was correctly processed
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost1"}}}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(cdrs))
	} else if !reflect.DeepEqual(argsEv.Event["Account"], cdrs[0].Account) {
		t.Errorf("Expecting: %+v, received: %+v", argsEv.Event["Account"], cdrs[0].Account)
	}
	cdrs[0].AnswerTime = cdrs[0].AnswerTime.UTC()
	if !reflect.DeepEqual(argsEv.Event["AnswerTime"], cdrs[0].AnswerTime) {
		t.Errorf("Expecting: %+v, received: %+v", argsEv.Event["AnswerTime"], cdrs[0].AnswerTime)
	} else if !reflect.DeepEqual(argsEv.Event["Destination"], cdrs[0].Destination) {
		t.Errorf("Expecting: %+v, received: %+v", argsEv.Event["Destination"], cdrs[0].Destination)
	} else if !reflect.DeepEqual(argsEv.Event["OriginID"], cdrs[0].OriginID) {
		t.Errorf("Expecting: %+v, received: %+v", argsEv.Event["OriginID"], cdrs[0].OriginID)
	} else if !reflect.DeepEqual(argsEv.Event["RequestType"], cdrs[0].RequestType) {
		t.Errorf("Expecting: %+v, received: %+v", argsEv.Event["RequestType"], cdrs[0].RequestType)
	} else if !reflect.DeepEqual(argsEv.Event["RunID"], cdrs[0].RunID) {
		t.Errorf("Expecting: %+v, received: %+v", argsEv.Event["RunID"], cdrs[0].RunID)
	} else if !reflect.DeepEqual(argsEv.Event["Usage"], cdrs[0].Usage) {
		t.Errorf("Expecting: %+v, received: %+v", argsEv.Event["Usage"], cdrs[0].Usage)
	} else if !reflect.DeepEqual(argsEv.Tenant, cdrs[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", argsEv.Tenant, cdrs[0].Tenant)
	} else if !reflect.DeepEqual(alsPrf.Attributes[0].Value[0].Rules, cdrs[0].Subject) {
		t.Errorf("Expecting: %+v, received: %+v", alsPrf.Attributes[0].Value[0].Rules, cdrs[0].Subject)
	}
}

func testV1CDRsProcessEventChrgS(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaChargers, "*attributes:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test2",
			Event: map[string]any{
				utils.RunID:        "testv1",
				utils.OriginID:     "test2_processEvent",
				utils.OriginHost:   "OriginHost2",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        2 * time.Minute,
			},
		},
	}
	var reply string
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost2"}}}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 3 {
		t.Errorf("Expecting: 3, received: %+v", len(cdrs))
	} else if cdrs[0].OriginID != "test2_processEvent" || cdrs[1].OriginID != "test2_processEvent" || cdrs[2].OriginID != "test2_processEvent" {
		t.Errorf("Expecting: test2_processEvent, received: %+v, %+v, %+v ", cdrs[0].OriginID, cdrs[1].OriginID, cdrs[2].OriginID)
	}
	sort.Slice(cdrs, func(i, j int) bool { return cdrs[i].RunID < cdrs[j].RunID })
	if cdrs[2].RunID != "raw" { // charger with RunID *raw
		t.Errorf("Expecting: %+v, received: %+v", "raw", cdrs[2].RunID)
	} else if cdrs[0].RunID != "CustomerCharges" {
		t.Errorf("Expecting: %+v, received: %+v", "CustomerCharges", cdrs[0].RunID)
	} else if cdrs[1].RunID != "SupplierCharges" {
		t.Errorf("Expecting: %+v, received: %+v", "SupplierCharges", cdrs[1].RunID)
	}
}

func testV1CDRsProcessEventRalS(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs, "*attributes:false", "*chargers:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test3",
			Event: map[string]any{
				utils.RunID:        "testv1",
				utils.OriginID:     "test3_processEvent",
				utils.OriginHost:   "OriginHost3",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        2 * time.Minute,
			},
		},
	}
	var reply string
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost3"}}}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(cdrs))
	} else if !reflect.DeepEqual(cdrs[0].Cost, 0.0204) {
		t.Errorf("Expected: %+v,\nreceived: %+v", 0.0204, utils.ToJSON(cdrs[0]))
	}
}

func testV1CDRsProcessEventSts(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaStats, "*rals:false", "*attributes:false", "*chargers:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test4",
			Event: map[string]any{
				utils.RunID:        "testv1",
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "test4_processEvent",
				utils.OriginHost:   "OriginHost4",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "+4986517174963",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        5 * time.Minute,
			},
		},
	}
	var reply string
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	eOut := []*engine.CDR{
		{
			CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
			RunID:       "testv1",
			OrderID:     0,
			OriginHost:  "OriginHost4",
			Source:      "",
			OriginID:    "test4_processEvent",
			ToR:         "*voice",
			RequestType: "*pseudoprepaid",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Date(2018, 01, 07, 17, 00, 00, 0, time.UTC),
			AnswerTime:  time.Date(2018, 01, 07, 17, 00, 10, 0, time.UTC),
			Usage:       300000000000,
			ExtraFields: map[string]string{},
			ExtraInfo:   "",
			Partial:     false,
			PreRated:    false,
			CostSource:  "",
			Cost:        -1,
			CostDetails: nil,
		},
	}
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost4"}}}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(cdrs))
	}
	cdrs[0].OrderID = 0
	cdrs[0].SetupTime = cdrs[0].SetupTime.UTC()
	cdrs[0].AnswerTime = cdrs[0].AnswerTime.UTC()
	if !reflect.DeepEqual(eOut[0], cdrs[0]) {
		t.Errorf("Expected: %+v,\nreceived: %+v", utils.ToJSON(eOut[0]), utils.ToJSON(cdrs[0]))
	}
	var metrics map[string]string
	statMetrics := map[string]string{
		utils.MetaACD: "2m30s",
		utils.MetaASR: "100%",
		utils.MetaTCD: "15m0s",
	}

	if err := pecdrsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}}, &metrics); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(statMetrics, metrics) {
		t.Errorf("expecting: %+v, received: %+v", statMetrics, metrics)
	}
}

func testV1CDRsProcessEventStore(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{"*store:false", "*attributes:false", "*chargers:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test5",
			Event: map[string]any{
				utils.RunID:        "testv1",
				utils.OriginID:     "test5_processEvent",
				utils.OriginHost:   "OriginHost5",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        2 * time.Minute,
			},
		},
	}
	var reply string
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost5"}}}, &cdrs); err == nil ||
		err.Error() != "SERVER_ERROR: NOT_FOUND" {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 0 {
		t.Errorf("Expecting: 0, received: %+v", len(cdrs))
	}
}

func testV1CDRsProcessEventThreshold(t *testing.T) {
	var reply string
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv2SetActions, &utils.AttrSetActions{
		ActionsId: "ACT_LOG",
		Actions: []*utils.TPAction{
			{Identifier: utils.MetaLog},
			{
				Identifier: utils.MetaTopUpReset, BalanceType: utils.MetaVoice,
				Units: "10", ExpiryTime: "*unlimited",
				DestinationIds: "*any", BalanceWeight: "10", Weight: 10},
		},
	}, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	tPrfl := engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_Test",
			FilterIDs: []string{
				"*lt:~*req.CostDetails.AccountSummary.BalanceSummaries[0].Value:10",
				"*string:~*req.Account:1005", // only for indexes
			},
			MaxHits:   -1,
			Weight:    30,
			ActionIDs: []string{"ACT_LOG"},
			Async:     true,
		},
	}
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	attrSetAcnt := v2.AttrSetAccount{
		Tenant:  "cgrates.org",
		Account: "1005",
		ExtraOptions: map[string]bool{
			utils.AllowNegative: true,
		},
	}
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv2SetAccount, &attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1005",
		BalanceType: utils.MetaMonetary,
		Value:       1,
		Balance: map[string]any{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 10.0,
		},
	}
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaThresholds, utils.MetaRALs, utils.ConcatenatedKey(utils.MetaChargers, "false"), "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testV2CDRsProcessCDRWithThreshold",
				utils.OriginHost:   "OriginHost6",
				utils.Source:       "testV2CDRsProcessCDRWithThreshold",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.Category:     "call",
				utils.AccountField: "1005",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        100 * time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	var cdrs []*engine.CDR
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost6"}}}, &cdrs); err != nil {
		t.Error("Unexpected error: ", err)
	} else if len(cdrs) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(cdrs))
	}
	var td engine.Threshold
	if err := pecdrsRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}}, &td); err != nil {
		t.Error(err)
	} else if td.Hits != 1 {
		t.Errorf("Expecting threshold to be hit once received: %v", td.Hits)
	}
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1005"}
	time.Sleep(50 * time.Millisecond)
	expectedVoice := 10.0
	if err := pecdrsRpc.Call(context.Background(), utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expectedVoice {
		t.Errorf("Expecting: %v, received: %v", expectedVoice, rply)
	}
}
func testV1CDRsProcessEventExport(t *testing.T) {
	var reply string
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaExport, "*store:false", "*attributes:false", "*chargers:false", "*stats:false", "*thresholds:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test7",
			Event: map[string]any{
				utils.RunID:        "testv1",
				utils.OriginID:     "test7_processEvent",
				utils.OriginHost:   "OriginHost7",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        2 * time.Minute,
			},
		},
	}
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err == nil ||
		err.Error() != utils.ErrPartiallyExecuted.Error() { // the export should fail as we test if the cdr is corectly writen in file
		t.Error("Unexpected error: ", err)
	}
}
func testV1CDRsProcessEventExportCheck(t *testing.T) {
	failoverContent := []byte(fmt.Sprintf(`{"CGRID":"%s"}`, utils.Sha1("test7_processEvent", "OriginHost7")))
	filesInDir, _ := os.ReadDir(pecdrsCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", pecdrsCfg.GeneralCfg().FailedPostsDir)
	}
	var foundFile bool
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName = file.Name()
		if strings.HasPrefix(fileName, "EEs|") {
			foundFile = true
			filePath := path.Join(pecdrsCfg.GeneralCfg().FailedPostsDir, fileName)
			ev, err := ees.NewExportEventsFromFile(filePath)
			if err != nil {
				t.Fatal(err)
			} else if len(ev.Events) == 0 {
				t.Fatal("Expected at least one event")
			}
			if !reflect.DeepEqual(failoverContent, ev.Events[0]) {
				t.Errorf("Expecting: %q, received: %q", string(failoverContent), ev.Events[0])
			}
		}
	}

	if !foundFile {
		t.Fatal("Could not find the file in folder")
	}
}

func testV1CDRsV2ProcessEventRalS(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs, "*attributes:false", "*chargers:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test101",
			Event: map[string]any{
				utils.RunID:        "testv1",
				utils.OriginID:     "test3_v2processEvent",
				utils.OriginHost:   "OriginHost101",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        2 * time.Minute,
			},
		},
	}
	expRply := []*utils.EventWithFlags{
		{
			Flags: []string{},
			Event: map[string]any{
				"Account":     "1001",
				"AnswerTime":  "2019-11-27T12:21:26Z",
				"CGRID":       "d13c705aa38164aaf297fb77d7700565a3cea04b",
				"Category":    "call",
				"Cost":        0.0204,
				"CostDetails": nil,
				"CostSource":  "*cdrs",
				"Destination": "+4986517174963",
				"ExtraInfo":   "",
				"OrderID":     0.,
				"OriginHost":  "OriginHost101",
				"OriginID":    "test3_v2processEvent",
				"Partial":     false,
				"PreRated":    false,
				"RequestType": "*pseudoprepaid",
				"RunID":       "testv1",
				"SetupTime":   "0001-01-01T00:00:00Z",
				"Source":      "",
				"Subject":     "1001",
				"Tenant":      "cgrates.org",
				"ToR":         "*voice",
				"Usage":       120000000000.,
			},
		},
	}
	var reply []*utils.EventWithFlags
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV2ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	}
	reply[0].Event["CostDetails"] = nil
	expRply[0].Event["CGRID"] = reply[0].Event["CGRID"]
	if *utils.Encoding == utils.MetaGOB { // gob encoding encodes 0 values of pointers to nil
		expRply[0].Flags = nil
		if utils.ToJSON(expRply) != utils.ToJSON(reply) {
			t.Errorf("Expected %s, received: %s ", utils.ToJSON(expRply), utils.ToJSON(reply))
		}
	} else {
		if !reflect.DeepEqual(reply[0], expRply[0]) {
			t.Errorf("Expected %s, received: %s ", utils.ToJSON(expRply), utils.ToJSON(reply))
		}
	}
	var cdrs []*engine.CDR
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost101"}}}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(cdrs))
	} else if !reflect.DeepEqual(cdrs[0].Cost, 0.0204) {
		t.Errorf("Expected: %+v,\nreceived: %+v", 0.0204, utils.ToJSON(cdrs[0]))
	}

	argsEv.Flags = append(argsEv.Flags, utils.MetaRerate)
	argsEv.CGREvent.ID = "test1002"
	argsEv.CGREvent.Event[utils.Usage] = time.Minute

	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV2ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	}
	expRply[0].Event["Usage"] = 60000000000.
	expRply[0].Event["Cost"] = 0.0102
	expRply[0].Flags = append(expRply[0].Flags, utils.MetaRefund)
	reply[0].Event["CostDetails"] = nil
	if *utils.Encoding == utils.MetaGOB { // gob encoding encodes 0 values of pointers to nil
		if utils.ToJSON(expRply) != utils.ToJSON(reply) {
			t.Errorf("Expected %s, received: %s ", utils.ToJSON(expRply), utils.ToJSON(reply))
		}
	} else {
		if !reflect.DeepEqual(reply[0], expRply[0]) {
			t.Errorf("Expected %s, received: %s ", utils.ToJSON(expRply), utils.ToJSON(reply))
		}
	}

	argsEv.CGREvent.Event[utils.Usage] = 30 * time.Second
	if err := pecdrsRpc.Call(context.Background(), utils.CDRsV2ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	}
	reply[0].Event["CostDetails"] = nil
	if *utils.Encoding == utils.MetaGOB { // gob encoding encodes 0 values of pointers to nil
		if utils.ToJSON(expRply) != utils.ToJSON(reply) {
			t.Errorf("Expected %s, received: %s ", utils.ToJSON(expRply), utils.ToJSON(reply))
		}
	} else {
		if !reflect.DeepEqual(reply[0], expRply[0]) {
			t.Errorf("Expected %s, received: %s ", utils.ToJSON(expRply), utils.ToJSON(reply))
		}
	}
}

func testV1CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
