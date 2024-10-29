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

// do not modify this code because it's generated
package dispatchers

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) CoreSv1Panic(ctx *context.Context, args *utils.PanicMessageArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Panic, args, reply)
}
func (dS *DispatcherService) CoreSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Ping, args, reply)
}
func (dS *DispatcherService) CoreSv1Shutdown(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Shutdown, args, reply)
}
func (dS *DispatcherService) CoreSv1Sleep(ctx *context.Context, args *utils.DurationArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Sleep, args, reply)
}
func (dS *DispatcherService) CoreSv1StartCPUProfiling(ctx *context.Context, args *utils.DirectoryArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StartCPUProfiling, args, reply)
}
func (dS *DispatcherService) CoreSv1StartMemoryProfiling(ctx *context.Context, args cores.MemoryProfilingParams, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := args.APIOpts
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StartMemoryProfiling, args, reply)
}
func (dS *DispatcherService) CoreSv1Status(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *map[string]any) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Status, args, reply)
}
func (dS *DispatcherService) CoreSv1StopCPUProfiling(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StopCPUProfiling, args, reply)
}
func (dS *DispatcherService) CoreSv1StopMemoryProfiling(ctx *context.Context, args utils.TenantWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := args.APIOpts
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StopMemoryProfiling, args, reply)
}
