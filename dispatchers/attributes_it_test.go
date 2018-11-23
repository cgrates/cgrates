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
	dspAttrCfgPath  string
	dspAttrCfg      *config.CGRConfig
	dspAttrRPC      *rpc.Client
	instAttrCfgPath string
	instAttrCfg     *config.CGRConfig
	instAttrRPC     *rpc.Client
)

var sTestsDspAttr = []func(t *testing.T){
	testDspAttrInitCfg,
	testDspAttrInitDataDb,
	testDspAttrResetStorDb,
	testDspAttrStartEngine,
	testDspAttrRPCConn,
	testDspAttrPing,
	testDspAttrLoadData,
	testDspAttrAddAttributesWithPermision,
	testDspAttrTestMissingApiKey,
	testDspAttrTestUnknownApiKey,
	testDspAttrTestAuthKey,
	testDspAttrAddAttributesWithPermision2,
	testDspAttrTestAuthKey2,
	testDspAttrKillEngine,
}

//Test start here
func TestDspAttributeS(t *testing.T) {
	for _, stest := range sTestsDspAttr {
		t.Run("", stest)
	}
}

func testDspAttrInitCfg(t *testing.T) {
	var err error
	dspAttrCfgPath = path.Join(dspDataDir, "conf", "samples", "dispatcher")
	dspAttrCfg, err = config.NewCGRConfigFromFolder(dspAttrCfgPath)
	if err != nil {
		t.Error(err)
	}
	dspAttrCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(dspAttrCfg)
	instAttrCfgPath = path.Join(dspDataDir, "conf", "samples", "tutmysql")
	instAttrCfg, err = config.NewCGRConfigFromFolder(instAttrCfgPath)
	if err != nil {
		t.Error(err)
	}
	instAttrCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(instAttrCfg)
}

func testDspAttrInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(instAttrCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDspAttrResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(instAttrCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDspAttrStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(instAttrCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(dspAttrCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDspAttrRPCConn(t *testing.T) {
	var err error
	instAttrRPC, err = jsonrpc.Dial("tcp", instAttrCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	dspAttrRPC, err = jsonrpc.Dial("tcp", dspAttrCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}

}

func testDspAttrPing(t *testing.T) {
	var reply string
	if err := instAttrRPC.Call(utils.AttributeSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dspAttrRPC.Call(utils.AttributeSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspAttrLoadData(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(dspDataDir, "tariffplans", "tutorial")}
	if err := instAttrRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testDspAttrAddAttributesWithPermision(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AuthKey",
		Contexts:  []string{utils.MetaAuth},
		FilterIDs: []string{"*string:APIKey:12345"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FieldName:  utils.APIMethods,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("ThresholdSv1.GetThAttrholdsForEvent", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	alsPrf.Compile()
	var Attrult string
	if err := instAttrRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &Attrult); err != nil {
		t.Error(err)
	} else if Attrult != utils.OK {
		t.Error("Unexpected reply returned", Attrult)
	}
	var reply *engine.AttributeProfile
	if err := instAttrRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
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
	if err := dspAttrRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err.Error() != utils.NewErrMandatoryIeMissing(utils.APIKey).Error() {
		t.Error(err)
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
	if err := dspAttrRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err.Error() != utils.ErrUnknownApiKey.Error() {
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
	if err := dspAttrRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspAttrAddAttributesWithPermision2(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AuthKey",
		Contexts:  []string{utils.MetaAuth},
		FilterIDs: []string{"*string:APIKey:12345"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FieldName:  utils.APIMethods,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("AttributeSv1.GetAttributeForEvent&AttributeSv1.ProcessEvent", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	alsPrf.Compile()
	if err := instAttrRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := instAttrRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspAttrTestAuthKey2(t *testing.T) {
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
	if err := dspAttrRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err != nil {
		t.Error(err)
	}
	attrReply.Compile()
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
	if err := dspAttrRPC.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testDspAttrKillEngine(t *testing.T) {
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
}
