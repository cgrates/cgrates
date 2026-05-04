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
                    "*routes": [{"ConnIDs": ["*localhost"]}]
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
		"*stats": [{"Tenant":"","FilterIDs":[],"ConnIDs":["*localhost"]}]
	}
},
"stats": {
	"enabled": true,
	"store_interval": "-1",
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
		&engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &engine.StatQueueProfile{
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
				Metrics: []*engine.MetricWithFilters{
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
				Tenant:  "cgrates.org",
				ID:      "noFlags",
				APIOpts: map[string]any{},
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
					utils.MetaStats: true,
					utils.MetaCost:  0.6,
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
					utils.MetaStats: true,
					utils.MetaCost:  1.2,
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
					utils.MetaStats: true,
					utils.MetaCost:  0.3,
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
		"*thresholds": [{"Tenant":"","FilterIDs":[],"ConnIDs":["*localhost"]}]
	}
},
"thresholds": {
	"enabled": true,
	"store_interval": "-1",
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
		&engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &engine.ThresholdProfile{
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
				Tenant:  "cgrates.org",
				ID:      "noFlags",
				APIOpts: map[string]any{},
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

func TestSessionSv1ProcessEventCDRs(t *testing.T) {
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
		"*cdrs":    [{"ConnIDs": ["*localhost"]}],
		"*chargers":[{"ConnIDs": ["*localhost"]}]
	}
},
"cdrs": {
	"enabled": true,
	"conns": {
		"*default": [{"ConnIDs": ["*localhost"]}]
	}
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

	t.Run("noFlags", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant:  "cgrates.org",
				ID:      "noCDRsFlag",
				APIOpts: map[string]any{},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2024-01-10T10:00:00Z",
					utils.AnswerTime:   "2024-01-10T10:00:01Z",
					utils.Usage:        "60s",
					utils.OriginID:     "noCDRsFlag_orig",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed without *cdrs flag: %v", err)
		}

		var cdrs []*utils.CDR
		experr := "retrieving CDRs failed: NOT_FOUND"
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs,
			&utils.CDRFilters{
				Tenant:    "cgrates.org",
				ID:        "checkNoCDR",
				FilterIDs: []string{"*string:~*req.OriginID:noCDRsFlag_orig"},
			}, &cdrs); err == nil || err.Error() != experr {
			t.Errorf("expected err <%v>, received <%v>", experr, err)
		}
	})

	t.Run("withCDRsFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withCDRsFlag",
				APIOpts: map[string]any{
					utils.MetaCDRs: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2024-01-10T10:00:00Z",
					utils.AnswerTime:   "2024-01-10T10:00:01Z",
					utils.Usage:        "60s",
					utils.OriginID:     "withCDRsFlag_orig",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent with *cdrs flag failed: %v", err)
		}

		var cdrs []*utils.CDR
		if err := client.Call(context.Background(), utils.AdminSv1GetCDRs,
			&utils.CDRFilters{
				Tenant:    "cgrates.org",
				ID:        "getCDRwithFlag",
				FilterIDs: []string{"*string:~*req.OriginID:withCDRsFlag_orig"},
			}, &cdrs); err != nil {
			t.Fatalf("AdminSv1GetCDRs failed after *cdrs flag: %v", err)
		}
		if len(cdrs) == 0 {
			t.Fatal("expected at least 1 CDR stored, got 0")
		}
		if cdrs[0].Event[utils.OriginID] != "withCDRsFlag_orig" {
			t.Errorf("CDR OriginID = %v, want withCDRsFlag_orig", cdrs[0].Event[utils.OriginID])
		}
	})

	t.Run("withCDRsFlagBlockerError", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "cdrBlocker",
				APIOpts: map[string]any{
					utils.MetaCDRs:            true,
					utils.OptsSesBlockerError: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2024-01-10T11:00:00Z",
					utils.AnswerTime:   "2024-01-10T11:00:01Z",
					utils.Usage:        "30s",
					utils.OriginID:     "withCDRsFlag_orig",
				},
			}, &rply)

		if err == nil {
			t.Error("expected an error with OptsSesBlockerError=true on duplicate CDR, got nil")
		}
	})

	t.Run("withCDRsFlagNonBlockerError", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "cdrNonBlocker",
				APIOpts: map[string]any{
					utils.MetaCDRs:            true,
					utils.OptsSesBlockerError: false,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2024-01-10T11:00:00Z",
					utils.AnswerTime:   "2024-01-10T11:00:01Z",
					utils.Usage:        "30s",
					utils.OriginID:     "withCDRsFlag_orig",
				},
			}, &rply)

		if err == nil {
			t.Error("expected PARTIALLY_EXECUTED, got nil")
		} else if err.Error() != utils.ErrPartiallyExecuted.Error() {
			t.Errorf("expected PARTIALLY_EXECUTED, got: %v", err)
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
		"*chargers": [{"ConnIDs": ["*localhost"]}],
		"*routes":   [{"ConnIDs": ["*localhost"]}]
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
			t.Errorf("expected RouteProfiles[*default] to be absent when *session=true, got: %v", rply.RouteProfiles)
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
					"*rates":    [{"ConnIDs": ["*localhost"]}],
					"*chargers": [{"ConnIDs": ["*localhost"]}]
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
				Tenant:  "cgrates.org",
				ID:      "noRatesFlagEmptyReply",
				APIOpts: map[string]any{},
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
					utils.MetaRates: true,
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
