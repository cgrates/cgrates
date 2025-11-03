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
	fltrCfgPath   string
	fltrCfg       *config.CGRConfig
	fltrRPC       *birpc.Client
	fltrConfigDIR string //run tests for specific configuration

	sTestsFltrs = []func(t *testing.T){
		testFiltersInitCfg,
		testFiltersInitDataDb,
		testFiltersStartEngine,
		testFiltersRPCConn,

		// tests for AdminSv1 APIs
		testFiltersGetFilterBeforeSet,
		testFiltersGetFilterIDsBeforeSet,
		testFiltersGetFilterCountBeforeSet,
		testFiltersGetFiltersBeforeSet,
		testFiltersSetFilters,
		testFiltersGetFilterAfterSet,
		testFiltersGetFilterIDsAfterSet,
		testFiltersGetFilterCountAfterSet,
		testFiltersGetFiltersAfterSet,
		testFiltersRemoveFilter,
		testFiltersGetFilterAfterRemove,
		testFiltersGetFilterIDsAfterRemove,
		testFiltersGetFilterCountAfterRemove,
		testFiltersGetFiltersAfterRemove,

		// tests for FiltersMatch API
		testFiltersSetInvalidFilter,
		testFiltersFiltersMatchTrue,
		testFiltersFiltersMatchFalse,
		testFiltersKillEngine,
	}
)

func TestFiltersIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		fltrConfigDIR = "tutinternal"
	case utils.MetaMongo:
		fltrConfigDIR = "tutmongo"
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		fltrConfigDIR = "tutmysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsFltrs {
		t.Run(fltrConfigDIR, stest)
	}
}

func testFiltersInitCfg(t *testing.T) {
	var err error
	fltrCfgPath = path.Join(*utils.DataDir, "conf", "samples", fltrConfigDIR)
	fltrCfg, err = config.NewCGRConfigFromPath(context.Background(), fltrCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testFiltersInitDataDb(t *testing.T) {
	if err := engine.InitDB(fltrCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testFiltersStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fltrCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testFiltersRPCConn(t *testing.T) {
	fltrRPC = engine.NewRPCClient(t, fltrCfg.ListenCfg(), *utils.Encoding)
}

func testFiltersGetFilterBeforeSet(t *testing.T) {
	var replyFilter engine.Filter
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_FILTER1",
			}}, &replyFilter); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFiltersGetFiltersBeforeSet(t *testing.T) {
	var replyFilters *[]*engine.Filter
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilters,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyFilters); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFiltersGetFilterIDsBeforeSet(t *testing.T) {
	var replyFilterIDs []string
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyFilterIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFiltersGetFilterCountBeforeSet(t *testing.T) {
	var replyCount int
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFiltersCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if replyCount != 0 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testFiltersSetFilters(t *testing.T) {
	filterProfiles := []*engine.FilterWithAPIOpts{
		{
			Filter: &engine.Filter{
				ID:     "TestA_FILTER1",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"1001"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"10"},
					},
				},
			},
		},
		{
			Filter: &engine.Filter{
				ID:     "TestA_FILTER2",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"1002"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"10"},
					},
				},
			},
		},
		{
			Filter: &engine.Filter{
				ID:     "TestA_FILTER3",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"1003"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"10"},
					},
				},
			},
		},
		{
			Filter: &engine.Filter{
				ID:     "TestB_FILTER1",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"2001"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"20"},
					},
				},
			},
		},
		{
			Filter: &engine.Filter{
				ID:     "TestB_FILTER2",
				Tenant: "cgrates.org",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Account",
						Values:  []string{"2002"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~*req.Destination",
						Values:  []string{"20"},
					},
				},
			},
		},
	}

	var reply string
	for _, filterProfile := range filterProfiles {
		if err := fltrRPC.Call(context.Background(), utils.AdminSv1SetFilter,
			filterProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testFiltersGetFilterAfterSet(t *testing.T) {
	expectedFilter := engine.Filter{
		ID:     "TestA_FILTER1",
		Tenant: "cgrates.org",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destination",
				Values:  []string{"10"},
			},
		},
	}
	var replyFilter engine.Filter
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_FILTER1",
			}}, &replyFilter); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyFilter, expectedFilter) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expectedFilter), utils.ToJSON(replyFilter))
	}
}

func testFiltersGetFilterIDsAfterSet(t *testing.T) {
	expectedIDs := []string{"TestA_FILTER1", "TestA_FILTER2", "TestA_FILTER3", "TestB_FILTER1", "TestB_FILTER2"}
	var replyFilterIDs []string
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyFilterIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyFilterIDs)
		if !slices.Equal(replyFilterIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyFilterIDs)
		}
	}

	expectedIDs = []string{"TestA_FILTER1", "TestA_FILTER2", "TestA_FILTER3"}
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyFilterIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyFilterIDs)
		if !slices.Equal(replyFilterIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyFilterIDs)
		}
	}

	expectedIDs = []string{"TestB_FILTER1", "TestB_FILTER2"}
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyFilterIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyFilterIDs)
		if !slices.Equal(replyFilterIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyFilterIDs)
		}
	}
}

