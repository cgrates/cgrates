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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dynThdCfgPath string
	dynThdCfg     *config.CGRConfig
	dynThdRpc     *birpc.Client
	dymThdConfDIR string
	dynThdDelay   int

	sTestsDynThd = []func(t *testing.T){
		testDynThdLoadConfig,
		testDynThdInitDataDb,
		testDynThdInitStorDb,
		testDynThdStartEngine,
		testDynThdRpcConn,
		testDynThdLoadTpsForTimingsAndRatingPlans,
		testDynThdCheckForThresholdProfile,
		testDynThdCheckForStatsProfile,
		testDynThdCheckForAttributeProfile,
		testDynThdCheckForActionPlan,
		testDynThdCheckForAction,
		testDynThdCheckForDestination,
		testDynThdCheckForFilter,
		testDynThdCheckForRoute,
		testDynThdCheckForRanking,
		testDynThdCheckForRatingProfile,
		testDynThdCheckForTrend,
		testDynThdCheckForResource,
		testDynThdCheckForActionTrigger,
		testDynThdSetLogAction,
		testDynThdSetAction,
		testDynThdSetThresholdProfile,
		testDynThdGetThresholdBeforeDebit,
		testDynThdSetBalance,
		testDynThdGetAccountBeforeDebit,
		testDynThdDebit1,
		testDynThdGetThresholdBeforeDebit,
		testDynThdDebit2,
		testDynThdGetAccountAfterDebit,
		testDynThdGetThresholdAfterDebit,
		testDynThdCheckForDynCreatedThresholdProfile,
		testDynThdCheckForDynCreatedStatQueueProfile,
		testDynThdCheckForDynCreatedAttributeProfile,
		testDynThdCheckForDynCreatedActionPlan,
		testDynThdCheckForDynCreatedAction,
		testDynThdCheckForDynCreatedDestination,
		testDynThdCheckForDynCreatedFilter,
		testDynThdCheckForDynCreatedRoute,
		testDynThdCheckForDynCreatedRanking,
		testDynThdCheckForDynRatingProfile,
		testDynThdCheckForDynCreatedTrend,
		testDynThdCheckForDynResource,
		testDynThdCheckForDynActionTrigger,
		testDynThdStopEngine,
	}
)

// Test starts here
func TestDynThdIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		dymThdConfDIR = "tutinternal"
	case utils.MetaMySQL:
		dymThdConfDIR = "tutmysql"
	case utils.MetaMongo:
		dymThdConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsDynThd {
		t.Run(dymThdConfDIR, stest)
	}
}

func testDynThdLoadConfig(t *testing.T) {
	var err error
	dynThdCfgPath = path.Join(*utils.DataDir, "conf", "samples", dymThdConfDIR)
	if dynThdCfg, err = config.NewCGRConfigFromPath(dynThdCfgPath); err != nil {
		t.Error(err)
	}
	dynThdDelay = 1000
}

func testDynThdInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(dynThdCfg); err != nil {
		t.Fatal(err)
	}
}

func testDynThdInitStorDb(t *testing.T) {
	if err := engine.InitStorDb(dynThdCfg); err != nil {
		t.Fatal(err)
	}
}

func testDynThdStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dynThdCfgPath, dynThdDelay); err != nil {
		t.Fatal(err)
	}
}

func testDynThdRpcConn(t *testing.T) {
	dynThdRpc = engine.NewRPCClient(t, dynThdCfg.ListenCfg())
}

func testDynThdLoadTpsForTimingsAndRatingPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
}

