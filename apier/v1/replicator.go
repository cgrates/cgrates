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
func (rplSv1 *ReplicatorSv1) GetAccount(id string, reply *engine.Account) error {
	if rcv, err := rplSv1.dm.GetAccount(id); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetDestination
func (rplSv1 *ReplicatorSv1) GetDestination(key string, reply *engine.Destination) error {
	if rcv, err := rplSv1.dm.DataDB().GetDestinationDrv(key, true, utils.NonTransactional); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetDestination
func (rplSv1 *ReplicatorSv1) GetReverseDestination(key string, reply *[]string) error {
	if rcv, err := rplSv1.dm.DataDB().GetReverseDestinationDrv(key, true, utils.NonTransactional); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetStatQueue
func (rplSv1 *ReplicatorSv1) GetStatQueue(tntID *utils.TenantID, reply *engine.StatQueue) error {
	if rcv, err := rplSv1.dm.DataDB().GetStatQueueDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetFilter
func (rplSv1 *ReplicatorSv1) GetFilter(tntID *utils.TenantID, reply *engine.Filter) error {
	if rcv, err := rplSv1.dm.DataDB().GetFilterDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetThreshold
func (rplSv1 *ReplicatorSv1) GetThreshold(tntID *utils.TenantID, reply *engine.Threshold) error {
	if rcv, err := rplSv1.dm.DataDB().GetThresholdDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetThresholdProfile
func (rplSv1 *ReplicatorSv1) GetThresholdProfile(tntID *utils.TenantID, reply *engine.ThresholdProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetThresholdProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetStatQueueProfile
func (rplSv1 *ReplicatorSv1) GetStatQueueProfile(tntID *utils.TenantID, reply *engine.StatQueueProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetStatQueueProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetTiming
func (rplSv1 *ReplicatorSv1) GetTiming(id string, reply *utils.TPTiming) error {
	if rcv, err := rplSv1.dm.DataDB().GetTimingDrv(id); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResource
func (rplSv1 *ReplicatorSv1) GetResource(tntID *utils.TenantID, reply *engine.Resource) error {
	if rcv, err := rplSv1.dm.DataDB().GetResourceDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetResourceProfile(tntID *utils.TenantID, reply *engine.ResourceProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetResourceProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetActionTriggers
func (rplSv1 *ReplicatorSv1) GetActionTriggers(id string, reply engine.ActionTriggers) error {
	if rcv, err := rplSv1.dm.DataDB().GetActionTriggersDrv(id); err != nil {
		return err
	} else {
		reply = rcv
	}
	return nil
}

//GetShareGroup
func (rplSv1 *ReplicatorSv1) GetShareGroup(id string, reply *engine.SharedGroup) error {
	if rcv, err := rplSv1.dm.DataDB().GetSharedGroupDrv(id); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetActions
func (rplSv1 *ReplicatorSv1) GetActions(id string, reply *engine.Actions) error {
	if rcv, err := rplSv1.dm.DataDB().GetActionsDrv(id); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetActions
func (rplSv1 *ReplicatorSv1) GetActionPlan(id string, reply *engine.ActionPlan) error {
	if rcv, err := rplSv1.dm.DataDB().GetActionPlanDrv(id); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetAllActionPlans
func (rplSv1 *ReplicatorSv1) GetAllActionPlans(_ string, reply *map[string]*engine.ActionPlan) error {
	if rcv, err := rplSv1.dm.DataDB().GetAllActionPlansDrv(); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetAccountActionPlans
func (rplSv1 *ReplicatorSv1) GetAccountActionPlans(id string, reply *[]string) error {
	if rcv, err := rplSv1.dm.DataDB().GetAccountActionPlansDrv(id); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetAllActionPlans
func (rplSv1 *ReplicatorSv1) GetRatingPlan(id string, reply *engine.RatingPlan) error {
	if rcv, err := rplSv1.dm.DataDB().GetRatingPlanDrv(id); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetAllActionPlans
func (rplSv1 *ReplicatorSv1) GetRatingProfile(id string, reply *engine.RatingProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetRatingProfileDrv(id); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetSupplierProfile(tntID *utils.TenantID, reply *engine.SupplierProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetSupplierProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetAttributeProfile(tntID *utils.TenantID, reply *engine.AttributeProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetAttributeProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetChargerProfile(tntID *utils.TenantID, reply *engine.ChargerProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetChargerProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetDispatcherProfile(tntID *utils.TenantID, reply *engine.DispatcherProfile) error {
	if rcv, err := rplSv1.dm.DataDB().GetDispatcherProfileDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetDispatcherHost(tntID *utils.TenantID, reply *engine.DispatcherHost) error {
	if rcv, err := rplSv1.dm.DataDB().GetDispatcherHostDrv(tntID.Tenant, tntID.ID); err != nil {
		return err
	} else {
		*reply = *rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetItemLoadIDs(itemID string, reply *map[string]int64) error {
	if rcv, err := rplSv1.dm.DataDB().GetItemLoadIDsDrv(itemID); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) GetFilterIndexes(args *utils.GetFilterIndexesArg, reply *map[string]utils.StringMap) error {
	if rcv, err := rplSv1.dm.DataDB().GetFilterIndexesDrv(args.CacheID, args.ItemIDPrefix,
		args.FilterType, args.FldNameVal); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

//GetResourceProfile
func (rplSv1 *ReplicatorSv1) MatchFilterIndex(args *utils.MatchFilterIndexArg, reply *utils.StringMap) error {
	if rcv, err := rplSv1.dm.DataDB().MatchFilterIndexDrv(args.CacheID, args.ItemIDPrefix,
		args.FilterType, args.FieldName, args.FieldVal); err != nil {
		return err
	} else {
		*reply = rcv
	}
	return nil
}

// SetThresholdProfile alters/creates a ThresholdProfile
func (rplSv1 *ReplicatorSv1) SetThresholdProfile(th *engine.ThresholdProfile, reply *string) error {
	if err := rplSv1.dm.DataDB().SetThresholdProfileDrv(th); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetThreshold
func (rplSv1 *ReplicatorSv1) SetThreshold(th *engine.Threshold, reply *string) error {
	if err := rplSv1.dm.DataDB().SetThresholdDrv(th); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetFilterIndexes
func (rplSv1 *ReplicatorSv1) SetFilterIndexes(args *utils.SetFilterIndexesArg, reply *string) error {
	if err := rplSv1.dm.DataDB().SetFilterIndexesDrv(args.CacheID, args.ItemIDPrefix,
		args.Indexes, true, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetAccount
func (rplSv1 *ReplicatorSv1) SetAccount(acc *engine.Account, reply *string) error {
	if err := rplSv1.dm.DataDB().SetAccountDrv(acc); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetDestination
func (rplSv1 *ReplicatorSv1) SetDestination(dst *engine.Destination, reply *string) error {
	if err := rplSv1.dm.DataDB().SetDestinationDrv(dst, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetDestination
func (rplSv1 *ReplicatorSv1) SetReverseDestination(dst *engine.Destination, reply *string) error {
	if err := rplSv1.dm.DataDB().SetReverseDestinationDrv(dst, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetDestination
func (rplSv1 *ReplicatorSv1) SetStatQueue(ssq *engine.StoredStatQueue, reply *string) error {
	if err := rplSv1.dm.DataDB().SetStatQueueDrv(ssq, nil); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetDestination
func (rplSv1 *ReplicatorSv1) SetFilter(fltr *engine.Filter, reply *string) error {
	if err := rplSv1.dm.DataDB().SetFilterDrv(fltr); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetStatQueueProfile(sq *engine.StatQueueProfile, reply *string) error {
	if err := rplSv1.dm.DataDB().SetStatQueueProfileDrv(sq); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetTiming(tm *utils.TPTiming, reply *string) error {
	if err := rplSv1.dm.DataDB().SetTimingDrv(tm); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetResource(rs *engine.Resource, reply *string) error {
	if err := rplSv1.dm.DataDB().SetResourceDrv(rs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetResourceProfile(rs *engine.ResourceProfile, reply *string) error {
	if err := rplSv1.dm.DataDB().SetResourceProfileDrv(rs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetActionTriggers(args *engine.SetActionTriggersArg, reply *string) error {
	if err := rplSv1.dm.DataDB().SetActionTriggersDrv(args.Key, args.Attrs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetSharedGroup(shg *engine.SharedGroup, reply *string) error {
	if err := rplSv1.dm.DataDB().SetSharedGroupDrv(shg); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetActions(args *engine.SetActionsArgs, reply *string) error {
	if err := rplSv1.dm.DataDB().SetActionsDrv(args.Key, args.Acs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetRatingPlan(rp *engine.RatingPlan, reply *string) error {
	if err := rplSv1.dm.DataDB().SetRatingPlanDrv(rp); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetRatingProfile(rp *engine.RatingProfile, reply *string) error {
	if err := rplSv1.dm.DataDB().SetRatingProfileDrv(rp); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetSupplierProfile(sp *engine.SupplierProfile, reply *string) error {
	if err := rplSv1.dm.DataDB().SetSupplierProfileDrv(sp); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetAttributeProfile(ap *engine.AttributeProfile, reply *string) error {
	if err := rplSv1.dm.DataDB().SetAttributeProfileDrv(ap); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetChargerProfile(cp *engine.ChargerProfile, reply *string) error {
	if err := rplSv1.dm.DataDB().SetChargerProfileDrv(cp); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetDispatcherProfile(dpp *engine.DispatcherProfile, reply *string) error {
	if err := rplSv1.dm.DataDB().SetDispatcherProfileDrv(dpp); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetActionPlan(args *engine.SetActionPlanArg, reply *string) error {
	if err := rplSv1.dm.DataDB().SetActionPlanDrv(args.Key, args.Ats); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetAccountActionPlans(args *engine.SetAccountActionPlansArg, reply *string) error {
	if err := rplSv1.dm.DataDB().SetAccountActionPlansDrv(args.AcntID, args.AplIDs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetDispatcherHost(dpp *engine.DispatcherHost, reply *string) error {
	if err := rplSv1.dm.DataDB().SetDispatcherHostDrv(dpp); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveThreshold(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveThresholdDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) SetLoadIDs(loadIDs map[string]int64, reply *string) error {
	if err := rplSv1.dm.DataDB().SetLoadIDsDrv(loadIDs); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveDestination(id string, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveDestinationDrv(id, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveAccount(id string, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveAccountDrv(id); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveStatQueue(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemStatQueueDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveFilter(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveFilterDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveThresholdProfile(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemThresholdProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveStatQueueProfile(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemStatQueueProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveTiming(id string, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveTimingDrv(id); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveResource(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveResourceDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveResourceProfile(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveResourceProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveActionTriggers(id string, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveActionTriggersDrv(id); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveSharedGroup(id string, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveSharedGroupDrv(id); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveActions(id string, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveActionsDrv(id); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveActionPlan(id string, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveActionPlanDrv(id); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemAccountActionPlans(args *engine.RemAccountActionPlansArgs, reply *string) error {
	if err := rplSv1.dm.DataDB().RemAccountActionPlansDrv(args.AcntID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveRatingPlan(id string, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveRatingPlanDrv(id); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveRatingProfile(id string, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveRatingProfileDrv(id); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveSupplierProfile(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveSupplierProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveAttributeProfile(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveAttributeProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveChargerProfile(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveChargerProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveDispatcherProfile(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveDispatcherProfileDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) RemoveDispatcherHost(args *utils.TenantID, reply *string) error {
	if err := rplSv1.dm.DataDB().RemoveDispatcherHostDrv(args.Tenant, args.ID); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
