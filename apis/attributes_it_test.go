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
	"fmt"
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
	attrCfgPath   string
	attrCfg       *config.CGRConfig
	attrRPC       *birpc.Client
	attrConfigDIR string //run tests for specific configuration

	sTestsAttr = []func(t *testing.T){
		testAttributesInitCfg,
		testAttributesInitDataDb,
		testAttributeSResetStorDb,
		testAttributesStartEngine,
		testAttributesRPCConn,

		// tests for AdminSv1 APIs
		testAttributesGetAttributeProfileBeforeSet,
		testAttributesGetAttributeProfileIDsBeforeSet,
		testAttributesGetAttributeProfileCountBeforeSet,
		testAttributesGetAttributeProfilesBeforeSet,
		testAttributesSetAttributeProfiles,
		testAttributesGetAttributeProfileAfterSet,
		testAttributesGetAttributeProfileIDsAfterSet,
		testAttributesGetAttributeProfileCountAfterSet,
		testAttributesGetAttributeProfilesAfterSet,
		testAttributesRemoveAttributeProfile,
		testAttributesGetAttributeProfileAfterRemove,
		testAttributesGetAttributeProfileIDsAfterRemove,
		testAttributesGetAttributeProfileCountAfterRemove,
		testAttributesGetAttributeProfilesAfterRemove,

		//tests for AttributeSv1 APIs
		testAttributeSetAttributeProfileBrokenReference,
		testAttributeSGetAttributeForEventMissingEvent,
		testAttributeSGetAttributeForEventAnyContext,
		testAttributeSGetAttributeForEventSameAnyContext,
		testAttributeSGetAttributeForEventNotFound,
		testAttributeSGetAttributeForEvent,
		testAttributeSetAttributeProfile,
		testAttributeProcessEvent,
		testAttributeProcessEventWithSearchAndReplace,
		testAttributeSProcessWithMultipleRuns,
		testAttributeSProcessWithMultipleRuns2,
		testAttributeGetAttributeProfileAllIDs,
		testAttributeGetAttributeProfileAllCount,
		testAttributeRemoveRemainAttributeProfiles,
		testAttributeGetAttributeProfileAfterRemove,

		// tests for attribute blocker behaviour
		testAttributeSetAttributeProfileWithAttrBlockers,
		testAttributeSetAttributeProfileWithAttrBlockers2,
		testAttributeSetAttributeProfileBlockersBothProfilesProcessRuns,

		// Testing index behaviour
		testAttributeSSetNonIndexedTypeFilter,
		testAttributeSSetIndexedTypeFilter,
		testAttributeSClearIndexes,
		testAttributeSCheckIndexesSetAttributeProfileWithoutFilters,

		// testAttributeSCheckIndexesAddNonIndexedFilter,
		testAttributeSCheckIndexesAddIndexedFilters,
		testAttributeSCheckIndexesModifyIndexedFilter,
		testAttributeSCheckIndexesRemoveAnIndexedFilter,
		testAttributeSCheckIndexesRemoveAttributeProfile,

		testAttributesKillEngine,
	}
)

func TestAttributesIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		attrConfigDIR = "attributes_internal"
	case utils.MetaMongo:
		attrConfigDIR = "attributes_mongo"
	case utils.MetaMySQL:
		attrConfigDIR = "attributes_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAttr {
		t.Run(attrConfigDIR, stest)
	}
}

