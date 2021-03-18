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

package dispatchers

import (
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// CacheSv1Ping interogates CacheSv1 server responsible to process the event
func (dS *DispatcherService) CacheSv1Ping(args *utils.CGREvent,
	reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1Ping, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaCaches, utils.CacheSv1Ping, args, reply)
}

// CacheSv1GetItemIDs returns the IDs for cacheID with given prefix
func (dS *DispatcherService) CacheSv1GetItemIDs(args *utils.ArgsGetCacheItemIDsWithAPIOpts,
	reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1GetItemIDs, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	}, utils.MetaCaches, utils.CacheSv1GetItemIDs, args, reply)
}

// CacheSv1HasItem verifies the existence of an Item in cache
func (dS *DispatcherService) CacheSv1HasItem(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *bool) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1HasItem, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}

	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	},
		utils.MetaCaches, utils.CacheSv1HasItem, args, reply)
}

// CacheSv1GetItemExpiryTime returns the expiryTime for an item
func (dS *DispatcherService) CacheSv1GetItemExpiryTime(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *time.Time) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1GetItemExpiryTime, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}

	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	}, utils.MetaCaches, utils.CacheSv1GetItemExpiryTime, args, reply)
}

// CacheSv1RemoveItem removes the Item with ID from cache
func (dS *DispatcherService) CacheSv1RemoveItem(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1RemoveItem, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	}, utils.MetaCaches, utils.CacheSv1RemoveItem, args, reply)
}

// CacheSv1RemoveItems removes the Item with ID from cache
func (dS *DispatcherService) CacheSv1RemoveItems(args utils.AttrReloadCacheWithAPIOpts,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1RemoveItems, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	}, utils.MetaCaches, utils.CacheSv1RemoveItems, args, reply)
}

// CacheSv1Clear will clear partitions in the cache (nil fol all, empty slice for none)
func (dS *DispatcherService) CacheSv1Clear(args *utils.AttrCacheIDsWithOpts,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1Clear, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	}, utils.MetaCaches, utils.CacheSv1Clear, args, reply)
}

// CacheSv1GetCacheStats returns CacheStats filtered by cacheIDs
func (dS *DispatcherService) CacheSv1GetCacheStats(args *utils.AttrCacheIDsWithOpts,
	reply *map[string]*ltcache.CacheStats) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1GetCacheStats, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	}, utils.MetaCaches, utils.CacheSv1GetCacheStats, args, reply)
}

// CacheSv1PrecacheStatus checks status of active precache processes
func (dS *DispatcherService) CacheSv1PrecacheStatus(args *utils.AttrCacheIDsWithOpts, reply *map[string]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1PrecacheStatus, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	}, utils.MetaCaches, utils.CacheSv1PrecacheStatus, args, reply)
}

// CacheSv1HasGroup checks existence of a group in cache
func (dS *DispatcherService) CacheSv1HasGroup(args *utils.ArgsGetGroupWithOpts,
	reply *bool) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1HasGroup, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.Opts,
	}, utils.MetaCaches, utils.CacheSv1HasGroup, args, reply)
}

// CacheSv1GetGroupItemIDs returns a list of itemIDs in a cache group
func (dS *DispatcherService) CacheSv1GetGroupItemIDs(args *utils.ArgsGetGroupWithOpts,
	reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1GetGroupItemIDs, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.Opts,
	}, utils.MetaCaches, utils.CacheSv1GetGroupItemIDs, args, reply)
}

// CacheSv1RemoveGroup will remove a group and all items belonging to it from cache
func (dS *DispatcherService) CacheSv1RemoveGroup(args *utils.ArgsGetGroupWithOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1RemoveGroup, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.Opts,
	}, utils.MetaCaches, utils.CacheSv1RemoveGroup, args, reply)
}

// CacheSv1ReloadCache reloads cache from DB for a prefix or completely
func (dS *DispatcherService) CacheSv1ReloadCache(args utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1ReloadCache, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	}, utils.MetaCaches, utils.CacheSv1ReloadCache, args, reply)
}

// CacheSv1LoadCache loads cache from DB for a prefix or completely
func (dS *DispatcherService) CacheSv1LoadCache(args utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1LoadCache, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.APIOpts,
	}, utils.MetaCaches, utils.CacheSv1LoadCache, args, reply)
}

// CacheSv1ReplicateRemove remove an item
func (dS *DispatcherService) CacheSv1ReplicateRemove(args *utils.ArgCacheReplicateRemove, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1ReplicateRemove, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.Opts,
	}, utils.MetaCaches, utils.CacheSv1ReplicateRemove, args, reply)
}

// CacheSv1ReplicateSet replicate an item
func (dS *DispatcherService) CacheSv1ReplicateSet(args *utils.ArgCacheReplicateSet, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CacheSv1ReplicateSet, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.Opts,
	}, utils.MetaCaches, utils.CacheSv1ReplicateSet, args, reply)
}
