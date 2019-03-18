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
func (dS *DispatcherService) ResponderPing(args *CGREvWithApiKey,
	reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderPing,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaResponder, args.RouteID,
		utils.ResponderPing, args.CGREvent, reply)
}

func (dS *DispatcherService) ResponderStatus(args *TntWithApiKey,
	reply *map[string]interface{}) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderStatus, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
	}, utils.MetaResponder, args.RouteID, utils.ResponderStatus,
		"", reply)
}

func (dS *DispatcherService) ResponderGetCost(args *CallDescriptorWithApiKey,
	reply *engine.CallCost) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderGetCost, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CallDescriptor.AsCGREvent(), utils.MetaResponder,
		args.RouteID, utils.ResponderGetCost, args.CallDescriptor, reply)
}

func (dS *DispatcherService) ResponderDebit(args *CallDescriptorWithApiKey,
	reply *engine.CallCost) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderDebit, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CallDescriptor.AsCGREvent(), utils.MetaResponder,
		args.RouteID, utils.ResponderDebit, args.CallDescriptor, reply)
}

func (dS *DispatcherService) ResponderMaxDebit(args *CallDescriptorWithApiKey,
	reply *engine.CallCost) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderMaxDebit, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CallDescriptor.AsCGREvent(), utils.MetaResponder,
		args.RouteID, utils.ResponderMaxDebit, args.CallDescriptor, reply)
}

func (dS *DispatcherService) ResponderRefundIncrements(args *CallDescriptorWithApiKey,
	reply *engine.Account) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderRefundIncrements, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CallDescriptor.AsCGREvent(), utils.MetaResponder,
		args.RouteID, utils.ResponderRefundIncrements, args.CallDescriptor, reply)
}

func (dS *DispatcherService) ResponderRefundRounding(args *CallDescriptorWithApiKey,
	reply *float64) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderRefundRounding, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CallDescriptor.AsCGREvent(), utils.MetaResponder,
		args.RouteID, utils.ResponderRefundRounding, args.CallDescriptor, reply)
}

func (dS *DispatcherService) ResponderGetMaxSessionTime(args *CallDescriptorWithApiKey,
	reply *time.Duration) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderGetMaxSessionTime, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CallDescriptor.AsCGREvent(), utils.MetaResponder,
		args.RouteID, utils.ResponderGetMaxSessionTime, args.CallDescriptor, reply)
}

func (dS *DispatcherService) ResponderShutdown(args *TntWithApiKey,
	reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderShutdown, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
	}, utils.MetaResponder, args.RouteID, utils.ResponderShutdown,
		"", reply)
}

func (dS *DispatcherService) ResponderGetTimeout(args *TntWithApiKey,
	reply *time.Duration) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.ResponderGetTimeout, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
	}, utils.MetaResponder, args.RouteID, utils.ResponderGetTimeout,
		0, reply)
}