func testAttributesInitCfg(t *testing.T) {
	var err error
	attrCfgPath = path.Join(*utils.DataDir, "conf", "samples", attrConfigDIR)
	attrCfg, err = config.NewCGRConfigFromPath(context.Background(), attrCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAttributesInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(attrCfg); err != nil {
		t.Fatal(err)
	}
}

func testAttributeSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(attrCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributesStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(attrCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testAttributesRPCConn(t *testing.T) {
	attrRPC = engine.NewRPCClient(t, attrCfg.ListenCfg(), *utils.Encoding)
}

func testAttributesGetAttributeProfileBeforeSet(t *testing.T) {
	var replyAttributeProfile engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_ATTRIBUTE1",
			}}, &replyAttributeProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testAttributesGetAttributeProfilesBeforeSet(t *testing.T) {
	var replyAttributeProfiles *[]*engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyAttributeProfiles); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testAttributesGetAttributeProfileIDsBeforeSet(t *testing.T) {
	var replyAttributeProfileIDs []string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyAttributeProfileIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testAttributesGetAttributeProfileCountBeforeSet(t *testing.T) {
	var replyCount int
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if replyCount != 0 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testAttributesSetAttributeProfiles(t *testing.T) {
	attributeProfiles := []*engine.APIAttributeProfileWithAPIOpts{
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "TestA_ATTRIBUTE1",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: false,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						Path:      "*req.Account",
						Type:      utils.MetaConstant,
						Value:     "1001",
						FilterIDs: []string{"fltr1"},
					},
				},
			},
		},
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "TestA_ATTRIBUTE2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: true,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						Path:      "*req.Account",
						Type:      utils.MetaConstant,
						Value:     "1002",
						FilterIDs: []string{"fltr2"},
					},
				},
			},
		},
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "TestA_ATTRIBUTE3",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						Path:      "*req.Account",
						Type:      utils.MetaConstant,
						Value:     "1003",
						FilterIDs: []string{"fltr3"},
					},
				},
			},
		},
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "TestB_ATTRIBUTE1",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						Path:      "*req.Account",
						Type:      utils.MetaConstant,
						Value:     "2001",
						FilterIDs: []string{"fltr4"},
					},
				},
			},
		},
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "TestB_ATTRIBUTE2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						Path:      "*req.Account",
						Type:      utils.MetaConstant,
						Value:     "2002",
						FilterIDs: []string{"fltr5"},
					},
				},
			},
		},
	}

	var reply string
	for _, attributeProfile := range attributeProfiles {
		if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
			attributeProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testAttributesGetAttributeProfileAfterSet(t *testing.T) {
	expectedAttributeProfile := engine.APIAttributeProfile{
		ID:        "TestA_ATTRIBUTE1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:      "*req.Account",
				Type:      utils.MetaConstant,
				Value:     "1001",
				FilterIDs: []string{"fltr1"},
			},
		},
	}
	var replyAttributeProfile engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_ATTRIBUTE1",
			}}, &replyAttributeProfile); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyAttributeProfile, expectedAttributeProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expectedAttributeProfile), utils.ToJSON(replyAttributeProfile))
	}
}

func testAttributesGetAttributeProfileIDsAfterSet(t *testing.T) {
	expectedIDs := []string{"TestA_ATTRIBUTE1", "TestA_ATTRIBUTE2", "TestA_ATTRIBUTE3", "TestB_ATTRIBUTE1", "TestB_ATTRIBUTE2"}
	var replyAttributeProfileIDs []string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyAttributeProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyAttributeProfileIDs)
		if !slices.Equal(replyAttributeProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyAttributeProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_ATTRIBUTE1", "TestA_ATTRIBUTE2", "TestA_ATTRIBUTE3"}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyAttributeProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyAttributeProfileIDs)
		if !slices.Equal(replyAttributeProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyAttributeProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_ATTRIBUTE1", "TestB_ATTRIBUTE2"}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyAttributeProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyAttributeProfileIDs)
		if !slices.Equal(replyAttributeProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyAttributeProfileIDs)
		}
	}
}

func testAttributesGetAttributeProfileCountAfterSet(t *testing.T) {
	var replyCount int
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 5 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 3 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testAttributesGetAttributeProfilesAfterSet(t *testing.T) {
	expectedAttributeProfiles := []*engine.APIAttributeProfile{
		{
			ID:        "TestA_ATTRIBUTE1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      "*req.Account",
					Type:      utils.MetaConstant,
					Value:     "1001",
					FilterIDs: []string{"fltr1"},
				},
			},
		},
		{
			ID:        "TestA_ATTRIBUTE2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      "*req.Account",
					Type:      utils.MetaConstant,
					Value:     "1002",
					FilterIDs: []string{"fltr2"},
				},
			},
		},
		{
			ID:        "TestA_ATTRIBUTE3",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      "*req.Account",
					Type:      utils.MetaConstant,
					Value:     "1003",
					FilterIDs: []string{"fltr3"},
				},
			},
		},
		{
			ID:        "TestB_ATTRIBUTE1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 5,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      "*req.Account",
					Type:      utils.MetaConstant,
					Value:     "2001",
					FilterIDs: []string{"fltr4"},
				},
			},
		},
		{
			ID:        "TestB_ATTRIBUTE2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 25,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      "*req.Account",
					Type:      utils.MetaConstant,
					Value:     "2002",
					FilterIDs: []string{"fltr5"},
				},
			},
		},
	}
	var replyAttributeProfiles []*engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyAttributeProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyAttributeProfiles, func(i, j int) bool {
			return replyAttributeProfiles[i].ID < replyAttributeProfiles[j].ID
		})
		if !reflect.DeepEqual(replyAttributeProfiles, expectedAttributeProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedAttributeProfiles), utils.ToJSON(replyAttributeProfiles))
		}
	}
}

