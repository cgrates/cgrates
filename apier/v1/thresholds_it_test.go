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
	tSv1CfgPath string
	tSv1Cfg     *config.CGRConfig
	tSv1Rpc     *rpc.Client
	tPrfl       *engine.ThresholdProfileWithAPIOpts
	tSv1ConfDIR string //run tests for specific configuration

	tEvs = []*engine.ThresholdsArgsProcessEvent{
		{
			CGREvent: &utils.CGREvent{ // hitting THD_ACNT_BALANCE_1
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.EventType:     utils.AccountUpdate,
					utils.AccountField:  "1002",
					utils.AllowNegative: true,
					utils.Disabled:      false,
					utils.Units:         12.3},
				Opts: map[string]interface{}{
					utils.MetaEventType: utils.AccountUpdate,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{ // hitting THD_ACNT_BALANCE_1
				Tenant: "cgrates.org",
				ID:     "event2",
				Event: map[string]interface{}{
					utils.EventType:    utils.BalanceUpdate,
					utils.AccountField: "1002",
					utils.BalanceID:    utils.MetaDefault,
					utils.Units:        12.3,
					utils.ExpiryTime:   time.Date(2009, 11, 10, 23, 00, 0, 0, time.UTC),
				},
				Opts: map[string]interface{}{
					utils.MetaEventType: utils.BalanceUpdate,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{ // hitting THD_STATS_1
				Tenant: "cgrates.org",
				ID:     "event3",
				Event: map[string]interface{}{
					utils.EventType:    utils.StatUpdate,
					utils.StatID:       "Stats1",
					utils.AccountField: "1002",
					"ASR":              35.0,
					"ACD":              "2m45s",
					"TCC":              12.7,
					"TCD":              "12m15s",
					"ACC":              0.75,
					"PDD":              "2s",
				},
				Opts: map[string]interface{}{
					utils.MetaEventType: utils.StatUpdate,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{ // hitting THD_STATS_1 and THD_STATS_2
				Tenant: "cgrates.org",
				ID:     "event4",
				Event: map[string]interface{}{
					utils.EventType:    utils.StatUpdate,
					utils.StatID:       "STATS_HOURLY_DE",
					utils.AccountField: "1002",
					"ASR":              35.0,
					"ACD":              "2m45s",
					"TCD":              "1h",
				},
				Opts: map[string]interface{}{
					utils.MetaEventType: utils.StatUpdate,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{ // hitting THD_STATS_3
				Tenant: "cgrates.org",
				ID:     "event5",
				Event: map[string]interface{}{
					utils.EventType:    utils.StatUpdate,
					utils.StatID:       "STATS_DAILY_DE",
					utils.AccountField: "1002",
					"ACD":              "2m45s",
					"TCD":              "3h1s",
				},
				Opts: map[string]interface{}{
					utils.MetaEventType: utils.StatUpdate,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{ // hitting THD_RES_1
				Tenant: "cgrates.org",
				ID:     "event6",
				Event: map[string]interface{}{
					utils.EventType:    utils.ResourceUpdate,
					utils.AccountField: "1002",
					utils.ResourceID:   "RES_GRP_1",
					utils.Usage:        10.0,
				},
				Opts: map[string]interface{}{
					utils.MetaEventType: utils.ResourceUpdate,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{ // hitting THD_RES_1
				Tenant: "cgrates.org",
				ID:     "event6",
				Event: map[string]interface{}{
					utils.EventType:    utils.ResourceUpdate,
					utils.AccountField: "1002",
					utils.ResourceID:   "RES_GRP_1",
					utils.Usage:        10.0,
				},
				Opts: map[string]interface{}{
					utils.MetaEventType: utils.ResourceUpdate,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{ // hitting THD_RES_1
				Tenant: "cgrates.org",
				ID:     "event6",
				Event: map[string]interface{}{
					utils.EventType:    utils.ResourceUpdate,
					utils.AccountField: "1002",
					utils.ResourceID:   "RES_GRP_1",
					utils.Usage:        10.0,
				},
				Opts: map[string]interface{}{
					utils.MetaEventType: utils.ResourceUpdate,
				},
			},
		},
		{
			CGREvent: &utils.CGREvent{ // hitting THD_CDRS_1
				Tenant: "cgrates.org",
				ID:     "cdrev1",
				Event: map[string]interface{}{
					"field_extr1":      "val_extr1",
					"fieldextr2":       "valextr2",
					utils.CGRID:        utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
					utils.RunID:        utils.MetaRaw,
					utils.OrderID:      123,
					utils.OriginHost:   "192.168.1.1",
					utils.Source:       utils.UnitTest,
					utils.OriginID:     "dsafdsaf",
					utils.ToR:          utils.MetaVoice,
					utils.RequestType:  utils.MetaRated,
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.AccountField: "1007",
					utils.Subject:      "1007",
					utils.Destination:  "+4986517174963",
					utils.SetupTime:    time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
					utils.PDD:          0 * time.Second,
					utils.AnswerTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
					utils.Usage:        10 * time.Second,
					utils.Route:        "SUPPL1",
					utils.Cost:         -1.0,
				},
				Opts: map[string]interface{}{
					utils.MetaEventType: utils.CDR,
				},
			},
		},
	}

	sTestsThresholdSV1 = []func(t *testing.T){
		testV1TSLoadConfig,
		testV1TSInitDataDb,
		testV1TSResetStorDb,
		testV1TSStartEngine,
		testV1TSRpcConn,
		testV1TSFromFolder,
		testV1TSGetThresholds,
		testV1TSProcessEvent,
		testV1TSGetThresholdsAfterProcess,
		testV1TSGetThresholdsAfterRestart,
		testv1TSGetThresholdProfileIDs,
		testv1TSGetThresholdProfileIDsCount,
		testV1TSSetThresholdProfileBrokenReference,
		testV1TSSetThresholdProfile,
		testV1TSUpdateThresholdProfile,
		testV1TSRemoveThresholdProfile,
		testV1TSMaxHits,
		testV1TSUpdateSnooze,
		testV1TSGetThresholdProfileWithoutTenant,
		testV1TSRemThresholdProfileWithoutTenant,
		testV1TSProcessEventWithoutTenant,
		testV1TSGetThresholdsWithoutTenant,
		testV1TSProcessAccountUpdateEvent,
		testV1TSResetThresholdsWithoutTenant,
		testV1TSStopEngine,
	}
)

// Test start here
func TestTSV1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		tSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		tSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsThresholdSV1 {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1TSLoadConfig(t *testing.T) {
	var err error
	tSv1CfgPath = path.Join(*dataDir, "conf", "samples", tSv1ConfDIR)
	if tSv1Cfg, err = config.NewCGRConfigFromPath(tSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1TSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1TSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1TSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1TSRpcConn(t *testing.T) {
	var err error
	tSv1Rpc, err = newRPCClient(tSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1TSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := tSv1Rpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1TSGetThresholds(t *testing.T) {
	var tIDs []string
	expectedIDs := []string{"THD_RES_1", "THD_STATS_2", "THD_STATS_1", "THD_ACNT_BALANCE_1", "THD_ACNT_EXPIRED", "THD_STATS_3", "THD_CDRS_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{}, &tIDs); err != nil {
		t.Error(err)
	} else if len(expectedIDs) != len(tIDs) {
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, tIDs)
	}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{Tenant: "cgrates.org"}, &tIDs); err != nil {
		t.Error(err)
	} else if len(expectedIDs) != len(tIDs) {
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, tIDs)
	}
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: expectedIDs[0]}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd, td) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
}

func testV1TSProcessEvent(t *testing.T) {
	var ids []string
	eIDs := []string{}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, &tEvs[0], &ids); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	eIDs = []string{"THD_ACNT_BALANCE_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, &tEvs[1], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_STATS_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, &tEvs[2], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_STATS_2", "THD_STATS_1"}
	eIDs2 := []string{"THD_STATS_1", "THD_STATS_2"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, &tEvs[3], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) && !reflect.DeepEqual(ids, eIDs2) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_STATS_3"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, &tEvs[4], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_RES_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, &tEvs[5], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, &tEvs[6], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, &tEvs[7], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_CDRS_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, &tEvs[8], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
}

func testV1TSGetThresholdsAfterProcess(t *testing.T) {
	var tIDs []string
	expectedIDs := []string{"THD_RES_1", "THD_STATS_2", "THD_STATS_1", "THD_ACNT_BALANCE_1", "THD_ACNT_EXPIRED"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{Tenant: "cgrates.org"}, &tIDs); err != nil {
		t.Error(err)
	} else if len(expectedIDs) != len(tIDs) { // THD_STATS_3 is not reccurent, so it was removed
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, tIDs)
	}
	var td engine.Threshold
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}}, &td); err != nil {
		t.Error(err)
	} else if td.Snooze.IsZero() { // make sure Snooze time was reset during execution
		t.Errorf("received: %+v", td)
	}
}

func testV1TSGetThresholdsAfterRestart(t *testing.T) {
	// in case of internal we skip this test
	if tSv1ConfDIR == "tutinternal" {
		t.SkipNow()
	}
	// time.Sleep(time.Second)
	if _, err := engine.StopStartEngine(tSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	var err error
	tSv1Rpc, err = newRPCClient(tSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	var td engine.Threshold
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}}, &td); err != nil {
		t.Error(err)
	} else if td.Snooze.IsZero() { // make sure Snooze time was reset during execution
		t.Errorf("received: %+v", td)
	}
}

func testV1TSSetThresholdProfileBrokenReference(t *testing.T) {
	var reply *engine.ThresholdProfile
	var result string
	tPrfl = &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_Test",
			FilterIDs: []string{"NonExistingFilter"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1"},
			Async:     true,
		},
	}
	expErr := "SERVER_ERROR: broken reference to filter: NonExistingFilter for item with ID: cgrates.org:THD_Test"
	if err := tSv1Rpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
}

func testv1TSGetThresholdProfileIDs(t *testing.T) {
	expected := []string{"THD_STATS_1", "THD_STATS_2", "THD_STATS_3", "THD_RES_1", "THD_CDRS_1", "THD_ACNT_BALANCE_1", "THD_ACNT_EXPIRED"}
	var result []string
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfileIDs, &utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfileIDs, &utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testV1TSSetThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	var result string
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_Test",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1"},
			Async:     true,
		},
	}
	if err := tSv1Rpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
}

func testV1TSUpdateThresholdProfile(t *testing.T) {
	var result string
	tPrfl.FilterIDs = []string{"*string:~Account:1001", "*prefix:~DST:10"}
	if err := tSv1Rpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.ThresholdProfile
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply.FilterIDs)
		sort.Strings(tPrfl.ThresholdProfile.FilterIDs)
		if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
			t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
		}
	}
}

