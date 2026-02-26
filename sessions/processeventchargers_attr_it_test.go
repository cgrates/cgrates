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
	"time"

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
    "chargers_conns": ["*localhost"],
},
"chargers": {
    "enabled": true,
    "attributes_conns": ["*localhost"]
},
"attributes": {
    "enabled": true
},
"admins": {
    "enabled": true
}
}`,
		TpFiles: map[string]string{
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,RunID,AttributeIDs
cgrates.org,CH_NO_ATTR,*string:~*req.Destination:1001,;10,,run_no_attr,*none
cgrates.org,CH_WITH_ATTR,*string:~*req.Destination:1002,;10,,run_with_attr,ATTR_SUBJECT`,

			utils.AttributesCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,AttributeFilterIDs,AttributeBlockers,Path,Type,Value
cgrates.org,ATTR_SUBJECT,*string:~*opts.*context:*chargers,;10,;false,,,*req.Subject,*constant,SUPPLIER1`,
		},
		DBCfg:    dbcfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),
	}
	// t.Cleanup(func() {
	// 	if ng.LogBuffer != nil {
	// 		fmt.Println(ng.LogBuffer)
	// 	}
	// })

	client, _ := ng.Run(t)
	time.Sleep(500 * time.Millisecond)

	// CH_NO_ATTR has AttributeIDs=*none, ChargerS skips AttributeS.
	// No attr profile modified the event, so run_no_attr should not appear in apiRply.Attributes.
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

	// CH_WITH_ATTR calls ATTR_SUBJECT which modifies the charger event.
	// Only run_with_attr should appear in apiRply.Attributes.
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
		attrRply, exists := rply.Attributes["run_with_attr"]
		if !exists {
			t.Fatalf("run_with_attr should appear in apiRply.Attributes when attr profile modified the event, got: %v", rply.Attributes)
		}
		if attrRply.CGREvent == nil {
			t.Fatal("CGREvent should not be nil")
		}
		// verify the event was modified with the expected value
		if attrRply.CGREvent.Event[utils.Subject] != "SUPPLIER1" {
			t.Errorf("Subject = %v, want SUPPLIER1", attrRply.CGREvent.Event[utils.Subject])
		}
		// verify only the profiles that modified the charger event appear in AlteredFields:
		// *default (ChargerS sets runID/chargeID/subsys) + cgrates.org:ATTR_SUBJECT (sets Subject)
		if len(attrRply.AlteredFields) != 2 {
			t.Errorf("expected 2 AlteredFields entries (*default + cgrates.org:ATTR_SUBJECT), got: %v", attrRply.AlteredFields)
		}
		// verify cgrates.org:ATTR_SUBJECT modified only *req.Subject
		if attrRply.AlteredFields[1].MatchedProfileID != "cgrates.org:ATTR_SUBJECT" {
			t.Errorf("expected cgrates.org:ATTR_SUBJECT, got: %v", attrRply.AlteredFields[1].MatchedProfileID)
		}
		if len(attrRply.AlteredFields[1].Fields) != 1 || attrRply.AlteredFields[1].Fields[0] != utils.MetaReq+utils.NestingSep+utils.Subject {
			t.Errorf("expected [*req.Subject], got: %v", attrRply.AlteredFields[1].Fields)
		}
	})

	// without *chargers flag, ChargerS is not called at all.
	// no charger events are processed, apiRply.Attributes should be empty.
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
}
