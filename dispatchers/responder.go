// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// ResponderPing interogates Responder server responsible to process the event
func (dS *DispatcherService) ResponderPing(ctx *context.Context, args *utils.CGREvent,
	reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderPing, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaResponder, utils.ResponderPing, args, reply)
}

func (dS *DispatcherService) ResponderGetCost(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts,
	reply *engine.CallCost) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderGetCost, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderGetCost, args, reply)
}

func (dS *DispatcherService) ResponderDebit(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts,
	reply *engine.CallCost) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderDebit, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderDebit, args, reply)
}

func (dS *DispatcherService) ResponderMaxDebit(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts,
	reply *engine.CallCost) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderMaxDebit, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderMaxDebit, args, reply)
}

func (dS *DispatcherService) ResponderRefundIncrements(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts,
	reply *engine.Account) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderRefundIncrements, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderRefundIncrements, args, reply)
}

func (dS *DispatcherService) ResponderRefundRounding(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts,
	reply *engine.Account) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderRefundRounding, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderRefundRounding, args, reply)
}

func (dS *DispatcherService) ResponderGetMaxSessionTime(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts,
	reply *time.Duration) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderGetMaxSessionTime, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderGetMaxSessionTime, args, reply)
}

func (dS *DispatcherService) ResponderShutdown(ctx *context.Context, args *utils.TenantWithAPIOpts,
	reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderShutdown, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaResponder, utils.ResponderShutdown, args, reply)
}

func (dS *DispatcherService) ResponderGetCostOnRatingPlans(ctx *context.Context, arg *utils.GetCostOnRatingPlansArgs, reply *map[string]any) (err error) {
	tnt := utils.FirstNonEmpty(arg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderGetCostOnRatingPlans, tnt,
			utils.IfaceAsString(arg.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: arg.APIOpts,
	}, utils.MetaResponder, utils.ResponderGetCostOnRatingPlans, arg, reply)
}

func (dS *DispatcherService) ResponderGetMaxSessionTimeOnAccounts(ctx *context.Context, arg *utils.GetMaxSessionTimeOnAccountsArgs, reply *map[string]any) (err error) {
	tnt := utils.FirstNonEmpty(arg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderGetMaxSessionTimeOnAccounts, tnt,
			utils.IfaceAsString(arg.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: arg.APIOpts,
	}, utils.MetaResponder, utils.ResponderGetMaxSessionTimeOnAccounts, arg, reply)
}
