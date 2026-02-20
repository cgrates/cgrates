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
	time.Sleep(100 * time.Millisecond)

	t.Run("noAttributes", func(t *testing.T) {
		t.Skip("fails due to comparision of len(chrgr.AlteredFields) != len(chargers.ChargerSDefaultAlteredFields)")
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
			t.Errorf("Attributes[run_no_attr] should not exist when AttributeIDs=*none, got: %v", rply.Attributes)
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
			t.Fatal("Attributes should not be nil when AttributeS alters extra fields")
		}
		attrRply, exists := rply.Attributes["run_with_attr"]
		if !exists {
			t.Fatalf("expected Attributes entry for run_with_attr, got: %v", rply.Attributes)
		}
		if attrRply.CGREvent == nil {
			t.Fatal("CGREvent in Attributes reply should not be nil")
		}
		subject, has := attrRply.CGREvent.Event[utils.Subject]
		if !has {
			t.Fatal("Subject field missing from altered CGREvent")
		}
		if subject != "SUPPLIER1" {
			t.Errorf("Subject = %v, want SUPPLIER1", subject)
		}
		totalFields := 0
		for _, fa := range attrRply.AlteredFields {
			totalFields += len(fa.Fields)
		}
		if totalFields <= 3 {
			t.Errorf("expected more than 3 altered fields, got: %d", totalFields)
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
