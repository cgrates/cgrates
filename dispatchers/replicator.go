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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ReplicatorSv1GetAccount(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.Account) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetAccount, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetActionProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ActionProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetActionProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetAttributeProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ChargerProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetChargerProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetDispatcherHost(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetDispatcherHost, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetDispatcherProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetDispatcherProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Filter) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetFilter, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *map[string]utils.StringSet) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetIndexes, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetItemLoadIDs(ctx *context.Context, args *utils.StringWithAPIOpts, reply *map[string]int64) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetItemLoadIDs, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetRateProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.RateProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetRateProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Resource) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetResource, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ResourceProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetResourceProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.RouteProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetRouteProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetStatQueue, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetStatQueueProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Threshold) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetThreshold, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetThresholdProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetTrend(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Trend) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetTrend, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1GetTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.TrendProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1GetTrendProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
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
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1Ping, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveAccount(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveAccount, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveActionProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActionProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveAttributeProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveChargerProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherHost(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDispatcherHost, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDispatcherProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveFilter, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveIndexes, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveRateProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRateProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveResource, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveResourceProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRouteProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveStatQueue, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveStatQueueProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveThreshold, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveThresholdProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveTrend(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveTrend, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1RemoveTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1RemoveTrendProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetAccount(ctx *context.Context, args *utils.AccountWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.Account != nil && len(args.Account.Tenant) != 0) {
		tnt = args.Account.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetAccount, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetActionProfile(ctx *context.Context, args *engine.ActionProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.ActionProfile != nil && len(args.ActionProfile.Tenant) != 0) {
		tnt = args.ActionProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetActionProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetAttributeProfile(ctx *context.Context, args *engine.AttributeProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.AttributeProfile != nil && len(args.AttributeProfile.Tenant) != 0) {
		tnt = args.AttributeProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetAttributeProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetChargerProfile(ctx *context.Context, args *engine.ChargerProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.ChargerProfile != nil && len(args.ChargerProfile.Tenant) != 0) {
		tnt = args.ChargerProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetChargerProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetDispatcherHost(ctx *context.Context, args *engine.DispatcherHostWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.DispatcherHost != nil && len(args.DispatcherHost.Tenant) != 0) {
		tnt = args.DispatcherHost.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetDispatcherHost, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetDispatcherProfile(ctx *context.Context, args *engine.DispatcherProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.DispatcherProfile != nil && len(args.DispatcherProfile.Tenant) != 0) {
		tnt = args.DispatcherProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetDispatcherProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetFilter(ctx *context.Context, args *engine.FilterWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.Filter != nil && len(args.Filter.Tenant) != 0) {
		tnt = args.Filter.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetFilter, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetIndexes(ctx *context.Context, args *utils.SetIndexesArg, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetIndexes, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetLoadIDs(ctx *context.Context, args *utils.LoadIDsWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetLoadIDs, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetRateProfile(ctx *context.Context, args *utils.RateProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.RateProfile != nil && len(args.RateProfile.Tenant) != 0) {
		tnt = args.RateProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetRateProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetResource(ctx *context.Context, args *engine.ResourceWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.Resource != nil && len(args.Resource.Tenant) != 0) {
		tnt = args.Resource.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetResource, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetResourceProfile(ctx *context.Context, args *engine.ResourceProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.ResourceProfile != nil && len(args.ResourceProfile.Tenant) != 0) {
		tnt = args.ResourceProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetResourceProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetRouteProfile(ctx *context.Context, args *engine.RouteProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.RouteProfile != nil && len(args.RouteProfile.Tenant) != 0) {
		tnt = args.RouteProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetRouteProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetStatQueue(ctx *context.Context, args *engine.StatQueueWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetStatQueue, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetStatQueueProfile(ctx *context.Context, args *engine.StatQueueProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.StatQueueProfile != nil && len(args.StatQueueProfile.Tenant) != 0) {
		tnt = args.StatQueueProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetStatQueueProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetThreshold(ctx *context.Context, args *engine.ThresholdWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.Threshold != nil && len(args.Threshold.Tenant) != 0) {
		tnt = args.Threshold.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetThreshold, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetThresholdProfile(ctx *context.Context, args *engine.ThresholdProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.ThresholdProfile != nil && len(args.ThresholdProfile.Tenant) != 0) {
		tnt = args.ThresholdProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetThresholdProfile, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetTrend(ctx *context.Context, args *engine.TrendWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.Trend != nil && len(args.Trend.Tenant) != 0) {
		tnt = args.Trend.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetTrend, args, reply)
}
func (dS *DispatcherService) ReplicatorSv1SetTrendProfile(ctx *context.Context, args *engine.TrendProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TrendProfile != nil && len(args.TrendProfile.Tenant) != 0) {
		tnt = args.TrendProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaReplicator, utils.ReplicatorSv1SetTrendProfile, args, reply)
}
