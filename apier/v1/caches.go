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

// GetItemExpiryTime returns the expiryTime for an item
func (chSv1 *CacheSv1) GetItemIDs(args *engine.ArgsGetCacheItemIDs,
	reply *[]string) error {
	return chSv1.cacheS.V1GetItemIDs(args, reply)
}

// HasItem verifies the existence of an Item in cache
func (chSv1 *CacheSv1) HasItem(args *engine.ArgsGetCacheItem,
	reply *bool) error {
	return chSv1.cacheS.V1HasItem(args, reply)
}

// GetItemExpiryTime returns the expiryTime for an item
func (chSv1 *CacheSv1) GetItemExpiryTime(args *engine.ArgsGetCacheItem,
	reply *time.Time) error {
	return chSv1.cacheS.V1GetItemExpiryTime(args, reply)
}

// RemoveItem removes the Item with ID from cache
func (chSv1 *CacheSv1) RemoveItem(args *engine.ArgsGetCacheItem,
	reply *string) error {
	return chSv1.cacheS.V1RemoveItem(args, reply)
}

// Clear will clear partitions in the cache (nil fol all, empty slice for none)
func (chSv1 *CacheSv1) Clear(cacheIDs []string,
	reply *string) error {
	return chSv1.cacheS.V1Clear(cacheIDs, reply)
}

// GetCacheStats returns CacheStats filtered by cacheIDs
func (chSv1 *CacheSv1) GetCacheStats(cacheIDs []string,
	rply *map[string]*ltcache.CacheStats) error {
	return chSv1.cacheS.V1GetCacheStats(cacheIDs, rply)
}

// PrecacheStatus checks status of active precache processes
func (chSv1 *CacheSv1) PrecacheStatus(cacheIDs []string, rply *map[string]string) error {
	return chSv1.cacheS.V1PrecacheStatus(cacheIDs, rply)
}

// HasGroup checks existence of a group in cache
func (chSv1 *CacheSv1) HasGroup(args *engine.ArgsGetGroup,
	rply *bool) (err error) {
	return chSv1.cacheS.V1HasGroup(args, rply)
}

// GetGroupItemIDs returns a list of itemIDs in a cache group
func (chSv1 *CacheSv1) GetGroupItemIDs(args *engine.ArgsGetGroup,
	rply *[]string) (err error) {
	return chSv1.cacheS.V1GetGroupItemIDs(args, rply)
}

// RemoveGroup will remove a group and all items belonging to it from cache
func (chSv1 *CacheSv1) RemoveGroup(args *engine.ArgsGetGroup,
	rply *string) (err error) {
	return chSv1.cacheS.V1RemoveGroup(args, rply)
}

func (chSv1 *CacheSv1) Ping(ign string, reply *string) error {
	*reply = utils.Pong
	return nil
}
