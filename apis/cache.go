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

package apis

import (
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

type CacheSv1 struct {
	ping
	cacheS *engine.CacheS
}

func NewCacheSv1(cacheS *engine.CacheS) *CacheSv1 {
	return &CacheSv1{cacheS: cacheS}
}

// GetItemIDs returns the IDs for cacheID with given prefix
func (chSv1 *CacheSv1) GetItemIDs(_ *context.Context, args *utils.ArgsGetCacheItemIDsWithAPIOpts,
	reply *[]string) error {
	return chSv1.cacheS.V1GetItemIDs(args, reply)
}

// HasItem verifies the existence of an Item in cache
func (chSv1 *CacheSv1) HasItem(_ *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *bool) error {
	return chSv1.cacheS.V1HasItem(args, reply)
}

// GetItemExpiryTime returns the expiryTime for an item
func (chSv1 *CacheSv1) GetItemExpiryTime(args *utils.ArgsGetCacheItemWithAPIOpts, //
	reply *time.Time) error {
	return chSv1.cacheS.V1GetItemExpiryTime(args, reply)
}

// RemoveItem removes the Item with ID from cache
func (chSv1 *CacheSv1) RemoveItem(_ *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *string) error {
	return chSv1.cacheS.V1RemoveItem(args, reply)
}

// RemoveItems removes the Items with ID from cache
func (chSv1 *CacheSv1) RemoveItems(_ *context.Context, args utils.AttrReloadCacheWithAPIOpts,
	reply *string) error {
	return chSv1.cacheS.V1RemoveItems(args, reply)
}

// Clear will clear partitions in the cache (nil fol all, empty slice for none)
func (chSv1 *CacheSv1) Clear(ctx *context.Context, args *utils.AttrCacheIDsWithAPIOpts,
	reply *string) error {
	return chSv1.cacheS.V1Clear(ctx, args, reply)
}

// GetCacheStats returns CacheStats filtered by cacheIDs
func (chSv1 *CacheSv1) GetCacheStats(_ *context.Context, args *utils.AttrCacheIDsWithAPIOpts, //
	rply *map[string]*ltcache.CacheStats) error {
	return chSv1.cacheS.V1GetCacheStats(args, rply)
}

// PrecacheStatus checks status of active precache processes
func (chSv1 *CacheSv1) PrecacheStatus(args *utils.AttrCacheIDsWithAPIOpts, rply *map[string]string) error { //
	return chSv1.cacheS.V1PrecacheStatus(args, rply)
}

// HasGroup checks existence of a group in cache
func (chSv1 *CacheSv1) HasGroup(args *utils.ArgsGetGroupWithAPIOpts, //
	rply *bool) (err error) {
	return chSv1.cacheS.V1HasGroup(args, rply)
}

// GetGroupItemIDs returns a list of itemIDs in a cache group
func (chSv1 *CacheSv1) GetGroupItemIDs(args *utils.ArgsGetGroupWithAPIOpts, //
	rply *[]string) (err error) {
	return chSv1.cacheS.V1GetGroupItemIDs(args, rply)
}

// RemoveGroup will remove a group and all items belonging to it from cache
func (chSv1 *CacheSv1) RemoveGroup(args *utils.ArgsGetGroupWithAPIOpts, //
	rply *string) (err error) {
	return chSv1.cacheS.V1RemoveGroup(args, rply)
}

// ReloadCache reloads cache from DB for a prefix or completely
func (chSv1 *CacheSv1) ReloadCache(ctx *context.Context, args *utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	return chSv1.cacheS.V1ReloadCache(ctx, *args, reply)
}

// LoadCache loads cache from DB for a prefix or completely
func (chSv1 *CacheSv1) LoadCache(ctx *context.Context, args *utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	return chSv1.cacheS.V1LoadCache(ctx, *args, reply)
}

// ReplicateSet replicate an item
func (chSv1 *CacheSv1) ReplicateSet(args *utils.ArgCacheReplicateSet, reply *string) (err error) { //
	return chSv1.cacheS.V1ReplicateSet(args, reply)
}

// ReplicateRemove remove an item
func (chSv1 *CacheSv1) ReplicateRemove(args *utils.ArgCacheReplicateRemove, reply *string) (err error) { //
	return chSv1.cacheS.V1ReplicateRemove(args, reply)
}