func testAttributesRemoveAttributeProfile(t *testing.T) {
	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_ATTRIBUTE2",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testAttributesGetAttributeProfileAfterRemove(t *testing.T) {
	var replyAttributeProfile engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Attribute2",
			}}, &replyAttributeProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testAttributesGetAttributeProfileIDsAfterRemove(t *testing.T) {
	expectedIDs := []string{"TestA_ATTRIBUTE1", "TestA_ATTRIBUTE3", "TestB_ATTRIBUTE1", "TestB_ATTRIBUTE2"}
	var replyAttributeProfileIDs []string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyAttributeProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyAttributeProfileIDs)
		if !slices.Equal(replyAttributeProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyAttributeProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_ATTRIBUTE1", "TestA_ATTRIBUTE3"}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyAttributeProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyAttributeProfileIDs)
		if !slices.Equal(replyAttributeProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyAttributeProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_ATTRIBUTE1", "TestB_ATTRIBUTE2"}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyAttributeProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyAttributeProfileIDs)
		if !slices.Equal(replyAttributeProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyAttributeProfileIDs)
		}
	}
}

func testAttributesGetAttributeProfileCountAfterRemove(t *testing.T) {
	var replyCount int
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 4 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testAttributesGetAttributeProfilesAfterRemove(t *testing.T) {
	expectedAttributeProfiles := []*engine.APIAttributeProfile{
		{
			ID:        "TestA_ATTRIBUTE1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      "*req.Account",
					Type:      utils.MetaConstant,
					Value:     "1001",
					FilterIDs: []string{"fltr1"},
				},
			},
		},
		{
			ID:        "TestA_ATTRIBUTE3",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      "*req.Account",
					Type:      utils.MetaConstant,
					Value:     "1003",
					FilterIDs: []string{"fltr3"},
				},
			},
		},
		{
			ID:        "TestB_ATTRIBUTE1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 5,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      "*req.Account",
					Type:      utils.MetaConstant,
					Value:     "2001",
					FilterIDs: []string{"fltr4"},
				},
			},
		},
		{
			ID:        "TestB_ATTRIBUTE2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 25,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      "*req.Account",
					Type:      utils.MetaConstant,
					Value:     "2002",
					FilterIDs: []string{"fltr5"},
				},
			},
		},
	}
	var replyAttributeProfiles []*engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyAttributeProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyAttributeProfiles, func(i, j int) bool {
			return replyAttributeProfiles[i].ID < replyAttributeProfiles[j].ID
		})
		if !reflect.DeepEqual(replyAttributeProfiles, expectedAttributeProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedAttributeProfiles), utils.ToJSON(replyAttributeProfiles))
		}
	}
}

func testAttributeSetAttributeProfileBrokenReference(t *testing.T) {
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST_SECOND",
			FilterIDs: []string{"invalid_filter_format", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
	}
	var reply string
	expectedErr := "SERVER_ERROR: broken reference to filter: <invalid_filter_format> for item with ID: cgrates.org:TEST_ATTRIBUTES_IT_TEST_SECOND"
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v \n, received %+v", expectedErr, err)
	}
}

func testAttributeSGetAttributeForEventMissingEvent(t *testing.T) {
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		nil, &rplyEv); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSGetAttributeForEventAnyContext(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
		Event: map[string]any{
			utils.AccountField: "dan",
			utils.Destination:  "+491511231234",
			utils.ToR:          "*voice",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:   10 * time.Second,
			utils.OptsContext: utils.MetaCDRs,
		},
	}
	eAttrPrf2 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    ev.Tenant,
			ID:        "ATTR_2",
			FilterIDs: []string{"*string:~*req.Account:dan", "*prefix:~*req.Destination:+4915", "*exists:~*opts.*usage:", "*notexists:~*req.RequestType:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1001",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var result string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		eAttrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_2"}}, &reply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eAttrPrf2.APIAttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", eAttrPrf2.APIAttributeProfile, reply)
	}
	var attrReply *engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Fatal(err)
	}

	expAttrFromEv := &engine.APIAttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:dan", "*prefix:~*req.Destination:+4915", "*exists:~*opts.*usage:", "*notexists:~*req.RequestType:"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:  utils.MetaConstant,
				Value: "1001",
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if !reflect.DeepEqual(expAttrFromEv, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(expAttrFromEv), utils.ToJSON(attrReply))
	}
}

func testAttributeSGetAttributeForEventSameAnyContext(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
		Event: map[string]any{
			utils.AccountField: "dan",
			utils.Destination:  "+491511231234",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: 10 * time.Second,
		},
	}
	var attrReply *engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Fatal(err)
	}

	expAttrFromEv := &engine.APIAttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:dan", "*prefix:~*req.Destination:+4915", "*exists:~*opts.*usage:", "*notexists:~*req.RequestType:"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:  utils.MetaConstant,
				Value: "1001",
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if !reflect.DeepEqual(expAttrFromEv, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(expAttrFromEv), utils.ToJSON(attrReply))
	}
}

