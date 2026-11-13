//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	"nodeID": "primary",
	"reconnects": 1
},
"listen": {
	"rpcJSON": ":4012",
	"rpcGOB": ":4013",
	"http": ":4080"
},
"db": {
	"dbConns": {
		"*default": {
			"replicationConns": ["rpl"],
			"replicationFailedDir": %q,
			"replicationInterval": "100ms",
		}
	},
	"items": {
		"*accounts": {"replicate": true},
		"*attributeProfiles": {"replicate": true}
	}
},
"rpcConns": {
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
	"nodeID": "target",
	"reconnects": 1
},
"listen": {
	"rpcJSON": ":4022",
	"rpcGOB": ":4023",
	"http": ":4090"
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
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

func TestReplicatorFailedPostsNoInterval(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	failedDir := t.TempDir()

	primaryCfg := fmt.Sprintf(`{
"general": {
	"nodeID": "primary",
	"reconnects": 1
},
"listen": {
	"rpcJSON": ":4112",
	"rpcGOB": ":4113",
	"http": ":4180"
},
"db": {
	"dbConns": {
		"*default": {
			"replicationConns": ["rpl"],
			"replicationFailedDir": %q
		}
	},
	"items": {
		"*accounts": {"replicate": true},
		"*attributeProfiles": {"replicate": true}
	}
},
"rpcConns": {
	"rpl": {
		"conns": [
			{
				"address": "127.0.0.1:4123",
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

	var primaryAcnt utils.Account
	if err := primaryClient.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
		},
		&primaryAcnt); err != nil {
		t.Fatalf("account 1001 not found on primary: %v", err)
	}

	var primaryAttr utils.APIAttributeProfile
	if err := primaryClient.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ATTR_1001",
			},
		},
		&primaryAttr); err != nil {
		t.Fatalf("attribute profile ATTR_1001 not found on primary: %v", err)
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
	"nodeID": "target",
	"reconnects": 1
},
"listen": {
	"rpcJSON": ":4122",
	"rpcGOB": ":4123",
	"http": ":4190"
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
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

func TestReplicatorFailedReplicationFiles(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	failedDir := t.TempDir()
	node1NG := engine.TestEngine{
		ConfigJSON: fmt.Sprintf(`{
"general": {
	"nodeID": "primary",
	"reconnects": 1
},
"listen": {
	"rpcJSON": ":4212",
	"rpcGOB": ":4213",
	"http": ":4280"
},
"db": {
	"dbConns": {
		"*default": {
			"replicationConns": ["node2Conn"],
			"replicationFailedDir": %q
		}
	},
	"items": {
		"*attributeProfiles": {"replicate": true}
	}
},
"rpcConns": {
	"node2Conn": {
		"conns": [
			{
				"address": "127.0.0.1:4223",
				"transport": "*gob"
			}
		]
	}
},
"admins": {
	"enabled": true
}
}`, failedDir),
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	client1, _ := node1NG.Run(t)

	node2NG := engine.TestEngine{
		ConfigJSON: `{
"general": {
	"nodeID": "node2",
	"reconnects": 1
},
"listen": {
	"rpcJSON": ":4222",
	"rpcGOB": ":4223",
	"http": ":4290"
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
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
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	client2, cfg2 := node2NG.Run(t)

	cleanFailedDir := func(t *testing.T) {
		t.Helper()
		entries, err := os.ReadDir(failedDir)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if err := os.RemoveAll(filepath.Join(failedDir, entry.Name())); err != nil {
				t.Fatal(err)
			}
		}
	}

	t.Run("ReplayRemoveUndecodable", func(t *testing.T) {
		cleanFailedDir(t)
		unDecFile := filepath.Join(failedDir, "file1"+utils.GOBSuffix)
		if err := os.WriteFile(unDecFile, []byte("gob file"), 0644); err != nil {
			t.Fatal(err)
		}
		var reply string
		if err := client1.Call(context.Background(), utils.AdminSv1ReplayFailedReplications,
			apis.ReplayFailedReplicationsArgs{}, &reply); err == nil {
			t.Fatal("expected replay to fail undecodable file")
		}
		if _, err := os.Stat(unDecFile); err != nil {
			t.Error("undecodable file was removed")
		}
	})

	t.Run("ReplaySaveFailedPathError", func(t *testing.T) {
		cleanFailedDir(t)
		node2NG.Stop(t)
		var reply string
		if err := client1.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
			&utils.APIAttributeProfileWithAPIOpts{
				APIAttributeProfile: &utils.APIAttributeProfile{
					Tenant: "cgrates.org",
					ID:     "ATTR_1",
					Attributes: []*utils.ExternalAttribute{
						{Path: "*req.Destination", Type: utils.MetaConstant, Value: "1001"},
					},
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}
		failedFile := filepath.Join(failedDir, utils.AttributeProfilePrefix+"cgrates.org:"+"ATTR_1"+utils.GOBSuffix)
		before, err := os.ReadFile(failedFile)
		if err != nil {
			t.Fatal("saved failed files doesn't exist")
		}
		if err := client1.Call(context.Background(), utils.AdminSv1ReplayFailedReplications,
			apis.ReplayFailedReplicationsArgs{
				FailedPath: filepath.Join(t.TempDir(), "missing"),
			}, &reply); err == nil {
			t.Fatal("expected replay to fail when FailedPath cannot be written")
		}
		after, err := os.ReadFile(failedFile)
		if err != nil {
			t.Fatal("saved failed files removed when FailedPath cannot be written")
		}
		if !bytes.Equal(before, after) {
			t.Error("saved failed files changed when FailedPath cannot be written")
		}
		node2NG.Start(t)
		client2 = engine.NewRPCClient(t, cfg2.ListenCfg(), *utils.Encoding)
		if err := client1.Call(context.Background(), utils.AdminSv1ReplayFailedReplications,
			apis.ReplayFailedReplicationsArgs{}, &reply); err != nil {
			t.Fatal(err)
		}
		var attr utils.APIAttributeProfile
		if err := client2.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1"}},
			&attr); err != nil {
			t.Fatal(err)
		}
		if len(attr.Attributes) != 1 || attr.Attributes[0].Value != "1001" {
			t.Errorf("got wrong attributes %+v", attr.Attributes)
		}
	})

}
