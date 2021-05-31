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
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accPrfCfgPath   string
	accPrfCfg       *config.CGRConfig
	accSRPC         *birpc.Client
	accPrfConfigDIR string //run tests for specific configuration

	sTestsAccPrf = []func(t *testing.T){
		testAccSInitCfg,
		testAccSInitDataDb,
		testAccSResetStorDb,
		testAccSStartEngine,
		testAccSRPCConn,
		testGetAccProfileBeforeSet,
		testAccSKillEngine,
	}
)

func TestAccSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		accPrfConfigDIR = "rates_internal"
	case utils.MetaMongo:
		accPrfConfigDIR = "rates_mongo"
	case utils.MetaMySQL:
		accPrfConfigDIR = "rates_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAccPrf {
		t.Run(accPrfConfigDIR, stest)
	}
}

func testAccSInitCfg(t *testing.T) {
	var err error
	accPrfCfgPath = path.Join(*dataDir, "conf", "samples", accPrfConfigDIR)
	accPrfCfg, err = config.NewCGRConfigFromPath(accPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAccSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(accPrfCfg); err != nil {
		t.Fatal(err)
	}
}

func testAccSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(accPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAccSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testAccSRPCConn(t *testing.T) {
	var err error
	accSRPC, err = newRPCClient(accPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testGetAccProfileBeforeSet(t *testing.T) {
	var reply *utils.TenantIDWithAPIOpts
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ACCOUNT_IT_TEST",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

/*
func testAccSetAccProfile(t *testing.T) {
	accPrf := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAcc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
		Opts:   map[string]interface{}{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    12,
					},
				},
				Type: "*abstract",
				Opts: map[string]interface{}{
					"Destination": 10,
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: nil,
				Weight:    10,
			},
		},
	}
	var result *utils.Account
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "test_ID1",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAcc) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expectedAcc), utils.ToJSON(result))
	}
}
*/
//Kill the engine when it is about to be finished
func testAccSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
