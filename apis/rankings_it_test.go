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

package apis

import (
	"path"
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	raCfgPath   string
	raCfg       *config.CGRConfig
	raRPC       *birpc.Client
	raConfigDIR string //run tests for specific configuration

	sTestsRa = []func(t *testing.T){
		testRankingSInitCfg,
		testRankingSInitDataDB,
		testRankingSResetStorDB,
		testRankingsStartEngine,
		testRankingsRPCConn,

		//tests for AdminSv1 APIs
		testRankingsGetRankingProfileBeforeSet,
		testRankingsGetRankingProfileIDsBeforeSet,
		testRankingsGetRankingProfileCountBeforeSet,
		testRankingsGetRankingProfilesBeforeSet,
		testRankingsSetRankingProfiles,
		testRankingsGetRankingProfileAfterSet,
		testRankingsGetRankingProfileIDsAfterSet,
		testRankingsGetRankingProfileCountAfterSet,
		testRankingsGetRankingProfilesAfterSet,
		testRankingsRemoveRankingProfile,
		testRankingsGetRankingProfileAfterRemove,
		testRankingsGetRankingProfileIDsAfterRemove,
		testRankingsGetRankingProfileCountAfterRemove,
		testRankingsGetRankingProfilesAfterRemove,

		testRankingsKillEngine,
	}
)

func TestRankingsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		raConfigDIR = "rankings_internal"
	case utils.MetaMongo:
		raConfigDIR = "rankings_mongo"
	case utils.MetaMySQL:
		raConfigDIR = "rankings_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRa {
		t.Run(raConfigDIR, stest)
	}
}

func testRankingSInitCfg(t *testing.T) {
	var err error
	raCfgPath = path.Join(*dataDir, "conf", "samples", raConfigDIR)
	raCfg, err = config.NewCGRConfigFromPath(context.Background(), raCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testRankingSInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(raCfg); err != nil {
		t.Fatal(err)
	}
}

func testRankingSResetStorDB(t *testing.T) {
	if err := engine.InitStorDB(raCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testRankingsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(raCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testRankingsRPCConn(t *testing.T) {
	var err error
	raRPC, err = engine.NewRPCClient(raCfg.ListenCfg(), *encoding) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testRankingsGetRankingProfileBeforeSet(t *testing.T) {
	var replyRankingProfile engine.RankingProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "Test_RNK1",
			}}, &replyRankingProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testRankingsGetRankingProfilesBeforeSet(t *testing.T) {
	var replyRankingProfiles *[]*engine.RankingProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRankingProfiles); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testRankingsGetRankingProfileIDsBeforeSet(t *testing.T) {
	var replyRankingProfileIDs []string
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRankingProfileIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testRankingsGetRankingProfileCountBeforeSet(t *testing.T) {
	var replyCount int
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if replyCount != 0 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testRankingsSetRankingProfiles(t *testing.T) {
	rankingProfiles := []*engine.RankingProfileWithAPIOpts{
		{
			RankingProfile: &engine.RankingProfile{
				ID:     "TestA_Ranking1",
				Tenant: "cgrates.org",

				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
			},
		},
		{
			RankingProfile: &engine.RankingProfile{
				ID:     "TestA_Ranking2",
				Tenant: "cgrates.org",

				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
			},
		},
		{
			RankingProfile: &engine.RankingProfile{
				ID:     "TestA_Ranking3",
				Tenant: "cgrates.org",

				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
			},
		},
		{
			RankingProfile: &engine.RankingProfile{
				ID:                "TestB_Ranking1",
				Tenant:            "cgrates.org",
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
			},
		},
		{
			RankingProfile: &engine.RankingProfile{
				ID:                "TestB_Ranking2",
				Tenant:            "cgrates.org",
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
			},
		},
	}

	var reply string
	for _, rankingProfile := range rankingProfiles {
		if err := raRPC.Call(context.Background(), utils.AdminSv1SetRankingProfile,
			rankingProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testRankingsGetRankingProfileAfterSet(t *testing.T) {
	expectedRankingProfile := engine.RankingProfile{
		ID:                "TestA_Ranking1",
		Tenant:            "cgrates.org",
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
	}
	var replyRankingProfile engine.RankingProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Ranking1",
			}}, &replyRankingProfile); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyRankingProfile, expectedRankingProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expectedRankingProfile), utils.ToJSON(replyRankingProfile))
	}
}

func testRankingsGetRankingProfileIDsAfterSet(t *testing.T) {
	expectedIDs := []string{"TestA_Ranking1", "TestA_Ranking2", "TestA_Ranking3", "TestB_Ranking1", "TestB_Ranking2"}
	var replyRankingProfileIDs []string
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRankingProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRankingProfileIDs)
		if !slices.Equal(replyRankingProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRankingProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_Ranking1", "TestA_Ranking2", "TestA_Ranking3"}
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyRankingProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRankingProfileIDs)
		if !slices.Equal(replyRankingProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRankingProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_Ranking1", "TestB_Ranking2"}
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyRankingProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRankingProfileIDs)
		if !slices.Equal(replyRankingProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRankingProfileIDs)
		}
	}
}

