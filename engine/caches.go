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
	"encoding/gob"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// Cache is the global cache used
var Cache *ltcache.TransCache

func init() {
	InitCache(nil)

	gob.Register(new(EventCost))

	// StatMetrics
	gob.Register(new(StatASR))
	gob.Register(new(StatACD))
	gob.Register(new(StatTCD))
	gob.Register(new(StatACC))
	gob.Register(new(StatTCC))
	gob.Register(new(StatPDD))
	gob.Register(new(StatDDC))
	gob.Register(new(StatSum))
	gob.Register(new(StatAverage))
	gob.Register(new(StatDistinct))
}

// InitCache will instantiate the cache with specific or default configuraiton
func InitCache(cfg config.CacheCfg) {
	if cfg == nil {
		cfg = config.CgrConfig().CacheCfg()
	}
	cfg.AddTmpCaches()
	Cache = ltcache.NewTransCache(cfg.AsTransCacheConfig())
}

// NewCacheS initializes the Cache service and executes the precaching
func NewCacheS(cfg *config.CGRConfig, dm *DataManager) (c *CacheS) {
	InitCache(cfg.CacheCfg()) // to make sure we start with correct config
	c = &CacheS{cfg: cfg, dm: dm,
		pcItems: make(map[string]chan struct{})}
	for cacheID := range cfg.CacheCfg() {
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

// GetPrecacheChannel returns the channel used to signal precaching
func (chS *CacheS) GetPrecacheChannel(chID string) chan struct{} {
	return chS.pcItems[chID]
}

// Precache loads data from DataDB into cache at engine start
func (chS *CacheS) Precache() (err error) {
	var wg sync.WaitGroup // wait for precache to finish
	errChan := make(chan error)
	doneChan := make(chan struct{})
	for cacheID, cacheCfg := range chS.cfg.CacheCfg() {
		if !cacheCfg.Precache {
			close(chS.pcItems[cacheID]) // no need of precache
			continue
		}
		wg.Add(1)
		go func(cacheID string) {
			errCache := chS.dm.CacheDataFromDB(
				utils.CacheInstanceToPrefix[cacheID], nil,
				false)
			if errCache != nil {
				errChan <- fmt.Errorf("precaching cacheID <%s>, got error: %s", cacheID, errCache)
			}
			close(chS.pcItems[cacheID])
			wg.Done()
		}(cacheID)
	}
	runtime.Gosched() // switch context
	go func() {       // report wg.Wait on doneChan
		wg.Wait()
		close(doneChan)
	}()
	select {
	case err = <-errChan:
	case <-doneChan:
	}
	return
}

// APIs start here

// Call gives the ability of CacheS to be passed as internal RPC
func (chS *CacheS) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(chS, serviceMethod, args, reply)
}

func (chS *CacheS) V1GetItemIDs(args *utils.ArgsGetCacheItemIDsWithArgDispatcher,
	reply *[]string) (err error) {
	itmIDs := Cache.GetItemIDs(args.CacheID, args.ItemIDPrefix)
	if len(itmIDs) == 0 {
		return utils.ErrNotFound
	}
	*reply = itmIDs
	return
}

func (chS *CacheS) V1HasItem(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *bool) (err error) {
	*reply = Cache.HasItem(args.CacheID, args.ItemID)
	return
}

func (chS *CacheS) V1GetItemExpiryTime(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *time.Time) (err error) {
	expTime, has := Cache.GetItemExpiryTime(args.CacheID, args.ItemID)
	if !has {
		return utils.ErrNotFound
	}
	*reply = expTime
	return
}

func (chS *CacheS) V1RemoveItem(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *string) (err error) {
	Cache.Remove(args.CacheID, args.ItemID, true, utils.NonTransactional)
	*reply = utils.OK
	return
}

func (chS *CacheS) V1Clear(args *utils.AttrCacheIDsWithArgDispatcher,
	reply *string) (err error) {
	Cache.Clear(args.CacheIDs)
	*reply = utils.OK
	return
}

func (chS *CacheS) V1GetCacheStats(args *utils.AttrCacheIDsWithArgDispatcher,
	rply *map[string]*ltcache.CacheStats) (err error) {
	cs := Cache.GetCacheStats(args.CacheIDs)
	*rply = cs
	return
}

func (chS *CacheS) V1PrecacheStatus(args *utils.AttrCacheIDsWithArgDispatcher, rply *map[string]string) (err error) {
	if len(args.CacheIDs) == 0 {
		args.CacheIDs = utils.CachePartitions.AsSlice()
	}
	pCacheStatus := make(map[string]string)
	for _, cacheID := range args.CacheIDs {
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

func (chS *CacheS) V1HasGroup(args *utils.ArgsGetGroupWithArgDispatcher,
	rply *bool) (err error) {
	*rply = Cache.HasGroup(args.CacheID, args.GroupID)
	return
}

func (chS *CacheS) V1GetGroupItemIDs(args *utils.ArgsGetGroupWithArgDispatcher,
	rply *[]string) (err error) {
	if has := Cache.HasGroup(args.CacheID, args.GroupID); !has {
		return utils.ErrNotFound
	}
	*rply = Cache.GetGroupItemIDs(args.CacheID, args.GroupID)
	return
}

func (chS *CacheS) V1RemoveGroup(args *utils.ArgsGetGroupWithArgDispatcher,
	rply *string) (err error) {
	Cache.RemoveGroup(args.CacheID, args.GroupID, true, utils.NonTransactional)
	*rply = utils.OK
	return
}

func (chS *CacheS) reloadCache(chID string, IDs []string) error {
	return chS.dm.CacheDataFromDB(chID, IDs, true)
}

func (chS *CacheS) V1ReloadCache(attrs utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
	if attrs.FlushAll {
		Cache.Clear(nil)
		return
	}
	if len(attrs.DestinationIDs) != 0 {
		// Reload Destinations
		if err = chS.reloadCache(utils.DESTINATION_PREFIX, attrs.DestinationIDs); err != nil {
			return
		}
	}
	if len(attrs.ReverseDestinationIDs) != 0 {
		// Reload ReverseDestinations
		if err = chS.reloadCache(utils.REVERSE_DESTINATION_PREFIX, attrs.ReverseDestinationIDs); err != nil {
			return
		}
	}
	if len(attrs.RatingPlanIDs) != 0 {
		// RatingPlans
		if err = chS.reloadCache(utils.RATING_PLAN_PREFIX, attrs.RatingPlanIDs); err != nil {
			return
		}
	}
	if len(attrs.RatingProfileIDs) != 0 {
		// RatingProfiles
		if err = chS.reloadCache(utils.RATING_PROFILE_PREFIX, attrs.RatingProfileIDs); err != nil {
			return
		}
	}
	if len(attrs.ActionIDs) != 0 {
		// Actions
		if err = chS.reloadCache(utils.ACTION_PREFIX, attrs.ActionIDs); err != nil {
			return
		}
	}
	if len(attrs.ActionPlanIDs) != 0 {
		// ActionPlans
		if err = chS.reloadCache(utils.ACTION_PLAN_PREFIX, attrs.ActionPlanIDs); err != nil {
			return
		}
	}
	if len(attrs.AccountActionPlanIDs) != 0 {
		// AccountActionPlans
		if err = chS.reloadCache(utils.AccountActionPlansPrefix, attrs.AccountActionPlanIDs); err != nil {
			return
		}
	}
	if len(attrs.ActionTriggerIDs) != 0 {
		// ActionTriggers
		if err = chS.reloadCache(utils.ACTION_TRIGGER_PREFIX, attrs.ActionTriggerIDs); err != nil {
			return
		}
	}
	if len(attrs.SharedGroupIDs) != 0 {
		// SharedGroups
		if err = chS.reloadCache(utils.SHARED_GROUP_PREFIX, attrs.SharedGroupIDs); err != nil {
			return
		}
	}
	if len(attrs.ResourceProfileIDs) != 0 {
		// ResourceProfiles
		if err = chS.reloadCache(utils.ResourceProfilesPrefix, attrs.ResourceProfileIDs); err != nil {
			return
		}
	}
	if len(attrs.ResourceIDs) != 0 {
		// Resources
		if err = chS.reloadCache(utils.ResourcesPrefix, attrs.ResourceIDs); err != nil {
			return
		}
	}
	if len(attrs.StatsQueueIDs) != 0 {
		// StatQueues
		if err = chS.reloadCache(utils.StatQueuePrefix, attrs.StatsQueueIDs); err != nil {
			return
		}
	}
	if len(attrs.StatsQueueProfileIDs) != 0 {
		// StatQueueProfiles
		if err = chS.reloadCache(utils.StatQueueProfilePrefix, attrs.StatsQueueProfileIDs); err != nil {
			return
		}
	}
	if len(attrs.ThresholdIDs) != 0 {
		// Thresholds
		if err = chS.reloadCache(utils.ThresholdPrefix, attrs.ThresholdIDs); err != nil {
			return
		}
	}
	if len(attrs.ThresholdProfileIDs) != 0 {
		// ThresholdProfiles
		if err = chS.reloadCache(utils.ThresholdProfilePrefix, attrs.ThresholdProfileIDs); err != nil {
			return
		}
	}
	if len(attrs.FilterIDs) != 0 {
		// Filters
		if err = chS.reloadCache(utils.FilterPrefix, attrs.FilterIDs); err != nil {
			return
		}
	}
	if len(attrs.SupplierProfileIDs) != 0 {
		// SupplierProfile
		if err = chS.reloadCache(utils.SupplierProfilePrefix, attrs.SupplierProfileIDs); err != nil {
			return
		}
	}
	if len(attrs.AttributeProfileIDs) != 0 {
		// AttributeProfile
		if err = chS.reloadCache(utils.AttributeProfilePrefix, attrs.AttributeProfileIDs); err != nil {
			return
		}
	}
	if len(attrs.ChargerProfileIDs) != 0 {
		// ChargerProfiles
		if err = chS.reloadCache(utils.ChargerProfilePrefix, attrs.ChargerProfileIDs); err != nil {
			return
		}
	}
	if len(attrs.DispatcherProfileIDs) != 0 {
		// DispatcherProfile
		if err = chS.reloadCache(utils.DispatcherProfilePrefix, attrs.DispatcherProfileIDs); err != nil {
			return
		}
	}
	if len(attrs.DispatcherHostIDs) != 0 {
		// DispatcherHosts
		if err = chS.reloadCache(utils.DispatcherHostPrefix, attrs.DispatcherHostIDs); err != nil {
			return
		}
	}

	//get loadIDs from database for all types
	loadIDs, err := chS.dm.GetItemLoadIDs(utils.EmptyString, false)
	if err != nil {
		if err == utils.ErrNotFound { // we can receive cache reload from LoaderS and we store the LoadID only after all Items was processed
			loadIDs = make(map[string]int64)
		} else {
			return err
		}
	}
	cacheLoadIDs := populateCacheLoadIDs(loadIDs, attrs.AttrReloadCache)
	for key, val := range cacheLoadIDs {
		Cache.Set(utils.CacheLoadIDs, key, val, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}

	*reply = utils.OK
	return nil
}

func toStringSlice(in *[]string) []string {
	if in == nil {
		return nil
	}
	return *in
}

func (chS *CacheS) V1LoadCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
	if args.FlushAll {
		Cache.Clear(nil)
	}
	if err := chS.dm.LoadDataDBCache(
		args.DestinationIDs,
		args.ReverseDestinationIDs,
		args.RatingPlanIDs,
		args.RatingProfileIDs,
		args.ActionIDs,
		args.ActionPlanIDs,
		args.AccountActionPlanIDs,
		args.ActionTriggerIDs,
		args.SharedGroupIDs,
		args.ResourceProfileIDs,
		args.ResourceIDs,
		args.StatsQueueIDs,
		args.StatsQueueProfileIDs,
		args.ThresholdIDs,
		args.ThresholdProfileIDs,
		args.FilterIDs,
		args.SupplierProfileIDs,
		args.AttributeProfileIDs,
		args.ChargerProfileIDs,
		args.DispatcherProfileIDs,
		args.DispatcherHostIDs,
	); err != nil {
		return utils.NewErrServerError(err)
	}
	//get loadIDs for all types
	loadIDs, err := chS.dm.GetItemLoadIDs(utils.EmptyString, false)
	if err != nil {
		if err == utils.ErrNotFound { // we can receive cache reload from LoaderS and we store the LoadID only after all Items was processed
			loadIDs = make(map[string]int64)
		} else {
			return err
		}
	}
	cacheLoadIDs := populateCacheLoadIDs(loadIDs, args.AttrReloadCache)
	for key, val := range cacheLoadIDs {
		Cache.Set(utils.CacheLoadIDs, key, val, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	*reply = utils.OK
	return nil
}

func flushCache(chID string, IDs []string) {
	for _, key := range IDs {
		Cache.Remove(chID, key, true, utils.NonTransactional)
	}
}

// V1FlushCache wipes out cache for a prefix or completely
func (chS *CacheS) V1FlushCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
	if args.FlushAll {
		Cache.Clear(nil)
		*reply = utils.OK
		return
	}
	flushCache(utils.CacheDestinations, args.DestinationIDs)
	flushCache(utils.CacheReverseDestinations, args.ReverseDestinationIDs)
	flushCache(utils.CacheRatingPlans, args.RatingPlanIDs)
	flushCache(utils.CacheRatingProfiles, args.RatingProfileIDs)
	flushCache(utils.CacheActions, args.ActionIDs)
	flushCache(utils.CacheActionPlans, args.ActionPlanIDs)
	flushCache(utils.CacheActionTriggers, args.ActionTriggerIDs)
	flushCache(utils.CacheSharedGroups, args.SharedGroupIDs)
	flushCache(utils.CacheResourceProfiles, args.ResourceProfileIDs)
	flushCache(utils.CacheResources, args.ResourceIDs)
	flushCache(utils.CacheStatQueues, args.StatsQueueIDs)
	flushCache(utils.CacheThresholdProfiles, args.StatsQueueProfileIDs)
	flushCache(utils.CacheThresholds, args.ThresholdIDs)
	flushCache(utils.CacheThresholdProfiles, args.ThresholdProfileIDs)
	flushCache(utils.CacheFilters, args.FilterIDs)
	flushCache(utils.CacheSupplierProfiles, args.SupplierProfileIDs)
	flushCache(utils.CacheAttributeProfiles, args.AttributeProfileIDs)
	flushCache(utils.CacheChargerProfiles, args.ChargerProfileIDs)
	flushCache(utils.CacheDispatcherProfiles, args.DispatcherProfileIDs)
	flushCache(utils.CacheDispatcherHosts, args.DispatcherHostIDs)
	flushCache(utils.CacheDispatcherRoutes, args.DispatcherRoutesIDs)
	//get loadIDs for all types
	loadIDs, err := chS.dm.GetItemLoadIDs(utils.EmptyString, false)
	if err != nil {
		return err
	}
	cacheLoadIDs := populateCacheLoadIDs(loadIDs, args.AttrReloadCache)
	for key, val := range cacheLoadIDs {
		Cache.Set(utils.CacheLoadIDs, key, val, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	*reply = utils.OK
	return
}

//populateCacheLoadIDs populate cacheLoadIDs based on attrs
func populateCacheLoadIDs(loadIDs map[string]int64, attrs utils.AttrReloadCache) (cacheLoadIDs map[string]int64) {
	cacheLoadIDs = make(map[string]int64)
	//based on IDs of each type populate cacheLoadIDs and add into cache
	if attrs.DestinationIDs == nil || len(attrs.DestinationIDs) != 0 {
		cacheLoadIDs[utils.CacheDestinations] = loadIDs[utils.CacheDestinations]
	}
	if attrs.ReverseDestinationIDs == nil || len(attrs.ReverseDestinationIDs) != 0 {
		cacheLoadIDs[utils.CacheReverseDestinations] = loadIDs[utils.CacheReverseDestinations]
	}
	if attrs.RatingPlanIDs == nil || len(attrs.RatingPlanIDs) != 0 {
		cacheLoadIDs[utils.CacheRatingPlans] = loadIDs[utils.CacheRatingPlans]
	}
	if attrs.RatingProfileIDs == nil || len(attrs.RatingProfileIDs) != 0 {
		cacheLoadIDs[utils.CacheRatingProfiles] = loadIDs[utils.CacheRatingProfiles]
	}
	if attrs.ActionIDs == nil || len(attrs.ActionIDs) != 0 {
		cacheLoadIDs[utils.CacheActions] = loadIDs[utils.CacheActions]
	}
	if attrs.ActionPlanIDs == nil || len(attrs.ActionPlanIDs) != 0 {
		cacheLoadIDs[utils.CacheActionPlans] = loadIDs[utils.CacheActionPlans]
	}
	if attrs.AccountActionPlanIDs == nil || len(attrs.AccountActionPlanIDs) != 0 {
		cacheLoadIDs[utils.CacheAccountActionPlans] = loadIDs[utils.CacheAccountActionPlans]
	}
	if attrs.ActionTriggerIDs == nil || len(attrs.ActionTriggerIDs) != 0 {
		cacheLoadIDs[utils.CacheActionTriggers] = loadIDs[utils.CacheActionTriggers]
	}
	if attrs.SharedGroupIDs == nil || len(attrs.SharedGroupIDs) != 0 {
		cacheLoadIDs[utils.CacheSharedGroups] = loadIDs[utils.CacheSharedGroups]
	}
	if attrs.ResourceProfileIDs == nil || len(attrs.ResourceProfileIDs) != 0 {
		cacheLoadIDs[utils.CacheResourceProfiles] = loadIDs[utils.CacheResourceProfiles]
	}
	if attrs.ResourceIDs == nil || len(attrs.ResourceIDs) != 0 {
		cacheLoadIDs[utils.CacheResources] = loadIDs[utils.CacheResources]
	}
	if attrs.StatsQueueProfileIDs == nil || len(attrs.StatsQueueProfileIDs) != 0 {
		cacheLoadIDs[utils.CacheStatQueueProfiles] = loadIDs[utils.CacheStatQueueProfiles]
	}
	if attrs.StatsQueueIDs == nil || len(attrs.StatsQueueIDs) != 0 {
		cacheLoadIDs[utils.CacheStatQueues] = loadIDs[utils.CacheStatQueues]
	}
	if attrs.ThresholdProfileIDs == nil || len(attrs.ThresholdProfileIDs) != 0 {
		cacheLoadIDs[utils.CacheThresholdProfiles] = loadIDs[utils.CacheThresholdProfiles]
	}
	if attrs.ThresholdIDs == nil || len(attrs.ThresholdIDs) != 0 {
		cacheLoadIDs[utils.CacheThresholds] = loadIDs[utils.CacheThresholds]
	}
	if attrs.FilterIDs == nil || len(attrs.FilterIDs) != 0 {
		cacheLoadIDs[utils.CacheFilters] = loadIDs[utils.CacheFilters]
	}
	if attrs.SupplierProfileIDs == nil || len(attrs.SupplierProfileIDs) != 0 {
		cacheLoadIDs[utils.CacheSupplierProfiles] = loadIDs[utils.CacheSupplierProfiles]
	}
	if attrs.AttributeProfileIDs == nil || len(attrs.AttributeProfileIDs) != 0 {
		cacheLoadIDs[utils.CacheAttributeProfiles] = loadIDs[utils.CacheAttributeProfiles]
	}
	if attrs.ChargerProfileIDs == nil || len(attrs.ChargerProfileIDs) != 0 {
		cacheLoadIDs[utils.CacheChargerProfiles] = loadIDs[utils.CacheChargerProfiles]
	}
	if attrs.DispatcherProfileIDs == nil || len(attrs.DispatcherProfileIDs) != 0 {
		cacheLoadIDs[utils.CacheDispatcherProfiles] = loadIDs[utils.CacheDispatcherProfiles]
	}
	return
}
