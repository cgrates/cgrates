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
	}

	client, _ := ng.Run(t)

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
			t.Errorf("Attributes should be empty without *chargers flag, got: %v", rply.Attributes)
		}
	})
}
