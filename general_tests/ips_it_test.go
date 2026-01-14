//go:build integration

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

package general_tests

import (
	"cmp"
	"net/netip"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// Expected IP profiles for testing
var (
	// Profile created via SetIPProfile API
	expectedIPsAPIProfile = &engine.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IPsAPI",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    15,
		TTL:       -1,
		Stored:    false,
		Pools: []*engine.IPPool{
			{
				ID:        "API_POOL",
				FilterIDs: []string{},
				Type:      "*ipv4",
				Range:     "10.100.0.1/32",
				Strategy:  "*ascending",
				Message:   "API created pool",
				Weight:    15,
				Blocker:   false,
			},
		},
	}

	// Profile from CSV test data (IPs1)
	expectedIPs1Profile = &engine.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IPs1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    10,
		TTL:       time.Second,
		Stored:    true,
		Pools: []*engine.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{"*string:~*req.Destination:2001"},
				Type:      "*ipv4",
				Range:     "172.16.1.1/32",
				Strategy:  "*ascending",
				Message:   "alloc_success",
				Weight:    15,
				Blocker:   false,
			},
			{
				ID:        "POOL2",
				FilterIDs: []string{"*string:~*req.Destination:2002"},
				Type:      "*ipv4",
				Range:     "192.168.122.1/32",
				Strategy:  "*random",
				Message:   "alloc_new",
				Weight:    25,
				Blocker:   true,
			},
		},
	}

	// Profile from CSV test data (IPs2)
	expectedIPs2Profile = &engine.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IPs2",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Weight:    20,
		TTL:       2 * time.Second,
		Stored:    false,
		Pools: []*engine.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{"*string:~*req.Destination:3001"},
				Type:      "*ipv4",
				Range:     "127.0.0.1/32",
				Strategy:  "*descending",
				Message:   "alloc_msg",
				Weight:    35,
				Blocker:   true,
			},
		},
	}
)

// TODO: move anything sessions related to sessions once ips implementation
// is complete.
func TestIPsIT(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbCfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbCfg = engine.PostgresDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
"sessions": {
	"enabled": true,
	"ips_conns": ["*localhost"]
},
"ips": {
    "enabled": true,
    "store_interval": "-1",
    "indexed_selects": true,
    "string_indexed_fields": ["*req.Account"],
    "prefix_indexed_fields": [],
    "suffix_indexed_fields": [],
    "exists_indexed_fields": []
},
"apiers": {
	"enabled": true
}
}`,
		TpFiles: map[string]string{
			utils.IPsCsv: `
