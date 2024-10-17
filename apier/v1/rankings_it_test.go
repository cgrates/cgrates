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
package v1

import (
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	rankingCfgPath   string
	rankingCfg       *config.CGRConfig
	rankingRPC       *birpc.Client
	rankingProfile   *engine.RankingProfileWithAPIOpts
	rankingConfigDIR string

	sTestsRanking = []func(t *testing.T){
		testRankingSInitCfg,
		testRankingSInitDataDb,
		testRankingSResetStorDb,
		testRankingSStartEngine,
		testRankingSRPCConn,
		testRankingSLoadAdd,
		testRankingSSetRankingProfile,
		testRankingSGetRankingProfileIDs,
		testRankingSUpdateRankingProfile,
		testRankingSRemRankingProfile,
		testRankingSKillEngine,
		//cache test
		testRankingSInitCfg,
		testRankingSInitDataDb,
		testRankingSResetStorDb,
		testRankingSStartEngine,
		testRankingSRPCConn,
		testRankingSCacheTestGetNotFound,
		testRankingSCacheTestSet,
		testRankingSCacheReload,
		testRankingSCacheTestGetFound,
		testRankingSKillEngine,
	}
)

func TestRankingSIT(t *testing.T) {

	switch *utils.DBType {
	case utils.MetaInternal:
		rankingConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		rankingConfigDIR = "tutmysql"
	case utils.MetaMongo:
		rankingConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRanking {
		t.Run(rankingConfigDIR, stest)
	}
}

func testRankingSInitCfg(t *testing.T) {
	var err error
	rankingCfgPath = path.Join(*utils.DataDir, "conf", "samples", rankingConfigDIR)
	rankingCfg, err = config.NewCGRConfigFromPath(rankingCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testRankingSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rankingCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testRankingSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rankingCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testRankingSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rankingCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testRankingSRPCConn(t *testing.T) {
	var err error
	rankingRPC, err = newRPCClient(rankingCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}
func testRankingSLoadAdd(t *testing.T) {
	rankingProfile := &engine.RankingProfileWithAPIOpts{
		RankingProfile: &engine.RankingProfile{
			Tenant:    "cgrates.org",
			ID:        "SG_Sum",
			Schedule:  "@every 15m",
			StatIDs:   []string{"SQProfile1"},
			MetricIDs: []string{"*sum"},
		},
	}

	var result string
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1SetRankingProfile, rankingProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testRankingSSetRankingProfile(t *testing.T) {
	var (
		reply  *engine.RankingProfileWithAPIOpts
		result string
	)
	rankingProfile = &engine.RankingProfileWithAPIOpts{
		RankingProfile: &engine.RankingProfile{
			Tenant:       "cgrates.org",
			ID:           "Ranking1",
			Schedule:     "@every 15m",
			StatIDs:      []string{"SQ1", "SQ2"},
			MetricIDs:    []string{"*acc", "*sum"},
			Sorting:      "*asc",
			ThresholdIDs: []string{"THD1", "THD2"}},
	}
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1GetRankingProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Ranking1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1SetRankingProfile, rankingProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expected: %v,Received: %v", utils.OK, result)
	}
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1GetRankingProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Ranking1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(rankingProfile, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profile (-expected +got):\n%s", diff)
	}
}
func testRankingSGetRankingProfileIDs(t *testing.T) {
	expected := []string{"Ranking1", "SG_Sum"}
	var result []string
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1GetRankingProfileIDs, utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testRankingSUpdateRankingProfile(t *testing.T) {
	rankingProfile.Sorting = "*desc"
	var (
		reply  *engine.RankingProfileWithAPIOpts
		result string
	)
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1SetRankingProfile, rankingProfile, &result); err != nil {
		t.Error(err)
	}
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1GetRankingProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Ranking1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(rankingProfile, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profile (-expected +got):\n%s", diff)
	}
}
func testRankingSRemRankingProfile(t *testing.T) {
	var (
		resp  string
		reply *engine.RankingProfile
	)
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1RemoveRankingProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Ranking1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

	if err := rankingRPC.Call(context.Background(), utils.APIerSv1GetRankingProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Ranking1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testRankingSKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}

func testRankingSCacheTestGetNotFound(t *testing.T) {
	var reply *engine.RankingProfile
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1GetRankingProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RANKINGPRF_CACHE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
}

func testRankingSCacheTestGetFound(t *testing.T) {
	var reply *engine.RankingProfile
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1GetRankingProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RANKINGPRF_CACHE"}, &reply); err != nil {
		t.Fatal(err)
	}
}

func testRankingSCacheTestSet(t *testing.T) {
	rankingProfile = &engine.RankingProfileWithAPIOpts{
		RankingProfile: &engine.RankingProfile{
			Tenant: "cgrates.org",
			ID:     "RANKINGPRF_CACHE",
		},
		APIOpts: map[string]any{
			utils.CacheOpt: utils.MetaNone,
		},
	}
	var reply string
	if err := rankingRPC.Call(context.Background(), utils.APIerSv1SetRankingProfile, rankingProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testRankingSCacheReload(t *testing.T) {
	cache := &utils.AttrReloadCacheWithAPIOpts{
		RankingProfileIDs: []string{"cgrates.org:RANKINGPRF_CACHE"},
	}
	var reply string
	if err := rankingRPC.Call(context.Background(), utils.CacheSv1ReloadCache, cache, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}
