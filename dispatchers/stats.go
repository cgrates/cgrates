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

func (dS *DispatcherService) StatSv1Ping(args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.StatSv1Ping,
			args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaStats, utils.StatSv1Ping, args, reply)
}

func (dS *DispatcherService) StatSv1GetStatQueuesForEvent(args *engine.StatsArgsProcessEvent,
	reply *[]string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.StatSv1GetStatQueuesForEvent,
			args.CGREvent.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaStats, utils.StatSv1GetStatQueuesForEvent, args, reply)
}

func (dS *DispatcherService) StatSv1GetQueueStringMetrics(args *utils.TenantIDWithAPIOpts,
	reply *map[string]string) (err error) {

	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.StatSv1GetQueueStringMetrics,
			args.TenantID.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaStats, utils.StatSv1GetQueueStringMetrics, args, reply)
}

func (dS *DispatcherService) StatSv1ProcessEvent(args *engine.StatsArgsProcessEvent,
	reply *[]string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.StatSv1ProcessEvent,
			args.CGREvent.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaStats, utils.StatSv1ProcessEvent, args, reply)
}

func (dS *DispatcherService) StatSv1GetQueueFloatMetrics(args *utils.TenantIDWithAPIOpts,
	reply *map[string]float64) (err error) {
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.StatSv1GetQueueFloatMetrics,
			args.TenantID.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaStats, utils.StatSv1GetQueueFloatMetrics, args, reply)
}

func (dS *DispatcherService) StatSv1GetQueueIDs(args *utils.TenantWithOpts,
	reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.StatSv1GetQueueIDs, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		Opts:   args.Opts,
	}, utils.MetaStats, utils.StatSv1GetQueueIDs, args, reply)
}
