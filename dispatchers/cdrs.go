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

// CDRsV1Ping interogates CDRsV1 server responsible to process the event
func (dS *DispatcherService) CDRsV1Ping(args *utils.CGREventWithArgDispatcher,
	reply *string) (err error) {
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing("ArgDispatcher")
	}
	if dS.attrS != nil {
		if err = dS.authorize(utils.CDRsV1Ping,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaCDRs, args.RouteID,
		utils.CDRsV1Ping, args.CGREvent, reply)
}

func (dS *DispatcherService) CDRsV1GetCDRs(args utils.RPCCDRsFilterWithArgDispatcher, reply *[]*engine.CDR) (err error) {
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing("ArgDispatcher")
	}
	if dS.attrS != nil {
		if err = dS.authorize(utils.CDRsV1GetCDRs,
			args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaCDRs, args.RouteID,
		utils.CDRsV1GetCDRs, args, reply)
}

func (dS *DispatcherService) CDRsV1CountCDRs(args *utils.RPCCDRsFilterWithArgDispatcher, reply *int64) (err error) {
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing("ArgDispatcher")
	}
	if dS.attrS != nil {
		if err = dS.authorize(utils.CDRsV1CountCDRs,
			args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaCDRs, args.RouteID,
		utils.CDRsV1CountCDRs, args, reply)
}

func (dS *DispatcherService) CDRsV1StoreSessionCost(args *engine.AttrCDRSStoreSMCost, reply *string) (err error) {
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing("ArgDispatcher")
	}
	if dS.attrS != nil {
		if err = dS.authorize(utils.CDRsV1StoreSessionCost,
			args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaCDRs, args.RouteID,
		utils.CDRsV1StoreSessionCost, args, reply)
}

func (dS *DispatcherService) CDRsV1RateCDRs(args *engine.ArgRateCDRs, reply *string) (err error) {
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing("ArgDispatcher")
	}
	if dS.attrS != nil {
		if err = dS.authorize(utils.CDRsV1RateCDRs,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaCDRs, args.RouteID,
		utils.CDRsV1RateCDRs, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessExternalCDR(args *engine.ExternalCDRWithArgDispatcher, reply *string) (err error) {
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing("ArgDispatcher")
	}
	if dS.attrS != nil {
		if err = dS.authorize(utils.CDRsV1ProcessExternalCDR,
			args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaCDRs, args.RouteID,
		utils.CDRsV1ProcessExternalCDR, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessEvent(args *engine.ArgV1ProcessEvent, reply *string) (err error) {
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing("ArgDispatcher")
	}
	if dS.attrS != nil {
		if err = dS.authorize(utils.CDRsV1ProcessEvent,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaCDRs, args.RouteID,
		utils.CDRsV1ProcessEvent, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessCDR(args *engine.CDRWithArgDispatcher, reply *string) (err error) {
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing("ArgDispatcher")
	}
	if dS.attrS != nil {
		if err = dS.authorize(utils.CDRsV1ProcessCDR,
			args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaCDRs, args.RouteID,
		utils.CDRsV1ProcessCDR, args, reply)
}
