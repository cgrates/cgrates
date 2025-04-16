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

package chargers

import (
	"path"
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	chargersCfgPath   string
	chargersCfg       *config.CGRConfig
	chargersRPC       *birpc.Client
	chargersConfigDIR string //run tests for specific configuration

	sTestsChargers = []func(t *testing.T){
		testChargersInitCfg,
		testChargersInitDataDb,
		testChargersResetStorDb,
		testChargersStartEngine,
		testChargersSRPCConn,

		// tests for AdminSv1 APIs
		testChargersGetChargerProfileBeforeSet,
		testChargersGetChargerProfileIDsBeforeSet,
		testChargersGetChargerProfileCountBeforeSet,
		testChargersGetChargerProfilesBeforeSet,
		testChargersSetChargerProfiles,
		testChargersGetChargerProfileAfterSet,
		testChargersGetChargerProfileIDsAfterSet,
		testChargersGetChargerProfileCountAfterSet,
		testChargersGetChargerProfilesAfterSet,
		testChargersRemoveChargerProfile,
		testChargersGetChargerProfileAfterRemove,
		testChargersGetChargerProfileIDsAfterRemove,
		testChargersGetChargerProfileCountAfterRemove,
		testChargersGetChargerProfilesAfterRemove,

		// tests for ChargerSv1 APIs
		testChargersSetGetChargerProfileEvent,
		testChargersGetChargersForEvent,
		testChargersProcessEvent,
		testChargersGetChargerProfilesWithPrefix,

		// blocker behaviour test
		testChargersBlockerRemoveChargerProfiles,
		testChargersBlockerSetChargerProfiles,
		testChargersBlockerGetChargersForEvent,

		testChargersSKillEngine,
	}
)

func TestChargersIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		chargersConfigDIR = "apis_chargers_internal"
	case utils.MetaMongo:
		chargersConfigDIR = "apis_chargers_mongo"
	case utils.MetaMySQL:
		chargersConfigDIR = "apis_chargers_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsChargers {
		t.Run(chargersConfigDIR, stest)
	}
}

func testChargersInitCfg(t *testing.T) {
	var err error
	chargersCfgPath = path.Join(*utils.DataDir, "conf", "samples", chargersConfigDIR)
	chargersCfg, err = config.NewCGRConfigFromPath(context.Background(), chargersCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testChargersInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(chargersCfg); err != nil {
		t.Fatal(err)
	}
}

func testChargersResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(chargersCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testChargersStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(chargersCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testChargersSRPCConn(t *testing.T) {
	chargersRPC = engine.NewRPCClient(t, chargersCfg.ListenCfg(), *utils.Encoding)
}

func testChargersGetChargerProfileBeforeSet(t *testing.T) {
	var replyChargerProfile utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_CHARGER1",
			}}, &replyChargerProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testChargersGetChargerProfilesBeforeSet(t *testing.T) {
	var replyChargerProfiles *[]*utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyChargerProfiles); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testChargersGetChargerProfileIDsBeforeSet(t *testing.T) {
	var replyChargerProfileIDs []string
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyChargerProfileIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testChargersGetChargerProfileCountBeforeSet(t *testing.T) {
	var replyCount int
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if replyCount != 0 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testChargersSetChargerProfiles(t *testing.T) {
	chargerProfiles := []*utils.ChargerProfileWithAPIOpts{
		{
			ChargerProfile: &utils.ChargerProfile{
				ID:        "TestA_CHARGER1",
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
				RunID:        "run1",
				AttributeIDs: []string{"ATTR_TEST1"},
			},
		},
		{
			ChargerProfile: &utils.ChargerProfile{
				ID:        "TestA_CHARGER2",
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
				RunID:        "run2",
				AttributeIDs: []string{"ATTR_TEST2"},
			},
		},
		{
			ChargerProfile: &utils.ChargerProfile{
				ID:        "TestA_CHARGER3",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				RunID:        "run3",
				AttributeIDs: []string{"ATTR_TEST3"},
			},
		},
		{
			ChargerProfile: &utils.ChargerProfile{
				ID:        "TestB_CHARGER1",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				RunID:        "run4",
				AttributeIDs: []string{"ATTR_TEST4"},
			},
		},
		{
			ChargerProfile: &utils.ChargerProfile{
				ID:        "TestB_CHARGER2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				RunID:        "run5",
				AttributeIDs: []string{"ATTR_TEST5"},
			},
		},
	}

	var reply string
	for _, chargerProfile := range chargerProfiles {
		if err := chargersRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
			chargerProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testChargersGetChargerProfileAfterSet(t *testing.T) {
	expectedChargerProfile := utils.ChargerProfile{
		ID:        "TestA_CHARGER1",
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
		RunID:        "run1",
		AttributeIDs: []string{"ATTR_TEST1"},
	}
	var replyChargerProfile utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_CHARGER1",
			}}, &replyChargerProfile); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyChargerProfile, expectedChargerProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expectedChargerProfile), utils.ToJSON(replyChargerProfile))
	}
}

