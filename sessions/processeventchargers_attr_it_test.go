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
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionSv1ProcessEventChargerAttributes(t *testing.T) {
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
		"*chargers": [{"ConnIDs": ["*localhost"]}]
	}
},
"chargers": {
	"enabled": true,
	"conns": {
		"*attributes": [{"ConnIDs": ["*localhost"]}]
	}
},
"attributes": {
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
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)

	var reply string

	if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		&utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_SUPPLIER1",
				FilterIDs: []string{"*string:~*opts.*context:*chargers"},
				Weights:   utils.DynamicWeights{{Weight: 10}},
				Attributes: []*utils.ExternalAttribute{
					{
						Path:  "*req.Subject",
						Type:  utils.MetaConstant,
						Value: "SUPPLIER1",
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAttributeProfile ATTR_SUPPLIER1 failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		&utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_SUPPLIER2_PREPAID",
				FilterIDs: []string{"*string:~*opts.*context:*chargers"},
				Weights:   utils.DynamicWeights{{Weight: 10}},
				Attributes: []*utils.ExternalAttribute{
					{
						Path:  "*req.Subject",
						Type:  utils.MetaConstant,
						Value: "SUPPLIER2",
					},
					{
						Path:  "*req.RequestType",
						Type:  utils.MetaConstant,
						Value: "*prepaid",
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAttributeProfile ATTR_SUPPLIER2_PREPAID failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		&utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_PREPAID_ON_SUPPLIER1",
				FilterIDs: []string{"*string:~*opts.*context:*chargers", "*string:~*req.Subject:SUPPLIER1"},
				Weights:   utils.DynamicWeights{{Weight: 10}},
				Attributes: []*utils.ExternalAttribute{
					{
						Path:  "*req.RequestType",
						Type:  utils.MetaConstant,
						Value: "*prepaid",
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAttributeProfile ATTR_PREPAID_ON_SUPPLIER1 failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CH_NO_ATTR",
				FilterIDs:    []string{"*string:~*req.Destination:1001"},
				RunID:        "run_no_attr",
				AttributeIDs: []string{utils.MetaNone},
				Weights:      utils.DynamicWeights{{Weight: 10}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile CH_NO_ATTR failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CH_WITH_ATTR",
				FilterIDs:    []string{"*string:~*req.Destination:1002"},
				RunID:        "run_with_attr",
				AttributeIDs: []string{"ATTR_SUPPLIER1"},
				Weights:      utils.DynamicWeights{{Weight: 10}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile CH_WITH_ATTR failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CH_MULTI_ATTR",
				FilterIDs:    []string{"*string:~*req.Destination:1003"},
				RunID:        "run_multi_attr",
				AttributeIDs: []string{"ATTR_SUPPLIER2_PREPAID"},
				Weights:      utils.DynamicWeights{{Weight: 10}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile CH_MULTI_ATTR failed: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CH_PROCESS_RUNS",
				FilterIDs:    []string{"*string:~*req.Destination:1004"},
				RunID:        "run_process_runs",
				AttributeIDs: []string{"ATTR_SUPPLIER1", "ATTR_PREPAID_ON_SUPPLIER1"},
				Weights:      utils.DynamicWeights{{Weight: 10}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile CH_PROCESS_RUNS failed: %v", err)
	}

	t.Run("noAttributes", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noAttr",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1001",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if _, exists := rply.Attributes["run_no_attr"]; exists {
			t.Errorf("run_no_attr should not appear in apiRply.Attributes when no attr profile modified the event, got: %v", rply.Attributes)
		}
	})

	t.Run("withAttributes", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withAttr",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if rply.Attributes == nil {
			t.Fatal("Attributes should not be nil")
		}
		attrRply, exists := rply.Attributes["run_with_attr"]
		if !exists {
			t.Fatalf("run_with_attr should appear in apiRply.Attributes, got: %v", rply.Attributes)
		}
		subject, has := attrRply.CGREvent.Event[utils.Subject]
		if !has {
			t.Fatal("Subject field missing from altered CGREvent")
		}
		if subject != "SUPPLIER1" {
			t.Errorf("Subject = %v, want SUPPLIER1", subject)
		}
		if len(attrRply.AlteredFields) != 1 {
			t.Errorf("expected 1 AlteredFields entry, got: %v", attrRply.AlteredFields)
		}
		if attrRply.AlteredFields[0].MatchedProfileID != "cgrates.org:ATTR_SUPPLIER1" {
			t.Errorf("expected cgrates.org:ATTR_SUPPLIER1, got: %v", attrRply.AlteredFields[0].MatchedProfileID)
		}
		if len(attrRply.AlteredFields[0].Fields) != 4 {
			t.Errorf("expected 4 fields (*opts.*runID, *opts.*chargeID, *opts.*subsys, *req.Subject), got: %v", attrRply.AlteredFields[0].Fields)
		}
	})

	t.Run("noChargersFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant:  "cgrates.org",
				ID:      "noChrgFlag",
				APIOpts: map[string]any{},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if len(rply.Attributes) > 0 {
			t.Errorf("apiRply.Attributes should be empty when *chargers flag is not set, got: %v", rply.Attributes)
		}
	})

	t.Run("multipleAttributes", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "multiAttr",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1003",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		attrRply, exists := rply.Attributes["run_multi_attr"]
		if !exists {
			t.Fatalf("run_multi_attr should appear in apiRply.Attributes, got: %v", rply.Attributes)
		}
		if len(attrRply.AlteredFields) != 1 {
			t.Errorf("expected 1 AlteredFields entry, got: %v", attrRply.AlteredFields)
		}
		if attrRply.AlteredFields[0].MatchedProfileID != "cgrates.org:ATTR_SUPPLIER2_PREPAID" {
			t.Errorf("expected cgrates.org:ATTR_SUPPLIER2_PREPAID, got: %v", attrRply.AlteredFields[0].MatchedProfileID)
		}
		if len(attrRply.AlteredFields[0].Fields) != 5 {
			t.Errorf("expected 5 fields (*opts.*runID, *opts.*chargeID, *opts.*subsys, *req.Subject, *req.RequestType), got: %v", attrRply.AlteredFields[0].Fields)
		}
		subject, has := attrRply.CGREvent.Event[utils.Subject]
		if !has {
			t.Fatal("Subject field missing from altered CGREvent")
		}
		if subject != "SUPPLIER2" {
			t.Errorf("Subject = %v, want SUPPLIER2", subject)
		}
		reqType, has := attrRply.CGREvent.Event[utils.RequestType]
		if !has {
			t.Fatal("RequestType field missing from altered CGREvent")
		}
		if reqType != "*prepaid" {
			t.Errorf("RequestType = %v, want *prepaid", reqType)
		}
	})

	t.Run("processRuns", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "processRuns",
				APIOpts: map[string]any{
					utils.MetaChargers:              true,
					utils.OptsAttributesProcessRuns: 2,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1004",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		attrRply, exists := rply.Attributes["run_process_runs"]
		if !exists {
			t.Fatalf("run_process_runs should appear in apiRply.Attributes, got: %v", rply.Attributes)
		}
		if len(attrRply.AlteredFields) != 2 {
			t.Errorf("expected 2 AlteredFields entries, got: %v", attrRply.AlteredFields)
		}
		if len(attrRply.AlteredFields[0].Fields) != 4 {
			t.Errorf("expected 4 fields in first entry (*opts.*runID, *opts.*chargeID, *opts.*subsys, *req.Subject), got: %v", attrRply.AlteredFields[0].Fields)
		}
		if len(attrRply.AlteredFields[1].Fields) != 1 {
			t.Errorf("expected 1 field in second entry (*req.RequestType), got: %v", attrRply.AlteredFields[1].Fields)
		}
	})
}
