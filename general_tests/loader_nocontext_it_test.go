//go:build integration
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

package general_tests

import (
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	ldrCtxCfgPath string
	ldrCtxCfg     *config.CGRConfig
	ldrCtxRPC     *birpc.Client
	ldrCtxConfDIR string //run tests for specific configuration
	ldrCtxDelay   int

	sTestsLdrCtx = []func(t *testing.T){
		testLoaderNoContextRemoveFolders,
		testLoaderNoContextCreateFolders,

		testLoaderNoContextLoadConfig,
		testLoaderNoContextInitDataDb,
		testLoaderNoContextResetStorDb,
		testLoaderNoContextStartEngine,
		testLoaderNoContextRpcConn,

		testLoaderNoContextWriteCSVs,
		testLoaderNoContextLoadTariffPlans,
		testLoaderNoContextGetFilterIndexesAfterLoad,
		testLoaderNoContextSetProfiles,
		testLoaderNoContextGetFilterIndexesAfterSet,

		testLoaderNoContextStopEngine,
		testLoaderNoContextRemoveFolders,
	}
)

func TestLoaderNoContextIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		ldrCtxConfDIR = "tutinternal"
	case utils.MetaMySQL:
		ldrCtxConfDIR = "tutmysql"
	case utils.MetaMongo:
		ldrCtxConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsLdrCtx {
		t.Run(ldrCtxConfDIR, stest)
	}
}

func testLoaderNoContextLoadConfig(t *testing.T) {
	var err error
	ldrCtxCfgPath = path.Join(*utils.DataDir, "conf", "samples", ldrCtxConfDIR)
	if ldrCtxCfg, err = config.NewCGRConfigFromPath(ldrCtxCfgPath); err != nil {
		t.Error(err)
	}
	ldrCtxDelay = 1000
}

func testLoaderNoContextInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(ldrCtxCfg); err != nil {
		t.Fatal(err)
	}
}

func testLoaderNoContextResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(ldrCtxCfg); err != nil {
		t.Fatal(err)
	}
}

func testLoaderNoContextStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ldrCtxCfgPath, ldrCtxDelay); err != nil {
		t.Fatal(err)
	}
}

func testLoaderNoContextRpcConn(t *testing.T) {
	ldrCtxRPC = engine.NewRPCClient(t, ldrCtxCfg.ListenCfg())
}

