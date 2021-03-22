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
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
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
		testDspDspv1GetProfileForEventWithMethod,
		testDspITStopCgrEngine,
	}
)

//Test start here

func TestDspDspv1(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		dispatcherConfigDIR = "dispatchers_mysql"
	case utils.MetaMongo:
		dispatcherConfigDIR = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	if *encoding == utils.MetaGOB {
		dispatcherConfigDIR += "_gob"
	}
	for _, stest := range sTestsDspDspv1 {
		t.Run(dispatcherConfigDIR, stest)
	}
}

func testDspITLoadConfig(t *testing.T) {
	var err error
	dspCfgPath = path.Join(*dataDir, "conf", "samples", "dispatchers", dispatcherConfigDIR)
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
	dspRPC, err = newRPCClient(dspCfg.ListenCfg())
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
	case <-time.After(time.Second):
		t.Errorf("cgr-loader failed: ")
	}
	time.Sleep(100 * time.Millisecond)
}

func testDspDspv1GetProfileForEvent(t *testing.T) {
	arg := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testDspv1",
		Event: map[string]interface{}{
			utils.EventName: "Event1",
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAny,
		},
	}
	var reply engine.DispatcherProfile
	expected := engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "EVENT1",
		Subsystems:     []string{utils.MetaAny},
		FilterIDs:      []string{"*string:~*req.EventName:Event1"},
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
	if *encoding == utils.MetaGOB { // in gob emtpty slice is encoded as nil
		expected.Hosts[0].FilterIDs = nil
		expected.Hosts[1].FilterIDs = nil
	}
	expected.Hosts.Sort()
	if err := dspRPC.Call(utils.DispatcherSv1GetProfileForEvent, &arg, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Hosts.Sort()
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("expected: %s ,\n received: %s", utils.ToJSON(expected), utils.ToJSON(reply))
	}

	arg2 := &utils.CGREvent{
		ID: "testDspvWithoutTenant",
		Event: map[string]interface{}{
			utils.EventName: "Event1",
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaAny,
		},
	}
	expected.Hosts.Sort()
	if err := dspRPC.Call(utils.DispatcherSv1GetProfileForEvent, &arg2, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Hosts.Sort()
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("expected: %s ,\n received: %s", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testDspDspv1GetProfileForEventWithMethod(t *testing.T) {
	arg := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testDspv2",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			utils.Subsys:               utils.MetaAny,
			utils.OptsDispatcherMethod: utils.DispatcherSv1GetProfileForEvent,
		},
	}
	var reply engine.DispatcherProfile
	expected := engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "EVENT6",
		Subsystems:     []string{utils.MetaAny},
		FilterIDs:      []string{"*string:~*opts.*method:DispatcherSv1.GetProfileForEvent"},
		StrategyParams: make(map[string]interface{}),
		Strategy:       utils.MetaWeight,
		Weight:         20,
		Hosts: engine.DispatcherHostProfiles{
			&engine.DispatcherHostProfile{
				ID:        "SELF",
				FilterIDs: []string{},
				Weight:    20,
				Params:    make(map[string]interface{}),
			},
		},
	}
	if *encoding == utils.MetaGOB { // in gob emtpty slice is encoded as nil
		expected.Hosts[0].FilterIDs = nil
	}
	expected.Hosts.Sort()
	if err := dspRPC.Call(utils.DispatcherSv1GetProfileForEvent, &arg, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Hosts.Sort()
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("expected: %s ,\n received: %s", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testDspITStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