func testAttributeSGetAttributeForEventNotFound(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
		Event: map[string]any{
			utils.AccountField: "dann",
			utils.Destination:  "+491511231234",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaCDRs,
		},
	}
	var attrReply *engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSGetAttributeForEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSGetAttributeForEvent",
		Event: map[string]any{
			utils.AccountField: "1007",
			utils.Destination:  "+491511231234",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:   10 * time.Second,
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eAttrPrf := &engine.APIAttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1007", "*string:~*opts.*context:*cdrs|*sessions", "*exists:~*opts.*usage:"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:      utils.MetaReq + utils.NestingSep + utils.AccountField,
				Value:     "1001",
				Type:      utils.MetaConstant,
				FilterIDs: []string{},
			},
			{
				Path:      utils.MetaReq + utils.NestingSep + utils.Subject,
				Value:     "1001",
				Type:      utils.MetaConstant,
				FilterIDs: []string{},
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if *utils.Encoding == utils.MetaGOB {
		eAttrPrf.Attributes[0].FilterIDs = nil
		eAttrPrf.Attributes[1].FilterIDs = nil
	}
	var result string
	eAttrPrfApi := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    ev.Tenant,
			ID:        "ATTR_1",
			FilterIDs: []string{"*string:~*req.Account:1007", "*string:~*opts.*context:*cdrs|*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value:     "1001",
					Type:      utils.MetaConstant,
					FilterIDs: []string{},
				},
				{
					Path:      utils.MetaReq + utils.NestingSep + utils.Subject,
					Value:     "1001",
					Type:      utils.MetaConstant,
					FilterIDs: []string{},
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		eAttrPrfApi, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var attrReply *engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	ev.Tenant = utils.EmptyString
	ev.ID = "randomID"
	var attrPrf *engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrPrf); err != nil {
		t.Fatal(err)
	}

	// Populate private variables in RSRParsers
	if !reflect.DeepEqual(eAttrPrf, attrPrf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(attrPrf))
	}
}

func testAttributeSetAttributeProfile(t *testing.T) {
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAttr := engine.APIAttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_IT_TEST",
		FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: "1002",
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: "cgrates.itsyscom",
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	var result engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAttr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedAttr), utils.ToJSON(result))
	}
}

func testAttributeProcessEvent(t *testing.T) {
	var reply *engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1"}}, &reply); err != nil {
		t.Fatal(err)
	}
	args := &utils.CGREvent{
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.AccountField: "1002",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:   10 * time.Second,
			utils.OptsContext: utils.MetaCDRs,
		},
	}
	expEvReply := &engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:TEST_ATTRIBUTES_IT_TEST",
				Fields:           []string{"*tenant", utils.AccountField},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.itsyscom",
			Event: map[string]any{
				utils.AccountField: "1002",
				utils.ToR:          utils.MetaVoice,
			},
			APIOpts: map[string]any{
				utils.MetaUsage:   float64(10 * time.Second),
				utils.OptsContext: utils.MetaCDRs,
			},
		},
	}
	evRply := &engine.AttrSProcessEventReply{}
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		args, &evRply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expEvReply.AlteredFields[0].Fields)
		sort.Strings(evRply.AlteredFields[0].Fields)
		if !reflect.DeepEqual(evRply, expEvReply) {
			t.Errorf("Expected %+v, received %+v", expEvReply, evRply)
		}
	}
}

