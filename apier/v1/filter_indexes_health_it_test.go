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
MERCHANTABILITY or FIdxTNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tFIdxHRpc *rpc.Client

	sTestsFilterIndexesSHealth = []func(t *testing.T){
		testV1FIdxHLoadConfig,
		testV1FIdxHdxInitDataDb,
		testV1FIdxHResetStorDb,
		testV1FIdxHStartEngine,
		testV1FIdxHRpcConn,
		testV1FIdxHLoadFromFolder,
		testV1FIdxHAccountActionPlansHealth,
		testV1FIdxHReverseDestinationHealth,

		testV1FIdxHStopEngine,
	}
)

// Test start here
func TestFIdxHealthIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		tSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		tSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsFilterIndexesSHealth {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1FIdxHLoadConfig(t *testing.T) {
	tSv1CfgPath = path.Join(*dataDir, "conf", "samples", tSv1ConfDIR)
	var err error
	if tSv1Cfg, err = config.NewCGRConfigFromPath(tSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1FIdxHdxInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1FIdxHResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxHStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxHRpcConn(t *testing.T) {
	var err error
	tFIdxHRpc, err = newRPCClient(tSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1FIdxHLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial2")}
	if err := tFIdxHRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1FIdxHAccountActionPlansHealth(t *testing.T) {
	var reply engine.AccountActionPlanIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetAccountActionPlansIndexHealth, engine.IndexHealthArgsWith2Ch{
		IndexCacheLimit:  -1,
		ObjectCacheLimit: -1,
	}, &reply); err != nil {
		t.Error(err)
	}
	exp := engine.AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{},
		BrokenReferences:          map[string][]string{},
	}
	if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testV1FIdxHReverseDestinationHealth(t *testing.T) {
	var reply engine.ReverseDestinationsIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetReverseDestinationsIndexHealth, engine.IndexHealthArgsWith2Ch{
		IndexCacheLimit:  -1,
		ObjectCacheLimit: -1,
	}, &reply); err != nil {
		t.Error(err)
	}
	exp := engine.ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{},
		BrokenReferences:           map[string][]string{},
	}
	if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testV1FIdxHStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
