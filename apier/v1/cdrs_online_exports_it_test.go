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
package v1

import (
	"os"
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdrExportsCfgPath string
	cdrExportsCfg     *config.CGRConfig
	cdrExportsRPC     *birpc.Client
)

var sTestsCDRsOnlineExports = []func(t *testing.T){
	testCDRsExportsInitCfg,
	testCDRsExportsInitDB,
	testCDRsExportsStartEngine,
	testCDRsExportsRPCConn,
	testCDRsOnlineExportsClearAndSet,
	testCDRsExportsKillEngine,
}

// ConfigSv1.SetConfig returns OK when clearing online_cdr_exports but the old values remain.
func TestCDRsOnlineExportsIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal:
	case utils.MetaMongo:
	default:
		t.SkipNow()
	}
	for _, stest := range sTestsCDRsOnlineExports {
		t.Run("TestCDRsOnlineExports", stest)
	}
}

func testCDRsExportsInitCfg(t *testing.T) {
	if err := os.MkdirAll("/tmp/cgrates_exports", 0755); err != nil {
		t.Fatal(err)
	}
	cdrExportsCfgPath = path.Join(*utils.DataDir, "conf", "samples", "tutmysql")
	var err error
	cdrExportsCfg, err = config.NewCGRConfigFromPath(cdrExportsCfgPath)
	if err != nil {
		t.Fatal(err)
	}
}

func testCDRsExportsInitDB(t *testing.T) {
	if err := engine.InitDataDB(cdrExportsCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(cdrExportsCfg); err != nil {
		t.Fatal(err)
	}
}

func testCDRsExportsStartEngine(t *testing.T) {
	if err := os.MkdirAll("/tmp/cgrates_exports", 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StopStartEngine(cdrExportsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testCDRsExportsRPCConn(t *testing.T) {
	var err error
	cdrExportsRPC, err = newRPCClient(cdrExportsCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

// testCDRsOnlineExportsClearAndSet sets online_cdr_exports to a non-empty list
// then clears it with an empty list and verifies the field is cleared.
func testCDRsOnlineExportsClearAndSet(t *testing.T) {
	t.Skip("online_cdr_exports not cleared")
	var reply string
	if err := cdrExportsRPC.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
		Config: map[string]any{
			"ees": map[string]any{
				"enabled": true,
				"exporters": []map[string]any{
					{
						"id":   "http_billing_event",
						"type": utils.MetaNone,
					},
				},
			},
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expected OK received: %s", reply)
	}

	if err := cdrExportsRPC.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
		Config: map[string]any{
			"cdrs": map[string]any{
				"online_cdr_exports": []string{"http_billing_event"},
			},
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expected OK received: %s", reply)
	}

	var rpl map[string]any
	if err := cdrExportsRPC.Call(context.Background(), utils.ConfigSv1GetConfig, &config.SectionWithAPIOpts{
		Section: "cdrs",
	}, &rpl); err != nil {
		t.Fatal(err)
	}
	exports := rpl["cdrs"].(map[string]any)["online_cdr_exports"].([]any)
	if len(exports) != 1 || exports[0] != "http_billing_event" {
		t.Fatalf("Expected [http_billing_event] received: %+v", exports)
	}

	// clearing online_cdr_exports with an empty list
	// SetConfig returns OK but the field is not updated in the running config
	if err := cdrExportsRPC.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
		Config: map[string]any{
			"cdrs": map[string]any{
				"online_cdr_exports": []string{},
			},
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Expected OK received: %s", reply)
	}

	if err := cdrExportsRPC.Call(context.Background(), utils.ConfigSv1GetConfig, &config.SectionWithAPIOpts{
		Section: "cdrs",
	}, &rpl); err != nil {
		t.Fatal(err)
	}
	exports = rpl["cdrs"].(map[string]any)["online_cdr_exports"].([]any)
	if len(exports) != 0 {
		t.Fatalf("Expected empty online_cdr_exports, received: %+v", exports)
	}
}

func testCDRsExportsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
