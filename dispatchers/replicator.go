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

func (dS *DispatcherService) ReplicatorSv1GetAccount(args *StringWithApiKey, rpl *engine.Account) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetDestination(args *StringWithApiKey, rpl *engine.Destination) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetReverseDestination(args *StringWithApiKey, rpl *[]string) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetTiming(args *StringWithApiKey, rpl *utils.TPTiming) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetActionTriggers(args *StringWithApiKey, rpl *engine.ActionTriggers) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetShareGroup(args *StringWithApiKey, rpl *engine.SharedGroup) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
	}
	args.TenantArg.Tenant = utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.ReplicatorSv1GetShareGroup, args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaReplicator, routeID,
		utils.ReplicatorSv1GetShareGroup, args, rpl)
}

func (dS *DispatcherService) ReplicatorSv1GetActions(args *StringWithApiKey, rpl *engine.Actions) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetActionPlan(args *StringWithApiKey, rpl *engine.ActionPlan) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetAllActionPlans(args *StringWithApiKey, rpl *map[string]*engine.ActionPlan) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetAccountActionPlans(args *StringWithApiKey, rpl *[]string) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetRatingPlan(args *StringWithApiKey, rpl *engine.RatingPlan) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetRatingProfile(args *StringWithApiKey, rpl *engine.RatingProfile) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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

func (dS *DispatcherService) ReplicatorSv1GetSupplierProfile(args *utils.TenantIDWithArgDispatcher, reply *engine.SupplierProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.TenantID != nil && args.TenantID.Tenant != utils.EmptyString {
		tnt = args.TenantID.Tenant
	}
	if args.ArgDispatcher == nil {
		return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ReplicatorSv1GetSupplierProfile, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	routeID := args.ArgDispatcher.RouteID
	return dS.Dispatch(&utils.CGREvent{
		Tenant: tnt,
		ID:     args.ID,
	}, utils.MetaReplicator, routeID, utils.ReplicatorSv1GetSupplierProfile, args, reply)
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

func (dS *DispatcherService) ReplicatorSv1GetItemLoadIDs(args *StringWithApiKey, rpl *map[string]int64) (err error) {
	if args == nil {
		args = &StringWithApiKey{}
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
