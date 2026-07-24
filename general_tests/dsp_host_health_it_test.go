//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"
	"time"

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
cgrates.org,RATER2,127.0.0.1:5112,*json,false
cgrates.org,DOWN,127.0.0.1:1,*json,false`,
		utils.DispatcherProfilesCsv: `#Tenant,ID,Subsystems,FilterIDs,ActivationInterval,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters,Weight
cgrates.org,DSP_ANY,*any,,,*weight,,RATER1,,10,false,,10
cgrates.org,DSP_ANY,,,,,,RATER2,,5,,,
cgrates.org,DSP_ANY,,,,,,DOWN,,1,,,`,
	}

	ngMgmt := TestEnvironment{
		ConfigJSON: cfgMgmt,
		TpFiles:    tpFiles,
	}
	ngMgmt.Setup(t, 0)

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

	var hostStatus map[string]string
	if err := clientDsp.Call(utils.DispatcherSv1CheckDispatcherProfileHosts,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DSP_ANY"},
		&hostStatus); err != nil {
		t.Fatalf("CheckDispatcherProfileHosts: %v", err)
	}
	for _, hostID := range []string{"RATER1", "RATER2"} {
		dur, ok := hostStatus[hostID]
		if !ok {
			t.Errorf("CheckDispatcherProfileHosts missing %s", hostID)
		} else if _, err := time.ParseDuration(dur); err != nil {
			t.Errorf("CheckDispatcherProfileHosts %s = %q, not a duration", hostID, dur)
		}
	}
	if got := hostStatus["DOWN"]; got != "-1" {
		t.Errorf("CheckDispatcherProfileHosts DOWN = %q, want %q", got, "-1")
	}
}
