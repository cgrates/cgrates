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

package trends

import (
	"path"
	"reflect"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	trCfgPath     string
	trCfg         *config.CGRConfig
	trRPC         *birpc.Client
	trConfigDIR   string //run tests for specific configuration
	trendProfiles []*utils.TrendProfileWithAPIOpts

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
	switch *utils.DBType {
	case utils.MetaInternal:
		trConfigDIR = "tutinternal"
	case utils.MetaMongo:
		trConfigDIR = "tutmongo"
	case utils.MetaMySQL:
		trConfigDIR = "tutmysql"
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
	trCfgPath = path.Join(*utils.DataDir, "conf", "samples", trConfigDIR)
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
	if _, err := engine.StopStartEngine(trCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testTrendsRPCConn(t *testing.T) {
	trRPC = engine.NewRPCClient(t, trCfg.ListenCfg(), *utils.Encoding)
}

func testTrendsGetTrendProfileBeforeSet(t *testing.T) {
	var replyTrendProfile utils.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "Trend1",
			}}, &replyTrendProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testTrendsGetTrendProfilesBeforeSet(t *testing.T) {
	var replyTrendProfiles *[]*utils.TrendProfile
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
	trendProfiles = []*utils.TrendProfileWithAPIOpts{
		{
			TrendProfile: &utils.TrendProfile{
				ID:              "Trend1",
				StatID:          "Stats1",
				Tenant:          "cgrates.org",
				Schedule:        "0 0 * * *",
				MinItems:        3,
				CorrelationType: "*last",
				QueueLength:     -1,
				TTL:             0,
				Tolerance:       0.5,
				Stored:          true,
				ThresholdIDs:    []string{"Th1"},
				Metrics:         []string{"*acc"},
			},
		},
		{
			TrendProfile: &utils.TrendProfile{
				ID:              "Trend2",
				StatID:          "Stats2",
				Tenant:          "cgrates.org",
				Schedule:        "12 30 * * *",
				MinItems:        1,
				TTL:             10 * time.Second,
				CorrelationType: "*average",
				Tolerance:       1.5,
				Stored:          false,
			},
		},
		{
			TrendProfile: &utils.TrendProfile{
				ID:              "Trend3",
				StatID:          "Stats3",
				Tenant:          "cgrates.org",
				Schedule:        "04 10 * * *",
				MinItems:        0,
				CorrelationType: "*last",
				Tolerance:       3,
				Stored:          true,
			},
		},
	}

	var reply string
	for _, trendProfile := range trendProfiles {
		if err := trRPC.Call(context.Background(), utils.AdminSv1SetTrendProfile,
			trendProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
	if err := trRPC.Call(context.Background(), utils.CacheSv1Clear,
		&utils.AttrCacheIDsWithAPIOpts{
			CacheIDs: nil,
		}, &reply); err != nil {
		t.Error(err)
	}
}

func testTrendsGetTrendProfileAfterSet(t *testing.T) {
	var replyTrendProfile *utils.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "Trend1",
			}}, &replyTrendProfile); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyTrendProfile, trendProfiles[0].TrendProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(trendProfiles[0].TrendProfile), utils.ToJSON(replyTrendProfile))
	}
}

func testTrendsGetTrendProfileIDsAfterSet(t *testing.T) {
	expectedIDs := []string{"Trend1", "Trend2", "Trend3"}
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

	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "Trend",
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
	} else if replyCount != 3 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "Trend",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 3 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "Trend1",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 1 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testTrendsGetTrendProfilesAfterSet(t *testing.T) {

	var replyTrendProfiles []*utils.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyTrendProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyTrendProfiles, func(i, j int) bool {
			return replyTrendProfiles[i].ID < replyTrendProfiles[j].ID
		})
		for i := range replyTrendProfiles {
			if !reflect.DeepEqual(replyTrendProfiles[i], trendProfiles[i].TrendProfile) {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>",
					utils.ToJSON(trendProfiles), utils.ToJSON(replyTrendProfiles))
			}
		}

	}
}

func testTrendsRemoveTrendProfile(t *testing.T) {
	var reply string
	if err := trRPC.Call(context.Background(), utils.AdminSv1RemoveTrendProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "Trend2",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testTrendsGetTrendProfileAfterRemove(t *testing.T) {
	var reply string
	if err := trRPC.Call(context.Background(), utils.CacheSv1Clear,
		&utils.AttrCacheIDsWithAPIOpts{
			CacheIDs: nil,
		}, &reply); err != nil {
		t.Error(err)
	}
	var replyTrendProfile utils.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "Trend2",
			}}, &replyTrendProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testTrendsGetTrendProfileIDsAfterRemove(t *testing.T) {
	expectedIDs := []string{"Trend1", "Trend3"}
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

	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "Trend",
		}, &replyTrendProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyTrendProfileIDs)
		if !slices.Equal(replyTrendProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyTrendProfileIDs)
		}
	}

	expectedIDs = []string{"Trend1"}
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "Trend1",
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
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 2, replyCount)
	}

	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "Trend",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 2, replyCount)
	}

}

func testTrendsGetTrendProfilesAfterRemove(t *testing.T) {
	expectedTrendProfiles := append(trendProfiles[:1], trendProfiles[2:]...)
	var replyTrendProfiles []*utils.TrendProfile
	if err := trRPC.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyTrendProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyTrendProfiles, func(i, j int) bool {
			return replyTrendProfiles[i].ID < replyTrendProfiles[j].ID
		})
		for i := range expectedTrendProfiles {
			if !reflect.DeepEqual(replyTrendProfiles[i], expectedTrendProfiles[i].TrendProfile) {
				t.Errorf("expected: <%+v>, \nreceived: <%+v>",
					utils.ToJSON(expectedTrendProfiles), utils.ToJSON(replyTrendProfiles))
			}
		}

	}
}

// Kill the engine when it is about to be finished
func testTrendsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
