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
	dspThCfgPath  string
	dspThCfg      *config.CGRConfig
	dspThRPC      *rpc.Client
	instThCfgPath string
	instThCfg     *config.CGRConfig
	instThRPC     *rpc.Client
)

var sTestsDspTh = []func(t *testing.T){
	testDspThInitCfg,
	testDspThInitDataDb,
	testDspThResetStorDb,
	testDspThStartEngine,
	testDspThRPCConn,
	testDspThPing,
	testDspThLoadData,
	testDspThAddAttributesWithPermision,
	testDspThTestAuthKey,
	testDspThAddAttributesWithPermision2,
	testDspThTestAuthKey2,
	testDspThKillEngine,
}

//Test start here
func TestDspThresholdS(t *testing.T) {
	for _, stest := range sTestsDspTh {
		t.Run("", stest)
	}
}

func testDspThInitCfg(t *testing.T) {
	var err error
	dspThCfgPath = path.Join(dspDataDir, "conf", "samples", "dispatcher")
	dspThCfg, err = config.NewCGRConfigFromFolder(dspThCfgPath)
	if err != nil {
		t.Error(err)
	}
	dspThCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(dspThCfg)
	instThCfgPath = path.Join(dspDataDir, "conf", "samples", "tutmysql")
	instThCfg, err = config.NewCGRConfigFromFolder(instThCfgPath)
	if err != nil {
		t.Error(err)
	}
	instThCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(instThCfg)
}

func testDspThInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(instThCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDspThResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(instThCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDspThStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(instThCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(dspThCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDspThRPCConn(t *testing.T) {
	var err error
	instThRPC, err = jsonrpc.Dial("tcp", instThCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	dspThRPC, err = jsonrpc.Dial("tcp", dspThCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDspThPing(t *testing.T) {
	var reply string
	if err := instThRPC.Call(utils.ThresholdSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dspThRPC.Call(utils.ThresholdSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspThLoadData(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(dspDataDir, "tariffplans", "tutorial")}
	if err := instThRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testDspThAddAttributesWithPermision(t *testing.T) {
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
	if err := instThRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instThRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspThTestAuthKey(t *testing.T) {
	var ids []string
	nowTime := time.Now()
	args := &ArgsProcessEventWithApiKey{
		APIKey: "12345",
		ArgsProcessEvent: engine.ArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Time:   &nowTime,
				Event: map[string]interface{}{
					utils.Account: "1002"},
			},
		},
	}

	if err := dspThRPC.Call(utils.ThresholdSv1ProcessEvent,
		args, &ids); err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
	var th *engine.Thresholds
	eTh := &engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
			Hits:   0,
		},
	}
	if err := dspThRPC.Call(utils.ThresholdSv1GetThresholdsForEvent, args, &th); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTh, th) {
		t.Errorf("expecting: %+v, received: %+v", eTh, th)
	}
}

func testDspThAddAttributesWithPermision2(t *testing.T) {
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
				Substitute: config.NewRSRParsersMustCompile("ThresholdSv1.ProcessEvent&ThresholdSv1.GetThresholdsForEvent", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := instThRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instThRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspThTestAuthKey2(t *testing.T) {
	var ids []string
	eIDs := []string{"THD_ACNT_1002"}
	nowTime := time.Now()
	args := &ArgsProcessEventWithApiKey{
		APIKey: "12345",
		ArgsProcessEvent: engine.ArgsProcessEvent{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Time:   &nowTime,
				Event: map[string]interface{}{
					utils.Account: "1002"},
			},
		},
	}

	if err := dspThRPC.Call(utils.ThresholdSv1ProcessEvent, args, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("expecting: %+v, received: %+v", eIDs, ids)
	}
	var th *engine.Thresholds
	eTh := &engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
			Hits:   1,
		},
	}
	if err := dspThRPC.Call(utils.ThresholdSv1GetThresholdsForEvent, args, &th); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual((*eTh)[0].Tenant, (*th)[0].Tenant) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].Tenant, (*th)[0].Tenant)
	} else if !reflect.DeepEqual((*eTh)[0].ID, (*th)[0].ID) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].ID, (*th)[0].ID)
	} else if !reflect.DeepEqual((*eTh)[0].Hits, (*th)[0].Hits) {
		t.Errorf("expecting: %+v, received: %+v", (*eTh)[0].Hits, (*th)[0].Hits)
	}
}

func testDspThKillEngine(t *testing.T) {
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
}
