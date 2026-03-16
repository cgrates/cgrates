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
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestReplicatorFailedPosts(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
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
	"http": ":4080",
	"birpc_json": ""
},
"data_db": {
	"db_type": "*internal",
	"replication_conns": ["rpl"],
	"replication_interval": "100ms",
	"replication_failed_dir": %q,
	"items": {
		"*accounts": {"replicate": true},
		"*destinations": {"replicate": true}
	}
},
"stor_db": {
	"db_type": "*internal"
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
"apiers": {
	"enabled": true
}
}`, failedDir)

	primaryNG := engine.TestEngine{
		ConfigJSON: primaryCfg,
		DBCfg:      engine.InternalDBCfg,
	}
	primaryClient, _ := primaryNG.Run(t)

	var reply string
	if err := primaryClient.Call(context.Background(), utils.APIerSv1SetAccount,
		&utils.AttrSetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		},
		&reply); err != nil {
		t.Fatal(err)
	}
	if err := primaryClient.Call(context.Background(), utils.APIerSv1SetDestination,
		&utils.AttrSetDestination{
			Id:       "DST_1001",
			Prefixes: []string{"+49"},
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
	"http": ":4090",
	"birpc_json": ""
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"apiers": {
	"enabled": true
}
}`
	targetNG := engine.TestEngine{
		ConfigJSON: targetCfg,
		DBCfg:      engine.InternalDBCfg,
	}
	targetClient, _ := targetNG.Run(t)

	if err := primaryClient.Call(context.Background(), utils.APIerSv1ReplayFailedReplications,
		v1.ReplayFailedReplicationsArgs{
			SourcePath: failedDir,
		},
		&reply); err != nil {
		t.Fatal(err)
	}

	var acnt *engine.Account
	if err := targetClient.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		},
		&acnt); err != nil {
		t.Fatalf("account 1001 not found on target: %v", err)
	}

	var dst engine.Destination
	if err := targetClient.Call(context.Background(), utils.APIerSv1GetDestination,
		"DST_1001",
		&dst); err != nil {
		t.Fatalf("destination DST_1001 not found on target: %v", err)
	}
	if !slices.Equal(dst.Prefixes, []string{"+49"}) {
		t.Errorf("expected destination prefixes %v, got %v", []string{"+49"}, dst.Prefixes)
	}

	entries, err = os.ReadDir(failedDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected failed dir to be empty, found %d entries", len(entries))
	}
}
