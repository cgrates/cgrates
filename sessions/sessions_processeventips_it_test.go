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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/birpc/context"
)

func TestSessionSv1ProcessEventIPsAuthorize(t *testing.T) {
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
	"ips_conns": ["*localhost"]
},
"ips": {
	"enabled": true,
	"store_interval": "-1",
	"indexed_selects": true,
	"string_indexed_fields": ["*req.Account"]
},
"admins": {
	"enabled": true
}
}`,
		TpFiles: map[string]string{
			utils.IPsCsv: `#Tenant[0],ID[1],FilterIDs[2],Weights[3],TTL[4],Stored[5],PoolID[6],PoolFilterIDs[7],PoolType[8],PoolRange[9],PoolStrategy[10],PoolMessage[11],PoolWeights[12],PoolBlockers[13]
cgrates.org,IPs1,*string:~*req.Account:1001,;10,1s,true,,,,,,,,
cgrates.org,IPs1,,,,,POOL1,*string:~*req.Destination:2001,*ipv4,172.16.1.1/32,*ascending,alloc_success,;15,`,
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
				Tenant:  "cgrates.org",
				ID:      "noFlags",
				APIOpts: map[string]any{},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed without IPs flags: %v", err)
		}

		if len(rply.IPsAllocation) > 0 {
			t.Errorf("IPsAllocation should be empty without IPs flags, got: %v",
				rply.IPsAllocation)
		}
	})

	t.Run("noAuthMatchingProfile", func(t *testing.T) {
		var reply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "NoAuthMatchingProfile",
				APIOpts: map[string]any{
					utils.MetaIPs: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &reply)

		if err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}

		if len(reply.IPsAllocation) > 0 {
			t.Errorf("IPsAllocation should be empty without authorization flags, got: %v",
				reply.IPsAllocation)
		}
	})

	t.Run("flagsIpsAuthorize", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "flagsIpsAuthorize",
				APIOpts: map[string]any{
					utils.MetaIPs:             true,
					utils.MetaAuthorize:       true,
					utils.OptsIPsAllocationID: "testallocid11",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed with *ips + *authorize flags: %v", err)
		}

		if rply.IPsAllocation == nil {
			t.Fatal("IPsAllocation should not be nil with *ips + *authorize flags")
		}

		authorizedIP, exists := rply.IPsAllocation[utils.MetaDefault]
		if !exists {
			t.Fatal("No IP authorization for *default runID with *ips + *authorize flags")
		}

		if authorizedIP.Address.String() != "172.16.1.1" {
			t.Errorf("Authorized IP = %s, want 172.16.1.1", authorizedIP.Address.String())
		}
		if authorizedIP.ProfileID != "IPs1" {
			t.Errorf("ProfileID = %s, want IPs1", authorizedIP.ProfileID)
		}
		if authorizedIP.PoolID != "POOL1" {
			t.Errorf("PoolID = %s, want POOL1", authorizedIP.PoolID)
		}
	})

	t.Run("ipsAuthorizeFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ipsAuthorizeFlag",
				APIOpts: map[string]any{
					utils.MetaIPsAuthorizeCfg: true,
					utils.OptsIPsAllocationID: "testallocid101",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed with *ipsAuthorize flag: %v", err)
		}

		if rply.IPsAllocation == nil {
			t.Fatal("IPsAllocation should not be nil with *ipsAuthorize flag")
		}

		authorizedIP, exists := rply.IPsAllocation[utils.MetaDefault]
		if !exists {
			t.Fatal("No IP authorization for *default runID with *ipsAuthorize flag")
		}

		if authorizedIP.Address.String() != "172.16.1.1" {
			t.Errorf("Authorized IP = %s, want 172.16.1.1", authorizedIP.Address.String())
		}
		if authorizedIP.ProfileID != "IPs1" {
			t.Errorf("ProfileID = %s, want IPs1", authorizedIP.ProfileID)
		}
		if authorizedIP.PoolID != "POOL1" {
			t.Errorf("PoolID = %s, want POOL1", authorizedIP.PoolID)
		}
	})

	t.Run("AuthMatchPrf", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "AuthMatchPrf",
				APIOpts: map[string]any{
					utils.MetaIPsAuthorizeCfg: true,
					utils.OptsIPsAllocationID: "testallocid21",
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

		if rply.IPsAllocation == nil {
			t.Fatal("IPsAllocation should not be nil for matching profile")
		}

		authorizedIP, exists := rply.IPsAllocation[utils.MetaDefault]
		if !exists {
			t.Fatal("No IP authorization found for matching profile")
		}

		if authorizedIP.ProfileID != "IPs1" {
			t.Errorf("ProfileID = %s, want IPs1", authorizedIP.ProfileID)
		}
		if authorizedIP.PoolID != "POOL1" {
			t.Errorf("PoolID = %s, want POOL1", authorizedIP.PoolID)
		}
		if authorizedIP.Address.String() != "172.16.1.1" {
			t.Errorf("Authorized IP = %s, want 172.16.1.1", authorizedIP.Address.String())
		}
		if authorizedIP.Message != "alloc_success" {
			t.Errorf("Message = %s, want alloc_success", authorizedIP.Message)
		}
	})

	t.Run("AuthNoMatchPrfl", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "AuthNoMatchPrfl",
				APIOpts: map[string]any{
					utils.MetaIPsAuthorizeCfg: true,
					utils.OptsIPsAllocationID: "testallocid212",
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "9999",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}

		if len(rply.IPsAllocation) > 0 {
			authorizedIP, exists := rply.IPsAllocation[utils.MetaDefault]
			if exists && authorizedIP.Address.IsValid() {
				t.Errorf("IPsAllocation should have no valid IP when no profile matches, got IP: %v",
					authorizedIP.Address)
			}
		}
	})
}
