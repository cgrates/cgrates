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
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdrsCfgPath string
	cdrsCfg     *config.CGRConfig
	cdrsRpc     *birpc.Client
	cdrsConfDIR string // run the tests for specific configuration

	// subtests to be executed for each confDIR
	sTestsCDRsIT = []func(t *testing.T){
		testCDRsInitConfig,
		testCDRsFlushDBs,
		testCDRsStartEngine,
		testCDRsRpcConn,
		testCDRsLoadTariffPlanFromFolder,
		//default process
		testCDRsProcessCDR,
		// //custom process
		testCDRsProcessCDR2,
		testCDRsProcessCDR3,

		testCDRsSetStats,
		testCDRsSetThresholdProfile,

		testCDRsProcessCDR4,
		testCDRsGetStats1,
		testCDRsGetThreshold1,
		testCDRsProcessCDR5,
		testCDRsGetStats2,
		testCDRsGetThreshold2,

		testCDRsKillEngine,
	}
)

// Tests starting here
func TestCDRsIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		cdrsConfDIR = "cdrsv2internal"
	case utils.MetaMySQL:
		cdrsConfDIR = "cdrsv2mysql"
	case utils.MetaMongo:
		cdrsConfDIR = "cdrsv2mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func testCDRsInitConfig(t *testing.T) {
	var err error
	cdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", cdrsConfDIR)
	if *utils.Encoding == utils.MetaGOB {
		cdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", cdrsConfDIR+"_gob")
	}
	if cdrsCfg, err = config.NewCGRConfigFromPath(context.Background(), cdrsCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testCDRsFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(cdrsCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testCDRsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCDRsRpcConn(t *testing.T) {
	cdrsRpc = engine.NewRPCClient(t, cdrsCfg.ListenCfg(), *utils.Encoding)
}

func testCDRsLoadTariffPlanFromFolder(t *testing.T) {
	caching := utils.MetaReload
	var rpl string
	if err := cdrsRpc.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaCache:       caching,
				utils.MetaStopOnError: true,
			},
		}, &rpl); err != nil {
		t.Error(err)
	} else if rpl != utils.OK {
		t.Error("Unexpected reply returned:", rpl)
	}
	var resp string
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *utils.ChargerProfile
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testCDRsProcessCDR(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.OriginID:     "testCDRsProcessCDR1",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "testCDRsProcessCDR",
			utils.RequestType:  utils.MetaRated,
			utils.Category:     "call",
			utils.AccountField: "testCDRsProcessCDR",
			utils.Subject:      "ANY2CNT",
			utils.Destination:  "+4986517174963",
			"field_extr1":      "val_extr1",
			"fieldextr2":       "valextr2",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:   "abcdef1",
			utils.MetaAttributes: true,
			utils.MetaChargers:   true,
			utils.MetaRates:      true,
			utils.OptsCDRsExport: false,
			utils.MetaAccounts:   false,
			utils.MetaUsage:      time.Minute + 11*time.Second,
			utils.MetaStartTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
		},
	}

	var reply []*utils.EventsWithOpts
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEventWithGet, args,
		&reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		sort.Slice(reply, func(i, j int) bool {
			return utils.IfaceAsString(reply[i].Opts[utils.MetaRunID]) >
				utils.IfaceAsString(reply[j].Opts[utils.MetaRunID])
		})
		if reply[0].Event["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				reply[0].Event["PayPalAccount"])
		}
		if reply[1].Opts[utils.MetaRateSCost].(map[string]any)[utils.Cost] != 0.4666666666666667 {
			t.Errorf("Unexpected cost for CDR: %f", reply[1].Opts[utils.MetaRateSCost].(map[string]any)[utils.Cost])
		}
		if reply[1].Event["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				reply[1].Event["PayPalAccount"])
		}
	}
}

