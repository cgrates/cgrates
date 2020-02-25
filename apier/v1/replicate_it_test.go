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
package v1

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sTestsInternalReplicateIT = []func(t *testing.T){
		testInternalReplicateITInitCfg,
		testInternalReplicateITDataFlush,
		testInternalReplicateITStartEngine,
		testInternalReplicateITRPCConn,
		testInternalReplicateLoadDataInEngineTwo,

		// testInternalReplicateSetDestination,

		testInternalReplicateITKillEngine,
	}
)

func TestInternalReplicateIT(t *testing.T) {
	internalCfgDirPath = "internal"
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		engineOneCfgDirPath = "engine1_redis"
		engineTwoCfgDirPath = "engine2_redis"
	case utils.MetaMongo:
		engineOneCfgDirPath = "engine1_mongo"
		engineTwoCfgDirPath = "engine2_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	if *encoding == utils.MetaGOB {
		internalCfgDirPath += "_gob"
	}
	for _, stest := range sTestsInternalReplicateIT {
		t.Run(*dbType, stest)
	}
}

func testInternalReplicateITInitCfg(t *testing.T) {
	var err error
	internalCfgPath = path.Join(*dataDir, "conf", "samples", "remote_replication", internalCfgDirPath)
	internalCfg, err = config.NewCGRConfigFromPath(internalCfgPath)
	if err != nil {
		t.Error(err)
	}
	internalCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(internalCfg)

	// prepare config for engine1
	engineOneCfgPath = path.Join(*dataDir, "conf", "samples",
		"remote_replication", engineOneCfgDirPath)
	engineOneCfg, err = config.NewCGRConfigFromPath(engineOneCfgPath)
	if err != nil {
		t.Error(err)
	}
	engineOneCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()

	// prepare config for engine2
	engineTwoCfgPath = path.Join(*dataDir, "conf", "samples",
		"remote_replication", engineTwoCfgDirPath)
	engineTwoCfg, err = config.NewCGRConfigFromPath(engineTwoCfgPath)
	if err != nil {
		t.Error(err)
	}
	engineTwoCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()

}

func testInternalReplicateITDataFlush(t *testing.T) {
	if err := engine.InitDataDb(engineOneCfg); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if err := engine.InitDataDb(engineTwoCfg); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testInternalReplicateITStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(engineOneCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(engineTwoCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(internalCfgPath, 500); err != nil {
		t.Fatal(err)
	}
}

func testInternalReplicateITRPCConn(t *testing.T) {
	var err error
	engineOneRPC, err = newRPCClient(engineOneCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	engineTwoRPC, err = newRPCClient(engineTwoCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	internalRPC, err = newRPCClient(internalCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testInternalReplicateLoadDataInEngineTwo(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := engineTwoRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testInternalReplicateSetDestination(t *testing.T) {
	//set
	attrs := utils.AttrSetDestination{Id: "TEST_SET_DESTINATION3", Prefixes: []string{"004", "005"}}
	var reply string
	if err := internalRPC.Call(utils.APIerSv1SetDestination, attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	eDst := &engine.Destination{
		Id:       "TEST_SET_DESTINATION3",
		Prefixes: []string{"004", "005"},
	}
	// check
	rpl := &engine.Destination{}
	if err := engineOneRPC.Call(utils.APIerSv1GetDestination, "TEST_SET_DESTINATION3", &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDst, rpl) {
		t.Errorf("Expected: %v,\n received: %v", eDst, rpl)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDestination, "TEST_SET_DESTINATION3", &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDst, rpl) {
		t.Errorf("Expected: %v,\n received: %v", eDst, rpl)
	}

	// remove
	attr := &AttrRemoveDestination{DestinationIDs: []string{"TEST_SET_DESTINATION"}, Prefixes: []string{"004", "005"}}
	if err := internalRPC.Call(utils.APIerSv1RemoveDestination, attr, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetDestination, "TEST_SET_DESTINATION", &rpl); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDestination, "TEST_SET_DESTINATION", &rpl); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

}

func testInternalReplicateITKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
