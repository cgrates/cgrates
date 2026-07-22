// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) IPsV1Ping(ctx *context.Context, args *utils.CGREvent, rpl *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1Ping, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaIPs, utils.IPsV1Ping, args, rpl)
}

func (dS *DispatcherService) IPsV1GetIPAllocationForEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.IPAllocations) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1GetIPAllocationForEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaIPs, utils.IPsV1GetIPAllocationForEvent, args, reply)
}

func (dS *DispatcherService) IPsV1AuthorizeIP(ctx *context.Context, args *utils.CGREvent,
	reply *engine.AllocatedIP) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1AuthorizeIP, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaIPs, utils.IPsV1AuthorizeIP, args, reply)
}

func (dS *DispatcherService) IPsV1AllocateIP(ctx *context.Context, args *utils.CGREvent,
	reply *engine.AllocatedIP) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1AllocateIP, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaIPs, utils.IPsV1AllocateIP, args, reply)
}

func (dS *DispatcherService) IPsV1ReleaseIP(ctx *context.Context, args *utils.CGREvent,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1ReleaseIP, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaIPs, utils.IPsV1ReleaseIP, args, reply)
}

func (dS *DispatcherService) IPsV1ClearIPAllocations(ctx *context.Context, args *engine.ClearIPAllocationsArgs,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1ClearIPAllocations, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaIPs, utils.IPsV1ClearIPAllocations, args, reply)
}

func (dS *DispatcherService) IPsV1GetIPAllocations(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.IPAllocations) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1GetIPAllocations, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaIPs, utils.IPsV1GetIPAllocations, args, reply)
}
