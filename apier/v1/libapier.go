// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// GetCacheOpt receive the apiOpt and compare with default value
// overwrite the default if it's present
// visible in APIerSv2
func GetCacheOpt(apiOpt *string) string {
	cacheOpt := config.CgrConfig().GeneralCfg().DefaultCaching
	if apiOpt != nil && *apiOpt != utils.EmptyString {
		cacheOpt = *apiOpt
	}
	return cacheOpt
}

// composeArgsReload add the ItemID to AttrReloadCache
// for a specific CacheID
func composeArgsReload(args utils.ArgsGetCacheItem) (rpl utils.AttrReloadCache) {
	rpl = utils.InitAttrReloadCache()
	switch args.CacheID {
	case utils.CacheResourceProfiles:
		rpl.ResourceProfileIDs = []string{args.ItemID}
	case utils.CacheResources:
		rpl.ResourceIDs = []string{args.ItemID}
	case utils.CacheStatQueues:
		rpl.StatsQueueIDs = []string{args.ItemID}
	case utils.CacheStatQueueProfiles:
		rpl.StatsQueueProfileIDs = []string{args.ItemID}
	case utils.CacheThresholds:
		rpl.ThresholdIDs = []string{args.ItemID}
	case utils.CacheThresholdProfiles:
		rpl.ThresholdProfileIDs = []string{args.ItemID}
	case utils.CacheFilters:
		rpl.FilterIDs = []string{args.ItemID}
	case utils.CacheSupplierProfiles:
		rpl.SupplierProfileIDs = []string{args.ItemID}
	case utils.CacheAttributeProfiles:
		rpl.AttributeProfileIDs = []string{args.ItemID}
	case utils.CacheChargerProfiles:
		rpl.ChargerProfileIDs = []string{args.ItemID}
	case utils.CacheDispatcherProfiles:
		rpl.DispatcherProfileIDs = []string{args.ItemID}
	case utils.CacheDispatcherHosts:
		rpl.DispatcherHostIDs = []string{args.ItemID}
	case utils.CacheRatingProfiles:
		rpl.RatingProfileIDs = []string{args.ItemID}
	}
	return
}
