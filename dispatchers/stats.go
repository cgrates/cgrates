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

func (dS *DispatcherService) StatSv1Ping(args *CGREvWithApiKey, reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.StatSv1Ping,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaStats, args.RouteID,
		utils.StatSv1Ping, args.CGREvent, reply)
}

func (dS *DispatcherService) StatSv1GetStatQueuesForEvent(args *ArgsStatProcessEventWithApiKey,
	reply *[]string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.StatSv1GetStatQueuesForEvent,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaStats, args.RouteID,
		utils.StatSv1GetStatQueuesForEvent, args.StatsArgsProcessEvent, reply)
}

func (dS *DispatcherService) StatSv1GetQueueStringMetrics(args *TntIDWithApiKey,
	reply *map[string]string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.StatSv1GetQueueStringMetrics,
			args.TenantID.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		ID:     args.ID,
	}, utils.MetaStats, args.RouteID, utils.StatSv1GetQueueStringMetrics,
		args.TenantID, reply)
}

func (dS *DispatcherService) StatSv1ProcessEvent(args *ArgsStatProcessEventWithApiKey,
	reply *[]string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.StatSv1ProcessEvent,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaStats, args.RouteID,
		utils.StatSv1ProcessEvent, args.StatsArgsProcessEvent, reply)
}
