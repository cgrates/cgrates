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

func TestSessionSv1ProcessEventResourcesAuthorize(t *testing.T) {
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
"logger": { "level": 7 },
"sessions": {
	"enabled": true,
	"resources_conns": ["*localhost"]
},
"resources": {
	"enabled": true,
	"store_interval": "-1"
},
"admins": { "enabled": true }
}`,
		TpFiles: map[string]string{
			utils.ResourcesCsv: `#Tenant[0],Id[1],FilterIDs[2],Weights[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],ThresholdIDs[9]
cgrates.org,RES1,*string:~*req.Account:1001,;10,1h,3,ResourceAllocationSuccess,false,true,`,
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

	t.Run("noFlags", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noFlags",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply)

		if err != nil {
			t.Fatal(err)
		}
		if len(rply.ResourceAllocation) != 0 {
			t.Fatalf("expected no allocation, got %v", rply.ResourceAllocation)
		}
	})

	t.Run("resourcesOnly", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "resourcesOnly",
				APIOpts: map[string]any{
					utils.MetaResources: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply)

		if err != nil {
			t.Fatal(err)
		}
		if len(rply.ResourceAllocation) != 0 {
			t.Fatalf("expected no allocation, got %v", rply.ResourceAllocation)
		}
	})

	t.Run("authorizeAndResources", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "authorizeAndResources",
				APIOpts: map[string]any{
					utils.MetaAuthorize:        true,
					utils.MetaResources:        true,
					utils.OptsResourcesUsageID: "usage1",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply)

		if err != nil {
			t.Fatal(err)
		}

		if msg := rply.ResourceAllocation[utils.MetaDefault]; msg != "ResourceAllocationSuccess" {
			t.Fatalf("unexpected allocation msg: %q", msg)
		}
	})

	t.Run("resourcesAuthorizeFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "resourcesAuthorizeFlag",
				APIOpts: map[string]any{
					utils.MetaResourcesAuthorizeCfg: true,
					utils.OptsResourcesUsageID:      "2",
					utils.OptsResourcesUnits:        1,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply)

		if err != nil {
			t.Fatal(err)
		}

		if msg := rply.ResourceAllocation[utils.MetaDefault]; msg != "ResourceAllocationSuccess" {
			t.Fatalf("unexpected allocation msg: %q", msg)
		}
	})

	t.Run("missingUsageID", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "missingUsageID",
				APIOpts: map[string]any{
					utils.MetaResourcesAuthorizeCfg: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, exists := rply.ResourceAllocation[utils.MetaDefault]; !exists {
			t.Fatalf("expected allocation entry")
		}
	})

	t.Run("noMatchingProfile", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingProfile",
				APIOpts: map[string]any{
					utils.MetaResourcesAuthorizeCfg: true,
					utils.OptsResourcesUsageID:      "usage-nomatch",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
				},
			}, &rply)

		if err != nil {
			t.Fatal(err)
		}

		if msg := rply.ResourceAllocation[utils.MetaDefault]; msg != "" {
			t.Fatalf("expected empty allocation msg, got %q", msg)
		}
	})
}
