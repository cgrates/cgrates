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

func (dS *DispatcherService) ThresholdSv1Ping(args *utils.CGREventWithArgDispatcher, reply *string) (err error) {
	if args == nil {
		args = utils.NewCGREventWithArgDispatcher()
	}
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ThresholdSv1Ping,
			args.CGREvent.Tenant,
			args.OptsAPIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.OptsRouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaThresholds, routeID,
		utils.ThresholdSv1Ping, args, reply)
}

func (dS *DispatcherService) ThresholdSv1GetThresholdsForEvent(args *engine.ArgsProcessEvent,
	t *engine.Thresholds) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ThresholdSv1GetThresholdsForEvent,
			args.CGREvent.Tenant,
			args.OptsAPIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.OptsRouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaThresholds, routeID,
		utils.ThresholdSv1GetThresholdsForEvent, args, t)
}

func (dS *DispatcherService) ThresholdSv1ProcessEvent(args *engine.ArgsProcessEvent,
	tIDs *[]string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ThresholdSv1ProcessEvent,
			args.CGREvent.Tenant,
			args.OptsAPIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.OptsRouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaThresholds, routeID,
		utils.ThresholdSv1ProcessEvent, args, tIDs)
}

func (dS *DispatcherService) ThresholdSv1GetThresholdIDs(args *utils.TenantWithArgDispatcher, tIDs *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantArg != nil && args.TenantArg.Tenant != utils.EmptyString {
		tnt = args.TenantArg.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ThresholdSv1GetThresholdIDs,
			tnt, args.OptsAPIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.OptsRouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaThresholds, routeID,
		utils.ThresholdSv1GetThresholdIDs, args, tIDs)
}

func (dS *DispatcherService) ThresholdSv1GetThreshold(args *utils.TenantIDWithArgDispatcher, th *engine.Threshold) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ThresholdSv1GetThreshold, tnt,
			args.OptsAPIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.OptsRouteID
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaThresholds, routeID, utils.ThresholdSv1GetThreshold, args, th)
}
