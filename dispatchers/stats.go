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

// do not modify this code because it's generated
package dispatchers

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) StatSv1GetQueueDecimalMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]*utils.Decimal) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.StatSv1GetQueueDecimalMetrics, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaStats, utils.StatSv1GetQueueDecimalMetrics, args, reply)
}
func (dS *DispatcherService) StatSv1GetQueueFloatMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]float64) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.StatSv1GetQueueFloatMetrics, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaStats, utils.StatSv1GetQueueFloatMetrics, args, reply)
}
func (dS *DispatcherService) StatSv1GetQueueIDs(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.StatSv1GetQueueIDs, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaStats, utils.StatSv1GetQueueIDs, args, reply)
}
func (dS *DispatcherService) StatSv1GetQueueStringMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.StatSv1GetQueueStringMetrics, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaStats, utils.StatSv1GetQueueStringMetrics, args, reply)
}
func (dS *DispatcherService) StatSv1GetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.StatSv1GetStatQueue, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaStats, utils.StatSv1GetStatQueue, args, reply)
}
func (dS *DispatcherService) StatSv1GetStatQueuesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.StatSv1GetStatQueuesForEvent, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaStats, utils.StatSv1GetStatQueuesForEvent, args, reply)
}
func (dS *DispatcherService) StatSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.StatSv1Ping, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaStats, utils.StatSv1Ping, args, reply)
}
func (dS *DispatcherService) StatSv1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.StatSv1ProcessEvent, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaStats, utils.StatSv1ProcessEvent, args, reply)
}
func (dS *DispatcherService) StatSv1ResetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.StatSv1ResetStatQueue, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaStats, utils.StatSv1ResetStatQueue, args, reply)
}