func testDynThdCheckForThresholdProfile(t *testing.T) {
	var rply *engine.ThresholdProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_THR_1002"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForStatsProfile(t *testing.T) {
	var rply *engine.StatQueueProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetStatQueueProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_STAT_1002"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForAttributeProfile(t *testing.T) {
	var rply *engine.StatQueueProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetAttributeProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_ATTR_1002"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForActionPlan(t *testing.T) {
	var rply *engine.ActionPlan
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetActionPlan, &v1.AttrGetActionPlan{ID: "DYNAMICLY_ACT_PLN_1002"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForAction(t *testing.T) {
	var rply map[string]engine.Actions
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2GetActions, &v2.AttrGetActions{ActionIDs: []string{"DYNAMICLY_ACT_1002"}}, &rply); err == nil || err.Error() != utils.NewErrServerError(utils.ErrNotFound).Error() {
		t.Error(err)
	}
}

func testDynThdCheckForDestination(t *testing.T) {
	var rply *engine.Destination
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetDestination, "DYNAMICLY_DST_1005", &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForFilter(t *testing.T) {
	var rply *engine.Filter
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_FLT_1002"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForRoute(t *testing.T) {
	var rply *engine.RouteProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetRouteProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_ROUTE_1002"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForRanking(t *testing.T) {
	var rply *engine.RankingProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetRankingProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_RANKING_1002"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForRatingProfile(t *testing.T) {
	var rply *engine.RatingProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetRatingProfile, &utils.AttrGetRatingProfile{Tenant: "cgrates.org", Category: "call", Subject: "1002"}, &rply); err != nil {
		t.Error(err)
	} else if len(rply.RatingPlanActivations) != 1 {
		t.Errorf("Expected only 1 RatingPlanActivations, received <%+v>", rply.RatingPlanActivations)
	}
}

func testDynThdCheckForTrend(t *testing.T) {
	var rply *engine.TrendProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetTrendProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_TREND_1002"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForResource(t *testing.T) {
	var rply *engine.ResourceProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetResourceProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_RESOURCE_1002"}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdCheckForActionTrigger(t *testing.T) {
	var rply *engine.ActionTriggers
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetActionTriggers, &v1.AttrGetActionTriggers{GroupIDs: []string{"DYNAMICLY_ACTTRGR_1002"}}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDynThdSetLogAction(t *testing.T) {
	var reply string

	act := &utils.AttrSetActions{
		ActionsId: "LOG_WARNING_1002",
		Actions: []*utils.TPAction{
			{
				Identifier:      utils.MetaLog,
				BalanceBlocker:  "false",
				BalanceDisabled: "false",
				Weight:          10,
			},
		}}
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2SetActions,
		act, &reply); err != nil {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
}

func testDynThdSetAction(t *testing.T) {
	var reply string

	act := &utils.AttrSetActions{
		ActionsId: "DYNAMIC_THRESHOLD_ACTION",
		Actions: []*utils.TPAction{
			{
				Identifier:      utils.MetaDynamicThreshold,
				ExtraParameters: "cgrates.org;DYNAMICLY_THR_<~*req.ID>;*string:~*opts.*eventType:AccountUpdate;;1;;;true;10;;true;;~*opts",
			},
			{
				Identifier:      utils.MetaDynamicStats,
				ExtraParameters: "*tenant;DYNAMICLY_STAT_<~*req.ID>;*string:~*opts.*eventType:AccountUpdate;*now;100;10s;0;*acd&*tcd&*asr;;false;true;30;*none;~*opts",
			},
			{
				Identifier:      utils.MetaDynamicAttribute,
				ExtraParameters: "*tenant;DYNAMICLY_ATTR_<~*req.ID>;*any;*string:~*opts.*eventType:AccountUpdate;*now;;*req.Subject;*constant;SUPPLIER1;true;10;~*opts",
			},
			{
				Identifier:      utils.MetaDynamicActionPlan,
				ExtraParameters: "DYNAMICLY_ACT_PLN_<~*req.ID>;LOG_WARNING_<~*req.ID>;*asap;10;",
			},
			{
				Identifier:      utils.MetaDynamicAction,
				ExtraParameters: "DYNAMICLY_ACT_<~*req.ID>;*cdrlog;\f{\"Account\":\"<~*req.ID>\",\"RequestType\":\"*pseudoprepaid\",\"Subject\":\"DifferentThanAccount\", \"ToR\":\"~ActionType:s/^\\*(.*)$/did_$1/\"}\f;*string:~*req.Account:<~*req.ID>&filter2;balID;*monetary;call&data;1002&1003;SPECIAL_1002;SHARED_A&SHARED_B;*unlimited;*daily;10;10;true;false;10;false",
			},
			{
				Identifier:      utils.MetaDynamicDestination,
				ExtraParameters: "DYNAMICLY_DST_1005;1005",
			},
			{
				Identifier:      utils.MetaDynamicFilter,
				ExtraParameters: "*tenant;DYNAMICLY_FLT_<~*req.ID>;*string;~*req.Account;<~*req.ID>;*now;",
			},
			{
				Identifier:      utils.MetaDynamicRoute,
				ExtraParameters: "*tenant;DYNAMICLY_ROUTE_<~*req.ID>;*string:~*req.Account:<~*req.ID>&*string:~*req.Destination:1003;*now;*weight;*acd&*tcc;route1;*string:~*req.Account:<~*req.ID>&*string:~*req.Destination:1003;<~*req.ID>;RP1&RP2;RS1&RS2;Stat_1&Stat_1_1;10;true;param;10;key:value",
			},
			{
				Identifier:      utils.MetaDynamicRanking,
				ExtraParameters: "*tenant;DYNAMICLY_RANKING_<~*req.ID>;@every 15m;Stats2&Stats3&Stats4;Metric1&Metric3;*asc;metricA:true&metricB:false;true;THD1&THD2;key:value",
			},
			{
				Identifier:      utils.MetaDynamicRatingProfile,
				ExtraParameters: "*tenant;call;~*req.ID;*now;RP_<~*req.ID>;;key:value",
			},
			{
				Identifier:      utils.MetaDynamicTrend,
				ExtraParameters: "*tenant;DYNAMICLY_TREND_<~*req.ID>;0 12 * * *;Stats2;*acc&*tcc;-1;-1;1;*average;2.1;true;TD1&TD2;key:value",
			},
			{
				Identifier:      utils.MetaDynamicResource,
				ExtraParameters: "*tenant;DYNAMICLY_RESOURCE_<~*req.ID>;*string:~*req.Account:<~*req.ID>;;1h;1;msg_<~*req.ID>;true;true;10;TD1&TD2;key:value",
			},
			{
				Identifier:      utils.MetaDynamicActionTrigger,
				ExtraParameters: "DYNAMICLY_ACTTRGR_<~*req.ID>;uid_<~*req.ID>;*max_balance;20;true;1s;2014-07-29T15:00:00Z;2014-07-29T15:00:00Z;*default;*monetary;call&data;DST_<~*req.ID>&DST2;SPECIAL_<~*req.ID>;SHRGroup_<~*req.ID>&SHRGroup2;2014-07-29T15:00:00Z;*asap;10;true;true;ACT_<~*req.ID>;20",
			},
		}}
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2SetActions,
		act, &reply); err != nil {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
}

func testDynThdSetThresholdProfile(t *testing.T) {
	ThdPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate", "*string:~*asm.ID:1002", "*lt:~*asm.BalanceSummaries.testBalanceID.Value:56m", "*gte:~*asm.BalanceSummaries.testBalanceID.Initial:58m"},
			ID:        "THD_ACNT_1002",
			MaxHits:   1,
			ActionIDs: []string{"DYNAMIC_THRESHOLD_ACTION"},
			Async:     true,
		},
	}
	var reply string
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, ThdPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
	}

	var result1 *engine.ThresholdProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile, args, &result1); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result1, ThdPrf.ThresholdProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(ThdPrf.ThresholdProfile), utils.ToJSON(result1))
	}
}

