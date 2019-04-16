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
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dspCfgPath string
	dspCfg     *config.CGRConfig
	dspRPC     *rpc.Client

	sTestsDspDspv1 = []func(t *testing.T){
		testDspITLoadConfig,
		testDspITResetDataDB,
		testDspITResetStorDb,
		testDspITStartEngine,
		testDspITRPCConn,
		testDspITLoadData,

		testDspDspv1GetProfileForEvent,

		testDspITStopCgrEngine,
	}
)

//Test start here
func TestDspDspv1SMySQL(t *testing.T) {
	dspCfgPath = path.Join(*dataDir, "conf", "samples", "dispatchers", "dispatchers")
	for _, stest := range sTestsDspDspv1 {
		t.Run("TestDspDspv1", stest)
	}
}

func TestDspDspv1SMongo(t *testing.T) {
	dspCfgPath = path.Join(*dataDir, "conf", "samples", "dispatchers", "dispatchers_mongo")
	for _, stest := range sTestsDspDspv1 {
		t.Run("TestDspDspv1", stest)
	}
}

func testDspITLoadConfig(t *testing.T) {
	var err error
	if dspCfg, err = config.NewCGRConfigFromPath(dspCfgPath); err != nil {
		t.Error(err)
	}
}

func testDspITResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(dspCfg); err != nil {
		t.Fatal(err)
	}
}

func testDspITResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(dspCfg); err != nil {
		t.Fatal(err)
	}
}

func testDspITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dspCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testDspITRPCConn(t *testing.T) {
	var err error
	dspRPC, err = jsonrpc.Dial("tcp", dspCfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func testDspITLoadData(t *testing.T) {
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", dspCfgPath, "-path", path.Join(*dataDir, "tariffplans", "dispatchers"))

		if err := loader.Start(); err != nil {
			t.Error(err)
		}
		loader.Wait()
		wchan <- struct{}{}
	}()
	select {
	case <-wchan:
	case <-time.After(5 * time.Second):
		t.Errorf("cgr-loader failed: ")
	}
	time.Sleep(500 * time.Millisecond)
}

func testDspDspv1GetProfileForEvent(t *testing.T) {
	arg := dispatchers.DispatcherEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testDspv1",
			Event: map[string]interface{}{
				utils.EVENT_NAME: "Event1",
			},
		},
		Subsystem: utils.META_ANY,
	}
	var reply engine.DispatcherProfile
	expected := engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "EVENT1",
		Subsystems:     []string{utils.META_ANY},
		FilterIDs:      []string{"*string:~EventName:Event1"},
		StrategyParams: make(map[string]interface{}),
		Strategy:       utils.MetaWeight,
		Weight:         30,
		Hosts: engine.DispatcherHostProfiles{
			&engine.DispatcherHostProfile{
				ID:        "ALL2",
				FilterIDs: []string{},
				Weight:    20,
				Params:    make(map[string]interface{}),
			},
			&engine.DispatcherHostProfile{
				ID:        "ALL",
				FilterIDs: []string{},
				Weight:    10,
				Params:    make(map[string]interface{}),
			},
		},
	}
	expected.Hosts.Sort()
	if err := dspRPC.Call(utils.DispatcherSv1GetProfileForEvent, &arg, &reply); err != nil {
		t.Error(err)
	}
	reply.Hosts.Sort()
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("expected: %s , received: %s", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testDspITStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
