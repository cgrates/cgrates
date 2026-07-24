// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) SchedulerSv1Ping(args *utils.CGREventWithArgDispatcher, reply *string) (err error) {
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
		if err = dS.authorize(utils.SchedulerSv1Ping,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaScheduler, routeID,
		utils.SchedulerSv1Ping, args, reply)
}

func (dS *DispatcherService) SchedulerSv1Reload(args *utils.CGREventWithArgDispatcher, reply *string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SchedulerSv1Ping,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaScheduler, routeID,
		utils.SchedulerSv1Reload, args, reply)
}
