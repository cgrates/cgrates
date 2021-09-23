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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ResourceSv1Ping(args *utils.CGREvent, rpl *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1Ping, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaResources, utils.ResourceSv1Ping, args, rpl)
}

func (dS *DispatcherService) ResourceSv1GetResourcesForEvent(args *utils.CGREvent,
	reply *engine.Resources) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1GetResourcesForEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaResources, utils.ResourceSv1GetResourcesForEvent, args, reply)
}

func (dS *DispatcherService) ResourceSv1AuthorizeResources(args *utils.CGREvent,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1AuthorizeResources, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaResources, utils.ResourceSv1AuthorizeResources, args, reply)
}

func (dS *DispatcherService) ResourceSv1AllocateResources(args *utils.CGREvent,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1AllocateResources, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaResources, utils.ResourceSv1AllocateResources, args, reply)
}

func (dS *DispatcherService) ResourceSv1ReleaseResources(args *utils.CGREvent,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1ReleaseResources, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaResources, utils.ResourceSv1ReleaseResources, args, reply)
}

func (dS *DispatcherService) ResourceSv1GetResource(args *utils.TenantIDWithAPIOpts, reply *engine.Resource) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1GetResource, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaResources, utils.ResourceSv1GetResource, args, reply)
}

func (dS *DispatcherService) ResourceSv1GetResourceWithConfig(args *utils.TenantIDWithAPIOpts, reply *engine.ResourceWithConfig) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResourceSv1GetResourceWithConfig, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaResources, utils.ResourceSv1GetResourceWithConfig, args, reply)
}
