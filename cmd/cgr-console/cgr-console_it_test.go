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
	"testing"

	"github.com/cgrates/cgrates/config"
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
		testConsoleItRatingProfileSet,
		testConsoleItRatingProfileIds,
		testConsoleItResources,
		testConsoleItResourcesProfileIds,
		testConsoleItResourcesProfile,
		testConsoleItResourcesRelease,
		testConsoleItResourcesProfileSet,
		testConsoleItResourcesForEvent,
		testConsoleItResourcesAllocate,

		testConsoleItResourcesAuthorize,
		testConsoleItRouteProfileIds,
		testConsoleItRoutesProfilesForEvent,
		testConsoleItRoutesProfile,
		testConsoleItRoutes,
		testConsoleItCacheReload,
		testConsoleItAttributesProfileIds,
		testConsoleItAttributesProfileSet,
		testConsoleItFilterIds,
		testConsoleItFilterSet,
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
		testConsoleItCacheRemoveItem,
		testConsoleItFilter,
		testConsoleItPing,
		testConsoleItSessionUpdate,
		testConsoleItCacheStats,
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

// func testConsoleItDispatchersProfileIds(t *testing.T) {

// }
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

func testConsoleItKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Fatal(err)
	}
}
