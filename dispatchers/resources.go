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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ResourceSv1Ping(args *utils.CGREventWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.CGREventWithOpts)
	}
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1Ping, args.CGREvent.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: args.CGREvent,
		Opts:     args.Opts,
	}, utils.MetaResources, utils.ResourceSv1Ping, args, rpl)
}

func (dS *DispatcherService) ResourceSv1GetResourcesForEvent(args utils.ArgRSv1ResourceUsage,
	reply *engine.Resources) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1GetResourcesForEvent, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: args.CGREvent,
		Opts:     args.Opts,
	}, utils.MetaResources, utils.ResourceSv1GetResourcesForEvent, args, reply)
}

func (dS *DispatcherService) ResourceSv1AuthorizeResources(args utils.ArgRSv1ResourceUsage,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1AuthorizeResources, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: args.CGREvent,
		Opts:     args.Opts,
	}, utils.MetaResources, utils.ResourceSv1AuthorizeResources, args, reply)
}

func (dS *DispatcherService) ResourceSv1AllocateResources(args utils.ArgRSv1ResourceUsage,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1AllocateResources, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: args.CGREvent,
		Opts:     args.Opts,
	}, utils.MetaResources, utils.ResourceSv1AllocateResources, args, reply)
}

func (dS *DispatcherService) ResourceSv1ReleaseResources(args utils.ArgRSv1ResourceUsage,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1ReleaseResources, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: args.CGREvent,
		Opts:     args.Opts,
	}, utils.MetaResources, utils.ResourceSv1ReleaseResources, args, reply)
}

func (dS *DispatcherService) ResourceSv1GetResource(args *utils.TenantIDWithOpts, reply *engine.Resource) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1GetResource, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaResources, utils.ResourceSv1GetResource, args, reply)
}
