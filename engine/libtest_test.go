// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestGetDefaultEmptyCacheStats(t *testing.T) {
	expected := map[string]*ltcache.CacheStats{
		utils.MetaDefault:                  {Items: 0, Groups: 0},
		utils.CacheAccountActionPlans:      {Items: 0, Groups: 0},
		utils.CacheActionPlans:             {Items: 0, Groups: 0},
		utils.CacheActionTriggers:          {Items: 0, Groups: 0},
		utils.CacheActions:                 {Items: 0, Groups: 0},
		utils.CacheAttributeFilterIndexes:  {Items: 0, Groups: 0},
		utils.CacheAttributeProfiles:       {Items: 0, Groups: 0},
		utils.CacheChargerFilterIndexes:    {Items: 0, Groups: 0},
		utils.CacheChargerProfiles:         {Items: 0, Groups: 0},
		utils.CacheDispatcherFilterIndexes: {Items: 0, Groups: 0},
		utils.CacheReverseFilterIndexes:    {Items: 0, Groups: 0},
		utils.CacheDispatcherProfiles:      {Items: 0, Groups: 0},
		utils.CacheDispatcherHosts:         {Items: 0, Groups: 0},
		utils.CacheDispatcherRoutes:        {Items: 0, Groups: 0},
		utils.CacheDestinations:            {Items: 0, Groups: 0},
		utils.CacheEventResources:          {Items: 0, Groups: 0},
		utils.CacheFilters:                 {Items: 0, Groups: 0},
		utils.CacheRatingPlans:             {Items: 0, Groups: 0},
		utils.CacheRatingProfiles:          {Items: 0, Groups: 0},
		utils.CacheResourceFilterIndexes:   {Items: 0, Groups: 0},
		utils.CacheResourceProfiles:        {Items: 0, Groups: 0},
		utils.CacheResources:               {Items: 0, Groups: 0},
		utils.CacheReverseDestinations:     {Items: 0, Groups: 0},
		utils.CacheRPCResponses:            {Items: 0, Groups: 0},
		utils.CacheSharedGroups:            {Items: 0, Groups: 0},
		utils.CacheStatFilterIndexes:       {Items: 0, Groups: 0},
		utils.CacheStatQueueProfiles:       {Items: 0, Groups: 0},
		utils.CacheStatQueues:              {Items: 0, Groups: 0},
		utils.CacheSupplierFilterIndexes:   {Items: 0, Groups: 0},
		utils.CacheSupplierProfiles:        {Items: 0, Groups: 0},
		utils.CacheThresholdFilterIndexes:  {Items: 0, Groups: 0},
		utils.CacheThresholdProfiles:       {Items: 0, Groups: 0},
		utils.CacheThresholds:              {Items: 0, Groups: 0},
		utils.CacheTimings:                 {Items: 0, Groups: 0},
		utils.CacheDiameterMessages:        {Items: 0, Groups: 0},
		utils.CacheClosedSessions:          {Items: 0, Groups: 0},
		utils.CacheLoadIDs:                 {Items: 0, Groups: 0},
		utils.CacheRPCConnections:          {Items: 0, Groups: 0},
		utils.CacheCDRIDs:                  {Items: 0, Groups: 0},
		utils.CacheRatingProfilesTmp:       {Items: 0, Groups: 0},
	}
	result := GetDefaultEmptyCacheStats()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
