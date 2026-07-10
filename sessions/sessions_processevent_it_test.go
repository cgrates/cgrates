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
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

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
                    "*routes": [{"connIDs": ["*localhost"]}]
                }
            },
            "routes": {
                "enabled": true,
                "indexedSelects": true,
                "stringIndexedFields": ["*req.Account"]
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
				Tenant: "cgrates.org",
				ID:     "noFlags",
				APIOpts: map[string]any{
					utils.MetaOriginID: "OriginID",
				},
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
					utils.MetaOriginID:        "OriginID",
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
					utils.MetaOriginID:        "OriginID",
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
					utils.MetaRoutes:   true,
					utils.MetaOriginID: "OriginID",
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
					utils.MetaRoutes:   true,
					utils.MetaOriginID: "OriginID",
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
func TestSessionSv1ProcessEventStats(t *testing.T) {
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
		"*stats": [{"tenant":"","filterIDs":[],"connIDs":["*localhost"]}]
	}
},
"stats": {
	"enabled": true,
	"storeInterval": "-1",
	"indexedSelects": true,
	"stringIndexedFields": ["*req.Account"]
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

	if err := client.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		&utils.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &utils.StatQueueProfile{
				Tenant:    "cgrates.org",
				ID:        "SQ_ACNT_1001",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{Weight: 10},
				},
				Blockers: utils.DynamicBlockers{
					{Blocker: false},
				},
				QueueLength:  100,
				TTL:          -1,
				MinItems:     0,
				Stored:       false,
				ThresholdIDs: []string{utils.MetaNone},
				Metrics: []*utils.MetricWithFilters{
					{MetricID: utils.MetaACC},
					{MetricID: utils.MetaTCC},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetStatQueueProfile failed: %v", err)
	}

	t.Run("noFlags", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noFlags",
				APIOpts: map[string]any{
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed without *stats flag: %v", err)
		}
		if len(rply.StatQueueIDs) > 0 {
			t.Errorf("StatQueueIDs should be empty without *stats flag, got: %v", rply.StatQueueIDs)
		}

		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "SQ_ACNT_1001",
				},
			}, &metrics); err != nil {
			t.Fatalf("StatSv1GetQueueFloatMetrics failed: %v", err)
		}
		if metrics[utils.MetaACC] != -1 {
			t.Errorf("*acc = %v, want -1 (N/A, no event processed yet)", metrics[utils.MetaACC])
		}
		if metrics[utils.MetaTCC] != -1 {
			t.Errorf("*tcc = %v, want -1 (N/A, no event processed yet)", metrics[utils.MetaTCC])
		}
	})

	t.Run("withStatsFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withStatsFlag",
				APIOpts: map[string]any{
					utils.MetaStats:    true,
					utils.MetaCost:     0.6,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed with *stats flag: %v", err)
		}
		if rply.StatQueueIDs == nil {
			t.Fatal("StatQueueIDs should not be nil with *stats flag")
		}
		sqIDs, ok := rply.StatQueueIDs[utils.MetaPrimary]
		if !ok || len(sqIDs) == 0 {
			t.Fatalf("expected StatQueueIDs[%s], got: %v", utils.MetaPrimary, rply.StatQueueIDs)
		}
		sort.Strings(sqIDs)
		if !reflect.DeepEqual(sqIDs, []string{"SQ_ACNT_1001"}) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", []string{"SQ_ACNT_1001"}, sqIDs)
		}

		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "SQ_ACNT_1001",
				},
			}, &metrics); err != nil {
			t.Fatalf("StatSv1GetQueueFloatMetrics failed: %v", err)
		}
		if metrics[utils.MetaACC] != 0.6 {
			t.Errorf("*acc = %v, want 0.6 (single event)", metrics[utils.MetaACC])
		}
		if metrics[utils.MetaTCC] != 0.6 {
			t.Errorf("*tcc = %v, want 0.6 (single event)", metrics[utils.MetaTCC])
		}
	})

	t.Run("secondEventAccumulatesState", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "secondEvent",
				APIOpts: map[string]any{
					utils.MetaStats:    true,
					utils.MetaCost:     1.2,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "120s",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent (second event) failed: %v", err)
		}
		if _, ok := rply.StatQueueIDs[utils.MetaPrimary]; !ok {
			t.Fatalf("expected StatQueueIDs[%s] on second event", utils.MetaPrimary)
		}

		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "SQ_ACNT_1001",
				},
			}, &metrics); err != nil {
			t.Fatalf("StatSv1GetQueueFloatMetrics failed: %v", err)
		}
		if metrics[utils.MetaACC] != 0.9 {
			t.Errorf("*acc = %v, want 0.9 (avg of 0.6 and 1.2)", metrics[utils.MetaACC])
		}
		if metrics[utils.MetaTCC] != 1.8 {
			t.Errorf("*tcc = %v, want 1.8 (sum of 0.6 and 1.2)", metrics[utils.MetaTCC])
		}
	})

	t.Run("noMatchingProfileNonBlockerError", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingProfileNonBlockerError",
				APIOpts: map[string]any{
					utils.MetaStats:           true,
					utils.OptsSesBlockerError: false,
					utils.MetaCost:            9.9,
					utils.MetaOriginID:        "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "9999",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)
		if err == nil {
			t.Error("expected PARTIALLY_EXECUTED, got nil")
		} else if err.Error() != utils.ErrPartiallyExecuted.Error() {
			t.Errorf("expected PARTIALLY_EXECUTED, got: %v", err)
		}
		if len(rply.StatQueueIDs) > 0 {
			t.Errorf("StatQueueIDs should be empty, got: %v", rply.StatQueueIDs)
		}

		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "SQ_ACNT_1001",
				},
			}, &metrics); err != nil {
			t.Fatalf("StatSv1GetQueueFloatMetrics failed: %v", err)
		}
		if metrics[utils.MetaACC] != 0.9 {
			t.Errorf("*acc = %v, want 0.9 (unmatched event must not affect queue)", metrics[utils.MetaACC])
		}
		if metrics[utils.MetaTCC] != 1.8 {
			t.Errorf("*tcc = %v, want 1.8 (unmatched event must not affect queue)", metrics[utils.MetaTCC])
		}
	})

	t.Run("noMatchingProfileBlockerError", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingProfileBlockerError",
				APIOpts: map[string]any{
					utils.MetaStats:           true,
					utils.OptsSesBlockerError: true,
					utils.MetaCost:            9.9,
					utils.MetaOriginID:        "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "9999",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)
		if err == nil {
			t.Error("expected NOT_FOUND, got nil")
		} else if err.Error() != utils.ErrNotFound.Error() {
			t.Errorf("expected NOT_FOUND, got: %v", err)
		}
		if len(rply.StatQueueIDs) > 0 {
			t.Errorf("StatQueueIDs should be empty on blocker error, got: %v", rply.StatQueueIDs)
		}

		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "SQ_ACNT_1001",
				},
			}, &metrics); err != nil {
			t.Fatalf("StatSv1GetQueueFloatMetrics failed: %v", err)
		}
		if metrics[utils.MetaACC] != 0.9 {
			t.Errorf("*acc = %v, want 0.9 (blocker error must not affect queue)", metrics[utils.MetaACC])
		}
		if metrics[utils.MetaTCC] != 1.8 {
			t.Errorf("*tcc = %v, want 1.8 (blocker error must not affect queue)", metrics[utils.MetaTCC])
		}
	})

	t.Run("withStatsFlagNoFilterMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withStatsFlagNoFilterMatch",
				APIOpts: map[string]any{
					utils.MetaStats:           true,
					utils.OptsSesBlockerError: false,
					utils.MetaCost:            5.0,
					utils.MetaOriginID:        "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "2002",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)
		if err == nil {
			t.Error("expected PARTIALLY_EXECUTED, got nil")
		} else if err.Error() != utils.ErrPartiallyExecuted.Error() {
			t.Errorf("expected PARTIALLY_EXECUTED, got: %v", err)
		}
		if len(rply.StatQueueIDs) > 0 {
			t.Errorf("StatQueueIDs should be empty when filter does not match, got: %v", rply.StatQueueIDs)
		}

		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "SQ_ACNT_1001",
				},
			}, &metrics); err != nil {
			t.Fatalf("StatSv1GetQueueFloatMetrics failed: %v", err)
		}
		if metrics[utils.MetaACC] != 0.9 {
			t.Errorf("*acc = %v, want 0.9 (non-matching account must not affect queue)", metrics[utils.MetaACC])
		}
		if metrics[utils.MetaTCC] != 1.8 {
			t.Errorf("*tcc = %v, want 1.8 (non-matching account must not affect queue)", metrics[utils.MetaTCC])
		}
	})

	t.Run("thirdEventFinalState", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "thirdEvent",
				APIOpts: map[string]any{
					utils.MetaStats:    true,
					utils.MetaCost:     0.3,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "30s",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent (third event) failed: %v", err)
		}
		if _, ok := rply.StatQueueIDs[utils.MetaPrimary]; !ok {
			t.Fatalf("expected StatQueueIDs[%s] on third event", utils.MetaPrimary)
		}

		time.Sleep(50 * time.Millisecond)

		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "SQ_ACNT_1001",
				},
			}, &metrics); err != nil {
			t.Fatalf("StatSv1GetQueueFloatMetrics failed: %v", err)
		}
		if metrics[utils.MetaACC] != 0.7 {
			t.Errorf("*acc = %v, want 0.7 (avg of 0.6, 1.2, 0.3)", metrics[utils.MetaACC])
		}
		if metrics[utils.MetaTCC] != 2.1 {
			t.Errorf("*tcc = %v, want 2.1 (sum of 0.6, 1.2, 0.3)", metrics[utils.MetaTCC])
		}
	})
}
func TestSessionSv1ProcessEventThresholds(t *testing.T) {
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
		"*thresholds": [{"tenant":"","filterIDs":[],"connIDs":["*localhost"]}]
	}
},
"thresholds": {
	"enabled": true,
	"storeInterval": "-1",
	"indexedSelects": true,
	"stringIndexedFields": ["*req.Account"]
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
	}

	client, _ := ng.Run(t)

	var reply string
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

	if err := client.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
		&utils.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &utils.ThresholdProfile{
				Tenant:    "cgrates.org",
				ID:        "THD_ACNT_1001",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{Weight: 10},
				},
				Blocker: false,
				MaxHits: -1,
				MinHits: 100,
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetThresholdProfile failed: %v", err)
	}

	t.Run("noFlags", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noFlags",
				APIOpts: map[string]any{
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed without *thresholds flag: %v", err)
		}
		if len(rply.ThresholdIDs) > 0 {
			t.Errorf("ThresholdIDs should be empty without *thresholds flag, got: %v", rply.ThresholdIDs)
		}
	})

	t.Run("withThresholdsFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withThresholdsFlag",
				APIOpts: map[string]any{
					utils.MetaThresholds: true,
					utils.MetaOriginID:   "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed with *thresholds flag: %v", err)
		}
		if rply.ThresholdIDs == nil {
			t.Fatal("ThresholdIDs should not be nil with *thresholds flag")
		}
		thdIDs, ok := rply.ThresholdIDs[utils.MetaPrimary]
		if !ok || len(thdIDs) == 0 {
			t.Fatalf("expected ThresholdIDs[%s], got: %v", utils.MetaPrimary, rply.ThresholdIDs)
		}
		sort.Strings(thdIDs)
		if !reflect.DeepEqual(thdIDs, []string{"THD_ACNT_1001"}) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", []string{"THD_ACNT_1001"}, thdIDs)
		}
	})

	t.Run("noMatchingProfileNonBlockerError", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingProfileNonBlockerError",
				APIOpts: map[string]any{
					utils.MetaThresholds:      true,
					utils.OptsSesBlockerError: false,
					utils.MetaOriginID:        "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "9999",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)
		if err == nil {
			t.Error("expected PARTIALLY_EXECUTED, got nil")
		} else if err.Error() != utils.ErrPartiallyExecuted.Error() {
			t.Errorf("expected PARTIALLY_EXECUTED, got: %v", err)
		}
		if len(rply.ThresholdIDs) > 0 {
			t.Errorf("ThresholdIDs should be empty, got: %v", rply.ThresholdIDs)
		}
	})

	t.Run("noMatchingProfileBlockerError", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingProfileBlockerError",
				APIOpts: map[string]any{
					utils.MetaThresholds:      true,
					utils.OptsSesBlockerError: true,
					utils.MetaOriginID:        "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "9999",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)
		if err == nil {
			t.Error("expected NOT_FOUND, got nil")
		} else if err.Error() != utils.ErrNotFound.Error() {
			t.Errorf("expected NOT_FOUND, got: %v", err)
		}
		if len(rply.ThresholdIDs) > 0 {
			t.Errorf("ThresholdIDs should be empty on blocker error, got: %v", rply.ThresholdIDs)
		}
	})

	t.Run("noFilterMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noFilterMatch",
				APIOpts: map[string]any{
					utils.MetaThresholds:      true,
					utils.OptsSesBlockerError: false,
					utils.MetaOriginID:        "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "2002",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)
		if err == nil {
			t.Error("expected PARTIALLY_EXECUTED, got nil")
		} else if err.Error() != utils.ErrPartiallyExecuted.Error() {
			t.Errorf("expected PARTIALLY_EXECUTED, got: %v", err)
		}
		if len(rply.ThresholdIDs) > 0 {
			t.Errorf("ThresholdIDs should be empty when filter does not match, got: %v", rply.ThresholdIDs)
		}
	})
}

