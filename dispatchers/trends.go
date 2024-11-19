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

func (dS *DispatcherService) TrendSv1GetScheduledTrends(ctx *context.Context, args *utils.ArgScheduledTrends, reply *[]utils.ScheduledTrend) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantIDWithAPIOpts.TenantID != nil && len(args.TenantIDWithAPIOpts.TenantID.Tenant) != 0) {
		tnt = args.TenantIDWithAPIOpts.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.TenantIDWithAPIOpts.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaTrends, utils.TrendSv1GetScheduledTrends, args, reply)
}
func (dS *DispatcherService) TrendSv1GetTrend(ctx *context.Context, args *utils.ArgGetTrend, reply *engine.Trend) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.TenantWithAPIOpts.Tenant) != 0 {
		tnt = args.TenantWithAPIOpts.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.TenantWithAPIOpts.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaTrends, utils.TrendSv1GetTrend, args, reply)
}
func (dS *DispatcherService) TrendSv1GetTrendSummary(ctx *context.Context, args utils.TenantIDWithAPIOpts, reply *engine.TrendSummary) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && len(args.TenantID.Tenant) != 0 {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := args.APIOpts
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaTrends, utils.TrendSv1GetTrendSummary, args, reply)
}
func (dS *DispatcherService) TrendSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaTrends, utils.TrendSv1Ping, args, reply)
}
func (dS *DispatcherService) TrendSv1ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleTrendQueries, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantIDWithAPIOpts.TenantID != nil && len(args.TenantIDWithAPIOpts.TenantID.Tenant) != 0) {
		tnt = args.TenantIDWithAPIOpts.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.TenantIDWithAPIOpts.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaTrends, utils.TrendSv1ScheduleQueries, args, reply)
}
