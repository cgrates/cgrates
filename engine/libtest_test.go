//go:build integration || flaky || offline || kafka || call || race || performance

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestGetDefaultEmptyCacheStats(t *testing.T) {
	cacheStats := GetDefaultEmptyCacheStats()
	expectedKeys := []string{
		utils.MetaDefault,
		utils.CacheAccountActionPlans,
		utils.CacheActionPlans,
		utils.CacheActionTriggers,
		utils.CacheActions,
		utils.CacheAttributeFilterIndexes,
		utils.CacheAttributeProfiles,
		utils.CacheChargerFilterIndexes,
		utils.CacheChargerProfiles,
		utils.CacheDispatcherFilterIndexes,
		utils.CacheDispatcherProfiles,
		utils.CacheDispatcherHosts,
		utils.CacheDispatcherRoutes,
		utils.CacheDispatcherLoads,
		utils.CacheDispatchers,
		utils.CacheDestinations,
		utils.CacheEventResources,
		utils.CacheEventIPs,
		utils.CacheFilters,
		utils.CacheRatingPlans,
		utils.CacheRatingProfiles,
		utils.CacheResourceFilterIndexes,
		utils.CacheResourceProfiles,
		utils.CacheResources,
		utils.CacheIPFilterIndexes,
		utils.CacheIPProfiles,
		utils.CacheIPAllocations,
		utils.CacheReverseDestinations,
		utils.CacheRPCResponses,
		utils.CacheSharedGroups,
		utils.CacheStatFilterIndexes,
		utils.CacheStatQueueProfiles,
		utils.CacheStatQueues,
		utils.CacheRankings,
		utils.CacheRankingProfiles,
		utils.CacheSTIR,
		utils.CacheRouteFilterIndexes,
		utils.CacheRouteProfiles,
		utils.CacheThresholdFilterIndexes,
		utils.CacheThresholdProfiles,
		utils.CacheThresholds,
		utils.CacheTimings,
		utils.CacheDiameterMessages,
		utils.CacheClosedSessions,
		utils.CacheLoadIDs,
		utils.CacheRPCConnections,
		utils.CacheCDRIDs,
		utils.CacheRatingProfilesTmp,
		utils.CacheUCH,
		utils.CacheEventCharges,
		utils.CacheTrendProfiles,
		utils.CacheTrends,
		utils.CacheReverseFilterIndexes,
		utils.MetaAPIBan,
		utils.MetaSentryPeer,
		utils.CacheCapsEvents,
		utils.CacheReplicationHosts,
		utils.CacheRadiusPackets,
	}

	if len(cacheStats) != len(expectedKeys) {
		t.Errorf("expected %d keys, got %d", len(expectedKeys), len(cacheStats))
	}
	for _, key := range expectedKeys {
		if _, exists := cacheStats[key]; !exists {
			t.Errorf("expected key %s to be present in the map", key)
		}
	}
}

func TestKillEngine(t *testing.T) {
	err := KillEngine(10)
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
	err = KillEngine(-1)
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func TestStopStartEngine(t *testing.T) {
	cmd, err := StopStartEngine("valid/path/to/config", 10)
	if err == nil {
		t.Errorf("expected error, got %v", err)
	}
	if cmd != nil {
		t.Errorf("expected a valid command, got nil")
	}
	cmd, err = StopStartEngine("valid/path/to/config", -1)
	if err == nil {
		t.Errorf("expected an error from KillEngine, got nil")
	}
	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}

	cmd, err = StopStartEngine("", 10)
	if err == nil {
		t.Errorf("expected an error from StartEngine, got nil")
	}
	if cmd != nil {
		t.Errorf("expected no command, got %v", cmd)
	}
}

func TestCallScript(t *testing.T) {
	err := CallScript("/path/to/nonexistent/script.sh", "arg", 100)
	if err == nil {
		t.Errorf("Expected an error when calling a non-existent script, got nil")
	}
	err = CallScript("/bin/echo", "script", 100)
	if err != nil {
		t.Errorf("Expected no error when calling a valid command, got: %v", err)
	}
	start := time.Now()
	err = CallScript("/bin/echo", "script", 100)
	duration := time.Since(start)
	if duration < 100*time.Millisecond {
		t.Errorf("Expected the delay to be at least 100ms, got %v", duration)
	}
}

func TestForceKillProc(t *testing.T) {

	cmd := exec.Command("sleep", "5")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start the sleep process: %v", err)
	}
	defer cmd.Process.Kill()
	err := ForceKillProcName("sleep", 100)
	if err != nil {
		t.Errorf("Expected no error when trying to kill an existing process, got: %v", err)
	}
	start := time.Now()
	err = ForceKillProcName("sleep", 100)
	if err != nil {
		t.Errorf("Expected no error when trying to kill an existing process, got: %v", err)
	}
	duration := time.Since(start)
	if duration < 100*time.Millisecond {
		t.Errorf("Expected the delay to be at least 100ms, got %v", duration)
	}
}

func TestRoutesIDs(t *testing.T) {
	t.Run("returns empty slice when SortedRoutesList is empty", func(t *testing.T) {
		var sRs SortedRoutesList
		routeIDs := sRs.RouteIDs()

		if len(routeIDs) != 0 {
			t.Errorf("expected empty slice, got %v", routeIDs)
		}
	})

	t.Run("returns single RouteID when there is one route", func(t *testing.T) {
		sRs := SortedRoutesList{
			{
				Routes: []*SortedRoute{
					{RouteID: "route1"},
				},
			},
		}
		routeIDs := sRs.RouteIDs()

		expected := []string{"route1"}
		if !reflect.DeepEqual(routeIDs, expected) {
			t.Errorf("expected %v, got %v", expected, routeIDs)
		}
	})

	t.Run("returns multiple RouteIDs for multiple routes", func(t *testing.T) {
		sRs := SortedRoutesList{
			{
				Routes: []*SortedRoute{
					{RouteID: "1001"},
					{RouteID: "1002"},
				},
			},
			{
				Routes: []*SortedRoute{
					{RouteID: "1003"},
					{RouteID: "1004"},
				},
			},
		}
		routeIDs := sRs.RouteIDs()

		expected := []string{"1001", "1002", "1003", "1004"}
		if !reflect.DeepEqual(routeIDs, expected) {
			t.Errorf("expected %v, got %v", expected, routeIDs)
		}
	})
}
