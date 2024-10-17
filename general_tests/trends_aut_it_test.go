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
	"os/exec"
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
	trendAuQCfgPath string
	trendAuQCfg     *config.CGRConfig
	trendAuQRpc     *birpc.Client
	trendAuQConfDIR string //run tests for specific configuration

	sTeststrendAuQEmpty = []func(t *testing.T){
		testtrendAuQLoadConfig,
		testtrendAuQInitDataDb,
		testtrendAuQFromFolder,
		testtrendAuQStartEngine,
		testtrendAuQRpcConn,
		testScheduledTrends,
		testtrendAuQStopEngine,
	}

	sTeststrendAuQSchedIDs = []func(t *testing.T){
		testtrendAuQLoadConfig,
		testtrendAuQInitDataDb,
		testtrendAuQFromFolder,
		testtrendAuQStartEngine,
		testtrendAuQRpcConn,
		testScheduledTrends2,
		testtrendAuQStopEngine,
	}
)

// Test start here
func TestTrendAuQEmptyScheduleIDs(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	case utils.MetaMySQL:
		trendAuQConfDIR = "trends_mysql"
	case utils.MetaMongo:
		trendAuQConfDIR = "trends_mongo"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTeststrendAuQEmpty {
		t.Run(trendAuQConfDIR, stest)
	}
}

func TestTrendAuQScheduleIDs(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	case utils.MetaMySQL:
		trendAuQConfDIR = "trends_schedIDs_mysql"
	case utils.MetaMongo:
		trendAuQConfDIR = "trends_schedIDs_mongo"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTeststrendAuQSchedIDs {
		t.Run(trendAuQConfDIR, stest)
	}
}

func testtrendAuQLoadConfig(t *testing.T) {
	var err error
	trendAuQCfgPath = path.Join(*utils.DataDir, "conf", "samples", "trends", trendAuQConfDIR)
	if trendAuQCfg, err = config.NewCGRConfigFromPath(trendAuQCfgPath); err != nil {
		t.Error(err)
	}
}

func testtrendAuQInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(trendAuQCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testtrendAuQResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(trendAuQCfg); err != nil {
		t.Fatal(err)
	}
}

func testtrendAuQStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(trendAuQCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testtrendAuQRpcConn(t *testing.T) {
	trendAuQRpc = engine.NewRPCClient(t, trendAuQCfg.ListenCfg())
}

func testtrendAuQFromFolder(t *testing.T) {
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", trendAuQCfgPath, "-path", path.Join(*utils.DataDir, "tariffplans", "tuttrends"))
		if err := loader.Start(); err != nil {
			t.Error(err)
		}
		loader.Wait()
		wchan <- struct{}{}
	}()
	select {
	case <-wchan:
	case <-time.After(1 * time.Second):
		t.Errorf("cgr-loader failed: ")
	}
}

func testScheduledTrends(t *testing.T) {

	// getting all scheduled trends for a tenant
	var schedTrends []utils.ScheduledTrend
	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}, TrendIDPrefixes: []string{}}, &schedTrends); err != nil {
		t.Error(err)
	} else if len(schedTrends) != 3 {
		t.Errorf("expected 3 schedTrends, got %d", len(schedTrends))
	}
	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "tenant1"}}, TrendIDPrefixes: []string{}}, &schedTrends); err != nil {
		t.Error(err)
	} else if len(schedTrends) != 2 {
		t.Errorf("expected 2 schedTrends, got %d", len(schedTrends))
	}
	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "tenant2"}}, TrendIDPrefixes: []string{}}, &schedTrends); err != nil {
		t.Error(err)
	} else if len(schedTrends) != 2 {
		t.Errorf("expected 2 schedTrends, got %d", len(schedTrends))
	}

	// getting scheduled trends by the prefix
	expTrends := []utils.ScheduledTrend{
		{
			TrendID: "TREND_1",
			Next:    time.Now().Add(1 * time.Second),
		},
		{
			TrendID: "TREND_2",
			Next:    time.Now().Add(4 * time.Second),
		},
	}

	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}, TrendIDPrefixes: []string{"TREND"}}, &schedTrends); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(schedTrends[0], expTrends[0], cmpopts.EquateApproxTime(4*time.Second), cmpopts.IgnoreFields(utils.ScheduledTrend{}, "Previous")); diff != utils.EmptyString {
		t.Errorf("unexpected scheduled trends (-want +got)\n%s", diff)
	}

	expTrends = []utils.ScheduledTrend{
		{
			TrendID: "TR_1hr",
		},
	}

	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "tenant1"}}, TrendIDPrefixes: []string{"TR_1"}}, &schedTrends); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(schedTrends, expTrends, cmpopts.IgnoreFields(utils.ScheduledTrend{}, "Next"), cmpopts.IgnoreFields(utils.ScheduledTrend{}, "Previous")); diff != utils.EmptyString {
		t.Errorf("unexpected scheduled trends (-want +got)\n%s", diff)
	}

}