func testLoaderNoContextWriteCSVs(t *testing.T) {
	writeFile := func(fileName, data string) error {
		csvFile, err := os.Create(path.Join("/tmp/TestLoaderNoContextIT", fileName))
		if err != nil {
			return err
		}
		defer csvFile.Close()
		_, err = csvFile.WriteString(data)
		if err != nil {
			return err

		}
		return csvFile.Sync()
	}

	// Create and populate Attributes.csv
	if err := writeFile(utils.AttributesCsv, `
#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_1,,*string:~*req.Field1:Value1,,,*req.Field1,*constant,Value2,false,10
cgrates.org,ATTR_2,,,,,*req.Field2,*constant,Value2,false,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate DispatcherProfiles.csv
	if err := writeFile(utils.DispatcherProfilesCsv, `
#Tenant,ID,Subsystems,FilterIDs,ActivationInterval,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters,Weight
cgrates.org,DSP1,,,,*weight,,ALL,,20,false,,10
cgrates.org,DSP1,,,,,,ALL2,,10,,,
cgrates.org,DSP2,,*string:~*req.Field1:Value1,,*weight,,connID,,20,false,,20
cgrates.org,DSP2,,,,,,ALL2,,10,,,
`); err != nil {
		t.Fatal(err)
	}
}

func testLoaderNoContextLoadTariffPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: "/tmp/TestLoaderNoContextIT"}
	if err := ldrCtxRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func testLoaderNoContextGetFilterIndexesAfterLoad(t *testing.T) {
	// check attribute profile filter indexes
	expIdx := []string{
		"*none:*any:*any:ATTR_2",
		"*string:*req.Field1:Value1:ATTR_1",
	}
	var result []string
	if err := ldrCtxRPC.Call(context.Background(), utils.APIerSv1GetFilterIndexes, &v1.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes,
		Tenant:   "cgrates.org",
		Context:  utils.MetaAny,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("expected: %+v,\nreceived: %+v", expIdx, result)
		}
	}

	// check dispatcher profile filter indexes
	expIdx = []string{
		"*none:*any:*any:DSP1",
		"*string:*req.Field1:Value1:DSP2",
	}
	if err := ldrCtxRPC.Call(context.Background(), utils.APIerSv1GetFilterIndexes, &v1.AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   "cgrates.org",
		Context:  utils.MetaAny,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("expected: %+v,\nreceived: %+v", expIdx, result)
		}
	}
}

func testLoaderNoContextSetProfiles(t *testing.T) {
	// set attribute profile
	attrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "cgrates.org",
			ID:     "ATTR_3",
			// Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.Field3:Value3"},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field3",
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile("Value4", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
	}
	attrPrf.Compile()
	var reply string
	if err := ldrCtxRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var attrReply *engine.AttributeProfile
	if err := ldrCtxRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_3"}}, &attrReply); err != nil {
		t.Error(err)
	} else {
		attrReply.Compile()
		attrPrf.AttributeProfile.Contexts = []string{utils.MetaAny}
		if !reflect.DeepEqual(attrPrf.AttributeProfile, attrReply) {
			t.Errorf("expected : %+v,\nreceived: %+v",
				utils.ToJSON(attrPrf.AttributeProfile), utils.ToJSON(attrReply))
		}
	}

	// set dispatcher profile
	dspPrf := &v1.DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "cgrates.org",
			// Subsystems: []string{utils.MetaAny},
			ID:        "DSP3",
			FilterIDs: []string{"*string:~*req.RandomField:RandomValue"},
			Strategy:  utils.MetaFirst,
			Weight:    20,
		},
	}

	if err := ldrCtxRPC.Call(context.Background(), utils.APIerSv1SetDispatcherProfile,
		dspPrf,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var dspReply *engine.DispatcherProfile
	if err := ldrCtxRPC.Call(context.Background(), utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DSP3"},
		&dspReply); err != nil {
		t.Error(err)
	} else {
		dspPrf.DispatcherProfile.Subsystems = []string{utils.MetaAny}
		if !reflect.DeepEqual(dspPrf.DispatcherProfile, dspReply) {
			t.Errorf("expected: %+v,\nreceived: %+v", utils.ToJSON(dspPrf.DispatcherProfile), utils.ToJSON(dspReply))
		}
	}
}

func testLoaderNoContextGetFilterIndexesAfterSet(t *testing.T) {
	// check attribute profile filter indexes
	expIdx := []string{
		"*none:*any:*any:ATTR_2",
		"*string:*req.Field1:Value1:ATTR_1",
		"*string:*req.Field3:Value3:ATTR_3",
	}
	var result []string
	if err := ldrCtxRPC.Call(context.Background(), utils.APIerSv1GetFilterIndexes, &v1.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes,
		Tenant:   "cgrates.org",
		Context:  utils.MetaAny,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("expected: %+v,\nreceived: %+v", expIdx, result)
		}
	}

	// check dispatcher profile filter indexes
	expIdx = []string{
		"*none:*any:*any:DSP1",
		"*string:*req.Field1:Value1:DSP2",
		"*string:*req.RandomField:RandomValue:DSP3",
	}
	if err := ldrCtxRPC.Call(context.Background(), utils.APIerSv1GetFilterIndexes, &v1.AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   "cgrates.org",
		Context:  utils.MetaAny,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("expected: %+v,\nreceived: %+v", expIdx, result)
		}
	}
}

func testLoaderNoContextStopEngine(t *testing.T) {
	if err := engine.KillEngine(ldrCtxDelay); err != nil {
		t.Error(err)
	}
}

func testLoaderNoContextCreateFolders(t *testing.T) {
	if err := os.MkdirAll("/tmp/TestLoaderNoContextIT", 0755); err != nil {
		t.Error(err)
	}
}

func testLoaderNoContextRemoveFolders(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestLoaderNoContextIT"); err != nil {
		t.Error(err)
	}
}
