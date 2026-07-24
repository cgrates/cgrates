// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// ResponderPing interogates Responder server responsible to process the event
func (dS *DispatcherService) ResponderPing(args *utils.CGREventWithArgDispatcher,
	reply *string) (err error) {
	if args == nil || (args.CGREvent == nil && args.ArgDispatcher == nil) {
		args = utils.NewCGREventWithArgDispatcher()
	} else if args.CGREvent == nil {
		args.CGREvent = new(utils.CGREvent)
	}
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderPing, args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaResponder,
		routeID, utils.ResponderPing, args, reply)
}

func (dS *DispatcherService) ResponderGetCost(args *engine.CallDescriptorWithArgDispatcher,
	reply *engine.CallCost) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderGetCost, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.AsCGREvent(), utils.MetaResponder,
		routeID, utils.ResponderGetCost, args, reply)
}

func (dS *DispatcherService) ResponderDebit(args *engine.CallDescriptorWithArgDispatcher,
	reply *engine.CallCost) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderDebit, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.AsCGREvent(), utils.MetaResponder,
		routeID, utils.ResponderDebit, args, reply)
}

func (dS *DispatcherService) ResponderMaxDebit(args *engine.CallDescriptorWithArgDispatcher,
	reply *engine.CallCost) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderMaxDebit, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.AsCGREvent(), utils.MetaResponder,
		routeID, utils.ResponderMaxDebit, args, reply)
}

func (dS *DispatcherService) ResponderRefundIncrements(args *engine.CallDescriptorWithArgDispatcher,
	reply *engine.Account) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderRefundIncrements, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.AsCGREvent(), utils.MetaResponder,
		routeID, utils.ResponderRefundIncrements, args, reply)
}

func (dS *DispatcherService) ResponderRefundRounding(args *engine.CallDescriptorWithArgDispatcher,
	reply *engine.Account) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderRefundRounding, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.AsCGREvent(), utils.MetaResponder,
		routeID, utils.ResponderRefundRounding, args, reply)
}

func (dS *DispatcherService) ResponderGetMaxSessionTime(args *engine.CallDescriptorWithArgDispatcher,
	reply *time.Duration) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderGetMaxSessionTime, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.AsCGREvent(), utils.MetaResponder,
		routeID, utils.ResponderGetMaxSessionTime, args, reply)
}

func (dS *DispatcherService) ResponderShutdown(args *utils.TenantWithArgDispatcher,
	reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderShutdown, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaResponder,
		routeID, utils.ResponderShutdown, args, reply)
}

func (dS *DispatcherService) ResponderGetCostOnRatingPlans(arg *utils.GetCostOnRatingPlansArgs, reply *map[string]any) (err error) {
	tnt := utils.FirstNonEmpty(arg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if arg.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderGetCostOnRatingPlans, tnt,
			arg.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if arg.ArgDispatcher != nil {
		routeID = arg.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaResponder,
		routeID, utils.ResponderGetCostOnRatingPlans, arg, reply)
}

func (dS *DispatcherService) ResponderGetMaxSessionTimeOnAccounts(args *utils.GetMaxSessionTimeOnAccountsArgs, reply *map[string]any) (err error) {
	tnt := utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ResponderGetMaxSessionTimeOnAccounts, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaResponder,
		routeID, utils.ResponderGetMaxSessionTimeOnAccounts, args, reply)
}
