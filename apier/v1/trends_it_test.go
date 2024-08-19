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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	trendCfgPath   string
	trendCfg       *config.CGRConfig
	trendRPC       *birpc.Client
	trendProfile   *engine.TrendProfileWithAPIOpts
	trendProfile2  *engine.TrendProfileWithAPIOpts
	trendConfigDIR string

	sTestsTrend = []func(t *testing.T){
		testTrendSInitCfg,
		testTrendSInitDataDb,
		testTrendSResetStorDb,
		testTrendSStartEngine,
		testTrendSRPCConn,
		testTrendSLoadAdd,
		testTrendSetTrendProfile,
		testTrendSGetTrendProfileIDs,
		testTrendSGetTrendProfiles,
		testTrendSUpdateTrendProfile,
		testTrendSRemTrendProfile,
		testTrendSKillEngine,
	}
)

func TestTrendSIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		trendConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		trendConfigDIR = "tutmysql"
	case utils.MetaMongo:
		trendConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTrend {
		t.Run(trendConfigDIR, stest)
	}
}

func testTrendSInitCfg(t *testing.T) {
	var err error
	trendCfgPath = path.Join(*utils.DataDir, "conf", "samples", trendConfigDIR)
	trendCfg, err = config.NewCGRConfigFromPath(trendCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testTrendSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(trendCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testTrendSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(trendCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTrendSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(trendCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTrendSRPCConn(t *testing.T) {
	var err error
	trendRPC, err = newRPCClient(trendCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}
func testTrendSLoadAdd(t *testing.T) {
	trendProfile = &engine.TrendProfileWithAPIOpts{
		TrendProfile: &engine.TrendProfile{
			Tenant: "cgrates.org",
			ID:     "TR_AVG",
			StatID: "Stat1",
		},
	}

	var result string
	if err := trendRPC.Call(context.Background(), utils.APIerSv1SetTrendProfile, trendProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testTrendSetTrendProfile(t *testing.T) {
	var (
		reply  *engine.TrendProfileWithAPIOpts
		result string
	)
	trendProfile2 = &engine.TrendProfileWithAPIOpts{
		TrendProfile: &engine.TrendProfile{
			Tenant:       "cgrates.org",
			ID:           "Trend1",
			ThresholdIDs: []string{"THD1"}},
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Trend1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1SetTrendProfile, trendProfile2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expected: %v,Received: %v", utils.OK, result)
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Trend1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(trendProfile2, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profile (-expected +got):\n%s", diff)
	}
}
func testTrendSGetTrendProfileIDs(t *testing.T) {
	expected := []string{"Trend1", "TR_AVG"}
	var result []string
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfileIDs, utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testTrendSGetTrendProfiles(t *testing.T) {
	trendPrf := &engine.TrendProfileWithAPIOpts{
		TrendProfile: &engine.TrendProfile{
			Tenant:   "tenant1",
			ID:       "Trend_Last",
			StatID:   "Stat1",
			Schedule: "*now",
		}}
	var result string
	if err := trendRPC.Call(context.Background(), utils.APIerSv1SetTrendProfile, trendPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expected: %v,Received: %v", utils.OK, result)
	}
	expRes := []*engine.TrendProfile{
		trendPrf.TrendProfile,
		trendProfile.TrendProfile,
		trendProfile2.TrendProfile,
	}
	var tResult []*engine.TrendProfile
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfiles, engine.TrendProfilesAPI{}, &tResult); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(expRes, tResult, cmpopts.SortSlices(func(a, b *engine.TrendProfile) bool { return a.ID < b.ID })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profiles (-expected +got):\n%s", diff)
	}

	expRes = []*engine.TrendProfile{
		trendProfile.TrendProfile,
		trendProfile2.TrendProfile,
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfiles, engine.TrendProfilesAPI{Tenant: "cgrates.org"}, &tResult); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(expRes, tResult, cmpopts.SortSlices(func(a, b *engine.TrendProfile) bool { return a.ID < b.ID })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profiles (-expected +got):\n%s", diff)
	}

	expRes = []*engine.TrendProfile{
		trendProfile2.TrendProfile,
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfiles, engine.TrendProfilesAPI{Tenant: "cgrates.org", TpIDs: []string{"Trend1"}}, &tResult); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(expRes, tResult, cmpopts.SortSlices(func(a, b *engine.TrendProfile) bool { return a.ID < b.ID })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profiles (-expected +got):\n%s", diff)
	}
}

func testTrendSUpdateTrendProfile(t *testing.T) {
	var (
		reply  *engine.TrendProfileWithAPIOpts
		result string
	)
	if err := trendRPC.Call(context.Background(), utils.APIerSv1SetTrendProfile, trendProfile, &result); err != nil {
		t.Error(err)
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Trend1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(trendProfile2, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profile (-expected +got):\n%s", diff)
	}
}

func testTrendSRemTrendProfile(t *testing.T) {
	var (
		resp  string
		reply *engine.TrendProfile
	)
	if err := trendRPC.Call(context.Background(), utils.APIerSv1RemoveTrendProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Trend1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Trend1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTrendSKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
