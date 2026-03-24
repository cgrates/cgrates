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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestReplicatorFailedPosts(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	failedDir := t.TempDir()

	// reconnects: 1 so ConnManager gives up quickly and flush writes .gob files
	primaryCfg := fmt.Sprintf(`{
"general": {
	"node_id": "primary",
	"reconnects": 1
},
"listen": {
	"rpc_json": ":4012",
	"rpc_gob": ":4013",
	"http": ":4080"
},
"db": {
	"db_conns": {
		"*default": {
			"replication_conns": ["rpl"],
			"replication_failed_dir": %q,
			"replication_interval": "100ms",
		}
	},
	"items": {
		"*accounts": {"replicate": true},
		"*attribute_profiles": {"replicate": true}
	}
},
"rpc_conns": {
	"rpl": {
		"conns": [
			{
				"address": "127.0.0.1:4023",
				"transport": "*gob"
			}
		]
	}
},
"admins": {
	"enabled": true
}
}`, failedDir)

	primaryNG := engine.TestEngine{
		ConfigJSON: primaryCfg,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
	}
	primaryClient, _ := primaryNG.Run(t)

	var reply string
	if err := primaryClient.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
		},
		&reply); err != nil {
		t.Fatal(err)
	}
	if err := primaryClient.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		&utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant: "cgrates.org",
				ID:     "ATTR_1001",
				Attributes: []*utils.ExternalAttribute{
					{Path: "*req.Account", Type: utils.MetaConstant, Value: "1001"},
				},
			},
		},
		&reply); err != nil {
		t.Fatal(err)
	}

	time.Sleep(300 * time.Millisecond)

	entries, err := os.ReadDir(failedDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected .gob files in failed dir, found none")
	}

	targetCfg := `{
"general": {
	"node_id": "target",
	"reconnects": 1
},
"listen": {
	"rpc_json": ":4022",
	"rpc_gob": ":4023",
	"http": ":4090"
},
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal",
			"opts": {
				"internalDBDumpInterval": "0s",
				"internalDBRewriteInterval": "0s"
			}
		}
	}
},
"admins": {
	"enabled": true
}
}`
	targetNG := engine.TestEngine{
		ConfigJSON: targetCfg,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
	}
	targetClient, _ := targetNG.Run(t)

	if err := primaryClient.Call(context.Background(), utils.AdminSv1ReplayFailedReplications,
		apis.ReplayFailedReplicationsArgs{
			SourcePath: failedDir,
		},
		&reply); err != nil {
		t.Fatal(err)
	}

	var acnt utils.Account
	if err := targetClient.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
		},
		&acnt); err != nil {
		t.Fatalf("account 1001 not found on target: %v", err)
	}

	var attr utils.APIAttributeProfile
	if err := targetClient.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ATTR_1001",
			},
		},
		&attr); err != nil {
		t.Fatalf("attribute profile ATTR_1001 not found on target: %v", err)
	}
	if len(attr.Attributes) != 1 || attr.Attributes[0].Value != "1001" {
		t.Errorf("expected attribute value 1001, got %+v", attr.Attributes)
	}

	entries, err = os.ReadDir(failedDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected failed dir to be empty, found %d entries", len(entries))
	}
}
