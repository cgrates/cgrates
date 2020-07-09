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

	"github.com/cgrates/cgrates/utils"
)

// ServiceManagerV1Ping interogates ServiceManager server responsible to process the event
func (dS *DispatcherService) ServiceManagerV1Ping(args *utils.CGREventWithArgDispatcher,
	reply *string) (err error) {
	if args == nil {
		args = utils.NewCGREventWithArgDispatcher()
	}
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ServiceManagerV1Ping,
			args.Tenant,
			args.OptsAPIKey, args.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.OptsRouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaServiceManager, routeID,
		utils.ServiceManagerV1Ping, args, reply)
}

func (dS *DispatcherService) ServiceManagerV1StartService(args ArgStartServiceWithApiKey,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantArg.Tenant != utils.EmptyString {
		tnt = args.TenantArg.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ServiceManagerV1StartService, tnt,
			args.OptsAPIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.OptsRouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaServiceManager, routeID,
		utils.ServiceManagerV1StartService, args, reply)
}

func (dS *DispatcherService) ServiceManagerV1StopService(args ArgStartServiceWithApiKey,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantArg.Tenant != utils.EmptyString {
		tnt = args.TenantArg.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ServiceManagerV1StopService, tnt,
			args.OptsAPIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.OptsRouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaServiceManager, routeID,
		utils.ServiceManagerV1StopService, args, reply)
}

func (dS *DispatcherService) ServiceManagerV1ServiceStatus(args ArgStartServiceWithApiKey,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantArg.Tenant != utils.EmptyString {
		tnt = args.TenantArg.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ServiceManagerV1ServiceStatus, tnt,
			args.OptsAPIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.OptsRouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaServiceManager, routeID,
		utils.ServiceManagerV1ServiceStatus, args, reply)
}
