//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"slices"
	"strings"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestFraudDetectionAccountDisable(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{
"general": {
	"log_level": 7
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"rals": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"stats_conns": ["*internal"],
	"thresholds_conns": ["*internal"]
},
"filters": {
	"stats_conns": ["*internal"]
},
"stats": {
	"enabled": true,
	"indexed_selects": false,
	"store_interval": "-1"
},
"thresholds": {
	"enabled": true,
	"indexed_selects": false,
	"store_interval": "-1"
},
"apiers": {
	"enabled": true
}
}`

	var logBuf bytes.Buffer
	testEnv := TestEnvironment{
		ConfigJSON: content,
		LogBuffer:  &logBuf,
	}
	client, _ := testEnv.Setup(t, *utils.WaitRater)
	t.Cleanup(func() {
		if t.Failed() {
			t.Log(logBuf.String())
		}
	})

	nightEvent := map[string]any{
		utils.Account:   "1001",
		utils.SetupTime: time.Date(2026, 11, 18, 23, 0, 0, 0, time.UTC),
	}

	if !t.Run("SetAccount", func(t *testing.T) {
		var reply string
		if err := client.Call(utils.APIerSv1SetBalance, &utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1001",
			BalanceType: utils.MONETARY,
			Value:       10,
			Balance:     map[string]any{utils.ID: "test"},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		var acnt engine.Account
		if err := client.Call(utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		if acnt.Disabled {
			t.Fatal("account 1001 should be enabled initially")
		}
	}) {
		return
	}

	if !t.Run("SetProfiles", func(t *testing.T) {
		var reply string

		if err := client.Call(utils.APIerSv1SetFilter, &v1.FilterWithCache{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_NIGHT",
				Rules: []*engine.FilterRule{{
					Type:    utils.MetaRSR,
					Element: utils.EmptyString,
					Values:  []string{`~*req.SetupTime(~T0[0-7]|T2[0-3])`},
				}},
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.APIerSv1SetFilter, &v1.FilterWithCache{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_FRAUD_10",
				Rules: []*engine.FilterRule{{
					Type:    utils.MetaGreaterOrEqual,
					Element: "~*stats.SQ_FRAUD.*sum:1",
					Values:  []string{"10"},
				}},
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}

		if err := client.Call(utils.APIerSv1SetStatQueueProfile, &engine.StatQueueWithCache{
			StatQueueProfile: &engine.StatQueueProfile{
				Tenant:      "cgrates.org",
				ID:          "SQ_FRAUD",
				FilterIDs:   []string{"*string:~*req.Account:1001", "FLTR_NIGHT"},
				QueueLength: 0,
				TTL:         3 * time.Second,
				MinItems:    1,
				Metrics:     []*engine.MetricWithFilters{{MetricID: "*sum:1"}},
				Weight:      20,
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}

		if err := client.Call(utils.APIerSv2SetActions, &utils.AttrSetActions{
			ActionsId: "ACT_FRAUD_DISABLE",
			Actions:   []*utils.TPAction{{Identifier: utils.DISABLE_ACCOUNT}, {Identifier: utils.LOG}},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.APIerSv2SetActions, &utils.AttrSetActions{
			ActionsId: "ACT_FRAUD_ENABLE",
			Actions:   []*utils.TPAction{{Identifier: utils.ENABLE_ACCOUNT}},
		}, &reply); err != nil {
			t.Fatal(err)
		}

		if err := client.Call(utils.APIerSv1SetThresholdProfile, &engine.ThresholdWithCache{
			ThresholdProfile: &engine.ThresholdProfile{
				Tenant:    "cgrates.org",
				ID:        "TH_FRAUD_TEMP",
				FilterIDs: []string{"FLTR_FRAUD_10", "*string:~*req.Account:1001"},
				MaxHits:   1,
				Weight:    20,
				ActionIDs: []string{"ACT_FRAUD_DISABLE"},
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.APIerSv1SetThresholdProfile, &engine.ThresholdWithCache{
			ThresholdProfile: &engine.ThresholdProfile{
				Tenant:    "cgrates.org",
				ID:        "TH_FRAUD_PERM",
				FilterIDs: []string{"FLTR_FRAUD_10", "*string:~*req.Account:1001"},
				MinHits:   2,
				MaxHits:   -1,
				Weight:    10,
				ActionIDs: []string{"ACT_FRAUD_DISABLE"},
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}
	}) {
		return
	}

	if !t.Run("FirstDisable", func(t *testing.T) {
		var reply sessions.V1ProcessEventReply

		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv1", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv2", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv3", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv4", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv5", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv6", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv7", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv8", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}

		cgrEv := &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv9", Event: nightEvent}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: cgrEv,
		}, &reply); err != nil {
			t.Fatal(err)
		}
		reply = sessions.V1ProcessEventReply{}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaThresholds},
			CGREvent: cgrEv,
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.ThresholdIDs != nil && len(*reply.ThresholdIDs) != 0 {
			t.Fatalf("expected no threshold IDs, received %v", *reply.ThresholdIDs)
		}
		var acnt engine.Account
		if err := client.Call(utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		if acnt.Disabled {
			t.Fatal("account 1001 should still be enabled after the ninth event")
		}

		cgrEv = &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv10", Event: nightEvent}
		reply = sessions.V1ProcessEventReply{}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: cgrEv,
		}, &reply); err != nil {
			t.Fatal(err)
		}
		reply = sessions.V1ProcessEventReply{}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaThresholds},
			CGREvent: cgrEv,
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.ThresholdIDs == nil {
			t.Fatal("expected threshold IDs, received none")
		}
		if expIDs := []string{"TH_FRAUD_TEMP", "TH_FRAUD_PERM"}; !slices.Equal(*reply.ThresholdIDs, expIDs) {
			t.Fatalf("expected threshold IDs %v, received %v", expIDs, *reply.ThresholdIDs)
		}

		if err := client.Call(utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		if !acnt.Disabled {
			t.Fatal("account 1001 should be disabled after the first trigger")
		}

		var th engine.Threshold
		if err := client.Call(utils.ThresholdSv1GetThreshold,
			&utils.TenantID{Tenant: "cgrates.org", ID: "TH_FRAUD_PERM"}, &th); err != nil {
			t.Fatal(err)
		}
		if th.Hits != 1 || !th.Snooze.IsZero() {
			t.Fatalf("TH_FRAUD_PERM: expected Hits 1 and zero Snooze, received Hits %d, Snooze %v", th.Hits, th.Snooze)
		}

		logs := logBuf.String()
		if !strings.Contains(logs, `LOG Account: {"ID":"cgrates.org:1001"`) ||
			!strings.Contains(logs, `"Disabled":true`) {
			t.Fatalf("expected a *log entry for the disabled account 1001, log buffer:\n%s", logs)
		}
	}) {
		return
	}

	if !t.Run("AdminReenables", func(t *testing.T) {
		var reply string
		if err := client.Call(utils.APIerSv1ExecuteAction, &utils.AttrExecuteAction{
			Tenant: "cgrates.org", Account: "1001", ActionsId: "ACT_FRAUD_ENABLE",
		}, &reply); err != nil {
			t.Fatal(err)
		}
		var acnt engine.Account
		if err := client.Call(utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		if acnt.Disabled {
			t.Fatal("account 1001 should be enabled again")
		}
	}) {
		return
	}

	t.Run("SecondDisable", func(t *testing.T) {
		time.Sleep(4 * time.Second)

		var reply sessions.V1ProcessEventReply

		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv11", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv12", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv13", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv14", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv15", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv16", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv17", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv18", Event: nightEvent},
		}, &reply); err != nil {
			t.Fatal(err)
		}

		cgrEv := &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv19", Event: nightEvent}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: cgrEv,
		}, &reply); err != nil {
			t.Fatal(err)
		}
		reply = sessions.V1ProcessEventReply{}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaThresholds},
			CGREvent: cgrEv,
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.ThresholdIDs != nil && len(*reply.ThresholdIDs) != 0 {
			t.Fatalf("expected no threshold IDs, received %v", *reply.ThresholdIDs)
		}

		cgrEv = &utils.CGREvent{Tenant: "cgrates.org", ID: "fraudEv20", Event: nightEvent}
		reply = sessions.V1ProcessEventReply{}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaStats},
			CGREvent: cgrEv,
		}, &reply); err != nil {
			t.Fatal(err)
		}
		reply = sessions.V1ProcessEventReply{}
		if err := client.Call(utils.SessionSv1ProcessEvent, &sessions.V1ProcessEventArgs{
			Flags:    []string{utils.MetaThresholds},
			CGREvent: cgrEv,
		}, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.ThresholdIDs == nil {
			t.Fatal("expected threshold IDs, received none")
		}
		if expIDs := []string{"TH_FRAUD_PERM"}; !slices.Equal(*reply.ThresholdIDs, expIDs) {
			t.Fatalf("expected threshold IDs %v, received %v", expIDs, *reply.ThresholdIDs)
		}

		var th engine.Threshold
		if err := client.Call(utils.ThresholdSv1GetThreshold,
			&utils.TenantID{Tenant: "cgrates.org", ID: "TH_FRAUD_PERM"}, &th); err != nil {
			t.Fatal(err)
		}
		if th.Hits != 2 || th.Snooze.IsZero() {
			t.Fatalf("TH_FRAUD_PERM: expected Hits 2 and nonzero Snooze, received Hits %d, Snooze %v", th.Hits, th.Snooze)
		}

		var acnt engine.Account
		if err := client.Call(utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &acnt); err != nil {
			t.Fatal(err)
		}
		if !acnt.Disabled {
			t.Fatal("account 1001 should be disabled after the second trigger")
		}
	})
}
