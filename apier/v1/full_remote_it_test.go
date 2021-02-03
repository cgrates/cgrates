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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

var (
	fullRemInternalCfgPath    string
	fullRemInternalCfgDirPath string
	fullRemInternalCfg        *config.CGRConfig
	fullRemInternalRPC        *rpc.Client

	fullRemEngineOneCfgPath    string
	fullRemEngineOneCfgDirPath string
	fullRemEngineOneCfg        *config.CGRConfig
	fullRemEngineOneRPC        *rpc.Client

	sTestsFullRemoteIT = []func(t *testing.T){
		testFullRemoteITInitCfg,
		testFullRemoteITDataFlush,
		testFullRemoteITStartEngine,
		testFullRemoteITRPCConn,

		testFullRemoteITAttribute,

		testFullRemoteITKillEngine,
	}
)

func TestFullRemoteIT(t *testing.T) {
	fullRemInternalCfgDirPath = "internal"
	fullRemEngineOneCfgDirPath = "remote"

	for _, stest := range sTestsFullRemoteIT {
		t.Run(*dbType, stest)
	}
}

func testFullRemoteITInitCfg(t *testing.T) {
	var err error
	fullRemInternalCfgPath = path.Join(*dataDir, "conf", "samples", "full_remote", fullRemInternalCfgDirPath)
	fullRemInternalCfg, err = config.NewCGRConfigFromPath(fullRemInternalCfgPath)
	if err != nil {
		t.Error(err)
	}

	// prepare config for engine1
	fullRemEngineOneCfgPath = path.Join(*dataDir, "conf", "samples",
		"full_remote", fullRemEngineOneCfgDirPath)
	fullRemEngineOneCfg, err = config.NewCGRConfigFromPath(fullRemEngineOneCfgPath)
	if err != nil {
		t.Error(err)
	}
	fullRemEngineOneCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()

}

func testFullRemoteITDataFlush(t *testing.T) {
	if err := engine.InitDataDb(fullRemEngineOneCfg); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testFullRemoteITStartEngine(t *testing.T) {
	engine.KillEngine(100)
	if _, err := engine.StartEngine(fullRemInternalCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(fullRemEngineOneCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testFullRemoteITRPCConn(t *testing.T) {
	var err error
	fullRemInternalRPC, err = newRPCClient(fullRemInternalCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	fullRemEngineOneRPC, err = newRPCClient(fullRemEngineOneCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testFullRemoteITAttribute(t *testing.T) {
	// verify for not found in internal
	var reply *engine.AttributeProfile
	if err := fullRemInternalRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}

	var replySet string
	alsPrf := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_1001_SIMPLEAUTH",
			Contexts:  []string{"simpleauth"},
			FilterIDs: []string{"*string:~*req.Account:1001"},

			Attributes: []*engine.Attribute{
				{
					Path:      utils.MetaReq + utils.NestingSep + "Password",
					FilterIDs: []string{},
					Type:      utils.MetaConstant,
					Value:     config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	// add an attribute profile in engine1 and verify it internal
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}},
		&reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(reply))
	}
	// update the attribute profile and verify it to be updated
	alsPrf.FilterIDs = []string{"*string:~*req.Account:1001", "*string:~*req.Destination:1002"}
	alsPrf.Compile()
	if err := fullRemEngineOneRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &replySet); err != nil {
		t.Error(err)
	} else if replySet != utils.OK {
		t.Error("Unexpected reply returned", replySet)
	}

	if err := fullRemInternalRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}},
		&reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(reply))
	}
}

func testFullRemoteITKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