func testDynThdGetThresholdBeforeDebit(t *testing.T) {
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
	}

	expThd := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
		Hits:   0,
	}

	var result2 *engine.Threshold
	if err := dynThdRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, &utils.TenantIDWithAPIOpts{TenantID: args}, &result2); err != nil {
		t.Error(err)
	} else if result2.Snooze = expThd.Snooze; !reflect.DeepEqual(result2, expThd) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expThd), utils.ToJSON(result2))
	}
}
func testDynThdSetBalance(t *testing.T) {
	args := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1002",
		BalanceType: utils.MetaVoice,
		Value:       float64(time.Hour),
		Balance: map[string]any{
			utils.ID:            "testBalanceID",
			utils.RatingSubject: "*zero1s",
		},
	}
	var reply string
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling SetBalance received: %s", reply)
	}
}

func testDynThdGetAccountBeforeDebit(t *testing.T) {
	exp := float64(time.Hour)
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v",
			exp, rply)
	}
}

func testDynThdDebit1(t *testing.T) {
	tStart := time.Date(2021, 5, 5, 12, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptor{
		Category:      utils.Call,
		Tenant:        "cgrates.org",
		Subject:       "1002",
		Account:       "1002",
		Destination:   "1003",
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(5 * time.Second),
		LoopIndex:     0,
		DurationIndex: 5 * time.Second,
		ToR:           utils.MetaVoice,
		CgrID:         "12345678911",
		RunID:         utils.MetaDefault,
	}
	cc := new(engine.CallCost)
	if err := dynThdRpc.Call(context.Background(), utils.ResponderMaxDebit, &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: cd,
	}, cc); err != nil {
		t.Error(err)
	}
}

