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
