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

package ees

import (
	"net/rpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	httpJSONMapConfigDir string
	httpJSONMapCfgPath   string
	httpJSONMapCfg       *config.CGRConfig
	httpJSONMapRpc       *rpc.Client

	sTestsHTTPJsonMap = []func(t *testing.T){
		testHTTPJsonMapLoadConfig,
		testHTTPJsonMapResetDataDB,
		testHTTPJsonMapResetStorDb,
		testHTTPJsonMapStartEngine,
		testHTTPJsonMapRPCConn,

		testStopCgrEngine,
	}
)

func TestHTTPJsonMapExport(t *testing.T) {
	httpJSONMapConfigDir = "ees"
	for _, stest := range sTestsHTTPJsonMap {
		t.Run(httpJSONMapConfigDir, stest)
	}
}

func testHTTPJsonMapLoadConfig(t *testing.T) {
	var err error
	httpJSONMapCfgPath = path.Join(*dataDir, "conf", "samples", httpJSONMapConfigDir)
	if httpJSONMapCfg, err = config.NewCGRConfigFromPath(httpJSONMapCfgPath); err != nil {
		t.Error(err)
	}
}

func testHTTPJsonMapResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(httpJSONMapCfg); err != nil {
		t.Fatal(err)
	}
}

func testHTTPJsonMapResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(httpJSONMapCfg); err != nil {
		t.Fatal(err)
	}
}

func testHTTPJsonMapStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(httpJSONMapCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testHTTPJsonMapRPCConn(t *testing.T) {
	var err error
	httpJSONMapRpc, err = newRPCClient(httpJSONMapCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}
