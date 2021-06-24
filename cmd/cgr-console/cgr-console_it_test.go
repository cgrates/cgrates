// +build integration

/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	dbType    = flag.String("dbtype", utils.MetaInternal, "The type of DataBase (Internal/Mongo/mySql)")
	waitRater = flag.Int("wait_rater", 100, "Number of milliseconds to wait for rater to start and cache")
	// encoding  = flag.String("rpc", utils.MetaJSON, "what encoding whould be used for rpc comunication")
)

var (
	cnslItCfgPath string
	cnslItDirPath string
	cnslItCfg     *config.CGRConfig
	cnslItTests   = []func(t *testing.T){
		testConsoleItLoadConfig,
		testConsoleItInitDataDB,
		testConsoleItInitStorDB,
		testConsoleItStartEngine,
		testConsoleItLoadTP,
		testConsoleItCacheClear,
		testConsoleItThreshold,
		testConsoleItThresholdsProfileIds,
		testConsoleItThresholdsProfileSet,
		testConsoleItThresholdsProfile,
		testConsoleItThresholdsProcessEvent,
		testConsoleItThresholdsForEvent,
		testConsoleItTriggersSet,
		testConsoleItTriggers,
		testConsoleItSchedulerReload,
		testConsoleItSchedulerExecute,
		testConsoleItAccountTriggersReset,
		testConsoleItAccountTriggersAdd,
		testConsoleItAccountTriggersRemove,
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
		testConsoleItChargersProfile,
		testConsoleItChargersProfileSet,
		testConsoleItChargersForEvent,
		testConsoleItChargersProfileIds,
		testConsoleItChargersProcessEvent,
		testConsoleItResourcesAuthorize,
		testConsoleItRouteProfileIds,
		testConsoleItRoutesProfilesForEvent,
		testConsoleItRoutesProfile,
		testConsoleItRoutes,
		testConsoleItRoutesProfileSet,
		testConsoleItComputeFilterIndexes,
		// testConsoleItFilterIndexes,
		testConsoleItCacheReload,
		testConsoleItAttributesProfileIds,
		testConsoleItAttributesProfileSet,
		testConsoleItActionPlanGet,
		testConsoleItActionPlanRemove,
		testConsoleItDispatchersProfileSet,
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
		testConsoleItGetJsonSection,
		testConsoleItStatus,
		testConsoleItSharedGroup,
		testConsoleItDatacost,
		testConsoleItCost,
		testConsoleItRatingPlanCost,
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
		testConsoleItPing,
		testConsoleItLoadTpFromFolder,
		testConsoleItImportTpFromFolder,
		testConsoleItLoadTpFromStordb,
		testConsoleItAccounts,
		testConsoleItAccountRemove,
		testConsoleItBalanceAdd,
		testConsoleItBalanceRemove,
		testConsoleItBalanceDebit,
		testConsoleItGetLoadTimes,
		testConsoleItGetLoadIds,
		testConsoleItSessionAuthorizeEvent,
		testConsoleItCachePrecacheStatus,
		testConsoleItSessionInitiate,
		testConsoleItSessionProcessMessage,
		testConsoleItSessionUpdate,
		testConsoleItSleep,
		testConsoleItCacheRemoveGroup,
		// testConsoleItCacheStats,
		testConsoleItReloadConfig,
		testConsoleItKillEngine,
	}
)

