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
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var (
	libengCfg       *config.CGRConfig
	libengRpc       *rpc.Client
	libengCfgPath   string
	libengConfigDIR string

	testLibEngineIT = []func(t *testing.T){
		testLibengITInitCfg,
		testLibengITInitDataDb,
		testLibengITInitStorDb,
		testLibengITStartEngine,
		testLibengITRPCConnection,
		testLibengITStopEngine,
	}
)

func TestLibEngineIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		libengConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		libengConfigDIR = "tutmysql"
	case utils.MetaMongo:
		libengConfigDIR = "tutmongo"
	default:
		t.Fatal("Unknown database type")
	}

	for _, test := range testLibEngineIT {
		t.Run(libengConfigDIR, test)
	}
}

func testLibengITInitCfg(t *testing.T) {
	libengCfgPath = path.Join(*dataDir, "conf", "samples", libengConfigDIR)
	var err error
	libengCfg, err = config.NewCGRConfigFromPath(libengCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testLibengITInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(libengCfg); err != nil {
		t.Fatal(err)
	}
}

func testLibengITInitStorDb(t *testing.T) {
	if err := engine.InitStorDb(libengCfg); err != nil {
		t.Fatal(err)
	}
}

func testLibengITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(libengCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

//
func testLibengITRPCConnection(t *testing.T) {
	cgrCfg := &config.RemoteHost{
		ID:              "a4f3f",
		Address:         "localhost:6012",
		Transport:       "*json",
		ConnectAttempts: 2,
		Reconnects:      5,
		ConnectTimeout:  1 * time.Second,
		ReplyTimeout:    25 * time.Millisecond,
		TLS:             false,
		// ClientKey:       "key1",
	}
	cM := engine.NewConnManager(config.NewDefaultCGRConfig(), nil)
	args := &utils.DurationArgs{
		Duration: 50 * time.Millisecond,
		APIOpts:  make(map[string]interface{}),
		Tenant:   "cgrates.org",
	}
	var reply string
	cM.Call([]string{"a4f3f"}, nil, utils.CoreSv1Sleep, args, &reply)
	_, err := engine.NewRPCConnection(cgrCfg, "", "", "",
		cgrCfg.ConnectAttempts, cgrCfg.Reconnects, cgrCfg.ConnectTimeout, cgrCfg.ReplyTimeout, nil, false, nil, "*localhost",
		"a4f3f", new(ltcache.Cache))
	if err != nil {
		t.Error(err)
	}
}

func testLibengITStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
