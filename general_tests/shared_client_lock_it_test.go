//go:build integration
// +build integration

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

package general_tests

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	clntLockCfgPath string
	clntLockCfg     *config.CGRConfig
	clntLockRPC     *birpc.Client
	clntLockDelay   int

	sTestsClntLock = []func(t *testing.T){
		testSharedClientLockLoadConfig,
		testSharedClientLockInitDataDb,
		testSharedClientLockResetStorDb,
		testSharedClientLockStartEngine,
		testSharedClientLockRpcConn,

		// Test sets the charger and attribute profiles.
		testSharedClientLockSetProfiles,

		// Test simulates a scenario that is occurring when using an older rpcclient library version
		// where a request is dispatched from CDRs to ChargerS via a *localhost connection.
		// The connection is read-locked until ChargerS responds. ChargerS, in turn, sends a request
		// to AttributeS using the same *localhost connection. However, AttributeS is currently unavailable,
		// leading to a "can't find rpc service" error.
		// This error is considered a network error, which prompts a reconnection attempt. The reconnection process
		// involves a lock operation during the disconnect function. As the connection is already read-locked from
		// the initial request, this results in a deadlock and the original request will time out.
		testSharedClientLockCDRsProcessEvent,
		testSharedClientLockStopEngine,
	}
)

func TestSharedClientLockIT(t *testing.T) {
	for _, stest := range sTestsClntLock {
		t.Run("shared client lock", stest)
	}
}

func testSharedClientLockLoadConfig(t *testing.T) {
	content := `{
"general": {
	"log_level": 7,
	"node_id": "shared_client_lock",
	// to notice the deadlock, reply_timeout should be increased
	"reply_timeout": "50ms" 
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"cdrs": {
	"enabled": true,
	"chargers_conns":["*localhost"]
},
"chargers": {
	"enabled": true,
	"attributes_conns": ["*localhost"]
},
"attributes": {
	"enabled": false
},
"apiers": {
	"enabled": true
}	
}`
	folderNameSuffix, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		t.Fatalf("could not generate random number for folder name suffix, err: %s", err.Error())
	}
	clntLockCfgPath = fmt.Sprintf("/tmp/config%d", folderNameSuffix)
	err = os.MkdirAll(clntLockCfgPath, 0755)
	if err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(clntLockCfgPath, "cgrates.json")
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	if clntLockCfg, err = config.NewCGRConfigFromPath(clntLockCfgPath); err != nil {
		t.Error(err)
	}
	clntLockDelay = 100
}

func testSharedClientLockInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(clntLockCfg); err != nil {
		t.Fatal(err)
	}
}

func testSharedClientLockResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(clntLockCfg); err != nil {
		t.Fatal(err)
	}
}

func testSharedClientLockStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(clntLockCfgPath, clntLockDelay); err != nil {
		t.Fatal(err)
	}
}

func testSharedClientLockRpcConn(t *testing.T) {
	clntLockRPC = engine.NewRPCClient(t, clntLockCfg.ListenCfg())
}

func testSharedClientLockSetProfiles(t *testing.T) {
	var reply string
	err := clntLockRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile,
		&v1.ChargerWithAPIOpts{
			ChargerProfile: &engine.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "CHARGER_TEST",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{"ATTR_TEST"},
			},
		}, &reply)
	if err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	err = clntLockRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile,
		&engine.AttributeProfileWithAPIOpts{
			AttributeProfile: &engine.AttributeProfile{
				Tenant: "cgrates.org",
				ID:     "ATTR_TEST",
				Attributes: []*engine.Attribute{
					{
						Path: "*req.Test",
						Type: utils.MetaConstant,
						Value: config.RSRParsers{
							&config.RSRParser{
								Rules: "TestValue",
							},
						},
					},
				},
			},
		}, &reply)
	if err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testSharedClientLockCDRsProcessEvent(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "cdr_test_event",
			Event:  make(map[string]any),
		},
	}
	var reply string
	if err := clntLockRPC.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err == nil ||
		err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("expected: <%v>,\nreceived: <%v>",
			utils.ErrPartiallyExecuted, err)
	}
}

func testSharedClientLockStopEngine(t *testing.T) {
	err := engine.KillEngine(clntLockDelay)
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(clntLockCfgPath)
	if err != nil {
		t.Error(err)
	}
}
