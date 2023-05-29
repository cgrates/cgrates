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
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) EeSv1ArchiveEventsInReply(ctx *context.Context, args *ees.ArchiveEventsArgs, reply *[]uint8) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaEEs, utils.EeSv1ArchiveEventsInReply, args, reply)
}
func (dS *DispatcherService) EeSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
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
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaEEs, utils.EeSv1Ping, args, reply)
}
func (dS *DispatcherService) EeSv1ProcessEvent(ctx *context.Context, args *utils.CGREventWithEeIDs, reply *map[string]map[string]any) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.CGREvent != nil && len(args.CGREvent.Tenant) != 0) {
		tnt = args.CGREvent.Tenant
	}
	ev := make(map[string]any)
	if args != nil && args.CGREvent != nil {
		ev = args.CGREvent.Event
	}
	opts := make(map[string]any)
	if args != nil && args.CGREvent != nil {
		opts = args.CGREvent.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaEEs, utils.EeSv1ProcessEvent, args, reply)
}
