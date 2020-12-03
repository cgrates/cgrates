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

package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewReplicatorSv1(dm *engine.DataManager) *ReplicatorSv1 {
	return &ReplicatorSv1{dm: dm}
}

// Exports RPC
type ReplicatorSv1 struct {
	dm *engine.DataManager
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (rplSv1 *ReplicatorSv1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(rplSv1, serviceMethod, args, reply)
}

//GetAccount
func (rplSv1 *ReplicatorSv1) GetAccount(args *utils.StringWithOpts, reply *engine.Account) error {
	if rcv, err := rplSv1.dm.GetAccount(args.Arg); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetDestination
func (rplSv1 *ReplicatorSv1) GetDestination(key *utils.StringWithOpts, reply *engine.Destination) error {
	if rcv, err := rplSv1.dm.DataDB().GetDestinationDrv(key.Arg, utils.NonTransactional); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetDestination
func (rplSv1 *ReplicatorSv1) GetReverseDestination(key *utils.StringWithOpts, reply *[]string) error {
	if rcv, err := rplSv1.dm.DataDB().GetReverseDestinationDrv(key.Arg, true, utils.NonTransactional); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetStatQueue
func (rplSv1 *ReplicatorSv1) GetStatQueue(tntID *utils.TenantIDWithOpts, reply *engine.StatQueue) error {
	if rcv, err := rplSv1.dm.DataDB().GetStatQueueDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetFilter
func (rplSv1 *ReplicatorSv1) GetFilter(tntID *utils.TenantIDWithOpts, reply *engine.Filter) error {
	if rcv, err := rplSv1.dm.DataDB().GetFilterDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetThreshold
func (rplSv1 *ReplicatorSv1) GetThreshold(tntID *utils.TenantIDWithOpts, reply *engine.Threshold) error {
	if rcv, err := rplSv1.dm.DataDB().GetThresholdDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetThresholdProfile
func (rplSv1 *ReplicatorSv1) GetThresholdProfile(tntID *utils.TenantIDWithOpts, reply *engine.ThresholdProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetThresholdProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetStatQueueProfile
func (rplSv1 *ReplicatorSv1) GetStatQueueProfile(tntID *utils.TenantIDWithOpts, reply *engine.StatQueueProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetStatQueueProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetTiming
func (rplSv1 *ReplicatorSv1) GetTiming(id *utils.StringWithOpts, reply *utils.TPTiming) error {
	if rcv, err := rplSv1.dm.DataDB().GetTimingDrv(id.Arg); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResource
func (rplSv1 *ReplicatorSv1) GetResource(tntID *utils.TenantIDWithOpts, reply *engine.Resource) error {
	if rcv, err := rplSv1.dm.DataDB().GetResourceDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetResourceProfile(tntID *utils.TenantIDWithOpts, reply *engine.ResourceProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetResourceProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetActionTriggers
func (rplSv1 *ReplicatorSv1) GetActionTriggers(id *utils.StringWithOpts, reply *engine.ActionTriggers) error {
	if rcv, err := rplSv1.dm.DataDB().GetActionTriggersDrv(id.Arg); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetSharedGroup
func (rplSv1 *ReplicatorSv1) GetSharedGroup(id *utils.StringWithOpts, reply *engine.SharedGroup) error {
	if rcv, err := rplSv1.dm.DataDB().GetSharedGroupDrv(id.Arg); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetActions
func (rplSv1 *ReplicatorSv1) GetActions(id *utils.StringWithOpts, reply *engine.Actions) error {
	if rcv, err := rplSv1.dm.DataDB().GetActionsDrv(id.Arg); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetActions
func (rplSv1 *ReplicatorSv1) GetActionPlan(id *utils.StringWithOpts, reply *engine.ActionPlan) error {
	if rcv, err := rplSv1.dm.DataDB().GetActionPlanDrv(id.Arg, true, utils.NonTransactional); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetAllActionPlans
func (rplSv1 *ReplicatorSv1) GetAllActionPlans(_ *utils.StringWithOpts, reply *map[string]*engine.ActionPlan) error {
	if rcv, err := rplSv1.dm.DataDB().GetAllActionPlansDrv(); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetAccountActionPlans
func (rplSv1 *ReplicatorSv1) GetAccountActionPlans(id *utils.StringWithOpts, reply *[]string) error {
	if rcv, err := rplSv1.dm.DataDB().GetAccountActionPlansDrv(id.Arg, false, utils.NonTransactional); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetAllActionPlans
func (rplSv1 *ReplicatorSv1) GetRatingPlan(id *utils.StringWithOpts, reply *engine.RatingPlan) error {
	if rcv, err := rplSv1.dm.DataDB().GetRatingPlanDrv(id.Arg); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetAllActionPlans
func (rplSv1 *ReplicatorSv1) GetRatingProfile(id *utils.StringWithOpts, reply *engine.RatingProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetRatingProfileDrv(id.Arg); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetRouteProfile(tntID *utils.TenantIDWithOpts, reply *engine.RouteProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetRouteProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetAttributeProfile(tntID *utils.TenantIDWithOpts, reply *engine.AttributeProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetAttributeProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetChargerProfile(tntID *utils.TenantIDWithOpts, reply *engine.ChargerProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetChargerProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetDispatcherProfile(tntID *utils.TenantIDWithOpts, reply *engine.DispatcherProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetDispatcherProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetDispatcherHost(tntID *utils.TenantIDWithOpts, reply *engine.DispatcherHost) error {
	if rcv, err := rplSv1.dm.DataDB().GetDispatcherHostDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

func (rplSv1 *ReplicatorSv1) GetRateProfile(tntID *utils.TenantIDWithOpts, reply *engine.RateProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetRateProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetItemLoadIDs(itemID *utils.StringWithOpts, reply *map[string]int64) error {
	if rcv, err := rplSv1.dm.DataDB().GetItemLoadIDsDrv(itemID.Arg); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

// SetThresholdProfile alters/creates a ThresholdProfile
func (rplSv1 *ReplicatorSv1) SetThresholdProfile(th *engine.ThresholdProfileWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetThresholdProfileDrv(th.ThresholdProfile); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetThreshold
func (rplSv1 *ReplicatorSv1) SetThreshold(th *engine.ThresholdWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetThresholdDrv(th.Threshold); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetAccount
func (rplSv1 *ReplicatorSv1) SetAccount(acc *engine.AccountWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetAccountDrv(acc.Account); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetDestination
func (rplSv1 *ReplicatorSv1) SetDestination(dst *engine.DestinationWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetDestinationDrv(dst.Destination, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetReverseDestination
func (rplSv1 *ReplicatorSv1) SetReverseDestination(dst *engine.DestinationWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetReverseDestinationDrv(dst.Destination, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetStatQueue
func (rplSv1 *ReplicatorSv1) SetStatQueue(ssq *engine.StoredStatQueueWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetStatQueueDrv(ssq.StoredStatQueue, nil); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetFilter
func (rplSv1 *ReplicatorSv1) SetFilter(fltr *engine.FilterWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetFilterDrv(fltr.Filter); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetStatQueueProfile
func (rplSv1 *ReplicatorSv1) SetStatQueueProfile(sq *engine.StatQueueProfileWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetStatQueueProfileDrv(sq.StatQueueProfile); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetTiming
func (rplSv1 *ReplicatorSv1) SetTiming(tm *utils.TPTimingWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetTimingDrv(tm.TPTiming); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetResource
func (rplSv1 *ReplicatorSv1) SetResource(rs *engine.ResourceWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetResourceDrv(rs.Resource); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetResourceProfile
func (rplSv1 *ReplicatorSv1) SetResourceProfile(rs *engine.ResourceProfileWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetResourceProfileDrv(rs.ResourceProfile); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetActionTriggers
func (rplSv1 *ReplicatorSv1) SetActionTriggers(args *engine.SetActionTriggersArgWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetActionTriggersDrv(args.Key, args.Attrs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetSharedGroup
func (rplSv1 *ReplicatorSv1) SetSharedGroup(shg *engine.SharedGroupWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetSharedGroupDrv(shg.SharedGroup); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetActions
func (rplSv1 *ReplicatorSv1) SetActions(args *engine.SetActionsArgsWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetActionsDrv(args.Key, args.Acs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetRatingPlan
func (rplSv1 *ReplicatorSv1) SetRatingPlan(rp *engine.RatingPlanWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetRatingPlanDrv(rp.RatingPlan); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetRatingProfile
func (rplSv1 *ReplicatorSv1) SetRatingProfile(rp *engine.RatingProfileWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetRatingProfileDrv(rp.RatingProfile); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetRouteProfile
func (rplSv1 *ReplicatorSv1) SetRouteProfile(sp *engine.RouteProfileWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetRouteProfileDrv(sp.RouteProfile); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetAttributeProfile
func (rplSv1 *ReplicatorSv1) SetAttributeProfile(ap *engine.AttributeProfileWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetAttributeProfileDrv(ap.AttributeProfile); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetChargerProfile
func (rplSv1 *ReplicatorSv1) SetChargerProfile(cp *engine.ChargerProfileWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetChargerProfileDrv(cp.ChargerProfile); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetDispatcherProfile
func (rplSv1 *ReplicatorSv1) SetDispatcherProfile(dpp *engine.DispatcherProfileWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetDispatcherProfileDrv(dpp.DispatcherProfile); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetActionPlan
func (rplSv1 *ReplicatorSv1) SetActionPlan(args *engine.SetActionPlanArgWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetActionPlanDrv(args.Key, args.Ats, args.Overwrite, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetAccountActionPlans
func (rplSv1 *ReplicatorSv1) SetAccountActionPlans(args *engine.SetAccountActionPlansArgWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetAccountActionPlansDrv(args.AcntID, args.AplIDs, args.Overwrite); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetDispatcherHost
func (rplSv1 *ReplicatorSv1) SetDispatcherHost(dpp *engine.DispatcherHostWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetDispatcherHostDrv(dpp.DispatcherHost); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetRateProfile(dpp *engine.RateProfileWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetRateProfileDrv(dpp.RateProfile); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveThreshold
func (rplSv1 *ReplicatorSv1) RemoveThreshold(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveThresholdDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetLoadIDs
func (rplSv1 *ReplicatorSv1) SetLoadIDs(args *utils.LoadIDsWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().SetLoadIDsDrv(args.LoadIDs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveDestination
func (rplSv1 *ReplicatorSv1) RemoveDestination(id *utils.StringWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveDestinationDrv(id.Arg, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveAccount
func (rplSv1 *ReplicatorSv1) RemoveAccount(id *utils.StringWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveAccountDrv(id.Arg); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveStatQueue
func (rplSv1 *ReplicatorSv1) RemoveStatQueue(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemStatQueueDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveFilter
func (rplSv1 *ReplicatorSv1) RemoveFilter(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveFilterDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveThresholdProfile
func (rplSv1 *ReplicatorSv1) RemoveThresholdProfile(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemThresholdProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveStatQueueProfile
func (rplSv1 *ReplicatorSv1) RemoveStatQueueProfile(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemStatQueueProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveTiming
func (rplSv1 *ReplicatorSv1) RemoveTiming(id *utils.StringWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveTimingDrv(id.Arg); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveResource
func (rplSv1 *ReplicatorSv1) RemoveResource(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveResourceDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveResourceProfile
func (rplSv1 *ReplicatorSv1) RemoveResourceProfile(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveResourceProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveActionTriggers(id *utils.StringWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveActionTriggersDrv(id.Arg); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveSharedGroup(id *utils.StringWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveSharedGroupDrv(id.Arg); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveActions(id *utils.StringWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveActionsDrv(id.Arg); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveActionPlan(id *utils.StringWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveActionPlanDrv(id.Arg, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemAccountActionPlans(args *engine.RemAccountActionPlansArgsWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemAccountActionPlansDrv(args.AcntID, args.ApIDs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveRatingPlan(id *utils.StringWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveRatingPlanDrv(id.Arg); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveRatingProfile(id *utils.StringWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveRatingProfileDrv(id.Arg); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveRouteProfile(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveRouteProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveAttributeProfile(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveAttributeProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveChargerProfile(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveChargerProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveDispatcherProfile(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveDispatcherProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveRateProfile(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveRateProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveDispatcherHost(args *utils.TenantIDWithOpts, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveDispatcherHostDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) Ping(ign *utils.CGREventWithOpts, reply *string) error {
	*reply = utils.Pong
	return nil
}

// GetIndexes .
func (rplSv1 *ReplicatorSv1) GetIndexes(args *utils.GetIndexesArg, reply *map[string]utils.StringSet) error {
	indx, err := rplSv1.dm.DataDB().GetIndexesDrv(args.IdxItmType, args.TntCtx, args.IdxKey)
	if err != nil {
		return err
	}
	*reply = indx
	return nil
}

// SetIndexes .
func (rplSv1 *ReplicatorSv1) SetIndexes(args *utils.SetIndexesArg, reply *string) error {
	if err := rplSv1.dm.DataDB().SetIndexesDrv(args.IdxItmType, args.TntCtx, args.Indexes, true, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// RemoveIndexes .
func (rplSv1 *ReplicatorSv1) RemoveIndexes(args *utils.GetIndexesArg, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveIndexesDrv(args.IdxItmType, args.TntCtx, args.IdxKey); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}
