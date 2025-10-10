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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestReplicationBatch(t *testing.T) {
	content1 := `{
    "general": {
        "log_level": 7,
    },
	"listen": {
		"rpc_json": ":2022",
		"rpc_gob": ":2023",
		"http": ":2280"
			 },
	"rpc_conns": {
		"conn2": {
			"strategy": "*first",
			"conns": [{"address": "127.0.0.1:2033", "transport":"*gob"}]
		},
		"conn3": {
			"strategy": "*first",
			"conns": [{"address": "127.0.0.1:2043", "transport":"*gob"}]
		}
				},
    "data_db": {
        "db_type": "redis",
        "db_port": 6379,
        "db_name": "10",
		"replication_conns":["conn2","conn3"],			
	    "replication_interval": "1s", 		
		"items":{
		"*accounts":{"remote":false,"replicate":true},
		"*resource_profiles":{"remote":false,"replicate":true},
		"*resources":{"remote":false,"replicate":true},
		"*statqueue_profiles": {"remote":false,"replicate":true},
		"*statqueues": {"remote":false,"replicate":true},
		"*threshold_profiles": {"remote":false,"replicate":true},
		"*thresholds": {"remote":false,"replicate":true},
		"*filters": {"remote":false,"replicate":true},
		"*route_profiles":{"remote":false,"replicate":true},
		"*attribute_profiles":{"remote":false,"replicate":true},
		"*charger_profiles": {"remote":false,"replicate":true},
		"*rate_profiles":{"remote":false,"replicate":true},
		"*load_ids":{"remote":false,"replicate":true},
		"*indexes":{"remote":false, "replicate":true}, 
		"*action_profiles":{"remote":false,"replicate":true},
		"*account_profiles":{"remote":false,"replicate":true}
	}
    },
    "stor_db": {
        "db_type": "*internal",
    },
    "stats": {
        "enabled": true,
        "store_interval": "-1",
    },
    "admins": {
        "enabled": true,
    }
}
`

	content2 := `{
 "general": {
        "log_level": 7,
    },
	"listen": {
		"rpc_json": ":2032",
		"rpc_gob": ":2033",
		"http": ":2380"
	},
	"rpc_conns": {
		"conn1": {
			"strategy": "*first",
			"conns": [{"address": "127.0.0.1:2023", "transport":"*gob"}]
			}
		},
    "data_db": {
        "db_type": "redis",
        "db_port": 6379,
        "db_name": "11",
		"replication_conns":["conn1"],			 	
		"items":{
		"*accounts":{"remote":false,"replicate":true},
		"*resource_profiles":{"remote":false,"replicate":true},
		"*resources":{"remote":false,"replicate":true},
		"*statqueue_profiles": {"remote":false,"replicate":true},
		"*statqueues": {"remote":false,"replicate":true},
		"*threshold_profiles": {"remote":false,"replicate":true},
		"*thresholds": {"remote":false,"replicate":true},
		"*filters": {"remote":false,"replicate":true},
		"*route_profiles":{"remote":false,"replicate":true},
		"*attribute_profiles":{"remote":false,"replicate":true},
		"*charger_profiles": {"remote":false,"replicate":true},
		"*rate_profiles":{"remote":false,"replicate":true},
		"*load_ids":{"remote":false,"replicate":true},
		"*indexes":{"remote":false, "replicate":true}, 
		"*action_profiles":{"remote":false,"replicate":true},
		"*account_profiles":{"remote":false,"replicate":true}
	}
    },
    "stor_db": {
        "db_type": "*internal",
    },
    "admins": {
        "enabled": true,
    }
}
`
	ng1 := engine.TestEngine{
		ConfigJSON:       content1,
		Encoding:         *utils.Encoding,
		GracefulShutdown: true,
	}
	ng2 := engine.TestEngine{
		ConfigJSON:       content2,
		Encoding:         *utils.Encoding,
		GracefulShutdown: true,
	}
	client1, _ := ng1.Run(t)
	client2, _ := ng2.Run(t)
	time.Sleep(100 * time.Millisecond)
	t.Run("LoadProfilesEngine1", func(t *testing.T) {
		attrPrf := &utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant:    utils.CGRateSorg,
				ID:        "ATTR_1",
				FilterIDs: []string{"*string:~*opts.*context:*sessions"},
				Attributes: []*utils.ExternalAttribute{
					{
						Path:  "*req.Destinations",
						Type:  utils.MetaConstant,
						Value: "1008",
					},
				},
			},
		}
		var reply string
		if err := client1.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
			attrPrf, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
		chgrsPrf := &utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CHARGER_1",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{"*none"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
			APIOpts: nil,
		}

		if err := client1.Call(context.Background(), utils.AdminSv1SetChargerProfile,
			chgrsPrf, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	})
	t.Run("Engine2GetProfiles", func(t *testing.T) {
		getProfilesShouldFail(t, client2)
		// waiting the replicator on first engine load the batch after interval
		time.Sleep(1 * time.Second)
		getProfilesShouldSucceed(t, client2)
	})
}
func getProfilesShouldFail(t *testing.T, client *birpc.Client) {
	t.Helper()

	var (
		attr utils.APIAttributeProfile
		chgr *utils.ChargerProfile
	)

	err := client.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1"}, &attr)
	if err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%v>, received: <%v>", utils.ErrNotFound, err)
	}

	err = client.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "CHARGER_1"}, &chgr)
	if err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%v>, received: <%v>", utils.ErrNotFound, err)
	}
}

func getProfilesShouldSucceed(t *testing.T, client *birpc.Client) {
	t.Helper()

	var (
		attr utils.APIAttributeProfile
		chgr *utils.ChargerProfile
	)

	if err := client.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1"}, &attr); err != nil {
		t.Errorf("failed to get attribute profile: %v", err)
	}
	if err := client.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "CHARGER_1"}, &chgr); err != nil {
		t.Errorf("failed to get charger profile: %v", err)
	}
}
