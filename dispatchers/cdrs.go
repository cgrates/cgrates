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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// CDRsV1Ping interogates CDRsV1 server responsible to process the event
func (dS *DispatcherService) CDRsV1Ping(args *utils.CGREvent,
	reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1Ping, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaCDRs,
		utils.CDRsV1Ping, args, reply)
}

func (dS *DispatcherService) CDRsV1ProcessEvent(args *engine.ArgV1ProcessEvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV1ProcessEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &args.CGREvent, utils.MetaCDRs,
		utils.CDRsV1ProcessEvent, args, reply)
}

func (dS *DispatcherService) CDRsV2ProcessEvent(args *engine.ArgV1ProcessEvent, reply *[]*utils.EventWithFlags) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = dS.cfg.GeneralCfg().DefaultTenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CDRsV2ProcessEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &args.CGREvent, utils.MetaCDRs,
		utils.CDRsV2ProcessEvent, args, reply)
}
