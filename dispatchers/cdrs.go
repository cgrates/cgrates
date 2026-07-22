// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// CDRsV1Ping interogates CDRsV1 server responsible to process the event
func (dS *DispatcherService) CDRsV1Ping(ctx *context.Context, args *utils.CGREvent,
	reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1Ping, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaCDRs,
		utils.CDRsV1Ping, args, reply)
}

// CDRsV1GetCDRs returns the CDRs that match the filter
func (dS *DispatcherService) CDRsV1GetCDRs(ctx *context.Context, args *utils.RPCCDRsFilterWithAPIOpts, reply *[]*engine.CDR) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1GetCDRs, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaCDRs, utils.CDRsV1GetCDRs, args, reply)
}

// CDRsV1GetCDRsCount counts the cdrs that match the filter
func (dS *DispatcherService) CDRsV1GetCDRsCount(ctx *context.Context, args *utils.RPCCDRsFilterWithAPIOpts, reply *int64) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1GetCDRsCount, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaCDRs, utils.CDRsV1GetCDRsCount, args, reply)
}

func (dS *DispatcherService) CDRsV1StoreSessionCost(ctx *context.Context, args *engine.AttrCDRSStoreSMCost, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1StoreSessionCost, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaCDRs, utils.CDRsV1StoreSessionCost, args, reply)
}

func (dS *DispatcherService) CDRsV1RateCDRs(ctx *context.Context, args *engine.ArgRateCDRs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1RateCDRs, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaCDRs, utils.CDRsV1RateCDRs, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessExternalCDR(ctx *context.Context, args *engine.ExternalCDRWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1ProcessExternalCDR, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaCDRs, utils.CDRsV1ProcessExternalCDR, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessEvent(ctx *context.Context, args *engine.ArgV1ProcessEvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1ProcessEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaCDRs,
		utils.CDRsV1ProcessEvent, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessCDR(ctx *context.Context, args *engine.CDRWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1ProcessCDR, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaCDRs, utils.CDRsV1ProcessCDR, args, reply)
}

func (dS *DispatcherService) CDRsV2ProcessEvent(ctx *context.Context, args *engine.ArgV1ProcessEvent, reply *[]*utils.EventWithFlags) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = dS.cfg.GeneralCfg().DefaultTenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV2ProcessEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaCDRs,
		utils.CDRsV2ProcessEvent, args, reply)
}

func (dS *DispatcherService) CDRsV2StoreSessionCost(ctx *context.Context, args *engine.ArgsV2CDRSStoreSMCost, reply *string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = dS.cfg.GeneralCfg().DefaultTenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV2StoreSessionCost, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaCDRs, utils.CDRsV2StoreSessionCost, args, reply)
}
