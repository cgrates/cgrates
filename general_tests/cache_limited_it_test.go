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
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCacheLimitedProfiles(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal, utils.MetaPostgres, utils.MetaMongo:
		t.SkipNow()
	case utils.MetaMySQL:
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{
		"general": {
			"log_level": 7,
		},
		
		"listen": {
			"rpc_json": ":2012",
			"rpc_gob": ":2013",
			"http": ":2080",
		},
	"data_db": {					
		"db_type": "*redis",			
		"db_host": "127.0.0.1",			
		"db_port": 6379, 			
		"db_name": "10", 			
		"db_user": "cgrates", 			
		},
		"rals": {
			"enabled": true,
		},
		"caches":{
	"partitions": {
		"*destinations": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		
		"*reverse_destinations": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*rating_plans": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		
		"*rating_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	 
		"*actions": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		
		"*action_plans": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		
		"*account_action_plans": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*action_triggers": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*timings": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		
		"*resource_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},
		"*resources": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		
		"*trend_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*trends": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		
		"*ranking_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*rankings": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	 
		"*statqueue_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*statqueues": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		
		"*threshold_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*thresholds": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		
		"*filters": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},		 
		"*route_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*attribute_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*charger_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	 
		"*dispatcher_profiles": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		"*dispatcher_hosts": {"limit": 10, "ttl": "3s", "static_ttl": false, "precache": false, "remote":false, "replicate": false},	
		},
		},
		"schedulers": {
			"enabled": true,
			"cdrs_conns":["*internal"],
		},
		"cdrs": {
			"enabled": true,
			"rals_conns": ["*internal"],
		},
		
		"chargers": {
			"enabled": true,
			"attributes_conns": ["*internal"],
		},
		"attributes": {
			"enabled": true,
		},
		}
		`

	ng := engine.TestEngine{
		ConfigJSON: content,
		LogBuffer:  bytes.NewBuffer(nil),
		TpFiles: map[string]string{
			utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,5555,PACKAGE_5555,,,`,
			utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
PACKAGE_5555,ACT_TOPUP_MON,*asap,`,
			utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_MON,*topup_reset,,,money1,*monetary,,*any,,,*unlimited,,250,10,false,false,10`,
			utils.DestinationsCsv: `#Id,Prefix
DST_1001,1001
DST_1002,1002
DST_1002,1002
DST_1003,1003
DST_1004,1004
DST_1005,1005
DST_1006,1006
DST_1007,1007
DST_1008,1008
DST_1009,1009
DST_10010,10010
DST_10011,10011
DST_10012,10012
DST_10013,10013
DST_10014,10014
DST_10015,10015
`,
			utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_1001,*any,RT_1,*up,0,0,
DR_1001,*any,RT_1,*up,0,0,
DR_1002,*any,RT_1,*up,0,0,
DR_1003,*any,RT_1,*up,0,0,
DR_1004,*any,RT_1,*up,0,0,
DR_1005,*any,RT_1,*up,0,0,
DR_1006,*any,RT_1,*up,0,0,
DR_1007,*any,RT_1,*up,0,0,
DR_1008,*any,RT_1,*up,0,0,
DR_1009,*any,RT_1,*up,0,0,
DR_10010,*any,RT_1,*up,0,0,
DR_10011,*any,RT_1,*up,0,0,
DR_10012,*any,RT_1,*up,0,0,
DR_10013,*any,RT_1,*up,0,0,
DR_10014,*any,RT_1,*up,0,0,
DR_10015,*any,RT_1,*up,0,0,
`,
			utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_1,0,2,1s,1s,0s`,
			utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_1001,*any,10
RP_ANY,DR_1002,*any,10
RP_ANY,DR_1003,*any,10
RP_ANY,DR_1004,*any,10
RP_ANY,DR_1005,*any,10
RP_ANY,DR_1006,*any,10
RP_ANY,DR_1007,*any,10
RP_ANY,DR_1008,*any,10
RP_ANY,DR_1009,*any,10
RP_ANY,DR_10010,*any,10
RP_ANY,DR_10011,*any,10
RP_ANY,DR_10012,*any,10
RP_ANY,DR_10013,*any,10
RP_ANY,DR_10014,*any,10
RP_ANY,DR_10015,*any,10`,
			utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,5555,,RP_ANY,`,
		},
	}
	t.SkipNow()
	client, _ := ng.Run(t)
	t.Run("ProcessCDR", func(t *testing.T) {
		for i := 1; i <= 15; i++ {
			var reply string
			uuidStr := utils.GenUUID()
			err := client.Call(context.Background(), utils.CDRsV1ProcessEvent,
				&engine.ArgV1ProcessEvent{
					Flags: []string{utils.MetaRALs},
					CGREvent: utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "event1",
						Event: map[string]any{
							utils.Tenant:       "cgrates.org",
							utils.Category:     "call",
							utils.ToR:          "*voice",
							utils.OriginID:     uuidStr,
							utils.RequestType:  "*rated",
							utils.AccountField: "5555",
							utils.Destination:  fmt.Sprintf("100%d", i),
							utils.SetupTime:    time.Date(2024, time.February, 2, 16, 14, 50, 0, time.UTC),
							utils.AnswerTime:   time.Date(2024, time.February, 2, 16, 15, 0, 0, time.UTC),
							utils.Usage:        10,
						},
					},
				}, &reply)
			if err != nil {
				t.Fatal(err)
			}

			var cdrs []*engine.CDR
			if err := client.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilter{
				OriginIDs: []string{uuidStr},
			}, &cdrs); err != nil {
				t.Errorf("CDRsV1GetCDRs failed unexpectedly: %v", err)
			}
		}
	})
	t.Run("CacheGetElements", func(t *testing.T) {
		args := &utils.ArgsGetCacheItemIDs{
			CacheID: utils.MetaDestinations,
		}
		var reply *[]string
		if err := client.Call(context.Background(), utils.CacheSv1GetItemIDs, args, &reply); err != nil {
			t.Error(err)
		} else if len(*reply) != 15 {
			t.Errorf("expected 15 items, got %d", len(*reply))
		}
	})

}
