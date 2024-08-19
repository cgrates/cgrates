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

package engine

import (
	"testing"

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
		utils.CacheFilters,
		utils.CacheRatingPlans,
		utils.CacheRatingProfiles,
		utils.CacheResourceFilterIndexes,
		utils.CacheResourceProfiles,
		utils.CacheResources,
		utils.CacheReverseDestinations,
		utils.CacheRPCResponses,
		utils.CacheSharedGroups,
		utils.CacheStatFilterIndexes,
		utils.CacheStatQueueProfiles,
		utils.CacheStatQueues,
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
		t.Errorf("expected no error, got %v", err)
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