func testDynThdDebit2(t *testing.T) {
	tStart := time.Date(2021, 5, 5, 12, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptor{
		Category:      utils.Call,
		Tenant:        "cgrates.org",
		Subject:       "1002",
		Account:       "1002",
		Destination:   "1003",
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(5 * time.Minute),
		LoopIndex:     0,
		DurationIndex: 5 * time.Minute,
		ToR:           utils.MetaVoice,
		CgrID:         "12345678910",
		RunID:         utils.MetaDefault,
	}
	cc := new(engine.CallCost)
	if err := dynThdRpc.Call(context.Background(), utils.ResponderMaxDebit, &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: cd,
	}, cc); err != nil {
		t.Error(err)
	}
}

func testDynThdGetAccountAfterDebit(t *testing.T) {
	exp := float64(time.Hour - 5*time.Minute - 5*time.Second)
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v",
			exp, rply)
	}
}

func testDynThdGetThresholdAfterDebit(t *testing.T) {
	expThd := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
		Hits:   1,
	}
	var result2 *engine.Threshold
	if err := dynThdRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1002"}}, &result2); err != nil {
		t.Error(err)
	} else if result2.Snooze = expThd.Snooze; !reflect.DeepEqual(result2, expThd) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expThd), utils.ToJSON(result2))
	}
}

