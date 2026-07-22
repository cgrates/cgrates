//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestDspTr = []func(t *testing.T){
	testDspTrPingFailover,
	testDspTrGetTrendFailover,
	testDspTrPing,
	testDspTrTestAuthKey,
	testDspTrTestAuthKey2,
}

func TestDspTrendS(t *testing.T) {
	var config1, config2, config3 string
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "trd_rnk_mysql"
		config2 = "trd_rnk2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "trd_rnk_mongo"
		config2 = "trd_rnk2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	testDsp(t, sTestDspTr, "TestDspTrendS", config1, config2, config3, "tuttrends", "testtp", "dispatchers")
}

func testDspTrPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.TrendSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsAPIKey: "tr12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.TrendSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	time.Sleep(200 * time.Millisecond)
	if err := dispEngine.RPC.Call(context.Background(), utils.TrendSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.TrendSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspTrGetTrendFailover(t *testing.T) {
	trSched := utils.ArgScheduleTrendQueries{TrendIDs: []string{"TREND_1"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}, APIOpts: map[string]any{utils.OptsAPIKey: "tr12345"}}}
	var scheduled int

	if err := dispEngine.RPC.Call(context.Background(), utils.TrendSv1ScheduleQueries, &trSched,
		&scheduled); err != nil {
		t.Error(err)
	} else if scheduled != 1 {
		t.Errorf("expected %d,received %d", 1, scheduled)
	}
	allEngine.stopEngine(t)
	time.Sleep(200 * time.Millisecond)
	if err := dispEngine.RPC.Call(context.Background(), utils.TrendSv1ScheduleQueries, &trSched,
		&scheduled); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("expected err %v,received %v", utils.ErrPartiallyExecuted, err)
	}
	allEngine.startEngine(t)
}

func testDspTrPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.TrendSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.TrendSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsAPIKey: "tr12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspTrTestAuthKey(t *testing.T) {
	var tr *engine.Trend
	args := &utils.ArgGetTrend{
		TenantWithAPIOpts: utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.OptsAPIKey: "12345",
			},
		},
		ID: "TREND_1",
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.TrendSv1GetTrend,
		args, &tr); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
	trSched := utils.ArgScheduleTrendQueries{TrendIDs: []string{"TREND_1", "TREND_2"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}, APIOpts: map[string]any{utils.OptsAPIKey: "12345"}}}
	var scheduled int
	if err := dispEngine.RPC.Call(context.Background(), utils.TrendSv1ScheduleQueries,
		&trSched, &scheduled); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}
func testDspTrTestAuthKey2(t *testing.T) {
	var schedTrends []utils.ScheduledTrend
	if err := dispEngine.RPC.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}, APIOpts: map[string]any{utils.OptsAPIKey: "tr12345"}}, TrendIDPrefixes: []string{"TREND_1"}}, &schedTrends); err != nil {
		t.Error(err)
	} else if len(schedTrends) != 1 {
		t.Errorf("expected 1 schedTrends, got %d", len(schedTrends))
	}
}
