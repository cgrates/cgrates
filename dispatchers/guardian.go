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

// GuardianSv1Ping interogates GuardianSv1 server responsible to process the event
func (dS *DispatcherService) GuardianSv1Ping(args *CGREvWithApiKey,
	reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.GuardianSv1Ping,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaGuardian, args.RouteID,
		utils.GuardianSv1Ping, args.CGREvent, reply)
}

// RemoteLock will lock a key from remote
func (dS *DispatcherService) GuardianSv1RemoteLock(args *AttrRemoteLockWithApiKey,
	reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.GuardianSv1RemoteLock,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaGuardian, args.RouteID,
		utils.GuardianSv1RemoteLock, args.AttrRemoteLock, reply)
}

// RemoteUnlock will unlock a key from remote based on reference ID
func (dS *DispatcherService) GuardianSv1RemoteUnlock(args *AttrRemoteUnlockWithApiKey,
	reply *[]string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.GuardianSv1RemoteUnlock,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaGuardian, args.RouteID,
		utils.GuardianSv1RemoteUnlock, args.RefID, reply)
}
