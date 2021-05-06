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
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var Cache *CacheS

func init() {
	Cache = NewCacheS(config.CgrConfig(), nil, nil)
	gob.Register(new(AttributeProfile))
	gob.Register(new(AttributeProfileWithAPIOpts))
	// Threshold
	gob.Register(new(Threshold))
	gob.Register(new(ThresholdProfile))
	gob.Register(new(ThresholdProfileWithAPIOpts))
	gob.Register(new(ThresholdWithAPIOpts))
	// Resource
	gob.Register(new(Resource))
	gob.Register(new(ResourceProfile))
	gob.Register(new(ResourceProfileWithAPIOpts))
	gob.Register(new(ResourceWithAPIOpts))
	// Stats
	gob.Register(new(StatQueue))
	gob.Register(new(StatQueueProfile))
	gob.Register(new(StatQueueProfileWithAPIOpts))
	gob.Register(new(StoredStatQueue))
	gob.Register(new(StatQueueProfileWithAPIOpts))
	// RouteS
	gob.Register(new(RouteProfile))
	gob.Register(new(RouteProfileWithAPIOpts))
	// Filters
	gob.Register(new(Filter))
	gob.Register(new(FilterWithAPIOpts))
	// Dispatcher
	gob.Register(new(DispatcherHost))
	gob.Register(new(DispatcherHostProfile))
	gob.Register(new(DispatcherHostWithAPIOpts))

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

	gob.Register(new(HTTPPosterRequest))

	gob.Register([]interface{}{})
	gob.Register([]map[string]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[string][]map[string]interface{}{})
	gob.Register(map[string]string{})
	gob.Register(time.Duration(0))
	gob.Register(time.Time{})
	gob.Register(url.Values{})
	gob.Register(json.RawMessage{})
	gob.Register(BalanceSummaries{})

	gob.Register(new(utils.ArgCacheReplicateSet))
	gob.Register(new(utils.ArgCacheReplicateRemove))

	gob.Register(utils.StringSet{})
}

