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

type DataDBMock struct{}

//Storage methods
func (dbM *DataDBMock) Close() {}

func (dbM *DataDBMock) Flush(string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetKeysForPrefix(string) ([]string, error) {
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

//DataDB methods
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

func (dbM *DataDBMock) GetActionPlanDrv(string, bool, string) (*ActionPlan, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetActionPlanDrv(string, *ActionPlan, bool, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveActionPlanDrv(key string, transactionID string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAllActionPlansDrv() (map[string]*ActionPlan, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAccountActionPlansDrv(acntID string, skipCache bool,
	transactionID string) (apIDs []string, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetAccountActionPlansDrv(acntID string, apIDs []string, overwrite bool) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemAccountActionPlansDrv(acntID string, apIDs []string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) PushTask(*Task) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) PopTask() (*Task, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAccountDrv(string) (*Account, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetAccountDrv(*Account) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveAccountDrv(string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetResourceProfileDrv(string, string) (*ResourceProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetResourceProfileDrv(*ResourceProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveResourceProfileDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetResourceDrv(string, string) (*Resource, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetResourceDrv(*Resource) error {
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
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetIndexesDrv(idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetStatQueueProfileDrv(tenant string, ID string) (sq *StatQueueProfile, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemStatQueueProfileDrv(tenant, id string) (err error) {
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

func (dbM *DataDBMock) GetThresholdProfileDrv(tenant string, ID string) (tp *ThresholdProfile, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemThresholdProfileDrv(tenant, id string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetThresholdDrv(string, string) (*Threshold, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetThresholdDrv(*Threshold) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveThresholdDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetFilterDrv(string, string) (*Filter, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetFilterDrv(*Filter) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveFilterDrv(string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetRouteProfileDrv(string, string) (*RouteProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetRouteProfileDrv(*RouteProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRouteProfileDrv(string, string) error {
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

func (dbM *DataDBMock) GetChargerProfileDrv(string, string) (*ChargerProfile, error) {
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