func TestConsoleItTests(t *testing.T) {
	switch *dbType {
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

func testConsoleItLoadConfig(t *testing.T) {
	var err error
	cnslItCfgPath = path.Join(*dataDir, "conf", "samples", cnslItDirPath)
	if cnslItCfg, err = config.NewCGRConfigFromPath(cnslItCfgPath); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItInitDataDB(t *testing.T) {
	if err := engine.InitDataDb(cnslItCfg); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItInitStorDB(t *testing.T) {
	if err := engine.InitStorDb(cnslItCfg); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(cnslItCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testConsoleItLoadTP(t *testing.T) {
	cmd := exec.Command("cgr-loader", "-config_path="+cnslItCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "tutorial"))
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
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
	expected := map[string]interface{}{
		"*tcc": "N/A",
		"*tcd": "N/A",
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
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
	expected := map[string]interface{}{
		"cores": map[string]interface{}{
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
	var rcv map[string]interface{}
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
	expected := []interface{}{
		map[string]interface{}{
			"Tenant":    "cgrates.org",
			"ID":        "ROUTE_ACNT_1001",
			"FilterIDs": []interface{}{"FLTR_ACNT_1001"},
			"ActivationInterval": map[string]interface{}{
				"ActivationTime": "2017-11-27T00:00:00Z",
				"ExpiryTime":     "0001-01-01T00:00:00Z",
			},
			"Sorting":           "*weight",
			"SortingParameters": []interface{}{},
			"Routes": []interface{}{
				map[string]interface{}{
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
				map[string]interface{}{
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
	var rcv []interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	sort.Slice(rcv[0].(map[string]interface{})["Routes"], func(i, j int) bool {
		return utils.IfaceAsString(rcv[0].(map[string]interface{})["Routes"].([]interface{})[i].(map[string]interface{})["ID"]) < utils.IfaceAsString(rcv[0].(map[string]interface{})["Routes"].([]interface{})[j].(map[string]interface{})["ID"])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %s \n but received \n %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItStatsProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "stats_profile", `Tenant="cgrates.org"`, `ID="Stats2"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]interface{}{
		"ActivationInterval": map[string]interface{}{
			"ActivationTime": "2014-07-29T15:00:00Z",
			"ExpiryTime":     "0001-01-01T00:00:00Z",
		},
		"Blocker":   true,
		"FilterIDs": []interface{}{"FLTR_ACNT_1001_1002"},
		"ID":        "Stats2",
		"Metrics": []interface{}{
			map[string]interface{}{
				"FilterIDs": nil,
				"MetricID":  "*tcc",
			},
			map[string]interface{}{
				"FilterIDs": nil,
				"MetricID":  "*tcd",
			},
		},
		"MinItems":     0.,
		"QueueLength":  100.,
		"Stored":       false,
		"TTL":          "-1ns",
		"Tenant":       "cgrates.org",
		"ThresholdIDs": []interface{}{"*none"},
		"Weight":       30.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(err)
	}
	sort.Slice(rcv["Metrics"].([]interface{}), func(i, j int) bool {
		return utils.IfaceAsString((rcv["Metrics"].([]interface{})[i].(map[string]interface{}))["MetricID"]) < utils.IfaceAsString((rcv["Metrics"].([]interface{})[j].(map[string]interface{}))["MetricID"])
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
		map[string]interface{}{
			"Tenant":    "cgrates.org",
			"ID":        "ROUTE_ACNT_1001",
			"FilterIDs": []interface{}{"FLTR_ACNT_1001"},
			"ActivationInterval": map[string]interface{}{
				"ActivationTime": "2017-11-27T00:00:00Z",
				"ExpiryTime":     "0001-01-01T00:00:00Z",
			},
			"Sorting":           "*weight",
			"SortingParameters": []interface{}{},
			"Routes": []interface{}{
				map[string]interface{}{
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
				map[string]interface{}{
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
	var rcv map[string]interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Error(err)
	}
	sort.Slice(rcv["Routes"], func(i, j int) bool {
		return utils.IfaceAsString(rcv["Routes"].([]interface{})[i].(map[string]interface{})["ID"]) < utils.IfaceAsString(rcv["Routes"].([]interface{})[j].(map[string]interface{})["ID"])
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
	expected := map[string]interface{}{
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
	var rcv map[string]interface{}
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
	expected := map[string]interface{}{
		"ActionIDs": []interface{}{"ACT_LOG_WARNING"},
		"ActivationInterval": map[string]interface{}{
			"ActivationTime": "2014-07-29T15:00:00Z",
			"ExpiryTime":     "0001-01-01T00:00:00Z",
		},
		"Async":     true,
		"Blocker":   false,
		"FilterIDs": []interface{}{"FLTR_ACNT_1001"},
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
	var rcv map[string]interface{}
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
	expected := []interface{}{":1001", "call:1001", "call:1002", "call:1003", "mms:*any", "sms:*any"}
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
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
	expected := []interface{}{"123", "Stats2", "Stats2_1"}
	cmd.Stdout = output
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
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
	expected := map[string]interface{}{
		"*account_action_plans": map[string]interface{}{
			"Items":  1.,
			"Groups": 0.,
		},
		"*accounts": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*action_plans": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*action_triggers": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*actions": map[string]interface{}{
			"Groups": 0.,
			"Items":  1.,
		},
		"*apiban": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*attribute_filter_indexes": map[string]interface{}{
			"Items":  10.,
			"Groups": 2.,
		},
		"*attribute_profiles": map[string]interface{}{
			"Items":  1.,
			"Groups": 0.,
		},
		"*caps_events": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*cdr_ids": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*cdrs": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*charger_filter_indexes": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*charger_profiles": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*closed_sessions": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*default": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*destinations": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*diameter_messages": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_filter_indexes": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_hosts": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_loads": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_profiles": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatcher_routes": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*dispatchers": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*event_charges": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*event_resources": map[string]interface{}{
			"Items":  1.,
			"Groups": 0.,
		},
		"*filters": map[string]interface{}{
			"Items":  4.,
			"Groups": 0.,
		},
		"*load_ids": map[string]interface{}{
			"Items":  13.,
			"Groups": 0.,
		},
		"*rating_plans": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*rating_profiles": map[string]interface{}{
			"Items":  1.,
			"Groups": 0.,
		},
		"*replication_hosts": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*resource_filter_indexes": map[string]interface{}{
			"Items":  2.,
			"Groups": 1.,
		},
		"*resource_profiles": map[string]interface{}{
			"Items":  2.,
			"Groups": 0.,
		},
		"*resources": map[string]interface{}{
			"Items":  2.,
			"Groups": 0.,
		},
		"*reverse_destinations": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*reverse_filter_indexes": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*route_filter_indexes": map[string]interface{}{
			"Items":  3.,
			"Groups": 1.,
		},
		"*route_profiles": map[string]interface{}{
			"Items":  1.,
			"Groups": 0.,
		},
		"*rpc_connections": map[string]interface{}{
			"Items":  3.,
			"Groups": 0.,
		},
		"*rpc_responses": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*session_costs": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*shared_groups": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*stat_filter_indexes": map[string]interface{}{
			"Items":  2.,
			"Groups": 1.,
		},
		"*statqueue_profiles": map[string]interface{}{
			"Items":  2.,
			"Groups": 0.,
		},
		"*statqueues": map[string]interface{}{
			"Items":  2.,
			"Groups": 0.,
		},
		"*stir": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*threshold_filter_indexes": map[string]interface{}{
			"Items":  9.,
			"Groups": 1.,
		},
		"*threshold_profiles": map[string]interface{}{
			"Items":  2.,
			"Groups": 0.,
		},
		"*thresholds": map[string]interface{}{
			"Items":  2.,
			"Groups": 0.,
		},
		"*timings": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tmp_rating_profiles": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_account_actions": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_action_plans": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_action_triggers": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_actions": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_attributes": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_chargers": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_destination_rates": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_destinations": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_dispatcher_hosts": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_dispatcher_profiles": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_filters": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_rates": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_rating_plans": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_rating_profiles": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_resources": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_routes": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_shared_groups": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*tp_stats": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_thresholds": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*tp_timings": map[string]interface{}{
			"Groups": 0.,
			"Items":  0.,
		},
		"*uch": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
		"*versions": map[string]interface{}{
			"Items":  0.,
			"Groups": 0.,
		},
	}

	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
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
	expected := []interface{}{
		map[string]interface{}{
			"Tenant": "cgrates.org",
			"ID":     "ResGroup1",
			"Usages": map[string]interface{}{},
			"TTLIdx": nil,
		},
		map[string]interface{}{
			"Tenant": "cgrates.org",
			"ID":     "123",
			"Usages": map[string]interface{}{},
			"TTLIdx": nil,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
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
	expected := []interface{}{"ATTR_1001_SESSIONAUTH", "ATTR_1001_SIMPLEAUTH", "ATTR_1002_SESSIONAUTH", "ATTR_1002_SIMPLEAUTH", "ATTR_1003_SESSIONAUTH", "ATTR_1003_SIMPLEAUTH", "ATTR_ACC_ALIAS"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
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
	expected := []interface{}{"123", "THD_ACNT_1001"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
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
	expected := map[string]interface{}{
		"Tenant": "cgrates.org",
		"ID":     "ResGroup1",
		"Usages": map[string]interface{}{},
		"TTLIdx": nil,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
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
	expected := map[string]interface{}{
		"ActivationInterval": map[string]interface{}{
			"ActivationTime": "2014-07-29T15:00:00Z",
			"ExpiryTime":     "0001-01-01T00:00:00Z",
		},
		"AllocationMessage": "",
		"Blocker":           false,
		"FilterIDs":         []interface{}{"FLTR_RES"},
		"ID":                "ResGroup1",
		"Limit":             7.,
		"Stored":            true,
		"Tenant":            "cgrates.org",
		"ThresholdIDs":      []interface{}{"*none"},
		"UsageTTL":          "-1ns",
		"Weight":            10.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
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
	expected := []interface{}{
		map[string]interface{}{
			"ProfileID": "ROUTE_ACNT_1001",
			"Sorting":   "*weight",
			"Routes": []interface{}{
				map[string]interface{}{
					"RouteID":         "route2",
					"RouteParameters": "",
					"SortingData": map[string]interface{}{
						"Weight": 20.,
					},
				},
				map[string]interface{}{
					"RouteID":         "route1",
					"RouteParameters": "",
					"SortingData": map[string]interface{}{
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
	var rcv []interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		fmt.Println(utils.IfaceAsString((rcv[0].(map[string]interface{})["Routes"].([]interface{})[i].(map[string]interface{})["RouteID"])))
		return utils.IfaceAsString((rcv[0].(map[string]interface{})["Routes"].([]interface{})[i].(map[string]interface{})["RouteID"])) < utils.IfaceAsString((rcv[0].(map[string]interface{})["Routes"].([]interface{})[j].(map[string]interface{})["RouteID"]))
		// return utils.IfaceAsString((rcv["Metrics"].([]interface{})[i].(map[string]interface{}))["MetricID"]) < utils.IfaceAsString((rcv["Metrics"].([]interface{})[j].(map[string]interface{}))["MetricID"])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItFilter(t *testing.T) {
	cmd := exec.Command("cgr-console", "filter", `ID="FLTR_RES"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]interface{}{
		"ActivationInterval": map[string]interface{}{
			"ActivationTime": "2014-07-29T15:00:00Z",
			"ExpiryTime":     "0001-01-01T00:00:00Z",
		},
		"Tenant": "cgrates.org",
		"ID":     "FLTR_RES",
		"Rules": []interface{}{
			map[string]interface{}{
				"Type":    "*string",
				"Element": "~*req.Account",
				"Values":  []interface{}{"1001", "1002", "1003"},
			},
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
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
	expected := []interface{}{
		map[string]interface{}{
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
	var rcv []interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv[0].(map[string]interface{})["Snooze"] = ""

	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+v \n but received \n %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItStatsForEvent(t *testing.T) {
	cmd := exec.Command("cgr-console", "stats_for_event", `Tenant="cgrates.org"`, `ID="Stats2"`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []interface{}{"123"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
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
	expected := []interface{}{"123"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
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
	expected := map[string]interface{}{
		"Attributes": map[string]interface{}{
			"APIOpts": map[string]interface{}{
				"*subsys": "*sessions",
			},
			"AlteredFields": []interface{}{"*req.LCRProfile", "*req.Password", "*req.PaypalAccount", "*req.RequestType"},
			"Event": map[string]interface{}{
				"Account":       "1001",
				"LCRProfile":    "premium_cli",
				"Password":      "CGRateS.org",
				"PaypalAccount": "cgrates@paypal.com",
				"RequestType":   "*prepaid",
			},
			"ID":              nil,
			"MatchedProfiles": []interface{}{"ATTR_1001_SESSIONAUTH"},
			"Tenant":          "cgrates.org",
			"Time":            nil,
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv["Attributes"].(map[string]interface{})["ID"] = nil
	rcv["Attributes"].(map[string]interface{})["Time"] = nil
	sort.Slice(rcv["Attributes"].(map[string]interface{})["AlteredFields"], func(i, j int) bool {
		return utils.IfaceAsString(rcv["Attributes"].(map[string]interface{})["AlteredFields"].([]interface{})[i]) < utils.IfaceAsString(rcv["Attributes"].(map[string]interface{})["AlteredFields"].([]interface{})[j])
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
	expected := map[string]interface{}{
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
	var rcv map[string]interface{}
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
	expected := map[string]interface{}{
		"Tenant":             "cgrates.org",
		"ID":                 "DEFAULT",
		"FilterIDs":          []interface{}{},
		"ActivationInterval": nil,
		"RunID":              "*default",
		"AttributeIDs":       []interface{}{"*none"},
		"Weight":             0.,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
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
	expected := map[string]interface{}{
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
	var rcv map[string]interface{}
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
	var rcv map[string]interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	expected := map[string]interface{}{
		"*account_action_plans":     rcv["*account_action_plans"],
		"*action_plans":             rcv["*action_plans"],
		"*action_triggers":          rcv["*action_triggers"],
		"*actions":                  rcv["*actions"],
		"*attribute_filter_indexes": rcv["*attribute_filter_indexes"],
		"*attribute_profiles":       rcv["*attribute_profiles"],
		"*charger_profiles":         rcv["*charger_profiles"],
		"*destinations":             rcv["*destinations"],
		"*dispatcher_profiles":      rcv["*dispatcher_profiles"],
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
	expected := []interface{}{
		map[string]interface{}{
			"Tenant":             "cgrates.org",
			"ID":                 "DEFAULT",
			"FilterIDs":          []interface{}{},
			"ActivationInterval": nil,
			"RunID":              "*default",
			"AttributeIDs":       []interface{}{"*none"},
			"Weight":             0.,
		},
		map[string]interface{}{
			"Tenant":             "cgrates.org",
			"ID":                 "Raw",
			"FilterIDs":          []interface{}{},
			"ActivationInterval": nil,
			"RunID":              "*raw",
			"AttributeIDs":       []interface{}{"*constant:*req.RequestType:*none"},
			"Weight":             0.,
		},
		map[string]interface{}{
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
	var rcv []interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return utils.IfaceAsString(rcv[i].(map[string]interface{})["ID"]) < utils.IfaceAsString(rcv[j].(map[string]interface{})["ID"])
	})
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItAccounts(t *testing.T) {
	cmd := exec.Command("cgr-console", "accounts", `AccountIDs=["1001"]`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []interface{}{
		map[string]interface{}{
			"ActionTriggers": nil,
			"AllowNegative":  false,
			"BalanceMap": map[string]interface{}{
				"*monetary": []interface{}{
					map[string]interface{}{
						"Blocker":        false,
						"Categories":     map[string]interface{}{},
						"DestinationIDs": map[string]interface{}{},
						"Disabled":       false,
						"ExpirationDate": "0001-01-01T00:00:00Z",
						"Factor":         nil,
						"ID":             "test",
						"RatingSubject":  "",
						"SharedGroups":   map[string]interface{}{},
						"TimingIDs":      map[string]interface{}{},
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
	var rcv []interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv[0].(map[string]interface{})["BalanceMap"].(map[string]interface{})["*monetary"].([]interface{})[0].(map[string]interface{})["Uuid"] = ""
	rcv[0].(map[string]interface{})["UpdateTime"] = ""
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
	expected := map[string]interface{}{
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
	var rcv map[string]interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItSessionInitiate(t *testing.T) {
	cmd := exec.Command("cgr-console", "session_initiate", `GetAttributes=true`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]interface{}{
		"AttributesDigest":   "LCRProfile:premium_cli,Password:CGRateS.org,PaypalAccount:cgrates@paypal.com,RequestType:*prepaid",
		"MaxUsage":           "0s",
		"ResourceAllocation": nil,
		"StatQueues":         nil,
		"Thresholds":         nil,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
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

func testConsoleItCachePrecacheStatus(t *testing.T) {
	cmd := exec.Command("cgr-console", "cache_precache_status")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]interface{}{
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
	var rcv map[string]interface{}
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
	expected := []interface{}{"DEFAULT", "Raw", "cps"}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
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
	expected := []interface{}{
		map[string]interface{}{
			"ChargerSProfile":    "DEFAULT",
			"AttributeSProfiles": nil,
			"AlteredFields":      []interface{}{"*req.RunID"},
			"CGREvent": map[string]interface{}{
				"APIOpts": map[string]interface{}{
					"*subsys": "*chargers",
				},
				"Tenant": "cgrates.org",
				"ID":     "",
				"Time":   "",
				"Event": map[string]interface{}{
					"Account": "1001",
					"RunID":   "*default",
				},
			},
		},
		map[string]interface{}{
			"ChargerSProfile":    "Raw",
			"AttributeSProfiles": []interface{}{"*constant:*req.RequestType:*none"},
			"AlteredFields":      []interface{}{"*req.RunID", "*req.RequestType"},
			"CGREvent": map[string]interface{}{
				"APIOpts": map[string]interface{}{
					"*subsys": "*chargers",
				},
				"Tenant": "cgrates.org",
				"ID":     "",
				"Time":   "",
				"Event": map[string]interface{}{
					"Account":     "1001",
					"RequestType": "*none",
					"RunID":       "*raw",
				},
			},
		},
		map[string]interface{}{
			"ChargerSProfile":    "cps",
			"AlteredFields":      []interface{}{"*req.RunID"},
			"AttributeSProfiles": []interface{}{},
			"CGREvent": map[string]interface{}{
				"APIOpts": map[string]interface{}{
					"*subsys": "*chargers",
				},
				"Event": map[string]interface{}{
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
	var rcv []interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return utils.IfaceAsString(rcv[i].(map[string]interface{})["ChargerSProfile"]) < utils.IfaceAsString(rcv[j].(map[string]interface{})["ChargerSProfile"])
	})
	rcv[0].(map[string]interface{})["CGREvent"].(map[string]interface{})["Time"] = ""
	rcv[1].(map[string]interface{})["CGREvent"].(map[string]interface{})["Time"] = ""
	rcv[2].(map[string]interface{})["CGREvent"].(map[string]interface{})["Time"] = ""
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItSessionProcessMessage(t *testing.T) {
	cmd := exec.Command("cgr-console", "session_process_message", `GetAttributes=true`, `Event={"Account":"1001"}`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]interface{}{
		"Attributes": map[string]interface{}{
			"APIOpts": map[string]interface{}{
				"*subsys": "*sessions",
			},
			"AlteredFields": []interface{}{"*req.LCRProfile", "*req.Password", "*req.PaypalAccount", "*req.RequestType"},
			"Event": map[string]interface{}{
				"Account":       "1001",
				"LCRProfile":    "premium_cli",
				"Password":      "CGRateS.org",
				"PaypalAccount": "cgrates@paypal.com",
				"RequestType":   "*prepaid",
			},
			"ID":              "",
			"MatchedProfiles": []interface{}{"ATTR_1001_SESSIONAUTH"},
			"Tenant":          "cgrates.org",
			"Time":            "",
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Log(output.String())
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv["Attributes"].(map[string]interface{})["AlteredFields"], func(i, j int) bool {
		return utils.IfaceAsString(rcv["Attributes"].(map[string]interface{})["AlteredFields"].([]interface{})[i]) < utils.IfaceAsString(rcv["Attributes"].(map[string]interface{})["AlteredFields"].([]interface{})[j])
	})
	rcv["Attributes"].(map[string]interface{})["ID"] = ""
	rcv["Attributes"].(map[string]interface{})["Time"] = ""
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
	expected := []interface{}{
		map[string]interface{}{
			"ActionsID":      "",
			"ActivationDate": "0001-01-01T00:00:00Z",
			"Balance": map[string]interface{}{
				"Blocker":        nil,
				"Categories":     nil,
				"DestinationIDs": nil,
				"Disabled":       nil,
				"ExpirationDate": nil,
				"Factor":         nil,
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
	var rcv []interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv[0].(map[string]interface{})["UniqueID"] = ""
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
	}
}

func testConsoleItRatingProfile(t *testing.T) {
	cmd := exec.Command("cgr-console", "ratingprofile", `Category="call"`, `Subject="1001"`)
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := map[string]interface{}{
		"Id": "*out:cgrates.org:call:1001",
		"RatingPlanActivations": []interface{}{
			map[string]interface{}{
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
	var rcv map[string]interface{}
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
	expected := []interface{}{
		map[string]interface{}{
			"Id":         "AP_PACKAGE_10",
			"AccountIDs": map[string]interface{}{},
			"ActionTimings": []interface{}{
				map[string]interface{}{
					"Uuid": "",
					"Timing": map[string]interface{}{
						"Timing": map[string]interface{}{
							"ID":        "*asap",
							"Years":     []interface{}{},
							"Months":    []interface{}{},
							"MonthDays": []interface{}{},
							"WeekDays":  []interface{}{},
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
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv[0].(map[string]interface{})["ActionTimings"].([]interface{})[0].(map[string]interface{})["Uuid"] = ""
	if !reflect.DeepEqual(rcv, expected) {
		t.Fatalf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func testConsoleItDestinations(t *testing.T) {
	cmd := exec.Command("cgr-console", "destinations")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	expected := []interface{}{
		map[string]interface{}{
			"Id":       "DST_1001",
			"Prefixes": []interface{}{"1001"},
		},
		map[string]interface{}{
			"Id":       "DST_1002",
			"Prefixes": []interface{}{"1002"},
		},
		map[string]interface{}{
			"Id":       "DST_1003",
			"Prefixes": []interface{}{"1003"},
		},
		map[string]interface{}{
			"Id":       "DST_FS",
			"Prefixes": []interface{}{"10"},
		},
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv []interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	sort.Slice(rcv, func(i, j int) bool {
		return rcv[i].(map[string]interface{})["Id"].(string) < rcv[j].(map[string]interface{})["Id"].(string)
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
	expected := map[string]interface{}{
		"Account":     "",
		"Category":    "call",
		"Cost":        0.,
		"DataSpans":   []interface{}{},
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
	var rcv map[string]interface{}
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
	expected := map[string]interface{}{
		"Id":                "",
		"AccountParameters": nil,
		"MemberIds":         nil,
	}
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Fatal(err)
	}
	var rcv map[string]interface{}
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
	var rcv map[string]interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	expected := map[string]interface{}{
		"*account_action_plans":     rcv["*account_action_plans"],
		"*action_plans":             rcv["*action_plans"],
		"*action_triggers":          rcv["*action_triggers"],
		"*actions":                  rcv["*actions"],
		"*attribute_filter_indexes": rcv["*attribute_filter_indexes"],
		"*attribute_profiles":       rcv["*attribute_profiles"],
		"*charger_profiles":         rcv["*charger_profiles"],
		"*destinations":             rcv["*destinations"],
		"*dispatcher_profiles":      rcv["*dispatcher_profiles"],
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
		t.Fatalf("Expected %+q \n but received \n %+q", expected, rcv)
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
	expected := map[string]interface{}{
		"Category":    "call",
		"Tenant":      "cgrates.org",
		"Subject":     "1001",
		"Account":     "1001",
		"Destination": "1002",
		"ToR":         "",
		"Cost":        0.,
		"Timespans":   nil,
		"RatedUsage":  0.,
		"AccountSummary": map[string]interface{}{
			"Tenant": "cgrates.org",
			"ID":     "1001",
			"BalanceSummaries": []interface{}{
				map[string]interface{}{
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
	var rcv map[string]interface{}
	if err := json.NewDecoder(output).Decode(&rcv); err != nil {
		t.Error(output.String())
		t.Fatal(err)
	}
	rcv["AccountSummary"].(map[string]interface{})["BalanceSummaries"].([]interface{})[0].(map[string]interface{})["UUID"] = ""
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

func reverseFormatting(i interface{}, s utils.StringSet) (_ interface{}, err error) {
	switch v := i.(type) {
	case map[string]interface{}:
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
	case []interface{}:
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
	var tmp map[string]interface{}
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

func testConsoleItKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Fatal(err)
	}
}
