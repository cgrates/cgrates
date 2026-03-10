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

package sessions

import (
	"strings"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/birpc/context"
)

func TestSessionSv1ProcessEventRoutes(t *testing.T) {
	var dbcfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.InternalDBCfg
	case utils.MetaRedis:
		dbcfg = engine.RedisDBCfg
	case utils.MetaMySQL:
		dbcfg = engine.MySQLDBCfg
	case utils.MetaMongo:
		dbcfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbcfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
            "logger": {
                "level": 7
            },
            "sessions": {
                "enabled": true,
                "conns": {
                    "*routes": [{"Values": ["*localhost"]}]
                }
            },
            "routes": {
                "enabled": true,
                "indexed_selects": true,
                "string_indexed_fields": ["*req.Account"]
            },
            "chargers": {
                "enabled": true
            },
            "admins": {
                "enabled": true
            }
            }`,
		DBCfg:    dbcfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),
	}

	client, _ := ng.Run(t)

	// t.Cleanup(func() {
	// 	if ng.LogBuffer != nil {
	// 		fmt.Println(ng.LogBuffer)
	// 	}
	// })

	var reply string
	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_DEST_1003",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Destination",
						Values:  []string{"1003"},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetRouteProfile,
		&utils.RouteProfileWithAPIOpts{
			RouteProfile: &utils.RouteProfile{
				Tenant:    "cgrates.org",
				ID:        "ROUTE_ACNT_1001",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{Weight: 0},
				},
				Sorting: utils.MetaWeight,
				Routes: []*utils.Route{
					{
						ID:        "vendor1",
						FilterIDs: []string{"FLTR_DEST_1003"},
						Weights:   utils.DynamicWeights{{Weight: 10}},
						Blockers:  utils.DynamicBlockers{{Blocker: false}},
					},
					{
						ID:       "vendor2",
						Weights:  utils.DynamicWeights{{Weight: 20}},
						Blockers: utils.DynamicBlockers{{Blocker: false}},
					},
					{
						ID:       "vendor3",
						Weights:  utils.DynamicWeights{{Weight: 40}},
						Blockers: utils.DynamicBlockers{{Blocker: false}},
					},
					{
						ID:       "vendor4",
						Weights:  utils.DynamicWeights{{Weight: 35}},
						Blockers: utils.DynamicBlockers{{Blocker: false}},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetRouteProfile failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CGR_DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
				Weights: utils.DynamicWeights{
					{Weight: 0},
				},
				Blockers: utils.DynamicBlockers{
					{Blocker: false},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile failed: %v", err)
	}

	t.Run("noFlags", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant:  "cgrates.org",
				ID:      "noFlags",
				APIOpts: map[string]any{},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed without Routes flags: %v", err)
		}
		if len(rply.RouteProfiles) > 0 {
			t.Errorf("RouteProfiles should be empty without *routes flag, got: %v", rply.RouteProfiles)
		}
	})

	t.Run("noMatchingProfileBlockerError", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingProfileBlockerError",
				APIOpts: map[string]any{
					utils.MetaRoutes:          true,
					utils.OptsSesBlockerError: true,
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "9999",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err == nil {
			t.Error("expected ROUTES_ERROR:NOT_FOUND, got nil")
		} else if !strings.Contains(err.Error(), "NOT_FOUND") {
			t.Errorf("expected ROUTES_ERROR:NOT_FOUND, got: %v", err)
		}
		if len(rply.RouteProfiles) > 0 {
			t.Errorf("RouteProfiles should be empty on error, got: %v", rply.RouteProfiles)
		}
	})

	t.Run("noMatchingProfileNonBlockerError", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingProfileNonBlockerError",
				APIOpts: map[string]any{
					utils.MetaRoutes:          true,
					utils.OptsSesBlockerError: false,
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "9999",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err == nil {
			t.Error("expected PARTIALLY_EXECUTED, got nil")
		} else if err.Error() != utils.ErrPartiallyExecuted.Error() {
			t.Errorf("expected PARTIALLY_EXECUTED, got: %v", err)
		}
		if len(rply.RouteProfiles) > 0 {
			t.Errorf("RouteProfiles should be empty on error, got: %v", rply.RouteProfiles)
		}
	})

	t.Run("withRoutesFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withRoutesFlag",
				APIOpts: map[string]any{
					utils.MetaRoutes: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed with *routes flag: %v", err)
		}
		if rply.RouteProfiles == nil {
			t.Fatal("RouteProfiles should not be nil with *routes flag")
		}

		sortedRoutes, ok := rply.RouteProfiles[utils.MetaPrimary]
		if !ok || len(sortedRoutes) == 0 {
			t.Fatal("no RouteProfiles entry for *primary RunID")
		}

		if sortedRoutes[0].ProfileID != "ROUTE_ACNT_1001" {
			t.Errorf("ProfileID = %s, want ROUTE_ACNT_1001", sortedRoutes[0].ProfileID)
		}
		if len(sortedRoutes[0].Routes) == 0 {
			t.Fatal("Expected at least one route in the sorted list")
		}
		if sortedRoutes[0].Routes[0].RouteID != "vendor3" {
			t.Errorf("First RouteID = %s, want vendor3 (highest weight 40)", sortedRoutes[0].Routes[0].RouteID)
		}
	})

	t.Run("withRoutesFlagNoFilterMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withRoutesFlagNoFilterMatch",
				APIOpts: map[string]any{
					utils.MetaRoutes: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if rply.RouteProfiles == nil {
			t.Fatal("RouteProfiles should not be nil")
		}

		sortedRoutes, ok := rply.RouteProfiles[utils.MetaPrimary]
		if !ok || len(sortedRoutes) == 0 {
			t.Fatal("no RouteProfiles entry for *primary RunID")
		}

		if len(sortedRoutes[0].Routes) != 3 {
			t.Errorf("Expected 3 routes (vendor1 excluded by filter), got %d", len(sortedRoutes[0].Routes))
		}
		for _, r := range sortedRoutes[0].Routes {
			if r.RouteID == "vendor1" {
				t.Error("vendor1 should be excluded (FLTR_DEST_1003 does not match destination 2001)")
			}
		}
	})
}
