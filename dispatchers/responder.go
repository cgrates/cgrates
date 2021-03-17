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
func (dS *DispatcherService) ResponderPing(args *utils.CGREvent,
	reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderPing, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaResponder, utils.ResponderPing, args, reply)
}

func (dS *DispatcherService) ResponderGetCost(args *engine.CallDescriptorWithAPIOpts,
	reply *engine.CallCost) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderGetCost, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderGetCost, args, reply)
}

func (dS *DispatcherService) ResponderDebit(args *engine.CallDescriptorWithAPIOpts,
	reply *engine.CallCost) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderDebit, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderDebit, args, reply)
}

func (dS *DispatcherService) ResponderMaxDebit(args *engine.CallDescriptorWithAPIOpts,
	reply *engine.CallCost) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderMaxDebit, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderMaxDebit, args, reply)
}

func (dS *DispatcherService) ResponderRefundIncrements(args *engine.CallDescriptorWithAPIOpts,
	reply *engine.Account) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderRefundIncrements, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderRefundIncrements, args, reply)
}

func (dS *DispatcherService) ResponderRefundRounding(args *engine.CallDescriptorWithAPIOpts,
	reply *float64) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderRefundRounding, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderRefundRounding, args, reply)
}

func (dS *DispatcherService) ResponderGetMaxSessionTime(args *engine.CallDescriptorWithAPIOpts,
	reply *time.Duration) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderGetMaxSessionTime, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(args.AsCGREvent(args.APIOpts), utils.MetaResponder, utils.ResponderGetMaxSessionTime, args, reply)
}

func (dS *DispatcherService) ResponderShutdown(args *utils.TenantWithOpts,
	reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ResponderShutdown, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.Opts,
	}, utils.MetaResponder, utils.ResponderShutdown, args, reply)
}