func testScheduledTrends2(t *testing.T) {

	// getting all scheduled trends for a tenant
	var schedTrends []utils.ScheduledTrend
	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}, TrendIDPrefixes: []string{}}, &schedTrends); err != nil {
		t.Error(err)
	} else if len(schedTrends) != 1 {
		t.Errorf("expected 1 schedTrends, got %d", len(schedTrends))
	}
	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "tenant1"}}, TrendIDPrefixes: []string{}}, &schedTrends); err != nil {
		t.Error(err)
	} else if len(schedTrends) != 1 {
		t.Errorf("expected 1 schedTrends, got %d", len(schedTrends))
	}
	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "tenant2"}}, TrendIDPrefixes: []string{}}, &schedTrends); err != nil {
		t.Error(err)
	} else if len(schedTrends) != 2 {
		t.Errorf("expected 2 schedTrends, got %d", len(schedTrends))
	}

	expTrends := []utils.ScheduledTrend{
		{
			TrendID: "TR_1min",
			Next:    time.Now().Add(1 * time.Minute),
		},
	}
	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org"}}, TrendIDPrefixes: []string{}}, &schedTrends); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(schedTrends, expTrends, cmpopts.EquateApproxTime(1*time.Minute), cmpopts.IgnoreFields(utils.ScheduledTrend{}, "Previous")); diff != utils.EmptyString {
		t.Errorf("unexpected scheduled trends (-want +got)\n%s", diff)
	}

	expTrends = []utils.ScheduledTrend{
		{
			TrendID: "TR_5min",
			Next:    time.Now().Add(5 * time.Minute),
		},
	}
	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "tenant1"}}, TrendIDPrefixes: []string{}}, &schedTrends); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(schedTrends, expTrends, cmpopts.EquateApproxTime(5*time.Minute), cmpopts.IgnoreFields(utils.ScheduledTrend{}, "Previous")); diff != utils.EmptyString {
		t.Errorf("unexpected scheduled trends (-want +got)\n%s", diff)
	}

	// getting scheduled trends by the prefix
	expTrends = []utils.ScheduledTrend{
		{
			TrendID: "Trend_avg",
			Next:    time.Now().Add(10 * time.Minute),
		},
		{
			TrendID: "Trend_avg_30min",
			Next:    time.Now().Add(30 * time.Minute),
		},
	}

	if err := trendAuQRpc.Call(context.Background(), utils.TrendSv1GetScheduledTrends, &utils.ArgScheduledTrends{TenantIDWithAPIOpts: utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "tenant2"}}, TrendIDPrefixes: []string{"Trend_avg"}}, &schedTrends); err != nil {
		t.Error(err)
	} else if diff := cmp.Diff(schedTrends, expTrends, cmpopts.EquateApproxTime(10*time.Minute), cmpopts.IgnoreFields(utils.ScheduledTrend{}, "Previous")); diff != utils.EmptyString {
		t.Errorf("unexpected scheduled trends (-want +got)\n%s", diff)
	}

}

func testtrendAuQStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
