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

func (dS *DispatcherService) ReplicatorSv1Ping(args *utils.CGREventWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = utils.NewCGREventWithArgDispatcher()
	}
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1Ping, args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1Ping, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAccount(args *utils.StringWithApiKey, rpl *engine.Account) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetAccount, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetDestination(args *utils.StringWithApiKey, rpl *engine.Destination) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetDestination, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetReverseDestination(args *utils.StringWithApiKey, rpl *[]string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetReverseDestination, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetReverseDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetStatQueue(args *utils.TenantIDWithArgDispatcher, reply *engine.StatQueue) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetStatQueue, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetStatQueue, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetFilter(args *utils.TenantIDWithArgDispatcher, reply *engine.Filter) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetFilter, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetFilter, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetThreshold(args *utils.TenantIDWithArgDispatcher, reply *engine.Threshold) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetThreshold, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetThreshold, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetThresholdProfile(args *utils.TenantIDWithArgDispatcher, reply *engine.ThresholdProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetThresholdProfile, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetThresholdProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetStatQueueProfile(args *utils.TenantIDWithArgDispatcher, reply *engine.StatQueueProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetStatQueueProfile, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetStatQueueProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetTiming(args *utils.StringWithApiKey, rpl *utils.TPTiming) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetTiming, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetResource(args *utils.TenantIDWithArgDispatcher, reply *engine.Resource) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetResource, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetResource, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetResourceProfile(args *utils.TenantIDWithArgDispatcher, reply *engine.ResourceProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetResourceProfile, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetResourceProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetActionTriggers(args *utils.StringWithApiKey, rpl *engine.ActionTriggers) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetActionTriggers, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetSharedGroup(args *utils.StringWithApiKey, rpl *engine.SharedGroup) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetSharedGroup, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetActions(args *utils.StringWithApiKey, rpl *engine.Actions) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetActions, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetActionPlan(args *utils.StringWithApiKey, rpl *engine.ActionPlan) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetActionPlan, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAllActionPlans(args *utils.StringWithApiKey, rpl *map[string]*engine.ActionPlan) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetAllActionPlans, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetAllActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetAccountActionPlans(args *utils.StringWithApiKey, rpl *[]string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetAccountActionPlans, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRatingPlan(args *utils.StringWithApiKey, rpl *engine.RatingPlan) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetRatingPlan, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRatingProfile(args *utils.StringWithApiKey, rpl *engine.RatingProfile) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetRatingProfile, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetRouteProfile(args *utils.TenantIDWithArgDispatcher, reply *engine.RouteProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRouteProfile, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetRouteProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetAttributeProfile(args *utils.TenantIDWithArgDispatcher, reply *engine.AttributeProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetAttributeProfile, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetAttributeProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetChargerProfile(args *utils.TenantIDWithArgDispatcher, reply *engine.ChargerProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetChargerProfile, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetChargerProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetDispatcherProfile(args *utils.TenantIDWithArgDispatcher, reply *engine.DispatcherProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetDispatcherProfile, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetDispatcherProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetDispatcherHost(args *utils.TenantIDWithArgDispatcher, reply *engine.DispatcherHost) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetDispatcherHost, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetDispatcherHost, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetRateProfile(args *utils.TenantIDWithArgDispatcher, reply *engine.RateProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetRateProfile, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetRateProfile, args, reply)
}

func (dS *DispatcherService) ReplicatorSv1GetItemLoadIDs(args *utils.StringWithApiKey, rpl *map[string]int64) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetItemLoadIDs, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetItemLoadIDs, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetFilterIndexes(args *utils.GetFilterIndexesArgWithArgDispatcher, rpl *map[string]utils.StringMap) (err error) {
	if args == nil {
		args = &utils.GetFilterIndexesArgWithArgDispatcher{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetFilterIndexes, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetFilterIndexes, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1MatchFilterIndex(args *utils.MatchFilterIndexArgWithArgDispatcher, rpl *utils.StringMap) (err error) {
	if args == nil {
		args = &utils.MatchFilterIndexArgWithArgDispatcher{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1MatchFilterIndex, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1MatchFilterIndex, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetThresholdProfile(args *engine.ThresholdProfileWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.ThresholdProfileWithArgDispatcher{}
	}
	args.ThresholdProfile.Tenant = utils.FirstNonEmpty(args.ThresholdProfile.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetThresholdProfile, args.ThresholdProfile.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.ThresholdProfile.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetThresholdProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetThreshold(args *engine.ThresholdWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.ThresholdWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetThreshold, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetThreshold, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetFilterIndexes(args *utils.SetFilterIndexesArgWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.SetFilterIndexesArgWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetFilterIndexes, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetFilterIndexes, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetDestination(args *engine.DestinationWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.DestinationWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetDestination, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetAccount(args *engine.AccountWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.AccountWithArgDispatcher{}
	}
	tenant := utils.FirstNonEmpty(utils.SplitConcatenatedKey(args.ID)[0], dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetAccount, tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetReverseDestination(args *engine.DestinationWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.DestinationWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetReverseDestination, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetReverseDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetStatQueue(args *engine.StoredStatQueueWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.StoredStatQueueWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetStatQueue, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetStatQueue, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetFilter(args *engine.FilterWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.FilterWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetFilter, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetFilter, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetStatQueueProfile(args *engine.StatQueueProfileWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.StatQueueProfileWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetStatQueueProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetStatQueueProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetTiming(args *utils.TPTimingWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TPTimingWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetTiming, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetResource(args *engine.ResourceWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.ResourceWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetResource, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetResource, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetResourceProfile(args *engine.ResourceProfileWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.ResourceProfileWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetResourceProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetResourceProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActionTriggers(args *engine.SetActionTriggersArgWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetActionTriggersArgWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetActionTriggers, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetSharedGroup(args *engine.SharedGroupWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.SharedGroupWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetSharedGroup, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActions(args *engine.SetActionsArgsWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetActionsArgsWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetActions, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRatingPlan(args *engine.RatingPlanWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.RatingPlanWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetRatingPlan, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRatingProfile(args *engine.RatingProfileWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.RatingProfileWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetRatingProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRouteProfile(args *engine.RouteProfileWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.RouteProfileWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetRouteProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetRouteProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetAttributeProfile(args *engine.AttributeProfileWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.AttributeProfileWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetAttributeProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetAttributeProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetChargerProfile(args *engine.ChargerProfileWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.ChargerProfileWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetChargerProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetChargerProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetDispatcherProfile(args *engine.DispatcherProfileWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.DispatcherProfileWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetDispatcherProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetDispatcherProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetRateProfile(args *engine.RateProfileWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.RateProfileWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetRateProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetRateProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetActionPlan(args *engine.SetActionPlanArgWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetActionPlanArgWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetActionPlan, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetAccountActionPlans(args *engine.SetAccountActionPlansArgWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.SetAccountActionPlansArgWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetAccountActionPlans, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetDispatcherHost(args *engine.DispatcherHostWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.DispatcherHostWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetDispatcherHost, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetDispatcherHost, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveThreshold(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveThreshold, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveThreshold, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDestination(args *utils.StringWithApiKey, rpl *string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveDestination, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveDestination, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1SetLoadIDs(args *utils.LoadIDsWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.LoadIDsWithArgDispatcher{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1SetLoadIDs, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1SetLoadIDs, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveAccount(args *utils.StringWithApiKey, rpl *string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveAccount, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveAccount, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveStatQueue(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveStatQueue, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveStatQueue, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveFilter(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveFilter, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveFilter, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveThresholdProfile(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveThresholdProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveThresholdProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveStatQueueProfile(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveStatQueueProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveStatQueueProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveTiming(args *utils.StringWithApiKey, rpl *string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveTiming, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveTiming, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveResource(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveResource, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveResource, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveResourceProfile(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveResourceProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveResourceProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActionTriggers(args *utils.StringWithApiKey, rpl *string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveActionTriggers, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveActionTriggers, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveSharedGroup(args *utils.StringWithApiKey, rpl *string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveSharedGroup, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveSharedGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActions(args *utils.StringWithApiKey, rpl *string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveActions, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveActions, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveActionPlan(args *utils.StringWithApiKey, rpl *string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveActionPlan, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveActionPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemAccountActionPlans(args *engine.RemAccountActionPlansArgsWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &engine.RemAccountActionPlansArgsWithArgDispatcher{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemAccountActionPlans, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemAccountActionPlans, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRatingPlan(args *utils.StringWithApiKey, rpl *string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveRatingPlan, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveRatingPlan, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRatingProfile(args *utils.StringWithApiKey, rpl *string) (err error) {
	if args == nil {
		args = &utils.StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveRatingProfile, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveRatingProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRouteProfile(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveRouteProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveRouteProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveAttributeProfile(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveAttributeProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveAttributeProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveChargerProfile(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveChargerProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveChargerProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherProfile(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveDispatcherProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveDispatcherProfile, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveDispatcherHost(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveDispatcherHost, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveDispatcherHost, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1RemoveRateProfile(args *utils.TenantIDWithArgDispatcher, rpl *string) (err error) {
	if args == nil {
		args = &utils.TenantIDWithArgDispatcher{}
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1RemoveRateProfile, args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1RemoveRateProfile, args, rpl)
}