func testAttributeProcessEventWithSearchAndReplace(t *testing.T) {
	attrPrf1 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_Search_and_replace",
			FilterIDs: []string{"*string:~*req.Category:call", "*string:~*opts.*context:*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Category",
					Value: "~*req.Category:s/(.*)/${1}_suffix/",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var result string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "HeaderEventForAttribute",
		Event: map[string]any{
			"Category": "call",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:   10 * time.Second,
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_Search_and_replace",
				Fields:           []string{"*req.Category"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "HeaderEventForAttribute",
			Event: map[string]any{
				"Category": "call_suffix",
			},
			APIOpts: map[string]any{
				utils.MetaUsage:   float64(10 * time.Second),
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var rplyEv *engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessWithMultipleRuns(t *testing.T) {
	attrPrf1 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*string:~*opts.*context:*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: "Value1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	attrPrf2 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_2",
			FilterIDs: []string{"*string:~*req.Field1:Value1", "*string:~*opts.*context:*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Value: "Value2",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	attrPrf3 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_3",
			FilterIDs: []string{"*string:~*req.NotFound:NotFound", "*string:~*opts.*context:*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field3",
					Value: "Value3",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	// Add attribute in DM
	var result string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                 10 * time.Second,
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{"*req.Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{"*req.Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{"*req.Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{"*req.Field2"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, rplyEv.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, rplyEv.AlteredFields)
	} else if !reflect.DeepEqual(eRply.CGREvent.Event, rplyEv.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, rplyEv.CGREvent.Event)
	}
}

func testAttributeSProcessWithMultipleRuns2(t *testing.T) {
	attrPrf1 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*string:~*opts.*context:*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: "Value1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	attrPrf2 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_2",
			FilterIDs: []string{"*string:~*req.Field1:Value1", "*string:~*opts.*context:*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Value: "Value2",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 25,
				},
			},
		},
	}
	attrPrf3 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_3",
			FilterIDs: []string{"*string:~*req.Field2:Value2", "*string:~*opts.*context:*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field3",
					Value: "Value3",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
		},
	}
	// Add attributeProfiles
	var result string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                 10 * time.Second,
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_1",
				Fields:           []string{"*req.Field1"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{"*req.Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_2",
				Fields:           []string{"*req.Field2"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_3",
				Fields:           []string{"*req.Field3"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
				"Field3":       "Value3",
			},
			APIOpts: map[string]any{
				utils.MetaUsage:                 10 * time.Second,
				utils.OptsContext:               utils.MetaSessionS,
				utils.OptsAttributesProcessRuns: 4,
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Slice(rplyEv.AlteredFields, func(i, j int) bool {
		return rplyEv.AlteredFields[i].MatchedProfileID < rplyEv.AlteredFields[j].MatchedProfileID
	})
	if !reflect.DeepEqual(eRply.AlteredFields, rplyEv.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eRply.AlteredFields), utils.ToJSON(rplyEv.AlteredFields))
	} else if !reflect.DeepEqual(eRply.CGREvent.Event, rplyEv.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, rplyEv.CGREvent.Event)
	}
}

func testAttributeGetAttributeProfileAllIDs(t *testing.T) {
	var rply []string
	expectedIds := []string{"ATTR_1", "ATTR_2", "ATTR_3", "ATTR_Search_and_replace", "TEST_ATTRIBUTES_IT_TEST", "TestA_ATTRIBUTE1",
		"TestA_ATTRIBUTE3", "TestB_ATTRIBUTE1", "TestB_ATTRIBUTE2"} //"TEST_ATTRIBUTES_IT_TEST_SECOND"}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &rply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(rply)
		sort.Strings(expectedIds)
		if !reflect.DeepEqual(expectedIds, rply) {
			t.Errorf("Expected %+v, received %+v", expectedIds, rply)
		}
	}
}

func testAttributeGetAttributeProfileAllCount(t *testing.T) {
	var rply int
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfilesCount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
			},
		}, &rply); err != nil {
		t.Error(err)
	} else if rply != 9 {
		t.Errorf("Expected %+v, received %+v", 9, rply)
	}
}

func testAttributeRemoveRemainAttributeProfiles(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "ATTR_1",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "ATTR_2",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "ATTR_3",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "ATTR_Search_and_replace",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
}

func testAttributeGetAttributeProfileAfterRemove(t *testing.T) {
	var reply *engine.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "ATTR_Search_and_replace",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "ATTR_1",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "ATTR_2",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "ATTR_3",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSetAttributeProfileWithAttrBlockers(t *testing.T) {
	// the blocker on the profile is false
	attrPrf1 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_WITH_BLOCKER_TRUE",
			FilterIDs: []string{"*string:~*req.Blockers:*exists", "*eq:~*opts.*attrProcessRuns:2"},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "ToR",
					Value: "*sms",
				},
				{
					Path:  utils.MetaOpts + utils.NestingSep + "*chargers",
					Value: "true",
				},
				{
					Blockers: utils.DynamicBlockers{
						{
							FilterIDs: []string{"*prefix:~*req.Destination:4433"},
							Blocker:   true,
						},
					},
					Path:  utils.MetaReq + utils.NestingSep + "RequestType",
					Value: "*rated",
				},
				{
					Path:  utils.MetaOpts + utils.NestingSep + "*usage",
					Value: "1m",
				},
			},
		},
	}
	// here the the blocker on the profile is true
	attrPrf2 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "ATTR_WITH_BLOCKER",
			FilterIDs: []string{"*string:~*req.Blockers:*exists",
				"*notexists:~*opts.*usage:"},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Blockers: utils.DynamicBlockers{
						{
							FilterIDs: []string{"*prefix:~*req.Destination:4433"},
							Blocker:   true,
						},
					},
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Value: "10093",
				},
				{
					Path:  utils.MetaOpts + utils.NestingSep + "*rates",
					Value: "true",
				},
			},
		},
	}
	// Add attributeProfiles
	var result string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// first we will process the second attribute with true BLOCKER on profile, but true on Attributes if the Destination prefix is 4433(FOR NOW THE DESTINATION WILL BE EMPTY)
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.AccountField: "1002",
			"Blockers":         "*exists",
		},
		APIOpts: map[string]any{},
	}
	expEvReply := &engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_WITH_BLOCKER",
				Fields:           []string{"*opts.*rates", "*req.Account"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "10093",
				"Blockers":         "*exists",
			},
			APIOpts: map[string]any{
				utils.MetaRates: "true",
			},
		},
	}
	evRply := &engine.AttrSProcessEventReply{}
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		args, &evRply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expEvReply.AlteredFields[0].Fields)
		sort.Strings(evRply.AlteredFields[0].Fields)
		if !reflect.DeepEqual(evRply, expEvReply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expEvReply), utils.ToJSON(evRply))
		}
	}
}

