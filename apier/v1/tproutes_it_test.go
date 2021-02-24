// +build offline

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
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpRouteCfgPath   string
	tpRouteCfg       *config.CGRConfig
	tpRouteRPC       *rpc.Client
	tpRoutePrf       *utils.TPRouteProfile
	tpRouteDelay     int
	tpRouteConfigDIR string //run tests for specific configuration
)

var sTestsTPRoute = []func(t *testing.T){
	testTPRouteInitCfg,
	testTPRouteResetStorDb,
	testTPRouteStartEngine,
	testTPRouteRPCConn,
	testTPRouteGetTPRouteBeforeSet,
	testTPRouteSetTPRoute,
	testTPRouteGetTPRouteAfterSet,
	testTPRouteGetTPRouteIDs,
	testTPRouteUpdateTPRoute,
	testTPRouteGetTPRouteAfterUpdate,
	testTPRouteRemTPRoute,
	testTPRouteGetTPRouteAfterRemove,
	testTPRouteKillEngine,
}

//Test start here
func TestTPRouteIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpRouteConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpRouteConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpRouteConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPRoute {
		t.Run(tpRouteConfigDIR, stest)
	}
}

func testTPRouteInitCfg(t *testing.T) {
	var err error
	tpRouteCfgPath = path.Join(*dataDir, "conf", "samples", tpRouteConfigDIR)
	tpRouteCfg, err = config.NewCGRConfigFromPath(tpRouteCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpRouteDelay = 1000

}

// Wipe out the cdr database
func testTPRouteResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpRouteCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPRouteStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpRouteCfgPath, tpRouteDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPRouteRPCConn(t *testing.T) {
	var err error
	tpRouteRPC, err = jsonrpc.Dial(utils.TCP, tpRouteCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPRouteGetTPRouteBeforeSet(t *testing.T) {
	var reply *utils.TPRoute
	if err := tpRouteRPC.Call(utils.APIerSv1GetTPRouteProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "RoutePrf"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRouteSetTPRoute(t *testing.T) {
	tpRoutePrf = &utils.TPRouteProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "RoutePrf",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		Sorting: "*lc",
		Routes: []*utils.TPRoute{
			&utils.TPRoute{
				ID:              "route1",
				FilterIDs:       []string{"FLTR_1"},
				AccountIDs:      []string{"Acc1", "Acc2"},
				RatingPlanIDs:   []string{"RPL_1"},
				ResourceIDs:     []string{"ResGroup1"},
				StatIDs:         []string{"Stat1"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: "SortingParam1",
			},
		},
		Weight: 20,
	}
	sort.Strings(tpRoutePrf.FilterIDs)
	var result string
	if err := tpRouteRPC.Call(utils.APIerSv1SetTPRouteProfile,
		tpRoutePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRouteGetTPRouteAfterSet(t *testing.T) {
	var reply *utils.TPRouteProfile
	if err := tpRouteRPC.Call(utils.APIerSv1GetTPRouteProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "RoutePrf"}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	if !reflect.DeepEqual(tpRoutePrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpRoutePrf), utils.ToJSON(reply))
	}
}

func testTPRouteGetTPRouteIDs(t *testing.T) {
	var result []string
	expectedTPID := []string{"cgrates.org:RoutePrf"}
	if err := tpRouteRPC.Call(utils.APIerSv1GetTPRouteProfileIDs,
		&AttrGetTPRouteProfileIDs{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}

}

func testTPRouteUpdateTPRoute(t *testing.T) {
	tpRoutePrf.Routes = []*utils.TPRoute{
		&utils.TPRoute{
			ID:              "route1",
			FilterIDs:       []string{"FLTR_1"},
			AccountIDs:      []string{"Acc1", "Acc2"},
			RatingPlanIDs:   []string{"RPL_1"},
			ResourceIDs:     []string{"ResGroup1"},
			StatIDs:         []string{"Stat1"},
			Weight:          10,
			Blocker:         true,
			RouteParameters: "SortingParam1",
		},
		&utils.TPRoute{
			ID:              "route2",
			FilterIDs:       []string{"FLTR_1"},
			AccountIDs:      []string{"Acc3"},
			RatingPlanIDs:   []string{"RPL_1"},
			ResourceIDs:     []string{"ResGroup1"},
			StatIDs:         []string{"Stat1"},
			Weight:          20,
			Blocker:         false,
			RouteParameters: "SortingParam2",
		},
	}
	var result string
	if err := tpRouteRPC.Call(utils.APIerSv1SetTPRouteProfile,
		tpRoutePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	sort.Slice(tpRoutePrf.Routes, func(i, j int) bool {
		return strings.Compare(tpRoutePrf.Routes[i].ID, tpRoutePrf.Routes[j].ID) == -1
	})
}

func testTPRouteGetTPRouteAfterUpdate(t *testing.T) {
	var reply *utils.TPRouteProfile
	if err := tpRouteRPC.Call(utils.APIerSv1GetTPRouteProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "RoutePrf"}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	sort.Slice(reply.Routes, func(i, j int) bool {
		return strings.Compare(reply.Routes[i].ID, reply.Routes[j].ID) == -1
	})
	if !reflect.DeepEqual(tpRoutePrf.Routes, reply.Routes) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(tpRoutePrf), utils.ToJSON(reply))
	}
}

func testTPRouteRemTPRoute(t *testing.T) {
	var resp string
	if err := tpRouteRPC.Call(utils.APIerSv1RemoveTPRouteProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "RoutePrf"},
		&resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPRouteGetTPRouteAfterRemove(t *testing.T) {
	var reply *utils.TPRouteProfile
	if err := tpRouteRPC.Call(utils.APIerSv1GetTPRouteProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "RoutePrf"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRouteKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpRouteDelay); err != nil {
		t.Error(err)
	}
}
