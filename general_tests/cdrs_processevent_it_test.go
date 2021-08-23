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
	"io/ioutil"
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	pecdrsCfgPath string
	pecdrsConfDIR string
	pecdrsCfg     *config.CGRConfig
	pecdrsRpc     *rpc.Client

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
		testV1CDRsKillEngine,
	}
)

func TestCDRsITPE(t *testing.T) {
	switch *dbType {
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
	pecdrsCfgPath = path.Join(*dataDir, "conf", "samples", pecdrsConfDIR)
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
	if _, err := engine.StopStartEngine(pecdrsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1CDRsRpcConn(t *testing.T) {
	var err error
	pecdrsRpc, err = newRPCClient(pecdrsCfg.ListenCfg())
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst string
	if err := pecdrsRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func testV1CDRsProcessEventAttrS(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "test1_processEvent"}
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      acntAttrs.Tenant,
		Account:     acntAttrs.Account,
		BalanceType: utils.VOICE,
		Value:       120000000000,
		Balance: map[string]interface{}{
			utils.ID:     "BALANCE1",
			utils.Weight: 20,
		},
	}
	var reply string
	if err := pecdrsRpc.Call(utils.APIerSv1SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: %s", reply)
	}
	expectedVoice := 120000000000.0
	if err := pecdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.VOICE].GetTotalValue(); rply != expectedVoice {
		t.Errorf("Expecting: %v, received: %v", expectedVoice, rply)
	}
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaAttributes, utils.MetaStore, "*chargers:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test1",
			Event: map[string]interface{}{
				utils.RunID:       "testv1",
				utils.OriginID:    "test1_processEvent",
				utils.OriginHost:  "OriginHost1",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "1001",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:       2 * time.Minute,
			},
		},
	}
	var cdrs []*engine.CDR
	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			Contexts:  []string{utils.META_ANY},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.NewRSRParsersMustCompile("1011", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var result string
	if err := pecdrsRpc.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAt *engine.AttributeProfile
	if err := pecdrsRpc.Call(utils.APIerSv1GetAttributeProfile, &utils.TenantIDWithArgDispatcher{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}}, &replyAt); err != nil {
		t.Fatal(err)
	}
	replyAt.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAt) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(replyAt))
	}
	if err := pecdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	// check if the CDR was correctly processed
	if err := pecdrsRpc.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
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
	return
}

func testV1CDRsProcessEventChrgS(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaChargers, "*attributes:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test2",
			Event: map[string]interface{}{
				utils.RunID:       "testv1",
				utils.OriginID:    "test2_processEvent",
				utils.OriginHost:  "OriginHost2",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "1001",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:       2 * time.Minute,
			},
		},
	}
	var reply string
	if err := pecdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	if err := pecdrsRpc.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost2"}}}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 3 {
		t.Errorf("Expecting: 3, received: %+v", len(cdrs))
	} else if cdrs[0].OriginID != "test2_processEvent" || cdrs[1].OriginID != "test2_processEvent" || cdrs[2].OriginID != "test2_processEvent" {
		t.Errorf("Expecting: test2_processEvent, received: %+v, %+v, %+v ", cdrs[0].OriginID, cdrs[1].OriginID, cdrs[2].OriginID)
	}
	sort.Slice(cdrs, func(i, j int) bool { return cdrs[i].RunID < cdrs[j].RunID })
	if cdrs[0].RunID != utils.MetaRaw { // charger with RunID *raw
		t.Errorf("Expecting: %+v, received: %+v", utils.MetaRaw, cdrs[0].RunID)
	} else if cdrs[1].RunID != "CustomerCharges" {
		t.Errorf("Expecting: %+v, received: %+v", "CustomerCharges", cdrs[1].RunID)
	} else if cdrs[2].RunID != "SupplierCharges" {
		t.Errorf("Expecting: %+v, received: %+v", "SupplierCharges", cdrs[2].RunID)
	}
}

func testV1CDRsProcessEventRalS(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs, "*attributes:false", "*chargers:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test3",
			Event: map[string]interface{}{
				utils.RunID:       "testv1",
				utils.OriginID:    "test3_processEvent",
				utils.OriginHost:  "OriginHost3",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "1001",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:       2 * time.Minute,
			},
		},
	}
	var reply string
	if err := pecdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	if err := pecdrsRpc.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost3"}}}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(cdrs))
	} else if !reflect.DeepEqual(cdrs[0].Cost, 0.0204) {
		t.Errorf("\nExpected: %+v,\nreceived: %+v", 0.0204, utils.ToJSON(cdrs[0]))
	}
}