func TestSessionSv1ProcessEventChargerSSessionTerminate(t *testing.T) {

	ng := engine.TestEngine{
		ConfigJSON: `{
"logger": {"level": 7},
"sessions": {
	"enabled": true,
	"conns": {
		"*chargers": [{"connIDs": ["*localhost"]}],
		"*routes":   [{"connIDs": ["*localhost"]}]
	}
},
"chargers": {
	"enabled": true
},
"routes": {
	"enabled": true
},
"admins": {
	"enabled": true
}
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),
	}

	// t.Cleanup(func() {
	// 	if ng.LogBuffer != nil {
	// 		fmt.Println(ng.LogBuffer)
	// 	}
	// })

	client, _ := ng.Run(t)

	var reply string

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CGR_DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
				Weights:      utils.DynamicWeights{{Weight: 10}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetRouteProfile,
		&utils.RouteProfileWithAPIOpts{
			RouteProfile: &utils.RouteProfile{
				Tenant:  "cgrates.org",
				ID:      "ROUTE_ALL",
				Weights: utils.DynamicWeights{{Weight: 10}},
				Sorting: utils.MetaWeight,
				Routes: []*utils.Route{
					{
						ID:       "vendor1",
						Weights:  utils.DynamicWeights{{Weight: 10}},
						Blockers: utils.DynamicBlockers{{Blocker: false}},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetRouteProfile failed: %v", err)
	}

	t.Run("standaloneEventChargersRun", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "standalone1",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaRoutes:   true,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if _, hasDefault := rply.RouteProfiles[utils.MetaDefault]; !hasDefault {
			t.Errorf("expected RouteProfiles[*default] to exist when ChargerS runs, got: %v", rply.RouteProfiles)
		}
	})

	t.Run("sessionEventChargersSkipped", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "session1",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaRoutes:   true,
					utils.MetaSession:  true,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if _, hasDefault := rply.RouteProfiles[utils.MetaDefault]; !hasDefault {
			t.Errorf("expected RouteProfiles[*default] to be present")
		}
	})

	t.Run("terminateEventChargersSkipped", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "terminate1",
				APIOpts: map[string]any{
					utils.MetaChargers:  true,
					utils.MetaRoutes:    true,
					utils.MetaTerminate: true,
					utils.MetaOriginID:  "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if _, hasDefault := rply.RouteProfiles[utils.MetaDefault]; hasDefault {
			t.Errorf("expected RouteProfiles[*default] to be absent when *terminate=true, got: %v", rply.RouteProfiles)
		}
	})
}

func TestSessionSv1ProcessEventRatesFlag(t *testing.T) {
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
			"logger": {"level": 7},
			"sessions": {
				"enabled": true,
				"conns": {
					"*rates":    [{"connIDs": ["*localhost"]}],
					"*chargers": [{"connIDs": ["*localhost"]}]
				}
			},
			"rates": {
				"enabled": true
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
	}

	client, _ := ng.Run(t)

	var reply string

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CGR_DEFAULT",
				RunID:        utils.MetaPrimary,
				AttributeIDs: []string{utils.MetaNone},
				Weights:      utils.DynamicWeights{{Weight: 0}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetRateProfile,
		&utils.RateProfileWithAPIOpts{
			RateProfile: &utils.RateProfile{
				Tenant:          "cgrates.org",
				ID:              "RP_1001",
				FilterIDs:       []string{"*string:~*req.Account:1001"},
				Weights:         utils.DynamicWeights{{Weight: 0}},
				MaxCostStrategy: utils.MetaMaxCostDisconnect,
				Rates: map[string]*utils.Rate{
					"RT_DEFAULT": {
						ID:              "RT_DEFAULT",
						FilterIDs:       []string{},
						Weights:         utils.DynamicWeights{{Weight: 0}},
						ActivationTimes: "* * * * *",
						IntervalRates: []*utils.IntervalRate{
							{
								IntervalStart: utils.NewDecimalFromFloat64(0),
								RecurrentFee:  utils.NewDecimalFromFloat64(0.01),
								Unit:          utils.NewDecimalFromFloat64(float64(time.Minute)),
								Increment:     utils.NewDecimalFromFloat64(float64(time.Minute)),
							},
						},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetRateProfile failed: %v", err)
	}

	t.Run("noRatesFlagEmptyReply", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noRatesFlagEmptyReply",
				APIOpts: map[string]any{
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent without *rates flag failed: %v", err)
		}
		if len(rply.RateSCost) != 0 {
			t.Errorf("RateSCost should be empty without *rates flag, got: %v", rply.RateSCost)
		}
	})

	t.Run("withRatesFlagPopulatesReply", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withRatesFlagPopulatesReply",
				APIOpts: map[string]any{
					utils.MetaRates:    true,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent with *rates flag failed: %v", err)
		}
		if len(rply.RateSCost) == 0 {
			t.Fatal("RateSCost should be populated when *rates flag is set")
		}
		cost, ok := rply.RateSCost[utils.MetaPrimary]
		if !ok {
			t.Fatalf("expected *primary runID in RateSCost, got keys: %v", rply.RateSCost)
		}
		if cost != 0.01 {
			t.Errorf("expected cost 0.01 (1min × 0.01/min), got: %v", cost)
		}
	})

	t.Run("noMatchingRateProfileNonBlocker", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingRateProfileNonBlocker",
				APIOpts: map[string]any{
					utils.MetaRates:           true,
					utils.OptsSesBlockerError: false,
					utils.MetaOriginID:        "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)

		if err == nil {
			t.Error("expected PARTIALLY_EXECUTED error, got nil")
		} else if err.Error() != utils.ErrPartiallyExecuted.Error() {
			t.Errorf("expected PARTIALLY_EXECUTED, got: %v", err)
		}
	})

	t.Run("noMatchingRateProfileBlockerError", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingRateProfileBlockerError",
				APIOpts: map[string]any{
					utils.MetaRates:           true,
					utils.OptsSesBlockerError: true,
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
					utils.Usage:        "60s",
				},
			}, &rply)

		if err == nil {
			t.Error("expected an error with blocker=true and no matching RateProfile, got nil")
		}
	})
}

func TestSessionSv1ProcessEventAccountsAuthorizeRates(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
			"logger": {"level": 7},
			"sessions": {
				"enabled": true,
				"conns": {
					"*accounts": [{"ConnIDs": ["*localhost"]}],
					"*rates":    [{"ConnIDs": ["*localhost"]}]
				}
			},
			"accounts": {
				"enabled": true,
				"indexed_selects": true,
				"conns": {
					"*rates": [{"ConnIDs": ["*localhost"]}]
				}
			},
			"rates":  {"enabled": true},
			"admins": {"enabled": true}
		}`,
		TpFiles: map[string]string{
			utils.RatesCsv: `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP_SIMPLE,,;10,,,,RT_SIMPLE,*string:~*req.Destination:1002,"* * * * *",;10,false,0s,0,1,1m,1m,`,
		},
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}

	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	t.Run("setAccount", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
			&utils.AccountWithAPIOpts{
				Account: &utils.Account{
					Tenant:    "cgrates.org",
					ID:        "1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   utils.DynamicWeights{{Weight: 20}},
					Balances: map[string]*utils.Balance{
						"AbstractBalance": {
							ID:             "AbstractBalance",
							Type:           utils.MetaAbstract,
							Weights:        utils.DynamicWeights{{Weight: 10}},
							Units:          utils.NewDecimal(int64(10*time.Minute), 0),
							RateProfileIDs: []string{"RP_SIMPLE"},
						},
						"ConcreteBalance": {
							ID:      "ConcreteBalance",
							Type:    utils.MetaConcrete,
							Weights: utils.DynamicWeights{{Weight: 5}},
							Units:   utils.NewDecimal(10, 0),
						},
					},
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetAccount failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetAccount reply = %s, want OK", reply)
		}
	})
	t.Run("noFlags", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noFlags",
				APIOpts: map[string]any{
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)
		if err != nil {
			t.Fatalf("ProcessEvent failed without flags: %v", err)
		}
		if len(rply.RateSCost) > 0 {
			t.Errorf("RateSCost should be empty without flags, got: %v", rply.RateSCost)
		}
		if len(rply.AccountSUsage) > 0 {
			t.Errorf("AccountSUsage should be empty without flags, got: %v", rply.AccountSUsage)
		}
	})

	t.Run("accountsAuthorizeFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "accountsAuthorizeFlag",
				APIOpts: map[string]any{
					utils.MetaAccounts:  true,
					utils.MetaAuthorize: true,
					utils.MetaUsage:     10 * time.Minute,
					utils.MetaOriginID:  "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)
		if err != nil {
			t.Fatalf("ProcessEvent failed with *accounts + *authorize: %v", err)
		}
		usage, ok := rply.AccountSUsage[utils.MetaPrimary]
		if !ok {
			t.Fatal("AccountSUsage missing *primary")
		}
		wantUsage := 10 * time.Minute
		if usage != wantUsage {
			t.Errorf("AccountSUsage[*primary] = %v, want %v",
				time.Duration(usage), time.Duration(wantUsage))
		}
		if len(rply.RateSCost) > 0 {
			t.Errorf("RateSCost should be empty without *rates flag, got: %v", rply.RateSCost)
		}
	})

	t.Run("ratesFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ratesFlag",
				APIOpts: map[string]any{
					utils.MetaRates:    true,
					utils.MetaUsage:    10 * time.Minute,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)
		if err != nil {
			t.Fatalf("ProcessEvent failed with *rates: %v", err)
		}
		cost, ok := rply.RateSCost[utils.MetaPrimary]
		if !ok {
			t.Fatalf("RateSCost missing *primary, got: %v", rply.RateSCost)
		}
		if cost != 10.0 {
			t.Errorf("RateSCost[*primary] = %g, want 10.0", cost)
		}
		if len(rply.AccountSUsage) > 0 {
			t.Errorf("AccountSUsage should be empty without *accounts flag, got: %v", rply.AccountSUsage)
		}
	})

	t.Run("accountsAuthorizeAndRates", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "accountsAuthorizeAndRates",
				APIOpts: map[string]any{
					utils.MetaAccounts:  true,
					utils.MetaAuthorize: true,
					utils.MetaRates:     true,
					utils.MetaUsage:     10 * time.Minute,
					utils.MetaOriginID:  "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)
		if err != nil {
			t.Fatalf("ProcessEvent failed with *accounts + *authorize + *rates: %v", err)
		}
		usage, ok := rply.AccountSUsage[utils.MetaPrimary]
		if !ok {
			t.Fatal("AccountSUsage missing *primary")
		}
		wantUsage := 10 * time.Minute
		if usage != wantUsage {
			t.Errorf("AccountSUsage[*primary] = %v, want %v", usage, wantUsage)
		}
		cost, ok := rply.RateSCost[utils.MetaPrimary]
		if !ok {
			t.Fatal("RateSCost missing *primary")
		}
		if cost != 10.0 {
			t.Errorf("RateSCost[*primary] = %g, want 10.0", cost)
		}
	})

	t.Run("noMatchingAccount", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingAccount",
				APIOpts: map[string]any{
					utils.MetaAccounts:  true,
					utils.MetaAuthorize: true,
					utils.MetaUsage:     10 * time.Minute,
					utils.MetaOriginID:  "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)
		if err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Fatalf("expected NOT_FOUND for unknown account, got: %v", err)
		}
	})

	t.Run("noMatchingRate", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingRate",
				APIOpts: map[string]any{
					utils.MetaRates:    true,
					utils.MetaUsage:    10 * time.Minute,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "9999",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)
		if err == nil {
			t.Fatal("expected error for unmatched destination, got none")
		}
	})
}

