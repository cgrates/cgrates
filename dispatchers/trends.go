// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) TrendSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.TrendSv1Ping, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaTrends, utils.TrendSv1Ping, args, reply)
}
func (dS *DispatcherService) TrendSv1ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleTrendQueries,
	scheduled *int) (err error) {
	tnt := utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.TrendSv1ScheduleQueries,
			tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts}, utils.MetaTrends, utils.TrendSv1ScheduleQueries, args, scheduled)
}

func (dS *DispatcherService) TrendSv1GetTrend(ctx *context.Context, args *utils.ArgGetTrend, trend *engine.Trend) (err error) {
	tnt := utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.TrendSv1GetTrend,
			args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]),
			utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaTrends, utils.TrendSv1GetTrend, args, trend)
}

func (dS *DispatcherService) TrendSv1GetScheduledTrends(ctx *context.Context, args *utils.ArgScheduledTrends, schedTrends *[]utils.ScheduledTrend) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.TrendSv1GetScheduledTrends,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaTrends, utils.TrendSv1GetScheduledTrends, args, schedTrends)
}

func (dS *DispatcherService) TrendSv1GetTrendSummary(ctx *context.Context, args utils.TenantIDWithAPIOpts, reply *engine.TrendSummary) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.TrendSv1GetTrendSummary, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaTrends, utils.TrendSv1GetTrendSummary, args, reply)
}
