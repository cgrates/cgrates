//go:build flaky

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	"nodeID": "shared_client_lock",
	"replyTimeout": "10m",
},
"logger": {
	"level": 7
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
				"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
	}
    	}
	},
},

"cdrs": {
	"enabled": true,
	"conns": {
		"*chargers": [{"connIDs": ["*localhost"]}]
	},
},
"chargers": {
	"enabled": true,
	"conns": {
		"*attributes": [{"connIDs": ["*localhost"]}]
	},
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
	if err := engine.InitDB(clntLockCfg); err != nil {
		t.Fatal(err)
	}
}

func testSharedClientLockStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(clntLockCfgPath, clntLockDelay); err != nil {
		t.Fatal(err)
	}
}

func testSharedClientLockRpcConn(t *testing.T) {
	clntLockRPC = engine.NewRPCClient(t, clntLockCfg.ListenCfg(), *utils.Encoding)
}

func testSharedClientLockSetProfiles(t *testing.T) {
	var reply string
	err := clntLockRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
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
		&utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant: "cgrates.org",
				ID:     "ATTR_TEST",
				Attributes: []*utils.ExternalAttribute{
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
