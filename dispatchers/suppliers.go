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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) SupplierSv1Ping(args *utils.CGREventWithArgDispatcher, reply *string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing("ArgDispatcher")
		}
		if err = dS.authorize(utils.SupplierSv1Ping,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSuppliers, routeID,
		utils.SupplierSv1Ping, args, reply)
}

func (dS *DispatcherService) SupplierSv1GetSuppliers(args *engine.ArgsGetSuppliers,
	reply *engine.SortedSuppliers) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing("ArgDispatcher")
		}
		if err = dS.authorize(utils.SupplierSv1GetSuppliers,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaSuppliers, routeID,
		utils.SupplierSv1GetSuppliers, args, reply)
}

func (dS *DispatcherService) SupplierSv1GetSupplierProfilesForEvent(args *utils.CGREventWithArgDispatcher,
	reply *[]*engine.SupplierProfile) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing("ArgDispatcher")
		}
		if err = dS.authorize(utils.SupplierSv1GetSupplierProfilesForEvent,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSuppliers, routeID,
		utils.SupplierSv1GetSupplierProfilesForEvent, args, reply)
}
