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

func (dS *DispatcherService) ReplicatorSv1Ping(args *utils.CGREvent, rpl *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1Ping, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaReplicator, utils.ReplicatorSv1Ping, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAccount(args *utils.StringWithAPIOpts, rpl *engine.Account) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetDestination(args *utils.StringWithAPIOpts, rpl *engine.Destination) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetReverseDestination(args *utils.StringWithAPIOpts, rpl *[]string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetReverseDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetStatQueue(args *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetStatQueue, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetFilter(args *utils.TenantIDWithAPIOpts, reply *engine.Filter) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetFilter, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetThreshold(args *utils.TenantIDWithAPIOpts, reply *engine.Threshold) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetThreshold, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetThresholdProfile(args *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetThresholdProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetStatQueueProfile(args *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetStatQueueProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetTiming(args *utils.StringWithAPIOpts, rpl *utils.TPTiming) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetResource(args *utils.TenantIDWithAPIOpts, reply *engine.Resource) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetResource, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetResourceProfile(args *utils.TenantIDWithAPIOpts, reply *engine.ResourceProfile) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetResourceProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetActionTriggers(args *utils.StringWithAPIOpts, rpl *engine.ActionTriggers) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetSharedGroup(args *utils.StringWithAPIOpts, rpl *engine.SharedGroup) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetActions(args *utils.StringWithAPIOpts, rpl *engine.Actions) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetActionPlan(args *utils.StringWithAPIOpts, rpl *engine.ActionPlan) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAllActionPlans(args *utils.StringWithAPIOpts, rpl *map[string]*engine.ActionPlan) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAllActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAccountActionPlans(args *utils.StringWithAPIOpts, rpl *[]string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRatingPlan(args *utils.StringWithAPIOpts, rpl *engine.RatingPlan) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRatingProfile(args *utils.StringWithAPIOpts, rpl *engine.RatingProfile) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRouteProfile(args *utils.TenantIDWithAPIOpts, reply *engine.RouteProfile) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRouteProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetAttributeProfile(args *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAttributeProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetChargerProfile(args *utils.TenantIDWithAPIOpts, reply *engine.ChargerProfile) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetChargerProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetDispatcherProfile(args *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetDispatcherProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetDispatcherHost(args *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) (err error) {
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
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetDispatcherHost, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetRateProfile(args *utils.TenantIDWithAPIOpts, reply *engine.RateProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRateProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetRateProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetActionProfile(args *utils.TenantIDWithAPIOpts, reply *engine.ActionProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetActionProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetActionProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetItemLoadIDs(args *utils.StringWithAPIOpts, rpl *map[string]int64) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetItemLoadIDs, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetThresholdProfile(args *engine.ThresholdProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ThresholdProfileWithAPIOpts{}
	}
	args.ThresholdProfile.Tenant = utils.FirstNonEmpty(args.ThresholdProfile.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetThresholdProfile, args.ThresholdProfile.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.ThresholdProfile.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetThresholdProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetThreshold(args *engine.ThresholdWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ThresholdWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetThreshold, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetReverseDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetStatQueue(args *engine.StatQueueWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.StatQueueWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetStatQueue, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetStatQueueProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetTiming(args *utils.TPTimingWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetResource(args *engine.ResourceWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ResourceWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetResource, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetDispatcherProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRateProfile(args *utils.RateProfileWithOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.RateProfileWithOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetRateProfile, args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetRateProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActionProfile(args *engine.ActionProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &engine.ActionProfileWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetActionProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetActionProfile, args, rpl)
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetDispatcherHost, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveThreshold(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveThreshold, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveThreshold, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDestination(args *utils.StringWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetLoadIDs, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveAccount(args *utils.StringWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveStatQueue(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveStatQueue, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveStatQueue, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveFilter(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveFilter, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveFilter, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveThresholdProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveThresholdProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveThresholdProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveStatQueueProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveStatQueueProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveStatQueueProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveTiming(args *utils.StringWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveResource(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveResource, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveResource, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveResourceProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveResourceProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveResourceProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActionTriggers(args *utils.StringWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveSharedGroup(args *utils.StringWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActions(args *utils.StringWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActionPlan(args *utils.StringWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRatingPlan(args *utils.StringWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRatingProfile(args *utils.StringWithAPIOpts, rpl *string) (err error) {
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
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRouteProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveRouteProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRouteProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveAttributeProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveAttributeProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveAttributeProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveChargerProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveChargerProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveChargerProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveDispatcherProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDispatcherProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherHost(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveDispatcherHost, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveDispatcherHost, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRateProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveRateProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveRateProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActionProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveActionProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveActionProfile, args, rpl)
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
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
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.Opts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveIndexes, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetAccountProfile(args *utils.TenantIDWithAPIOpts, reply *utils.AccountProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAccountProfile, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1GetAccountProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1SetAccountProfile(args *utils.AccountProfileWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.AccountProfileWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1SetAccountProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1SetAccountProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveAccountProfile(args *utils.TenantIDWithAPIOpts, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithAPIOpts{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1RemoveAccountProfile, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{
		Tenant: args.Tenant,
		Opts:   args.APIOpts,
	}, utils.MetaReplicator, utils.ReplicatorSv1RemoveAccountProfile, args, rpl)
}
