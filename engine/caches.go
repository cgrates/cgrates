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
	"fmt"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var Cache *ltcache.TransCache

func init() {
	InitCache(nil)
}

var precachedPartitions = []string{
	utils.CacheDestinations,
	utils.CacheReverseDestinations,
	utils.CacheRatingPlans,
	utils.CacheRatingProfiles,
	utils.CacheActions,
	utils.CacheActionPlans,
	utils.CacheAccountActionPlans,
	utils.CacheActionTriggers,
	utils.CacheSharedGroups,
	utils.CacheAliases,
	utils.CacheReverseAliases,
	utils.CacheDerivedChargers,
	utils.CacheResourceProfiles,
	utils.CacheResources,
	utils.CacheEventResources,
	utils.CacheTimings,
	utils.CacheStatQueueProfiles,
	utils.CacheStatQueues,
	utils.CacheThresholdProfiles,
	utils.CacheThresholds,
	utils.CacheFilters,
	utils.CacheSupplierProfiles,
	utils.CacheAttributeProfiles,
	utils.CacheChargerProfiles,
}

// InitCache will instantiate the cache with specific or default configuraiton
func InitCache(cfg config.CacheCfg) {
	if cfg == nil {
		cfg = config.CgrConfig().CacheCfg()
	}
	Cache = ltcache.NewTransCache(cfg.AsTransCacheConfig())
}

// NewCacheS initializes the Cache service
func NewCacheS(cfg *config.CGRConfig, dm *DataManager) (c *CacheS) {
	InitCache(cfg.CacheCfg()) // to make sure we start with correct config
	c = &CacheS{cfg: cfg, dm: dm,
		pcItems: make(map[string]chan struct{})}
	for cacheID := range cfg.CacheCfg() {
		if !utils.IsSliceMember(precachedPartitions, cacheID) {
			continue
		}
		c.pcItems[cacheID] = make(chan struct{})
	}
	return
}

// CacheS deals with cache preload and other cache related tasks/APIs
type CacheS struct {
	cfg     *config.CGRConfig
	dm      *DataManager
	pcItems map[string]chan struct{} // signal precaching
}

// GetChannel returns the channel used to signal precaching
func (chS *CacheS) GetPrecacheChannel(chID string) chan struct{} {
	return chS.pcItems[chID]
}

// Precache loads data from DataDB into cache at engine start
func (chS *CacheS) Precache() (err error) {
	for cacheID, cacheCfg := range chS.cfg.CacheCfg() {
		if !utils.IsSliceMember(precachedPartitions, cacheID) {
			continue
		}
		if cacheCfg.Precache {
			if err = chS.dm.CacheDataFromDB(
				utils.CacheInstanceToPrefix[cacheID], nil,
				false); err != nil {
				return
			}
		}
		close(chS.pcItems[cacheID])
	}
	return
}

type ArgsGetCacheItemIDs struct {
	CacheID      string
	ItemIDPrefix string
}

func (chS *CacheS) V1GetItemIDs(args *ArgsGetCacheItemIDs,
	reply *[]string) (err error) {
	if itmIDs := Cache.GetItemIDs(args.CacheID, args.ItemIDPrefix); len(itmIDs) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = itmIDs
	}
	return
}

type ArgsGetCacheItem struct {
	CacheID string
	ItemID  string
}

func (chS *CacheS) V1HasItem(args *ArgsGetCacheItem,
	reply *bool) (err error) {
	*reply = Cache.HasItem(args.CacheID, args.ItemID)
	return
}

func (chS *CacheS) V1GetItemExpiryTime(args *ArgsGetCacheItem,
	reply *time.Time) (err error) {
	if expTime, has := Cache.GetItemExpiryTime(args.CacheID, args.ItemID); !has {
		return utils.ErrNotFound
	} else {
		*reply = expTime
	}
	return
}

func (chS *CacheS) V1RemoveItem(args *ArgsGetCacheItem,
	reply *string) (err error) {
	Cache.Remove(args.CacheID, args.ItemID, true, utils.NonTransactional)
	*reply = utils.OK
	return
}

func (chS *CacheS) V1Clear(cacheIDs []string,
	reply *string) (err error) {
	Cache.Clear(cacheIDs)
	*reply = utils.OK
	return
}

func (chS *CacheS) V1GetCacheStats(cacheIDs []string,
	rply *map[string]*ltcache.CacheStats) (err error) {
	cs := Cache.GetCacheStats(cacheIDs)
	*rply = cs
	return
}

func (chS *CacheS) V1PrecacheStatus(cacheIDs []string, rply *map[string]string) (err error) {
	if len(cacheIDs) == 0 {
		for _, cacheID := range precachedPartitions {
			cacheIDs = append(cacheIDs, cacheID)
		}
	}
	pCacheStatus := make(map[string]string)
	for _, cacheID := range cacheIDs {
		if _, has := chS.pcItems[cacheID]; !has {
			return fmt.Errorf("unknown cacheID: %s", cacheID)
		}
		select {
		case <-chS.GetPrecacheChannel(cacheID):
			pCacheStatus[cacheID] = utils.MetaReady
		default:
			pCacheStatus[cacheID] = utils.MetaPrecaching
		}
	}
	*rply = pCacheStatus
	return
}

type ArgsGetGroup struct {
	CacheID string
	GroupID string
}

func (chS *CacheS) V1HasGroup(args *ArgsGetGroup,
	rply *bool) (err error) {
	*rply = Cache.HasGroup(args.CacheID, args.GroupID)
	return
}

func (chS *CacheS) V1GetGroupItemIDs(args *ArgsGetGroup,
	rply *[]string) (err error) {
	if has := Cache.HasGroup(args.CacheID, args.GroupID); !has {
		return utils.ErrNotFound
	}
	*rply = Cache.GetGroupItemIDs(args.CacheID, args.GroupID)
	return
}

func (chS *CacheS) V1RemoveGroup(args *ArgsGetGroup,
	rply *string) (err error) {
	Cache.RemoveGroup(args.CacheID, args.GroupID, true, utils.NonTransactional)
	*rply = utils.OK
	return
}
