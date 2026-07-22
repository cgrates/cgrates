//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"fmt"
	"os"
	"slices"
	"strings"
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

func TestReplicatorFailedPostsNoInterval(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	failedDir := t.TempDir()

	primaryCfg := fmt.Sprintf(`{
"general": {
	"node_id": "primary",
	"reconnects": 1
},
"listen": {
	"rpc_json": ":4112",
	"rpc_gob": ":4113",
	"http": ":4180",
	"birpc_json": ""
},
"data_db": {
	"db_type": "*internal",
	"replication_conns": ["rpl"],
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
				"address": "127.0.0.1:4123",
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
		&reply); err == nil || !strings.Contains(err.Error(), "connect: connection refused") {
		t.Fatal("expected connection refused error")
	}
	if err := primaryClient.Call(context.Background(), utils.APIerSv1SetDestination,
		&utils.AttrSetDestination{
			Id:       "DST_1001",
			Prefixes: []string{"+49"},
		},
		&reply); err == nil || !strings.Contains(err.Error(), "connect: connection refused") {
		t.Fatal("expected connection refused error")
	}

	entries, err := os.ReadDir(failedDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatal("expected 2 .gob files in failed dir")
	}

	targetCfg := `{
"general": {
	"node_id": "target",
	"reconnects": 1
},
"listen": {
	"rpc_json": ":4122",
	"rpc_gob": ":4123",
	"http": ":4190",
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
