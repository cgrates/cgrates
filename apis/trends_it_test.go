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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	trCfgPath   string
	trCfg       *config.CGRConfig
	trRPC       *birpc.Client
	trConfigDIR string //run tests for specific configuration

	sTestsTr = []func(t *testing.T){
		testTrendSInitCfg,
		testTrendSInitDataDB,
		testTrendSResetStorDB,
		testTrendsStartEngine,
		testTrendsRPCConn,

		//tests for AdminSv1 APIs
		testTrendsGetTrendProfileBeforeSet,
		testTrendsGetTrendProfileIDsBeforeSet,
		testTrendsGetTrendProfileCountBeforeSet,
		testTrendsGetTrendProfilesBeforeSet,
		testTrendsSetTrendProfiles,
		testTrendsGetTrendProfileAfterSet,
		testTrendsGetTrendProfileIDsAfterSet,
		testTrendsGetTrendProfileCountAfterSet,
		testTrendsGetTrendProfilesAfterSet,
		testTrendsRemoveTrendProfile,
		testTrendsGetTrendProfileAfterRemove,
		testTrendsGetTrendProfileIDsAfterRemove,
		testTrendsGetTrendProfileCountAfterRemove,
		testTrendsGetTrendProfilesAfterRemove,

		testTrendsKillEngine,
	}
)

func TestTrendsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		trConfigDIR = "trends_internal"
	case utils.MetaMongo:
		trConfigDIR = "trends_mongo"
	case utils.MetaMySQL:
		trConfigDIR = "trends_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTr {
		t.Run(trConfigDIR, stest)
	}
}

func testTrendSInitCfg(t *testing.T) {
	var err error
	trCfgPath = path.Join(*dataDir, "conf", "samples", trConfigDIR)
	trCfg, err = config.NewCGRConfigFromPath(context.Background(), trCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testTrendSInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(trCfg); err != nil {
		t.Fatal(err)
	}
}

func testTrendSResetStorDB(t *testing.T) {
	if err := engine.InitStorDB(trCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTrendsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(trCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testTrendsRPCConn(t *testing.T) {
	var err error
	trRPC, err = newRPCClient(trCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTrendsGetTrendProfileBeforeSet(t *testing.T) {
	var replyTrendProfile engine.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "Test_RNK1",
			}}, &replyTrendProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testTrendsGetTrendProfilesBeforeSet(t *testing.T) {
	var replyTrendProfiles *[]*engine.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyTrendProfiles); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testTrendsGetTrendProfileIDsBeforeSet(t *testing.T) {
	var replyTrendProfileIDs []string
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyTrendProfileIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testTrendsGetTrendProfileCountBeforeSet(t *testing.T) {
	var replyCount int
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if replyCount != 0 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testTrendsSetTrendProfiles(t *testing.T) {
	TrendProfiles := []*engine.TrendProfileWithAPIOpts{
		{
			TrendProfile: &engine.TrendProfile{
				ID:     "TestA_Trend1",
				Tenant: "cgrates.org",
			},
		},
		{
			TrendProfile: &engine.TrendProfile{
				ID:     "TestA_Trend2",
				Tenant: "cgrates.org",
			},
		},
		{
			TrendProfile: &engine.TrendProfile{
				ID:     "TestA_Trend3",
				Tenant: "cgrates.org",
			},
		},
		{
			TrendProfile: &engine.TrendProfile{
				ID:     "TestB_Trend1",
				Tenant: "cgrates.org",
			},
		},
		{
			TrendProfile: &engine.TrendProfile{
				ID:     "TestB_Trend2",
				Tenant: "cgrates.org",
			},
		},
	}

	var reply string
	for _, TrendProfile := range TrendProfiles {
		if err := trRPC.Call(context.Background(), utils.AdminSv1SetTrendProfile,
			TrendProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testTrendsGetTrendProfileAfterSet(t *testing.T) {
	expectedTrendProfile := engine.TrendProfile{
		ID:     "TestA_Trend1",
		Tenant: "cgrates.org",
	}
	var replyTrendProfile engine.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Trend1",
			}}, &replyTrendProfile); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyTrendProfile, expectedTrendProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expectedTrendProfile), utils.ToJSON(replyTrendProfile))
	}
}

func testTrendsGetTrendProfileIDsAfterSet(t *testing.T) {
	expectedIDs := []string{"TestA_Trend1", "TestA_Trend2", "TestA_Trend3", "TestB_Trend1", "TestB_Trend2"}
	var replyTrendProfileIDs []string
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyTrendProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyTrendProfileIDs)
		if !slices.Equal(replyTrendProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyTrendProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_Trend1", "TestA_Trend2", "TestA_Trend3"}
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyTrendProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyTrendProfileIDs)
		if !slices.Equal(replyTrendProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyTrendProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_Trend1", "TestB_Trend2"}
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyTrendProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyTrendProfileIDs)
		if !slices.Equal(replyTrendProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyTrendProfileIDs)
		}
	}
}

func testTrendsGetTrendProfileCountAfterSet(t *testing.T) {
	var replyCount int
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 5 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 3 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testTrendsGetTrendProfilesAfterSet(t *testing.T) {
	expectedTrendProfiles := []*engine.TrendProfile{
		{
			ID:     "TestA_Trend1",
			Tenant: "cgrates.org",
		},
		{
			ID:     "TestA_Trend2",
			Tenant: "cgrates.org",
		},
		{
			ID:     "TestA_Trend3",
			Tenant: "cgrates.org",
		},
		{
			ID:     "TestB_Trend1",
			Tenant: "cgrates.org",
		},
		{
			ID:     "TestB_Trend2",
			Tenant: "cgrates.org",
		},
	}
	var replyTrendProfiles []*engine.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyTrendProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyTrendProfiles, func(i, j int) bool {
			return replyTrendProfiles[i].ID < replyTrendProfiles[j].ID
		})
		if !reflect.DeepEqual(replyTrendProfiles, expectedTrendProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedTrendProfiles), utils.ToJSON(replyTrendProfiles))
		}
	}
}

