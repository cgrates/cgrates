//go:build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package ips

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NOTE: this test is incomplete. Currently used only for the API samples.
// TODO: move anything sessions related to sessions once ips implementation
// is complete.
func TestIPsIT(t *testing.T) {
	// t.Skip("ips test currently incomplete, skipping...")
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL: // redis is already the default
	case utils.MetaMongo:
		dbCfg = engine.MongoDBCfg
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
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
		"*usageID": [
			{
				"Tenant": "cgrates.org",
				"FilterIDs": ["*string:~*req.Account:1001"],
				"Value": "cfg_usage"
			}
		],
		// "*ttl": [
		//     {
		// 		"Tenant": "*any",
		// 		"FilterIDs": [],
		// 		"Value": "72h"
		//     }
		// ],
		// "*units": [
		//     {
		// 		"Tenant": "*any",
		// 		"FilterIDs": [],
		// 		"Value": 1
		//     }
		// ]
    }
},
"admins": {
	"enabled": true
}
}`,
		TpFiles: map[string]string{
			utils.IPsCsv: `
#Tenant[0],Id[1],FilterIDs[2],Weights[3],TTL[4],Type[5],AddressPool[6],Allocation[7],Stored[8]
cgrates.org,IPs1,*string:~*req.Account:1001,;10,1s,true,,,,,,,,
cgrates.org,IPs1,,,,,POOL1,*string:~*req.Destination:2001,*ipv4,172.16.1.1/24,*ascending,alloc_success,;15,
cgrates.org,IPs1,,,,,POOL1,,,,,,*exists:~*req.GimmeMoreWeight:;50,*exists:~*req.ShouldBlock:;true
cgrates.org,IPs1,,,,,POOL2,*string:~*req.Destination:2002,*ipv4,192.168.122.1/24,*random,alloc_new,;25,;true
cgrates.org,IPs2,*string:~*req.Account:1002,;20,2s,false,POOL1,*string:~*req.Destination:3001,*ipv4,127.0.0.1/24,*descending,alloc_msg,;35,;true`,
		},
		DBCfg:            dbCfg,
		Encoding:         *utils.Encoding,
		LogBuffer:        new(bytes.Buffer),
		GracefulShutdown: true,
	}
	t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
	client, _ := ng.Run(t)

	t.Run("admins apis", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.AdminSv1SetIPProfile,
			&utils.IPProfileWithAPIOpts{
				IPProfile: &utils.IPProfile{
					Tenant:    "cgrates.org",
					ID:        "IPsAPI",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							Weight: 15,
						},
					},
					TTL:    -1,
					Stored: false,
					Pools: []*utils.IPPool{
						{
							ID:        "FIRST_POOL",
							FilterIDs: []string{},
							Type:      "*ipv4",
							Range:     "192.168.122.1/24",
							Strategy:  "*ascending",
							Message:   "Some message",
							Weights: utils.DynamicWeights{
								{
									Weight: 15,
								},
							},
							Blockers: utils.DynamicBlockers{
								{
									Blocker: false,
								},
							},
						},
					},
				},
			}, &reply); err != nil {
			t.Error(err)
		}

		var ipp utils.IPProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetIPProfile,
			utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "IPsAPI",
			}, &ipp); err != nil {
			t.Error(err)
		}

		var ipps []*utils.IPProfile
		if err := client.Call(context.Background(), utils.AdminSv1GetIPProfiles,
			&utils.ArgsItemIDs{
				Tenant:      "cgrates.org",
				ItemsPrefix: "IPs",
			}, &ipps); err != nil {
			t.Error(err)
		}

		if err := client.Call(context.Background(), utils.AdminSv1RemoveIPProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "IPsAPI",
				},
			}, &reply); err != nil {
			t.Error(err)
		}

		var no int
		if err := client.Call(context.Background(), utils.AdminSv1GetIPProfilesCount,
			&utils.TenantWithAPIOpts{
				Tenant: "cgrates.org",
			}, &no); err != nil {
			t.Error(err)
		}
	})

	t.Run("ips apis", func(t *testing.T) {
		var ip utils.IPAllocations
		if err := client.Call(context.Background(), utils.IPsV1GetIPAllocations,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "IPs1",
				},
			}, &ip); err != nil {
			t.Error(err)
		}

		usageID := "api_usage"
		var evIPs IPAllocationsList
		if err := client.Call(context.Background(), utils.IPsV1GetIPAllocationsForEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "GetIPsForEvent1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.OptsIPsUsageID: usageID,
				},
			}, &evIPs); err != nil {
			t.Error(err)
		}

		var reply string
		if err := client.Call(context.Background(), utils.IPsV1AuthorizeIP,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "AuthorizeIP1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					// utils.OptsIPsUsageID: usageID,
					// utils.OptsIPsTTL:   time.Second,
					// utils.OptsIPsUnits: 2,
				},
			}, &reply); err != nil {
			t.Error(err)
		}

		if err := client.Call(context.Background(), utils.IPsV1AllocateIP,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "AllocateIP1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					// utils.OptsIPsUsageID: usageID,
					// utils.OptsIPsTTL:   time.Second,
					// utils.OptsIPsUnits: 2,
				},
			}, &reply); err != nil {
			t.Error(err)
		}

		if err := client.Call(context.Background(), utils.IPsV1ReleaseIP,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ReleaseIP1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					// utils.OptsIPsUsageID: usageID,
					// utils.OptsIPsTTL:   time.Second,
					// utils.OptsIPsUnits: 2,
				},
			}, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("sessions ips apis", func(t *testing.T) {
		// NOTE: reply is of type any to avoid having to import sessions just for
		// this test in order to prevent future cyclic imports. Any sessions
		// related test should be moved to sessions when ips implementation is
		// complete.
		var reply any
		if err := client.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				APIOpts: map[string]any{
					utils.MetaIPs:      true,
					utils.MetaOriginID: "session1",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &reply); err != nil {
			t.Error(err)
		}
		if err := client.Call(context.Background(), utils.SessionSv1InitiateSession,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				APIOpts: map[string]any{
					utils.MetaIPs:      true,
					utils.MetaOriginID: "session1",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    "2018-01-07T17:00:00Z",
				},
			}, &reply); err != nil {
			t.Error(err)
		}
	})
}
