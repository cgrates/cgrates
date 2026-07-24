//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspConfig = []func(t *testing.T){
	testDspConfigSv1GetJSONSection,
}

// Test start here
func TestDspConfigIT(t *testing.T) {
	var config1, config2, config3 string
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "all_mysql"
		config2 = "all2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "all_mongo"
		config2 = "all2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	dispDIR := "dispatchers"
	if *utils.Encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspConfig, "TestDspConfigIT", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspConfigSv1GetJSONSection(t *testing.T) {
	expected := map[string]any{
		"HTTPListen":       ":6080",
		"HTTPTLSListen":    "127.0.0.1:2280",
		"RPCGOBListen":     ":6013",
		"RPCGOBTLSListen":  "127.0.0.1:2023",
		"RPCJSONListen":    ":6012",
		"RPCJSONTLSListen": "127.0.0.1:2022",
	}
	var reply map[string]any
	if err := dispEngine.RPC.Call(utils.ConfigSv1GetJSONSection, &config.StringWithArgDispatcher{
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("cfg12345"),
		},
		Section: "listen",
	}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %+v, received: %+v", expected, reply)
	}
}
