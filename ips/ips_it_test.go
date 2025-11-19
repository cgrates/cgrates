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

package ips

import (
	"net/netip"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Expected IP profiles for testing
var (
	// Profile created via SetIPProfile API
	expectedIPsAPIProfile = &utils.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IPsAPI",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights:   utils.DynamicWeights{{Weight: 15}},
		TTL:       -1,
		Stored:    false,
		Pools: []*utils.IPPool{
			{
				ID:        "API_POOL",
				FilterIDs: []string{},
				Type:      "*ipv4",
				Range:     "10.100.0.1/32",
				Strategy:  "*ascending",
				Message:   "API created pool",
				Weights:   utils.DynamicWeights{{Weight: 15}},
				Blockers:  utils.DynamicBlockers{{Blocker: false}},
			},
		},
	}

	// Profile from CSV test data (IPs1)
	expectedIPs1Profile = &utils.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IPs1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights:   utils.DynamicWeights{{Weight: 10}},
		TTL:       time.Second,
		Stored:    true,
		Pools: []*utils.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{"*string:~*req.Destination:2001"},
				Type:      "*ipv4",
				Range:     "172.16.1.1/32",
				Strategy:  "*ascending",
				Message:   "alloc_success",
				Weights:   utils.DynamicWeights{{Weight: 15}, {FilterIDs: []string{"*exists:~*req.NeedMoreWeight:"}, Weight: 50}},
				Blockers:  utils.DynamicBlockers{{FilterIDs: []string{"*exists:~*req.ShouldBlock:"}, Blocker: true}},
			},
			{
				ID:        "POOL2",
				FilterIDs: []string{"*string:~*req.Destination:2002"},
				Type:      "*ipv4",
				Range:     "192.168.122.1/32",
				Strategy:  "*random",
				Message:   "alloc_new",
				Weights:   utils.DynamicWeights{{Weight: 25}},
				Blockers:  utils.DynamicBlockers{{Blocker: true}},
			},
		},
	}

	// Profile from CSV test data (IPs2)
	expectedIPs2Profile = &utils.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IPs2",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Weights:   utils.DynamicWeights{{Weight: 20}},
		TTL:       2 * time.Second,
		Stored:    false,
		Pools: []*utils.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{"*string:~*req.Destination:3001"},
				Type:      "*ipv4",
				Range:     "127.0.0.1/32",
				Strategy:  "*descending",
				Message:   "alloc_msg",
				Weights:   utils.DynamicWeights{{Weight: 35}},
				Blockers:  utils.DynamicBlockers{{Blocker: true}},
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
	case utils.MetaRedis:
		dbCfg = engine.RedisDBCfg
	case utils.MetaMySQL:
		dbCfg = engine.MySQLDBCfg
	case utils.MetaMongo:
		dbCfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		dbCfg = engine.PostgresDBCfg
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
	"ips_conns": ["*localhost"],
	"opts": {
		// "*ips": [
		// 	{
		// 		"Tenant": "*any",
		// 		"FilterIDs": [],
		// 		"Value": false
		// 	}
		// "*ipsAuthorize": [
		// 	{
		// 		"Tenant": "*any",
		// 		"FilterIDs": [],
		// 		"Value": false
		// 	}
		// ],
		// "*ipsAllocate": [
		// 	{
		// 		"Tenant": "*any",
		// 		"FilterIDs": [],
		// 		"Value": false
		// 	}
		// ],
		// "*ipsRelease": [
		// 	{
		// 		"Tenant": "*any",
		// 		"FilterIDs": [],
		// 		"Value": false
		// 	}
		// ]
	}
},
"ips": {
    "enabled": true,
    "store_interval": "-1",
    "indexed_selects": true,
    "string_indexed_fields": ["*req.Account"],
    "prefix_indexed_fields": [],
    "suffix_indexed_fields": [],
    "exists_indexed_fields": [],
    "notexists_indexed_fields": [],
    "opts":{
		// "*allocationID": [
		// 	{
		// 		"Tenant": "cgrates.org",
		// 		"FilterIDs": ["*string:~*req.Account:1001"],
		// 		"Value": "cfg_allocation"
		// 	}
		// ]
    }
},
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal",
				"opts":{
			"internalDBRewriteInterval": "0s",
			"internalDBDumpInterval": "0s"
		}
    	},
	},
},
"admins": {
	"enabled": true
}
}`,
		TpFiles: map[string]string{
			utils.IPsCsv: `