// Disable Attributes process
func testCDRsProcessCDR2(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.OriginID:     "testCDRsProcessCDR2",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "testCDRsProcessCDR2",
			utils.RequestType:  utils.MetaRated,
			utils.Category:     "call",
			utils.AccountField: "testCDRsProcessCDR2",
			utils.Subject:      "ANY2CNT",
			utils.Destination:  "+4986517174963",
			"field_extr1":      "val_extr1",
			"fieldextr2":       "valextr2",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:   "abcdef2",
			utils.MetaAttributes: false,
			utils.MetaChargers:   true,
			utils.MetaRates:      true,
			utils.OptsCDRsExport: false,
			utils.MetaAccounts:   false,
			utils.MetaUsage:      time.Minute + 11*time.Second,
			utils.MetaStartTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
		},
	}

	var reply []*utils.EventsWithOpts
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEventWithGet, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		sort.Slice(reply, func(i, j int) bool {
			return utils.IfaceAsString(reply[i].Opts[utils.MetaRunID]) >
				utils.IfaceAsString(reply[j].Opts[utils.MetaRunID])
		})
		// we disable the connection to AttributeS and PayPalAccount shouldn't be present
		if _, has := reply[0].Event["PayPalAccount"]; has {
			t.Errorf("PayPalAccount should NOT be added by AttributeS, have: %s",
				reply[0].Event["PayPalAccount"])
		}
		if reply[1].Opts[utils.MetaRateSCost].(map[string]any)[utils.Cost] != 0.4666666666666667 {
			t.Errorf("Unexpected cost for CDR: %f", reply[1].Opts[utils.MetaRateSCost].(map[string]any)[utils.Cost])
		}
		//we disable the connection to AttributeS and PayPalAccount shouldn't be present
		if _, has := reply[1].Event["PayPalAccount"]; has {
			t.Errorf("PayPalAccount should NOT be added by AttributeS, have: %s",
				reply[1].Event["PayPalAccount"])
		}
	}
}

// Disable Attributes and Charger process
func testCDRsProcessCDR3(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.OriginID:     "testCDRsProcessCDR3",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "testCDRsProcessCDR3",
			utils.RequestType:  utils.MetaRated,
			utils.Category:     "call",
			utils.AccountField: "testCDRsProcessCDR3",
			utils.Subject:      "ANY2CNT",
			utils.Destination:  "+4986517174963",
			"field_extr1":      "val_extr1",
			"fieldextr2":       "valextr2",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:   "abcdef3",
			utils.MetaAttributes: false,
			utils.MetaChargers:   false,
			utils.MetaRates:      true,
			utils.OptsCDRsExport: false,
			utils.MetaAccounts:   false,
			utils.MetaUsage:      time.Minute + 11*time.Second,
			utils.MetaStartTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
		},
	}

	var reply []*utils.EventsWithOpts
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEventWithGet, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		sort.Slice(reply, func(i, j int) bool {
			return utils.IfaceAsString(reply[i].Opts[utils.MetaRunID]) >
				utils.IfaceAsString(reply[j].Opts[utils.MetaRunID])
		})
		// we disable the connection to AttributeS and PayPalAccount shouldn't be present
		if _, has := reply[0].Event["PayPalAccount"]; has {
			t.Errorf("PayPalAccount should NOT be added by AttributeS, have: %s",
				reply[0].Event["PayPalAccount"])
		}
	}
}

func testCDRsSetStats(t *testing.T) {
	var reply *engine.StatQueueProfile
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "STS_ProcessCDR"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "STS_ProcessCDR",
			FilterIDs: []string{"*string:~*req.OriginID:testCDRsProcessCDR4"},
			// QueueLength: 10,
			Metrics: []*engine.MetricWithFilters{{
				MetricID: "*sum#~*opts.*usage",
			}},
			ThresholdIDs: []string{utils.MetaNone},
			Blockers:     utils.DynamicBlockers{{Blocker: true}},
			Stored:       true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			MinItems: 0,
		},
	}
	var result string
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "STS_ProcessCDR"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
}

