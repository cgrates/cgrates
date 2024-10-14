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
	loadersIDBIdxCfgDir                        string
	loadersIDBIdxCfgPath                       string
	loadersIDBIdxCfgPathInternal               = path.Join(*utils.DataDir, "conf", "samples", "loaders_indexes_internal_db")
	loadersIDBIdxCfg, loadersIDBIdxCfgInternal *config.CGRConfig
	loadersIDBIdxRPC, loadersIDBIdxRPCInternal *birpc.Client

	LoadersIDBIdxTests = []func(t *testing.T){
		testLoadersIDBIdxItLoadConfig,
		testLoadersIDBIdxItDB,
		testLoadersIDBIdxItStartEngines,
		testLoadersIDBIdxItRPCConn,
		testLoadersIDBIdxItLoad,
		testLoadersIDBIdxCheckAttributes,
		testLoadersIDBIdxCheckAttributesIndexes,
		testLoadersIDBIdxItStopCgrEngine,
	}
)

func TestLoadersIDBIdxIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		loadersIDBIdxCfgDir = "tutinternal"
	case utils.MetaMySQL:
		loadersIDBIdxCfgDir = "tutmysql"
	case utils.MetaMongo:
		loadersIDBIdxCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range LoadersIDBIdxTests {
		t.Run(loadersIDBIdxCfgDir, stest)
	}
}

func testLoadersIDBIdxItLoadConfig(t *testing.T) {
	var err error
	loadersIDBIdxCfgPath = path.Join(*utils.DataDir, "conf", "samples", loadersIDBIdxCfgDir)
	if loadersIDBIdxCfg, err = config.NewCGRConfigFromPath(loadersIDBIdxCfgPath); err != nil {
		t.Error(err)
	}
	if loadersIDBIdxCfgInternal, err = config.NewCGRConfigFromPath(loadersIDBIdxCfgPathInternal); err != nil {
		t.Error(err)
	}
}

func testLoadersIDBIdxItDB(t *testing.T) {
	if err := engine.InitDataDb(loadersIDBIdxCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(loadersIDBIdxCfg); err != nil {
		t.Fatal(err)
	}
}

func testLoadersIDBIdxItStartEngines(t *testing.T) {
	if _, err := engine.StopStartEngine(loadersIDBIdxCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(loadersIDBIdxCfgPathInternal, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testLoadersIDBIdxItRPCConn(t *testing.T) {
	loadersIDBIdxRPC = engine.NewRPCClient(t, loadersIDBIdxCfg.ListenCfg())
	loadersIDBIdxRPCInternal = engine.NewRPCClient(t, loadersIDBIdxCfgInternal.ListenCfg())
}

func testLoadersIDBIdxItLoad(t *testing.T) {
	var loadInst utils.LoadInstance
	if err := loadersIDBIdxRPCInternal.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")},
		&loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testLoadersIDBIdxCheckAttributes(t *testing.T) {
	exp := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1001_SIMPLEAUTH",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Password",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
			},
		},
		Weight: 20.0,
	}

	var reply *engine.AttributeProfile
	if err := loadersIDBIdxRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply.Compile(); !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testLoadersIDBIdxCheckAttributesIndexes(t *testing.T) {
	expIdx := []string{
		"*string:*req.Account:1001:ATTR_1001_SIMPLEAUTH",
		"*string:*req.Account:1002:ATTR_1002_SIMPLEAUTH",
		"*string:*req.Account:1003:ATTR_1003_SIMPLEAUTH",
	}
	var indexes []string
	if err := loadersIDBIdxRPC.Call(context.Background(), utils.APIerSv1GetFilterIndexes, &v1.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaString,
		Context: "simpleauth"},
		&indexes); err != nil {
		t.Error(err)
	} else if sort.Strings(indexes); !reflect.DeepEqual(indexes, expIdx) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(expIdx), utils.ToJSON(indexes))
	}
}

func testLoadersIDBIdxItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
