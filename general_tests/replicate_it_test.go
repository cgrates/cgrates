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

package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestReplicateMultipleDB(t *testing.T) {
	cfg1 := `
{
"general": {
	"nodeID": "InternalEngine"
},

"logger": {
    "level": 7
},

"listen": {
	"rpcJSON": ":2012",
	"rpcGOB": ":2013",
	"http": ":2080"
},

"rpcConns": {
	"conn2": {
		"strategy": "*broadcast_sync",
		"conns": [
			{"id": "engine1", "address": "127.0.0.1:2023", "transport":"*gob"},
		]
	}
},

"db": {
	"dbConns": {
		"*redis": {
			"dbType": "redis",						
			"dbPort": 6379, 						
			"dbName": "10", 
			"replicationInterval": "-1",
			"replicationConns": ["conn2"],
		},
		 "*default": {
		 	"replicationConns": ["conn2"],
			"replicationInterval": "-1",
			"opts":{
				"internalDBRewriteInterval": "0s",
				"internalDBDumpInterval": "0s"
			}
		 }
		
	},
	"items":{
		"*thresholdProfiles": {"remote":false,"replicate":true},
		"*attributeProfiles":{"remote":false,"replicate":true,"dbConn": "*redis"},
	},
},


"thresholds": {
	"enabled": true,
	"store_interval": "-1"
},

"admins": {
	"enabled": true
},
}
`

	cfg2 := `
{
"general": {
	"nodeID": "InternalEngine2"
},

"logger": {
    "level": 7
},

"listen": {
	"rpcJSON": ":2022",
	"rpcGOB": ":2023",
	"http": ":2280"
},

"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
			"opts":{
				"internalDBRewriteInterval": "0s",
				"internalDBDumpInterval": "0s"
			}
		},
		"*redis": {
			"dbType": "redis",					
			"dbPort": 6379, 						
			"dbName": "13", 
		},
	},
	"items":{
		"*attributeProfiles":{"dbConn": "*redis"},
	},
},

"thresholds": {
	"enabled": true,
	"store_interval": "-1"
},

"admins": {
	"enabled": true
},

}
`
	tpFiles := map[string]string{
		utils.AttributesCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,AttributeFilterIDs,AttributeBlockers,Path,Type,Value
cgrates.org,ATTR_ACNT_1001,*string:~*opts.*context:*sessions,;10,;false,,,*req.OfficeGroup,*constant,Marketing
`,
		utils.ThresholdsCsv: `#Tenant[0],Id[1],FilterIDs[2],Weight[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],AttributeIDs[8],ActionProfileIDs[9],Async[10],EeIDs[11]
cgrates.org,THD_ACNT_1001,*string:~*req.Account:1001,;10,-1,0,0,false,,TOPUP_MONETARY_10,false,`,
	}
	ng := engine.TestEngine{
		ConfigJSON: cfg1,
		Encoding:   *utils.Encoding,
		TpFiles:    tpFiles,
	}
	ng2 := engine.TestEngine{
		ConfigJSON: cfg2,
		Encoding:   *utils.Encoding,
	}
	client2, _ := ng2.Run(t)
	client1, _ := ng.Run(t)
	time.Sleep(500 * time.Millisecond)
	t.Run("GetReplicatedProfiles", func(t *testing.T) {
		var replyAttributeProfile utils.APIAttributeProfile
		if err := client1.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "ATTR_ACNT_1001",
				}}, &replyAttributeProfile); err != nil {
			t.Error(err)
		}

		//replicated profile in second engine
		var replyAttributeProfile2 utils.APIAttributeProfile
		if err := client2.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "ATTR_ACNT_1001",
				}}, &replyAttributeProfile2); err != nil {
			t.Error(err)
		} else if diff := cmp.Diff(replyAttributeProfile, replyAttributeProfile2); diff != "" {
			t.Error(diff)
		}

		var rcvTHP *engine.ThresholdProfile
		if err := client1.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "THD_ACNT_1001",
				},
			}, &rcvTHP); err != nil {
			t.Error(err)
		}

		var rcvTHP2 *engine.ThresholdProfile
		if err := client2.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "THD_ACNT_1001",
				},
			}, &rcvTHP2); err != nil {
			t.Error(err)
		} else if diff := cmp.Diff(rcvTHP, rcvTHP2, cmpopts.IgnoreUnexported(engine.ThresholdProfile{})); diff != "" {
			t.Error(diff)
		}

	})

}
