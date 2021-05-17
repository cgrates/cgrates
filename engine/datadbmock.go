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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type DataDBMock struct {
	GetKeysForPrefixF     func(*context.Context, string) ([]string, error)
	GetChargerProfileDrvF func(string, string) (*ChargerProfile, error)
	GetFilterDrvF         func(string, string) (*Filter, error)
	GetIndexesDrvF        func(idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error)
}

//Storage methods
func (dbM *DataDBMock) Close() {}

func (dbM *DataDBMock) Flush(string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetKeysForPrefix(ctx *context.Context, prf string) ([]string, error) {
	if dbM.GetKeysForPrefixF != nil {
		return dbM.GetKeysForPrefixF(ctx, prf)
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

//DataDB methods
func (dbM *DataDBMock) HasDataDrv(*context.Context, string, string, string) (bool, error) {
	return false, utils.ErrNotImplemented
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

func (dbM *DataDBMock) GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) AddLoadHistory(*utils.LoadInstance, int, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetIndexesDrv(ctx *context.Context, idxItmType, tntCtx string,
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

func (dbM *DataDBMock) GetFilterDrv(*context.Context, string, string) (*Filter, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetFilterDrv(*context.Context, *Filter) error {
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

func (dbM *DataDBMock) GetAttributeProfileDrv(*context.Context, string, string) (*AttributeProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetAttributeProfileDrv(*context.Context, *AttributeProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveAttributeProfileDrv(*context.Context, string, string) error {
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

func (dbM *DataDBMock) SetLoadIDsDrv(ctx *context.Context, loadIDs map[string]int64) error {
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

func (dbM *DataDBMock) GetRateProfileDrv(*context.Context, string, string) (*utils.RateProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetRateProfileDrv(*context.Context, *utils.RateProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRateProfileDrv(*context.Context, string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetActionProfileDrv(*context.Context, string, string) (*ActionProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetActionProfileDrv(*context.Context, *ActionProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveActionProfileDrv(*context.Context, string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAccountDrv(*context.Context, string, string) (*utils.Account, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetAccountDrv(ctx *context.Context, profile *utils.Account) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveAccountDrv(*context.Context, string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetVersions(vrs Versions, overwrite bool) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRatingProfileDrv(string) error {
	return utils.ErrNotImplemented
}
