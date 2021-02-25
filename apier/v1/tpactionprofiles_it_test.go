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
	tpActPrfCfgPath   string
	tpActPrfCfg       *config.CGRConfig
	tpActPrfRPC       *rpc.Client
	tpActPrf          *utils.TPActionProfile
	tpActPrfDelay     int
	tpActPrfConfigDIR string //run tests for specific configuration
)

var sTestsTPActPrf = []func(t *testing.T){
	testTPActPrfInitCfg,
	testTPActPrfResetStorDb,
	testTPActPrfStartEngine,
	testTPActPrfRPCConn,
	testTPActPrfGetTPActPrfBeforeSet,
	testTPActPrfSetTPActPrf,
	testTPActPrfGetTPActPrfAfterSet,
	testTPActPrfGetTPActPrfIDs,
	testTPActPrfUpdateTPActPrf,
	testTPActPrfGetTPActPrfAfterUpdate,
	testTPActPrfRemTPActPrf,
	testTPActPrfGetTPActPrfAfterRemove,
	testTPActPrfKillEngine,
}

//Test start here
func TestTPActPrfIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpActPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpActPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpActPrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPActPrf {
		t.Run(tpActPrfConfigDIR, stest)
	}
}

func testTPActPrfInitCfg(t *testing.T) {
	var err error
	tpActPrfCfgPath = path.Join(*dataDir, "conf", "samples", tpActPrfConfigDIR)
	tpActPrfCfg, err = config.NewCGRConfigFromPath(tpActPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpActPrfDelay = 1000
}

// Wipe out the cdr database
func testTPActPrfResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpActPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPActPrfStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpActPrfCfgPath, tpActPrfDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPActPrfRPCConn(t *testing.T) {
	var err error
	tpActPrfRPC, err = jsonrpc.Dial(utils.TCP, tpActPrfCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPActPrfGetTPActPrfBeforeSet(t *testing.T) {
	var reply *utils.TPActionProfile
	if err := tpActPrfRPC.Call(utils.APIerSv1GetTPActionProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Attr1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPActPrfSetTPActPrf(t *testing.T) {
	tpActPrf = &utils.TPActionProfile{
		TPid:     "TP1",
		Tenant:   "cgrates.org",
		ID:       "ONE_TIME_ACT",
		Weight:   10,
		Schedule: utils.MetaASAP,
		Targets: []*utils.TPActionTarget{
			{
				TargetType: utils.MetaAccounts,
				TargetIDs:  []string{"1001"},
			},
		},
		Actions: []*utils.TPAPAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*utils.TPAPDiktat{{
					Path:  "~*balance.TestBalance.Value",
					Value: "10",
				}},
			},
		},
	}
	sort.Strings(tpActPrf.FilterIDs)
	var result string
	if err := tpActPrfRPC.Call(utils.APIerSv1SetTPActionProfile, tpActPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPActPrfGetTPActPrfAfterSet(t *testing.T) {
	var reply *utils.TPActionProfile
	if err := tpActPrfRPC.Call(utils.APIerSv1GetTPActionProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "ONE_TIME_ACT"}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	if !reflect.DeepEqual(tpActPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpActPrf), utils.ToJSON(reply))
	}
}

func testTPActPrfGetTPActPrfIDs(t *testing.T) {
	var result []string
	expectedTPID := []string{"cgrates.org:ONE_TIME_ACT"}
	if err := tpActPrfRPC.Call(utils.APIerSv1GetTPActionProfileIDs,
		&AttrGetTPActionProfileIDs{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPActPrfUpdateTPActPrf(t *testing.T) {
	tpActPrf.Actions = []*utils.TPAPAction{
		{
			ID:        "new_TOPUP",
			FilterIDs: []string{},
			Type:      "*topup",
			Diktats: []*utils.TPAPDiktat{{
				Path:  "~*balance.TestBalance.Value",
				Value: "10",
			}},
		},
	}
	var result string
	if err := tpActPrfRPC.Call(utils.APIerSv1SetTPActionProfile, tpActPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPActPrfGetTPActPrfAfterUpdate(t *testing.T) {
	var reply *utils.TPActionProfile
	revTPActPrf := &utils.TPActionProfile{
		TPid:     "TP1",
		Tenant:   "cgrates.org",
		ID:       "ONE_TIME_ACT",
		Weight:   10,
		Schedule: utils.MetaASAP,
		Targets: []*utils.TPActionTarget{
			{
				TargetType: utils.MetaAccounts,
				TargetIDs:  []string{"1001"},
			},
		},
		Actions: []*utils.TPAPAction{
			{
				ID:        "new_TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*utils.TPAPDiktat{{
					Path:  "~*balance.TestBalance.Value",
					Value: "10",
				}},
			},
		},
	}
	sort.Strings(revTPActPrf.FilterIDs)
	sort.Slice(revTPActPrf.Actions, func(i, j int) bool {
		return strings.Compare(revTPActPrf.Actions[i].ID, revTPActPrf.Actions[j].ID) == -1
	})
	if err := tpActPrfRPC.Call(utils.APIerSv1GetTPActionProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "ONE_TIME_ACT"}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	sort.Slice(reply.Actions, func(i, j int) bool {
		return strings.Compare(reply.Actions[i].ID, reply.Actions[j].ID) == -1
	})
	if !reflect.DeepEqual(tpActPrf, reply) && !reflect.DeepEqual(revTPActPrf, reply) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(tpActPrf), utils.ToJSON(reply))
	}
}

func testTPActPrfRemTPActPrf(t *testing.T) {
	var resp string
	if err := tpActPrfRPC.Call(utils.APIerSv1RemoveTPActionProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "ONE_TIME_ACT"},
		&resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPActPrfGetTPActPrfAfterRemove(t *testing.T) {
	var reply *utils.TPActionProfile
	if err := tpActPrfRPC.Call(utils.APIerSv1GetTPActionProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "ONE_TIME_ACT"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPActPrfKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpActPrfDelay); err != nil {
		t.Error(err)
	}
}
