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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dspStsCfgPath  string
	dspStsCfg      *config.CGRConfig
	dspStsRPC      *rpc.Client
	instStsCfgPath string
	instStsCfg     *config.CGRConfig
	instStsRPC     *rpc.Client
)

var sTestsDspSts = []func(t *testing.T){
	testDspStsInitCfg,
	testDspStsInitDataDb,
	testDspStsResetStorDb,
	testDspStsStartEngine,
	testDspStsRPCConn,
	testDspStsPing,
	testDspStsLoadData,
	testDspStsAddStsibutesWithPermision,
	testDspStsTestAuthKey,
	testDspStsAddStsibutesWithPermision2,
	testDspStsTestAuthKey2,
	testDspStsKillEngine,
}

//Test start here
func TestDspStatS(t *testing.T) {
	for _, stest := range sTestsDspSts {
		t.Run("", stest)
	}
}

func testDspStsInitCfg(t *testing.T) {
	var err error
	dspStsCfgPath = path.Join(dspDataDir, "conf", "samples", "dispatcher")
	dspStsCfg, err = config.NewCGRConfigFromFolder(dspStsCfgPath)
	if err != nil {
		t.Error(err)
	}
	dspStsCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(dspStsCfg)
	instStsCfgPath = path.Join(dspDataDir, "conf", "samples", "tutmysql")
	instStsCfg, err = config.NewCGRConfigFromFolder(instStsCfgPath)
	if err != nil {
		t.Error(err)
	}
	instStsCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(instStsCfg)
}

func testDspStsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(instStsCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDspStsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(instStsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDspStsStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(instStsCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(dspStsCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDspStsRPCConn(t *testing.T) {
	var err error
	instStsRPC, err = jsonrpc.Dial("tcp", instStsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	dspStsRPC, err = jsonrpc.Dial("tcp", dspStsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}

}

func testDspStsPing(t *testing.T) {
	var reply string
	if err := instStsRPC.Call(utils.StatSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dspStsRPC.Call(utils.StatSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspStsLoadData(t *testing.T) {
	var reply string
	Stss := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(dspDataDir, "tariffplans", "tutorial")}
	if err := instStsRPC.Call("ApierV1.LoadTariffPlanFromFolder", Stss, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testDspStsAddStsibutesWithPermision(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AuthKey",
		Contexts:  []string{utils.MetaAuth},
		FilterIDs: []string{"*string:APIKey:12345"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FieldName:  utils.APIMethods,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("ThresholdSv1.GetThSessionholdsForEvent", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := instStsRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instStsRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspStsTestAuthKey(t *testing.T) {
	var reply []string
	args := ArgsStatProcessEventWithApiKey{
		APIKey: "12345",
		StatsArgsProcessEvent: engine.StatsArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account:    "1001",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:      time.Duration(135 * time.Second),
					utils.COST:       123.0,
					utils.PDD:        time.Duration(12 * time.Second)}},
		}}
	if err := dspStsRPC.Call(utils.StatSv1ProcessEvent,
		args, &reply); err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}

	args2 := TntIDWithApiKey{
		APIKey: "12345",
		TenantID: utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		},
	}

	var metrics map[string]string
	if err := dspStsRPC.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspStsAddStsibutesWithPermision2(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AuthKey",
		Contexts:  []string{utils.MetaAuth},
		FilterIDs: []string{"*string:APIKey:12345"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FieldName:  utils.APIMethods,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("StatSv1.ProcessEvent&StatSv1.GetQueueStringMetrics", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := instStsRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instStsRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspStsTestAuthKey2(t *testing.T) {
	var reply []string
	var metrics map[string]string
	expected := []string{"Stats2"}
	args := ArgsStatProcessEventWithApiKey{
		APIKey: "12345",
		StatsArgsProcessEvent: engine.StatsArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.AnswerTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:       time.Duration(135 * time.Second),
					utils.COST:        123.0,
					utils.RunID:       utils.DEFAULT_RUNID,
					utils.Destination: "1002"},
			},
		},
	}
	if err := dspStsRPC.Call(utils.StatSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	args2 := TntIDWithApiKey{
		APIKey: "12345",
		TenantID: utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		},
	}
	expectedMetrics := map[string]string{
		utils.MetaTCC: "123",
		utils.MetaTCD: "2m15s",
	}

	if err := dspStsRPC.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}

	args = ArgsStatProcessEventWithApiKey{
		APIKey: "12345",
		StatsArgsProcessEvent: engine.StatsArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account:     "1002",
					utils.AnswerTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:       time.Duration(45 * time.Second),
					utils.RunID:       utils.DEFAULT_RUNID,
					utils.COST:        10.0,
					utils.Destination: "1001",
				},
			},
		},
	}
	if err := dspStsRPC.Call(utils.StatSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expectedMetrics = map[string]string{
		utils.MetaTCC: "133",
		utils.MetaTCD: "3m0s",
	}
	if err := dspStsRPC.Call(utils.StatSv1GetQueueStringMetrics,
		args2, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testDspStsKillEngine(t *testing.T) {
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
}