func testDynThdCheckForDynCreatedThresholdProfile(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	exp := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "DYNAMICLY_THR_1002",
		FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Time{},
			ExpiryTime:     time.Time{},
		},
		MaxHits: 1,
		Blocker: true,
		Weight:  10,
		Async:   true,
	}
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "DYNAMICLY_THR_1002",
	}
	var result1 *engine.ThresholdProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile, args, &result1); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result1, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynCreatedStatQueueProfile(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "DYNAMICLY_STAT_1002",
	}
	var result1 *engine.StatQueueProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetStatQueueProfile, args, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "DYNAMICLY_STAT_1002",
		FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: result1.ActivationInterval.ActivationTime,
		},
		QueueLength: 100,
		TTL:         10000000000,
		MinItems:    0,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: "*acd",
			},
			{
				MetricID: "*tcd",
			},
			{
				MetricID: "*asr",
			},
		},
		Stored:       false,
		Blocker:      true,
		Weight:       30,
		ThresholdIDs: []string{utils.MetaNone},
	}
	if !reflect.DeepEqual(result1, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynCreatedAttributeProfile(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "DYNAMICLY_ATTR_1002",
	}
	var result1 *engine.AttributeProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetAttributeProfile, args, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "DYNAMICLY_ATTR_1002",
		FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: result1.ActivationInterval.ActivationTime,
		},
		Contexts: []string{utils.MetaAny},
		Attributes: []*engine.Attribute{
			{
				Path:  "*req.Subject",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("SUPPLIER1", "&"),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	if !reflect.DeepEqual(utils.ToJSON(result1), utils.ToJSON(exp)) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynCreatedActionPlan(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	args := &v1.AttrGetActionPlan{ID: "DYNAMICLY_ACT_PLN_1002"}
	var result1 []*engine.ActionPlan
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetActionPlan, args, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &[]*engine.ActionPlan{
		{
			Id:         "DYNAMICLY_ACT_PLN_1002",
			AccountIDs: nil,
			ActionTimings: []*engine.ActionTiming{
				{
					Uuid: result1[0].ActionTimings[0].Uuid,
					Timing: &engine.RateInterval{
						Timing: &engine.RITiming{
							ID:        utils.MetaASAP,
							Years:     utils.Years{},
							Months:    utils.Months{},
							MonthDays: utils.MonthDays{},
							WeekDays:  utils.WeekDays{},
							StartTime: utils.MetaASAP,
							EndTime:   utils.EmptyString,
						},
						Rating: nil,
						Weight: 0,
					},
					ActionsID: "LOG_WARNING_1002",
					ExtraData: nil,
					Weight:    10,
				},
			},
		},
	}
	if !reflect.DeepEqual(utils.ToJSON(exp), utils.ToJSON(result1)) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynCreatedAction(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	args := &v2.AttrGetActions{ActionIDs: []string{"DYNAMICLY_ACT_1002"}}
	result1 := make(map[string]engine.Actions)
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv2GetActions, args, &result1); err != nil {
		t.Fatal(err)
	}
	exp := map[string]engine.Actions{
		"DYNAMICLY_ACT_1002": {
			{
				Id:               "DYNAMICLY_ACT_1002",
				ActionType:       utils.CDRLog,
				ExtraParameters:  "{\"Account\":\"1002\",\"RequestType\":\"*pseudoprepaid\",\"Subject\":\"DifferentThanAccount\", \"ToR\":\"~ActionType:s/^\\*(.*)$/did_$1/\"}",
				Filters:          []string{"*string:~*req.Account:1002", "filter2"},
				ExpirationString: utils.MetaUnlimited,
				Weight:           10,
				Balance: &engine.BalanceFilter{
					Uuid: result1["DYNAMICLY_ACT_1002"][0].Balance.Uuid,
					ID:   utils.StringPointer("balID"),
					Type: utils.StringPointer(utils.MetaMonetary),
					Value: &utils.ValueFormula{
						Method: utils.EmptyString,
						Params: nil,
						Static: 10,
					},
					ExpirationDate: nil,
					Weight:         utils.Float64Pointer(10),
					DestinationIDs: &utils.StringMap{"1002": true, "1003": true},
					RatingSubject:  utils.StringPointer("SPECIAL_1002"),
					Categories:     &utils.StringMap{"call": true, "data": true},
					SharedGroups:   &utils.StringMap{"SHARED_A": true, "SHARED_B": true},
					TimingIDs:      &utils.StringMap{utils.MetaDaily: true},
					Timings: []*engine.RITiming{{
						ID:        utils.MetaDaily,
						Years:     utils.Years{},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: result1["DYNAMICLY_ACT_1002"][0].Balance.Timings[0].StartTime, //depends on time it ran
					}},
					Disabled: utils.BoolPointer(false),
					Factors:  nil,
					Blocker:  utils.BoolPointer(true),
				},
			},
		},
	}
	if !reflect.DeepEqual(exp, result1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynCreatedDestination(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	var result1 *engine.Destination
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetDestination, "DYNAMICLY_DST_1005", &result1); err != nil {
		t.Error(err)
	}
	exp := &engine.Destination{
		Id:       "DYNAMICLY_DST_1005",
		Prefixes: []string{"1005"},
	}
	if !reflect.DeepEqual(exp, result1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynCreatedFilter(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	var result1 *engine.Filter
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_FLT_1002"}, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "DYNAMICLY_FLT_1002",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1002"},
			},
		},
		ActivationInterval: result1.ActivationInterval,
	}
	if !reflect.DeepEqual(exp, result1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynCreatedRoute(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	var result1 *engine.RouteProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetRouteProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_ROUTE_1002"}, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &engine.RouteProfile{
		Tenant: "cgrates.org",
		ID:     "DYNAMICLY_ROUTE_1002",
		FilterIDs: []string{
			"*string:~*req.Account:1002",
			"*string:~*req.Destination:1003",
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: result1.ActivationInterval.ActivationTime,
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{utils.MetaACD, utils.MetaTCC},
		Routes: []*engine.Route{
			{
				ID: "route1",
				FilterIDs: []string{
					"*string:~*req.Account:1002",
					"*string:~*req.Destination:1003",
				},
				AccountIDs: []string{"1002"},
				RatingPlanIDs: []string{
					"RP1",
					"RP2",
				},
				ResourceIDs: []string{
					"RS1",
					"RS2",
				},
				StatIDs: []string{
					"Stat_1",
					"Stat_1_1",
				},
				Weight:          10,
				Blocker:         true,
				RouteParameters: "param",
			},
		},
		Weight: 10,
	}

	if !reflect.DeepEqual(exp, result1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynCreatedRanking(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	var result1 *engine.RankingProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetRankingProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_RANKING_1002"}, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &engine.RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "DYNAMICLY_RANKING_1002",
		Schedule:          "@every 15m",
		StatIDs:           []string{"Stats2", "Stats3", "Stats4"},
		MetricIDs:         []string{"Metric1", "Metric3"},
		Sorting:           "*asc",
		SortingParameters: []string{"metricA:true", "metricB:false"},
		Stored:            true,
		ThresholdIDs:      []string{"THD1", "THD2"},
	}

	if !reflect.DeepEqual(exp, result1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynRatingProfile(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	var result1 *engine.RatingProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetRatingProfile, &utils.AttrGetRatingProfile{Tenant: "cgrates.org", Category: "call", Subject: "1002"}, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &engine.RatingProfile{
		Id: "*out:cgrates.org:call:1002",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
				RatingPlanId:   "RP_1002",
				FallbackKeys:   nil,
			},
			{
				ActivationTime: result1.RatingPlanActivations[1].ActivationTime,
				RatingPlanId:   "RP_1002",
				FallbackKeys:   nil,
			},
		},
	}

	if !reflect.DeepEqual(exp, result1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynCreatedTrend(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	var result1 *engine.TrendProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetTrendProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_TREND_1002"}, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &engine.TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "DYNAMICLY_TREND_1002",
		Schedule:        "0 12 * * *",
		StatID:          "Stats2",
		Metrics:         []string{"*acc", "*tcc"},
		TTL:             -1,
		QueueLength:     -1,
		MinItems:        1,
		CorrelationType: "*average",
		Tolerance:       2.1,
		Stored:          true,
		ThresholdIDs:    []string{"TD1", "TD2"},
	}
	if !reflect.DeepEqual(exp, result1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynResource(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	var result1 *engine.ResourceProfile
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetResourceProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "DYNAMICLY_RESOURCE_1002"}, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &engine.ResourceProfile{
		Tenant: "cgrates.org",
		ID:     "DYNAMICLY_RESOURCE_1002",
		FilterIDs: []string{
			"*string:~*req.Account:1002",
		},
		ActivationInterval: &utils.ActivationInterval{},
		UsageTTL:           time.Hour,
		Limit:              1,
		AllocationMessage:  "msg_1002",
		Blocker:            true,
		Stored:             true,
		Weight:             10,
		ThresholdIDs:       []string{"TD1", "TD2"},
	}

	if !reflect.DeepEqual(exp, result1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdCheckForDynActionTrigger(t *testing.T) {
	time.Sleep(50 * time.Millisecond)
	var result1 *engine.ActionTriggers
	if err := dynThdRpc.Call(context.Background(), utils.APIerSv1GetActionTriggers, &v1.AttrGetActionTriggers{GroupIDs: []string{"DYNAMICLY_ACTTRGR_1002"}}, &result1); err != nil {
		t.Fatal(err)
	}
	exp := &engine.ActionTriggers{
		{
			ID:             "DYNAMICLY_ACTTRGR_1002",
			UniqueID:       "uid_1002",
			ThresholdType:  "*max_balance",
			ThresholdValue: 20,
			Recurrent:      true,
			MinSleep:       time.Second,
			ExpirationDate: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			ActivationDate: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			Balance: &engine.BalanceFilter{
				Uuid:           nil,
				ID:             utils.StringPointer(utils.MetaDefault),
				Type:           utils.StringPointer(utils.MetaMonetary),
				Value:          nil,
				ExpirationDate: utils.TimePointer(time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC)),
				Weight:         utils.Float64Pointer(10),
				DestinationIDs: &utils.StringMap{"DST2": true, "DST_1002": true},
				RatingSubject:  utils.StringPointer("SPECIAL_1002"),
				Categories:     &utils.StringMap{"call": true, "data": true},
				SharedGroups:   &utils.StringMap{"SHRGroup2": true, "SHRGroup_1002": true},
				TimingIDs:      &utils.StringMap{utils.MetaASAP: true},
				Timings:        nil,
				Disabled:       utils.BoolPointer(true),
				Factors:        nil,
				Blocker:        utils.BoolPointer(true),
			},
			Weight:    20,
			ActionsID: "ACT_1002",
		},
	}

	if !reflect.DeepEqual(exp, result1) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(result1))
	}
}

func testDynThdStopEngine(t *testing.T) {
	if err := engine.KillEngine(dynThdDelay); err != nil {
		t.Error(err)
	}
}