func testAttributeSetAttributeProfileWithAttrBlockers2(t *testing.T) {
	// first we will process the second attribute with true BLOCKER on profile, but true on Attributes if the Destination prefix is 4433(NOW WE WILL POPULATE THE DESTINATION, AND THE BLOCKER WILL STOP THE NEXT ATTRIBUTES)
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.AccountField: "1002",
			"Blockers":         "*exists",
			utils.Destination:  "4433254",
		},
		APIOpts: map[string]any{},
	}
	expEvReply := &engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_WITH_BLOCKER",
				Fields:           []string{"*req.Account"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "10093",
				"Blockers":         "*exists",
				utils.Destination:  "4433254",
			},
			// now *rates was not processde ebcause the blocker amtched the filter of Destination
			APIOpts: map[string]any{},
		},
	}
	evRply := &engine.AttrSProcessEventReply{}
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		args, &evRply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expEvReply.AlteredFields[0].Fields)
		sort.Strings(evRply.AlteredFields[0].Fields)
		if !reflect.DeepEqual(evRply, expEvReply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expEvReply), utils.ToJSON(evRply))
		}
	}
}

func testAttributeSetAttributeProfileBlockersBothProfilesProcessRuns(t *testing.T) {
	// now we will process both attributes that matched, but blokcers on the attributes will be felt and not all attributes from the list were processed
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.AccountField: "1002",
			"Blockers":         "*exists",
			utils.Destination:  "4433254",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	expEvReply := &engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_WITH_BLOCKER_TRUE",
				Fields:           []string{"*opts.*chargers", "*req.RequestType", "*req.ToR"},
			},
			{
				MatchedProfileID: "cgrates.org:ATTR_WITH_BLOCKER",
				Fields:           []string{"*req.Account"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.ToR:          "*sms",
				utils.AccountField: "10093",
				"Blockers":         "*exists",
				utils.Destination:  "4433254",
				utils.RequestType:  "*rated",
			},
			// now *rates was not processde ebcause the blocker amtched the filter of Destination
			APIOpts: map[string]any{
				utils.OptsAttributesProcessRuns: 2.,
				utils.MetaChargers:              "true",
			},
		},
	}
	evRply := &engine.AttrSProcessEventReply{}
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		args, &evRply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expEvReply.AlteredFields[0].Fields)
		sort.Strings(evRply.AlteredFields[0].Fields)
		if !reflect.DeepEqual(evRply, expEvReply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expEvReply), utils.ToJSON(evRply))
		}
	}
}

func testAttributeSSetNonIndexedTypeFilter(t *testing.T) {
	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "NONINDEXED_FLTR_TYPE",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaGreaterThan,
					Element: utils.Cost,
					Values:  []string{"10"},
				},
			},
		},
	}
	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetFilter, filter,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	args := &utils.CGREvent{
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.AccountField: "1002",
			"Blockers":         "*exists",
			utils.Destination:  "44322",
		},
		APIOpts: map[string]any{},
	}
	expEvReply := &engine.AttrSProcessEventReply{
		AlteredFields: []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_WITH_BLOCKER",
				Fields:           []string{"*opts.*rates", "*req.Account"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "10093",
				"Blockers":         "*exists",
				utils.Destination:  "44322",
			},
			APIOpts: map[string]any{
				utils.MetaRates: "true",
			},
		},
	}
	evRply := &engine.AttrSProcessEventReply{}
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		args, &evRply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expEvReply.AlteredFields[0].Fields)
		sort.Strings(evRply.AlteredFields[0].Fields)
		if !reflect.DeepEqual(evRply, expEvReply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expEvReply), utils.ToJSON(evRply))
		}
	}
}

func testAttributeSSetIndexedTypeFilter(t *testing.T) {
	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "INDEXED_FLTR_TYPE",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Account",
					Values:  []string{"10"},
				},
			},
		},
	}
	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetFilter, filter,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply: ", reply)
	}
}