#Tenant[0],ID[1],FilterIDs[2],ActivationInterval[3],TTL[4],Stored[5],Weight[6],PoolID[7],PoolFilterIDs[8],PoolType[9],PoolRange[10],PoolStrategy[11],PoolMessage[12],PoolWeight[13],PoolBlocker[14]
cgrates.org,IPs1,*string:~*req.Account:1001,,1s,true,10,,,,,,,,
cgrates.org,IPs1,,,,,,POOL1,*string:~*req.Destination:2001,*ipv4,172.16.1.1/32,*ascending,alloc_success,15,
cgrates.org,IPs1,,,,,,POOL2,*string:~*req.Destination:2002,*ipv4,192.168.122.1/32,*random,alloc_new,25,true
cgrates.org,IPs2,*string:~*req.Account:1002,,2s,false,20,POOL1,*string:~*req.Destination:3001,*ipv4,127.0.0.1/32,*descending,alloc_msg,35,true`,
		},
		DBCfg: dbCfg,
		// LogBuffer: new(bytes.Buffer),
		// GracefulShutdown: true,
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond) // wait for all services

	t.Run("admins apis", func(t *testing.T) {
		var reply string
		err := client.Call(context.Background(), utils.APIerSv1SetIPProfile,
			&engine.IPProfileWithAPIOpts{
				IPProfile: &engine.IPProfile{
					Tenant:    "cgrates.org",
					ID:        "IPsAPI",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weight:    15,
					TTL:       -1,
					Stored:    false,
					Pools: []*engine.IPPool{
						{
							ID:        "API_POOL",
							FilterIDs: []string{},
							Type:      "*ipv4",
							Range:     "10.100.0.1/32",
							Strategy:  "*ascending",
							Message:   "API created pool",
							Weight:    15,
							Blocker:   false,
						},
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}

		var profile engine.IPProfile
		err = client.Call(context.Background(), utils.APIerSv1GetIPProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "IPsAPI",
				},
			}, &profile)
		if err != nil {
			t.Fatal(err)
		}
		verifyIPProfile(t, &profile, expectedIPsAPIProfile)

		err = client.Call(context.Background(), utils.APIerSv1GetIPProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "IPs1",
				},
			}, &profile)
		if err != nil {
			t.Fatal(err)
		}
		slices.SortFunc(profile.Pools, func(a, b *engine.IPPool) int {
			return cmp.Compare(a.ID, b.ID)
		})
		verifyIPProfile(t, &profile, expectedIPs1Profile)

		err = client.Call(context.Background(), utils.APIerSv1GetIPProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "IPs2",
				},
			}, &profile)
		if err != nil {
			t.Fatal(err)
		}
		verifyIPProfile(t, &profile, expectedIPs2Profile)

		err = client.Call(context.Background(), utils.APIerSv1RemoveIPProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "IPsAPI",
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}

		// Verify profile is gone
		err = client.Call(context.Background(), utils.APIerSv1GetIPProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "IPsAPI",
				},
			}, &engine.IPProfile{})
		verifySpecificError(t, err, utils.ErrNotFound.Error())
	})

	t.Run("ips core functionality", func(t *testing.T) {
		// Account "1001" + Destination "2001" should match IPs1/POOL1 -> 172.16.1.1
		allocID := "test_allocation"

		// no allocs yet
		verifyAllocations(t, client, "IPs1")

		var eventAllocs engine.IPAllocations
		err := client.Call(context.Background(), utils.IPsV1GetIPAllocationForEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "GetIPsForEvent1",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
				},
				APIOpts: map[string]any{
					utils.OptsIPsAllocationID: allocID,
				},
			}, &eventAllocs)
		if err != nil {
			t.Fatal(err)
		}
		if eventAllocs.ID != "IPs1" {
			t.Fatalf("want profile IPs1, got %s", eventAllocs.ID)
		}

		expectedAuthorizedIP := &engine.AllocatedIP{
			ProfileID: "IPs1",
			PoolID:    "POOL1",
			Message:   "alloc_success",
			Address:   netip.MustParseAddr("172.16.1.1"),
		}
		var authorizedIP engine.AllocatedIP
		err = client.Call(context.Background(), utils.IPsV1AuthorizeIP,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "AuthorizeIP1",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
				},
				APIOpts: map[string]any{
					utils.OptsIPsAllocationID: allocID,
				},
			}, &authorizedIP)
		if err != nil {
			t.Fatal(err)
		}
		verifyAllocatedIP(t, &authorizedIP, expectedAuthorizedIP)

		// verify allocation doesn't exist yet (authorize is dry run).
		verifyAllocations(t, client, "IPs1")

		expectedAllocatedIP := &engine.AllocatedIP{
			ProfileID: "IPs1",
			PoolID:    "POOL1",
			Message:   "alloc_success",
			Address:   netip.MustParseAddr("172.16.1.1"),
		}
		var allocatedIP engine.AllocatedIP
		err = client.Call(context.Background(), utils.IPsV1AllocateIP,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "AllocateIP1",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
				},
				APIOpts: map[string]any{
					utils.OptsIPsAllocationID: allocID,
				},
			}, &allocatedIP)
		if err != nil {
			t.Fatal(err)
		}
		verifyAllocatedIP(t, &allocatedIP, expectedAllocatedIP)

		// verify allocation exists now
		verifyAllocations(t, client, "IPs1", allocID)
		verifyIPAllocation(t, client, "IPs1", allocID, expectedAuthorizedIP.Address.String())

		// double allocation should fail
		err = client.Call(context.Background(), utils.IPsV1AllocateIP,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "AllocateIP2",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
				},
				APIOpts: map[string]any{
					utils.OptsIPsAllocationID: "different_alloc",
				},
			}, &engine.AllocatedIP{})
		verifySpecificError(t, err,
			`allocation failed for pool "POOL1", IP "172.16.1.1": IP_ALREADY_ALLOCATED (allocated to "test_allocation")`)

		var reply string
		err = client.Call(context.Background(), utils.IPsV1ReleaseIP,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ReleaseIP1",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
				},
				APIOpts: map[string]any{
					utils.OptsIPsAllocationID: allocID,
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}

		// allocation should be gone.
		verifyAllocations(t, client, "IPs1")

		// test ClearAllocations API
		alloc1, alloc2 := "test_alloc1", "test_alloc2"

		// allocate first IP (should work after release)
		err = client.Call(context.Background(), utils.IPsV1AllocateIP,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "AllocateForClear1",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
				},
				APIOpts: map[string]any{
					utils.OptsIPsAllocationID: alloc1,
				},
			}, &engine.AllocatedIP{})
		if err != nil {
			t.Fatal(err)
		}
		verifyAllocations(t, client, "IPs1", alloc1)

		// try second allocation on different pool (should get POOL2)
		err = client.Call(context.Background(), utils.IPsV1AllocateIP,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "AllocateForClear2",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2002", // different destination for POOL2
				},
				APIOpts: map[string]any{
					utils.OptsIPsAllocationID: alloc2,
				},
			}, &engine.AllocatedIP{})
		if err != nil {
			t.Fatal(err)
		}
		verifyAllocations(t, client, "IPs1", alloc1, alloc2)

		// clear only second allocation
		err = client.Call(context.Background(), utils.IPsV1ClearIPAllocations,
			&engine.ClearIPAllocationsArgs{
				Tenant:        "cgrates.org",
				ID:            "IPs1",
				AllocationIDs: []string{alloc2},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		verifyAllocations(t, client, "IPs1", alloc1)

		// clear the rest (by not specifying AllocationIDs)
		err = client.Call(context.Background(), utils.IPsV1ClearIPAllocations,
			&engine.ClearIPAllocationsArgs{
				Tenant: "cgrates.org",
				ID:     "IPs1",
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		verifyAllocations(t, client, "IPs1")
	})

	t.Run("sessions integration", func(t *testing.T) {
		var authReply sessions.V1AuthorizeReply
		err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
			&sessions.V1AuthorizeArgs{
				AuthorizeIP: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.OriginID:     "session_auth_test",
						utils.AccountField: "1001",
						utils.Destination:  "2001",
						utils.SetupTime:    "2018-01-07T17:00:00Z",
					},
				},
			}, &authReply)
		if err != nil {
			t.Fatal(err)
		}
		if authReply.AllocatedIP == nil {
			t.Fatal("SessionSv1AuthorizeEvent reply does not contain AllocatedIP")
		}

		var initReply sessions.V1InitSessionReply
		err = client.Call(context.Background(), utils.SessionSv1InitiateSession,
			&sessions.V1InitSessionArgs{
				AllocateIP: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					Event: map[string]any{
						utils.OriginID:     "session_init_test",
						utils.AccountField: "1001",
						utils.Destination:  "2001",
						utils.SetupTime:    "2018-01-07T17:00:00Z",
					},
				},
			}, &initReply)
		if err != nil {
			t.Fatal(err)
		}
		if initReply.AllocatedIP == nil {
			t.Fatal("SessionSv1InitiateSession reply does not contain AllocatedIP")
		}
	})
}

// Helper functions for testing

func getIPAllocations(t *testing.T, client *birpc.Client, profileID string) *engine.IPAllocations {
	t.Helper()
	var allocs engine.IPAllocations
	if err := client.Call(context.Background(), utils.IPsV1GetIPAllocations,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     profileID,
			},
		}, &allocs); err != nil {
		t.Fatal(err)
	}
	return &allocs
}

func verifyAllocatedIP(t *testing.T, got, want *engine.AllocatedIP) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AllocatedIP mismatch:\nwant: %s\ngot: %s",
			utils.ToJSON(want), utils.ToJSON(got))
	}
}

func verifySpecificError(t *testing.T, got error, expected string) {
	t.Helper()
	if got == nil {
		t.Fatalf("want error %v, got nil", expected)
	}
	if got.Error() != expected {
		t.Fatalf("want error %v, got %v", expected, got)
	}
}

func verifyAllocations(t *testing.T, client *birpc.Client, profileID string, wantAllocs ...string) {
	t.Helper()
	allocs := getIPAllocations(t, client, profileID)

	if len(allocs.Allocations) != len(wantAllocs) {
		t.Fatalf("Expected %d allocations, got %d: %s",
			len(wantAllocs), len(allocs.Allocations), utils.ToJSON(allocs))
	}

	for _, allocID := range wantAllocs {
		if _, exists := allocs.Allocations[allocID]; !exists {
			t.Fatalf("Allocation %s not found in profile %s: %s",
				allocID, profileID, utils.ToJSON(allocs))
		}
	}
}

func verifyIPAllocation(t *testing.T, client *birpc.Client, profileID, allocID, expectedIP string) {
	t.Helper()
	allocs := getIPAllocations(t, client, profileID)

	alloc, exists := allocs.Allocations[allocID]
	if !exists {
		t.Fatalf("Allocation %s not found in profile %s", allocID, profileID)
	}
	if alloc.Address.String() != expectedIP {
		t.Fatalf("Expected IP %s, got %s for allocation %s",
			expectedIP, alloc.Address.String(), allocID)
	}
}

func verifyIPProfile(t *testing.T, got, want *engine.IPProfile) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Profile mismatch:\nwant: %s\ngot: %s",
			utils.ToJSON(want), utils.ToJSON(got))
	}
}
