//go:build flaky

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
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
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
	"node_id": "shared_client_lock",
	"reply_timeout": "10m",
},
"logger": {
	"level": 7
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"cdrs": {
	"enabled": true,
	"chargers_conns":["*localhost"],
},
"chargers": {
	"enabled": true,
	"attributes_conns": ["*localhost"]
},
"attributes": {
	"enabled": false,
},
"admins": {
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
	if clntLockCfg, err = config.NewCGRConfigFromPath(context.Background(), clntLockCfgPath); err != nil {
		t.Error(err)
	}
	clntLockDelay = 100
}

func testSharedClientLockInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(clntLockCfg); err != nil {
		t.Fatal(err)
	}
}

func testSharedClientLockResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(clntLockCfg); err != nil {
		t.Fatal(err)
	}
}

func testSharedClientLockStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(clntLockCfgPath, clntLockDelay); err != nil {
		t.Fatal(err)
	}
}

func testSharedClientLockRpcConn(t *testing.T) {
	var err error
	clntLockRPC, err = engine.NewRPCClient(clntLockCfg.ListenCfg(), *utils.Encoding)
	if err != nil {
		t.Fatal("Could not connect to engine: ", err.Error())
	}
}

func testSharedClientLockSetProfiles(t *testing.T) {
	var reply string
	err := clntLockRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&apis.ChargerWithAPIOpts{
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

	err = clntLockRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		&engine.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &engine.APIAttributeProfile{
				Tenant: "cgrates.org",
				ID:     "ATTR_TEST",
				Attributes: []*engine.ExternalAttribute{
					{
						Path:  "*req.Test",
						Type:  utils.MetaConstant,
						Value: "TestValue",
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
	argsEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "cdr_test_event",
		Event:  make(map[string]any),
		APIOpts: map[string]any{
			utils.MetaChargers: true,
		},
	}
	var reply string
	err := clntLockRPC.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply)
	if err == nil || !strings.Contains(err.Error(), "use of closed network connection") {
		t.Error("Unexpected error returned", err)
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