func testAttributeSClearIndexes(t *testing.T) {
	args := &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
	}
	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply: ", reply)
	}
}

func testAttributeSCheckIndexesSetAttributeProfileWithoutFilters(t *testing.T) {
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant: "cgrates.org",
			ID:     "ATTR_TEST",
			Attributes: []*engine.ExternalAttribute{
				{
					Type:  utils.MetaConstant,
					Path:  "~*req.Account",
					Value: "1002",
				},
			},
		},
	}

	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expIdx := []string{"*none:*any:*any:ATTR_TEST"}
	args := &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
	}
	var replyIdx []string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		args, &replyIdx); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyIdx, expIdx) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
	}
}

// func testAttributeSCheckIndexesAddNonIndexedFilter(t *testing.T) {
// 	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
// 		APIAttributeProfile: &engine.APIAttributeProfile{
// 			Tenant:    "cgrates.org",
// 			ID:        "ATTR_TEST",
// 			FilterIDs: []string{"NONINDEXED_FLTR_TYPE"},
// 			Attributes: []*engine.ExternalAttribute{
// 				{
// 					Type:  utils.MetaConstant,
// 					Path:  "~*req.Account",
// 					Value: "1002",
// 				},
// 			},
// 		},
// 	}

// 	var reply string
// 	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
// 		attrPrf, &reply); err != nil {
// 		t.Error(err)
// 	} else if reply != utils.OK {
// 		t.Error(err)
// 	}

// 	expIdx := []string{"*none:*any:*any:ATTR_TEST"}
// 	args := &AttrGetFilterIndexes{
// 		Tenant:   "cgrates.org",
// 		ItemType: utils.MetaAttributes,
// 	}
// 	var replyIdx []string
// 	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
// 		args, &replyIdx); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(replyIdx, expIdx) {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
// 	}
// }

func testAttributeSCheckIndexesAddIndexedFilters(t *testing.T) {
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_TEST",
			FilterIDs: []string{"NONINDEXED_FLTR_TYPE", "INDEXED_FLTR_TYPE"},
			Attributes: []*engine.ExternalAttribute{
				{
					Type:  utils.MetaConstant,
					Path:  "~*req.Account",
					Value: "1002",
				},
			},
		},
	}

	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expIdx := []string{"*prefix:*req.Account:10:ATTR_TEST"}
	args := &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
	}
	var replyIdx []string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		args, &replyIdx); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyIdx, expIdx) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
	}

	attrPrf = &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_TEST",
			FilterIDs: []string{"NONINDEXED_FLTR_TYPE", "INDEXED_FLTR_TYPE", "*string:~*req.Category:call"},
			Attributes: []*engine.ExternalAttribute{
				{
					Type:  utils.MetaConstant,
					Path:  "~*req.Account",
					Value: "1002",
				},
			},
		},
	}

	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expIdx = []string{"*prefix:*req.Account:10:ATTR_TEST", "*string:*req.Category:call:ATTR_TEST"}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		args, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(replyIdx, expIdx) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
		}
	}
}

func testAttributeSCheckIndexesModifyIndexedFilter(t *testing.T) {
	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "INDEXED_FLTR_TYPE",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaSuffix,
					Element: "~*req.Subject",
					Values:  []string{"01"},
				},
			},
		},
	}
	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetFilter, filter,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply: ", reply)
	}

	expIdx := []string{"*string:*req.Category:call:ATTR_TEST", "*suffix:*req.Subject:01:ATTR_TEST"}
	args := &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
	}
	var replyIdx []string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		args, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(replyIdx, expIdx) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
		}
	}
}

func testAttributeSCheckIndexesRemoveAnIndexedFilter(t *testing.T) {
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_TEST",
			FilterIDs: []string{"NONINDEXED_FLTR_TYPE", "INDEXED_FLTR_TYPE"},
			Attributes: []*engine.ExternalAttribute{
				{
					Type:  utils.MetaConstant,
					Path:  "~*req.Account",
					Value: "1002",
				},
			},
		},
	}

	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expIdx := []string{"*suffix:*req.Subject:01:ATTR_TEST"}
	args := &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
	}
	var replyIdx []string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		args, &replyIdx); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyIdx, expIdx) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
	}
}
func testAttributeSCheckIndexesRemoveAttributeProfile(t *testing.T) {
	argsRem := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ATTR_TEST",
		},
	}
	var reply string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		argsRem, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	args := &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
	}
	var replyIdx []string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		args, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

