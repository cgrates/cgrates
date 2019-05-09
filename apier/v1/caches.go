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
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func NewCacheSv1(cacheS *engine.CacheS) *CacheSv1 {
	return &CacheSv1{cacheS: cacheS}
}

// Exports RPC from CacheS
type CacheSv1 struct {
	cacheS *engine.CacheS
}

// GetItemIDs returns the IDs for cacheID with given prefix
func (chSv1 *CacheSv1) GetItemIDs(args *utils.ArgsGetCacheItemIDsWithArgDispatcher,
	reply *[]string) error {
	return chSv1.cacheS.V1GetItemIDs(args, reply)
}

// HasItem verifies the existence of an Item in cache
func (chSv1 *CacheSv1) HasItem(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *bool) error {
	return chSv1.cacheS.V1HasItem(args, reply)
}

// GetItemExpiryTime returns the expiryTime for an item
func (chSv1 *CacheSv1) GetItemExpiryTime(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *time.Time) error {
	return chSv1.cacheS.V1GetItemExpiryTime(args, reply)
}

// RemoveItem removes the Item with ID from cache
func (chSv1 *CacheSv1) RemoveItem(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *string) error {
	return chSv1.cacheS.V1RemoveItem(args, reply)
}

// Clear will clear partitions in the cache (nil fol all, empty slice for none)
func (chSv1 *CacheSv1) Clear(args *utils.AttrCacheIDsWithArgDispatcher,
	reply *string) error {
	return chSv1.cacheS.V1Clear(args, reply)
}

// FlushCache wipes out cache for a prefix or completely
func (chSv1 *CacheSv1) FlushCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
	return chSv1.cacheS.V1FlushCache(args, reply)
}

// GetCacheStats returns CacheStats filtered by cacheIDs
func (chSv1 *CacheSv1) GetCacheStats(args *utils.AttrCacheIDsWithArgDispatcher,
	rply *map[string]*ltcache.CacheStats) error {
	return chSv1.cacheS.V1GetCacheStats(args, rply)
}

// PrecacheStatus checks status of active precache processes
func (chSv1 *CacheSv1) PrecacheStatus(args *utils.AttrCacheIDsWithArgDispatcher, rply *map[string]string) error {
	return chSv1.cacheS.V1PrecacheStatus(args, rply)
}

// HasGroup checks existence of a group in cache
func (chSv1 *CacheSv1) HasGroup(args *utils.ArgsGetGroupWithArgDispatcher,
	rply *bool) (err error) {
	return chSv1.cacheS.V1HasGroup(args, rply)
}

// GetGroupItemIDs returns a list of itemIDs in a cache group
func (chSv1 *CacheSv1) GetGroupItemIDs(args *utils.ArgsGetGroupWithArgDispatcher,
	rply *[]string) (err error) {
	return chSv1.cacheS.V1GetGroupItemIDs(args, rply)
}

// RemoveGroup will remove a group and all items belonging to it from cache
func (chSv1 *CacheSv1) RemoveGroup(args *utils.ArgsGetGroupWithArgDispatcher,
	rply *string) (err error) {
	return chSv1.cacheS.V1RemoveGroup(args, rply)
}

// ReloadCache reloads cache from DB for a prefix or completely
func (chSv1 *CacheSv1) ReloadCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
	return chSv1.cacheS.V1ReloadCache(args, reply)
}

// LoadCache loads cache from DB for a prefix or completely
func (chSv1 *CacheSv1) LoadCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
	return chSv1.cacheS.V1LoadCache(args, reply)
}

// Ping used to detreminate if component is active
func (chSv1 *CacheSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
func (chSv1 *CacheSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(chSv1, serviceMethod, args, reply)
}
