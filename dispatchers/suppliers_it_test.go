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
	dspSupCfgPath  string
	dspSupCfg      *config.CGRConfig
	dspSupRPC      *rpc.Client
	instSupCfgPath string
	instSupCfg     *config.CGRConfig
	instSupRPC     *rpc.Client
)

var sTestsDspSup = []func(t *testing.T){
	testDspSupInitCfg,
	testDspSupInitDataDb,
	testDspSupResetStorDb,
	testDspSupStartEngine,
	testDspSupRPCConn,
	testDspSupPing,
	testDspSupLoadData,
	testDspSupAddAttributesWithPermision,
	testDspSupTestAuthKey,
	testDspSupAddAttributesWithPermision2,
	testDspSupTestAuthKey2,
	testDspSupKillEngine,
}

//Test start here
func TestDspSupplierS(t *testing.T) {
	for _, stest := range sTestsDspSup {
		t.Run("", stest)
	}
}

func testDspSupInitCfg(t *testing.T) {
	var err error
	dspSupCfgPath = path.Join(dspDataDir, "conf", "samples", "dispatcher")
	dspSupCfg, err = config.NewCGRConfigFromFolder(dspSupCfgPath)
	if err != nil {
		t.Error(err)
	}
	dspSupCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(dspSupCfg)
	instSupCfgPath = path.Join(dspDataDir, "conf", "samples", "tutmysql")
	instSupCfg, err = config.NewCGRConfigFromFolder(instSupCfgPath)
	if err != nil {
		t.Error(err)
	}
	instSupCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(instSupCfg)
}

func testDspSupInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(instSupCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDspSupResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(instSupCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDspSupStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(instSupCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(dspSupCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDspSupRPCConn(t *testing.T) {
	var err error
	instSupRPC, err = jsonrpc.Dial("tcp", instSupCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	dspSupRPC, err = jsonrpc.Dial("tcp", dspSupCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}

}

func testDspSupPing(t *testing.T) {
	var reply string
	if err := instSupRPC.Call(utils.SupplierSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dspSupRPC.Call(utils.SupplierSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspSupLoadData(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(dspDataDir, "tariffplans", "tutorial")}
	if err := instSupRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testDspSupAddAttributesWithPermision(t *testing.T) {
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
	if err := instSupRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instSupRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspSupTestAuthKey(t *testing.T) {
	var rpl *engine.SortedSuppliers
	args := &ArgsGetSuppliersWithApiKey{
		APIKey: "12345",
		ArgsGetSuppliers: engine.ArgsGetSuppliers{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Time:   &nowTime,
				Event: map[string]interface{}{
					utils.Account:     "1002",
					utils.Subject:     "1002",
					utils.Destination: "1001",
					utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
					utils.Usage:       "1m20s",
				},
			},
		},
	}
	if err := dspSupRPC.Call(utils.SupplierSv1GetSuppliers,
		args, &rpl); err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspSupAddAttributesWithPermision2(t *testing.T) {
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
				Substitute: config.NewRSRParsersMustCompile("ThresholdSv1.ProcessEvent&SupplierSv1.GetSuppliers", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := instSupRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instSupRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspSupTestAuthKey2(t *testing.T) {
	var rpl *engine.SortedSuppliers
	eRpl := &engine.SortedSuppliers{
		ProfileID: "SPL_ACNT_1002",
		Sorting:   utils.MetaLeastCost,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID:         "supplier1",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Cost:         0.1166,
					utils.RatingPlanID: "RP_1002_LOW",
					utils.Weight:       10.0,
				},
			},
			{
				SupplierID:         "supplier2",
				SupplierParameters: "",
				SortingData: map[string]interface{}{
					utils.Cost:         0.2334,
					utils.RatingPlanID: "RP_1002",
					utils.Weight:       20.0,
				},
			},
		},
	}
	args := &ArgsGetSuppliersWithApiKey{
		APIKey: "12345",
		ArgsGetSuppliers: engine.ArgsGetSuppliers{
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Time:   &nowTime,
				Event: map[string]interface{}{
					utils.Account:     "1002",
					utils.Subject:     "1002",
					utils.Destination: "1001",
					utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
					utils.Usage:       "1m20s",
				},
			},
		},
	}
	if err := dspSupRPC.Call(utils.SupplierSv1GetSuppliers,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl), utils.ToJSON(rpl))
	}
}

func testDspSupKillEngine(t *testing.T) {
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
}