func testV1CDRsProcessEventSts(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaStatS, "*rals:false", "*attributes:false", "*chargers:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test4",
			Event: map[string]interface{}{
				utils.RunID:       "testv1",
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "test4_processEvent",
				utils.OriginHost:  "OriginHost4",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "1001",
				utils.Destination: "+4986517174963",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       5 * time.Minute,
			},
		},
	}
	var reply string
	if err := pecdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	eOut := []*engine.CDR{
		&engine.CDR{
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
	if err := pecdrsRpc.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost4"}}}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(cdrs))
	}
	cdrs[0].OrderID = 0
	cdrs[0].SetupTime = cdrs[0].SetupTime.UTC()
	cdrs[0].AnswerTime = cdrs[0].AnswerTime.UTC()
	if !reflect.DeepEqual(eOut[0], cdrs[0]) {
		t.Errorf("\nExpected: %+v,\nreceived: %+v", utils.ToJSON(eOut[0]), utils.ToJSON(cdrs[0]))
	}
	var metrics map[string]string
	statMetrics := map[string]string{
		utils.MetaACD: "2m30s",
		utils.MetaASR: "100%",
		utils.MetaTCD: "15m0s",
	}

	if err := pecdrsRpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}}, &metrics); err != nil {
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
			Event: map[string]interface{}{
				utils.RunID:       "testv1",
				utils.OriginID:    "test5_processEvent",
				utils.OriginHost:  "OriginHost5",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "1001",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:       2 * time.Minute,
			},
		},
	}
	var reply string
	if err := pecdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	if err := pecdrsRpc.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost5"}}}, &cdrs); err == nil ||
		err.Error() != "SERVER_ERROR: NOT_FOUND" {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 0 {
		t.Errorf("Expecting: 0, received: %+v", len(cdrs))
	}
}

func testV1CDRsProcessEventThreshold(t *testing.T) {
	var reply string
	if err := pecdrsRpc.Call(utils.APIerSv2SetActions, &utils.AttrSetActions{
		ActionsId: "ACT_LOG",
		Actions: []*utils.TPAction{
			&utils.TPAction{
				Identifier: utils.LOG},
			&utils.TPAction{
				Identifier: utils.TOPUP_RESET, BalanceType: utils.VOICE,
				Units: "10", ExpiryTime: "*unlimited",
				DestinationIds: "*any", BalanceWeight: "10", Weight: 10},
		},
	}, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	tPrfl := engine.ThresholdWithCache{
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
	if err := pecdrsRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &reply); err != nil {
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
	if err := pecdrsRpc.Call(utils.APIerSv2SetAccount, attrSetAcnt, &reply); err != nil {
		t.Fatal(err)
	}
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1005",
		BalanceType: utils.MONETARY,
		Value:       1,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 10.0,
		},
	}
	if err := pecdrsRpc.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaThresholds, utils.MetaRALs, utils.ConcatenatedKey(utils.MetaChargers, "false"), "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDRsProcessCDRWithThreshold",
				utils.OriginHost:  "OriginHost6",
				utils.Source:      "testV2CDRsProcessCDRWithThreshold",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Category:    "call",
				utils.Account:     "1005",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       100 * time.Minute,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
	}
	if err := pecdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	var cdrs []*engine.CDR
	if err := pecdrsRpc.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{OriginHosts: []string{"OriginHost6"}}}, &cdrs); err != nil {
		t.Error("Unexpected error: ", err)
	} else if len(cdrs) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(cdrs))
	}
	var td engine.Threshold
	if err := pecdrsRpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}}, &td); err != nil {
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
	if err := pecdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.VOICE].GetTotalValue(); rply != expectedVoice {
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
			Event: map[string]interface{}{
				utils.RunID:       "testv1",
				utils.OriginID:    "test7_processEvent",
				utils.OriginHost:  "OriginHost7",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "1001",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:       2 * time.Minute,
			},
		},
	}
	if err := pecdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}
func testV1CDRsProcessEventExportCheck(t *testing.T) {
	failoverContent := []byte(fmt.Sprintf(`{"CGRID":"%s"}`, utils.Sha1("test7_processEvent", "OriginHost7")))
	filesInDir, _ := ioutil.ReadDir(pecdrsCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", pecdrsCfg.GeneralCfg().FailedPostsDir)
	}
	var foundFile bool
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName = file.Name()
		if strings.HasPrefix(fileName, "cdr|") {
			foundFile = true
			filePath := path.Join(pecdrsCfg.GeneralCfg().FailedPostsDir, fileName)
			ev, err := engine.NewExportEventsFromFile(filePath)
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
func testV1CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