func testV1TSRemoveThresholdProfile(t *testing.T) {
	var resp string
	if err := tSv1Rpc.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.ThresholdProfile
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Received %s and the error:%+v", utils.ToJSON(sqp), err)
	}
	if err := tSv1Rpc.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}}, &resp); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v received: %v", utils.ErrNotFound, err)
	}
}

func testV1TSMaxHits(t *testing.T) {
	var reply string
	// check if exist
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TH3"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:  "cgrates.org",
			ID:      "TH3",
			MaxHits: 3,
		},
	}
	//set
	if err := tSv1Rpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var ids []string
	eIDs := []string{"TH3"}
	thEvent := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{ // hitting TH3
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
	}
	//process event
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	//check threshold after first process ( hits : 1)
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "TH3", Hits: 1}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TH3"}}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
	//process event
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}

	//check threshold for event
	var ths engine.Thresholds
	eTd.Hits = 2
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThresholdsForEvent,
		thEvent, &ths); err != nil {
		t.Error(err)
	} else if len(ths) != 1 {
		t.Errorf("expecting: 1, received: %+v", utils.ToJSON(ths))
	} else if !reflect.DeepEqual(eTd.TenantID(), ths[0].TenantID()) {
		t.Errorf("expecting: %+v, received: %+v", eTd.TenantID(), ths[0].TenantID())
	} else if !reflect.DeepEqual(eTd.Hits, ths[0].Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd.Hits, ths[0].Hits)
	}

	//check threshold for event without tenant
	thEvent.Tenant = utils.EmptyString
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThresholdsForEvent,
		thEvent, &ths); err != nil {
		t.Error(err)
	} else if len(ths) != 1 {
		t.Errorf("expecting: 1, received: %+v", utils.ToJSON(ths))
	} else if !reflect.DeepEqual(eTd.TenantID(), ths[0].TenantID()) {
		t.Errorf("expecting: %+v, received: %+v", eTd.TenantID(), ths[0].TenantID())
	} else if !reflect.DeepEqual(eTd.Hits, ths[0].Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd.Hits, ths[0].Hits)
	}

	//check threshold after second process ( hits : 2)
	eTd.Hits = 2
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TH3"}}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}

	//process event
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	//check threshold after third process (reached the maximum hits and should be removed)
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TH3"}}, &td); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Err : %+v \n, td : %+v", err, utils.ToJSON(td))
	}
}

