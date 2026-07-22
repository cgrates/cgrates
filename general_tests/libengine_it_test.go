//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"errors"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var (
	libengCfg       *config.CGRConfig
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
	switch *utils.DBType {
	case utils.MetaInternal:
		libengConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		libengConfigDIR = "tutmysql"
	case utils.MetaMongo:
		libengConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		libengConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown database type")
	}

	for _, test := range testLibEngineIT {
		t.Run(libengConfigDIR, test)
	}
}

func testLibengITInitCfg(t *testing.T) {
	libengCfgPath = path.Join(*utils.DataDir, "conf", "samples", libengConfigDIR)
	var err error
	libengCfg, err = config.NewCGRConfigFromPath(libengCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testLibengITInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(libengCfg); err != nil {
		t.Fatal(err)
	}
}

func testLibengITInitStorDb(t *testing.T) {
	if err := engine.InitStorDb(libengCfg); err != nil {
		t.Fatal(err)
	}
}

func testLibengITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(libengCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testLibengITRPCConnection(t *testing.T) {
	cgrCfg := &config.RemoteHost{
		ID:              "a4f3f",
		Address:         "localhost:2012",
		Transport:       "*json",
		ConnectAttempts: 2,
		Reconnects:      5,
		ConnectTimeout:  1 * time.Second,
		ReplyTimeout:    25 * time.Millisecond,
		TLS:             false,
	}
	args := &utils.DurationArgs{
		Duration: 50 * time.Millisecond,
		APIOpts:  make(map[string]any),
		Tenant:   "cgrates.org",
	}
	var reply string
	conn, err := engine.NewRPCConnection(context.Background(),
		cgrCfg, "", "", "",
		cgrCfg.ConnectAttempts, cgrCfg.Reconnects,
		cgrCfg.MaxReconnectInterval, cgrCfg.ConnectTimeout,
		cgrCfg.ReplyTimeout, nil, false, "*localhost",
		"a4f3f", new(ltcache.Cache))
	if err != nil {
		t.Error(err)
	}
	//We check if we get a reply timeout error when calling a sleep bigger than the reply timeout from connection config.
	if err := conn.Call(context.Background(), utils.CoreSv1Sleep, args,
		&reply); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected: %v, received %v", context.DeadlineExceeded, err)
	}

}

func testLibengITStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
