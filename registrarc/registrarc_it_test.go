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

package registrarc

import (
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestRegistrarC(t *testing.T) {
	dbCfg := engine.DBCfg{
		StorDB: &engine.DBParams{
			Type: utils.StringPointer(utils.MetaInternal),
		},
	}
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaMongo:
		dbCfg.DataDB = engine.MongoDBCfg.DataDB
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	const (
		dspCfg = `{
"general": {
	"node_id": "dispatcher",
	"reconnects": 1
},
"caches": {
	"partitions": {
		"*dispatcher_hosts": {
			"limit": -1,
			"ttl": "150ms"
		}
	}
},
"dispatchers": {
	"enabled": true
}
}`
		workerCfg = `{
"general": {
        "node_id": "%s"
},
"listen": {
        "rpc_json": ":%[2]d12",
        "rpc_gob": ":%[2]d13",
        "http": ":%[2]d80"
},
"rpc_conns": {
        "dispConn": {
                "strategy": "*first",
                "conns": [{
                        "address": "http://127.0.0.1:2080/registrar",
                        "transport": "*http_jsonrpc"
                }]
        }
},
"registrarc": {
        "dispatchers": {
                "enabled": true,
                "registrars_conns": ["dispConn"],
                "hosts": [{
                        "Tenant": "*default",
                        "ID": "hostB",
                        "transport": "*json",
                        "tls": false
                }],
                "refresh_interval": "1s"
        }
}
}`
	)

	disp := engine.TestEngine{
		ConfigJSON: dspCfg,
		DBCfg:      dbCfg,
	}
	client, cfg := disp.Run(t)

	tpFiles := map[string]string{
		utils.DispatcherProfilesCsv: `#Tenant,ID,Subsystems,FilterIDs,ActivationInterval,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters,Weight
cgrates.org,dsp_test,,,,*weight,,hostA,,20,,,
cgrates.org,dsp_test,,,,,,hostB,,10,,,`,
	}

	engine.LoadCSVsWithCGRLoader(t, cfg.ConfigPath, "", nil, tpFiles, "-caches_address=")

	checkNodeID := func(t *testing.T, expected string) {
		t.Helper()
		var status map[string]any
		err := client.Call(context.Background(), utils.CoreSv1Status,
			utils.TenantWithAPIOpts{
				Tenant:  "cgrates.org",
				APIOpts: map[string]any{},
			}, &status)
		if err != nil && expected != "" {
			t.Fatalf("DispatcherSv1.RemoteStatus unexpected err: %v", err)
		}
		nodeID := utils.IfaceAsString(status[utils.NodeID])
		if expected == "" &&
			(err == nil || err.Error() != utils.ErrDSPHostNotFound.Error()) {
			t.Errorf("DispatcherSv1.RemoteStatus err=%q, want %q", err, utils.ErrDSPHostNotFound)
		}
		if nodeID != expected {
			t.Errorf("DispatcherSv1.RemoteStatus nodeID=%q, want %q", nodeID, expected)
		}
	}

	/*
		Currently, only a dispatcher profile can be found in dataDB.
		It references 2 hosts that don't exist yet: hostA (weight=20) and hostB (weight=10).
		Its sorting strategy is "*weight".
	*/

	checkNodeID(t, "") // no hosts registered yet; will fail

	// Workers will be automatically closed at the end of the subtest.
	t.Run("start workers and dispatch", func(t *testing.T) {
		workerB := engine.TestEngine{
			ConfigJSON:     fmt.Sprintf(workerCfg, "workerB", 70),
			DBCfg:          dbCfg,
			PreserveDataDB: true,
			PreserveStorDB: true,
		}
		workerB.Run(t)

		// workerB is now active and has registered hostB.
		// The status request will be dispatched to hostB, because
		// hostA, which should have had priority, has not yet been
		// registered.
		checkNodeID(t, "workerB")

		workerA := engine.TestEngine{
			ConfigJSON:     fmt.Sprintf(workerCfg, "workerA", 60),
			DBCfg:          dbCfg,
			PreserveDataDB: true,
			PreserveStorDB: true,
		}
		workerA.Run(t)

		// workerA is now active and has overwritten hostB's port with
		// its own, instead of registering hostA. The request will be
		// dispatched based on hostB again.
		checkNodeID(t, "workerA")
	})

	time.Sleep(150 * time.Millisecond) // wait for cached hosts to expire
	checkNodeID(t, "")                 // no hosts left
}
