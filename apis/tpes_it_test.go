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

package apis

import (
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/tpes"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpesCfgPath   string
	tpesCfg       *config.CGRConfig
	tpeSRPC       *birpc.Client
	tpeSConfigDIR string //run tests for specific configuration

	sTestTpes = []func(t *testing.T){
		testTPeSInitCfg,
		testTPeSInitDataDb,
		testTPeSStartEngine,
		testTPeSRPCConn,
		testTPeSPing,
		testTPeSExportTariffPlan,
		testTPeSKillEngine,
	}
)

func TestTPeSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpeSConfigDIR = "tutinternal"
	case utils.MetaMongo:
		tpeSConfigDIR = "tutmongo"
	case utils.MetaMySQL:
		tpeSConfigDIR = "tutmysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestTpes {
		t.Run(tpeSConfigDIR, stest)
	}
}

func testTPeSInitCfg(t *testing.T) {
	var err error
	tpesCfgPath = path.Join(*dataDir, "conf", "samples", tpeSConfigDIR)
	tpesCfg, err = config.NewCGRConfigFromPath(context.Background(), tpesCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testTPeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(tpesCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpesCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testTPeSPing(t *testing.T) {
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.TPeSv1Ping, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
}

func testTPeSExportTariffPlan(t *testing.T) {
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
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
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	attrPrf1 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST_SECOND",
			FilterIDs: []string{"*string:~*opts.*context:*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
	}
	var reply1 string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != utils.OK {
		t.Error(err)
	}
	rsPrf1 := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "ResGroup1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			Limit:             10,
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				}},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	var replystr string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		rsPrf1, &replystr); err != nil {
		t.Error(err)
	} else if replystr != utils.OK {
		t.Error("Unexpected reply returned", replystr)
	}

	rsPrf2 := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "ResGroup2",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			Limit:             10,
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		rsPrf2, &replystr); err != nil {
		t.Error(err)
	} else if replystr != utils.OK {
		t.Error("Unexpected reply returned", replystr)
	}
	var replyBts []byte
	if err := tpeSRPC.Call(context.Background(), utils.TPeSv1ExportTariffPlan, &tpes.ArgsExportTP{
		Tenant: "cgrates.org",
		ExportItems: map[string][]string{
			utils.MetaAttributes: {"TEST_ATTRIBUTES_IT_TEST", "TEST_ATTRIBUTES_IT_TEST_SECOND"},
			utils.MetaResources:  {"ResGroup1", "ResGroup2"},
		},
	}, &replyBts); err != nil {
		t.Error(err)
	} /*
		t.Errorf("received: %v", string(replyBts))

		rdr, err := zip.NewReader(bytes.NewReader(replyBts), int64(len(reply)))
		if err != nil {
			t.Error(err)
		}
		csvRply := make([][]string, 6)
		for _, f := range rdr.File {
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			info := csv.NewReader(rc)
			//info.FieldsPerRecord = -1
			csvRply, err = info.ReadAll()
			if err != nil {
				t.Error(err)
			}
			rc.Close()
		}
		t.Errorf("received: %v", utils.ToJSON(csvRply))
	*/

}

func testTPeSRPCConn(t *testing.T) {
	var err error
	tpeSRPC, err = newRPCClient(tpesCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testTPeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
