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
package general_tests

import (
	"fmt"
	"net/rpc"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

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
		testV1CDRsRemoveFolders,
		testV1CDRsCreateFolders,
		testV1CDRsInitConfig,
		testV1CDRsInitDataDb,
		testV1CDRsInitCdrDb,
		testV1CDRsStartEngine,
		testV1CDRsRpcConn,
		testV1CDRsLoadTPs,

		testV1CDRsProcessEventExport,
		testV1CDRsProcessEventAttrS,
		testV1CDRsProcessEventChrgS,
		testV1CDRsProcessEventRalS,
		testV1CDRsProcessEventSts,
		testV1CDRsProcessEventStore,
		testV1CDRsProcessEventThreshold,
		testV1CDRsProcessEventExportCheck,

		testV1CDRsKillEngine,
		testV1CDRsRemoveFolders,
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
		t.Error(err)
	}

	if _, err := engine.StopStartEngine(pecdrsCfgPath, *utils.WaitRater); err != nil {
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

func testV1CDRsLoadTPs(t *testing.T) {
	writeFile := func(fileName, data string) error {
		csvFile, err := os.Create(path.Join("/tmp/TestCDRsITPE", fileName))
		if err != nil {
			return err
		}
		defer csvFile.Close()
		_, err = csvFile.WriteString(data)
		if err != nil {
			return err

		}
		return csvFile.Sync()
	}

	// Create and populate AccountActions.csv
	if err := writeFile(utils.AccountActionsCsv, `
#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,PACKAGE_1001,,,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate ActionPlans.csv
	if err := writeFile(utils.ActionPlansCsv, `
#Id,ActionsId,TimingId,Weight
PACKAGE_1001,TOPUP_RST_MONETARY_10,*asap,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Actions.csv
	if err := writeFile(utils.ActionsCsv, `
#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
TOPUP_RST_MONETARY_10,*topup_reset,,,,*monetary,,*any,,,*unlimited,,10,10,false,false,10
TOPUP_MONETARY_10,*topup,,,,*monetary,,*any,,,*unlimited,,10,10,false,false,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Attributes.csv
	if err := writeFile(utils.AttributesCsv, `
#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_SUPPLIER1,*chargers,,,,*req.Subject,*constant,SUPPLIER1,false,10
cgrates.org,ATTR_SubjChange,,,,,*req.Subject,*constant,1011,false,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Chargers.csv
	if err := writeFile(utils.ChargersCsv, `
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,Raw,,,*raw,*constant:*req.RequestType:*none,20
cgrates.org,CustomerCharges,,,CustomerCharges,*none,20
cgrates.org,SupplierCharges,,,SupplierCharges,ATTR_SUPPLIER1,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate DestinationRates.csv
	if err := writeFile(utils.DestinationRatesCsv, `
#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY_1CNT,*any,RT_1CNT,*up,5,0,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Rates.csv
	if err := writeFile(utils.RatesCsv, `
#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_1CNT,0,0.01,60s,1s,0s
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RatingPlans.csv
	if err := writeFile(utils.RatingPlansCsv, `
#Id,DestinationRatesId,TimingTag,Weight
RP_TESTIT1,DR_ANY_1CNT,*any,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RatingProfiles.csv
	if err := writeFile(utils.RatingProfilesCsv, `
#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,*any,2018-01-01T00:00:00Z,RP_TESTIT1,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Stats.csv
	if err := writeFile(utils.StatsCsv, `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,Stat_1,*string:~*req.Account:1001,2014-07-29T15:00:00Z,100,5s,0,*acd;*tcd;*asr,,false,true,30,*none
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Thresholds.csv
	if err := writeFile(utils.ThresholdsCsv, `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,THD_ACNT_1001,*string:~*req.Account:1001,2014-07-29T15:00:00Z,-1,0,0,false,10,TOPUP_MONETARY_10,false
`); err != nil {
		t.Fatal(err)
	}

	var loadInst string
	if err := pecdrsRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: "/tmp/TestCDRsITPE"}, &loadInst); err != nil {
		t.Error(err)
	}
}

func testV1CDRsProcessEventAttrS(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaAttributes, utils.MetaStore, "*chargers:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test1",
			Event: map[string]any{
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
	var reply string
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
	} else if cdrs[0].Subject != "1011" {
		t.Errorf("Expecting: %+v, received: %+v", "1011", cdrs[0].Subject)
	}
}

func testV1CDRsProcessEventChrgS(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaChargers, "*attributes:false", "*export:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test2",
			Event: map[string]any{
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
			Event: map[string]any{
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
			Event: map[string]any{
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
			Event: map[string]any{
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
			{
				Identifier: utils.LOG},
			{
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
		Balance: map[string]any{
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
			Event: map[string]any{
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
			Event: map[string]any{
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
	filesInDir, _ := os.ReadDir(pecdrsCfg.GeneralCfg().FailedPostsDir)
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
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}

func testV1CDRsCreateFolders(t *testing.T) {
	if err := os.MkdirAll("/tmp/TestCDRsITPE", 0755); err != nil {
		t.Error(err)
	}
}

func testV1CDRsRemoveFolders(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestCDRsITPE"); err != nil {
		t.Error(err)
	}
}