func testChargersGetChargerProfileIDsAfterSet(t *testing.T) {
	expectedIDs := []string{"TestA_CHARGER1", "TestA_CHARGER2", "TestA_CHARGER3", "TestB_CHARGER1", "TestB_CHARGER2"}
	var replyChargerProfileIDs []string
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyChargerProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyChargerProfileIDs)
		if !slices.Equal(replyChargerProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyChargerProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_CHARGER1", "TestA_CHARGER2", "TestA_CHARGER3"}
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyChargerProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyChargerProfileIDs)
		if !slices.Equal(replyChargerProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyChargerProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_CHARGER1", "TestB_CHARGER2"}
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyChargerProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyChargerProfileIDs)
		if !slices.Equal(replyChargerProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyChargerProfileIDs)
		}
	}
}

func testChargersGetChargerProfileCountAfterSet(t *testing.T) {
	var replyCount int
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 5 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 3 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testChargersGetChargerProfilesAfterSet(t *testing.T) {
	expectedChargerProfiles := []*utils.ChargerProfile{
		{
			ID:        "TestA_CHARGER1",
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
			RunID:        "run1",
			AttributeIDs: []string{"ATTR_TEST1"},
		},
		{
			ID:        "TestA_CHARGER2",
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
			RunID:        "run2",
			AttributeIDs: []string{"ATTR_TEST2"},
		},
		{
			ID:        "TestA_CHARGER3",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			RunID:        "run3",
			AttributeIDs: []string{"ATTR_TEST3"},
		},
		{
			ID:        "TestB_CHARGER1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 5,
				},
			},
			RunID:        "run4",
			AttributeIDs: []string{"ATTR_TEST4"},
		},
		{
			ID:        "TestB_CHARGER2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 25,
				},
			},
			RunID:        "run5",
			AttributeIDs: []string{"ATTR_TEST5"},
		},
	}
	var replyChargerProfiles []*utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyChargerProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyChargerProfiles, func(i, j int) bool {
			return replyChargerProfiles[i].ID < replyChargerProfiles[j].ID
		})
		if !reflect.DeepEqual(replyChargerProfiles, expectedChargerProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedChargerProfiles), utils.ToJSON(replyChargerProfiles))
		}
	}
}

func testChargersRemoveChargerProfile(t *testing.T) {
	var reply string
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_CHARGER2",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testChargersGetChargerProfileAfterRemove(t *testing.T) {
	var replyChargerProfile utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Charger2",
			}}, &replyChargerProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testChargersGetChargerProfileIDsAfterRemove(t *testing.T) {
	expectedIDs := []string{"TestA_CHARGER1", "TestA_CHARGER3", "TestB_CHARGER1", "TestB_CHARGER2"}
	var replyChargerProfileIDs []string
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyChargerProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyChargerProfileIDs)
		if !slices.Equal(replyChargerProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyChargerProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_CHARGER1", "TestA_CHARGER3"}
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyChargerProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyChargerProfileIDs)
		if !slices.Equal(replyChargerProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyChargerProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_CHARGER1", "TestB_CHARGER2"}
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyChargerProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyChargerProfileIDs)
		if !slices.Equal(replyChargerProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyChargerProfileIDs)
		}
	}
}

func testChargersGetChargerProfileCountAfterRemove(t *testing.T) {
	var replyCount int
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 4 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testChargersGetChargerProfilesAfterRemove(t *testing.T) {
	expectedChargerProfiles := []*utils.ChargerProfile{
		{
			ID:        "TestA_CHARGER1",
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
			RunID:        "run1",
			AttributeIDs: []string{"ATTR_TEST1"},
		},
		{
			ID:        "TestA_CHARGER3",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			RunID:        "run3",
			AttributeIDs: []string{"ATTR_TEST3"},
		},
		{
			ID:        "TestB_CHARGER1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 5,
				},
			},
			RunID:        "run4",
			AttributeIDs: []string{"ATTR_TEST4"},
		},
		{
			ID:        "TestB_CHARGER2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 25,
				},
			},
			RunID:        "run5",
			AttributeIDs: []string{"ATTR_TEST5"},
		},
	}
	var replyChargerProfiles []*utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyChargerProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyChargerProfiles, func(i, j int) bool {
			return replyChargerProfiles[i].ID < replyChargerProfiles[j].ID
		})
		if !reflect.DeepEqual(replyChargerProfiles, expectedChargerProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedChargerProfiles), utils.ToJSON(replyChargerProfiles))
		}
	}
}

func testChargersGetChargerProfilesWithPrefix(t *testing.T) {
	chgrsPrf := &utils.ChargerProfileWithAPIOpts{
		ChargerProfile: &utils.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "aTEST_CHARGERS_IT_TEST",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}
	var reply string
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chgrsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedChargerPrf := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "aTEST_CHARGERS_IT_TEST",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	var result *utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "aTEST_CHARGERS_IT_TEST",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedChargerPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedChargerPrf), utils.ToJSON(result))
	}
	var reply2 []*utils.ChargerProfile
	args := &utils.ArgsItemIDs{
		ItemsPrefix: "aTEST",
	}
	expected := []*utils.ChargerProfile{
		{
			Tenant:       "cgrates.org",
			ID:           "aTEST_CHARGERS_IT_TEST",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
		args, &reply2); err != nil {
		t.Error(err)
	}
	sort.Slice(reply2, func(i, j int) bool {
		return reply2[i].ID < reply2[j].ID
	})
	if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply2))
	}
}