func testCDRsSetThresholdProfile(t *testing.T) {
	// Set Action
	var reply1 string
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "ACT_THD_ProcessCDR",
			Actions: []*utils.APAction{
				{
					Type: utils.MetaLog,
				},
			},
		},
	}
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1SetActionProfile, actPrf, &reply1); err != nil &&
		err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on AdminSv1.SetActionProfile: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Calling AdminSv1.SetActionProfile received: %s", reply1)
	}

	// Set Account
	var reply2 string
	accnt := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "testCDRsProcessCDR4",
		},
	}
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1SetAccount, accnt, &reply2); err != nil {
		t.Error("Got error on AdminSv1.SetAccount: ", err.Error())
	} else if reply2 != utils.OK {
		t.Errorf("Calling AdminSv1.SetAccount received: %s", reply2)
	}

	// Set Threshold
	var reply *engine.ThresholdProfile
	var result string
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ProcessCDR"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_ProcessCDR",
			FilterIDs: []string{"*string:~*req.OriginID:testCDRsProcessCDR4"},
			MaxHits:   -1,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			ActionProfileIDs: []string{"ACT_THD_ProcessCDR"},
			Async:            false,
		},
	}
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ProcessCDR"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
}

func testCDRsProcessCDR4(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.OriginID:     "testCDRsProcessCDR4",
			utils.OriginHost:   "192.168.1.1",
			utils.Source:       "testCDRsProcessCDR4",
			utils.RequestType:  utils.MetaRated,
			utils.Category:     "call",
			utils.AccountField: "testCDRsProcessCDR4",
			utils.Subject:      "ANY2CNT2",
			utils.Destination:  "+4986517174963",
			"field_extr1":      "val_extr1",
			"fieldextr2":       "valextr2",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:   "abcdef4",
			utils.MetaAttributes: true,
			utils.MetaChargers:   true,
			utils.MetaRates:      true,
			utils.OptsCDRsExport: false,
			utils.MetaAccounts:   false,
			utils.MetaStats:      false,
			utils.MetaThresholds: false,
			utils.MetaUsage:      time.Minute + 11*time.Second,
			utils.MetaStartTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
		},
	}

	var reply string
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testCDRsGetStats1(t *testing.T) {
	expectedIDs := []string{"STS_ProcessCDR"}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaOpts + utils.NestingSep + utils.MetaUsage: utils.NotAvailable,
	}
	if err := cdrsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]},
		}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testCDRsGetThreshold1(t *testing.T) {
	expected := []string{"THD_ACNT_1001", "THD_ProcessCDR"}
	var result []string
	if err := cdrsRpc.Call(context.Background(), utils.AdminSv1GetThresholdProfileIDs,
		&utils.ArgsItemIDs{Tenant: "cgrates.org"}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	var td engine.Threshold
	if err := cdrsRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ProcessCDR"},
		}, &td); err != nil {
		t.Error(err)
	} else if td.Hits != 0 {
		t.Errorf("received: %+v", td)
	}
}

func testCDRsProcessCDR5(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.OriginID:     "testCDRsProcessCDR4",
			utils.OriginHost:   "192.168.1.2",
			utils.Source:       "testCDRsProcessCDR5",
			utils.RequestType:  utils.MetaRated,
			utils.Category:     "call",
			utils.AccountField: "testCDRsProcessCDR5",
			utils.Subject:      "ANY2CNT",
			utils.Destination:  "+4986517174963",
			"field_extr1":      "val_extr1",
			"fieldextr2":       "valextr2",
		},
		APIOpts: map[string]any{
			utils.MetaOriginID:   "abcdef5",
			utils.MetaAttributes: true,
			utils.MetaChargers:   true,
			utils.MetaRates:      true,
			utils.OptsCDRsExport: false,
			utils.MetaAccounts:   false,
			utils.MetaStats:      true,
			utils.MetaThresholds: true,
			utils.MetaUsage:      time.Minute + 11*time.Second,
			utils.MetaStartTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
		},
	}

	var reply string
	if err := cdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testCDRsGetStats2(t *testing.T) {
	expectedIDs := []string{"STS_ProcessCDR"}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaOpts + utils.NestingSep + utils.MetaUsage: "142000000000",
	}
	if err := cdrsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]},
		}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testCDRsGetThreshold2(t *testing.T) {
	var td engine.Threshold
	if err := cdrsRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ProcessCDR"},
		}, &td); err != nil {
		t.Error(err)
	} else if td.Hits != 2 { // 2 Chargers
		t.Errorf("received: %+v", td)
	}
}

func testCDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