func TestSessionSv1ProcessEventMultipleFlags(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
            "logger": {"level": 7},
            "sessions": {
                "enabled": true,
                "conns": {
                    "*attributes": [{"ConnIDs": ["*localhost"]}],
                    "*chargers":   [{"ConnIDs": ["*localhost"]}],
                    "*routes":     [{"ConnIDs": ["*localhost"]}],
                    "*rates":      [{"ConnIDs": ["*localhost"]}],
                    "*accounts":   [{"ConnIDs": ["*localhost"]}],
                    "*resources":  [{"ConnIDs": ["*localhost"]}],
                    "*ips":        [{"ConnIDs": ["*localhost"]}]
                }
            },
            "attributes": {"enabled": true},
            "chargers": {
                "enabled": true,
                "conns": {
                    "*attributes": [{"ConnIDs": ["*localhost"]}]
                }
            },
            "routes":  {"enabled": true},
            "rates":   {"enabled": true},
            "accounts": {
                "enabled": true,
                "conns": {
                    "*rates": [{"ConnIDs": ["*localhost"]}]
                }
            },
            "resources": {"enabled": true},
            "ips":       {"enabled": true},
            "admins":    {"enabled": true}
        }`,
		TpFiles: map[string]string{
			utils.RatesCsv: `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP_SIMPLE,,;10,,,,RT_SIMPLE,*string:~*req.Destination:1002,"* * * * *",;10,false,0s,0,1,1m,1m,`,
		},
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}

	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	var reply string

	t.Run("setAttributeProfile", func(t *testing.T) {
		if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
			&utils.APIAttributeProfileWithAPIOpts{
				APIAttributeProfile: &utils.APIAttributeProfile{
					Tenant:    "cgrates.org",
					ID:        "ATTR_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   utils.DynamicWeights{{Weight: 10}},
					Attributes: []*utils.ExternalAttribute{
						{
							Path:  "*req.Subject",
							Type:  utils.MetaConstant,
							Value: "1001",
						},
					},
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetAttributeProfile failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetAttributeProfile reply = %s, want OK", reply)
		}
	})

	t.Run("setChargerProfile", func(t *testing.T) {
		if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
			&utils.ChargerProfileWithAPIOpts{
				ChargerProfile: &utils.ChargerProfile{
					Tenant:       "cgrates.org",
					ID:           "DEFAULT",
					Weights:      utils.DynamicWeights{{Weight: 0}},
					RunID:        utils.MetaDefault,
					AttributeIDs: []string{utils.MetaNone},
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetChargerProfile failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetChargerProfile reply = %s, want OK", reply)
		}
	})

	t.Run("setRouteProfile", func(t *testing.T) {
		if err := client.Call(context.Background(), utils.AdminSv1SetRouteProfile,
			&utils.RouteProfileWithAPIOpts{
				RouteProfile: &utils.RouteProfile{
					Tenant:    "cgrates.org",
					ID:        "ROUTE_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   utils.DynamicWeights{{Weight: 10}},
					Sorting:   utils.MetaWeight,
					Routes: []*utils.Route{
						{
							ID:       "route1",
							Weights:  utils.DynamicWeights{{Weight: 20}},
							Blockers: utils.DynamicBlockers{{Blocker: false}},
						},
					},
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetRouteProfile failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetRouteProfile reply = %s, want OK", reply)
		}
	})

	t.Run("setResourceProfile", func(t *testing.T) {
		if err := client.Call(context.Background(), utils.AdminSv1SetResourceProfile,
			&utils.ResourceProfileWithAPIOpts{
				ResourceProfile: &utils.ResourceProfile{
					Tenant:            "cgrates.org",
					ID:                "RES_1001",
					FilterIDs:         []string{"*string:~*req.Account:1001"},
					Weights:           utils.DynamicWeights{{Weight: 10}},
					UsageTTL:          time.Hour,
					Limit:             10,
					AllocationMessage: "RES_1001_ALLOC",
					Blocker:           false,
					Stored:            true,
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetResourceProfile failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetResourceProfile reply = %s, want OK", reply)
		}
	})

	t.Run("setIPProfile", func(t *testing.T) {
		if err := client.Call(context.Background(), utils.AdminSv1SetIPProfile,
			&utils.IPProfileWithAPIOpts{
				IPProfile: &utils.IPProfile{
					Tenant:    "cgrates.org",
					ID:        "IP_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					TTL:       10 * time.Minute,
					Pools: []*utils.IPPool{
						{
							ID:      "POOL_1001",
							Range:   "10.10.10.1/32",
							Message: "Allocated by test",
						},
					},
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetIPProfile failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetIPProfile reply = %s, want OK", reply)
		}
	})

	t.Run("setAccount", func(t *testing.T) {
		if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
			&utils.AccountWithAPIOpts{
				Account: &utils.Account{
					Tenant:    "cgrates.org",
					ID:        "1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   utils.DynamicWeights{{Weight: 20}},
					Balances: map[string]*utils.Balance{
						"AbstractBalance": {
							ID:             "AbstractBalance",
							Type:           utils.MetaAbstract,
							Weights:        utils.DynamicWeights{{Weight: 10}},
							Units:          utils.NewDecimal(int64(10*time.Minute), 0),
							RateProfileIDs: []string{"RP_SIMPLE"},
						},
						"ConcreteBalance": {
							ID:      "ConcreteBalance",
							Type:    utils.MetaConcrete,
							Weights: utils.DynamicWeights{{Weight: 5}},
							Units:   utils.NewDecimal(10, 0),
						},
					},
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetAccount failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetAccount reply = %s, want OK", reply)
		}
	})

	t.Run("allFlagsOneRequest", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "allFlagsOneRequest",
				APIOpts: map[string]any{
					utils.MetaAttributes:       true,
					utils.MetaChargers:         true,
					utils.MetaRoutes:           true,
					utils.MetaRates:            true,
					utils.MetaAccounts:         true,
					utils.MetaResources:        true,
					utils.MetaIPs:              true,
					utils.MetaAuthorize:        true,
					utils.MetaUsage:            10 * time.Minute,
					utils.MetaOriginID:         "originIDprocessing",
					utils.OptsResourcesUsageID: "resourceUsageTest",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2026-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent allFlagsOneRequest failed: %v", err)
		}
		if len(rply.Attributes) == 0 {
			t.Error("Attributes should be populated")
		}
		if len(rply.RouteProfiles) == 0 {
			t.Error("RouteProfiles should be populated")
		}
		cost, ok := rply.RateSCost[utils.MetaDefault]
		if !ok {
			t.Fatalf("RateSCost missing *default, got: %v", rply.RateSCost)
		}
		if cost != 10.0 {
			t.Errorf("RateSCost[*default] = %g, want 10.0", cost)
		}
		usage, ok := rply.AccountSUsage[utils.MetaDefault]
		if !ok {
			t.Fatalf("AccountSUsage missing *default, got: %v", rply.AccountSUsage)
		}
		if usage != 10*time.Minute {
			t.Errorf("AccountSUsage[*default] = %v, want %v", usage, 10*time.Minute)
		}
		resID, ok := rply.ResourceAllocation[utils.MetaDefault]
		if !ok {
			t.Fatalf("ResourceAllocation missing *default, got: %v", rply.ResourceAllocation)
		}
		if resID != "RES_1001_ALLOC" {
			t.Errorf("ResourceAllocation[*default] = %q, want RES_1001_ALLOC", resID)
		}
		if len(rply.IPsAllocation) == 0 {
			t.Error("IPsAllocation should be populated")
		}
	})
}

func TestSessionSv1ProcessEventEEs(t *testing.T) {
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
			"logger": {"level": 7},
			"sessions": {
				"enabled": true,
				"conns": {
					"*ees": [{"connIDs": ["*localhost"]}]
				}
			},
			"ees": {
				"enabled": true,
				"exporters": [
					{
						"id": "LogExporter",
						"type": "*log",
						"attempts": 1
					}
				]
			},
			"chargers": {"enabled": true},
			"admins": {"enabled": true}
		}`,
		DBCfg:    dbcfg,
		Encoding: *utils.Encoding,
	}
	client, _ := ng.Run(t)

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CGR_DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
				Weights:      utils.DynamicWeights{{Weight: 0}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, new(string)); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile failed: %v", err)
	}

	t.Run("noFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noFlag",
				APIOpts: map[string]any{
					utils.MetaOriginID: "OriginID_noFlag",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if len(rply.EventExporters) != 0 {
			t.Errorf("EventExporters should be empty without *ees flag, got: %v", rply.EventExporters)
		}
	})

	t.Run("withEEsFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withEEsFlag",
				APIOpts: map[string]any{
					utils.MetaEEs:      true,
					utils.MetaOriginID: "OriginID_withEEsFlag",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent with *ees flag failed: %v", err)
		}
		if len(rply.EventExporters) == 0 {
			t.Fatal("EventExporters should not be empty with *ees flag")
		}
		for runID, eesIDs := range rply.EventExporters {
			if len(eesIDs) == 0 {
				t.Errorf("runID %q has no exporter IDs", runID)
			}
		}
	})
}

