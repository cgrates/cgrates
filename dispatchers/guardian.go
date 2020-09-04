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
func (dS *DispatcherService) GuardianSv1Ping(args *utils.CGREventWithOpts,
	reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREventWithOpts)
	}
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.GuardianSv1Ping, args.CGREvent.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaGuardian, utils.GuardianSv1Ping, args, reply)
}

// GuardianSv1RemoteLock will lock a key from remote
func (dS *DispatcherService) GuardianSv1RemoteLock(args AttrRemoteLockWithOpts,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.GuardianSv1RemoteLock, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
		},
		Opts: args.Opts,
	}, utils.MetaGuardian, utils.GuardianSv1RemoteLock, args, reply)
}

// GuardianSv1RemoteUnlock will unlock a key from remote based on reference ID
func (dS *DispatcherService) GuardianSv1RemoteUnlock(args AttrRemoteUnlockWithOpts,
	reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.GuardianSv1RemoteUnlock, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
		},
		Opts: args.Opts,
	}, utils.MetaGuardian, utils.GuardianSv1RemoteUnlock, args, reply)
}
