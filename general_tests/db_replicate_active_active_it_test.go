//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/utils"
)

// this tests db replication where both engines are active and scheduler runs actions from only 1 host/engine
func TestDBReplicationActiveActive(t *testing.T) {

	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	failedDirNG1 := t.TempDir()

	primaryCfg := fmt.Sprintf(`{
"general": {
	"log_level": 7,
	"node_id": "ng1",
},
"listen": {
	"rpc_json": ":4112",
	"rpc_gob": ":4113",
	"http": ":4180",
	"birpc_json": ""
},
"data_db": {
	"db_type": "*internal",
	"replication_conns": ["rpl"],
	"replication_failed_dir": %q,
	"items": {
		"*accounts": {"replicate":true},
		"*reverse_destinations": {"replicate":true},
		"*destinations": {"replicate":true},
		"*rating_plans": {"replicate":true},
		"*rating_profiles": {"replicate":true},
		"*actions": {"replicate":true},
		"*action_plans": {"replicate":true},
		"*account_action_plans": {"replicate":true},
		"*action_triggers": {"replicate":true},
		"*shared_groups": {"replicate":true},
		"*timings": {"replicate":true},
		"*resource_profiles": {"replicate":true},
		"*resources": {"replicate":true},
		"*ip_profiles": {"replicate":true},
		"*ip_allocations": {"replicate":true},
		"*ranking_profiles": {"replicate":true},
		"*rankings": {"replicate":true},
		"*trend_profiles": {"replicate":true},
		"*trends": {"replicate":true},
		"*statqueue_profiles": {"replicate":true},
		"*statqueues": {"replicate":true},
		"*threshold_profiles": {"replicate":true},
		"*thresholds": {"replicate":true},
		"*filters": {"replicate":true},
		"*route_profiles": {"replicate":true},
		"*attribute_profiles": {"replicate":true},
		"*charger_profiles": {"replicate":true},
		"*dispatcher_profiles": {"replicate":true},
		"*dispatcher_hosts": {"replicate":true},
		"*load_ids": {"replicate":true},
		"*versions": {"replicate":true},
		"*resource_filter_indexes" : {"replicate":true},
		"*ip_filter_indexes" : {"replicate":true},
		"*stat_filter_indexes" : {"replicate":true},
		"*threshold_filter_indexes" : {"replicate":true},
		"*route_filter_indexes" : {"replicate":true},
		"*attribute_filter_indexes" : {"replicate":true},
		"*charger_filter_indexes" : {"replicate":true},
		"*dispatcher_filter_indexes" : {"replicate":true},
		"*reverse_filter_indexes" : {"replicate":true},
		"*sessions_backup": {"replicate":true}, 
	}
},
"stor_db": {
	"db_type": "*internal"
},
"rpc_conns": {
	"rpl": {
		"conns": [
			{
				"address": "127.0.0.1:4123",
				"transport": "*gob"
			}
		]
	}
},
"schedulers": {
	"enabled": true
},

"apiers": {
	"enabled": true,
	"scheduler_conns": ["*localhost"]
}
}`, failedDirNG1)

	failedDirNG2 := t.TempDir()
	targetCfg := fmt.Sprintf(`{
	"general": {
		"log_level": 7,
		"node_id": "ng2",
	},
	"listen": {
		"rpc_json": ":4122",
		"rpc_gob": ":4123",
		"http": ":4190",
		"birpc_json": ""
	},
	"rpc_conns": {
		"rpl": {
			"conns": [
				{
					"address": "127.0.0.1:4113",
					"transport": "*gob"
				}
			]
		}
	},
	"data_db": {
		"db_type": "*internal",
		"replication_conns": ["rpl"],
		"replication_failed_dir": %q,
		"items": {
			"*accounts": {"replicate":true},
			"*reverse_destinations": {"replicate":true},
			"*destinations": {"replicate":true},
			"*rating_plans": {"replicate":true},
			"*rating_profiles": {"replicate":true},
			"*actions": {"replicate":true},
			"*action_plans": {"replicate":true},
			"*account_action_plans": {"replicate":true},
			"*action_triggers": {"replicate":true},
			"*shared_groups": {"replicate":true},
			"*timings": {"replicate":true},
			"*resource_profiles": {"replicate":true},
			"*resources": {"replicate":true},
			"*ip_profiles": {"replicate":true},
			"*ip_allocations": {"replicate":true},
			"*ranking_profiles": {"replicate":true},
			"*rankings": {"replicate":true},
			"*trend_profiles": {"replicate":true},
			"*trends": {"replicate":true},
			"*statqueue_profiles": {"replicate":true},
			"*statqueues": {"replicate":true},
			"*threshold_profiles": {"replicate":true},
			"*thresholds": {"replicate":true},
			"*filters": {"replicate":true},
			"*route_profiles": {"replicate":true},
			"*attribute_profiles": {"replicate":true},
			"*charger_profiles": {"replicate":true},
			"*dispatcher_profiles": {"replicate":true},
			"*dispatcher_hosts": {"replicate":true},
			"*load_ids": {"replicate":true},
			"*versions": {"replicate":true},
			"*resource_filter_indexes" : {"replicate":true},
			"*ip_filter_indexes" : {"replicate":true},
			"*stat_filter_indexes" : {"replicate":true},
			"*threshold_filter_indexes" : {"replicate":true},
			"*route_filter_indexes" : {"replicate":true},
			"*attribute_filter_indexes" : {"replicate":true},
			"*charger_filter_indexes" : {"replicate":true},
			"*dispatcher_filter_indexes" : {"replicate":true},
			"*reverse_filter_indexes" : {"replicate":true},
			"*sessions_backup": {"replicate":true}, 
		}
	},
	"stor_db": {
		"db_type": "*internal"
	},
	"schedulers": {
		"enabled": true
	},

	"apiers": {
		"enabled": true,
		"scheduler_conns": ["*localhost"]
	}
}`, failedDirNG2)
	ng1 := engine.TestEngine{
		ConfigJSON: primaryCfg,
	}

	ng2 := engine.TestEngine{
		ConfigJSON: targetCfg,
	}

	client1, _ := ng1.Run(t)
	time.Sleep(200 * time.Millisecond)

	client2, _ := ng2.Run(t)
	time.Sleep(200 * time.Millisecond)

	var acc1 *engine.Account
	var acc2 *engine.Account
	t.Run("Engine1SetData", func(t *testing.T) {
		var reply string
		topupAction := &utils.AttrSetActions{ActionsId: "TOPUP_ACTION", Actions: []*utils.TPAction{
			{Identifier: utils.MetaTopUpReset, Filters: "*string:~*cfg.general.node_id:ng1", BalanceId: "BalID1", ExpiryTime: "+10s",
				BalanceType: utils.MetaMonetary, Units: "5", Weight: 10.0},
		}}

		if err := client1.Call(context.Background(), utils.APIerSv2SetActions, topupAction, &reply); err != nil {
			t.Fatal(err)
		}
		atms1 := &engine.AttrSetActionPlan{
			Id:              "APRecurring",
			ReloadScheduler: true,
			ActionPlan: []*engine.AttrActionPlan{
				{
					ActionsId: "TOPUP_ACTION",
					Time:      "+1s",
					Weight:    20.0,
					TimingID:  "+1s",
				},
				{
					ActionsId: "TOPUP_ACTION",
					Time:      utils.MetaASAP,
					Weight:    21.0,
					TimingID:  utils.MetaASAP,
				},
			},
		}
		if err := client1.Call(context.Background(), utils.APIerSv1SetActionPlan, &atms1, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client1.Call(context.Background(), utils.APIerSv2SetAccount,
			engine.AttrSetAccount{
				Tenant:          "cgrates.org",
				Account:         "1001",
				ActionPlanIDs:   []string{"APRecurring"},
				ReloadScheduler: true,
			},
			&reply); err != nil {
			t.Fatal(err)
		}
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		if err := client1.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acc1); err != nil {
			t.Error(err)
		}
		t.Run("GetSchedulers", func(t *testing.T) {
			// engine 2 shouldnt have schdeduled actions since filtered by node_id
			var rply []*scheduler.ScheduledAction
			if err := client1.Call(context.Background(), utils.APIerSv1GetScheduledActions,
				scheduler.ArgsGetScheduledActions{}, &rply); err != nil {
				t.Error("Unexpected error: ", err)
			} else if len(rply) != 1 && rply[0].Accounts != 1 && rply[0].ActionPlanID != "APRecurring" && rply[0].ActionsID != "TOPUP_ACTION" {
				t.Errorf("ScheduledActions: %+v", utils.ToJSON(rply))
			}
			if err := client2.Call(context.Background(), utils.APIerSv1GetScheduledActions,
				scheduler.ArgsGetScheduledActions{}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
				t.Error("Unexpected error: ", err)
			}
		})
		if err := client1.Call(context.Background(), utils.APIerSv1SetDestination,
			&utils.AttrSetDestination{
				Id:       "DST_1001",
				Prefixes: []string{"+49"},
			},
			&reply); err != nil {
			t.Fatal(err)
		}
		attrs = &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}
		if err := client1.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acc1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Engine2GetData", func(t *testing.T) {
		var acnt *engine.Account
		if err := client2.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: "1001",
			},
			&acnt); err != nil {
			t.Fatalf("account 1001 not found on target: %v", err)
		}
		if !reflect.DeepEqual(acnt, acc1) {
			t.Errorf("expected <%#v>, \ngot <%#v>", acc1, acnt)
			t.Errorf("expected <%#v>, \ngot <%#v>", acc1.BalanceMap[utils.MetaMonetary].String(), acnt.BalanceMap[utils.MetaMonetary].String())
		}

		var dst engine.Destination
		if err := client2.Call(context.Background(), utils.APIerSv1GetDestination,
			"DST_1001",
			&dst); err != nil {
			t.Fatalf("destination DST_1001 not found on target: %v", err)
		}
		if !slices.Equal(dst.Prefixes, []string{"+49"}) {
			t.Errorf("expected destination prefixes %v, got %v", []string{"+49"}, dst.Prefixes)
		}
	})

	t.Run("Engine2SetData", func(t *testing.T) {
		var reply string
		topupAction := &utils.AttrSetActions{ActionsId: "TOPUP_ACTION2", Actions: []*utils.TPAction{
			{Identifier: utils.MetaTopUpReset, Filters: "*string:~*cfg.general.node_id:ng2", BalanceId: "BalID2", ExpiryTime: "+10s",
				BalanceType: utils.MetaMonetary, Units: "5", Weight: 10.0},
		}}

		if err := client2.Call(context.Background(), utils.APIerSv2SetActions, topupAction, &reply); err != nil {
			t.Fatal(err)
		}
		atms1 := &engine.AttrSetActionPlan{
			Id:              "APRecurring2",
			ReloadScheduler: true,
			ActionPlan: []*engine.AttrActionPlan{
				{
					ActionsId: "TOPUP_ACTION2",
					Time:      "+1s",
					Weight:    20.0,
					TimingID:  "+1s",
				},
				{
					ActionsId: "TOPUP_ACTION2",
					Time:      utils.MetaASAP,
					Weight:    21.0,
					TimingID:  utils.MetaASAP,
				},
			},
		}
		if err := client2.Call(context.Background(), utils.APIerSv1SetActionPlan, atms1, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client2.Call(context.Background(), utils.APIerSv2SetAccount,
			engine.AttrSetAccount{
				Tenant:          "cgrates.org",
				Account:         "1002",
				ActionPlanIDs:   []string{"APRecurring2"},
				ReloadScheduler: true,
			},
			&reply); err != nil {
			t.Fatal(err)
		}
		attrs := &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1002",
		}
		if err := client2.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acc2); err != nil {
			t.Error(err)
		}
		if err := client2.Call(context.Background(), utils.APIerSv1SetDestination,
			&utils.AttrSetDestination{
				Id:       "DST_1002",
				Prefixes: []string{"+45"},
			},
			&reply); err != nil {
			t.Fatal(err)
		}
		attrs = &utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1002",
		}
		if err := client2.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acc2); err != nil {
			t.Error(err)
		}
	})

	t.Run("Engine1GetData", func(t *testing.T) {
		var acnt *engine.Account
		if err := client1.Call(context.Background(), utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: "1002",
			},
			&acnt); err != nil {
			t.Fatalf("account 1002 not found on target: %v", err)
		}
		if !reflect.DeepEqual(acnt, acc2) {
			t.Errorf("expected <%#v>, \ngot <%#v>", acc2, acnt)
			t.Errorf("expected <%#v>, \ngot <%#v>", acc2.BalanceMap[utils.MetaMonetary].String(), acnt.BalanceMap[utils.MetaMonetary].String())
		}

		var dst engine.Destination
		if err := client1.Call(context.Background(), utils.APIerSv1GetDestination,
			"DST_1002",
			&dst); err != nil {
			t.Fatalf("destination DST_1002 not found on target: %v", err)
		}
		if !slices.Equal(dst.Prefixes, []string{"+45"}) {
			t.Errorf("expected destination prefixes %v, got %v", []string{"+45"}, dst.Prefixes)
		}
	})

	t.Run("GetSchedulers", func(t *testing.T) {
		var rply []*scheduler.ScheduledAction
		if err := client1.Call(context.Background(), utils.APIerSv1GetScheduledActions,
			scheduler.ArgsGetScheduledActions{}, &rply); err != nil {
			t.Error("Unexpected error: ", err)
		} else if len(rply) != 1 && rply[0].Accounts != 1 && rply[0].ActionPlanID != "APRecurring" && rply[0].ActionsID != "TOPUP_ACTION" {
			t.Errorf("ScheduledActions: %+v", utils.ToJSON(rply))
		}
		if err := client2.Call(context.Background(), utils.APIerSv1GetScheduledActions,
			scheduler.ArgsGetScheduledActions{}, &rply); err != nil {
			t.Error("Unexpected error: ", err)
		} else if len(rply) != 1 && rply[0].Accounts != 1 && rply[0].ActionPlanID != "APRecurring" && rply[0].ActionsID != "TOPUP_ACTION" {
			t.Errorf("ScheduledActions: %+v", utils.ToJSON(rply))
		}
	})

	entries, err := os.ReadDir(failedDirNG1)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected failed dir to be empty, found %d entries", len(entries))
	}

	entries2, err := os.ReadDir(failedDirNG2)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries2) != 0 {
		t.Fatal("expected 0 .gob files in failed dir, received", len(entries2))
	}

}
