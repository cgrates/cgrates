/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

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

func (dS *DispatcherService) IPsV1GetIPsForEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.IPs) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1GetIPsForEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaIPs, utils.IPsV1GetIPsForEvent, args, reply)
}

func (dS *DispatcherService) IPsV1AuthorizeIPs(ctx *context.Context, args *utils.CGREvent,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1AuthorizeIPs, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaIPs, utils.IPsV1AuthorizeIPs, args, reply)
}

func (dS *DispatcherService) IPsV1AllocateIPs(ctx *context.Context, args *utils.CGREvent,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1AllocateIPs, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaIPs, utils.IPsV1AllocateIPs, args, reply)
}

func (dS *DispatcherService) IPsV1ReleaseIPs(ctx *context.Context, args *utils.CGREvent,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1ReleaseIPs, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaIPs, utils.IPsV1ReleaseIPs, args, reply)
}

func (dS *DispatcherService) IPsV1GetIP(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.IP) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.IPsV1GetIP, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaIPs, utils.IPsV1GetIP, args, reply)
}
