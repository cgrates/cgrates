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

package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	filterCfgPath   string
	filterCfg       *config.CGRConfig
	filterRPC       *rpc.Client
	filterDataDir   = "/usr/share/cgrates"
	filter          *FilterWithCache
	filterConfigDIR string //run tests for specific configuration

	sTestsFilter = []func(t *testing.T){
		testFilterInitCfg,
		testFilterResetDataDB,
		testFilterStartEngine,
		testFilterRpcConn,
		testFilterGetFilterBeforeSet,
		testFilterSetFilter,
		testFilterGetFilterAfterSet,
		testFilterGetFilterIDs,
		testFilterUpdateFilter,
		testFilterGetFilterAfterUpdate,
		testFilterRemoveFilter,
		testFilterGetFilterAfterRemove,
		testFilterKillEngine,
	}
)

//Test start here
func TestFilterIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		filterConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		filterConfigDIR = "tutmysql"
	case utils.MetaMongo:
		filterConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsFilter {
		t.Run(filterConfigDIR, stest)
	}
}

func testFilterInitCfg(t *testing.T) {
	var err error
	filterCfgPath = path.Join(filterDataDir, "conf", "samples", filterConfigDIR)
	filterCfg, err = config.NewCGRConfigFromPath(filterCfgPath)
	if err != nil {
		t.Error(err)
	}
	filterCfg.DataFolderPath = filterDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(filterCfg)
}

// Wipe out the cdr database
func testFilterResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(filterCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testFilterStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(filterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testFilterRpcConn(t *testing.T) {
	var err error
	filterRPC, err = newRPCClient(filterCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testFilterGetFilterBeforeSet(t *testing.T) {
	var reply *engine.Filter
	if err := filterRPC.Call(utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testFilterSetFilter(t *testing.T) {
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "Filter1",
			Rules: []*engine.FilterRule{
				{
					Element: utils.MetaString,
					Type:    "~Account",
					Values:  []string{"1001", "1002"},
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}

	var result string
	if err := filterRPC.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterGetFilterIDs(t *testing.T) {
	expected := []string{"Filter1"}
	var result []string
	if err := filterRPC.Call(utils.APIerSv1GetFilterIDs, utils.TenantArgWithPaginator{TenantArg: utils.TenantArg{Tenant: "cgrates.org"}}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testFilterGetFilterAfterSet(t *testing.T) {
	var reply *engine.Filter
	if err := filterRPC.Call(utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(filter.Filter, reply) {
		t.Errorf("Expecting : %+v, received: %+v", filter.Filter, reply)
	}
}

func testFilterUpdateFilter(t *testing.T) {
	filter.Rules = []*engine.FilterRule{
		{
			Element: utils.MetaString,
			Type:    "~Account",
			Values:  []string{"1001", "1002"},
		},
		{
			Element: utils.MetaPrefix,
			Type:    "~Destination",
			Values:  []string{"10", "20"},
		},
	}
	var result string
	if err := filterRPC.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterGetFilterAfterUpdate(t *testing.T) {
	var reply *engine.Filter
	if err := filterRPC.Call(utils.APIerSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(filter.Filter, reply) {
		t.Errorf("Expecting : %+v, received: %+v", filter.Filter, reply)
	}
}

func testFilterRemoveFilter(t *testing.T) {
	var resp string
	if err := filterRPC.Call(utils.APIerSv1RemoveFilter,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "Filter1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testFilterGetFilterAfterRemove(t *testing.T) {
	var reply *engine.Filter
	if err := filterRPC.Call(utils.APIerSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testFilterKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
