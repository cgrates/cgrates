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
	"os"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Tests to recover from DB files on new machine
func TestOfflineInternalReplication(t *testing.T) {
	cfg1 := `
{
"general": {
	"node_id": "InternalEngine"
},

"logger": {
    "level": 7
},

"listen": {
	"rpc_json": ":2012",
	"rpc_gob": ":2013",
	"http": ":2080"
},

"rpc_conns": {
	"conn2": {
		"strategy": "*broadcast_sync",
		"conns": [
			{"id": "engine1", "address": "127.0.0.1:2023", "transport":"*gob"},
		]
	}
},

"db": {
	"db_conns": {
		 "*default": {
		 	"replication_conns": ["conn2"],
			"replication_interval": "-1",
			"opts":{
				"internalDBDumpPath": "/tmp/internal_db_master/db",
	        	"internalDBBackupPath": "/tmp/internal_db_master/backup/db",
	        	"internalDBDumpInterval": "500ms",
	        	"internalDBRewriteInterval": "500ms",
	        	"internalDBFileSizeLimit": "3.3KB",
			}
		 }
	},
	"items":{
		"*threshold_profiles": {"remote":false,"replicate":true},
		"*attribute_profiles":{"remote":false,"replicate":true},
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
	"node_id": "InternalEngine2"
},

"logger": {
    "level": 7
},

"listen": {
	"rpc_json": ":2022",
	"rpc_gob": ":2023",
	"http": ":2280"
},

"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal",
			"opts":{
				"internalDBDumpPath": "/tmp/internal_db_slave/db",
	        	"internalDBBackupPath": "/tmp/internal_db_slave/backup/db",
	        	"internalDBDumpInterval": "500ms",
	        	"internalDBRewriteInterval": "500ms",
	        	"internalDBFileSizeLimit": "3.3KB",
			}
		},
	},
},

"thresholds": {
	"enabled": true,
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
cgrates.org,THD_ACNT_1001,*string:~*req.Account:1001,;10,-1,0,0,false,,TOPUP_MONETARY_10,true,`,
	}

	if err := os.MkdirAll("/tmp/internal_db_master/db", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("/tmp/internal_db_master/backup/db", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("/tmp/internal_db_slave/db", 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll("/tmp/internal_db_master"); err != nil {
			t.Error(err)
		}
		if err := os.RemoveAll("/tmp/internal_db_slave"); err != nil {
			t.Error(err)
		}
	})

	ngMaster := engine.TestEngine{
		ConfigJSON:       cfg1,
		Encoding:         *utils.Encoding,
		TpFiles:          tpFiles,
		GracefulShutdown: true,
	}
	ngSlave := engine.TestEngine{
		ConfigJSON:       cfg2,
		Encoding:         *utils.Encoding,
		GracefulShutdown: true,
	}
	clientSlave, _ := ngSlave.Run(t)
	clientMaster, _ := ngMaster.Run(t)
	time.Sleep(500 * time.Millisecond)
	t.Run("GetReplicatedProfiles", func(t *testing.T) {
		var replyAttributeProfile utils.APIAttributeProfile
		if err := clientMaster.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "ATTR_ACNT_1001",
				}}, &replyAttributeProfile); err != nil {
			t.Error(err)
		}

		//replicated profile in second engine
		var replyAttributeProfile2 utils.APIAttributeProfile
		if err := clientSlave.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
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
		if err := clientMaster.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "THD_ACNT_1001",
				},
			}, &rcvTHP); err != nil {
			t.Error(err)
		}

		var rcvTHP2 *engine.ThresholdProfile
		if err := clientSlave.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
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

	t.Run("BackupMasterDB", func(t *testing.T) {
		var rply string
		if err := clientMaster.Call(context.Background(), utils.AdminSv1BackupDB, &apis.BackupParams{}, &rply); err != nil {
			t.Fatal(err)
		}
	})

	// Dump files are populated

	t.Run("KillEngines", func(t *testing.T) {
		ngSlave.Stop(t)
		ngMaster.Stop(t)
	})

	ngMaster2 := engine.TestEngine{
		ConfigJSON:       cfg1,
		Encoding:         *utils.Encoding,
		GracefulShutdown: true,
		PreserveDB:       true,
	}
	ngSlave2 := engine.TestEngine{
		ConfigJSON:       cfg2,
		Encoding:         *utils.Encoding,
		GracefulShutdown: true,
		PreserveDB:       true,
	}

	clientSlave2, _ := ngSlave2.Run(t)
	clientMaster2, _ := ngMaster2.Run(t)
	time.Sleep(500 * time.Millisecond)

	t.Run("CheckProfilesRecoveredForBothEngines", func(t *testing.T) {
		var replyAttributeProfile utils.APIAttributeProfile
		if err := clientMaster2.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "ATTR_ACNT_1001",
				}}, &replyAttributeProfile); err != nil {
			t.Error(err)
		}

		//replicated profile in second engine
		var replyAttributeProfile2 utils.APIAttributeProfile
		if err := clientSlave2.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
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
		if err := clientMaster2.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "THD_ACNT_1001",
				},
			}, &rcvTHP); err != nil {
			t.Error(err)
		}

		var rcvTHP2 *engine.ThresholdProfile
		if err := clientSlave2.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
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

	t.Run("KillEngines", func(t *testing.T) {
		ngMaster2.Stop(t)
		ngSlave2.Stop(t)
	})

	// // Create 2 fresh engines
	if err := os.MkdirAll("/tmp/internal_db2_master/db", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("/tmp/internal_db2_slave/db", 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll("/tmp/internal_db2_master"); err != nil {
			t.Error(err)
		}
		if err := os.RemoveAll("/tmp/internal_db2_slave"); err != nil {
			t.Error(err)
		}
	})

	ngMaster3 := engine.TestEngine{
		ConfigJSON: cfg1,
		DBCfg: engine.DBCfg{DB: &engine.DBParams{
			DBConns: map[string]engine.DBConn{
				utils.MetaDefault: {
					ReplicationConns:    utils.SliceStringPointer([]string{"conn2"}),
					ReplicationInterval: utils.StringPointer("-1"),
					Opts: engine.DBConnOpts{
						InternalDBDumpPath:        utils.StringPointer("/tmp/internal_db2_master/db"),
						InternalDBDumpInterval:    utils.StringPointer("500ms"),
						InternalDBRewriteInterval: utils.StringPointer("500ms"),
					},
				},
			},
			Items: map[string]engine.Item{
				utils.MetaThresholdProfiles: {Replicate: utils.BoolPointer(true)},
				utils.MetaAttributeProfiles: {Replicate: utils.BoolPointer(true)},
			},
		}},
		Encoding:         *utils.Encoding,
		GracefulShutdown: true,
	}
	ngSlave3 := engine.TestEngine{
		ConfigJSON: cfg2,
		DBCfg: engine.DBCfg{DB: &engine.DBParams{
			DBConns: map[string]engine.DBConn{
				utils.MetaDefault: {
					Opts: engine.DBConnOpts{
						InternalDBDumpPath:        utils.StringPointer("/tmp/internal_db2_slave/db"),
						InternalDBDumpInterval:    utils.StringPointer("500ms"),
						InternalDBRewriteInterval: utils.StringPointer("500ms"),
					},
				},
			},
		}},
		Encoding:         *utils.Encoding,
		GracefulShutdown: true,
	}

	clientSlave3, _ := ngSlave3.Run(t)
	clientMaster3, _ := ngMaster3.Run(t)
	time.Sleep(500 * time.Millisecond)

	t.Run("RestoreMasterInternalDB", func(t *testing.T) {
		var rply string
		if err := clientMaster3.Call(context.Background(), utils.AdminSv1RestoreDB, "/tmp/internal_db_master/backup/db", &rply); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CheckRestoredProfiles", func(t *testing.T) {
		var replyAttributeProfile utils.APIAttributeProfile
		if err := clientMaster3.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "ATTR_ACNT_1001",
				}}, &replyAttributeProfile); err != nil {
			t.Error(err)
		}

		// Slave engines dont replicate recoveries
		var replyAttributeProfile2 utils.APIAttributeProfile
		if err := clientSlave3.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "ATTR_ACNT_1001",
				}}, &replyAttributeProfile2); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}

		var rcvTHP *engine.ThresholdProfile
		if err := clientMaster3.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "THD_ACNT_1001",
				},
			}, &rcvTHP); err != nil {
			t.Error(err)
		}

		// Slave engines dont replicate recoveries
		var rcvTHP2 *engine.ThresholdProfile
		if err := clientSlave3.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "THD_ACNT_1001",
				},
			}, &rcvTHP2); err == nil || err.Error() != utils.ErrNotFound.Error() {
			t.Error(err)
		}
	})

}