#Tenant[0],ID[1],FilterIDs[2],Weights[3],TTL[4],Stored[5],PoolID[6],PoolFilterIDs[7],PoolType[8],PoolRange[9],PoolStrategy[10],PoolMessage[11],PoolWeights[12],PoolBlockers[13]
cgrates.org,IPs1,*string:~*req.Account:1001,;10,1s,true,,,,,,,,
cgrates.org,IPs1,,,,,POOL1,*string:~*req.Destination:2001,*ipv4,172.16.1.1/32,*ascending,alloc_success,;15,
cgrates.org,IPs1,,,,,POOL1,,,,,,*exists:~*req.NeedMoreWeight:;50,*exists:~*req.ShouldBlock:;true
cgrates.org,IPs1,,,,,POOL2,*string:~*req.Destination:2002,*ipv4,192.168.122.1/32,*random,alloc_new,;25,;true
cgrates.org,IPs2,*string:~*req.Account:1002,;20,2s,false,POOL1,*string:~*req.Destination:3001,*ipv4,127.0.0.1/32,*descending,alloc_msg,;35,;true`,
		},
		DBCfg:    dbCfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),
		// GracefulShutdown: true,
	}
	// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)
	time.Sleep(50 * time.Millisecond) // wait for all services

	t.Run("admins apis", func(t *testing.T) {
		var reply string
		err := client.Call(context.Background(), utils.AdminSv1SetIPProfile,
			&utils.IPProfileWithAPIOpts{
				IPProfile: &utils.IPProfile{
					Tenant:    "cgrates.org",
					ID:        "IPsAPI",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   utils.DynamicWeights{{Weight: 15}},
					TTL:       -1,
					Stored:    false,
					Pools: []*utils.IPPool{
						{
							ID:        "API_POOL",
							FilterIDs: []string{},
							Type:      "*ipv4",
							Range:     "10.100.0.1/32",
							Strategy:  "*ascending",
							Message:   "API created pool",
							Weights:   utils.DynamicWeights{{Weight: 15}},
							Blockers:  utils.DynamicBlockers{{Blocker: false}},
						},
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}

		var profile utils.IPProfile
		err = client.Call(context.Background(), utils.AdminSv1GetIPProfile,
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

		var ipps []*utils.IPProfile
		err = client.Call(context.Background(), utils.AdminSv1GetIPProfiles,
			&utils.ArgsItemIDs{
				Tenant:      "cgrates.org",
				ItemsPrefix: "IPs",
			}, &ipps)
		if err != nil {
			t.Fatal(err)
		}
		if len(ipps) != 3 { // IPs1, IPs2 from CSV + IPsAPI from SetIPProfile
			t.Fatalf("want exactly 3 profiles, got %d", len(ipps))
		}

		// Verify each expected profile exists and matches exactly
		expectedProfiles := []*utils.IPProfile{
			expectedIPs1Profile,
			expectedIPs2Profile,
			expectedIPsAPIProfile,
		}
		for _, expected := range expectedProfiles {
			found := findProfileByID(ipps, expected.ID)
			if found == nil {
				t.Fatalf("profile %q not found in GetIPProfiles results", expected.ID)
			}
			verifyIPProfile(t, found, expected)
		}

		var count int
		err = client.Call(context.Background(), utils.AdminSv1GetIPProfilesCount,
			&utils.TenantWithAPIOpts{
				Tenant: "cgrates.org",
			}, &count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 3 {
			t.Fatalf("want exactly 3 profiles in count, got %d", count)
		}

		err = client.Call(context.Background(), utils.AdminSv1RemoveIPProfile,
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
		err = client.Call(context.Background(), utils.AdminSv1GetIPProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "IPsAPI",
				},
			}, &utils.IPProfile{})
		verifySpecificError(t, err, utils.ErrNotFound.Error())
	})

	t.Run("ips core functionality", func(t *testing.T) {
		// Account "1001" + Destination "2001" should match IPs1/POOL1 -> 172.16.1.1
		allocID := "test_allocation"

		// no allocs yet
		verifyAllocations(t, client, "IPs1")

		var eventAllocs utils.IPAllocations
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

		expectedAuthorizedIP := &utils.AllocatedIP{
			ProfileID: "IPs1",
			PoolID:    "POOL1",
			Message:   "alloc_success",
			Address:   netip.MustParseAddr("172.16.1.1"),
		}
		var authorizedIP utils.AllocatedIP
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

		expectedAllocatedIP := &utils.AllocatedIP{
			ProfileID: "IPs1",
			PoolID:    "POOL1",
			Message:   "alloc_success",
			Address:   netip.MustParseAddr("172.16.1.1"),
		}
		var allocatedIP utils.AllocatedIP
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
			}, &utils.AllocatedIP{})
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
			}, &utils.AllocatedIP{})
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
			}, &utils.AllocatedIP{})
		if err != nil {
			t.Fatal(err)
		}
		verifyAllocations(t, client, "IPs1", alloc1, alloc2)

		// clear only second allocation
		err = client.Call(context.Background(), utils.IPsV1ClearIPAllocations,
			&utils.ClearIPAllocationsArgs{
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
			&utils.ClearIPAllocationsArgs{
				Tenant: "cgrates.org",
				ID:     "IPs1",
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		verifyAllocations(t, client, "IPs1")

	})

	t.Run("sessions integration", func(t *testing.T) {
		// NOTE: reply is of type any to avoid having to import sessions just for
		// this test in order to prevent future cyclic imports. Any sessions
		// related test should be moved to sessions when ips implementation is
		// complete.

		var authReply any
		err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				APIOpts: map[string]any{
					utils.MetaIPs:      true,
					utils.MetaOriginID: "session_auth_test",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &authReply)
		if err != nil {
			t.Fatal(err)
		}
		if authReply == nil {
			t.Fatal("SessionSv1AuthorizeEvent returned nil reply")
		}

		var initReply any
		err = client.Call(context.Background(), utils.SessionSv1InitiateSession,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				APIOpts: map[string]any{
					utils.MetaIPs:      true,
					utils.MetaOriginID: "session_init_test",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "2001",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &initReply)
		if err != nil {
			t.Fatal(err)
		}
		if initReply == nil {
			t.Fatal("SessionSv1InitiateSession returned nil reply")
		}
	})
}

// Helper functions for testing

func getIPAllocations(t *testing.T, client *birpc.Client, profileID string) *utils.IPAllocations {
	t.Helper()
	var allocs utils.IPAllocations
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

func verifyAllocatedIP(t *testing.T, got, expected *utils.AllocatedIP) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("AllocatedIP mismatch:\nExpected: %s\nGot: %s",
			utils.ToIJSON(expected), utils.ToIJSON(got))
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

func verifyIPProfile(t *testing.T, got, expected *utils.IPProfile) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Profile mismatch:\nwant: %s\ngot: %s",
			utils.ToIJSON(expected), utils.ToIJSON(got))
	}
}

func findProfileByID(profiles []*utils.IPProfile, id string) *utils.IPProfile {
	for _, p := range profiles {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func BenchmarkIPsAuthorize(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cm := engine.NewConnManager(cfg)
	ipService := NewIPService(dm, cfg, fltrs, cm)

	ctx := context.Background()
	profile := &utils.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IP1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights:   utils.DynamicWeights{{Weight: 10}},
		TTL:       time.Second,
		Stored:    false,
		Pools: []*utils.IPPool{{
			ID:        "POOL1",
			FilterIDs: []string{},
			Type:      "*ipv4",
			Range:     "192.168.122.1/32",
			Strategy:  "*ascending",
			Message:   "bench pool",
			Weights:   utils.DynamicWeights{{Weight: 10}},
			Blockers:  utils.DynamicBlockers{{Blocker: false}},
		}},
	}
	if err := dm.SetIPProfile(ctx, profile, true); err != nil {
		b.Fatalf("Failed to set IPProfile: %v", err)
	}

	for b.Loop() {
		b.StopTimer()
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsIPsAllocationID: "alloc1",
			},
		}
		var allocIP utils.AllocatedIP
		b.StartTimer()
		if err := ipService.V1AuthorizeIP(ctx, args, &allocIP); err != nil {
			b.Error("AuthorizeIP failed:", err)
		}
	}
}

func BenchmarkIPsAllocate(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cm := engine.NewConnManager(cfg)
	ipService := NewIPService(dm, cfg, fltrs, cm)

	ctx := context.Background()
	profile := &utils.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IP2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights:   utils.DynamicWeights{{Weight: 10}},
		TTL:       time.Second,
		Stored:    false,
		Pools: []*utils.IPPool{{
			ID:        "POOL1",
			FilterIDs: []string{},
			Type:      "*ipv4",
			Range:     "192.168.122.1/32",
			Strategy:  "*ascending",
			Message:   "bench pool",
			Weights:   utils.DynamicWeights{{Weight: 10}},
			Blockers:  utils.DynamicBlockers{{Blocker: false}},
		}},
	}
	if err := dm.SetIPProfile(ctx, profile, true); err != nil {
		b.Fatalf("Failed to set IPProfile: %v", err)
	}
	for b.Loop() {
		b.StopTimer()
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsIPsAllocationID: "alloc1",
			},
		}
		var allocIP utils.AllocatedIP
		b.StartTimer()
		if err := ipService.V1AllocateIP(ctx, args, &allocIP); err != nil {
			b.Error("AuthorizeIP failed:", err)
		}
	}
}
func BenchmarkIPsRelease(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	cm := engine.NewConnManager(cfg)
	ipService := NewIPService(dm, cfg, fltrs, cm)
	ctx := context.Background()
	profile := &utils.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IP3",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights:   utils.DynamicWeights{{Weight: 10}},
		TTL:       time.Second,
		Stored:    false,
		Pools: []*utils.IPPool{{
			ID:        "POOL1",
			FilterIDs: []string{},
			Type:      "*ipv4",
			Range:     "192.168.122.1/32",
			Strategy:  "*ascending",
			Message:   "bench pool",
			Weights:   utils.DynamicWeights{{Weight: 10}},
			Blockers:  utils.DynamicBlockers{{Blocker: false}},
		}},
	}
	if err := dm.SetIPProfile(ctx, profile, true); err != nil {
		b.Fatalf("Failed to set IPProfile: %v", err)
	}

	for b.Loop() {
		b.StopTimer()
		args := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsIPsAllocationID: "alloc1",
			},
		}
		var reply string
		b.StartTimer()
		if err := ipService.V1ReleaseIP(ctx, args, &reply); err != nil {
			b.Error("AuthorizeIP failed:", err)
		}
	}
}