func testTrendsRemoveTrendProfile(t *testing.T) {
	var reply string
	if err := trRPC.Call(context.Background(), utils.AdminSv1RemoveTrendProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Trend2",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testTrendsGetTrendProfileAfterRemove(t *testing.T) {
	var replyTrendProfile engine.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Trend2",
			}}, &replyTrendProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testTrendsGetTrendProfileIDsAfterRemove(t *testing.T) {
	expectedIDs := []string{"TestA_Trend1", "TestA_Trend3", "TestB_Trend1", "TestB_Trend2"}
	var replyTrendProfileIDs []string
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyTrendProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyTrendProfileIDs)
		if !slices.Equal(replyTrendProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyTrendProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_Trend1", "TestA_Trend3"}
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyTrendProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyTrendProfileIDs)
		if !slices.Equal(replyTrendProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyTrendProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_Trend1", "TestB_Trend2"}
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyTrendProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyTrendProfileIDs)
		if !slices.Equal(replyTrendProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyTrendProfileIDs)
		}
	}
}

func testTrendsGetTrendProfileCountAfterRemove(t *testing.T) {
	var replyCount int
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 4 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testTrendsGetTrendProfilesAfterRemove(t *testing.T) {
	expectedTrendProfiles := []*engine.TrendProfile{
		{
			ID:     "TestA_Trend1",
			Tenant: "cgrates.org",
		},
		{
			ID:     "TestA_Trend3",
			Tenant: "cgrates.org",
		},
		{
			ID:     "TestB_Trend1",
			Tenant: "cgrates.org",
		},
		{
			ID:     "TestB_Trend2",
			Tenant: "cgrates.org",
		},
	}
	var replyTrendProfiles []*engine.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyTrendProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyTrendProfiles, func(i, j int) bool {
			return replyTrendProfiles[i].ID < replyTrendProfiles[j].ID
		})
		if !reflect.DeepEqual(replyTrendProfiles, expectedTrendProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedTrendProfiles), utils.ToJSON(replyTrendProfiles))
		}
	}
}

// Kill the engine when it is about to be finished
func testTrendsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
