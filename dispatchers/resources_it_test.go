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
	dspResCfgPath  string
	dspResCfg      *config.CGRConfig
	dspResRPC      *rpc.Client
	instResCfgPath string
	instResCfg     *config.CGRConfig
	instResRPC     *rpc.Client
)

var sTestsDspRes = []func(t *testing.T){
	testDspResInitCfg,
	testDspResInitDataDb,
	testDspResResetStorDb,
	testDspResStartEngine,
	testDspResRPCConn,
	testDspResPing,
	testDspResLoadData,
	testDspResAddAttributesWithPermision,
	testDspResTestAuthKey,
	testDspResAddAttributesWithPermision2,
	testDspResTestAuthKey2,
	testDspResKillEngine,
}

//Test start here
func TestDspResourceS(t *testing.T) {
	for _, stest := range sTestsDspRes {
		t.Run("", stest)
	}
}

func testDspResInitCfg(t *testing.T) {
	var err error
	dspResCfgPath = path.Join(dspDataDir, "conf", "samples", "dispatcher")
	dspResCfg, err = config.NewCGRConfigFromFolder(dspResCfgPath)
	if err != nil {
		t.Error(err)
	}
	dspResCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(dspResCfg)
	instResCfgPath = path.Join(dspDataDir, "conf", "samples", "tutmysql")
	instResCfg, err = config.NewCGRConfigFromFolder(instResCfgPath)
	if err != nil {
		t.Error(err)
	}
	instResCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(instResCfg)
}

func testDspResInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(instResCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDspResResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(instResCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDspResStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(instResCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(dspResCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDspResRPCConn(t *testing.T) {
	var err error
	instResRPC, err = jsonrpc.Dial("tcp", instResCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	dspResRPC, err = jsonrpc.Dial("tcp", dspResCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}

}

func testDspResPing(t *testing.T) {
	var reply string
	if err := instResRPC.Call(utils.ResourceSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dspResRPC.Call(utils.ResourceSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspResLoadData(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(dspDataDir, "tariffplans", "tutorial")}
	if err := instResRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testDspResAddAttributesWithPermision(t *testing.T) {
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
	if err := instResRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instResRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspResTestAuthKey(t *testing.T) {
	var rs *engine.Resources
	args := &ArgsV1ResUsageWithApiKey{
		APIKey: "12345",
		ArgRSv1ResourceUsage: utils.ArgRSv1ResourceUsage{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Destination: "1002",
				},
			},
		},
	}

	if err := dspResRPC.Call(utils.ResourceSv1GetResourcesForEvent,
		args, &rs); err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspResAddAttributesWithPermision2(t *testing.T) {
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
				Substitute: config.NewRSRParsersMustCompile("ThresholdSv1.ProcessEvent&ResourceSv1.GetResourcesForEvent", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := instResRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instResRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspResTestAuthKey2(t *testing.T) {
	var rs *engine.Resources
	args := &ArgsV1ResUsageWithApiKey{
		APIKey: "12345",
		ArgRSv1ResourceUsage: utils.ArgRSv1ResourceUsage{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Destination: "1002",
				},
			},
		},
	}
	eRs := &engine.Resources{
		&engine.Resource{
			Tenant: "cgrates.org",
			ID:     "ResGroup1",
			Usages: map[string]*engine.ResourceUsage{},
		},
	}

	if err := dspResRPC.Call(utils.ResourceSv1GetResourcesForEvent,
		args, &rs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRs, rs) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRs), utils.ToJSON(rs))
	}
}

func testDspResKillEngine(t *testing.T) {
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
}