// Kill the engine when it is about to be finished
func testAttributesKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func TestAttributesArith(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	tpInDir := t.TempDir()

	content := fmt.Sprintf(`{

"general": {
	"log_level": 7
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"attributes": {
	"enabled": true
},
"admins": {
	"enabled": true
},
"loaders": [
	{
		"id": "*default",
		"enabled": true,
		"run_delay": "-1",
		"tp_in_dir": "%s",
		"tp_out_dir": "",
		"action": "*store",
		"opts": {
			"*stopOnError": true
		}
	}
]
}`, tpInDir)

	tpFiles := map[string]string{
		utils.AttributesCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,AttributeFilterIDs,AttributeBlockers,Path,Type,Value
cgrates.org,ATTR_ARITH,*string:~*req.AttrSource:csv,,,,,,,
cgrates.org,ATTR_ARITH,,,,,,*req.3*4,*multiply,3;4
cgrates.org,ATTR_ARITH,,,,,,*req.12/4,*divide,12;4
cgrates.org,ATTR_ARITH,,,,,,*req.3+4,*sum,3;4
cgrates.org,ATTR_ARITH,,,,,,*req.3-4,*difference,3;4
cgrates.org,ATTR_ARITH,,,,,,*req.MultiplyBetweenVariables,*multiply,~*req.Elem1;~*req.Elem2`,
	}

	testEnv := engine.TestEnvironment{
		ConfigJSON: content,
		TpPath:     tpInDir,
		TpFiles:    tpFiles,
		Encoding:   *utils.Encoding,
	}
	client, _ := testEnv.Setup(t, context.Background())

	t.Run("SetAttributesThroughAPI", func(t *testing.T) {
		var reply string
		attrPrf := &engine.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &engine.APIAttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_API",
				FilterIDs: []string{"*string:~*req.AttrSource:api"},
				Attributes: []*engine.ExternalAttribute{
					{
						Path:  "*req.3*4",
						Type:  utils.MetaMultiply,
						Value: "3;4",
					},
					{
						Path:  "*req.12/4",
						Type:  utils.MetaDivide,
						Value: "12;4",
					},
					{
						Path:  "*req.3+4",
						Type:  utils.MetaSum,
						Value: "3;4",
					},
					{
						Path:  "*req.3-4",
						Type:  utils.MetaDifference,
						Value: "3;4",
					},
					{
						Path:  "*req.MultiplyBetweenVariables",
						Type:  utils.MetaMultiply,
						Value: "~*req.Elem1;~*req.Elem2",
					},
				},
			},
		}
		if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile, attrPrf, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckFieldsAlteredByAttributeS", func(t *testing.T) {
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event_test",
			Event: map[string]any{
				"AttrSource": "csv",
				"Elem1":      "3",
				"Elem2":      "4",
			},
		}
		expected := engine.AttrSProcessEventReply{
			AlteredFields: []*engine.FieldsAltered{
				{
					MatchedProfileID: "cgrates.org:ATTR_ARITH",
					Fields:           []string{"*req.12/4", "*req.3*4", "*req.3+4", "*req.3-4", "*req.MultiplyBetweenVariables"},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event_test",
				Event: map[string]any{
					"AttrSource":               "csv",
					"Elem1":                    "3",
					"Elem2":                    "4",
					"12/4":                     "3",
					"3*4":                      "12",
					"3+4":                      "7",
					"3-4":                      "-1",
					"MultiplyBetweenVariables": "12",
				},
				APIOpts: make(map[string]any),
			},
		}

		var reply engine.AttrSProcessEventReply
		if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
			ev, &reply); err != nil {
			t.Error(err)
		} else {
			sort.Strings(expected.AlteredFields[0].Fields)
			sort.Strings(reply.AlteredFields[0].Fields)
			if !reflect.DeepEqual(expected, reply) {
				t.Errorf("\nexpected: %s, \nreceived: %s",
					utils.ToJSON(expected), utils.ToJSON(reply))
			}
		}

		ev.Event["AttrSource"] = "api"
		expected.AlteredFields = []*engine.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_API",
				Fields:           []string{"*req.12/4", "*req.3*4", "*req.3+4", "*req.3-4", "*req.MultiplyBetweenVariables"},
			},
		}
		expected.CGREvent.Event = map[string]any{
			"AttrSource":               "api",
			"Elem1":                    "3",
			"Elem2":                    "4",
			"12/4":                     "3",
			"3*4":                      "12",
			"3+4":                      "7",
			"3-4":                      "-1",
			"MultiplyBetweenVariables": "12",
		}

		reply = engine.AttrSProcessEventReply{}
		if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
			ev, &reply); err != nil {
			t.Error(err)
		} else {
			sort.Strings(expected.AlteredFields[0].Fields)
			sort.Strings(reply.AlteredFields[0].Fields)
			if !reflect.DeepEqual(expected, reply) {
				t.Errorf("\nexpected: %s, \nreceived: %s",
					utils.ToJSON(expected), utils.ToJSON(reply))
			}
		}
	})
}
