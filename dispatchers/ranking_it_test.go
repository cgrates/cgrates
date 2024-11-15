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

package dispatchers

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestDspRn = []func(t *testing.T){
	testDspRnPingFailover,
	testDspRnGetRankingFailover,
	testDspRnPing,
	testDspRnTestAuthKey,
	testDspRnTestAuthKey2,
}

func TestDspRankingS(t *testing.T) {
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
	testDsp(t, sTestDspRn, "TestDspRankingS", config1, config2, config3, "tutrankings", "testtp", "dispatchers")
}

func testDspRnPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.RankingSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsAPIKey: "rn12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RankingSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	time.Sleep(200 * time.Millisecond)
	if err := dispEngine.RPC.Call(context.Background(), utils.RankingSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.RankingSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspRnGetRankingFailover(t *testing.T) {
	trSched := utils.ArgScheduleRankingQueries{RankingIDs: []string{"RANK2"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}, APIOpts: map[string]any{utils.OptsAPIKey: "rn12345"}}}
	var scheduled int
	if err := dispEngine.RPC.Call(context.Background(), utils.RankingSv1ScheduleQueries, &trSched,
		&scheduled); err != nil {
		t.Error(err)
	} else if scheduled != 1 {
		t.Errorf("expected %d,received %d", 1, scheduled)
	}
	allEngine.stopEngine(t)
	time.Sleep(200 * time.Millisecond)
	if err := dispEngine.RPC.Call(context.Background(), utils.RankingSv1ScheduleQueries, &trSched,
		&scheduled); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("expected err %v,received %v", utils.ErrPartiallyExecuted, err)
	}
	allEngine.startEngine(t)
}

func testDspRnPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.RankingSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RankingSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsAPIKey: "rn12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspRnTestAuthKey(t *testing.T) {
	var rn *engine.Ranking
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RANK1",
		},
		APIOpts: map[string]any{
			utils.OptsAPIKey: "12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RankingSv1GetRanking,
		args, &rn); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
	rnSched := utils.ArgScheduleRankingQueries{RankingIDs: []string{"RANK1"}, TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}, APIOpts: map[string]any{utils.OptsAPIKey: "12345"}}}
	var scheduled int
	if err := dispEngine.RPC.Call(context.Background(), utils.RankingSv1ScheduleQueries,
		&rnSched, &scheduled); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspRnTestAuthKey2(t *testing.T) {
	var schedRankings []utils.ScheduledRanking
	if err := dispEngine.RPC.Call(context.Background(), utils.RankingSv1GetSchedule, &utils.ArgScheduledRankings{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}, APIOpts: map[string]any{utils.OptsAPIKey: "rn12345"}}, RankingIDPrefixes: []string{"RANK2"}}, &schedRankings); err != nil {
		t.Error(err)
	} else if len(schedRankings) != 1 {
		t.Errorf("expected 1 schedTrends, got %d", len(schedRankings))
	}
}