func testChargersSetGetChargerProfileEvent(t *testing.T) {
	chgrsPrf := &utils.ChargerProfileWithAPIOpts{
		ChargerProfile: &utils.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "TEST_CHARGERS_IT_TEST",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}
	var reply string
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chgrsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedChargerPrf := &utils.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "TEST_CHARGERS_IT_TEST",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	var result *utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_CHARGERS_IT_TEST",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedChargerPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedChargerPrf), utils.ToJSON(result))
	}
}
func testChargersGetChargersForEvent(t *testing.T) {

	expected := []*utils.ChargerProfile{
		{
			Tenant:       "cgrates.org",
			ID:           "TEST_CHARGERS_IT_TEST",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "eventCharger",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]any{},
	}
	var reply []*utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.ChargerSv1GetChargersForEvent,
		cgrEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("\nExpected %+v, \nreceived %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testChargersProcessEvent(t *testing.T) {
	expected := &[]*ChrgSProcessEventReply{
		{
			ChargerSProfile: "TEST_CHARGERS_IT_TEST",
			AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "eventCharger",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
				},
				APIOpts: map[string]any{
					utils.MetaChargeID: utils.UUIDSha1Prefix(),
					"*subsys":          "*chargers",
					utils.MetaRunID:    utils.MetaDefault,
				},
			},
		},
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "eventCharger",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]any{},
	}
	reply := &[]*ChrgSProcessEventReply{}
	if err := chargersRPC.Call(context.Background(), utils.ChargerSv1ProcessEvent,
		cgrEv, &reply); err != nil {
		t.Error(err)
	} else {
		(*reply)[0].CGREvent.APIOpts[utils.MetaChargeID] = (*expected)[0].CGREvent.APIOpts[utils.MetaChargeID]
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("\nExpected %+v, \nreceived %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}
}

func testChargersBlockerRemoveChargerProfiles(t *testing.T) {
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_CHARGERS_IT_TEST", "TestA_CHARGER1", "TestA_CHARGER3", "TestB_CHARGER1", "TestB_CHARGER2", "aTEST_CHARGERS_IT_TEST"}
	var chargerProfileIDs []string
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs, args, &chargerProfileIDs); err != nil {
		t.Fatal(err)
	} else {
		sort.Strings(chargerProfileIDs)
		if !slices.Equal(chargerProfileIDs, expected) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, chargerProfileIDs)
		}
	}
	var reply string
	for _, chargerProfileID := range chargerProfileIDs {
		argsRem := utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     chargerProfileID,
			},
		}
		if err := chargersRPC.Call(context.Background(), utils.AdminSv1RemoveChargerProfile, argsRem, &reply); err != nil {
			t.Fatal(err)
		}
	}
	if err := chargersRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs, args, &chargerProfileIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testChargersBlockerSetChargerProfiles(t *testing.T) {
	chargerProfiles := []*utils.ChargerProfileWithAPIOpts{
		{
			ChargerProfile: &utils.ChargerProfile{
				ID:        "CHARGER_TEST_1",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:BlockerBehaviour"},
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
				RunID: "run1",
			},
		},
		{
			ChargerProfile: &utils.ChargerProfile{
				ID:        "CHARGER_TEST_2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:BlockerBehaviour"},
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
				RunID: "run2",
			},
		},
		{
			ChargerProfile: &utils.ChargerProfile{
				ID:        "CHARGER_TEST_3",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:BlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				RunID: "run3",
			},
		},
		{
			ChargerProfile: &utils.ChargerProfile{
				ID:        "CHARGER_TEST_4",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:BlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				RunID: "run4",
			},
		},
	}

	var reply string
	for _, chargerProfile := range chargerProfiles {
		if err := chargersRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
			chargerProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testChargersBlockerGetChargersForEvent(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventGetChargerProfiles",
		Event: map[string]any{
			"TestCase": "BlockerBehaviour",
		},
		APIOpts: map[string]any{},
	}
	expected := []*utils.ChargerProfile{
		{
			ID:        "CHARGER_TEST_1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:BlockerBehaviour"},
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
			RunID: "run1",
		},
		{
			ID:        "CHARGER_TEST_3",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:BlockerBehaviour"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			RunID: "run3",
		},
		{
			ID:        "CHARGER_TEST_2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:BlockerBehaviour"},
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
			RunID: "run2",
		},
	}
	var reply []*utils.ChargerProfile
	if err := chargersRPC.Call(context.Background(), utils.ChargerSv1GetChargersForEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

// Kill the engine when it is about to be finished
func testChargersSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
