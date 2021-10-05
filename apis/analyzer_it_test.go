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

package apis

import (
	"os"
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	anzCfgPath string
	anzCfg     *config.CGRConfig
	anzBiRPC   *birpc.Client

	sTestsAnz = []func(t *testing.T){}
)

func TestAnalyzerSIT(t *testing.T) {
	for _, stest := range sTestsAnz {
		t.Run("TestAnalyzerSIT", stest)
	}
}

func testAnalyzerSInitCfg(t *testing.T) {
	var err error
	if err := os.RemoveAll("/tmp/analyzers/"); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll("/tmp/analyzers/", 0700); err != nil {
		t.Fatal(err)
	}
	anzCfgPath = path.Join(*dataDir, "conf", "samples", "analyzers")
	anzCfg, err = config.NewCGRConfigFromPath(context.Background(), anzCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAnalyzerSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(anzCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAnalyzerSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(anzCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAnalyzerSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(anzCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testAnzBiSRPCConn(t *testing.T) {
	var err error
	anzBiRPC, err = newRPCClient(anzCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}
