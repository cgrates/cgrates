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
	switch *dbType {
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
	if *encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspConfig, "TestDspConfigIT", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspConfigSv1GetJSONSection(t *testing.T) {
	expected := map[string]interface{}{
		"http":         ":6080",
		"http_tls":     "127.0.0.1:2280",
		"rpc_gob":      ":6013",
		"rpc_gob_tls":  "127.0.0.1:2023",
		"rpc_json":     ":6012",
		"rpc_json_tls": "127.0.0.1:2022",
	}
	expected = map[string]interface{}{
		"listen": expected,
	}
	var reply map[string]interface{}
	if err := dispEngine.RPC.Call(utils.ConfigSv1GetConfig, &config.SectionWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cfg12345",
		},
		Section: "listen",
	}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}