func TestSessionSv1ProcessEventEEsWithRates(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
			"logger": {"level": 7},
			"sessions": {
				"enabled": true,
				"conns": {
					"*accounts": [{"connIDs": ["*localhost"]}],
					"*rates":    [{"connIDs": ["*localhost"]}],
					"*ees":      [{"connIDs": ["*localhost"]}]
				}
			},
			"accounts": {
				"enabled": true,
				"conns": {
					"*rates": [{"connIDs": ["*localhost"]}]
				}
			},
			"rates":  {"enabled": true},
			"ees": {
				"enabled": true,
				"exporters": [
					{
						"id": "LogExporter",
						"type": "*log",
						"attempts": 1
					}
				]
			},
			"chargers": {"enabled": true},
			"admins": {"enabled": true}
		}`,
		TpFiles: map[string]string{
			utils.RatesCsv: `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP_SIMPLE,,;10,,,,RT_SIMPLE,*string:~*req.Destination:1002,"* * * * *",;10,false,0s,0,1,1m,1m,`,
		},
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),
	}

	// t.Cleanup(func() {
	// 	if ng.LogBuffer != nil {
	// 		fmt.Println(ng.LogBuffer)
	// 	}
	// })

	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CGR_DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
				Weights:      utils.DynamicWeights{{Weight: 0}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, new(string)); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile failed: %v", err)
	}

	t.Run("setAccount", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
			&utils.AccountWithAPIOpts{
				Account: &utils.Account{
					Tenant:    "cgrates.org",
					ID:        "1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   utils.DynamicWeights{{Weight: 20}},
					Balances: map[string]*utils.Balance{
						"AbstractBalance": {
							ID:             "AbstractBalance",
							Type:           utils.MetaAbstract,
							Weights:        utils.DynamicWeights{{Weight: 10}},
							Units:          utils.NewDecimal(int64(10*time.Minute), 0),
							RateProfileIDs: []string{"RP_SIMPLE"},
						},
						"ConcreteBalance": {
							ID:      "ConcreteBalance",
							Type:    utils.MetaConcrete,
							Weights: utils.DynamicWeights{{Weight: 5}},
							Units:   utils.NewDecimal(10, 0),
						},
					},
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetAccount failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetAccount reply = %s, want OK", reply)
		}
	})

	t.Run("ratesAndEEs", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ratesAndEEs",
				APIOpts: map[string]any{
					utils.MetaRates:    true,
					utils.MetaEEs:      true,
					utils.MetaUsage:    10 * time.Minute,
					utils.MetaOriginID: "OriginID_ratesAndEEs",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent with *rates+*ees failed: %v", err)
		}
		cost, ok := rply.RateSCost[utils.MetaPrimary]
		if !ok {
			t.Fatalf("RateSCost missing *primary, got: %v", rply.RateSCost)
		}
		if cost != 10.0 {
			t.Errorf("RateSCost[*primary] = %g, want 10.0", cost)
		}
		if len(rply.EventExporters) == 0 {
			t.Fatal("EventExporters should not be empty with *ees flag")
		}
		for runID, eesIDs := range rply.EventExporters {
			if len(eesIDs) == 0 {
				t.Errorf("runID %q has no exporter IDs", runID)
			}
		}
	})

	t.Run("accountsAuthorizeRatesAndEEs", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "accountsAuthorizeRatesAndEEs",
				APIOpts: map[string]any{
					utils.MetaAccounts:  true,
					utils.MetaAuthorize: true,
					utils.MetaRates:     true,
					utils.MetaEEs:       true,
					utils.MetaUsage:     10 * time.Minute,
					utils.MetaOriginID:  "OriginID_accountsAuthorizeRatesAndEEs",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent with *accounts+*authorize+*rates+*ees failed: %v", err)
		}
		usage, ok := rply.AccountSUsage[utils.MetaPrimary]
		if !ok {
			t.Fatalf("AccountSUsage missing *default, got: %v", rply.AccountSUsage)
		}
		if usage != 10*time.Minute {
			t.Errorf("AccountSUsage[*default] = %v, want %v", usage, 10*time.Minute)
		}
		cost, ok := rply.RateSCost[utils.MetaPrimary]
		if !ok {
			t.Fatalf("RateSCost missing *default, got: %v", rply.RateSCost)
		}
		if cost != 10.0 {
			t.Errorf("RateSCost[*default] = %g, want 10.0", cost)
		}
		if len(rply.EventExporters) == 0 {
			t.Fatal("EventExporters should not be empty with *ees flag")
		}
		for runID, eesIDs := range rply.EventExporters {
			if len(eesIDs) == 0 {
				t.Errorf("runID %q has no exporter IDs", runID)
			}
		}
	})
}

func TestSessionSv1ProcessEventAttributesAndEEs(t *testing.T) {
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
			"logger": {"level": 7},
			"sessions": {
				"enabled": true,
				"conns": {
					"*attributes": [{"connIDs": ["*localhost"]}],
					"*ees":        [{"connIDs": ["*localhost"]}]
				}
			},
			"attributes": {"enabled": true},
			"ees": {
				"enabled": true,
				"exporters": [
					{
						"id": "LogExporter",
						"type": "*log",
						"attempts": 1
					}
				]
			},
			"chargers": {"enabled": true},
			"admins": {"enabled": true}
		}`,
		DBCfg:    dbcfg,
		Encoding: *utils.Encoding,
	}
	client, _ := ng.Run(t)

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       utils.CGRateSorg,
				ID:           "CGR_DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
				Weights:      utils.DynamicWeights{{Weight: 0}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, new(string)); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile failed: %v", err)
	}

	var reply string

	if err := client.Call(
		context.Background(),
		utils.AdminSv1SetAttributeProfile,
		&utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant:    utils.CGRateSorg,
				ID:        "ATTR_MODIFY_ACCOUNT",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights:   utils.DynamicWeights{{Weight: 10}},
				Attributes: []*utils.ExternalAttribute{
					{
						Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
						Type:  utils.MetaConstant,
						Value: "1002",
					},
				},
			},
		},
		&reply,
	); err != nil {
		t.Fatalf("AdminSv1SetAttributeProfile failed: %v", err)
	}

	t.Run("attributesModifyEvent", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: utils.CGRateSorg,
				ID:     "attributesModifyEvent",
				APIOpts: map[string]any{
					utils.MetaAttributes: true,
					utils.MetaEEs:        true,
					utils.MetaOriginID:   "OriginID_attributesModifyEvent",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent with *attributes+*ees failed: %v", err)
		}
		if len(rply.Attributes) == 0 {
			t.Fatal("Attributes should not be empty with *attributes flag")
		}
		attrRply, ok := rply.Attributes[utils.MetaPrimary]
		if !ok {
			t.Fatalf("Attributes missing *primary runID, got: %v", rply.Attributes)
		}
		if len(attrRply.AlteredFields) == 0 {
			t.Fatal("AlteredFields should not be empty")
		}
		altered := attrRply.AlteredFields[0]
		if altered.MatchedProfileID != utils.ConcatenatedKey(utils.CGRateSorg, "ATTR_MODIFY_ACCOUNT") {
			t.Errorf("MatchedProfileID = %q, want %q", altered.MatchedProfileID,
				utils.ConcatenatedKey(utils.CGRateSorg, "ATTR_MODIFY_ACCOUNT"))
		}
		if !slices.Contains(altered.Fields, utils.MetaReq+utils.NestingSep+utils.AccountField) {
			t.Errorf("expected %q in Fields, got: %v",
				utils.MetaReq+utils.NestingSep+utils.AccountField, altered.Fields)
		}
		if len(rply.EventExporters) == 0 {
			t.Fatal("EventExporters should not be empty with *ees flag")
		}
		for runID, eesIDs := range rply.EventExporters {
			if len(eesIDs) == 0 {
				t.Errorf("runID %q has no exporter IDs", runID)
			}
		}
	})

	t.Run("attributesNoMatch", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: utils.CGRateSorg,
				ID:     "attributesNoMatch",
				APIOpts: map[string]any{
					utils.MetaAttributes: true,
					utils.MetaEEs:        true,
					utils.MetaOriginID:   "OriginID_attributesNoMatch",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "1003",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent attributesNoMatch failed: %v", err)
		}
		if rply.Attributes != nil {
			t.Errorf("Attributes should be nil when no profile matches, got: %v", rply.Attributes)
		}
		if len(rply.EventExporters) == 0 {
			t.Fatal("EventExporters should still be populated even when attributes do not match")
		}
		for runID, eesIDs := range rply.EventExporters {
			if len(eesIDs) == 0 {
				t.Errorf("runID %q has no exporter IDs", runID)
			}
		}
	})
}