func testRankingsGetRankingProfileCountAfterSet(t *testing.T) {
	var replyCount int
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 5 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 3 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testRankingsGetRankingProfilesAfterSet(t *testing.T) {
	expectedRankingProfiles := []*engine.RankingProfile{
		{
			ID:                "TestA_Ranking1",
			Tenant:            "cgrates.org",
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
		},
		{
			ID:     "TestA_Ranking2",
			Tenant: "cgrates.org",

			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
		},
		{
			ID:                "TestA_Ranking3",
			Tenant:            "cgrates.org",
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
		},
		{
			ID:                "TestB_Ranking1",
			Tenant:            "cgrates.org",
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
		},
		{
			ID:                "TestB_Ranking2",
			Tenant:            "cgrates.org",
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
		},
	}
	var replyRankingProfiles []*engine.RankingProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRankingProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyRankingProfiles, func(i, j int) bool {
			return replyRankingProfiles[i].ID < replyRankingProfiles[j].ID
		})
		if !reflect.DeepEqual(replyRankingProfiles, expectedRankingProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedRankingProfiles), utils.ToJSON(replyRankingProfiles))
		}
	}
}

func testRankingsRemoveRankingProfile(t *testing.T) {
	var reply string
	if err := raRPC.Call(context.Background(), utils.AdminSv1RemoveRankingProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Ranking2",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testRankingsGetRankingProfileAfterRemove(t *testing.T) {
	var replyRankingProfile engine.RankingProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Ranking2",
			}}, &replyRankingProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testRankingsGetRankingProfileIDsAfterRemove(t *testing.T) {
	expectedIDs := []string{"TestA_Ranking1", "TestA_Ranking3", "TestB_Ranking1", "TestB_Ranking2"}
	var replyRankingProfileIDs []string
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRankingProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRankingProfileIDs)
		if !slices.Equal(replyRankingProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRankingProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_Ranking1", "TestA_Ranking3"}
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyRankingProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRankingProfileIDs)
		if !slices.Equal(replyRankingProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRankingProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_Ranking1", "TestB_Ranking2"}
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyRankingProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRankingProfileIDs)
		if !slices.Equal(replyRankingProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRankingProfileIDs)
		}
	}
}

func testRankingsGetRankingProfileCountAfterRemove(t *testing.T) {
	var replyCount int
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 4 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testRankingsGetRankingProfilesAfterRemove(t *testing.T) {
	expectedRankingProfiles := []*engine.RankingProfile{
		{
			ID:                "TestA_Ranking1",
			Tenant:            "cgrates.org",
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
		},
		{
			ID:                "TestA_Ranking3",
			Tenant:            "cgrates.org",
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
		},
		{
			ID:                "TestB_Ranking1",
			Tenant:            "cgrates.org",
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
		},
		{
			ID:                "TestB_Ranking2",
			Tenant:            "cgrates.org",
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
		},
	}
	var replyRankingProfiles []*engine.RankingProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRankingProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRankingProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyRankingProfiles, func(i, j int) bool {
			return replyRankingProfiles[i].ID < replyRankingProfiles[j].ID
		})
		if !reflect.DeepEqual(replyRankingProfiles, expectedRankingProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedRankingProfiles), utils.ToJSON(replyRankingProfiles))
		}
	}
}

// Kill the engine when it is about to be finished
func testRankingsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
