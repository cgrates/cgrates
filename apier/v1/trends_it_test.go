//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	trendCfgPath   string
	trendCfg       *config.CGRConfig
	trendRPC       *birpc.Client
	trendProfile   *engine.TrendProfileWithAPIOpts
	trendConfigDIR string

	sTestsTrend = []func(t *testing.T){
		testTrendSInitCfg,
		testTrendSInitDataDb,
		testTrendSResetStorDb,
		testTrendSStartEngine,
		testTrendSRPCConn,
		testTrendSFromFolder,
		testTrendSetTrendProfile,
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
	if err := engine.InitDataDB(trendCfg); err != nil {
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
func testTrendSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testTrendSetTrendProfile(t *testing.T) {
	var (
		reply  *engine.TrendProfileWithAPIOpts
		result string
	)
	trendProfile = &engine.TrendProfileWithAPIOpts{
		TrendProfile: &engine.TrendProfile{
			Tenant:          "cgrates.org",
			ID:              "Trend1",
			ThresholdIDs:    []string{"THD1"},
			CorrelationType: utils.MetaAverage,
			MinItems:        1,
			QueueLength:     10,
			StatID:          "Stats1",
			Schedule:        "@every 4s",
		},
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Trend1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1SetTrendProfile, trendProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expected: %v,Received: %v", utils.OK, result)
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Trend1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(trendProfile, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
		t.Errorf("Unnexpected profile (-expected +got):\n%s", diff)
	}
}

func testTrendSUpdateTrendProfile(t *testing.T) {
	trendProfile.MinItems = 4
	trendProfile.QueueLength = 100
	trendProfile.CorrelationType = utils.MetaLast
	trendProfile.Schedule = "@every 1s"
	trendProfile.StatID = "Stats2"
	var (
		reply  *engine.TrendProfileWithAPIOpts
		result string
	)
	expTrn := &engine.TrendProfileWithAPIOpts{
		TrendProfile: &engine.TrendProfile{
			Tenant:          "cgrates.org",
			ID:              "Trend1",
			ThresholdIDs:    []string{"THD1"},
			CorrelationType: utils.MetaLast,
			MinItems:        4,
			QueueLength:     100,
			StatID:          "Stats2",
			Schedule:        "@every 1s",
		},
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1SetTrendProfile, trendProfile, &result); err != nil {
		t.Error(err)
	}
	if err := trendRPC.Call(context.Background(), utils.APIerSv1GetTrendProfile, &utils.TenantID{Tenant: "cgrates.org", ID: "Trend1"}, &reply); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(expTrn, reply, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != utils.EmptyString {
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
