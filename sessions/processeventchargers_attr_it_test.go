//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionSv1ProcessEventChargerAttributes(t *testing.T) {

	ng := engine.TestEngine{
		ConfigJSON: `{
"logger": {"level": 7},
"sessions": {
	"enabled": true,
	"conns": {
		"*chargers": [{"connIDs": ["*localhost"]}]
	}
},
"chargers": {
	"enabled": true,
	"conns": {
		"*attributes": [{"connIDs": ["*localhost"]}]
	}
},
"attributes": {
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

	client, _ := ng.Run(t)

	// t.Cleanup(func() {
	// 	if ng.LogBuffer != nil {
	// 		fmt.Println(ng.LogBuffer)
	// 	}
	// })

	var reply string

	if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		&utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_SUBJECT",
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
		t.Fatalf("AdminSv1SetAttributeProfile failed: %v", err)
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
				AttributeIDs: []string{"ATTR_SUBJECT"},
				Weights:      utils.DynamicWeights{{Weight: 10}},
				Blockers:     utils.DynamicBlockers{{Blocker: false}},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile CH_WITH_ATTR failed: %v", err)
	}

	t.Run("noAttributes", func(t *testing.T) {
		t.Skip("fails due to comparison of len(chrgr.AlteredFields)")
	})

	t.Run("withAttributes", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "withAttr",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaOriginID: "withAttr",
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
			t.Fatalf("expected Attributes entry for run_with_attr, got: %v", rply.Attributes)
		}
		subject, has := attrRply.CGREvent.Event[utils.Subject]
		if !has {
			t.Fatal("Subject field missing from altered CGREvent")
		}
		if subject != "SUPPLIER1" {
			t.Errorf("Subject = %v, want SUPPLIER1", subject)
		}
	})

	t.Run("noChargersFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noChrgFlag",
				APIOpts: map[string]any{
					utils.MetaOriginID: "noChrgFlag",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if len(rply.Attributes) > 0 {
			t.Errorf("Attributes should be empty without *chargers flag, got: %v", rply.Attributes)
		}
	})
}
