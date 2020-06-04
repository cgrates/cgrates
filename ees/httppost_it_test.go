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
	httpPostConfigDir string
	httpPostCfgPath   string
	httpPostCfg       *config.CGRConfig
	httpPostRpc       *rpc.Client

	sTestsHTTPPost = []func(t *testing.T){
		testHTTPPostLoadConfig,
		testHTTPPostResetDataDB,
		testHTTPPostResetStorDb,
		testHTTPPostStartEngine,
		testHTTPPostRPCConn,

		testStopCgrEngine,
	}
)

func TestHTTPPostExport(t *testing.T) {
	httpPostConfigDir = "ees"
	for _, stest := range sTestsHTTPPost {
		t.Run(httpPostConfigDir, stest)
	}
}

func testHTTPPostLoadConfig(t *testing.T) {
	var err error
	httpPostCfgPath = path.Join(*dataDir, "conf", "samples", httpPostConfigDir)
	if httpPostCfg, err = config.NewCGRConfigFromPath(httpPostCfgPath); err != nil {
		t.Error(err)
	}
}

func testHTTPPostResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(httpPostCfg); err != nil {
		t.Fatal(err)
	}
}

func testHTTPPostResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(httpPostCfg); err != nil {
		t.Fatal(err)
	}
}

func testHTTPPostStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(httpPostCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testHTTPPostRPCConn(t *testing.T) {
	var err error
	httpPostRpc, err = newRPCClient(httpPostCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}
