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

func (dS *DispatcherService) CoreSv1Panic(args *utils.PanicMessageArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Panic, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Panic, args, reply)
}
func (dS *DispatcherService) CoreSv1Ping(args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Ping, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Ping, args, reply)
}

func (dS *DispatcherService) CoreSv1Sleep(args *utils.DurationArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Sleep, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Sleep, args, reply)
}

func (dS *DispatcherService) CoreSv1StartCPUProfiling(args *utils.DirectoryArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1StartCPUProfiling, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StartCPUProfiling, args, reply)
}
func (dS *DispatcherService) CoreSv1StartMemoryProfiling(args *utils.MemoryPrf, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1StartMemoryProfiling, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StartMemoryProfiling, args, reply)
}
func (dS *DispatcherService) CoreSv1Status(args *utils.TenantWithAPIOpts, reply *map[string]interface{}) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Status, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Status, args, reply)
}
func (dS *DispatcherService) CoreSv1StopCPUProfiling(args *utils.TenantWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1StopCPUProfiling,
			tnt, utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StopCPUProfiling, args, reply)
}
func (dS *DispatcherService) CoreSv1StopMemoryProfiling(args *utils.TenantWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1StopMemoryProfiling,
			tnt, utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StopMemoryProfiling, args, reply)
}
