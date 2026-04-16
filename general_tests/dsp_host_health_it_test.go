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
	"net/rpc/jsonrpc"
	"slices"
	"sort"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestDispatcherHostHealth(t *testing.T) {
	cfgMgmt := `{
"general": { 
	"node_id": "mgmt"
},
"listen": {
	"rpc_json": "127.0.0.1:6012",
	"rpc_gob": "127.0.0.1:6013",
	"http": "127.0.0.1:6080"
},
"stor_db": {
	"db_type": "*internal"
},
"apiers": {
	"enabled": true
}
}`

	cfgDsp := `{
"general": {
	"node_id": "dsp"
},
"listen": {
	"rpc_json": "127.0.0.1:3012",
	"rpc_gob": "127.0.0.1:3013",
	"http": "127.0.0.1:3080"
},
"stor_db": {
	"db_type": "*internal"
},
"dispatchers": {
	"enabled": true
}
}`

	cfgR1 := `{
"general": {
	"node_id": "rater1"
},
"listen": {
	"rpc_json": "127.0.0.1:5012",
	"rpc_gob": "127.0.0.1:5013",
	"http": "127.0.0.1:5080"
},
"stor_db": {
	"db_type": "*internal"
},
"rals": {
	"enabled": true
}
}`

	cfgR2 := `{
"general": {
	"node_id": "rater2"
},
"listen": {
	"rpc_json": "127.0.0.1:5112",
	"rpc_gob": "127.0.0.1:5113",
	"http": "127.0.0.1:5180"
},
"stor_db": {
	"db_type": "*internal"
},
"rals": {
	"enabled": true
}
}`

	tpFiles := map[string]string{
		utils.DispatcherHostsCsv: `#Tenant,ID,Address,Transport,TLS
cgrates.org,RATER1,127.0.0.1:5012,*json,false
cgrates.org,RATER2,127.0.0.1:5112,*json,false`,
		utils.DispatcherProfilesCsv: `#Tenant,ID,Subsystems,FilterIDs,ActivationInterval,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters,Weight
cgrates.org,DSP_ANY,*any,,,*weight,,RATER1,,10,false,,10
cgrates.org,DSP_ANY,,,,,,RATER2,,5,,,`,
	}

	ngMgmt := TestEnvironment{
		ConfigJSON: cfgMgmt,
		TpFiles:    tpFiles,
	}
	clientMgmt, _ := ngMgmt.Setup(t, 0)

	ngR1 := TestEnvironment{
		ConfigJSON:     cfgR1,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngR1.Setup(t, 0)

	ngR2 := TestEnvironment{
		ConfigJSON:     cfgR2,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	ngR2.Setup(t, 0)

	ngDsp := TestEnvironment{
		ConfigJSON:     cfgDsp,
		PreserveDataDB: true,
		PreserveStorDB: true,
	}
	clientDsp, _ := ngDsp.Setup(t, 0)

	// nil CGREvent used to panic
	var reply string
	if err := clientDsp.Call(utils.CoreSv1Ping,
		&utils.CGREventWithArgDispatcher{}, &reply); err != nil {
		t.Fatalf("Ping dispatcher (nil CGREvent): %v", err)
	}
	if reply != utils.Pong {
		t.Errorf("Ping dispatcher (nil CGREvent) = %q, want %q", reply, utils.Pong)
	}

	// same but with tenant set explicitly
	reply = ""
	if err := clientDsp.Call(utils.CoreSv1Ping,
		&utils.CGREventWithArgDispatcher{
			CGREvent: &utils.CGREvent{Tenant: "cgrates.org"},
		}, &reply); err != nil {
		t.Fatalf("Ping dispatcher (with tenant): %v", err)
	}
	if reply != utils.Pong {
		t.Errorf("Ping dispatcher (with tenant) = %q, want %q", reply, utils.Pong)
	}

	// get hosts and their addresses, connect directly, call Status
	var hostIDs []string
	if err := clientMgmt.Call(utils.APIerSv1GetDispatcherHostIDs,
		utils.TenantArgWithPaginator{
			TenantArg: utils.TenantArg{Tenant: "cgrates.org"},
		}, &hostIDs); err != nil {
		t.Fatalf("GetDispatcherHostIDs: %v", err)
	}
	sort.Strings(hostIDs)
	wantHostIDs := []string{"RATER1", "RATER2"}
	if !slices.Equal(hostIDs, wantHostIDs) {
		t.Fatalf("GetDispatcherHostIDs = %v, want %v", hostIDs, wantHostIDs)
	}

	wantNodeID := map[string]string{
		"RATER1": "rater1",
		"RATER2": "rater2",
	}
	for _, id := range hostIDs {
		var host engine.DispatcherHost
		if err := clientMgmt.Call(utils.APIerSv1GetDispatcherHost,
			&utils.TenantID{Tenant: "cgrates.org", ID: id},
			&host); err != nil {
			t.Fatalf("GetDispatcherHost(%s): %v", id, err)
		}
		if len(host.Conns) == 0 {
			t.Fatalf("host %s has no connections", id)
		}

		addr := host.Conns[0].Address
		client, err := jsonrpc.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("dial %s (%s): %v", id, addr, err)
		}
		defer client.Close()

		var hostStatus map[string]any
		if err := client.Call(utils.CoreSv1Status, nil, &hostStatus); err != nil {
			t.Fatalf("Status on %s (%s): %v", id, addr, err)
		}
		hostNodeID, _ := hostStatus[utils.NodeID].(string)
		if hostNodeID != wantNodeID[id] {
			t.Errorf("host %s: NodeID = %q, want %q", id, hostNodeID, wantNodeID[id])
		}
	}
}
