//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspTh = []func(t *testing.T){
	testDspThPingFailover,
	testDspThProcessEventFailover,

	testDspThPing,
	testDspThTestAuthKey,
	testDspThTestAuthKey2,
	testDspThTestAuthKey3,
}

// Test start here
func TestDspThresholdS(t *testing.T) {
	var config1, config2, config3 string
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "all_mysql"
		config2 = "all2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "all_mongo"
		config2 = "all2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	dispDIR := "dispatchers"
	if *utils.Encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspTh, "TestDspThresholdS", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspThPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.ThresholdSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsAPIKey: "thr12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspThProcessEventFailover(t *testing.T) {
	var ids []string
	eIDs := []string{"THD_ACNT_1001"}
	nowTime := time.Now()
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   &nowTime,
		Event: map[string]any{
			utils.EventName:    "Event1",
			utils.AccountField: "1001"},

		APIOpts: map[string]any{
			utils.OptsAPIKey: "thr12345",
		},
	}

	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1ProcessEvent, args,
		&ids); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error NOT_FOUND but received %v and reply %v\n", err, ids)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1ProcessEvent, args, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
	allEngine2.startEngine(t)
}

func testDspThPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.ThresholdSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsAPIKey: "thr12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspThTestAuthKey(t *testing.T) {
	var ids []string
	nowTime := time.Now()
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   &nowTime,
		Event: map[string]any{
			utils.AccountField: "1002"},

		APIOpts: map[string]any{
			utils.OptsAPIKey: "12345",
		},
	}

	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1ProcessEvent,
		args, &ids); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
	var th *engine.Thresholds
	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent, args,
		&th); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspThTestAuthKey2(t *testing.T) {
	var ids []string
	eIDs := []string{"THD_ACNT_1002"}
	nowTime := time.Now()
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   &nowTime,
		Event: map[string]any{
			utils.AccountField: "1002"},

		APIOpts: map[string]any{
			utils.OptsAPIKey: "thr12345",
		},
	}

	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1ProcessEvent, args, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
	var th *engine.Thresholds
	eTh := &engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
			Hits:   1,
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent, args, &th); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual((*eTh)[0].Tenant, (*th)[0].Tenant) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].Tenant, (*th)[0].Tenant)
	} else if !reflect.DeepEqual((*eTh)[0].ID, (*th)[0].ID) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].ID, (*th)[0].ID)
	} else if !reflect.DeepEqual((*eTh)[0].Hits, (*th)[0].Hits) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].Hits, (*th)[0].Hits)
	}
}

func testDspThTestAuthKey3(t *testing.T) {
	var th *engine.Threshold
	eTh := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1002",
		Hits:   1,
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1GetThreshold, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
		},
		APIOpts: map[string]any{
			utils.OptsAPIKey: "thr12345",
		},
	}, &th); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual((*eTh).Tenant, (*th).Tenant) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh).Tenant, (*th).Tenant)
	} else if !reflect.DeepEqual((*eTh).ID, (*th).ID) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh).ID, (*th).ID)
	} else if !reflect.DeepEqual((*eTh).Hits, (*th).Hits) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh).Hits, (*th).Hits)
	}

	var ids []string
	eIDs := []string{"THD_ACNT_1001", "THD_ACNT_1002"}

	if err := dispEngine.RPC.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs, &utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsAPIKey: "thr12345",
		},
	}, &ids); err != nil {
		t.Fatal(err)
	}
	sort.Strings(ids)
	if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
}
