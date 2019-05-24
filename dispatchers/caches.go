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
func (dS *DispatcherService) CacheSv1Ping(args *utils.CGREventWithArgDispatcher,
	reply *string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1Ping, args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaCaches, routeID,
		utils.CacheSv1Ping, args, reply)
}

// GetItemIDs returns the IDs for cacheID with given prefix
func (dS *DispatcherService) CacheSv1GetItemIDs(args *utils.ArgsGetCacheItemIDsWithArgDispatcher,
	reply *[]string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1GetItemIDs, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1GetItemIDs, &args.ArgsGetCacheItemIDs, reply)
}

// HasItem verifies the existence of an Item in cache
func (dS *DispatcherService) CacheSv1HasItem(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *bool) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1HasItem, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1HasItem, &args.ArgsGetCacheItem, reply)
}

// GetItemExpiryTime returns the expiryTime for an item
func (dS *DispatcherService) CacheSv1GetItemExpiryTime(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *time.Time) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1GetItemExpiryTime, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1GetItemExpiryTime, &args.ArgsGetCacheItem, reply)
}

// RemoveItem removes the Item with ID from cache
func (dS *DispatcherService) CacheSv1RemoveItem(args *utils.ArgsGetCacheItemWithArgDispatcher,
	reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1RemoveItem, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1RemoveItem, &args.ArgsGetCacheItem, reply)
}

// Clear will clear partitions in the cache (nil fol all, empty slice for none)
func (dS *DispatcherService) CacheSv1Clear(args *utils.AttrCacheIDsWithArgDispatcher,
	reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1Clear, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1Clear, args.CacheIDs, reply)
}

// FlushCache wipes out cache for a prefix or completely
func (dS *DispatcherService) CacheSv1FlushCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1FlushCache, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1FlushCache, args.AttrReloadCache, reply)
}

// GetCacheStats returns CacheStats filtered by cacheIDs
func (dS *DispatcherService) CacheSv1GetCacheStats(args *utils.AttrCacheIDsWithArgDispatcher,
	reply *map[string]*ltcache.CacheStats) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1GetCacheStats, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1GetCacheStats, args.CacheIDs, reply)
}

// PrecacheStatus checks status of active precache processes
func (dS *DispatcherService) CacheSv1PrecacheStatus(args *utils.AttrCacheIDsWithArgDispatcher, reply *map[string]string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1PrecacheStatus, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1PrecacheStatus, args.CacheIDs, reply)
}

// HasGroup checks existence of a group in cache
func (dS *DispatcherService) CacheSv1HasGroup(args *utils.ArgsGetGroupWithArgDispatcher,
	reply *bool) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1HasGroup, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1HasGroup, args.ArgsGetGroup, reply)
}

// GetGroupItemIDs returns a list of itemIDs in a cache group
func (dS *DispatcherService) CacheSv1GetGroupItemIDs(args *utils.ArgsGetGroupWithArgDispatcher,
	reply *[]string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1GetGroupItemIDs, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1GetGroupItemIDs, args.ArgsGetGroup, reply)
}

// RemoveGroup will remove a group and all items belonging to it from cache
func (dS *DispatcherService) CacheSv1RemoveGroup(args *utils.ArgsGetGroupWithArgDispatcher, reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1RemoveGroup, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1RemoveGroup, args.ArgsGetGroup, reply)
}

// ReloadCache reloads cache from DB for a prefix or completely
func (dS *DispatcherService) CacheSv1ReloadCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1ReloadCache, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1ReloadCache, args.AttrReloadCache, reply)
}

// LoadCache loads cache from DB for a prefix or completely
func (dS *DispatcherService) CacheSv1LoadCache(args utils.AttrReloadCacheWithArgDispatcher, reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CacheSv1LoadCache, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCaches, routeID,
		utils.CacheSv1LoadCache, args.AttrReloadCache, reply)
}
