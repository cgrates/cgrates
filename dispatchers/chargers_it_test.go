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
	dspCppCfgPath  string
	dspCppCfg      *config.CGRConfig
	dspCppRPC      *rpc.Client
	instCppCfgPath string
	instCppCfg     *config.CGRConfig
	instCppRPC     *rpc.Client
)

var sTestsDspCpp = []func(t *testing.T){
	testDspCppInitCfg,
	testDspCppInitDataDb,
	testDspCppResetStorDb,
	testDspCppStartEngine,
	testDspCppRPCConn,
	testDspCppPing,
	testDspCppLoadData,
	testDspCppAddAttributeWithPermision,
	testDspCppTestAuthKey,
	testDspCppAddAttributesWithPermision2,
	testDspCppTestAuthKey2,
	testDspCppKillEngine,
}

//Test start here
func TestDspChargerS(t *testing.T) {
	for _, stest := range sTestsDspCpp {
		t.Run("", stest)
	}
}

func testDspCppInitCfg(t *testing.T) {
	var err error
	dspCppCfgPath = path.Join(dspDataDir, "conf", "samples", "dispatcher")
	dspCppCfg, err = config.NewCGRConfigFromFolder(dspCppCfgPath)
	if err != nil {
		t.Error(err)
	}
	dspCppCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(dspCppCfg)
	instCppCfgPath = path.Join(dspDataDir, "conf", "samples", "tutmysql")
	instCppCfg, err = config.NewCGRConfigFromFolder(instCppCfgPath)
	if err != nil {
		t.Error(err)
	}
	instCppCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(instCppCfg)
}

func testDspCppInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(instCppCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDspCppResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(instCppCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDspCppStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(instCppCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(dspCppCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDspCppRPCConn(t *testing.T) {
	var err error
	instCppRPC, err = jsonrpc.Dial("tcp", instCppCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	dspCppRPC, err = jsonrpc.Dial("tcp", dspCppCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDspCppPing(t *testing.T) {
	var reply string
	if err := instCppRPC.Call(utils.ChargerSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dspCppRPC.Call(utils.ChargerSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCppLoadData(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(dspDataDir, "tariffplans", "tutorial")}
	if err := instCppRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testDspCppAddAttributeWithPermision(t *testing.T) {
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
				Substitute: config.NewRSRParsersMustCompile("ThresholdSv1.GetThresholdsForEvent", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := instCppRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instCppRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspCppTestAuthKey(t *testing.T) {
	args := CGREvWithApiKey{
		APIKey: "12345",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}
	var reply *engine.ChargerProfiles
	if err := dspCppRPC.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspCppAddAttributesWithPermision2(t *testing.T) {
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
				Substitute: config.NewRSRParsersMustCompile("ThresholdSv1.ProcessEvent&ChargerSv1.GetChargersForEvent", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := instCppRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instCppRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspCppTestAuthKey2(t *testing.T) {
	args := CGREvWithApiKey{
		APIKey: "12345",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}
	eChargers := &engine.ChargerProfiles{
		&engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "DEFAULT",
			FilterIDs:    []string{},
			RunID:        "*default",
			AttributeIDs: []string{"*none"},
			Weight:       0,
		},
	}
	var reply *engine.ChargerProfiles
	if err := dspCppRPC.Call(utils.ChargerSv1GetChargersForEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eChargers, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eChargers), utils.ToJSON(reply))
	}
}

func testDspCppKillEngine(t *testing.T) {
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
}