func TestSessionSv1ProcessEventStatsAndThresholds(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
            "logger": {"level": 7},
            "sessions": {
                "enabled": true,
                "conns": {
                    "*stats":      [{"ConnIDs": ["*localhost"]}],
                    "*thresholds": [{"ConnIDs": ["*localhost"]}],
                    "*ees":        [{"ConnIDs": ["*localhost"]}]
                }
            },
            "stats": {"enabled": true},
            "thresholds": {
                "enabled": true,
                "conns": {
                    "*actions": [{"ConnIDs": ["*localhost"]}]
                }
            },
            "actions": {"enabled": true},
            "ees": {
                "enabled": true,
                "exporters": [
                    {"id": "LogExporter", "type": "*log", "attempts": 1}
                ]
            },
            "chargers": {"enabled": true},
            "admins":   {"enabled": true}
        }`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}

	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	var reply string

	t.Run("setChargerProfile", func(t *testing.T) {
		if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
			&utils.ChargerProfileWithAPIOpts{
				ChargerProfile: &utils.ChargerProfile{
					Tenant:       "cgrates.org",
					ID:           "CGR_DEFAULT",
					RunID:        utils.MetaDefault,
					AttributeIDs: []string{utils.MetaNone},
					Weights:      utils.DynamicWeights{{Weight: 0}},
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetChargerProfile failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetChargerProfile reply = %s, want OK", reply)
		}
	})

	t.Run("setStatQueueProfile", func(t *testing.T) {
		if err := client.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
			&utils.StatQueueProfileWithAPIOpts{
				StatQueueProfile: &utils.StatQueueProfile{
					Tenant:      "cgrates.org",
					ID:          "SQ_1001",
					FilterIDs:   []string{"*string:~*req.Account:1001"},
					Weights:     utils.DynamicWeights{{Weight: 10}},
					QueueLength: 10,
					TTL:         time.Hour,
					Metrics: []*utils.MetricWithFilters{
						{MetricID: utils.MetaTCD},
					},
					Stored: true,
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetStatQueueProfile failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetStatQueueProfile reply = %s, want OK", reply)
		}
	})

	t.Run("setThresholdProfile", func(t *testing.T) {
		if err := client.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
			&utils.ThresholdProfileWithAPIOpts{
				ThresholdProfile: &utils.ThresholdProfile{
					Tenant:    "cgrates.org",
					ID:        "THD_1001",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   utils.DynamicWeights{{Weight: 10}},
					MaxHits:   -1,
					Async:     true,
				},
			}, &reply); err != nil {
			t.Fatalf("AdminSv1SetThresholdProfile failed: %v", err)
		} else if reply != utils.OK {
			t.Fatalf("AdminSv1SetThresholdProfile reply = %s, want OK", reply)
		}
	})

	t.Run("statsAndThresholdsFlags", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "statsAndThresholdsEvent",
				APIOpts: map[string]any{
					utils.MetaStats:      true,
					utils.MetaThresholds: true,
					utils.MetaEEs:        true,
					utils.MetaOriginID:   "OriginID_statsAndThresholds",
					utils.MetaUsage:      10 * time.Minute,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.Usage:        10 * time.Minute,
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent with *stats+*thresholds failed: %v", err)
		}
		if len(rply.StatQueueIDs) == 0 {
			t.Error("StatQueueIDs should be populated with *stats flag")
		}
		sqIDs, ok := rply.StatQueueIDs[utils.MetaPrimary]
		if !ok || !slices.Contains(sqIDs, "SQ_1001") {
			t.Errorf("StatQueueIDs = %v, want to contain SQ_1001 under *primary", rply.StatQueueIDs)
		}
		if len(rply.ThresholdIDs) == 0 {
			t.Error("ThresholdIDs should be populated with *thresholds flag")
		}
		thIDs, ok := rply.ThresholdIDs[utils.MetaPrimary]
		if !ok || !slices.Contains(thIDs, "THD_1001") {
			t.Errorf("ThresholdIDs = %v, want to contain THD_1001 under *primary", rply.ThresholdIDs)
		}
		if len(rply.EventExporters) == 0 {
			t.Fatal("EventExporters should not be empty with *ees flag")
		}
	})

	t.Run("noMatchStatsAndThresholds", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchEvent",
				APIOpts: map[string]any{
					utils.MetaStats:      true,
					utils.MetaThresholds: true,
					utils.MetaOriginID:   "OriginID_noMatch",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "1002",
					utils.Usage:        10 * time.Minute,
				},
			}, &rply)
		if err != nil {
			t.Logf("ProcessEvent noMatch returned: %v (expected when no profiles match)", err)
		}
		if len(rply.StatQueueIDs) != 0 {
			t.Errorf("StatQueueIDs should be empty when no profile matches, got: %v", rply.StatQueueIDs)
		}
		if len(rply.ThresholdIDs) != 0 {
			t.Errorf("ThresholdIDs should be empty when no profile matches, got: %v", rply.ThresholdIDs)
		}
	})
}
