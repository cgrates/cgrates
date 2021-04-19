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

package apis

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	alsPrfCfgPath   string
	alsPrfCfg       *config.CGRConfig
	attrSRPC        *birpc.Client
	alsPrf          *engine.AttributeProfileWithAPIOpts
	alsPrfConfigDIR string //run tests for specific configuration

	sTestsAlsPrf = []func(t *testing.T){
		testAttributeSInitCfg,
		testAttributeSInitDataDb,
		testAttributeSResetStorDb,
		testAttributeSStartEngine,
		testAttributeSRPCConn,
		testGetAttributeProfileBeforeSet,
		//testAttributeSLoadFromFolder,
		testAttributeSetAttributeProfile,
		testAttributeSKillEngine,
	}
)

func TestAttributeSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		alsPrfConfigDIR = "attributes_internal"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAlsPrf {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func testAttributeSInitCfg(t *testing.T) {
	var err error
	alsPrfCfgPath = path.Join(*dataDir, "conf", "samples", alsPrfConfigDIR)
	alsPrfCfg, err = config.NewCGRConfigFromPath(alsPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAttributeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

func testAttributeSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(alsPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testAttributeSRPCConn(t *testing.T) {
	var err error
	attrSRPC, err = newRPCClient(alsPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testGetAttributeProfileBeforeSet(t *testing.T) {
	var reply *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := attrSRPC.Call(context.Background(),
		utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAttributeSetAttributeProfile(t *testing.T) {
	time.Sleep(1000 * time.Millisecond)
	attrPrf := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			Contexts:  []string{"*any"},
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
	}
	var reply string
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAttr := &engine.APIAttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_IT_TEST",
		Contexts:  []string{"*any"},
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: "1002",
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: "cgrates.itsyscom",
			},
		},
	}
	var result *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAttr) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedAttr), utils.ToJSON(result))
	}
}

//Kill the engine when it is about to be finished
func testAttributeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
