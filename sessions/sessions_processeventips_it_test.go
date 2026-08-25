//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/birpc/context"
)

func TestSessionSv1ProcessEventIPsAuthorize(t *testing.T) {

	ng := engine.TestEngine{
		ConfigJSON: `{
"logger": {
	"level": 7
},
"sessions": {
	"enabled": true,
	"conns": {
			"*ips": [{"tenant":"","filterIDs":[],"connIDs":["*localhost"]}]
		},
	"opts": {
	}
},
"ips": {
	"enabled": true,
	"storeInterval": "-1",
	"indexedSelects": true,
	"stringIndexedFields": ["*req.Account"]
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
					utils.MetaIPs:      true,
					utils.MetaOriginID: "OriginID",
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
					utils.MetaOriginID:        "OriginID",
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

		authorizedIP, exists := rply.IPsAllocation[utils.MetaPrimary]
		if !exists {
			t.Fatal("No IP authorization for *primary runID with *ips + *authorize flags")
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
					utils.MetaOriginID:        "OriginID",
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

		authorizedIP, exists := rply.IPsAllocation[utils.MetaPrimary]
		if !exists {
			t.Fatal("No IP authorization for *primary runID with *ipsAuthorize flag")
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
					utils.MetaOriginID:        "OriginID",
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

		authorizedIP, exists := rply.IPsAllocation[utils.MetaPrimary]
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
					utils.MetaOriginID:        "OriginID",
					utils.OptsSesBlockerError: true,
				},
				Event: map[string]any{
					utils.AccountField: "9999",
					utils.Destination:  "9999",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Fatalf("ProcessEvent expected NOT_FOUND error, got: %v", err)
		}
	})
}

func TestSessionSv1ProcessEventIPsAllocateRelease(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
"logger": {
	"level": 7
},
"sessions": {
	"enabled": true,
	"conns": {
		"*ips": [{"tenant":"","filterIDs":[],"connIDs":["*localhost"]}]
	},
	"opts": {
	}
},
"ips": {
	"enabled": true,
	"storeInterval": "-1",
	"indexedSelects": true,
	"stringIndexedFields": ["*req.Account"]
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
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}

	client, _ := ng.Run(t)
	time.Sleep(100 * time.Millisecond)

	getAllocations := func(t *testing.T) utils.IPAllocations {
		t.Helper()
		var allocations utils.IPAllocations
		if err := client.Call(context.Background(), utils.IPsV1GetIPAllocations,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "IPs1",
				},
			}, &allocations); err != nil {
			t.Fatalf("GetIPAllocations failed: %v", err)
		}
		return allocations
	}

	allocationID := "ProcessEventAllocation"
	t.Run("allocate", func(t *testing.T) {
		var reply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ipsAllocate",
				APIOpts: map[string]any{
					utils.MetaIPsAllocateCfg:  true,
					utils.MetaOriginID:        allocationID,
					utils.OptsSesBlockerError: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &reply)
		if err != nil {
			t.Fatalf("ProcessEvent allocation failed: %v", err)
		}

		allocatedIP := reply.IPsAllocation[utils.MetaPrimary]
		if allocatedIP == nil {
			t.Fatal("ProcessEvent allocation returned no IP")
		}
		if allocatedIP.ProfileID != "IPs1" || allocatedIP.PoolID != "POOL1" ||
			allocatedIP.Address.String() != "172.16.1.1" {
			t.Errorf("unexpected allocated IP: %s", utils.ToJSON(allocatedIP))
		}

		allocations := getAllocations(t)
		if len(allocations.Allocations) != 1 {
			t.Fatalf("stored allocations = %s, want one", utils.ToJSON(allocations))
		}
		storedAllocation, exists := allocations.Allocations[allocationID]
		if !exists {
			t.Fatalf("stored allocations = %s, want allocation %q", utils.ToJSON(allocations), allocationID)
		}
		if storedAllocation.PoolID != "POOL1" || storedAllocation.Address.String() != "172.16.1.1" {
			t.Errorf("unexpected stored allocation: %s", utils.ToJSON(storedAllocation))
		}
	})

	t.Run("release", func(t *testing.T) {
		var reply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ipsRelease",
				APIOpts: map[string]any{
					utils.MetaIPsReleaseCfg:   true,
					utils.MetaOriginID:        allocationID,
					utils.OptsSesBlockerError: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &reply)
		if err != nil {
			t.Fatalf("ProcessEvent release failed: %v", err)
		}

		if allocations := getAllocations(t); len(allocations.Allocations) != 0 {
			t.Errorf("stored allocations after release = %s, want none", utils.ToJSON(allocations))
		}
	})

	errorTests := []struct {
		name    string
		option  string
		blocker bool
	}{
		{name: "allocateBlocker", option: utils.MetaIPsAllocateCfg, blocker: true},
		{name: "allocateNonBlocker", option: utils.MetaIPsAllocateCfg},
		{name: "releaseBlocker", option: utils.MetaIPsReleaseCfg, blocker: true},
		{name: "releaseNonBlocker", option: utils.MetaIPsReleaseCfg},
	}
	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			var reply V1ProcessEventReply
			err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
				&utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     test.name,
					APIOpts: map[string]any{
						test.option:               true,
						utils.MetaCGRid:           utils.Sha1(test.name, ""),
						utils.MetaOriginID:        "",
						utils.OptsSesBlockerError: test.blocker,
					},
					Event: map[string]any{
						utils.AccountField: "1001",
						utils.Destination:  "2001",
						utils.SetupTime:    "2018-01-07T17:00:00Z",
					},
				}, &reply)

			var expectedErr error = utils.ErrPartiallyExecuted
			if test.blocker {
				expectedErr = utils.NewErrMandatoryIeMissing(utils.MetaOriginID)
			}
			if err == nil || err.Error() != expectedErr.Error() {
				t.Errorf("ProcessEvent error = %v, want %v", err, expectedErr)
			}
		})
	}
}