func testV1TSUpdateSnooze(t *testing.T) {
	var reply string
	// check if exist
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TH4"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	customTh := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH4",
			FilterIDs: []string{"*string:~*req.CustomEv:SnoozeEv"},
			MinSleep:  10 * time.Minute,
			Weight:    100,
		},
	}
	//set
	if err := tSv1Rpc.Call(utils.APIerSv1SetThresholdProfile, customTh, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var ids []string
	eIDs := []string{"TH4"}
	thEvent := &engine.ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH4"},
		CGREvent: &utils.CGREvent{ // hitting TH4
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				"CustomEv": "SnoozeEv",
			},
		},
	}
	tNow := time.Now()
	//process event
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	//check threshold after first process ( hits : 1)
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "TH4", Hits: 1}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TH4"}}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	} else if !(td.Snooze.After(tNow.Add(9*time.Minute)) && td.Snooze.Before(tNow.Add(11*time.Minute))) { // Snooze time should be between time.Now + 9 min and time.Now + 11 min
		t.Errorf("expecting: %+v, received: %+v", tNow.Add(10*time.Minute), td.Snooze)
	}

	customTh2 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH4",
			FilterIDs: []string{"*string:~*req.CustomEv:SnoozeEv"},
			MinSleep:  5 * time.Minute,
			Weight:    100,
		},
	}
	//set
	if err := tSv1Rpc.Call(utils.APIerSv1SetThresholdProfile, customTh2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TH4"}}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	} else if !(td.Snooze.After(tNow.Add(4*time.Minute)) && td.Snooze.Before(tNow.Add(6*time.Minute))) { // Snooze time should be between time.Now + 9 min and time.Now + 11 min
		t.Errorf("expecting: %+v, received: %+v", tNow.Add(10*time.Minute), td.Snooze)
	}

}

