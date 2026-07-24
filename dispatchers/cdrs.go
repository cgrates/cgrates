// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// CDRsV1Ping interogates CDRsV1 server responsible to process the event
func (dS *DispatcherService) CDRsV1Ping(args *utils.CGREventWithArgDispatcher,
	reply *string) (err error) {
	if args == nil || (args.CGREvent == nil && args.ArgDispatcher == nil) {
		args = utils.NewCGREventWithArgDispatcher()
	} else if args.CGREvent == nil {
		args.CGREvent = new(utils.CGREvent)
	}
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CDRsV1Ping, tnt,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaCDRs, routeID,
		utils.CDRsV1Ping, args, reply)
}

func (dS *DispatcherService) CDRsV1GetCDRs(args utils.RPCCDRsFilterWithArgDispatcher, reply *[]*engine.CDR) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantArg != nil && args.TenantArg.Tenant != utils.EmptyString {
		tnt = args.TenantArg.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CDRsV1GetCDRs, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCDRs, routeID,
		utils.CDRsV1GetCDRs, args, reply)
}

func (dS *DispatcherService) CDRsV1GetCDRsCount(args *utils.RPCCDRsFilterWithArgDispatcher, reply *int64) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantArg != nil && args.TenantArg.Tenant != utils.EmptyString {
		tnt = args.TenantArg.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CDRsV1GetCDRsCount, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCDRs, routeID,
		utils.CDRsV1GetCDRsCount, args, reply)
}

func (dS *DispatcherService) CDRsV1StoreSessionCost(args *engine.AttrCDRSStoreSMCost, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantArg != nil && args.TenantArg.Tenant != utils.EmptyString {
		tnt = args.TenantArg.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CDRsV1StoreSessionCost, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCDRs, routeID,
		utils.CDRsV1StoreSessionCost, args, reply)
}

func (dS *DispatcherService) CDRsV1RateCDRs(args *engine.ArgRateCDRs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantArg != nil && args.TenantArg.Tenant != utils.EmptyString {
		tnt = args.TenantArg.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CDRsV1RateCDRs, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCDRs, routeID,
		utils.CDRsV1RateCDRs, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessExternalCDR(args *engine.ExternalCDRWithArgDispatcher, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CDRsV1ProcessExternalCDR, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCDRs, routeID,
		utils.CDRsV1ProcessExternalCDR, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessEvent(args *engine.ArgV1ProcessEvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CDRsV1ProcessEvent, tnt,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaCDRs, routeID,
		utils.CDRsV1ProcessEvent, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessCDR(args *engine.CDRWithArgDispatcher, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.CDRsV1ProcessCDR, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaCDRs, routeID,
		utils.CDRsV1ProcessCDR, args, reply)
}
