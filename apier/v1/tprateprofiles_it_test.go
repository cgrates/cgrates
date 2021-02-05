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
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpRatePrfCfgPath   string
	tpRatePrfCfg       *config.CGRConfig
	tpRatePrfRPC       *rpc.Client
	tpRatePrfDataDir   = "/usr/share/cgrates"
	tpRatePrf          *utils.TPRateProfile
	tpRatePrfDelay     int
	tpRatePrfConfigDIR string //run tests for specific configuration
)

var sTestsTPRatePrf = []func(t *testing.T){
	testTPRatePrfInitCfg,
	testTPRatePrfResetStorDb,
	testTPRatePrfStartEngine,
	testTPRatePrfRPCConn,
	testTPRatePrfGetTPRatePrfBeforeSet,
	testTPRatePrfSetTPRatePrf,
	testTPRatePrfGetTPRatePrfAfterSet,
	testTPRatePrfGetTPRatePrfIDs,
	testTPRatePrfUpdateTPRatePrf,
	testTPRatePrfGetTPRatePrfAfterUpdate,
	testTPRatePrfRemTPRatePrf,
	testTPRatePrfGetTPRatePrfAfterRemove,
	testTPRatePrfKillEngine,
}

//Test start here
func TestTPRatePrfIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpRatePrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpRatePrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpRatePrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPRatePrf {
		t.Run(tpRatePrfConfigDIR, stest)
	}
}

func testTPRatePrfInitCfg(t *testing.T) {
	var err error
	tpRatePrfCfgPath = path.Join(tpRatePrfDataDir, "conf", "samples", tpRatePrfConfigDIR)
	tpRatePrfCfg, err = config.NewCGRConfigFromPath(tpRatePrfCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpRatePrfDelay = 1000
}

// Wipe out the cdr database
func testTPRatePrfResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpRatePrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPRatePrfStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpRatePrfCfgPath, tpRatePrfDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPRatePrfRPCConn(t *testing.T) {
	var err error
	tpRatePrfRPC, err = jsonrpc.Dial(utils.TCP, tpRatePrfCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPRatePrfGetTPRatePrfBeforeSet(t *testing.T) {
	var reply *utils.TPRateProfile
	if err := tpRatePrfRPC.Call(utils.APIerSv1GetTPRateProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Attr1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatePrfSetTPRatePrf(t *testing.T) {
	tpRatePrf = &utils.TPRateProfile{
		TPid:   "TP1",
		Tenant: "cgrates.org",
		ID:     "RT_SPECIAL_1002",
		Weight: 10,
		Rates: map[string]*utils.TPRate{
			"RT_ALWAYS": {
				ID:        "RT_ALWAYS",
				FilterIDs: []string{"* * * * *"},
				Weight:    0,
				Blocker:   false,
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.01,
						Unit:          "1m",
						Increment:     "1s",
					},
				},
			},
		},
	}
	var result string
	if err := tpRatePrfRPC.Call(utils.APIerSv1SetTPRateProfile, tpRatePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatePrfGetTPRatePrfAfterSet(t *testing.T) {
	var reply *utils.TPRateProfile
	if err := tpRatePrfRPC.Call(utils.APIerSv1GetTPRateProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "RT_SPECIAL_1002"}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	if !reflect.DeepEqual(tpRatePrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpRatePrf), utils.ToJSON(reply))
	}
}

func testTPRatePrfGetTPRatePrfIDs(t *testing.T) {
	var result []string
	expectedTPID := []string{"cgrates.org:RT_SPECIAL_1002"}
	if err := tpRatePrfRPC.Call(utils.APIerSv1GetTPRateProfileIds,
		&AttrGetTPRateProfileIds{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPRatePrfUpdateTPRatePrf(t *testing.T) {
	tpRatePrf.Rates = map[string]*utils.TPRate{
		"RT_ALWAYS": {
			ID:        "RT_ALWAYS",
			FilterIDs: []string{"* * * * *"},
			Weight:    0,
			Blocker:   false,
			IntervalRates: []*utils.TPIntervalRate{
				{
					IntervalStart: "0s",
					RecurrentFee:  0.01,
					Unit:          "1m",
					Increment:     "1s",
				},
			},
		},
	}
	var result string
	if err := tpRatePrfRPC.Call(utils.APIerSv1SetTPRateProfile, tpRatePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatePrfGetTPRatePrfAfterUpdate(t *testing.T) {
	var reply *utils.TPRateProfile
	revTPRatePrf := &utils.TPRateProfile{
		TPid:   "TP1",
		Tenant: "cgrates.org",
		ID:     "RT_SPECIAL_1002",
		Weight: 10,
	}

	if err := tpRatePrfRPC.Call(utils.APIerSv1GetTPRateProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "RT_SPECIAL_1002"}, &reply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tpRatePrf, reply) && !reflect.DeepEqual(revTPRatePrf, reply) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(tpRatePrf), utils.ToJSON(reply))
	}
}

func testTPRatePrfRemTPRatePrf(t *testing.T) {
	var resp string
	if err := tpRatePrfRPC.Call(utils.APIerSv1RemoveTPRateProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "RT_SPECIAL_1002"},
		&resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPRatePrfGetTPRatePrfAfterRemove(t *testing.T) {
	var reply *utils.TPActionProfile
	if err := tpRatePrfRPC.Call(utils.APIerSv1GetTPRateProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "RT_SPECIAL_1002"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatePrfKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpRatePrfDelay); err != nil {
		t.Error(err)
	}
}