// NewCacheS initializes the Cache service and executes the precaching
func NewCacheS(cfg *config.CGRConfig, dm *DataManager, cpS *CapsStats) (c *CacheS) {
	cfg.CacheCfg().AddTmpCaches()
	tCache := cfg.CacheCfg().AsTransCacheConfig()
	if len(cfg.CacheCfg().ReplicationConns) != 0 {
		var reply string
		for k, val := range tCache {
			if !cfg.CacheCfg().Partitions[k].Replicate ||
				k == utils.CacheCapsEvents {
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

	if _, has := tCache[utils.CacheCapsEvents]; has && cpS != nil {
		tCache[utils.CacheCapsEvents].OnEvicted = cpS.OnEvict
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

// SetWithoutReplicate is an exported method from TransCache
// handled Replicate functionality
func (chS *CacheS) SetWithoutReplicate(chID, itmID string, value interface{},
	groupIDs []string, commit bool, transID string) {
	chS.tCache.Set(chID, itmID, value, groupIDs, commit, transID)
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

// RemoveWithoutReplicate is an exported method from TransCache
func (chS *CacheS) RemoveWithoutReplicate(chID, itmID string, commit bool, transID string) {
	chS.tCache.Remove(chID, itmID, commit, transID)
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
				utils.CacheInstanceToPrefix[cacheID],
				[]string{utils.MetaAny},
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

func (chS *CacheS) V1GetItemIDs(args *utils.ArgsGetCacheItemIDsWithAPIOpts,
	reply *[]string) (err error) {
	itmIDs := chS.tCache.GetItemIDs(args.CacheID, args.ItemIDPrefix)
	if len(itmIDs) == 0 {
		return utils.ErrNotFound
	}
	*reply = itmIDs
	return
}

func (chS *CacheS) V1HasItem(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *bool) (err error) {
	*reply = chS.tCache.HasItem(args.CacheID, args.ItemID)
	return
}

func (chS *CacheS) V1GetItemExpiryTime(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *time.Time) (err error) {
	expTime, has := chS.tCache.GetItemExpiryTime(args.CacheID, args.ItemID)
	if !has {
		return utils.ErrNotFound
	}
	*reply = expTime
	return
}

func (chS *CacheS) V1RemoveItem(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *string) (err error) {
	chS.tCache.Remove(args.CacheID, args.ItemID, true, utils.NonTransactional)
	*reply = utils.OK
	return
}

func (chS *CacheS) V1RemoveItems(args utils.AttrReloadCacheWithAPIOpts,
	reply *string) (err error) {
	for key, ids := range args.ArgsCache {
		if cacheID, has := utils.ArgCacheToInstance[key]; has {
			for _, id := range ids {
				chS.tCache.Remove(cacheID, id, true, utils.NonTransactional)
			}
		}
	}
	*reply = utils.OK
	return
}

func (chS *CacheS) V1Clear(args *utils.AttrCacheIDsWithAPIOpts,
	reply *string) (err error) {
	chS.tCache.Clear(args.CacheIDs)
	*reply = utils.OK
	return
}

func (chS *CacheS) V1GetCacheStats(args *utils.AttrCacheIDsWithAPIOpts,
	rply *map[string]*ltcache.CacheStats) (err error) {
	cs := chS.tCache.GetCacheStats(args.CacheIDs)
	*rply = cs
	return
}

func (chS *CacheS) V1PrecacheStatus(args *utils.AttrCacheIDsWithAPIOpts, rply *map[string]string) (err error) {
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

func (chS *CacheS) V1HasGroup(args *utils.ArgsGetGroupWithAPIOpts,
	rply *bool) (err error) {
	*rply = chS.tCache.HasGroup(args.CacheID, args.GroupID)
	return
}

func (chS *CacheS) V1GetGroupItemIDs(args *utils.ArgsGetGroupWithAPIOpts,
	rply *[]string) (err error) {
	if has := chS.tCache.HasGroup(args.CacheID, args.GroupID); !has {
		return utils.ErrNotFound
	}
	*rply = chS.tCache.GetGroupItemIDs(args.CacheID, args.GroupID)
	return
}

func (chS *CacheS) V1RemoveGroup(args *utils.ArgsGetGroupWithAPIOpts,
	rply *string) (err error) {
	chS.tCache.RemoveGroup(args.CacheID, args.GroupID, true, utils.NonTransactional)
	*rply = utils.OK
	return
}

func (chS *CacheS) V1ReloadCache(attrs utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	return chS.cacheDataFromDB(attrs, reply, true)
}

func (chS *CacheS) V1LoadCache(attrs utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	return chS.cacheDataFromDB(attrs, reply, false)
}

func (chS *CacheS) cacheDataFromDB(attrs utils.AttrReloadCacheWithAPIOpts, reply *string, mustBeCached bool) (err error) {
	for key, ids := range attrs.ArgsCache {
		if prfx, has := utils.ArgCacheToPrefix[key]; has {
			if err = chS.dm.CacheDataFromDB(prfx, ids, true); err != nil {
				return
			}
		}
	}
	//get loadIDs from database for all types
	var loadIDs map[string]int64
	if loadIDs, err = chS.dm.GetItemLoadIDs(utils.EmptyString, false); err != nil {
		if err != utils.ErrNotFound { // we can receive cache reload from LoaderS and we store the LoadID only after all Items was processed
			return
		}
		loadIDs = make(map[string]int64)
	}
	for key, val := range populateCacheLoadIDs(loadIDs, attrs.ArgsCache) {
		chS.tCache.Set(utils.CacheLoadIDs, key, val, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	*reply = utils.OK
	return
}

//populateCacheLoadIDs populate cacheLoadIDs based on attrs
func populateCacheLoadIDs(loadIDs map[string]int64, attrs map[string][]string) (cacheLoadIDs map[string]int64) {
	cacheLoadIDs = make(map[string]int64)
	//based on IDs of each type populate cacheLoadIDs and add into cache
	for key, ids := range attrs {
		if inst, has := utils.ArgCacheToInstance[key]; has &&
			(ids == nil || len(ids) != 0) {
			cacheLoadIDs[inst] = loadIDs[inst]
		}
	}
	return
}

// ReplicateSet replicate an item to ReplicationConns
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
	if cmp, canCast := args.Value.(utils.Compiler); canCast {
		if err = cmp.Compile(); err != nil {
			return
		}
	}
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
