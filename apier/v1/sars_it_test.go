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
	sarCfgPath   string
	sarCfg       *config.CGRConfig
	sarRPC       *birpc.Client
	sarProfile   *engine.SarProfileWithAPIOpts
	sarConfigDIR string

	sTestsSar = []func(t *testing.T){
		testSarSInitCfg,
		testSarSInitDataDb,
		testSarSResetStorDb,
		testSarSStartEngine,
		testSarSRPCConn,
		testSarSLoadAdd,
		testSarSetSarProfile,
		testSarSGetSarProfileIDs,
		testSarSUpdateSarProfile,
		testSarSRemSarProfile,
		testSarSKillEngine,
	}
)

func TestSarSIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sarConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		sarConfigDIR = "tutmysql"
	case utils.MetaMongo:
		sarConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsSar {
		t.Run(sarConfigDIR, stest)
	}
}

func testSarSInitCfg(t *testing.T) {
	var err error
	sarCfgPath = path.Join(*utils.DataDir, "conf", "samples", sarConfigDIR)
	sarCfg, err = config.NewCGRConfigFromPath(sarCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testSarSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(sarCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testSarSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sarCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSarSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sarCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSarSRPCConn(t *testing.T) {
	var err error
	sarRPC, err = newRPCClient(sarCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}
func testSarSLoadAdd(t *testing.T) {
	sarProfile := &engine.SarProfileWithAPIOpts{
		SarProfile: &engine.SarProfile{
			Tenant:        "cgrates.org",
			ID:            "SR_AVG",
			QueryInterval: 2 * time.Minute,
			StatID:        "Stat1",
			Trend:         "*average",
		},
	}

	var result string
	if err := sarRPC.Call(context.Background(), utils.APIerSv1SetSarProfile, sarProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testSarSetSarProfile(t *testing.T) {
	var (
		reply  *engine.SarProfileWithAPIOpts
		result string
	)
	sarProfile = &engine.SarProfileWithAPIOpts{
		SarProfile: &engine.SarProfile{
			Tenant:        "cgrates.org",
			ID:            "Sar1",
			QueryInterval: time.Second * 15,
			ThresholdIDs:  []string{"THD1", "THD2"}},
	}
	if err := sarRPC.Call(context.Background(), utils.APIerSv1GetSarProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Sar1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
	if err := sarRPC.Call(context.Background(), utils.APIerSv1SetSarProfile, sarProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expected: %v,Received: %v", utils.OK, result)
	}
	if err := sarRPC.Call(context.Background(), utils.APIerSv1GetSarProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Sar1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(sarProfile, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profile (-expected +got):\n%s", diff)
	}
}
func testSarSGetSarProfileIDs(t *testing.T) {
	expected := []string{"Sag1", "SG_Sum"}
	var result []string
	if err := sarRPC.Call(context.Background(), utils.APIerSv1GetSarProfileIDs, utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testSarSUpdateSarProfile(t *testing.T) {
	var (
		reply  *engine.SarProfileWithAPIOpts
		result string
	)
	if err := sarRPC.Call(context.Background(), utils.APIerSv1SetSarProfile, sarProfile, &result); err != nil {
		t.Error(err)
	}
	if err := sarRPC.Call(context.Background(), utils.APIerSv1GetSarProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Sar1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(sarProfile, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profile (-expected +got):\n%s", diff)
	}
}
func testSarSRemSarProfile(t *testing.T) {
	var (
		resp  string
		reply *engine.SarProfile
	)
	if err := sarRPC.Call(context.Background(), utils.APIerSv1RemoveSarProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Sar1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

	if err := sarRPC.Call(context.Background(), utils.APIerSv1GetSarProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Sar1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testSarSKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
