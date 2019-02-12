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
	"os/exec"
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	attrEngine *testDispatcher
	dispEngine *testDispatcher
	allEngine  *testDispatcher
	allEngine2 *testDispatcher
)

var sTestsDspAttr = []func(t *testing.T){
	testDspAttrPingFailover,
	testDspAttrGetAttrFailover,

	testDspAttrPing,
	testDspAttrTestMissingApiKey,
	testDspAttrTestUnknownApiKey,
	testDspAttrTestAuthKey,
	testDspAttrTestAuthKey2,
	testDspAttrTestAuthKey3,
}

type testDispatcher struct {
	CfgParh string
	Cfg     *config.CGRConfig
	RCP     *rpc.Client
	cmd     *exec.Cmd
}

func newTestEngine(t *testing.T, cfgPath string, initDataDB, intitStoreDB bool) (d *testDispatcher) {
	d = new(testDispatcher)
	d.CfgParh = cfgPath
	var err error
	d.Cfg, err = config.NewCGRConfigFromFolder(d.CfgParh)
	if err != nil {
		t.Fatalf("Error at config init :%v\n", err)
	}
	d.Cfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()

	if initDataDB {
		d.initDataDb(t)
	}

	if intitStoreDB {
		d.resetStorDb(t)
	}
	d.startEngine(t)
	return d
}

func (d *testDispatcher) startEngine(t *testing.T) {
	var err error
	if d.cmd, err = engine.StartEngine(d.CfgParh, dspDelay); err != nil {
		t.Fatalf("Error at engine start:%v\n", err)
	}

	if d.RCP, err = jsonrpc.Dial("tcp", d.Cfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatalf("Error at dialing rcp client:%v\n", err)
	}
}

func (d *testDispatcher) stopEngine(t *testing.T) {
	pid := strconv.Itoa(d.cmd.Process.Pid)
	if err := exec.Command("kill", "-9", pid).Run(); err != nil {
		t.Fatalf("Error at stop engine:%v\n", err)
	}
	// // if err := d.cmd.Process.Kill(); err != nil {
	// // 	t.Fatalf("Error at stop engine:%v\n", err)
	// }
}

func (d *testDispatcher) initDataDb(t *testing.T) {
	if err := engine.InitDataDb(d.Cfg); err != nil {
		t.Fatalf("Error at DataDB init:%v\n", err)
	}
}

// Wipe out the cdr database
func (d *testDispatcher) resetStorDb(t *testing.T) {
	if err := engine.InitStorDb(d.Cfg); err != nil {
		t.Fatalf("Error at DataDB init:%v\n", err)
	}
}
func (d *testDispatcher) loadData(t *testing.T, path string) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path}
	if err := d.RCP.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Errorf("Error at loading data from folder:%v", err)
	}
}

//Test start here
func TestDspAttributeS(t *testing.T) {
	allEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all"), true, true)
	allEngine2 = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "all2"), true, true)
	attrEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "attributes"), true, true)
	dispEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", "dispatchers"), true, true)
	allEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "tutorial"))
	allEngine2.loadData(t, path.Join(dspDataDir, "tariffplans", "oldtutorial"))
	attrEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "dispatchers"))
	time.Sleep(500 * time.Millisecond)
	for _, stest := range sTestsDspAttr {
		t.Run("TestDspAttributeS", stest)
	}
	attrEngine.stopEngine(t)
	dispEngine.stopEngine(t)
	allEngine.stopEngine(t)
	allEngine2.stopEngine(t)
}

func testDspAttrPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.AttributeSv1Ping, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = ""
	if err := allEngine2.RCP.Call(utils.AttributeSv1Ping, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = ""
	if err := dispEngine.RCP.Call(utils.AttributeSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		APIKey: "attr12345",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	reply = ""
	if err := dispEngine.RCP.Call(utils.AttributeSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			ID:     "PING",
			Tenant: "cgrates.org",
		},
		APIKey: "attr12345",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	reply = ""
	if err := dispEngine.RCP.Call(utils.AttributeSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			ID:     "PING",
			Tenant: "cgrates.org",
		},
		APIKey: "attr12345",
	}, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspAttrGetAttrFailover(t *testing.T) {
	args := &CGREvWithApiKey{
		APIKey: "attr12345",
		CGREvent: utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSGetAttributeForEvent",
			Context: utils.StringPointer("simpleauth"),
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.EVENT_NAME: "Event1",
			},
		},
	}
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    args.Tenant,
		ID:        "ATTR_1002_SIMPLEAUTH",
		FilterIDs: []string{"*string:Account:1002"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FieldName:  "Password",
				Initial:    utils.ANY,
				Substitute: config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20.0,
	}
	eAttrPrf.Compile()

	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1002_SIMPLEAUTH"},
		AlteredFields:   []string{"Password"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSGetAttributeForEvent",
			Context: utils.StringPointer("simpleauth"),
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.EVENT_NAME: "Event1",
				"Password":       "CGRateS.org",
			},
		},
	}

	var attrReply *engine.AttributeProfile
	var rplyEv engine.AttrSProcessEventReply
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := dispEngine.RCP.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	} else if reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}

	allEngine2.stopEngine(t)

	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err != nil {
		t.Error(err)
	}
	if attrReply != nil {
		attrReply.Compile()
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	if err := dispEngine.RCP.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}

	allEngine2.startEngine(t)
}

func testDspAttrPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.AttributeSv1Ping, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if dispEngine.RCP == nil {
		t.Fatal(dispEngine.RCP)
	}
	if err := dispEngine.RCP.Call(utils.AttributeSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		APIKey: "attr12345",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspAttrTestMissingApiKey(t *testing.T) {
	args := &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSGetAttributeForEvent",
			Context: utils.StringPointer("simpleauth"),
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.NewErrMandatoryIeMissing(utils.APIKey).Error() {
		t.Errorf("Error:%v rply=%s", err, utils.ToJSON(attrReply))
	}
}

func testDspAttrTestUnknownApiKey(t *testing.T) {
	args := &CGREvWithApiKey{
		APIKey: "1234",
		CGREvent: utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSGetAttributeForEvent",
			Context: utils.StringPointer("simpleauth"),
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrUnknownApiKey.Error() {
		t.Error(err)
	}
}

func testDspAttrTestAuthKey(t *testing.T) {
	args := &CGREvWithApiKey{
		APIKey: "12345",
		CGREvent: utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSGetAttributeForEvent",
			Context: utils.StringPointer("simpleauth"),
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspAttrTestAuthKey2(t *testing.T) {
	args := &CGREvWithApiKey{
		APIKey: "attr12345",
		CGREvent: utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSGetAttributeForEvent",
			Context: utils.StringPointer("simpleauth"),
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    args.Tenant,
		ID:        "ATTR_1001_SIMPLEAUTH",
		FilterIDs: []string{"*string:Account:1001"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FieldName:  "Password",
				Initial:    utils.ANY,
				Substitute: config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20.0,
	}
	eAttrPrf.Compile()
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err != nil {
		t.Error(err)
	}
	if attrReply != nil {
		attrReply.Compile()
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1001_SIMPLEAUTH"},
		AlteredFields:   []string{"Password"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSGetAttributeForEvent",
			Context: utils.StringPointer("simpleauth"),
			Event: map[string]interface{}{
				utils.Account: "1001",
				"Password":    "CGRateS.org",
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := dispEngine.RCP.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testDspAttrTestAuthKey3(t *testing.T) {
	args := &CGREvWithApiKey{
		APIKey: "attr12345",
		CGREvent: utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSGetAttributeForEvent",
			Context: utils.StringPointer("simpleauth"),
			Event: map[string]interface{}{
				utils.Account:    "1001",
				utils.EVENT_NAME: "Event1",
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RCP.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}
