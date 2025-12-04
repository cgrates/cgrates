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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ReplicatorSv1Ping(ctx *context.Context, args *utils.CGREvent, rpl *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1Ping, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaReplicator, utils.ReplicatorSv1Ping, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAccount(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *engine.Account) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAccount, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetDestination(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *engine.Destination) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetDestination, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetReverseDestination(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *[]string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetReverseDestination, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetReverseDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetStatQueue, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetStatQueue, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Filter) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetFilter, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetFilter, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Threshold) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetThreshold, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetThreshold, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetThresholdProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetThresholdProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetStatQueueProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetStatQueueProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetTiming(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *utils.TPTiming) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetTiming, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Resource) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetResource, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetResource, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ResourceProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetResourceProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetResourceProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetIPAllocations(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.IPAllocations) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetIPAllocations, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetIPAllocations, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetIPProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.IPProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetIPProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetIPProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetActionTriggers(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *engine.ActionTriggers) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetActionTriggers, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetSharedGroup(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *engine.SharedGroup) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetSharedGroup, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetActions(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *engine.Actions) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetActions, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetActionPlan(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *engine.ActionPlan) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetActionPlan, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAllActionPlans(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *map[string]*engine.ActionPlan) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAllActionPlans, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAllActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAccountActionPlans(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *[]string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAccountActionPlans, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRatingPlan(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *engine.RatingPlan) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRatingPlan, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRatingProfile(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *engine.RatingProfile) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRatingProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.RouteProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRouteProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRouteProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAttributeProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAttributeProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ChargerProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetChargerProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetChargerProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetDispatcherProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetDispatcherProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetDispatcherProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetDispatcherHost(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetDispatcherHost, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tnt,
		ID:      args.ID,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetDispatcherHost, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetItemLoadIDs(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *map[string]int64) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetItemLoadIDs, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetItemLoadIDs, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetThresholdProfile(ctx *context.Context, args *engine.ThresholdProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ThresholdProfileWithAPIOpts{
			ThresholdProfile: &engine.ThresholdProfile{},
		}
	}
	args.ThresholdProfile.Tenant = utils.FirstNonEmpty(args.ThresholdProfile.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetThresholdProfile, args.ThresholdProfile.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.ThresholdProfile.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetThresholdProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetThreshold(ctx *context.Context, args *engine.ThresholdWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ThresholdWithAPIOpts{
			Threshold: &engine.Threshold{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetThreshold, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetThreshold, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetDestination(ctx *context.Context, args *engine.DestinationWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.DestinationWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetDestination, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetAccount(ctx *context.Context, args *engine.AccountWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.AccountWithAPIOpts{
			Account: &engine.Account{},
		}
	}
	tenant := utils.FirstNonEmpty(utils.SplitConcatenatedKey(args.ID)[0], dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetAccount, tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetReverseDestination(ctx *context.Context, args *engine.DestinationWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.DestinationWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetReverseDestination, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetReverseDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetStatQueue(ctx *context.Context, args *engine.StatQueueWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.StatQueueWithAPIOpts{
			StatQueue: &engine.StatQueue{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetStatQueue, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetStatQueue, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetFilter(ctx *context.Context, args *engine.FilterWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.FilterWithAPIOpts{
			Filter: &engine.Filter{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetFilter, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetFilter, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetStatQueueProfile(ctx *context.Context, args *engine.StatQueueProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.StatQueueProfileWithAPIOpts{
			StatQueueProfile: &engine.StatQueueProfile{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetStatQueueProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetStatQueueProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetTiming(ctx *context.Context, args *utils.TPTimingWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TPTimingWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetTiming, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetResource(ctx *context.Context, args *engine.ResourceWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ResourceWithAPIOpts{
			Resource: &engine.Resource{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetResource, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetResource, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetResourceProfile(ctx *context.Context, args *engine.ResourceProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ResourceProfileWithAPIOpts{
			ResourceProfile: &engine.ResourceProfile{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetResourceProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetResourceProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetIPAllocations(ctx *context.Context, args *engine.IPAllocationsWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.IPAllocationsWithAPIOpts{
			IPAllocations: &engine.IPAllocations{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetIPAllocations, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetIPAllocations, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetIPProfile(ctx *context.Context, args *engine.IPProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.IPProfileWithAPIOpts{
			IPProfile: &engine.IPProfile{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetIPProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetIPProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActionTriggers(ctx *context.Context, args *engine.SetActionTriggersArgWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetActionTriggersArgWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetActionTriggers, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetSharedGroup(ctx *context.Context, args *engine.SharedGroupWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SharedGroupWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetSharedGroup, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActions(ctx *context.Context, args *engine.SetActionsArgsWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetActionsArgsWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetActions, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRatingPlan(ctx *context.Context, args *engine.RatingPlanWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.RatingPlanWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetRatingPlan, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRatingProfile(ctx *context.Context, args *engine.RatingProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.RatingProfileWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetRatingProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRouteProfile(ctx *context.Context, args *engine.RouteProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.RouteProfileWithAPIOpts{
			RouteProfile: &engine.RouteProfile{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetRouteProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetRouteProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetAttributeProfile(ctx *context.Context, args *engine.AttributeProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.AttributeProfileWithAPIOpts{
			AttributeProfile: &engine.AttributeProfile{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetAttributeProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetAttributeProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetChargerProfile(ctx *context.Context, args *engine.ChargerProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ChargerProfileWithAPIOpts{
			ChargerProfile: &engine.ChargerProfile{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetChargerProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetChargerProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetDispatcherProfile(ctx *context.Context, args *engine.DispatcherProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.DispatcherProfileWithAPIOpts{
			DispatcherProfile: &engine.DispatcherProfile{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetDispatcherProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetDispatcherProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActionPlan(ctx *context.Context, args *engine.SetActionPlanArgWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetActionPlanArgWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetActionPlan, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetAccountActionPlans(ctx *context.Context, args *engine.SetAccountActionPlansArgWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetAccountActionPlansArgWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetAccountActionPlans, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetDispatcherHost(ctx *context.Context, args *engine.DispatcherHostWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.DispatcherHostWithAPIOpts{
			DispatcherHost: &engine.DispatcherHost{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetDispatcherHost, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetDispatcherHost, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveThreshold, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveThreshold, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDestination(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveDestination, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetLoadIDs(ctx *context.Context, args *utils.LoadIDsWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.LoadIDsWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetLoadIDs, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetLoadIDs, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveAccount(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveAccount, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveStatQueue, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveStatQueue, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveFilter, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveFilter, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveThresholdProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveThresholdProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveStatQueueProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveStatQueueProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveTiming(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveTiming, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveResource, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveResource, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveResourceProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveResourceProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveIPAllocations(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveIPAllocations, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveIPAllocations, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveIPProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveIPProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveIPProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActionTriggers(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveActionTriggers, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveSharedGroup(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveSharedGroup, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActions(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveActions, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActionPlan(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveActionPlan, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemAccountActionPlans(ctx *context.Context, args *engine.RemAccountActionPlansArgsWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.RemAccountActionPlansArgsWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemAccountActionPlans, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRatingPlan(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveRatingPlan, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRatingProfile(ctx *context.Context, args *utils.StringWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithAPIOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveRatingProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveRouteProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRouteProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveAttributeProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveAttributeProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveChargerProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveChargerProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveDispatcherProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDispatcherProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherHost(ctx *context.Context, args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
		}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveDispatcherHost, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDispatcherHost, args, rpl)
}

// ReplicatorSv1GetIndexes .
func (dS *DispatcherService) ReplicatorSv1GetIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *map[string]utils.StringSet) (err error) {
	if args == nil {
		args = &utils.GetIndexesArg{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetIndexes, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetIndexes, args, reply)
}

// ReplicatorSv1SetIndexes .
func (dS *DispatcherService) ReplicatorSv1SetIndexes(ctx *context.Context, args *utils.SetIndexesArg, reply *string) (err error) {
	if args == nil {
		args = &utils.SetIndexesArg{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetIndexes, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetIndexes, args, reply)
}

// ReplicatorSv1RemoveIndexes .
func (dS *DispatcherService) ReplicatorSv1RemoveIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *string) (err error) {
	if args == nil {
		args = &utils.GetIndexesArg{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveIndexes, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant:  args.Tenant,
		APIOpts: args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveIndexes, args, reply)
}