func testV1TSStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func testV1TSGetThresholdProfileWithoutTenant(t *testing.T) {
	tPrfl = &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			ID:        "randomID",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1"},
			Async:     true,
		},
	}
	var reply string
	if err := tSv1Rpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	tPrfl.ThresholdProfile.Tenant = "cgrates.org"
	var result *engine.ThresholdProfile
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: "randomID"},
		&result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, result) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(tPrfl.ThresholdProfile), utils.ToJSON(result))
	}
}

func testV1TSRemThresholdProfileWithoutTenant(t *testing.T) {
	var reply string
	if err := tSv1Rpc.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "randomID"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *engine.ThresholdProfile
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{ID: "randomID"},
		&result); err == nil || utils.ErrNotFound.Error() != err.Error() {
		t.Error(err)
	}
}

func testv1TSGetThresholdProfileIDsCount(t *testing.T) {
	var reply int
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfileIDsCount,
		&utils.TenantWithAPIOpts{},
		&reply); err != nil {
		t.Error(err)
	} else if reply != 7 {
		t.Errorf("Expected 7, received %+v", reply)
	}
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfileIDsCount,
		&utils.TenantWithAPIOpts{Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != 7 {
		t.Errorf("Expected 7, received %+v", reply)
	}
}

func testV1TSProcessEventWithoutTenant(t *testing.T) {
	var ids []string
	eIDs := []string{"TH4"}
	thEvent := &engine.ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"TH4"},
		CGREvent: &utils.CGREvent{ // hitting TH4
			ID: "event1",
			Event: map[string]interface{}{
				"CustomEv": "SnoozeEv",
			},
		},
	}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
}

