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

package engine

import (
	"github.com/cgrates/cgrates/utils"
)

type DataDBMock struct {
	GetKeysForPrefixF         func(string) ([]string, error)
	GetChargerProfileDrvF     func(string, string) (*ChargerProfile, error)
	GetFilterDrvF             func(string, string) (*Filter, error)
	GetIndexesDrvF            func(idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error)
	GetThresholdProfileDrvF   func(tenant, id string) (tp *ThresholdProfile, err error)
	SetThresholdProfileDrvF   func(tp *ThresholdProfile) (err error)
	RemThresholdProfileDrvF   func(tenant, id string) (err error)
	GetThresholdDrvF          func(tenant, id string) (*Threshold, error)
	GetResourceProfileDrvF    func(tnt, id string) (*ResourceProfile, error)
	SetResourceProfileDrvF    func(rp *ResourceProfile) error
	RemoveResourceProfileDrvF func(tnt, id string) error
	SetResourceDrvF           func(r *Resource) error
	GetStatQueueProfileDrvF   func(tenant, id string) (sq *StatQueueProfile, err error)
	SetStatQueueProfileDrvF   func(sq *StatQueueProfile) (err error)
	RemStatQueueProfileDrvF   func(tenant, id string) (err error)
	SetRankingProfileDrvF     func(sq *RankingProfile) (err error)
	GetRankingProfileDrvF     func(tenant string, id string) (sq *RankingProfile, err error)
	RemRankingProfileDrvF     func(tenant string, id string) (err error)
	SetTrendProfileDrvF       func(sq *TrendProfile) (err error)
	GetTrendProfileDrvF       func(tenant string, id string) (sq *TrendProfile, err error)
	RemTrendProfileDrvF       func(tenant string, id string) (err error)
	GetActionPlanDrvF         func(key string) (ap *ActionPlan, err error)
	SetActionPlanDrvF         func(key string, ap *ActionPlan) (err error)
	RemoveActionPlanDrvF      func(key string) (err error)
	GetRouteProfileDrvF       func(tenant, id string) (rp *RouteProfile, err error)
	RemoveRouteProfileDrvF    func(tenant, id string) error
	GetAccountDrvF            func(id string) (*Account, error)
	SetAccountDrvF            func(acc *Account) error
	RemoveAccountDrvF         func(id string) error
}

// Storage methods
func (dbM *DataDBMock) Close() {}

