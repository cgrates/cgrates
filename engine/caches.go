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
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var Cache *CacheS

func init() {
	Cache = NewCacheS(config.CgrConfig(), nil)
	// Threshold
	gob.Register(new(Threshold))
	gob.Register(new(ThresholdProfile))
	gob.Register(new(ThresholdProfileWithArgDispatcher))
	gob.Register(new(ThresholdWithArgDispatcher))
	// Resource
	gob.Register(new(Resource))
	gob.Register(new(ResourceProfile))
	gob.Register(new(ResourceProfileWithArgDispatcher))
	gob.Register(new(ResourceWithArgDispatcher))
	// Stats
	gob.Register(new(StatQueue))
	gob.Register(new(StatQueueProfile))
	gob.Register(new(StatQueueProfileWithArgDispatcher))
	gob.Register(new(StoredStatQueue))
	gob.Register(new(StatQueueProfileWithArgDispatcher))
	// Suppliers
	gob.Register(new(RouteProfile))
	gob.Register(new(RouteProfileWithArgDispatcher))
	// Filters
	gob.Register(new(Filter))
	gob.Register(new(FilterWithArgDispatcher))
	// Dispatcher
	gob.Register(new(DispatcherHost))
	gob.Register(new(DispatcherHostProfile))
	gob.Register(new(DispatcherHostWithArgDispatcher))

	// CDRs
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

//SetCache shared the cache from other subsystems
func SetCache(chS *CacheS) {
	Cache = chS
}

// NewCacheS initializes the Cache service and executes the precaching
func NewCacheS(cfg *config.CGRConfig, dm *DataManager) (c *CacheS) {
	cfg.CacheCfg().AddTmpCaches()
	tCache := cfg.CacheCfg().AsTransCacheConfig()
	if len(cfg.CacheCfg().ReplicationConns) != 0 {
		var reply string
		for k, val := range tCache {
			if !cfg.CacheCfg().Partitions[k].Replicate {
				continue
			}
			val.OnEvicted = func(itmID string, value interface{}) {
				if err := connMgr.Call(cfg.CacheCfg().ReplicationConns, nil, utils.CacheSv1ReplicateRemove,
					&utils.ArgCacheReplicateRemove{
						CacheID: k,
						ItemID:  itmID,
					}, &reply); err != nil {
					utils.Logger.Warning(fmt.Sprintf("error: %+v when autoexpired item: %+v from: %+v", err, itmID, k))
				}
			}
		}
	}

	c = &CacheS{
		cfg:     cfg,
		dm:      dm,
		pcItems: make(map[string]chan struct{}),
		tCache:  ltcache.NewTransCache(tCache),
	}
	for cacheID := range cfg.CacheCfg().Partitions {
		c.pcItems[cacheID] = make(chan struct{})
	}
	return
}

// CacheS deals with cache preload and other cache related tasks/APIs
type CacheS struct {
	cfg     *config.CGRConfig
	dm      *DataManager
	pcItems map[string]chan struct{} // signal precaching
	tCache  *ltcache.TransCache
}

// Set is an exported method from TransCache
// handled Replicate functionality
func (chS *CacheS) Set(chID, itmID string, value interface{},
	groupIDs []string, commit bool, transID string) (err error) {
	chS.tCache.Set(chID, itmID, value, groupIDs, commit, transID)
	return chS.ReplicateSet(chID, itmID, value)
}

// HasItem is an exported method from TransCache
func (chS *CacheS) HasItem(chID, itmID string) (has bool) {
	return chS.tCache.HasItem(chID, itmID)
}

// Get is an exported method from TransCache
func (chS *CacheS) Get(chID, itmID string) (interface{}, bool) {
	return chS.tCache.Get(chID, itmID)
}

// GetItemIDs is an exported method from TransCache
func (chS *CacheS) GetItemIDs(chID, prfx string) (itmIDs []string) {
	return chS.tCache.GetItemIDs(chID, prfx)
}

// Remove is an exported method from TransCache
func (chS *CacheS) Remove(chID, itmID string, commit bool, transID string) (err error) {
	chS.tCache.Remove(chID, itmID, commit, transID)
	return chS.ReplicateRemove(chID, itmID)
}

// Clear is an exported method from TransCache
func (chS *CacheS) Clear(chIDs []string) {
	chS.tCache.Clear(chIDs)
}

// BeginTransaction is an exported method from TransCache
func (chS *CacheS) BeginTransaction() string {
	return chS.tCache.BeginTransaction()
}

// RollbackTransaction is an exported method from TransCache
func (chS *CacheS) RollbackTransaction(transID string) {
	chS.tCache.RollbackTransaction(transID)
}

// CommitTransaction is an exported method from TransCache
func (chS *CacheS) CommitTransaction(transID string) {
	chS.tCache.CommitTransaction(transID)
}

// GetCloned is an exported method from TransCache
func (chS *CacheS) GetCloned(chID, itmID string) (cln interface{}, err error) {
	return chS.tCache.GetCloned(chID, itmID)
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
	for cacheID, cacheCfg := range chS.cfg.CacheCfg().Partitions {
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
	time.Sleep(1) // switch context
	go func() {   // report wg.Wait on doneChan
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
	itmIDs := chS.tCache.GetItemIDs(args.CacheID, args.ItemIDPrefix)
	if len(itmIDs) == 0 {
		return utils.ErrNotFound
	}
	*reply = itmIDs
	return
}

func (chS *CacheS) V1HasItem(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *bool) (err error) {
	*reply = chS.tCache.HasItem(args.CacheID, args.ItemID)
	return
}

func (chS *CacheS) V1GetItemExpiryTime(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *time.Time) (err error) {
	expTime, has := chS.tCache.GetItemExpiryTime(args.CacheID, args.ItemID)
	if !has {
		return utils.ErrNotFound
	}
	*reply = expTime
	return
}

func (chS *CacheS) V1RemoveItem(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *string) (err error) {
	chS.tCache.Remove(args.CacheID, args.ItemID, true, utils.NonTransactional)
	*reply = utils.OK
	return
}

func (chS *CacheS) V1Clear(args *utils.AttrCacheIDsWithArgDispatcher,
	reply *string) (err error) {
	chS.tCache.Clear(args.CacheIDs)
	*reply = utils.OK
	return
}

func (chS *CacheS) V1GetCacheStats(args *utils.AttrCacheIDsWithArgDispatcher,
	rply *map[string]*ltcache.CacheStats) (err error) {
	cs := chS.tCache.GetCacheStats(args.CacheIDs)
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
	*rply = chS.tCache.HasGroup(args.CacheID, args.GroupID)
	return
}

func (chS *CacheS) V1GetGroupItemIDs(args *utils.ArgsGetGroupWithArgDispatcher,
	rply *[]string) (err error) {
	if has := chS.tCache.HasGroup(args.CacheID, args.GroupID); !has {
		return utils.ErrNotFound
	}
	*rply = chS.tCache.GetGroupItemIDs(args.CacheID, args.GroupID)
	return
}

func (chS *CacheS) V1RemoveGroup(args *utils.ArgsGetGroupWithArgDispatcher,
	rply *string) (err error) {
	chS.tCache.RemoveGroup(args.CacheID, args.GroupID, true, utils.NonTransactional)
	*rply = utils.OK
	return
}

func (chS *CacheS) reloadCache(chID string, IDs []string) error {
	return chS.dm.CacheDataFromDB(chID, IDs, true)
}

func (chS *CacheS) V1ReloadCache(attrs utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
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
	if len(attrs.RouteProfileIDs) != 0 {
		// RouteProfiles
		if err = chS.reloadCache(utils.RouteProfilePrefix, attrs.RouteProfileIDs); err != nil {
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
	cacheLoadIDs := populateCacheLoadIDs(loadIDs, attrs.ArgsCache)
	for key, val := range cacheLoadIDs {
		chS.tCache.Set(utils.CacheLoadIDs, key, val, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}

	*reply = utils.OK
	return nil
}

func (chS *CacheS) V1LoadCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
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
		args.RouteProfileIDs,
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
	cacheLoadIDs := populateCacheLoadIDs(loadIDs, args.ArgsCache)
	for key, val := range cacheLoadIDs {
		chS.tCache.Set(utils.CacheLoadIDs, key, val, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	*reply = utils.OK
	return nil
}

//populateCacheLoadIDs populate cacheLoadIDs based on attrs
func populateCacheLoadIDs(loadIDs map[string]int64, attrs utils.ArgsCache) (cacheLoadIDs map[string]int64) {
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
	if attrs.RouteProfileIDs == nil || len(attrs.RouteProfileIDs) != 0 {
		cacheLoadIDs[utils.CacheRouteProfiles] = loadIDs[utils.CacheRouteProfiles]
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

// Replicate replicate an item to ReplicationConns
func (chS *CacheS) ReplicateSet(chID, itmID string, value interface{}) (err error) {
	if len(chS.cfg.CacheCfg().ReplicationConns) == 0 ||
		!chS.cfg.CacheCfg().Partitions[chID].Replicate {
		return
	}
	var reply string
	return connMgr.Call(chS.cfg.CacheCfg().ReplicationConns, nil, utils.CacheSv1ReplicateSet,
		&utils.ArgCacheReplicateSet{
			CacheID: chID,
			ItemID:  itmID,
			Value:   value,
		}, &reply)
}

// V1ReplicateSet replicate an item
func (chS *CacheS) V1ReplicateSet(args *utils.ArgCacheReplicateSet, reply *string) (err error) {
	chS.tCache.Set(args.CacheID, args.ItemID, args.Value, nil, true, utils.EmptyString)
	*reply = utils.OK
	return
}

// ReplicateRemove replicate an item to ReplicationConns
func (chS *CacheS) ReplicateRemove(chID, itmID string) (err error) {
	if len(chS.cfg.CacheCfg().ReplicationConns) == 0 ||
		!chS.cfg.CacheCfg().Partitions[chID].Replicate {
		return
	}
	var reply string
	return connMgr.Call(chS.cfg.CacheCfg().ReplicationConns, nil, utils.CacheSv1ReplicateRemove,
		&utils.ArgCacheReplicateRemove{
			CacheID: chID,
			ItemID:  itmID,
		}, &reply)
}

// V1ReplicateRemove replicate an item
func (chS *CacheS) V1ReplicateRemove(args *utils.ArgCacheReplicateRemove, reply *string) (err error) {
	chS.tCache.Remove(args.CacheID, args.ItemID, true, utils.EmptyString)
	*reply = utils.OK
	return
}
