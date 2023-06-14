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

func (dS *DispatcherService) StatSv1Ping(args *utils.CGREventWithArgDispatcher, reply *string) (err error) {
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
		if err = dS.authorize(utils.StatSv1Ping,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaStats, routeID,
		utils.StatSv1Ping, args, reply)
}

func (dS *DispatcherService) StatSv1GetStatQueuesForEvent(args *engine.StatsArgsProcessEvent,
	reply *[]string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.StatSv1GetStatQueuesForEvent,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaStats, routeID,
		utils.StatSv1GetStatQueuesForEvent, args, reply)
}

func (dS *DispatcherService) StatSv1GetQueueStringMetrics(args *utils.TenantIDWithArgDispatcher,
	reply *map[string]string) (err error) {

	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.StatSv1GetQueueStringMetrics,
			args.TenantID.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		ID:     args.ID,
	}, utils.MetaStats, routeID, utils.StatSv1GetQueueStringMetrics,
		args, reply)
}

func (dS *DispatcherService) StatSv1ProcessEvent(args *engine.StatsArgsProcessEvent,
	reply *[]string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.StatSv1ProcessEvent,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaStats, routeID,
		utils.StatSv1ProcessEvent, args, reply)
}

func (dS *DispatcherService) StatSv1GetQueueFloatMetrics(args *utils.TenantIDWithArgDispatcher,
	reply *map[string]float64) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.StatSv1GetQueueFloatMetrics,
			args.TenantID.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		ID:     args.ID,
	}, utils.MetaStats, routeID, utils.StatSv1GetQueueFloatMetrics,
		args, reply)
}

func (dS *DispatcherService) StatSv1GetQueueIDs(args *utils.TenantWithArgDispatcher,
	reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantArg != nil && args.TenantArg.Tenant != utils.EmptyString {
		tnt = args.TenantArg.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.StatSv1GetQueueIDs, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt},
		utils.MetaStats, routeID, utils.StatSv1GetQueueIDs,
		args, reply)
}