func (dbM *DataDBMock) Flush(string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetKeysForPrefix(prf string) ([]string, error) {
	if dbM.GetKeysForPrefixF != nil {
		return dbM.GetKeysForPrefixF(prf)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveKeysForPrefix(string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetVersions(itm string) (vrs Versions, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveVersions(vrs Versions) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) SelectDatabase(dbName string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetStorageType() string {
	return utils.EmptyString
}

func (dbM *DataDBMock) IsDBEmpty() (resp bool, err error) {
	return false, utils.ErrNotImplemented
}

// DataDB methods
func (dbM *DataDBMock) HasDataDrv(string, string, string) (bool, error) {
	return false, utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetRatingPlanDrv(string) (*RatingPlan, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetRatingPlanDrv(*RatingPlan) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRatingPlanDrv(key string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetRatingProfileDrv(string) (*RatingProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetRatingProfileDrv(*RatingProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetDestinationDrv(string, string) (*Destination, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetDestinationDrv(*Destination, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveDestinationDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveReverseDestinationDrv(string, string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetReverseDestinationDrv(string, []string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetReverseDestinationDrv(string, string) ([]string, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetActionsDrv(string) (Actions, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetActionsDrv(string, Actions) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveActionsDrv(string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetSharedGroupDrv(string) (*SharedGroup, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetSharedGroupDrv(*SharedGroup) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveSharedGroupDrv(id string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetActionTriggersDrv(string) (ActionTriggers, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetActionTriggersDrv(string, ActionTriggers) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveActionTriggersDrv(string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetActionPlanDrv(key string) (*ActionPlan, error) {
	if dbM.GetActionPlanDrvF != nil {
		return dbM.GetActionPlanDrvF(key)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetActionPlanDrv(key string, ap *ActionPlan) error {
	if dbM.GetActionPlanDrvF != nil {
		return dbM.SetActionPlanDrvF(key, ap)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveActionPlanDrv(key string) error {
	if dbM.RemoveActionPlanDrvF != nil {
		return dbM.RemoveActionPlanDrv(key)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAllActionPlansDrv() (map[string]*ActionPlan, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAccountActionPlansDrv(acntID string) (apIDs []string, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetAccountActionPlansDrv(acntID string, apIDs []string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemAccountActionPlansDrv(acntID string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) PushTask(*Task) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) PopTask() (*Task, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAccountDrv(id string) (*Account, error) {
	if dbM.GetAccountDrvF != nil {
		return dbM.GetAccountDrvF(id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetAccountDrv(acc *Account) error {
	if dbM.SetAccountDrvF != nil {
		return dbM.SetAccountDrvF(acc)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveAccountDrv(id string) error {
	if dbM.RemoveAccountDrvF != nil {
		return dbM.RemoveAccountDrvF(id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetResourceProfileDrv(tnt, id string) (*ResourceProfile, error) {
	if dbM.GetResourceProfileDrvF != nil {
		return dbM.GetResourceProfileDrvF(tnt, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetResourceProfileDrv(rp *ResourceProfile) error {
	if dbM.SetResourceProfileDrvF != nil {
		return dbM.SetResourceProfileDrvF(rp)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveResourceProfileDrv(tnt, id string) error {
	if dbM.RemoveResourceProfileDrvF != nil {
		return dbM.RemoveResourceProfileDrvF(tnt, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetResourceDrv(string, string) (*Resource, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetResourceDrv(r *Resource) error {
	if dbM.SetResourceDrvF != nil {
		return dbM.SetResourceDrvF(r)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveResourceDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetTimingDrv(string) (*utils.TPTiming, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetTimingDrv(*utils.TPTiming) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveTimingDrv(string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) AddLoadHistory(*utils.LoadInstance, int, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetIndexesDrv(idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
	if dbM.GetIndexesDrvF != nil {
		return dbM.GetIndexesDrvF(idxItmType, tntCtx, idxKey)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetIndexesDrv(idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetStatQueueProfileDrv(tenant, id string) (sq *StatQueueProfile, err error) {
	if dbM.GetStatQueueProfileDrvF != nil {
		return dbM.GetStatQueueProfileDrvF(tenant, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	if dbM.SetStatQueueProfileDrvF != nil {
		return dbM.SetStatQueueProfileDrvF(sq)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemStatQueueProfileDrv(tenant, id string) (err error) {
	if dbM.RemStatQueueProfileDrvF != nil {
		return dbM.RemStatQueueProfileDrvF(tenant, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetRankingProfileDrv(tenant, id string) (sg *RankingProfile, err error) {
	if dbM.GetStatQueueProfileDrvF != nil {
		return dbM.GetRankingProfileDrvF(tenant, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetRankingProfileDrv(sg *RankingProfile) (err error) {
	if dbM.SetRankingProfileDrvF(sg) != nil {
		return dbM.SetRankingProfileDrvF(sg)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemRankingProfileDrv(tenant string, id string) (err error) {
	if dbM.RemRankingProfileDrvF != nil {
		return dbM.RemRankingProfileDrvF(tenant, id)
	}
	return utils.ErrNotImplemented
}
func (dbM *DataDBMock) GetTrendProfileDrv(tenant, id string) (sg *TrendProfile, err error) {
	if dbM.GetStatQueueProfileDrvF != nil {
		return dbM.GetTrendProfileDrvF(tenant, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetTrendProfileDrv(trend *TrendProfile) (err error) {
	if dbM.SetTrendProfileDrvF(trend) != nil {
		return dbM.SetTrendProfileDrvF(trend)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemTrendProfileDrv(tenant string, id string) (err error) {
	if dbM.RemTrendProfileDrvF != nil {
		return dbM.RemTrendProfileDrvF(tenant, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetTrendDrv(tenant, id string) (*Trend, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetTrendDrv(*Trend) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveTrendDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetRankingDrv(tenant, id string) (*Ranking, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetRankingDrv(*Ranking) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRankingDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetStatQueueDrv(tenant, id string) (sq *StatQueue, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetStatQueueDrv(ssq *StoredStatQueue, sq *StatQueue) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemStatQueueDrv(tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetThresholdProfileDrv(tenant, id string) (tp *ThresholdProfile, err error) {
	if dbM.GetThresholdProfileDrvF != nil {
		return dbM.GetThresholdProfileDrvF(tenant, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	if dbM.SetThresholdProfileDrvF != nil {
		return dbM.SetThresholdProfileDrvF(tp)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemThresholdProfileDrv(tenant, id string) (err error) {
	if dbM.RemThresholdProfileDrvF != nil {
		return dbM.RemThresholdProfileDrvF(tenant, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetThresholdDrv(tenant, id string) (*Threshold, error) {
	if dbM.GetThresholdDrvF != nil {
		return dbM.GetThresholdDrvF(tenant, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetThresholdDrv(*Threshold) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveThresholdDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetFilterDrv(tnt string, id string) (*Filter, error) {
	if dbM.GetFilterDrvF != nil {
		return dbM.GetFilterDrvF(tnt, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetFilterDrv(*Filter) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveFilterDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetRouteProfileDrv(tenant, id string) (*RouteProfile, error) {
	if dbM.GetRouteProfileDrvF != nil {
		return dbM.GetRouteProfileDrvF(tenant, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetRouteProfileDrv(*RouteProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRouteProfileDrv(tenant, id string) error {

	if dbM.RemoveRouteProfileDrvF != nil {
		return dbM.RemoveRouteProfileDrvF(tenant, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAttributeProfileDrv(string, string) (*AttributeProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetAttributeProfileDrv(*AttributeProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveAttributeProfileDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetChargerProfileDrv(tnt string, id string) (*ChargerProfile, error) {
	if dbM.GetChargerProfileDrvF != nil {
		return dbM.GetChargerProfileDrvF(tnt, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetChargerProfileDrv(*ChargerProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveChargerProfileDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetDispatcherProfileDrv(string, string) (*DispatcherProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetDispatcherProfileDrv(*DispatcherProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveDispatcherProfileDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetLoadIDsDrv(loadIDs map[string]int64) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveLoadIDsDrv() error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetDispatcherHostDrv(string, string) (*DispatcherHost, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetDispatcherHostDrv(*DispatcherHost) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveDispatcherHostDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetVersions(vrs Versions, overwrite bool) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRatingProfileDrv(string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetBackupSessionsDrv(nodeID string, tnt string, storedSessions []*StoredSession) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetSessionsBackupDrv(nodeID string, tnt string) ([]*StoredSession, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveSessionsBackupDrv(nodeID, tnt, cgrid string) error {
	return utils.ErrNotImplemented
}
