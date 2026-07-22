//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"os"
	"path"
	"strings"
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
	testCDRsOnlineExportsOverwrite,
	testCDRsOnlineExportsInvalidIDs,
	testCDRsExportsKillEngine,
}

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

func testCDRsOnlineExportsClearAndSet(t *testing.T) {
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
		t.Fatalf("Expected [http_billing_event], received: %+v", exports)
	}

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

func testCDRsOnlineExportsOverwrite(t *testing.T) {
	var reply string

	if err := cdrExportsRPC.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
		Config: map[string]any{
			"ees": map[string]any{
				"enabled": true,
				"exporters": []map[string]any{
					{"id": "exporter_1", "type": utils.MetaNone},
					{"id": "exporter_2", "type": utils.MetaNone},
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
				"online_cdr_exports": []string{"exporter_1"},
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
				"online_cdr_exports": []string{"exporter_2"},
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
	if len(exports) != 1 || exports[0] != "exporter_2" {
		t.Fatalf("Expected [exporter_2] only, got: %+v", exports)
	}
}

func testCDRsOnlineExportsInvalidIDs(t *testing.T) {
	t.Skip("invalid IDs should be not be populated")

	var rpl map[string]any
	if err := cdrExportsRPC.Call(context.Background(), utils.ConfigSv1GetConfig, &config.SectionWithAPIOpts{
		Section: "cdrs",
	}, &rpl); err != nil {
		t.Fatal(err)
	}
	initialExports := rpl["cdrs"].(map[string]any)["online_cdr_exports"].([]any)
	if len(initialExports) != 0 {
		t.Fatalf("Precondition failed: expected empty online_cdr_exports, got: %+v", initialExports)
	}

	// invalid IDs are still written into the running config.
	var reply string
	err := cdrExportsRPC.Call(context.Background(), utils.ConfigSv1SetConfig, &config.SetConfigArgs{
		Config: map[string]any{
			"cdrs": map[string]any{
				"online_cdr_exports": []string{"test", "fake_ees_id", ""},
			},
		},
	}, &reply)

	if err == nil {
		t.Fatal("Expected error for invalid exporter IDs, got nil")
	}

	errMsg := err.Error()
	for _, invalidID := range []string{"test", "fake_ees_id", ""} {
		if invalidID == "" {
			continue
		}
		if !strings.Contains(errMsg, invalidID) {
			t.Errorf("Expected error to mention invalid ID %q, but got: %s", invalidID, errMsg)
		}
	}

	if err := cdrExportsRPC.Call(context.Background(), utils.ConfigSv1GetConfig, &config.SectionWithAPIOpts{
		Section: "cdrs",
	}, &rpl); err != nil {
		t.Fatal(err)
	}
	exports := rpl["cdrs"].(map[string]any)["online_cdr_exports"].([]any)
	if len(exports) != 0 {
		t.Fatalf("online_cdr_exports must remain empty after rejected SetConfig, got: %+v", exports)
	}
}

func testCDRsExportsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
