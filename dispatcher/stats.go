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

package dispatcher

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) StatSv1Ping(ign string, reply *string) error {
	if dS.statS == nil {
		return utils.NewErrNotConnected(utils.StatS)
	}
	return dS.statS.Call(utils.StatSv1Ping, ign, reply)
}

func (dS *DispatcherService) StatSv1GetStatQueuesForEvent(args *CGREvWithApiKey,
	reply *[]string) (err error) {
	if dS.statS == nil {
		return utils.NewErrNotConnected(utils.StatS)
	}
	if err = dS.authorizeMethod(args.APIKey, args.CGREvent.Tenant,
		utils.StatSv1GetStatQueuesForEvent, args.CGREvent.Time); err != nil {
		return
	}
	return dS.statS.Call(utils.StatSv1GetStatQueuesForEvent, args.CGREvent, reply)
}

func (dS *DispatcherService) StatSv1GetQueueStringMetrics(args *TntIDWithApiKey,
	reply *map[string]string) (err error) {
	if dS.statS == nil {
		return utils.NewErrNotConnected(utils.StatS)
	}
	nowTime := time.Now()
	if err = dS.authorizeMethod(args.APIKey, args.TenantID.Tenant,
		utils.StatSv1GetQueueStringMetrics, &nowTime); err != nil {
		return
	}
	return dS.statS.Call(utils.StatSv1GetQueueStringMetrics, args.TenantID, reply)
}

func (dS *DispatcherService) StatSv1ProcessEvent(args *CGREvWithApiKey,
	reply *[]string) (err error) {
	if dS.statS == nil {
		return utils.NewErrNotConnected(utils.StatS)
	}
	if err = dS.authorizeMethod(args.APIKey, args.CGREvent.Tenant,
		utils.StatSv1ProcessEvent, args.CGREvent.Time); err != nil {
		return
	}
	return dS.statS.Call(utils.StatSv1ProcessEvent, args.CGREvent, reply)
}
