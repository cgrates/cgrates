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
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	sagCfgPath   string
	sagCfg       *config.CGRConfig
	sagRPC       *birpc.Client
	sagProfile   *engine.SagProfileWithAPIOpts
	sagConfigDIR string

	sTestsSag = []func(t *testing.T){
		testSagSInitCfg,
		testSagSInitDataDb,
		testSagSResetStorDb,
		testSagSStartEngine,
		testSagSRPCConn,
		testSagSLoadAdd,
		testSagSSetSagProfile,
		testSagSGetSagProfileIDs,
		testSagSUpdateSagProfile,
		testSagSRemSagProfile,
		testSagSKillEngine,
		//cache test
		testSagSInitCfg,
		testSagSInitDataDb,
		testSagSResetStorDb,
		testSagSStartEngine,
		testSagSRPCConn,
		testSagSCacheTestGetNotFound,
		testSagSCacheTestSet,
		testSagSCacheReload,
		testSagSCacheTestGetFound,
		testSagSKillEngine,
	}
)

func TestSagSIT(t *testing.T) {

	switch *utils.DBType {
	case utils.MetaInternal:
		sagConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		sagConfigDIR = "tutmysql"
	case utils.MetaMongo:
		sagConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsSag {
		t.Run(sagConfigDIR, stest)
	}
}

func testSagSInitCfg(t *testing.T) {
	var err error
	sagCfgPath = path.Join(*utils.DataDir, "conf", "samples", sagConfigDIR)
	sagCfg, err = config.NewCGRConfigFromPath(sagCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testSagSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(sagCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testSagSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sagCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSagSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sagCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSagSRPCConn(t *testing.T) {
	var err error
	sagRPC, err = newRPCClient(sagCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}
func testSagSLoadAdd(t *testing.T) {
	sagProfile := &engine.SagProfileWithAPIOpts{
		SagProfile: &engine.SagProfile{
			Tenant:        "cgrates.org",
			ID:            "SG_Sum",
			QueryInterval: 2 * time.Minute,
			StatIDs:       []string{"SQProfile1"},
			MetricIDs:     []string{"*sum"},
		},
	}

	var result string
	if err := sagRPC.Call(context.Background(), utils.APIerSv1SetSagProfile, sagProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testSagSSetSagProfile(t *testing.T) {
	var (
		reply  *engine.SagProfileWithAPIOpts
		result string
	)
	sagProfile = &engine.SagProfileWithAPIOpts{
		SagProfile: &engine.SagProfile{
			Tenant:        "cgrates.org",
			ID:            "Sag1",
			QueryInterval: time.Second * 15,
			StatIDs:       []string{"SQ1", "SQ2"},
			MetricIDs:     []string{"*acc", "*sum"},
			Sorting:       "*asc",
			ThresholdIDs:  []string{"THD1", "THD2"}},
	}
	if err := sagRPC.Call(context.Background(), utils.APIerSv1GetSagProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Sag1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
	if err := sagRPC.Call(context.Background(), utils.APIerSv1SetSagProfile, sagProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expected: %v,Received: %v", utils.OK, result)
	}
	if err := sagRPC.Call(context.Background(), utils.APIerSv1GetSagProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Sag1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(sagProfile, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profile (-expected +got):\n%s", diff)
	}
}
func testSagSGetSagProfileIDs(t *testing.T) {
	expected := []string{"Sag1", "SG_Sum"}
	var result []string
	if err := sagRPC.Call(context.Background(), utils.APIerSv1GetSagProfileIDs, utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testSagSUpdateSagProfile(t *testing.T) {
	sagProfile.Sorting = "*desc"
	var (
		reply  *engine.SagProfileWithAPIOpts
		result string
	)
	if err := sagRPC.Call(context.Background(), utils.APIerSv1SetSagProfile, sagProfile, &result); err != nil {
		t.Error(err)
	}
	if err := sagRPC.Call(context.Background(), utils.APIerSv1GetSagProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Sag1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(sagProfile, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profile (-expected +got):\n%s", diff)
	}
}
func testSagSRemSagProfile(t *testing.T) {
	var (
		resp  string
		reply *engine.SagProfile
	)
	if err := sagRPC.Call(context.Background(), utils.APIerSv1RemoveSagProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Sag1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

	if err := sagRPC.Call(context.Background(), utils.APIerSv1GetSagProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Sag1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testSagSKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}

func testSagSCacheTestGetNotFound(t *testing.T) {
	var reply *engine.SagProfile
	if err := sagRPC.Call(context.Background(), utils.APIerSv1GetSagProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SAGS_CACHE"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
}

func testSagSCacheTestGetFound(t *testing.T) {
	var reply *engine.SagProfile
	if err := sagRPC.Call(context.Background(), utils.APIerSv1GetSagProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SAGS_CACHE"}, &reply); err != nil {
		t.Fatal(err)
	}
}

func testSagSCacheTestSet(t *testing.T) {
	sagProfile = &engine.SagProfileWithAPIOpts{
		SagProfile: &engine.SagProfile{
			Tenant: "cgrates.org",
			ID:     "SAGS_CACHE",
		},
		APIOpts: map[string]any{
			utils.CacheOpt: utils.MetaNone,
		},
	}
	var reply string
	if err := sagRPC.Call(context.Background(), utils.APIerSv1SetSagProfile, sagProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testSagSCacheReload(t *testing.T) {
	cache := &utils.AttrReloadCacheWithAPIOpts{
		SagProfileIDs: []string{"cgrates.org:SAGS_CACHE"},
	}
	var reply string
	if err := sagRPC.Call(context.Background(), utils.CacheSv1ReloadCache, cache, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}
