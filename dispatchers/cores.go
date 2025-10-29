/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package dispatchers

import (
	"time"

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
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Panic, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Panic, args, reply)
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
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Ping, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Ping, args, reply)
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
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1Sleep, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1Sleep, args, reply)
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
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1StartCPUProfiling, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StartCPUProfiling, args, reply)
}
func (dS *DispatcherService) CoreSv1StartMemoryProfiling(ctx *context.Context, params cores.MemoryProfilingParams, reply *string) (err error) {
	if params.Tenant == utils.EmptyString {
		params.Tenant = dS.cfg.GeneralCfg().DefaultTenant
	}
	ev := make(map[string]any)
	if params.APIOpts == nil {
		params.APIOpts = make(map[string]any)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1StartMemoryProfiling, params.Tenant,
			utils.IfaceAsString(params.APIOpts[utils.OptsAPIKey]),
			utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(
		&utils.CGREvent{
			Tenant:  params.Tenant,
			Event:   ev,
			APIOpts: params.APIOpts,
		}, utils.MetaCore,
		utils.CoreSv1StartMemoryProfiling, params, reply,
	)
}
func (dS *DispatcherService) CoreSv1Status(ctx *context.Context, params *cores.V1StatusParams, reply *map[string]any) error {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if params != nil && params.Tenant != utils.EmptyString {
		tnt = params.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if params != nil && params.APIOpts != nil {
		opts = params.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err := dS.authorize(utils.CoreSv1Status, tnt,
			utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return err
		}
	}
	return dS.Dispatch(
		&utils.CGREvent{
			Tenant:  tnt,
			Event:   ev,
			APIOpts: opts,
		}, utils.MetaCore, utils.CoreSv1Status, params, reply)
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
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1StopCPUProfiling,
			tnt, utils.IfaceAsString(opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaCore, utils.CoreSv1StopCPUProfiling, args, reply)
}
func (dS *DispatcherService) CoreSv1StopMemoryProfiling(ctx *context.Context, params utils.TenantWithAPIOpts, reply *string) (err error) {
	if params.Tenant == utils.EmptyString {
		params.Tenant = dS.cfg.GeneralCfg().DefaultTenant
	}
	ev := make(map[string]any)
	if params.APIOpts == nil {
		params.APIOpts = make(map[string]any)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.CoreSv1StopMemoryProfiling,
			params.Tenant, utils.IfaceAsString(params.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(
		&utils.CGREvent{
			Tenant:  params.Tenant,
			Event:   ev,
			APIOpts: params.APIOpts,
		}, utils.MetaCore,
		utils.CoreSv1StopMemoryProfiling, params, reply,
	)
}
