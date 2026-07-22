// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) RALsV1Ping(ctx *context.Context, args *utils.CGREvent, rpl *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RALsV1Ping, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaRALs, utils.RALsV1Ping, args, rpl)
}

func (dS *DispatcherService) RALsV1GetRatingPlansCost(ctx *context.Context, args *utils.RatingPlanCostArg, rpl *RatingPlanCost) (err error) {
	tenant := dS.cfg.GeneralCfg().DefaultTenant
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RALsV1GetRatingPlansCost, tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaRALs, utils.RALsV1GetRatingPlansCost, args, rpl)
}
