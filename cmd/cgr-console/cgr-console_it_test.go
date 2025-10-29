//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var cnslRPC *birpc.Client

var (
	cnslItCfgPath string
	cnslItDirPath string
	cnslItCfg     *config.CGRConfig
	cnslItTests   = []func(t *testing.T){
		testConsoleItLoadConfig,
		testConsoleItInitDataDB,
		testConsoleItInitStorDB,
		testConsoleItStartEngine,
		testConsoleItRPCConn,
		testConsoleItLoadTP,
		testConsoleItCacheClear,
		testConsoleItDebitMax,
		testConsoleItThreshold,
		testConsoleItThresholdsProfileIds,
		testConsoleItThresholdsProfileSet,
		testConsoleItThresholdsProfile,
		testConsoleItCacheItemExpiryTime,
		testConsoleItThresholdsProcessEvent,
		testConsoleItThresholdsForEvent,
		testConsoleItThresholdsProfileRemove,
		testConsoleItTriggersSet,
		testConsoleItTriggers,
		testConsoleItSchedulerReload,
		testConsoleItSchedulerExecute,
		testConsoleItActionExecute,
		testConsoleItAccountTriggersReset,
		testConsoleItAccountTriggersAdd,
		testConsoleItAccountTriggersRemove,
		testConsoleItAccountTriggersSet,
		testConsoleItRatingProfileSet,
		testConsoleItRatingProfileIds,
		testConsoleItRatingProfile,
		testConsoleItResources,
		testConsoleItResourcesProfileIds,
		testConsoleItResourcesProfile,
		testConsoleItResourcesRelease,
		testConsoleItResourcesProfileSet,
		testConsoleItResourcesForEvent,
		testConsoleItResourcesAllocate,
		testConsoleItResourcesProfileRemove,
		testConsoleItChargersProfile,
		testConsoleItChargersProfileSet,
		testConsoleItChargersForEvent,
		testConsoleItChargersProfileIds,
		testConsoleItChargersProcessEvent,
		testConsoleItChargersProfileRemove,
		testConsoleItResourcesAuthorize,
		testConsoleItRouteProfileIds,
		testConsoleItRoutesProfilesForEvent,
		testConsoleItRoutesProfile,
		testConsoleItRoutes,
		testConsoleItRoutesProfileSet,
		testConsoleItRoutesProfileRemove,
		testConsoleItComputeFilterIndexes,
		testConsoleItComputeActionplanIndexes,
		testConsoleItFilterIndexes,
		testConsoleItCacheReload,
		testConsoleItAttributesForEvent,
		testConsoleItAttributesProcessEvent,
		testConsoleItAttributesProfileIds,
		testConsoleItAttributesProfileSet,
		testConsoleItAttributesProfile,
		testConsoleItAttributesProfileRemove,
		testConsoleItActionPlanSet,
		testConsoleItActionPlanGet,
		testConsoleItActionPlanRemove,
		testConsoleItFilterIds,
		testConsoleItFilterSet,
		testConsoleItFilterIndexesRemove,
		testConsoleItAccountSet,
		testConsoleItCacheHasItem,
		testConsoleItStatsMetrics,
		testConsoleItStatsProfileSet,
		testConsoleItStatsProfile,
		testConsoleItStatsForEvent,
		testConsoleItStatsProfileIds,
		testConsoleItStatsProcessEvent,
		testConsoleItStatsProfileRemove,
		testConsoleItGetJsonSection,
		testConsoleItStatus,
		testConsoleItSharedGroup,
		testConsoleItDatacost,
		testConsoleItCost,
		testConsoleItMaxUsage,
		testConsoleItSessionProcessCdr,
		testConsoleItCostDetails,
		testConsoleItCdrs,
		testConsoleItRatingPlanCost,
		testConsoleItRatingProfileRemove,
		testConsoleItDebit,
		testConsoleItDestinations,
		testConsoleItDestinationSet,
		testConsoleItSetStordbVersions,
		testConsoleItSetDatadbVersions,
		testConsoleItStordbVersions,
		testConsoleItDataDbVersions,
		testConsoleItCacheRemoveItem,
		testConsoleItCacheHasGroup,
		testConsoleItFilter,
		testConsoleItFilterRemove,
		testConsoleItCacheGroupItemIds,
		testConsoleItPing,
		testConsoleItLoadTpFromFolder,
		testConsoleItImportTpFromFolder,
		testConsoleItLoadTpFromStordb,
		testConsoleItAccounts,
		testConsoleItMaxDuration,
		testConsoleItAccountRemove,
		testConsoleItBalanceAdd,
		testconsoleItBalanceSet,
		testConsoleItBalanceRemove,
		testConsoleItBalanceDebit,
		testConsoleItGetLoadTimes,
		testConsoleItGetLoadIds,
		testConsoleItSessionAuthorizeEvent,
		testConsoleItCachePrecacheStatus,
		testConsoleItDispatchersProfileSet,
		testConsoleItDispatchersProfileIds,
		testConsoleItDispatchersProfile,
		testConsoleItDispatchersProfileRemove,
		testConsoleItDispatchersHostSet,
		testConsoleItDispatchersHostIds,
		testConsoleItDispatchersHost,
		testConsoleItDispatchersHostRemove,
		testConsoleItAccountActionPlanGet,
		testConsoleItCacheItemIds,
		testConsoleItCacheItemExpiryTime,
		testConsoleItSessionProcessMessage,
		testConsoleItSessionUpdate,
		testConsoleItSessionInitiate,
		testConsoleItActiveSessions,
		testConsoleItPassiveSessions,
		testConsoleItSleep,
		testConsoleItCacheRemoveGroup,
		testConsoleItParse,
		testConsoleItSchedulerQueue,
		testConsoleItCacheStats,
		testConsoleItReloadConfig,
		testConsoleItKillEngine,
	}
	cnslItDispatchersTests = []func(t *testing.T){
		testConsoleItLoadConfig,
		testConsoleItInitDataDB,
		testConsoleItInitStorDB,
		testConsoleItStartEngine,
		testConsoleItDispatchersLoadTP,
		testConsoleItDispatchesForEvent,
		testConsoleItKillEngine,
	}
	cnslItLoadersTests = []func(t *testing.T){
		testConsoleItLoadersLoadConfig,
		testConsoleItInitDataDB,
		testConsoleItInitStorDB,
		testConsoleItLoadersStartEngine,
		testConsoleItLoadTP,
		testConsoleItLoaderLoad,
		testConsoleItLoaderRemove,
		testConsoleItLoadersKillEngine,
	}
)

func TestConsoleItTests(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		cnslItDirPath = "tutmysql"
	case utils.MetaMongo:
		cnslItDirPath = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown database type")
	}
	for _, test := range cnslItTests {
		t.Run("TestConsoleItTests", test)
	}
}

func TestConsoleItDispatchersTests(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		cnslItDirPath = path.Join("dispatchers", "dispatchers_mysql")
	case utils.MetaMongo:
		cnslItDirPath = path.Join("dispatchers", "dispatchers_mongo")
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown database type")
	}
	for _, test := range cnslItDispatchersTests {
		t.Run("TestConsoleItDispatchersTests", test)
	}
}

func TestConsoleItLoadersTests(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		cnslItDirPath = path.Join("loaders", "loaders_console_mysql")
	case utils.MetaMongo:
		cnslItDirPath = path.Join("loaders", "loaders_console_mongo")
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown database type")
	}
	for _, test := range cnslItLoadersTests {
		t.Run("TestConsoleItLoadersTests", test)
	}
}

func testConsoleItLoadConfig(t *testing.T) {
	var err error
	cnslItCfgPath = path.Join(*utils.DataDir, "conf", "samples", cnslItDirPath)
	if cnslItCfg, err = config.NewCGRConfigFromPath(cnslItCfgPath); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItLoadersLoadConfig(t *testing.T) {
	fldPathIn := "/tmp/In"
	fldPathOut := "/tmp/Out"
	if err := os.MkdirAll(fldPathIn, 0777); err != nil {
		t.Error(err)
	}
	if err := os.MkdirAll(fldPathOut, 0777); err != nil {
		t.Error(err)
	}

	var err error
	cnslItCfgPath = path.Join(*utils.DataDir, "conf", "samples", cnslItDirPath)
	if cnslItCfg, err = config.NewCGRConfigFromPath(cnslItCfgPath); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(cnslItCfg); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItInitStorDB(t *testing.T) {
	if err := engine.InitStorDb(cnslItCfg); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(cnslItCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func newRPCClient(cfg *config.ListenCfg) (c *birpc.Client, err error) {
	switch *utils.Encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

// make rpc
func testConsoleItRPCConn(t *testing.T) {
	var err error
	if cnslRPC, err = newRPCClient(cnslItCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItLoadersStartEngine(t *testing.T) {
	fldPathIn := "/tmp/In"
	fldPathOut := "/tmp/Out"
	if err := os.MkdirAll(fldPathIn, 0777); err != nil {
		t.Error(err)
	}
	if err := os.MkdirAll(fldPathOut, 0777); err != nil {
		t.Error(err)
	}
	if _, err := engine.StartEngine(cnslItCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItLoadTP(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+cnslItCfgPath, "-path="+path.Join(*utils.DataDir, "tariffplans", "tutorial"))
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
}

func testConsoleItDispatchersLoadTP(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+cnslItCfgPath, "-path="+path.Join(*utils.DataDir, "tariffplans", "dispatchers"), `-caches_address=`, `-scheduler_address=`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
}

func testConsoleItCacheClear(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_clear")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItThresholdsProfileIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "thresholds_profile_ids", `Tenant="cgrates.org"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []string{"THD_ACNT_1001", "THD_ACNT_1002"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Strings(rcv)
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}

}

func testConsoleItResourcesProfileIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "resources_profile_ids", "Tenant", "cgrates.org")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []string{"ResGroup1"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Strings(rcv)
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
	// fmt.Println(output.String())
}

func testConsoleItRatingProfileSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "ratingprofile_set", `Tenant="cgrates.org"`, `ID="123"`, `Subject="1001"`, `RatingPlanActivations=[{"ActivationTime":"2012-01-01T00:00:00Z", "RatingPlanId":"RP_1001", "FallbackSubjects":"dan2"}]`)
	output := bytes.NewBuffer(nil)
	expected := "OK"
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %s \n but received \n %s", expected, rcv)
	}
}

func testConsoleItRouteProfileIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "route_profile_ids", "Tenant", "cgrates.org")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []string{"ROUTE_ACNT_1001", "ROUTE_ACNT_1002", "ROUTE_ACNT_1003"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Fatal(err)
	}
	sort.Strings(rcv)
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItCacheReload(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_reload", "Tenant", "cgrates.org")
	output := bytes.NewBuffer(nil)
	expected := "OK"
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %s \n but received \n %s", expected, rcv)
	}
}

func testConsoleItFilterIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "filter_ids", "Tenant", "cgrates.org")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []string{"FLTR_ACNT_1001", "FLTR_ACNT_1001_1002", "FLTR_ACNT_1002", "FLTR_ACNT_1003", "FLTR_ACNT_1003_1001", "FLTR_DST_FS", "FLTR_RES"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Fatal(err)
	}
	sort.Strings(rcv)
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItCacheHasItem(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_has_item", "Tenant", "cgrates.org")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := false
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv bool
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", expected, rcv)
	}
}

