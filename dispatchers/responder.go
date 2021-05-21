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
	} else if args.ArgDispatcher == nil {
		args.ArgDispatcher = new(utils.ArgDispatcher)
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
	reply *float64) (err error) {
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

func (dS *DispatcherService) ResponderGetCostOnRatingPlans(arg *utils.GetCostOnRatingPlansArgs, reply *map[string]interface{}) (err error) {
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
func (dS *DispatcherService) ResponderGetMaxSessionTimeOnAccounts(args *utils.GetMaxSessionTimeOnAccountsArgs, reply *map[string]interface{}) (err error) {
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
