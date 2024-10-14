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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fltrSepCfgPath string
	fltrSepCfg     *config.CGRConfig
	fltrSepRPC     *birpc.Client
	fltrSepDelay   int
	fltrSepConfDIR string //run tests for specific configuration

	sTestsFltrSep = []func(t *testing.T){
		testFltrSepLoadConfig,
		testFltrSepInitDataDb,
		testFltrSepResetStorDb,
		testFltrSepStartEngine,
		testFltrSepRpcConn,
		testFltrSepLoadTarrifPlans,
		testFltrSepFilterSeparation,
		testFltrSepStopEngine,
	}
)

// Test start here
func TestFltrSepIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		fltrSepConfDIR = "tutinternal"
	case utils.MetaMySQL:
		fltrSepConfDIR = "tutmysql"
	case utils.MetaMongo:
		fltrSepConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsFltrSep {
		t.Run(fltrSepConfDIR, stest)
	}
}

func testFltrSepLoadConfig(t *testing.T) {
	var err error
	fltrSepCfgPath = path.Join(*utils.DataDir, "conf", "samples", fltrSepConfDIR)
	if fltrSepCfg, err = config.NewCGRConfigFromPath(fltrSepCfgPath); err != nil {
		t.Error(err)
	}
	fltrSepDelay = 1000
}

func testFltrSepInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(fltrSepCfg); err != nil {
		t.Fatal(err)
	}
}

func testFltrSepResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(fltrSepCfg); err != nil {
		t.Fatal(err)
	}
}

func testFltrSepStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fltrSepCfgPath, fltrSepDelay); err != nil {
		t.Fatal(err)
	}
}

func testFltrSepRpcConn(t *testing.T) {
	fltrSepRPC = engine.NewRPCClient(t, fltrSepCfg.ListenCfg())
}

func testFltrSepLoadTarrifPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "fltr_sep")}
	if err := fltrSepRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testFltrSepFilterSeparation(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "filter_separation_test",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaAny,
		},
	}

	eAttrPrf := &engine.AttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_FLTR_TEST",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003|1101", "*prefix:~*req.Account:10"},
		Contexts:  []string{utils.MetaAny},
		Attributes: []*engine.Attribute{
			{
				Path:      utils.MetaReq + utils.NestingSep + "TestField",
				Value:     config.NewRSRParsersMustCompile("testValue", utils.InfieldSep),
				Type:      utils.MetaConstant,
				FilterIDs: []string{},
			},
		},
		Weight: 20.0,
	}

	eAttrPrf.Compile()
	var attrReply *engine.AttributeProfile

	// first option of the first filter and the second filter match
	if err := fltrSepRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Error(err)
	} else {
		attrReply.Compile()
		sort.Slice(attrReply.FilterIDs, func(i, j int) bool {
			return attrReply.FilterIDs[i] > attrReply.FilterIDs[j]
		})
		if !reflect.DeepEqual(eAttrPrf, attrReply) {
			t.Errorf("expected: %+v, \nreceived: %+v",
				utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
		}
	}

	// third option of the first filter and the second filter match
	ev.Event[utils.AccountField] = "1003"
	if err := fltrSepRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Error(err)
	} else {
		attrReply.Compile()
		sort.Slice(attrReply.FilterIDs, func(i, j int) bool {
			return attrReply.FilterIDs[i] > attrReply.FilterIDs[j]
		})
		if !reflect.DeepEqual(eAttrPrf, attrReply) {
			t.Errorf("expected: %+v, \nreceived: %+v",
				utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
		}
	}

	// the second filter matches while none of the options from the first filter match
	ev.Event[utils.AccountField] = "1004"
	if err := fltrSepRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	// fourth option of the first filter matches while the second filter doesn't
	ev.Event[utils.AccountField] = "1101"
	if err := fltrSepRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testFltrSepStopEngine(t *testing.T) {
	if err := engine.KillEngine(fltrSepDelay); err != nil {
		t.Error(err)
	}
}