func testConsoleItStatsMetrics(t *testing.T) {
	cmd := exec.Command("cgr-console", "stats_metrics", `ID="Stats2"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"*tcc": "N/A",
		"*tcd": "N/A",
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItGetJsonSection(t *testing.T) {
	cmd := exec.Command("cgr-console", "get_json_section", `Section="cores"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"cores": map[string]any{
			"caps":                0.,
			"caps_stats_interval": "0",
			"caps_strategy":       "*busy",
			"shutdown_timeout":    "1s",
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItResourcesAuthorize(t *testing.T) {
	cmd := exec.Command("cgr-console", "resources_authorize", `Tenant="cgrates.org"`, `ID="123"`, `Event={"Account":"1001"}`, `UsageID="usageID"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "123"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItStatsProfileSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "stats_profile_set", `Tenant="cgrates.org"`, `ID="123"`)
	output := bytes.NewBuffer(nil)
	expected := "OK"
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %s \n but received \n %s", expected, rcv)
	}
}

func testConsoleItResourcesRelease(t *testing.T) {
	cmd := exec.Command("cgr-console", "resources_release", `Tenant="cgrates.org"`, `ID="123"`, `Event={"Account":"1001"}`, `UsageID="usageID"`)
	output := bytes.NewBuffer(nil)
	expected := "OK"
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %s \n but received \n %s", expected, rcv)
	}
}

func testConsoleItRoutesProfilesForEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "routes_profiles_for_event", `ID="123"`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"Tenant":    "cgrates.org",
			"ID":        "ROUTE_ACNT_1001",
			"FilterIDs": []any{"FLTR_ACNT_1001"},
			"ActivationInterval": map[string]any{
				"ActivationTime": "2017-11-27T00:00:00Z",
				"ExpiryTime":     "0001-01-01T00:00:00Z",
			},
			"Sorting":           "*weight",
			"SortingParameters": []any{},
			"Routes": []any{
				map[string]any{
					"ID":              "route1",
					"FilterIDs":       nil,
					"AccountIDs":      nil,
					"RatingPlanIDs":   nil,
					"ResourceIDs":     nil,
					"StatIDs":         nil,
					"Weight":          10.,
					"Blocker":         false,
					"RouteParameters": "",
				},
				map[string]any{
					"ID":              "route2",
					"FilterIDs":       nil,
					"AccountIDs":      nil,
					"RatingPlanIDs":   nil,
					"ResourceIDs":     nil,
					"StatIDs":         nil,
					"Weight":          20.,
					"Blocker":         false,
					"RouteParameters": "",
				},
			},
			"Weight": 20.,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	sort.Slice(rcv[0].(map[string]any)["Routes"], func(i, j int) bool {
		return utils.IfaceAsString(rcv[0].(map[string]any)["Routes"].([]any)[i].(map[string]any)["ID"]) < utils.IfaceAsString(rcv[0].(map[string]any)["Routes"].([]any)[j].(map[string]any)["ID"])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %s \n but received \n %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItStatsProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "stats_profile", `Tenant="cgrates.org"`, `ID="Stats2"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"ActivationInterval": map[string]any{
			"ActivationTime": "2014-07-29T15:00:00Z",
			"ExpiryTime":     "0001-01-01T00:00:00Z",
		},
		"Blocker":   true,
		"FilterIDs": []any{"FLTR_ACNT_1001_1002"},
		"ID":        "Stats2",
		"Metrics": []any{
			map[string]any{
				"FilterIDs": nil,
				"MetricID":  "*tcc",
			},
			map[string]any{
				"FilterIDs": nil,
				"MetricID":  "*tcd",
			},
		},
		"MinItems":     0.,
		"QueueLength":  100.,
		"Stored":       false,
		"TTL":          "-1ns",
		"Tenant":       "cgrates.org",
		"ThresholdIDs": []any{"*none"},
		"Weight":       30.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	sort.Slice(rcv["Metrics"].([]any), func(i, j int) bool {
		return utils.IfaceAsString((rcv["Metrics"].([]any)[i].(map[string]any))["MetricID"]) < utils.IfaceAsString((rcv["Metrics"].([]any)[j].(map[string]any))["MetricID"])
	})

	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItRoutesProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "routes_profile", `Tenant="cgrates.org"`, `ID="ROUTE_ACNT_1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected :=
		map[string]any{
			"Tenant":    "cgrates.org",
			"ID":        "ROUTE_ACNT_1001",
			"FilterIDs": []any{"FLTR_ACNT_1001"},
			"ActivationInterval": map[string]any{
				"ActivationTime": "2017-11-27T00:00:00Z",
				"ExpiryTime":     "0001-01-01T00:00:00Z",
			},
			"Sorting":           "*weight",
			"SortingParameters": []any{},
			"Routes": []any{
				map[string]any{
					"ID":              "route1",
					"FilterIDs":       nil,
					"AccountIDs":      nil,
					"RatingPlanIDs":   nil,
					"ResourceIDs":     nil,
					"StatIDs":         nil,
					"Weight":          10.,
					"Blocker":         false,
					"RouteParameters": "",
				},
				map[string]any{
					"ID":              "route2",
					"FilterIDs":       nil,
					"AccountIDs":      nil,
					"RatingPlanIDs":   nil,
					"ResourceIDs":     nil,
					"StatIDs":         nil,
					"Weight":          20.,
					"Blocker":         false,
					"RouteParameters": "",
				},
			},
			"Weight": 20.,
		}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Error(err)
	}
	sort.Slice(rcv["Routes"], func(i, j int) bool {
		return utils.IfaceAsString(rcv["Routes"].([]any)[i].(map[string]any)["ID"]) < utils.IfaceAsString(rcv["Routes"].([]any)[j].(map[string]any)["ID"])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %s \n but received \n %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

/* Snooze is different everytime, it uses current time */
func testConsoleItThreshold(t *testing.T) {
	cmd := exec.Command("cgr-console", "threshold", `Tenant="cgrates.org"`, `ID="THD_ACNT_1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Hits":   0.,
		"ID":     "THD_ACNT_1001",
		"Snooze": "0001-01-01T00:00:00Z",
		"Tenant": "cgrates.org",
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %s \n but received \n %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItThresholdsProfileSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "thresholds_profile_set", `Tenant="cgrates.org"`, `ID="123"`)
	output := bytes.NewBuffer(nil)
	expected := "OK"
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %s \n but received \n %s", expected, rcv)
	}
}

func testConsoleItThresholdsProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "thresholds_profile", `Tenant="cgrates.org"`, `ID="THD_ACNT_1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"ActionIDs": []any{"ACT_LOG_WARNING"},
		"ActivationInterval": map[string]any{
			"ActivationTime": "2014-07-29T15:00:00Z",
			"ExpiryTime":     "0001-01-01T00:00:00Z",
		},
		"Async":     true,
		"Blocker":   false,
		"FilterIDs": []any{"FLTR_ACNT_1001"},
		"ID":        "THD_ACNT_1001",
		"MaxHits":   1.,
		"MinHits":   1.,
		"MinSleep":  "1s",
		"Tenant":    "cgrates.org",
		"Weight":    10.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", expected, rcv)
	}
}

func testConsoleItRatingProfileIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "ratingprofile_ids", `Tenant="cgrates.org"`)
	output := bytes.NewBuffer(nil)
	expected := []any{":1001", "call:1001", "call:1002", "call:1003", "mms:*any", "sms:*any"}
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Error(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(string) < rcv[j].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", expected, rcv)
	}
}

func testConsoleItStatsProfileIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "stats_profile_ids", `Tenant="cgrates.org"`)
	output := bytes.NewBuffer(nil)
	expected := []any{"123", "Stats2", "Stats2_1"}
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(string) < rcv[j].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", expected, rcv)
	}
}

func testConsoleItStatus(t *testing.T) {
	cmd := exec.Command("cgr-console", "status")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
}

func testConsoleItCacheStats(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_stats")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"*account_action_plans": map[string]any{
			"Items":  1.,
			"Groups": 0.,
		},
		"*accounts": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*action_plans": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*action_triggers": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*actions": map[string]any{
			"Groups": 0.,
			"Items":  1.,
		},
		"*apiban": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*attribute_filter_indexes": map[string]any{
			"Items":  10.,
			"Groups": 2.,
		},
		"*attribute_profiles": map[string]any{
			"Items":  1.,
			"Groups": 0.,
		},
		"*caps_events": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*cdr_ids": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*cdrs": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*charger_filter_indexes": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*charger_profiles": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*closed_sessions": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*default": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*destinations": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*diameter_messages": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_filter_indexes": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_hosts": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_loads": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_profiles": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_routes": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatchers": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*event_charges": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*event_resources": map[string]any{
			"Items":  1.,
			"Groups": 0.,
		},
		"*filters": map[string]any{
			"Items":  4.,
			"Groups": 0.,
		},
		"*load_ids": map[string]any{
			"Items":  13.,
			"Groups": 0.,
		},
		"*rating_plans": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*rating_profiles": map[string]any{
			"Items":  1.,
			"Groups": 0.,
		},
		"*replication_hosts": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*resource_filter_indexes": map[string]any{
			"Items":  2.,
			"Groups": 1.,
		},
		"*resource_profiles": map[string]any{
			"Items":  2.,
			"Groups": 0.,
		},
		"*resources": map[string]any{
			"Items":  2.,
			"Groups": 0.,
		},
		"*reverse_destinations": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*reverse_filter_indexes": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*route_filter_indexes": map[string]any{
			"Items":  3.,
			"Groups": 1.,
		},
		"*route_profiles": map[string]any{
			"Items":  1.,
			"Groups": 0.,
		},
		"*rpc_connections": map[string]any{
			"Items":  3.,
			"Groups": 0.,
		},
		"*rpc_responses": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*session_costs": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*shared_groups": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*stat_filter_indexes": map[string]any{
			"Items":  2.,
			"Groups": 1.,
		},
		"*statqueue_profiles": map[string]any{
			"Items":  2.,
			"Groups": 0.,
		},
		"*statqueues": map[string]any{
			"Items":  2.,
			"Groups": 0.,
		},
		"*stir": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*threshold_filter_indexes": map[string]any{
			"Items":  9.,
			"Groups": 1.,
		},
		"*threshold_profiles": map[string]any{
			"Items":  2.,
			"Groups": 0.,
		},
		"*thresholds": map[string]any{
			"Items":  2.,
			"Groups": 0.,
		},
		"*timings": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tmp_rating_profiles": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_account_actions": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_action_plans": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_action_triggers": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_actions": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_attributes": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_chargers": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_destination_rates": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_destinations": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_dispatcher_hosts": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_dispatcher_profiles": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_filters": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_rates": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_rating_plans": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_rating_profiles": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_resources": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_ips": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_routes": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_shared_groups": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_stats": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_thresholds": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_timings": map[string]any{
			"Groups": 0.,
			"Items":  0.,
		},
		"*uch": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
		"*versions": map[string]any{
			"Items":  0.,
			"Groups": 0.,
		},
	}

	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItResourcesProfileSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "resources_profile_set", `ID="123"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItResourcesAllocate(t *testing.T) {
	cmd := exec.Command("cgr-console", "resources_allocate", `Tenant="cgrates.org"`, `ID="123"`, `Event={"Account":"1001"}`, `UsageID="usageID"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "ResGroup1"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItResourcesForEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "resources_for_event", `Tenant="cgrates.org"`, `ID="123"`, `Event={"Account":"1001"}`, `UsageID="usageID"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"Tenant": "cgrates.org",
			"ID":     "ResGroup1",
			"Usages": map[string]any{},
			"TTLIdx": nil,
		},
		map[string]any{
			"Tenant": "cgrates.org",
			"ID":     "123",
			"Usages": map[string]any{},
			"TTLIdx": nil,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItAttributesProfileIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "attributes_profile_ids", `Tenant="cgrates.org"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{"ATTR_1001_SESSIONAUTH", "ATTR_1001_SIMPLEAUTH", "ATTR_1002_SESSIONAUTH", "ATTR_1002_SIMPLEAUTH", "ATTR_1003_SESSIONAUTH", "ATTR_1003_SIMPLEAUTH", "ATTR_ACC_ALIAS"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(string) < rcv[j].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItThresholdsProcessEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "thresholds_process_event", `ID="123"`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{"123", "THD_ACNT_1001"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(string) < rcv[j].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItCacheRemoveItem(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_remove_item")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItFilterSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "filter_set", `ID="123"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItResources(t *testing.T) {
	cmd := exec.Command("cgr-console", "resources", `ID="ResGroup1"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Tenant": "cgrates.org",
		"ID":     "ResGroup1",
		"Usages": map[string]any{},
		"TTLIdx": nil,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItResourcesProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "resources_profile", `ID="ResGroup1"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"ActivationInterval": map[string]any{
			"ActivationTime": "2014-07-29T15:00:00Z",
			"ExpiryTime":     "0001-01-01T00:00:00Z",
		},
		"AllocationMessage": "",
		"Blocker":           false,
		"FilterIDs":         []any{"FLTR_RES"},
		"ID":                "ResGroup1",
		"Limit":             7.,
		"Stored":            true,
		"Tenant":            "cgrates.org",
		"ThresholdIDs":      []any{"*none"},
		"UsageTTL":          "-1ns",
		"Weight":            10.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", expected, rcv)
	}
}

func testConsoleItAccountSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "account_set", `Tenant="cgrates.org"`, `Account="1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItRoutes(t *testing.T) {
	cmd := exec.Command("cgr-console", "routes", `ID="ROUTE_ACNT_1001"`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"ProfileID": "ROUTE_ACNT_1001",
			"Sorting":   "*weight",
			"Routes": []any{
				map[string]any{
					"RouteID":         "route2",
					"RouteParameters": "",
					"SortingData": map[string]any{
						"Weight": 20.,
					},
				},
				map[string]any{
					"RouteID":         "route1",
					"RouteParameters": "",
					"SortingData": map[string]any{
						"Weight": 10.,
					},
				},
			},
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		fmt.Println(utils.IfaceAsString((rcv[0].(map[string]any)["Routes"].([]any)[i].(map[string]any)["RouteID"])))
		return utils.IfaceAsString((rcv[0].(map[string]any)["Routes"].([]any)[i].(map[string]any)["RouteID"])) < utils.IfaceAsString((rcv[0].(map[string]any)["Routes"].([]any)[j].(map[string]any)["RouteID"]))
		// return utils.IfaceAsString((rcv["Metrics"].([]any)[i].(map[string]any))["MetricID"]) < utils.IfaceAsString((rcv["Metrics"].([]any)[j].(map[string]any))["MetricID"])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItFilter(t *testing.T) {
	cmd := exec.Command("cgr-console", "filter", `ID="FLTR_RES"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"ActivationInterval": map[string]any{
			"ActivationTime": "2014-07-29T15:00:00Z",
			"ExpiryTime":     "0001-01-01T00:00:00Z",
		},
		"Tenant": "cgrates.org",
		"ID":     "FLTR_RES",
		"Rules": []any{
			map[string]any{
				"Type":    "*string",
				"Element": "~*req.Account",
				"Values":  []any{"1001", "1002", "1003"},
			},
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

/* Snooze is different everytime, it uses current time */
func testConsoleItThresholdsForEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "thresholds_for_event", `Tenant="cgrates.org"`, `ID="123"`, `Event={"Account":"1001"}`, `UsageID="usageID"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"Tenant": "cgrates.org",
			"ID":     "123",
			"Hits":   1.,
			"Snooze": "",
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv[0].(map[string]any)["Snooze"] = ""

	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItStatsForEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "stats_for_event", `Tenant="cgrates.org"`, `ID="Stats2"`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{"123"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItStatsProcessEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "stats_process_event", `Tenant="cgrates.org"`, `ID="123"`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{"123"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItReloadConfig(t *testing.T) {
	cmd := exec.Command("cgr-console", "reload_config", `Path="/usr/share/cgrates/conf/samples/tutmongo"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItAttributesProfileSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "attributes_profile_set", `Tenant="cgrates.org"`, `ID="attrID"`, `Attributes=[{"Path":"*req.Account", "Value":"1001"}]`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItPing(t *testing.T) {
	cmd := exec.Command("cgr-console", "ping", "attributes")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "Pong"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSessionUpdate(t *testing.T) {
	cmd := exec.Command("cgr-console", "session_update", `GetAttributes=true`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Attributes": map[string]any{
			"APIOpts": map[string]any{
				"*subsys": "*sessions",
			},
			"AlteredFields": []any{"*req.LCRProfile", "*req.Password", "*req.PaypalAccount", "*req.RequestType"},
			"Event": map[string]any{
				"Account":       "1001",
				"LCRProfile":    "premium_cli",
				"Password":      "CGRateS.org",
				"PaypalAccount": "cgrates@paypal.com",
				"RequestType":   "*prepaid",
			},
			"ID":              nil,
			"MatchedProfiles": []any{"ATTR_1001_SESSIONAUTH"},
			"Tenant":          "cgrates.org",
			"Time":            nil,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv["Attributes"].(map[string]any)["ID"] = nil
	rcv["Attributes"].(map[string]any)["Time"] = nil
	sort.Slice(rcv["Attributes"].(map[string]any)["AlteredFields"], func(i, j int) bool {
		return utils.IfaceAsString(rcv["Attributes"].(map[string]any)["AlteredFields"].([]any)[i]) < utils.IfaceAsString(rcv["Attributes"].(map[string]any)["AlteredFields"].([]any)[j])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItLoadTpFromFolder(t *testing.T) {
	cmd := exec.Command("cgr-console", "load_tp_from_folder", `FolderPath="/usr/share/cgrates/tariffplans/tutorial"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSessionAuthorizeEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "session_authorize_event", `GetAttributes=true`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"AttributesDigest":   "LCRProfile:premium_cli,Password:CGRateS.org,PaypalAccount:cgrates@paypal.com,RequestType:*prepaid",
		"ResourceAllocation": nil,
		"MaxUsage":           0.,
		"RoutesDigest":       nil,
		"Thresholds":         nil,
		"StatQueues":         nil,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	s := strings.Split(rcv["AttributesDigest"].(string), ",")
	sort.Strings(s)
	rcv["AttributesDigest"] = strings.Join(s, ",")
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItCacheRemoveGroup(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_remove_group", `Tenant="cgrates.org"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItChargersProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "chargers_profile", `ID="DEFAULT"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Tenant":             "cgrates.org",
		"ID":                 "DEFAULT",
		"FilterIDs":          []any{},
		"ActivationInterval": nil,
		"RunID":              "*default",
		"AttributeIDs":       []any{"*none"},
		"Weight":             0.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItComputeFilterIndexes(t *testing.T) {
	cmd := exec.Command("cgr-console", "compute_filter_indexes", `AttributeS=true`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItStordbVersions(t *testing.T) {
	cmd := exec.Command("cgr-console", "stordb_versions")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Accounts":            3.,
		"ActionPlans":         3.,
		"ActionTriggers":      2.,
		"Actions":             2.,
		"Attributes":          6.,
		"CDRs":                2.,
		"Chargers":            2.,
		"CostDetails":         2.,
		"Destinations":        1.,
		"Dispatchers":         2.,
		"LoadIDs":             1.,
		"RQF":                 5.,
		"RatingPlan":          1.,
		"RatingProfile":       1.,
		"Resource":            1.,
		"ReverseDestinations": 1.,
		"Routes":              2.,
		"SessionSCosts":       3.,
		"SharedGroups":        2.,
		"Stats":               4.,
		"Subscribers":         1.,
		"Thresholds":          4.,
		"Timing":              1.,
		"TpAccountActions":    1.,
		"TpActionPlans":       1.,
		"TpActionTriggers":    1.,
		"TpActions":           1.,
		"TpChargers":          1.,
		"TpDestinationRates":  1.,
		"TpDestinations":      1.,
		"TpDispatchers":       1.,
		"TpFilters":           1.,
		"TpRates":             1.,
		"TpRatingPlan":        1.,
		"TpRatingPlans":       1.,
		"TpRatingProfile":     1.,
		"TpRatingProfiles":    1.,
		"TpResource":          1.,
		"TpResources":         1.,
		"TpRoutes":            1.,
		"TpSharedGroups":      1.,
		"TpStats":             1.,
		"TpThresholds":        1.,
		"TpTiming":            1.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItImportTpFromFolder(t *testing.T) {
	cmd := exec.Command("cgr-console", "import_tp_from_folder", `FolderPath="/usr/share/cgrates/tariffplans/tutorial"`, `TPid="1"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItGetLoadIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "get_load_ids")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	expected := map[string]any{
		"*account_action_plans":     rcv["*account_action_plans"],
		"*action_plans":             rcv["*action_plans"],
		"*action_triggers":          rcv["*action_triggers"],
		"*actions":                  rcv["*actions"],
		"*attribute_filter_indexes": rcv["*attribute_filter_indexes"],
		"*attribute_profiles":       rcv["*attribute_profiles"],
		"*charger_profiles":         rcv["*charger_profiles"],
		"*destinations":             rcv["*destinations"],
		"*filters":                  rcv["*filters"],
		"*rating_plans":             rcv["*rating_plans"],
		"*rating_profiles":          rcv["*rating_profiles"],
		"*resource_profiles":        rcv["*resource_profiles"],
		"*resources":                rcv["*resources"],
		"*reverse_destinations":     rcv["*reverse_destinations"],
		"*route_profiles":           rcv["*route_profiles"],
		"*statqueue_profiles":       rcv["*statqueue_profiles"],
		"*statqueues":               rcv["*statqueues"],
		"*threshold_profiles":       rcv["*threshold_profiles"],
		"*thresholds":               rcv["*thresholds"],
		"*timings":                  rcv["*timings"],
		"test":                      rcv["test"],
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItChargersForEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "chargers_for_event")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"Tenant":             "cgrates.org",
			"ID":                 "DEFAULT",
			"FilterIDs":          []any{},
			"ActivationInterval": nil,
			"RunID":              "*default",
			"AttributeIDs":       []any{"*none"},
			"Weight":             0.,
		},
		map[string]any{
			"Tenant":             "cgrates.org",
			"ID":                 "Raw",
			"FilterIDs":          []any{},
			"ActivationInterval": nil,
			"RunID":              "*raw",
			"AttributeIDs":       []any{"*constant:*req.RequestType:*none"},
			"Weight":             0.,
		},
		map[string]any{
			"ActivationInterval": nil,
			"AttributeIDs":       nil,
			"FilterIDs":          nil,
			"ID":                 "cps",
			"RunID":              "",
			"Tenant":             "cgrates.org",
			"Weight":             0.,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return utils.IfaceAsString(rcv[i].(map[string]any)["ID"]) < utils.IfaceAsString(rcv[j].(map[string]any)["ID"])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItAccounts(t *testing.T) {
	cmd := exec.Command("cgr-console", "accounts", `AccountIDs=["1001"]`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"ActionTriggers": nil,
			"AllowNegative":  false,
			"BalanceMap": map[string]any{
				"*monetary": []any{
					map[string]any{
						"Blocker":        false,
						"Categories":     map[string]any{},
						"DestinationIDs": map[string]any{},
						"Disabled":       false,
						"ExpirationDate": "0001-01-01T00:00:00Z",
						"Factors":        nil,
						"ID":             "test",
						"RatingSubject":  "",
						"SharedGroups":   map[string]any{},
						"TimingIDs":      map[string]any{},
						"Timings":        nil,
						"Uuid":           "",
						"Value":          10.,
						"Weight":         10.,
					},
				},
			},
			"Disabled":     false,
			"ID":           "cgrates.org:1001",
			"UnitCounters": nil,
			"UpdateTime":   "",
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv[0].(map[string]any)["BalanceMap"].(map[string]any)["*monetary"].([]any)[0].(map[string]any)["Uuid"] = ""
	rcv[0].(map[string]any)["UpdateTime"] = ""
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItAccountRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "account_remove", `Account="1002"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItCacheHasGroup(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_has_group")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := false
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv bool
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", expected, rcv)
	}
}

func testConsoleItDataDbVersions(t *testing.T) {
	cmd := exec.Command("cgr-console", "datadb_versions")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Accounts":            3.,
		"ActionPlans":         3.,
		"ActionTriggers":      2.,
		"Actions":             2.,
		"Attributes":          6.,
		"Chargers":            2.,
		"Destinations":        1.,
		"Dispatchers":         2.,
		"LoadIDs":             1.,
		"RQF":                 5.,
		"RatingPlan":          1.,
		"RatingProfile":       1.,
		"Resource":            1.,
		"ReverseDestinations": 1.,
		"Routes":              2.,
		"SharedGroups":        2.,
		"Stats":               4.,
		"Subscribers":         1.,
		"Thresholds":          4.,
		"Timing":              1.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSessionInitiate(t *testing.T) {
	cmd := exec.Command("cgr-console", "session_initiate", `InitSession=true`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"AttributesDigest":   nil,
		"MaxUsage":           "10.8µs",
		"ResourceAllocation": nil,
		"StatQueues":         nil,
		"Thresholds":         nil,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}

	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	// s := strings.Split(rcv["AttributesDigest"].(string), ",")
	// sort.Strings(s)
	// rcv["AttributesDigest"] = strings.Join(s, ",")
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItCachePrecacheStatus(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_precache_status")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"*account_action_plans":      "*ready",
		"*accounts":                  "*ready",
		"*action_plans":              "*ready",
		"*action_triggers":           "*ready",
		"*actions":                   "*ready",
		"*apiban":                    "*ready",
		"*attribute_filter_indexes":  "*ready",
		"*attribute_profiles":        "*ready",
		"*caps_events":               "*ready",
		"*cdr_ids":                   "*ready",
		"*cdrs":                      "*ready",
		"*charger_filter_indexes":    "*ready",
		"*charger_profiles":          "*ready",
		"*closed_sessions":           "*ready",
		"*destinations":              "*ready",
		"*diameter_messages":         "*ready",
		"*dispatcher_filter_indexes": "*ready",
		"*dispatcher_hosts":          "*ready",
		"*dispatcher_loads":          "*ready",
		"*dispatcher_profiles":       "*ready",
		"*dispatcher_routes":         "*ready",
		"*dispatchers":               "*ready",
		"*event_charges":             "*ready",
		"*event_resources":           "*ready",
		"*filters":                   "*ready",
		"*load_ids":                  "*ready",
		"*rating_plans":              "*ready",
		"*rating_profiles":           "*ready",
		"*replication_hosts":         "*ready",
		"*resource_filter_indexes":   "*ready",
		"*resource_profiles":         "*ready",
		"*resources":                 "*ready",
		"*reverse_destinations":      "*ready",
		"*reverse_filter_indexes":    "*ready",
		"*route_filter_indexes":      "*ready",
		"*route_profiles":            "*ready",
		"*rpc_connections":           "*ready",
		"*rpc_responses":             "*ready",
		"*session_costs":             "*ready",
		"*shared_groups":             "*ready",
		"*stat_filter_indexes":       "*ready",
		"*statqueue_profiles":        "*ready",
		"*statqueues":                "*ready",
		"*stir":                      "*ready",
		"*threshold_filter_indexes":  "*ready",
		"*threshold_profiles":        "*ready",
		"*thresholds":                "*ready",
		"*timings":                   "*ready",
		"*tmp_rating_profiles":       "*ready",
		"*tp_account_actions":        "*ready",
		"*tp_action_plans":           "*ready",
		"*tp_action_triggers":        "*ready",
		"*tp_actions":                "*ready",
		"*tp_attributes":             "*ready",
		"*tp_chargers":               "*ready",
		"*tp_destination_rates":      "*ready",
		"*tp_destinations":           "*ready",
		"*tp_dispatcher_hosts":       "*ready",
		"*tp_dispatcher_profiles":    "*ready",
		"*tp_filters":                "*ready",
		"*tp_rates":                  "*ready",
		"*tp_rating_plans":           "*ready",
		"*tp_rating_profiles":        "*ready",
		"*tp_resources":              "*ready",
		"*tp_ips":                    "*ready",
		"*tp_routes":                 "*ready",
		"*tp_shared_groups":          "*ready",
		"*tp_stats":                  "*ready",
		"*tp_thresholds":             "*ready",
		"*tp_timings":                "*ready",
		"*uch":                       "*ready",
		"*versions":                  "*ready",
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItChargersProfileIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "chargers_profile_ids")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{"DEFAULT", "Raw", "cps"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(string) < rcv[j].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItChargersProcessEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "chargers_process_event", `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"ChargerSProfile":    "DEFAULT",
			"AttributeSProfiles": nil,
			"AlteredFields":      []any{"*req.RunID"},
			"CGREvent": map[string]any{
				"APIOpts": map[string]any{
					"*subsys": "*chargers",
				},
				"Tenant": "cgrates.org",
				"ID":     "",
				"Time":   "",
				"Event": map[string]any{
					"Account": "1001",
					"RunID":   "*default",
				},
			},
		},
		map[string]any{
			"ChargerSProfile":    "Raw",
			"AttributeSProfiles": []any{"*constant:*req.RequestType:*none"},
			"AlteredFields":      []any{"*req.RunID", "*req.RequestType"},
			"CGREvent": map[string]any{
				"APIOpts": map[string]any{
					"*subsys": "*chargers",
				},
				"Tenant": "cgrates.org",
				"ID":     "",
				"Time":   "",
				"Event": map[string]any{
					"Account":     "1001",
					"RequestType": "*none",
					"RunID":       "*raw",
				},
			},
		},
		map[string]any{
			"ChargerSProfile":    "cps",
			"AlteredFields":      []any{"*req.RunID"},
			"AttributeSProfiles": []any{},
			"CGREvent": map[string]any{
				"APIOpts": map[string]any{
					"*subsys": "*chargers",
				},
				"Event": map[string]any{
					"Account": "1001",
					"RunID":   "",
				},
				"ID":     "",
				"Tenant": "cgrates.org",
				"Time":   "",
			},
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return utils.IfaceAsString(rcv[i].(map[string]any)["ChargerSProfile"]) < utils.IfaceAsString(rcv[j].(map[string]any)["ChargerSProfile"])
	})
	rcv[0].(map[string]any)["CGREvent"].(map[string]any)["Time"] = ""
	rcv[1].(map[string]any)["CGREvent"].(map[string]any)["Time"] = ""
	rcv[2].(map[string]any)["CGREvent"].(map[string]any)["Time"] = ""
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItSessionProcessMessage(t *testing.T) {
	cmd := exec.Command("cgr-console", "session_process_message", `GetAttributes=true`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Attributes": map[string]any{
			"APIOpts": map[string]any{
				"*subsys": "*sessions",
			},
			"AlteredFields": []any{"*req.LCRProfile", "*req.Password", "*req.PaypalAccount", "*req.RequestType"},
			"Event": map[string]any{
				"Account":       "1001",
				"LCRProfile":    "premium_cli",
				"Password":      "CGRateS.org",
				"PaypalAccount": "cgrates@paypal.com",
				"RequestType":   "*prepaid",
			},
			"ID":              "",
			"MatchedProfiles": []any{"ATTR_1001_SESSIONAUTH"},
			"Tenant":          "cgrates.org",
			"Time":            "",
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv["Attributes"].(map[string]any)["AlteredFields"], func(i, j int) bool {
		return utils.IfaceAsString(rcv["Attributes"].(map[string]any)["AlteredFields"].([]any)[i]) < utils.IfaceAsString(rcv["Attributes"].(map[string]any)["AlteredFields"].([]any)[j])
	})
	rcv["Attributes"].(map[string]any)["ID"] = ""
	rcv["Attributes"].(map[string]any)["Time"] = ""
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItTriggersSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "triggers_set", `GroupID="123"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItTriggers(t *testing.T) {
	cmd := exec.Command("cgr-console", "triggers")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"ActionsID":      "",
			"ActivationDate": "0001-01-01T00:00:00Z",
			"Balance": map[string]any{
				"Blocker":        nil,
				"Categories":     nil,
				"DestinationIDs": nil,
				"Disabled":       nil,
				"ExpirationDate": nil,
				"Factors":        nil,
				"ID":             nil,
				"RatingSubject":  nil,
				"SharedGroups":   nil,
				"TimingIDs":      nil,
				"Timings":        nil,
				"Type":           nil,
				"Uuid":           nil,
				"Value":          nil,
				"Weight":         nil,
			},
			"Executed":          false,
			"ExpirationDate":    "0001-01-01T00:00:00Z",
			"ID":                "123",
			"LastExecutionTime": "0001-01-01T00:00:00Z",
			"MinQueuedItems":    0.,
			"MinSleep":          "0s",
			"Recurrent":         false,
			"ThresholdType":     "",
			"ThresholdValue":    0.,
			"UniqueID":          "",
			"Weight":            0.,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv[0].(map[string]any)["UniqueID"] = ""
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItRatingProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "ratingprofile", `Category="call"`, `Subject="1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Id": "*out:cgrates.org:call:1001",
		"RatingPlanActivations": []any{
			map[string]any{
				"ActivationTime": "2014-01-14T00:00:00Z",
				"RatingPlanId":   "RP_1001",
				"FallbackKeys":   nil,
			},
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItAccountTriggersReset(t *testing.T) {
	cmd := exec.Command("cgr-console", "account_triggers_reset", `Account="1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItBalanceAdd(t *testing.T) {
	cmd := exec.Command("cgr-console", "balance_add", `Account="1001"`, `Value=12`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItRoutesProfileSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "routes_profile_set", `ID="rps"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSleep(t *testing.T) {
	cmd := exec.Command("cgr-console", "sleep")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItActionPlanGet(t *testing.T) {
	cmd := exec.Command("cgr-console", "actionplan_get")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"Id":         "AP_PACKAGE_10",
			"AccountIDs": map[string]any{},
			"ActionTimings": []any{
				map[string]any{
					"Uuid": "",
					"Timing": map[string]any{
						"Timing": map[string]any{
							"ID":        "*asap",
							"Years":     []any{},
							"Months":    []any{},
							"MonthDays": []any{},
							"WeekDays":  []any{},
							"StartTime": "*asap",
							"EndTime":   "",
						},
						"Rating": nil,
						"Weight": 0.,
					},
					"ActionsID": "ACT_TOPUP_RST_10",
					"ExtraData": nil,
					"Weight":    10.,
				},
			},
		},
		map[string]any{
			"Id":         "AP_TEST",
			"AccountIDs": nil,
			"ActionTimings": []any{
				map[string]any{
					"Uuid": "",
					"Timing": map[string]any{
						"Timing": map[string]any{
							"ID":        "",
							"Years":     []any{},
							"Months":    []any{},
							"MonthDays": []any{1.},
							"WeekDays":  []any{},
							"StartTime": "00:00:00",
							"EndTime":   "",
						},
						"Rating": nil,
						"Weight": 0.,
					},
					"ActionsID": "ACT_TOPUP_RST_10",
					"ExtraData": nil,
					"Weight":    20.,
				},
			},
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv[0].(map[string]any)["ActionTimings"].([]any)[0].(map[string]any)["Uuid"] = ""
	rcv[1].(map[string]any)["ActionTimings"].([]any)[0].(map[string]any)["Uuid"] = ""
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(map[string]any)["Id"].(string) < rcv[j].(map[string]any)["Id"].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItDestinations(t *testing.T) {
	cmd := exec.Command("cgr-console", "destinations")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"Id":       "DST_1001",
			"Prefixes": []any{"1001"},
		},
		map[string]any{
			"Id":       "DST_1002",
			"Prefixes": []any{"1002"},
		},
		map[string]any{
			"Id":       "DST_1003",
			"Prefixes": []any{"1003"},
		},
		map[string]any{
			"Id":       "DST_FS",
			"Prefixes": []any{"10"},
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(map[string]any)["Id"].(string) < rcv[j].(map[string]any)["Id"].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItSchedulerReload(t *testing.T) {
	cmd := exec.Command("cgr-console", "scheduler_reload")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItActionPlanRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "actionplan_remove", `ID="AP_PACKAGE_10"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDispatchersProfileSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "dispatchers_profile_set", `ID="dps"`, `Subsystems=["attributes"]`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItAccountTriggersAdd(t *testing.T) {
	cmd := exec.Command("cgr-console", "account_triggers_add", `Account="1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItBalanceRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "balance_remove", `Account="1001"`, `Value=2`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDatacost(t *testing.T) {
	cmd := exec.Command("cgr-console", "datacost")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Account":     "",
		"Category":    "call",
		"Cost":        0.,
		"DataSpans":   []any{},
		"Destination": "",
		"Subject":     "",
		"Tenant":      "cgrates.org",
		"ToR":         "*data",
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDestinationSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "destination_set", `Id="DST_1004"`, `Prefixes=["1004"]`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSchedulerExecute(t *testing.T) {
	cmd := exec.Command("cgr-console", "scheduler_execute")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItAccountTriggersRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "account_triggers_remove", `Account="1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSetStordbVersions(t *testing.T) {
	cmd := exec.Command("cgr-console", "set_stordb_versions")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItChargersProfileSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "chargers_profile_set", `ID="cps"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSetDatadbVersions(t *testing.T) {
	cmd := exec.Command("cgr-console", "set_datadb_versions")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSharedGroup(t *testing.T) {
	cmd := exec.Command("cgr-console", "sharedgroup")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Id":                "",
		"AccountParameters": nil,
		"MemberIds":         nil,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItGetLoadTimes(t *testing.T) {
	cmd := exec.Command("cgr-console", "get_load_times", `Timezone="Local"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	expected := map[string]any{
		"*account_action_plans":     rcv["*account_action_plans"],
		"*action_plans":             rcv["*action_plans"],
		"*action_triggers":          rcv["*action_triggers"],
		"*actions":                  rcv["*actions"],
		"*attribute_filter_indexes": rcv["*attribute_filter_indexes"],
		"*attribute_profiles":       rcv["*attribute_profiles"],
		"*charger_profiles":         rcv["*charger_profiles"],
		"*destinations":             rcv["*destinations"],
		"*filters":                  rcv["*filters"],
		"*rating_plans":             rcv["*rating_plans"],
		"*rating_profiles":          rcv["*rating_profiles"],
		"*resource_profiles":        rcv["*resource_profiles"],
		"*resources":                rcv["*resources"],
		"*reverse_destinations":     rcv["*reverse_destinations"],
		"*route_profiles":           rcv["*route_profiles"],
		"*statqueue_profiles":       rcv["*statqueue_profiles"],
		"*statqueues":               rcv["*statqueues"],
		"*threshold_profiles":       rcv["*threshold_profiles"],
		"*thresholds":               rcv["*thresholds"],
		"*timings":                  rcv["*timings"],
		"test":                      rcv["test"],
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItLoadTpFromStordb(t *testing.T) {
	cmd := exec.Command("cgr-console", "load_tp_from_stordb", `TPid="TEST_SQL"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDebit(t *testing.T) {
	cmd := exec.Command("cgr-console", "debit", `Category="call"`, `Subject="1001"`, `Account="1001"`, `Destination="1002"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Category":    "call",
		"Tenant":      "cgrates.org",
		"Subject":     "1001",
		"Account":     "1001",
		"Destination": "1002",
		"ToR":         "",
		"Cost":        0.,
		"Timespans":   nil,
		"RatedUsage":  0.,
		"AccountSummary": map[string]any{
			"Tenant": "cgrates.org",
			"ID":     "1001",
			"BalanceSummaries": []any{
				map[string]any{
					"UUID":     "",
					"ID":       "test",
					"Type":     "*monetary",
					"Initial":  10.,
					"Value":    10.,
					"Disabled": false,
				},
			},
			"AllowNegative": false,
			"Disabled":      false,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv["AccountSummary"].(map[string]any)["BalanceSummaries"].([]any)[0].(map[string]any)["UUID"] = ""
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItFilterIndexesRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "filter_indexes_remove", `ItemType="test"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItBalanceDebit(t *testing.T) {
	cmd := exec.Command("cgr-console", "balance_debit", `Account="1001"`, `BalanceType="*monetary"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func reverseFormatting(i any, s utils.StringSet) (_ any, err error) {
	switch v := i.(type) {
	case map[string]any:
		for key, value := range v {
			if s.Has(key) {
				tmval, err := utils.IfaceAsDuration(value)
				if err != nil {
					return nil, err
				}
				v[key] = tmval
				continue
			}
			if v[key], err = reverseFormatting(value, s); err != nil {
				return nil, err
			}
		}
		return v, nil
	case []any:
		for i, value := range v {
			if v[i], err = reverseFormatting(value, s); err != nil {
				return
			}
		}
		return v, nil
	default:
		return i, nil
	}
}

func testConsoleItCost(t *testing.T) {
	cmd := exec.Command("cgr-console", "cost", `Tenant="cgrates.org"`, `Category="call"`, `Subject="1001"`, `AnswerTime="*now"`, `Destination="1002"`, `Usage="2m"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var tmp map[string]any
	if err := json.NewDecoder(output).Decode(&tmp); err != nil {
		t.Log(output.String())
		t.Error(output.String())
		t.Fatal(err)
	}
	reverseMap, err := reverseFormatting(tmp, utils.StringSet{
		utils.Usage:              {},
		utils.GroupIntervalStart: {},
		utils.RateIncrement:      {},
		utils.RateUnit:           {},
	})
	if err != nil {
		t.Fatal(err)
	}
	var rcv engine.EventCost
	if err := json.Unmarshal([]byte(utils.ToJSON(reverseMap)), &rcv); err != nil {
		t.Log(output.String())
		t.Error(output.String())
		t.Fatal(err)
	}
	usage := 2 * time.Minute
	// expected := new(engine.EventCost)
	expected := engine.EventCost{
		CGRID:     "",
		RunID:     "",
		StartTime: time.Time{},
		Usage:     &usage,
		Cost:      utils.Float64Pointer(0.7002),
		Charges: []*engine.ChargingInterval{
			{
				RatingID: rcv.Charges[0].RatingID,
				Increments: []*engine.ChargingIncrement{
					{
						Usage:          0 * time.Second,
						Cost:           0.4,
						AccountingID:   "",
						CompressFactor: 1.,
					},
					{
						Usage:          60 * time.Second,
						Cost:           0.2,
						AccountingID:   "",
						CompressFactor: 1.,
					},
				},
				CompressFactor: 1.,
			},
			{
				RatingID: rcv.Charges[0].RatingID,
				Increments: []*engine.ChargingIncrement{

					{
						Usage:          1 * time.Second,
						Cost:           0.00167,
						AccountingID:   "",
						CompressFactor: 60.,
					},
				},
				CompressFactor: 1.,
			},
		},
		Accounting:     engine.Accounting{},
		AccountSummary: nil,
		Rating: engine.Rating{
			rcv.Charges[0].RatingID: {
				ConnectFee:       0.4,
				MaxCost:          0.,
				MaxCostStrategy:  "",
				RatesID:          rcv.Rating[rcv.Charges[0].RatingID].RatesID,
				RatingFiltersID:  rcv.Rating[rcv.Charges[0].RatingID].RatingFiltersID,
				RoundingDecimals: 4.,
				RoundingMethod:   "*up",
				TimingID:         rcv.Rating[rcv.Charges[0].RatingID].TimingID,
			},
		},
		RatingFilters: engine.RatingFilters{
			rcv.Rating[rcv.Charges[0].RatingID].RatingFiltersID: engine.RatingMatchedFilters{
				"DestinationID":     "DST_1002",
				"DestinationPrefix": "1002",
				"RatingPlanID":      "RP_1001",
				"Subject":           "*out:cgrates.org:call:1001",
			},
		},
		Rates: engine.ChargedRates{
			rcv.Rating[rcv.Charges[0].RatingID].RatesID: {
				{
					GroupIntervalStart: 0 * time.Second,
					RateIncrement:      1 * time.Minute,
					RateUnit:           1 * time.Minute,
					Value:              0.2,
				},
				{
					GroupIntervalStart: 1 * time.Minute,
					RateIncrement:      1 * time.Second,
					RateUnit:           60 * time.Second,
					Value:              0.1,
				},
			},
		},
		Timings: engine.ChargedTimings{
			rcv.Rating[rcv.Charges[0].RatingID].TimingID: {
				MonthDays: utils.MonthDays{},
				Months:    utils.Months{},
				StartTime: "00:00:00",
				WeekDays:  utils.WeekDays{},
				Years:     utils.Years{},
			},
		},
	}
	rcv.StartTime = time.Time{}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItRatingPlanCost(t *testing.T) {
	cmd := exec.Command("cgr-console", "ratingplan_cost", `RatingPlanIDs=["RP_1001"]`, `SetupTime="*now"`, `Destination="1002"`, `Usage="2m0s"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv dispatchers.RatingPlanCost
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Fatal(err)
	}
	usage := 2 * time.Minute
	expected := dispatchers.RatingPlanCost{
		EventCost: &engine.EventCost{
			CGRID:     "",
			RunID:     "",
			StartTime: time.Time{},
			Usage:     &usage,
			Cost:      utils.Float64Pointer(0.7002),
			Charges: []*engine.ChargingInterval{
				{
					RatingID: rcv.EventCost.Charges[0].RatingID,
					Increments: []*engine.ChargingIncrement{
						{
							Usage:          0 * time.Second,
							Cost:           0.4,
							AccountingID:   "",
							CompressFactor: 1.,
						},
						{
							Usage:          60 * time.Second,
							Cost:           0.2,
							AccountingID:   "",
							CompressFactor: 1.,
						},
					},
					CompressFactor: 1.,
				},
				{
					RatingID: rcv.EventCost.Charges[0].RatingID,
					Increments: []*engine.ChargingIncrement{

						{
							Usage:          1 * time.Second,
							Cost:           0.00167,
							AccountingID:   "",
							CompressFactor: 60.,
						},
					},
					CompressFactor: 1.,
				},
			},
			Accounting:     engine.Accounting{},
			AccountSummary: nil,
			Rating: engine.Rating{
				rcv.EventCost.Charges[0].RatingID: {
					ConnectFee:       0.4,
					MaxCost:          0.,
					MaxCostStrategy:  "",
					RatesID:          rcv.EventCost.Rating[rcv.EventCost.Charges[0].RatingID].RatesID,
					RatingFiltersID:  rcv.EventCost.Rating[rcv.EventCost.Charges[0].RatingID].RatingFiltersID,
					RoundingDecimals: 4.,
					RoundingMethod:   "*up",
					TimingID:         rcv.EventCost.Rating[rcv.EventCost.Charges[0].RatingID].TimingID,
				},
			},
			RatingFilters: engine.RatingFilters{
				rcv.EventCost.Rating[rcv.EventCost.Charges[0].RatingID].RatingFiltersID: engine.RatingMatchedFilters{
					"DestinationID":     "DST_1002",
					"DestinationPrefix": "1002",
					"RatingPlanID":      "RP_1001",
					"Subject":           "*out:cgrates.org:call:1001",
				},
			},
			Rates: engine.ChargedRates{
				rcv.EventCost.Rating[rcv.EventCost.Charges[0].RatingID].RatesID: {
					{
						GroupIntervalStart: 0 * time.Second,
						RateIncrement:      1 * time.Minute,
						RateUnit:           1 * time.Minute,
						Value:              0.2,
					},
					{
						GroupIntervalStart: 1 * time.Minute,
						RateIncrement:      1 * time.Second,
						RateUnit:           60 * time.Second,
						Value:              0.1,
					},
				},
			},
			Timings: engine.ChargedTimings{
				rcv.EventCost.Rating[rcv.EventCost.Charges[0].RatingID].TimingID: {
					MonthDays: utils.MonthDays{},
					Months:    utils.Months{},
					StartTime: "00:00:00",
					WeekDays:  utils.WeekDays{},
					Years:     utils.Years{},
				},
			},
		},
		RatingPlanID: "RP_1001",
	}
	rcv.EventCost.StartTime = time.Time{}
	if !reflect.DeepEqual(rcv, expected) && !strings.Contains(rcv.EventCost.RatingFilters[rcv.EventCost.Rating[rcv.EventCost.Charges[0].RatingID].RatingFiltersID]["Subject"].(string), "rating_plan_cost") {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItActiveSessions(t *testing.T) {
	cmd := exec.Command("cgr-console", "active_sessions")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"Account":       "1001",
			"AnswerTime":    "0001-01-01T00:00:00Z",
			"CGRID":         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			"Category":      "",
			"DebitInterval": "0s",
			"Destination":   "",
			"DurationIndex": "0s",
			"ExtraFields":   map[string]any{},
			"LoopIndex":     0.,
			"MaxCostSoFar":  0.,
			"MaxRate":       0.,
			"MaxRateUnit":   "0s",
			"NextAutoDebit": "0001-01-01T00:00:00Z",
			"NodeID":        "",
			"OriginHost":    "",
			"OriginID":      "",
			"RequestType":   "",
			"RunID":         "",
			"SetupTime":     "0001-01-01T00:00:00Z",
			"Source":        "SessionS_",
			"Subject":       "",
			"Tenant":        "cgrates.org",
			"ToR":           "",
			"Usage":         "0s",
		},
		map[string]any{
			"Account":       "1001",
			"AnswerTime":    "0001-01-01T00:00:00Z",
			"CGRID":         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			"Category":      "",
			"DebitInterval": "0s",
			"Destination":   "",
			"DurationIndex": "0s",
			"ExtraFields":   map[string]any{},
			"LoopIndex":     0.,
			"MaxCostSoFar":  0.,
			"MaxRate":       0.,
			"MaxRateUnit":   "0s",
			"NextAutoDebit": "0001-01-01T00:00:00Z",
			"NodeID":        "",
			"OriginHost":    "",
			"OriginID":      "",
			"RequestType":   "",
			"RunID":         "*default",
			"SetupTime":     "0001-01-01T00:00:00Z",
			"Source":        "SessionS_",
			"Subject":       "",
			"Tenant":        "cgrates.org",
			"ToR":           "",
			"Usage":         "0s",
		},
		map[string]any{
			"Account":       "1001",
			"AnswerTime":    "0001-01-01T00:00:00Z",
			"CGRID":         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			"Category":      "",
			"DebitInterval": "0s",
			"Destination":   "",
			"DurationIndex": "0s",
			"ExtraFields":   map[string]any{},
			"LoopIndex":     0.,
			"MaxCostSoFar":  0.,
			"MaxRate":       0.,
			"MaxRateUnit":   "0s",
			"NextAutoDebit": "0001-01-01T00:00:00Z",
			"NodeID":        "",
			"OriginHost":    "",
			"OriginID":      "",
			"RequestType":   "",
			"RunID":         "*default",
			"SetupTime":     "0001-01-01T00:00:00Z",
			"Source":        "SessionS_",
			"Subject":       "",
			"Tenant":        "cgrates.org",
			"ToR":           "",
			"Usage":         "0s",
		},
		map[string]any{
			"Account":       "1001",
			"AnswerTime":    "0001-01-01T00:00:00Z",
			"CGRID":         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			"Category":      "",
			"DebitInterval": "0s",
			"Destination":   "",
			"DurationIndex": "0s",
			"ExtraFields":   map[string]any{},
			"LoopIndex":     0.,
			"MaxCostSoFar":  0.,
			"MaxRate":       0.,
			"MaxRateUnit":   "0s",
			"NextAutoDebit": "0001-01-01T00:00:00Z",
			"NodeID":        "",
			"OriginHost":    "",
			"OriginID":      "",
			"RequestType":   "*none",
			"RunID":         "*raw",
			"SetupTime":     "0001-01-01T00:00:00Z",
			"Source":        "SessionS_",
			"Subject":       "",
			"Tenant":        "cgrates.org",
			"ToR":           "",
			"Usage":         "0s",
		},
		map[string]any{
			"Account":       "1001",
			"AnswerTime":    "0001-01-01T00:00:00Z",
			"CGRID":         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			"Category":      "",
			"DebitInterval": "0s",
			"Destination":   "",
			"DurationIndex": "0s",
			"ExtraFields":   map[string]any{},
			"LoopIndex":     0.,
			"MaxCostSoFar":  0.,
			"MaxRate":       0.,
			"MaxRateUnit":   "0s",
			"NextAutoDebit": "0001-01-01T00:00:00Z",
			"NodeID":        "",
			"OriginHost":    "",
			"OriginID":      "",
			"RequestType":   "",
			"RunID":         "reseller1",
			"SetupTime":     "0001-01-01T00:00:00Z",
			"Source":        "SessionS_",
			"Subject":       "",
			"Tenant":        "cgrates.org",
			"ToR":           "",
			"Usage":         "0s",
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Error(output.String())
		t.Fatal(err)
	}
	for i := range rcv {
		rcv[i].(map[string]any)["NodeID"] = ""
		rcv[i].(map[string]any)["DurationIndex"] = "0s"
	}
	sort.Slice(rcv, func(i, j int) bool {
		return utils.IfaceAsString(rcv[i].(map[string]any)["RunID"]) < utils.IfaceAsString(rcv[j].(map[string]any)["RunID"])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItPassiveSessions(t *testing.T) {
	var reply string
	err := cnslRPC.Call(context.Background(), utils.SessionSv1DeactivateSessions, &utils.SessionIDsWithArgsDispatcher{}, &reply)
	if err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: OK, received : %+v", reply)
	}
	args := &utils.SessionFilter{
		APIOpts: make(map[string]any),
	}
	var reply2 []*sessions.ExternalSession
	if err := cnslRPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions, args, &reply2); err != nil {
		t.Error(err)
	}
	expected := []*sessions.ExternalSession{
		{
			Account:       "1001",
			AnswerTime:    time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			CGRID:         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			Category:      "",
			DebitInterval: 0 * time.Second,
			Destination:   "",
			DurationIndex: 0 * time.Second,
			ExtraFields:   map[string]string{},
			LoopIndex:     0.,
			MaxCostSoFar:  0.,
			MaxRate:       0.,
			MaxRateUnit:   0 * time.Second,
			NextAutoDebit: time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			NodeID:        "",
			OriginHost:    "",
			OriginID:      "",
			RequestType:   "",
			RunID:         "",
			SetupTime:     time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:        "SessionS_",
			Subject:       "",
			Tenant:        "cgrates.org",
			ToR:           "",
			Usage:         0 * time.Second,
		},
		{
			Account:       "1001",
			AnswerTime:    time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			CGRID:         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			Category:      "",
			DebitInterval: 0 * time.Second,
			Destination:   "",
			DurationIndex: 0 * time.Second,
			ExtraFields:   map[string]string{},
			LoopIndex:     0.,
			MaxCostSoFar:  0.,
			MaxRate:       0.,
			MaxRateUnit:   0 * time.Second,
			NextAutoDebit: time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			NodeID:        "",
			OriginHost:    "",
			OriginID:      "",
			RequestType:   "",
			RunID:         "*default",
			SetupTime:     time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:        "SessionS_",
			Subject:       "",
			Tenant:        "cgrates.org",
			ToR:           "",
			Usage:         0 * time.Second,
		},
		{
			Account:       "1001",
			AnswerTime:    time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			CGRID:         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			Category:      "",
			DebitInterval: 0 * time.Second,
			Destination:   "",
			DurationIndex: 0 * time.Second,
			ExtraFields:   map[string]string{},
			LoopIndex:     0.,
			MaxCostSoFar:  0.,
			MaxRate:       0.,
			MaxRateUnit:   0 * time.Second,
			NextAutoDebit: time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			NodeID:        "",
			OriginHost:    "",
			OriginID:      "",
			RequestType:   "",
			RunID:         "*default",
			SetupTime:     time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:        "SessionS_",
			Subject:       "",
			Tenant:        "cgrates.org",
			ToR:           "",
			Usage:         0 * time.Second,
		},
		{
			Account:       "1001",
			AnswerTime:    time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			CGRID:         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			Category:      "",
			DebitInterval: 0 * time.Second,
			Destination:   "",
			DurationIndex: 0 * time.Second,
			ExtraFields:   map[string]string{},
			LoopIndex:     0.,
			MaxCostSoFar:  0.,
			MaxRate:       0.,
			MaxRateUnit:   0 * time.Second,
			NextAutoDebit: time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			NodeID:        "",
			OriginHost:    "",
			OriginID:      "",
			RequestType:   "*none",
			RunID:         "*raw",
			SetupTime:     time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:        "SessionS_",
			Subject:       "",
			Tenant:        "cgrates.org",
			ToR:           "",
			Usage:         0 * time.Second,
		},
		{
			Account:       "1001",
			AnswerTime:    time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			CGRID:         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			Category:      "",
			DebitInterval: 0 * time.Second,
			Destination:   "",
			DurationIndex: 0 * time.Second,
			ExtraFields:   map[string]string{},
			LoopIndex:     0.,
			MaxCostSoFar:  0.,
			MaxRate:       0.,
			MaxRateUnit:   0 * time.Second,
			NextAutoDebit: time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			NodeID:        "",
			OriginHost:    "",
			OriginID:      "",
			RequestType:   "",
			RunID:         "reseller1",
			SetupTime:     time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			Source:        "SessionS_",
			Subject:       "",
			Tenant:        "cgrates.org",
			ToR:           "",
			Usage:         0 * time.Second,
		},
	}
	for i := range reply2 {
		reply2[i].NodeID = ""
		reply2[i].DurationIndex = 0 * time.Second
	}
	sort.Slice(reply2, func(i, j int) bool {
		return reply2[i].RunID < reply2[j].RunID
	})
	if !reflect.DeepEqual(reply2, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(reply2))
	}
}

func testConsoleItRatingProfileRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "ratingprofile_remove", `Tenant="cgrates.org"`, `Category="call"`, `Subject="1001"`)
	expected := "OK"
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItActionPlanSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "actionplan_set", `ID="AP_TEST"`, `ActionPlan=[{"ActionsId":"ACT_TOPUP_RST_10", "MonthDays":"1", "Time":"00:00:00", "Weight": 20.0}]`)
	expected := "OK"
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSchedulerQueue(t *testing.T) {
	cmd := exec.Command("cgr-console", "scheduler_queue")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"NextRunTime":      "",
			"Accounts":         0.,
			"ActionPlanID":     "AP_TEST",
			"ActionTimingUUID": "",
			"ActionsID":        "ACT_TOPUP_RST_10",
		},
		map[string]any{
			"NextRunTime":      "",
			"Accounts":         3.,
			"ActionPlanID":     "STANDARD_PLAN",
			"ActionTimingUUID": "",
			"ActionsID":        "TOPUP_RST_1024_DATA",
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	for i := range rcv {
		rcv[i].(map[string]any)["NextRunTime"] = ""
		rcv[i].(map[string]any)["ActionTimingUUID"] = ""
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(map[string]any)["ActionPlanID"].(string) < rcv[j].(map[string]any)["ActionPlanID"].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItMaxDuration(t *testing.T) {
	cmd := exec.Command("cgr-console", "maxduration", `Account="1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "0s"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItActionExecute(t *testing.T) {
	cmd := exec.Command("cgr-console", "action_execute", `Account="1001"`, `ActionsId="ACT_TOPUP_RST_10"`)
	expected := "OK"
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItAttributesProcessEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "attributes_process_event", `Tenant="cgrates.org"`, `Event={"Account":"1001"}`, `Context="*sessions"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"APIOpts":       map[string]any{},
		"AlteredFields": []any{"*req.LCRProfile", "*req.Password", "*req.PaypalAccount", "*req.RequestType"},
		"Event": map[string]any{
			"Account":       "1001",
			"LCRProfile":    "premium_cli",
			"Password":      "CGRateS.org",
			"PaypalAccount": "cgrates@paypal.com",
			"RequestType":   "*prepaid",
		},
		"ID":              "",
		"MatchedProfiles": []any{"ATTR_1001_SESSIONAUTH"},
		"Tenant":          "cgrates.org",
		"Time":            "",
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv["Time"] = ""
	sort.Slice(rcv["AlteredFields"], func(i, j int) bool {
		return utils.IfaceAsString(rcv["AlteredFields"].([]any)[i]) < utils.IfaceAsString(rcv["AlteredFields"].([]any)[j])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItAttributesForEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "attributes_for_event", `AttributeIDs=["ATTR_1001_SESSIONAUTH"]`, `Event={"Account":"1001"}`, `Context="*sessions"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Tenant":             "cgrates.org",
		"ID":                 "ATTR_1001_SESSIONAUTH",
		"Contexts":           []any{"*sessions"},
		"FilterIDs":          []any{"*string:~*req.Account:1001"},
		"ActivationInterval": nil,
		"Attributes": []any{
			map[string]any{
				"FilterIDs": []any{},
				"Path":      "*req.Password",
				"Type":      "*constant",
				"Value": []any{
					map[string]any{
						"Rules": "CGRateS.org",
					},
				},
			},
			map[string]any{
				"FilterIDs": []any{},
				"Path":      "*req.RequestType",
				"Type":      "*constant",
				"Value": []any{
					map[string]any{
						"Rules": "*prepaid",
					},
				},
			},
			map[string]any{
				"FilterIDs": []any{},
				"Path":      "*req.PaypalAccount",
				"Type":      "*constant",
				"Value": []any{
					map[string]any{
						"Rules": "cgrates@paypal.com",
					},
				},
			},
			map[string]any{
				"FilterIDs": []any{},
				"Path":      "*req.LCRProfile",
				"Type":      "*constant",
				"Value": []any{
					map[string]any{
						"Rules": "premium_cli",
					},
				},
			},
		},
		"Blocker": false,
		"Weight":  10.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItChargersProfileRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "chargers_profile_remove", `ID="DEFAULT"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItThresholdsProfileRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "thresholds_profile_remove", `ID="THD_ACNT_1002"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItResourcesProfileRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "resources_profile_remove", `ID="ResGroup1"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDispatchersHostSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "dispatchers_host_set", `ID="DHS_SET"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDispatchersHost(t *testing.T) {
	cmd := exec.Command("cgr-console", "dispatchers_host", `ID="DHS_SET"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Tenant":      "cgrates.org",
		"ID":          "DHS_SET",
		"Address":     "",
		"Transport":   "",
		"Synchronous": false,
		"TLS":         false,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItDispatchersHostRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "dispatchers_host_remove", `ID="DHS_SET"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDispatchersProfileIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "dispatchers_profile_ids")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{"dps"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDispatchersProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "dispatchers_profile", `ID="dps"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Tenant":             "cgrates.org",
		"ID":                 "dps",
		"Subsystems":         []any{"attributes"},
		"FilterIDs":          nil,
		"ActivationInterval": nil,
		"Strategy":           "",
		"StrategyParams":     nil,
		"Weight":             0.,
		"Hosts":              nil,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItAttributesProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "attributes_profile", `ID="ATTR_1001_SIMPLEAUTH"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"Tenant":             "cgrates.org",
		"ID":                 "ATTR_1001_SIMPLEAUTH",
		"Contexts":           []any{"simpleauth"},
		"FilterIDs":          []any{"*string:~*req.Account:1001"},
		"ActivationInterval": nil,
		"Attributes": []any{
			map[string]any{
				"FilterIDs": []any{},
				"Path":      "*req.Password",
				"Type":      "*constant",
				"Value": []any{
					map[string]any{
						"Rules": "CGRateS.org",
					},
				},
			},
		},
		"Blocker": false,
		"Weight":  20.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItRoutesProfileRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "routes_profile_remove", `ID="rps"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItStatsProfileRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "stats_profile_remove", `ID="123"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItAttributesProfileRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "attributes_profile_remove", `ID="attrID"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItFilterRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "filter_remove", `ID="FLTR_ACNT_1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "Error executing command: SERVER_ERROR: cannot remove filter <cgrates.org:FLTR_ACNT_1001> because will broken the reference to following items: {\"*route_filter_indexes\":{\"ROUTE_ACNT_1001\":{}},\"*threshold_filter_indexes\":{\"THD_ACNT_1001\":{}}}\n"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(output.String(), expected) {
		fmt.Printf("%T and %T", output.String(), expected)
		t.Fatalf("Expected %+q \n but received \n %+q", expected, output.String())
	}
}

func testConsoleItMaxUsage(t *testing.T) {
	cmd := exec.Command("cgr-console", "maxusage", `ToR="*prepaid"`, `Account="1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := 0.
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv float64
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", expected, rcv)
	}
}

func testConsoleItSessionProcessCdr(t *testing.T) {
	cmd := exec.Command("cgr-console", "session_process_cdr", `Event={"Account":"1001", "Source":"*sessions", "Usage":"2m30s", "CostDetails":{"CGRID":"2", "RunID":"*test"}, "OriginID":"169"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItCdrs(t *testing.T) {
	cmd := exec.Command("cgr-console", "cdrs")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"Account":    "1001",
			"AnswerTime": "0001-01-01T00:00:00Z",
			"CGRID":      "2659fc519890c924f82b4475ddd71b058178d02b",
			"Category":   "call",
			"Cost":       -1.,
			"CostDetails": map[string]any{
				"AccountSummary": nil,
				"Accounting":     nil,
				"CGRID":          "2",
				"Charges":        nil,
				"Cost":           nil,
				"Rates":          nil,
				"Rating":         nil,
				"RatingFilters":  nil,
				"RunID":          "*test",
				"StartTime":      "0001-01-01T00:00:00Z",
				"Timings":        nil,
				"Usage":          0.,
			},
			"CostSource":  "",
			"Destination": "",
			"ExtraFields": map[string]any{},
			"ExtraInfo":   "NOT_CONNECTED: RALs",
			"OrderID":     nil,
			"OriginHost":  "",
			"OriginID":    "169",
			"Partial":     false,
			"PreRated":    false,
			"RequestType": "*rated",
			"RunID":       "*default",
			"SetupTime":   "0001-01-01T00:00:00Z",
			"Source":      "*sessions",
			"Subject":     "1001",
			"Tenant":      "cgrates.org",
			"ToR":         "*voice",
			"Usage":       150000000000.,
		},
		map[string]any{
			"Account":    "1001",
			"AnswerTime": "0001-01-01T00:00:00Z",
			"CGRID":      "2659fc519890c924f82b4475ddd71b058178d02b",
			"Category":   "call",
			"Cost":       -1.,
			"CostDetails": map[string]any{
				"AccountSummary": nil,
				"Accounting":     nil,
				"CGRID":          "2",
				"Charges":        nil,
				"Cost":           nil,
				"Rates":          nil,
				"Rating":         nil,
				"RatingFilters":  nil,
				"RunID":          "*test",
				"StartTime":      "0001-01-01T00:00:00Z",
				"Timings":        nil,
				"Usage":          nil,
			},
			"CostSource":  "",
			"Destination": "",
			"ExtraFields": map[string]any{},
			"ExtraInfo":   "",
			"OrderID":     nil,
			"OriginHost":  "",
			"OriginID":    "169",
			"Partial":     false,
			"PreRated":    false,
			"RequestType": "*none",
			"RunID":       "*raw",
			"SetupTime":   "0001-01-01T00:00:00Z",
			"Source":      "*sessions",
			"Subject":     "1001",
			"Tenant":      "cgrates.org",
			"ToR":         "*voice",
			"Usage":       150000000000.,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	// // rcv[0].(map[string]any)["CGRID"] = ""
	rcv[0].(map[string]any)["OrderID"] = nil
	// // rcv[1].(map[string]any)["CGRID"] = ""
	rcv[1].(map[string]any)["OrderID"] = nil
	sort.Slice(rcv, func(i, j int) bool {
		return utils.IfaceAsString(rcv[i].(map[string]any)["RunID"]) < utils.IfaceAsString(rcv[j].(map[string]any)["RunID"])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItDispatchersHostIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "dispatchers_host_ids")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{"DHS_SET"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItDebitMax(t *testing.T) {
	cmd := exec.Command("cgr-console", "debit_max", `Category="call"`, `Account="1001"`, `Destination="1002"`, `TimeStart="2016-09-01T00:00:00Z"`, `TimeEnd="2016-09-01T00:00:01Z"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := &engine.CallCost{
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		ToR:         "*voice",
		Cost:        0.6,
		Timespans: engine.TimeSpans{
			{
				TimeStart: time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC),
				TimeEnd:   time.Date(2016, 9, 1, 0, 1, 0, 0, time.UTC),
				Cost:      0.6,
				RateInterval: &engine.RateInterval{
					Timing: &engine.RITiming{
						ID:        "*any",
						Years:     utils.Years{},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: "00:00:00",
						EndTime:   "",
					},
					Rating: &engine.RIRate{
						ConnectFee:       0.4,
						RoundingMethod:   utils.MetaRoundingUp,
						RoundingDecimals: 4.,
						MaxCost:          0.,
						MaxCostStrategy:  "",
						Rates: engine.RateGroups{
							{
								GroupIntervalStart: 0 * time.Second,
								Value:              0.2,
								RateIncrement:      60000000000 * time.Nanosecond,
								RateUnit:           60000000000 * time.Nanosecond,
							},
							{
								GroupIntervalStart: 60000000000 * time.Nanosecond,
								Value:              0.1,
								RateIncrement:      1000000000 * time.Nanosecond,
								RateUnit:           60000000000 * time.Nanosecond,
							},
						},
					},
					Weight: 10.,
				},
				DurationIndex: 60000000000 * time.Nanosecond,
				Increments: engine.Increments{
					{
						Duration: 0.,
						Cost:     0.4,
						BalanceInfo: &engine.DebitInfo{
							Unit: nil,
							Monetary: &engine.MonetaryInfo{
								UUID:         "",
								ID:           "test",
								Value:        9.6,
								RateInterval: nil,
							},
							AccountID: "cgrates.org:1001",
						},
						CompressFactor: 1.,
					},
					{
						Duration: 60000000000.,
						Cost:     0.2,
						BalanceInfo: &engine.DebitInfo{
							Unit: nil,
							Monetary: &engine.MonetaryInfo{
								UUID:         "",
								ID:           "test",
								Value:        9.4,
								RateInterval: nil,
							},
							AccountID: "cgrates.org:1001",
						},
						CompressFactor: 1.,
					},
				},
				RoundIncrement: nil,
				MatchedSubject: "*out:cgrates.org:call:1001",
				MatchedPrefix:  "1002",
				MatchedDestId:  "DST_1002",
				RatingPlanId:   "RP_1001",
				CompressFactor: 1.,
			},
		},
		RatedUsage: 60000000000.,
		AccountSummary: &engine.AccountSummary{
			Tenant: "cgrates.org",
			ID:     "1001",
			BalanceSummaries: engine.BalanceSummaries{
				{
					UUID:     "",
					ID:       "test",
					Type:     "*monetary",
					Initial:  10.,
					Value:    9.4,
					Disabled: false,
				},
			},
			AllowNegative: false,
			Disabled:      false,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv *engine.CallCost
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv.Timespans[0].Increments[0].BalanceInfo.Monetary.UUID = ""
	rcv.Timespans[0].Increments[1].BalanceInfo.Monetary.UUID = ""
	rcv.AccountSummary.BalanceSummaries[0].UUID = ""
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItAccountTriggersSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "account_triggers_set", `Account="1001"`, `GroupID="ATS_TEST"`, `ActionTrigger={"ThresholdType":"*min_balance", "ThresholdValue":2}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItAccountActionPlanGet(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+cnslItCfgPath, "-path="+path.Join(*utils.DataDir, "tariffplans", "tutorial2"))
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	cmd = exec.Command("cgr-console", "account_actionplan_get", `Account="1001"`)
	output = bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"ActionPlanId": "STANDARD_PLAN",
			"ActionsId":    "TOPUP_RST_MONETARY_10",
			"NextExecTime": "",
			"Uuid":         "",
		},
		map[string]any{
			"ActionPlanId": "STANDARD_PLAN",
			"ActionsId":    "TOPUP_RST_5M_VOICE",
			"NextExecTime": "",
			"Uuid":         "",
		},
		map[string]any{
			"ActionPlanId": "STANDARD_PLAN",
			"ActionsId":    "TOPUP_RST_10M_VOICE",
			"NextExecTime": "",
			"Uuid":         "",
		},
		map[string]any{
			"ActionPlanId": "STANDARD_PLAN",
			"ActionsId":    "TOPUP_RST_100_SMS",
			"NextExecTime": "",
			"Uuid":         "",
		},
		map[string]any{
			"ActionPlanId": "STANDARD_PLAN",
			"ActionsId":    "TOPUP_RST_1024_DATA",
			"NextExecTime": "",
			"Uuid":         "",
		},
		map[string]any{
			"ActionPlanId": "STANDARD_PLAN",
			"ActionsId":    "TOPUP_RST_1024_DATA",
			"NextExecTime": "",
			"Uuid":         "",
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	for i := range rcv {
		rcv[i].(map[string]any)["Uuid"] = ""
		rcv[i].(map[string]any)["NextExecTime"] = ""
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
	cmd = exec.Command("cgr-loader", "-config_path="+cnslItCfgPath, "-path="+path.Join(*utils.DataDir, "tariffplans", "tutorial"))
	output = bytes.NewBuffer(nil)
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
}

func testConsoleItCacheItemIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_item_ids", `CacheID="*threshold_profiles"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{"cgrates.org:123", "cgrates.org:THD_ACNT_1001"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(string) < rcv[j].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItCacheItemExpiryTime(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_item_expiry_time", `CacheID="*threshold_profiles"`, `ItemID="cgrates.org:THD_ACNT_1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)

		t.Fatal(err)
	}
	t.Log(output.String())
	var rcv time.Time
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDispatchersProfileRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "dispatchers_profile_remove", `ID="dps"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItFilterIndexes(t *testing.T) {
	cmd := exec.Command("cgr-console", "filter_indexes", `ItemType="*attributes"`, `FilterType="*string"`, `Tenant="cgrates.org"`, `Context="*sessions"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		"*string:*req.Account:1001:ATTR_1001_SESSIONAUTH",
		"*string:*req.Account:1002:ATTR_1002_SESSIONAUTH",
		"*string:*req.Account:1003:ATTR_1003_SESSIONAUTH",
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(string) < rcv[j].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItDispatchesForEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "dispatches_for_event", `Tenant="cgrates.org"`, `Event={"EventName":"Event1"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		map[string]any{
			"Tenant":             "cgrates.org",
			"ID":                 "EVENT1",
			"Subsystems":         []any{"*any"},
			"FilterIDs":          []any{"*string:~*req.EventName:Event1"},
			"ActivationInterval": nil,
			"Strategy":           "*weight",
			"StrategyParams":     map[string]any{},
			"Weight":             30.,
			"Hosts": []any{
				map[string]any{
					"ID":        "ALL2",
					"FilterIDs": []any{},
					"Weight":    20.,
					"Params":    map[string]any{},
					"Blocker":   false,
				},
				map[string]any{
					"ID":        "ALL",
					"FilterIDs": []any{},
					"Weight":    10.,
					"Params":    map[string]any{},
					"Blocker":   false,
				},
			},
		},
		map[string]any{
			"Tenant":             "cgrates.org",
			"ID":                 "PING1",
			"Subsystems":         []any{"*any"},
			"FilterIDs":          []any{},
			"ActivationInterval": nil,
			"Strategy":           "*weight",
			"StrategyParams":     map[string]any{},
			"Weight":             10.,
			"Hosts": []any{
				map[string]any{
					"ID":        "ALL",
					"FilterIDs": []any{},
					"Weight":    20.,
					"Params":    map[string]any{},
					"Blocker":   false,
				},
				map[string]any{
					"ID":        "ALL2",
					"FilterIDs": []any{},
					"Weight":    10.,
					"Params":    map[string]any{},
					"Blocker":   false,
				},
			},
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(map[string]any)["ID"].(string) < rcv[j].(map[string]any)["ID"].(string)
	})
	// sort.Slice(rcv[])
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItLoaderLoad(t *testing.T) {
	cmd := exec.Command("cgr-console", "loader_load", `LoaderID="CustomLoader"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItLoaderRemove(t *testing.T) {
	cmd := exec.Command("cgr-console", "loader_remove", `LoaderID="CustomLoader"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItComputeActionplanIndexes(t *testing.T) {
	cmd := exec.Command("cgr-console", "compute_actionplan_indexes")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItParse(t *testing.T) {
	cmd := exec.Command("cgr-console", "parse", `Expression="~1:s/intra_(.*)/super_${1}/"`, `Value="intra_33"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "super_33\n"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)

		t.Fatal(err)
	}
	t.Log(output.String())
	if !reflect.DeepEqual(output.String(), expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, output.String())
	}
}

func testconsoleItBalanceSet(t *testing.T) {
	cmd := exec.Command("cgr-console", "balance_set", `Account="1001"`, `Value=10`, `Balance={"ID":"2"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := "OK"
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv string
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItCacheGroupItemIds(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_group_item_ids", `CacheID="*reverse_filter_indexes"`, `GroupID="cgrates.org:FLTR_ACNT_1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []any{
		"cgrates.org:FLTR_ACNT_1001:*route_filter_indexes",
		"cgrates.org:FLTR_ACNT_1001:*threshold_filter_indexes",
		// "cgrates.org:FLTR_ACNT_1001:fii_cgrates.org:FLTR_ACNT_1001_1002:*stat_filter_indexes",
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(string) < rcv[j].(string)
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItCostDetails(t *testing.T) {
	cmd := exec.Command("cgr-console", "cost_details", `CgrId="2659fc519890c924f82b4475ddd71b058178d02b"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]any{
		"AccountSummary": nil,
		"Accounting":     nil,
		"CGRID":          "2",
		"Charges":        nil,
		"Cost":           nil,
		"Rates":          nil,
		"Rating":         nil,
		"RatingFilters":  nil,
		"RunID":          "*test",
		"StartTime":      "0001-01-01T00:00:00Z",
		"Timings":        nil,
		"Usage":          "0s",
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)

		t.Fatal(err)
	}
	t.Log(output.String())
	var rcv map[string]any
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItLoadersKillEngine(t *testing.T) {
	fldPathIn := "/tmp/In"
	fldPathOut := "/tmp/Out"
	if err := os.Remove(fldPathIn); err != nil {
		t.Error(err)
	}
	if err := os.Remove(fldPathOut); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}