func testFiltersGetFilterCountAfterSet(t *testing.T) {
	var replyCount int
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFiltersCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 5 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFiltersCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 3 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFiltersCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testFiltersGetFiltersAfterSet(t *testing.T) {
	expectedFilters := []*engine.Filter{
		{
			ID:     "TestA_FILTER1",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"10"},
				},
			},
		},
		{
			ID:     "TestA_FILTER2",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1002"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"10"},
				},
			},
		},
		{
			ID:     "TestA_FILTER3",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1003"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"10"},
				},
			},
		},
		{
			ID:     "TestB_FILTER1",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"2001"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"20"},
				},
			},
		},
		{
			ID:     "TestB_FILTER2",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"2002"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"20"},
				},
			},
		},
	}
	var replyFilters []*engine.Filter
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilters,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyFilters); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyFilters, func(i, j int) bool {
			return replyFilters[i].ID < replyFilters[j].ID
		})
		if !reflect.DeepEqual(replyFilters, expectedFilters) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedFilters), utils.ToJSON(replyFilters))
		}
	}
}

func testFiltersRemoveFilter(t *testing.T) {
	var reply string
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1RemoveFilter,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_FILTER2",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testFiltersGetFilterAfterRemove(t *testing.T) {
	var replyFilter engine.Filter
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Filter2",
			}}, &replyFilter); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFiltersGetFilterIDsAfterRemove(t *testing.T) {
	expectedIDs := []string{"TestA_FILTER1", "TestA_FILTER3", "TestB_FILTER1", "TestB_FILTER2"}
	var replyFilterIDs []string
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyFilterIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyFilterIDs)
		if !slices.Equal(replyFilterIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyFilterIDs)
		}
	}

	expectedIDs = []string{"TestA_FILTER1", "TestA_FILTER3"}
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyFilterIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyFilterIDs)
		if !slices.Equal(replyFilterIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyFilterIDs)
		}
	}

	expectedIDs = []string{"TestB_FILTER1", "TestB_FILTER2"}
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyFilterIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyFilterIDs)
		if !slices.Equal(replyFilterIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyFilterIDs)
		}
	}
}

func testFiltersGetFilterCountAfterRemove(t *testing.T) {
	var replyCount int
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFiltersCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 4 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFiltersCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFiltersCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testFiltersGetFiltersAfterRemove(t *testing.T) {
	expectedFilters := []*engine.Filter{
		{
			ID:     "TestA_FILTER1",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"10"},
				},
			},
		},
		{
			ID:     "TestA_FILTER3",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1003"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"10"},
				},
			},
		},
		{
			ID:     "TestB_FILTER1",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"2001"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"20"},
				},
			},
		},
		{
			ID:     "TestB_FILTER2",
			Tenant: "cgrates.org",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"2002"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"20"},
				},
			},
		},
	}
	var replyFilters []*engine.Filter
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilters,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyFilters); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyFilters, func(i, j int) bool {
			return replyFilters[i].ID < replyFilters[j].ID
		})
		if !reflect.DeepEqual(replyFilters, expectedFilters) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedFilters), utils.ToJSON(replyFilters))
		}
	}
}

func testFiltersSetInvalidFilter(t *testing.T) {
	fltrPrf := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "invalid_filter",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "",
					Values:  []string{},
				},
			},
		},
	}
	experr := `SERVER_ERROR: there exists at least one filter rule that is not valid`
	var reply string
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrPrf, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	var result *engine.Filter
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "invalid_filter",
		}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFiltersFiltersMatchTrue(t *testing.T) {
	args := &engine.ArgsFiltersMatch{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
			},
		},
		FilterIDs: []string{
			"*string:~*req.Account:1001",
			"*prefix:~*req.Destination:10",
		},
	}

	var reply bool
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1FiltersMatch, args,
		&reply); err != nil {
		t.Error(err)
	} else if reply != true {
		t.Error("expected reply to be", true)
	}
}

func testFiltersFiltersMatchFalse(t *testing.T) {
	args := &engine.ArgsFiltersMatch{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventTest",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "2002",
			},
		},
		FilterIDs: []string{
			"*string:~*req.Account:1001",
			"*prefix:~*req.Destination:10",
		},
	}

	var reply bool
	if err := fltrRPC.Call(context.Background(), utils.AdminSv1FiltersMatch, args,
		&reply); err != nil {
		t.Error(err)
	} else if reply != false {
		t.Error("expected reply to be", false)
	}
}

// Kill the engine when it is about to be finished
func testFiltersKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