func testV1TSGetThresholdsWithoutTenant(t *testing.T) {
	expectedThreshold := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_BALANCE_1",
		Hits:   1,
	}
	var reply *engine.Threshold
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "THD_ACNT_BALANCE_1"}},
		&reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedThreshold.ID, reply.ID) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedThreshold.ID), utils.ToJSON(reply.ID))
	} else if !reflect.DeepEqual(expectedThreshold.Tenant, reply.Tenant) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedThreshold.Tenant), utils.ToJSON(reply.Tenant))
	} else if !reflect.DeepEqual(expectedThreshold.ID, reply.ID) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedThreshold.Tenant), utils.ToJSON(reply.Tenant))
	}
}

func testV1TSProcessAccountUpdateEvent(t *testing.T) {
	var result string
	thAcntUpdate := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "TH_ACNT_UPDATE_EV",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.ID:testV1TSProcessAccountUpdateEvent",
				"*gt:~*asm.BalanceSummaries.HolidayBalance.Value:1.0",
			},
			MaxHits:   10,
			MinSleep:  10 * time.Millisecond,
			Weight:    20.0,
			ActionIDs: []string{"LOG_WARNING"},
			Async:     true,
		},
	}
	if err := tSv1Rpc.Call(utils.APIerSv1SetThresholdProfile, thAcntUpdate, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.ThresholdProfile
	if err := tSv1Rpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TH_ACNT_UPDATE_EV"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thAcntUpdate.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", thAcntUpdate.ThresholdProfile, reply)
	}

	attrSetBalance := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testV1TSProcessAccountUpdateEvent",
		BalanceType: "*monetary",
		Value:       1.5,
		Balance: map[string]interface{}{
			utils.ID: "HolidayBalance",
		},
	}
	if err := tSv1Rpc.Call(utils.APIerSv1SetBalance, attrSetBalance, &result); err != nil {
		t.Error("Got error on APIerSv1.SetBalance: ", err.Error())
	} else if result != utils.OK {
		t.Errorf("Calling APIerSv1.SetBalance received: %s", result)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testV1TSProcessAccountUpdateEvent",
	}
	if err := tSv1Rpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	}

	acntUpdateEv := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{ // hitting TH_ACNT_UPDATE_EV
			Tenant: "cgrates.org",
			ID:     "SIMULATE_ACNT_UPDATE_EV",
			Event:  acnt.AsAccountSummary().AsMapInterface(),
			Opts: map[string]interface{}{
				utils.MetaEventType: utils.AccountUpdate,
			},
		},
	}

	var ids []string
	eIDs := []string{"TH_ACNT_UPDATE_EV"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, acntUpdateEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}

}

func testV1TSResetThresholdsWithoutTenant(t *testing.T) {
	expectedThreshold := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_BALANCE_1",
		Hits:   1,
	}
	var reply *engine.Threshold
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "THD_ACNT_BALANCE_1"}},
		&reply); err != nil {
		t.Fatal(err)
	}
	reply.Snooze = expectedThreshold.Snooze
	if !reflect.DeepEqual(expectedThreshold, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedThreshold), utils.ToJSON(reply))
	}
	var result string
	if err := tSv1Rpc.Call(utils.ThresholdSv1ResetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "THD_ACNT_BALANCE_1"}},
		&result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, result)
	}
	expectedThreshold.Hits = 0
	reply = nil
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "THD_ACNT_BALANCE_1"}},
		&reply); err != nil {
		t.Fatal(err)
	}
	reply.Snooze = expectedThreshold.Snooze
	if !reflect.DeepEqual(expectedThreshold, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedThreshold), utils.ToJSON(reply))
	}
}
