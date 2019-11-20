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

// Call implements rpcclient.RpcClientConnection interface for internal RPC
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
func (rplSv1 *ReplicatorSv1) GetStatQueue(tntID *utils.TenantID, reply *engine.StoredStatQueue) error {
	if rcv, err := rplSv1.dm.DataDB().GetStoredStatQueueDrv(tntID.Tenant, tntID.ID); err != nil {
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
	if rcv, err := rplSv1.dm.DataDB().GetActionPlanDrv(id, true, utils.NonTransactional); err != nil {
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
	if rcv, err := rplSv1.dm.DataDB().GetAccountActionPlansDrv(id, false, utils.NonTransactional); err != nil {
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

// SetThresholdProfile alters/creates a ThresholdProfile
func (rplSv1 *ReplicatorSv1) SetThreshold(th *engine.Threshold, reply *string) error {
	if err := rplSv1.dm.DataDB().SetThresholdDrv(th); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// SetThresholdProfile alters/creates a ThresholdProfile
func (rplSv1 *ReplicatorSv1) SetFilterIndexes(args *utils.SetFilterIndexesArg, reply *string) error {
	if err := rplSv1.dm.SetFilterIndexes(args.CacheID, args.ItemIDPrefix,
		args.Indexes, true, utils.NonTransactional); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rplSv1 *ReplicatorSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
