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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ReplicatorSv1Ping(args *utils.CGREventWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.CGREventWithOpts)
	}
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1Ping, args.CGREvent.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaReplicator, utils.ReplicatorSv1Ping, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAccount(args *utils.StringWithOpts, rpl *engine.Account) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAccount, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetDestination(args *utils.StringWithOpts, rpl *engine.Destination) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetDestination, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetReverseDestination(args *utils.StringWithOpts, rpl *[]string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetReverseDestination, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetReverseDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetStatQueue(args *utils.TenantIDWithOpts, reply *engine.StatQueue) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetStatQueue, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetStatQueue, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetFilter(args *utils.TenantIDWithOpts, reply *engine.Filter) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetFilter, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetFilter, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetThreshold(args *utils.TenantIDWithOpts, reply *engine.Threshold) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetThreshold, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetThreshold, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetThresholdProfile(args *utils.TenantIDWithOpts, reply *engine.ThresholdProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetThresholdProfile, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetThresholdProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetStatQueueProfile(args *utils.TenantIDWithOpts, reply *engine.StatQueueProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetStatQueueProfile, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetStatQueueProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetTiming(args *utils.StringWithOpts, rpl *utils.TPTiming) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetTiming, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetResource(args *utils.TenantIDWithOpts, reply *engine.Resource) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetResource, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetResource, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetResourceProfile(args *utils.TenantIDWithOpts, reply *engine.ResourceProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetResourceProfile, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetResourceProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetActionTriggers(args *utils.StringWithOpts, rpl *engine.ActionTriggers) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetActionTriggers, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetSharedGroup(args *utils.StringWithOpts, rpl *engine.SharedGroup) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetSharedGroup, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetActions(args *utils.StringWithOpts, rpl *engine.Actions) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetActions, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetActionPlan(args *utils.StringWithOpts, rpl *engine.ActionPlan) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetActionPlan, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAllActionPlans(args *utils.StringWithOpts, rpl *map[string]*engine.ActionPlan) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAllActionPlans, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAllActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAccountActionPlans(args *utils.StringWithOpts, rpl *[]string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAccountActionPlans, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRatingPlan(args *utils.StringWithOpts, rpl *engine.RatingPlan) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRatingPlan, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRatingProfile(args *utils.StringWithOpts, rpl *engine.RatingProfile) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRatingProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRouteProfile(args *utils.TenantIDWithOpts, reply *engine.RouteProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRouteProfile, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRouteProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetAttributeProfile(args *utils.TenantIDWithOpts, reply *engine.AttributeProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAttributeProfile, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAttributeProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetChargerProfile(args *utils.TenantIDWithOpts, reply *engine.ChargerProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetChargerProfile, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetChargerProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetDispatcherProfile(args *utils.TenantIDWithOpts, reply *engine.DispatcherProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetDispatcherProfile, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetDispatcherProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetDispatcherHost(args *utils.TenantIDWithOpts, reply *engine.DispatcherHost) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetDispatcherHost, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetDispatcherHost, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetRateProfile(args *utils.TenantIDWithOpts, reply *engine.RateProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRateProfile, tnt,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tnt,
			ID:     args.ID,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRateProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetItemLoadIDs(args *utils.StringWithOpts, rpl *map[string]int64) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetItemLoadIDs, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetItemLoadIDs, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetThresholdProfile(args *engine.ThresholdProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ThresholdProfileWithOpts{}
	}
	args.ThresholdProfile.Tenant = utils.FirstNonEmpty(args.ThresholdProfile.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetThresholdProfile, args.ThresholdProfile.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.ThresholdProfile.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetThresholdProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetThreshold(args *engine.ThresholdWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ThresholdWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetThreshold, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetThreshold, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetDestination(args *engine.DestinationWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.DestinationWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetDestination, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetAccount(args *engine.AccountWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.AccountWithOpts{}
	}
	tenant := utils.FirstNonEmpty(utils.SplitConcatenatedKey(args.ID)[0], dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetAccount, tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetReverseDestination(args *engine.DestinationWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.DestinationWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetReverseDestination, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetReverseDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetStatQueue(args *engine.StoredStatQueueWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.StoredStatQueueWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetStatQueue, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetStatQueue, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetFilter(args *engine.FilterWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.FilterWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetFilter, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetFilter, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetStatQueueProfile(args *engine.StatQueueProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.StatQueueProfileWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetStatQueueProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetStatQueueProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetTiming(args *utils.TPTimingWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TPTimingWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetTiming, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetResource(args *engine.ResourceWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ResourceWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetResource, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetResource, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetResourceProfile(args *engine.ResourceProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ResourceProfileWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetResourceProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetResourceProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActionTriggers(args *engine.SetActionTriggersArgWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetActionTriggersArgWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetActionTriggers, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetSharedGroup(args *engine.SharedGroupWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SharedGroupWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetSharedGroup, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActions(args *engine.SetActionsArgsWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetActionsArgsWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetActions, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRatingPlan(args *engine.RatingPlanWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.RatingPlanWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetRatingPlan, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRatingProfile(args *engine.RatingProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.RatingProfileWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetRatingProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRouteProfile(args *engine.RouteProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.RouteProfileWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetRouteProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetRouteProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetAttributeProfile(args *engine.AttributeProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.AttributeProfileWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetAttributeProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetAttributeProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetChargerProfile(args *engine.ChargerProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ChargerProfileWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetChargerProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetChargerProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetDispatcherProfile(args *engine.DispatcherProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.DispatcherProfileWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetDispatcherProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetDispatcherProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRateProfile(args *engine.RateProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.RateProfileWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetRateProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetRateProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActionPlan(args *engine.SetActionPlanArgWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetActionPlanArgWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetActionPlan, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetAccountActionPlans(args *engine.SetAccountActionPlansArgWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetAccountActionPlansArgWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetAccountActionPlans, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetDispatcherHost(args *engine.DispatcherHostWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.DispatcherHostWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetDispatcherHost, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetDispatcherHost, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveThreshold(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveThreshold, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveThreshold, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDestination(args *utils.StringWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveDestination, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetLoadIDs(args *utils.LoadIDsWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.LoadIDsWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetLoadIDs, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetLoadIDs, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveAccount(args *utils.StringWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveAccount, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveStatQueue(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveStatQueue, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveStatQueue, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveFilter(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveFilter, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveFilter, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveThresholdProfile(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveThresholdProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveThresholdProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveStatQueueProfile(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveStatQueueProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveStatQueueProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveTiming(args *utils.StringWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveTiming, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveResource(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveResource, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveResource, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveResourceProfile(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveResourceProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveResourceProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActionTriggers(args *utils.StringWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveActionTriggers, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveSharedGroup(args *utils.StringWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveSharedGroup, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActions(args *utils.StringWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveActions, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActionPlan(args *utils.StringWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveActionPlan, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemAccountActionPlans(args *engine.RemAccountActionPlansArgsWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.RemAccountActionPlansArgsWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemAccountActionPlans, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRatingPlan(args *utils.StringWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveRatingPlan, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRatingProfile(args *utils.StringWithOpts, rpl *string) (err error) {
	if args == nil {
		args = new(utils.StringWithOpts)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveRatingProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRouteProfile(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveRouteProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRouteProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveAttributeProfile(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveAttributeProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveAttributeProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveChargerProfile(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveChargerProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveChargerProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherProfile(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveDispatcherProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDispatcherProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherHost(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveDispatcherHost, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDispatcherHost, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRateProfile(args *utils.TenantIDWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveRateProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRateProfile, args, rpl)
}

// ReplicatorSv1GetIndexes .
func (dS *DispatcherService) ReplicatorSv1GetIndexes(args *utils.GetIndexesArg, reply *map[string]utils.StringSet) (err error) {
	if args == nil {
		args = &utils.GetIndexesArg{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetIndexes, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetIndexes, args, reply)
}

// ReplicatorSv1SetIndexes .
func (dS *DispatcherService) ReplicatorSv1SetIndexes(args *utils.SetIndexesArg, reply *string) (err error) {
	if args == nil {
		args = &utils.SetIndexesArg{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetIndexes, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetIndexes, args, reply)
}

// ReplicatorSv1RemoveIndexes .
func (dS *DispatcherService) ReplicatorSv1RemoveIndexes(args *utils.GetIndexesArg, reply *string) (err error) {
	if args == nil {
		args = &utils.GetIndexesArg{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveIndexes, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: args.Tenant,
		},
		Opts: args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveIndexes, args, reply)
}
