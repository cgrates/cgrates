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
	RemoveRateProfileDrvF      func(ctx *context.Context, str1 string, str2 string) error
	SetRateProfileDrvF         func(*context.Context, *utils.RateProfile) error
	GetRateProfileDrvF         func(*context.Context, string, string) (*utils.RateProfile, error)
	GetKeysForPrefixF          func(*context.Context, string) ([]string, error)
	GetIndexesDrvF             func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error)
	SetIndexesDrvF             func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error)
	GetAttributeProfileDrvF    func(ctx *context.Context, str1 string, str2 string) (*AttributeProfile, error)
	SetAttributeProfileDrvF    func(ctx *context.Context, attr *AttributeProfile) error
	RemoveAttributeProfileDrvF func(ctx *context.Context, str1 string, str2 string) error
	SetLoadIDsDrvF             func(ctx *context.Context, loadIDs map[string]int64) error
	GetFilterDrvF              func(ctx *context.Context, str1 string, str2 string) (*Filter, error)
	GetChargerProfileDrvF      func(ctx *context.Context, tnt, id string) (*ChargerProfile, error)
	SetChargerProfileDrvF      func(ctx *context.Context, chr *ChargerProfile) (err error)
	GetThresholdProfileDrvF    func(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error)
	SetThresholdProfileDrvF    func(ctx *context.Context, tp *ThresholdProfile) (err error)
	RemThresholdProfileDrvF    func(ctx *context.Context, tenant, id string) (err error)
	GetThresholdDrvF           func(ctx *context.Context, tenant, id string) (*Threshold, error)
	RemoveThresholdDrvF        func(ctx *context.Context, tnt, id string) error
	GetResourceProfileDrvF     func(ctx *context.Context, tnt, id string) (*ResourceProfile, error)
	SetResourceProfileDrvF     func(ctx *context.Context, rp *ResourceProfile) error
	RemoveResourceProfileDrvF  func(ctx *context.Context, tnt, id string) error
	RemoveResourceDrvF         func(ctx *context.Context, tnt, id string) error
	GetStatQueueProfileDrvF    func(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error)
	SetStatQueueProfileDrvF    func(ctx *context.Context, sq *StatQueueProfile) (err error)
	RemStatQueueProfileDrvF    func(ctx *context.Context, tenant, id string) (err error)
	RemStatQueueDrvF           func(ctx *context.Context, tenant, id string) (err error)
	SetFilterDrvF              func(ctx *context.Context, fltr *Filter) error
	GetActionProfileDrvF       func(ctx *context.Context, tenant string, ID string) (*ActionProfile, error)
	SetActionProfileDrvF       func(ctx *context.Context, ap *ActionProfile) error
	RemoveActionProfileDrvF    func(ctx *context.Context, tenant string, ID string) error
	RemoveFilterDrvF           func(str1 string, str2 string) error
	SetAccountDrvF             func(ctx *context.Context, profile *utils.Account) error
	GetAccountDrvF             func(ctx *context.Context, str1 string, str2 string) (*utils.Account, error)
	RemoveAccountDrvF          func(ctx *context.Context, str1 string, str2 string) error
	GetRouteProfileDrvF        func(ctx *context.Context, tnt, id string) (*RouteProfile, error)
	SetRouteProfileDrvF        func(ctx *context.Context, rtPrf *RouteProfile) error
	RemoveRouteProfileDrvF     func(ctx *context.Context, tnt, id string) error
	RemoveChargerProfileDrvF   func(ctx *context.Context, chr string, rpl string) error
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

func (dbM *DataDBMock) GetResourceProfileDrv(ctx *context.Context, tnt, id string) (*ResourceProfile, error) {
	if dbM.GetResourceProfileDrvF != nil {
		return dbM.GetResourceProfileDrvF(ctx, tnt, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetResourceProfileDrv(ctx *context.Context, resPrf *ResourceProfile) error {
	if dbM.SetResourceProfileDrvF != nil {
		return dbM.SetResourceProfileDrvF(ctx, resPrf)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveResourceProfileDrv(ctx *context.Context, tnt string, id string) error {
	if dbM.RemoveResourceProfileDrvF != nil {
		return dbM.RemoveResourceProfileDrvF(ctx, tnt, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetResourceDrv(*context.Context, string, string) (*Resource, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetResourceDrv(*context.Context, *Resource) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveResourceDrv(ctx *context.Context, tnt, id string) error {
	if dbM.RemoveResourceDrvF != nil {
		return dbM.RemoveResourceDrvF(ctx, tnt, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) AddLoadHistory(*utils.LoadInstance, int, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetIndexesDrv(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
	if dbM.GetIndexesDrvF != nil {
		return dbM.GetIndexesDrvF(ctx, idxItmType, tntCtx, idxKey)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetIndexesDrv(ctx *context.Context, idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	if dbM.SetIndexesDrvF != nil {
		return dbM.SetIndexesDrvF(ctx, idxItmType, tntCtx, indexes, commit, transactionID)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetStatQueueProfileDrv(ctx *context.Context, tenant, id string) (sq *StatQueueProfile, err error) {
	if dbM.GetStatQueueProfileDrvF != nil {
		return dbM.GetStatQueueProfileDrvF(ctx, tenant, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetStatQueueProfileDrv(ctx *context.Context, sq *StatQueueProfile) (err error) {
	if dbM.SetStatQueueProfileDrvF != nil {
		return dbM.SetStatQueueProfileDrvF(ctx, sq)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemStatQueueProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	if dbM.RemStatQueueProfileDrvF != nil {
		return dbM.RemStatQueueProfileDrvF(ctx, tenant, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetStatQueueDrv(ctx *context.Context, tenant, id string) (sq *StatQueue, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetStatQueueDrv(ctx *context.Context, ssq *StoredStatQueue, sq *StatQueue) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemStatQueueDrv(ctx *context.Context, tenant, id string) (err error) {
	if dbM.RemStatQueueDrvF != nil {
		return dbM.RemStatQueueDrvF(ctx, tenant, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetThresholdProfileDrv(ctx *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
	if dbM.GetThresholdProfileDrvF != nil {
		return dbM.GetThresholdProfileDrvF(ctx, tenant, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetThresholdProfileDrv(ctx *context.Context, tp *ThresholdProfile) (err error) {
	if dbM.SetThresholdProfileDrvF != nil {
		return dbM.SetThresholdProfileDrvF(ctx, tp)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemThresholdProfileDrv(ctx *context.Context, tenant, id string) (err error) {
	if dbM.RemThresholdProfileDrvF != nil {
		return dbM.RemThresholdProfileDrvF(ctx, tenant, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetThresholdDrv(ctx *context.Context, tenant, id string) (*Threshold, error) {
	if dbM.GetThresholdDrvF != nil {
		return dbM.GetThresholdDrvF(ctx, tenant, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetThresholdDrv(*context.Context, *Threshold) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveThresholdDrv(ctx *context.Context, tnt, id string) error {
	if dbM.RemoveThresholdDrvF != nil {
		return dbM.RemoveThresholdDrvF(ctx, tnt, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetFilterDrv(ctx *context.Context, str1 string, str2 string) (*Filter, error) {
	if dbM.GetFilterDrvF != nil {
		return dbM.GetFilterDrvF(ctx, str1, str2)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetFilterDrv(ctx *context.Context, fltr *Filter) error {
	if dbM.SetFilterDrvF != nil {
		return dbM.SetFilterDrvF(ctx, fltr)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveFilterDrv(str1 string, str2 string) error {
	if dbM.RemoveFilterDrvF != nil {
		return dbM.RemoveFilterDrvF(str1, str2)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetRouteProfileDrv(ctx *context.Context, tnt, id string) (*RouteProfile, error) {
	if dbM.GetRouteProfileDrvF != nil {
		return dbM.GetRouteProfileDrvF(ctx, tnt, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetRouteProfileDrv(ctx *context.Context, rtPrf *RouteProfile) error {
	if dbM.SetRouteProfileDrvF != nil {
		return dbM.SetRouteProfileDrvF(ctx, rtPrf)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRouteProfileDrv(ctx *context.Context, tnt, id string) error {
	if dbM.RemoveRouteProfileDrvF != nil {
		return dbM.RemoveRouteProfileDrvF(ctx, tnt, id)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAttributeProfileDrv(ctx *context.Context, str1 string, str2 string) (*AttributeProfile, error) {
	if dbM.GetAttributeProfileDrvF != nil {
		return dbM.GetAttributeProfileDrvF(ctx, str1, str2)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetAttributeProfileDrv(ctx *context.Context, attr *AttributeProfile) error {
	if dbM.SetAttributeProfileDrvF != nil {
		return dbM.SetAttributeProfileDrvF(ctx, attr)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveAttributeProfileDrv(ctx *context.Context, str1 string, str2 string) error {
	if dbM.RemoveAttributeProfileDrvF != nil {
		return dbM.RemoveAttributeProfileDrvF(ctx, str1, str2)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetChargerProfileDrv(ctx *context.Context, tnt, id string) (*ChargerProfile, error) {
	if dbM.GetChargerProfileDrvF != nil {
		return dbM.GetChargerProfileDrvF(ctx, tnt, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetChargerProfileDrv(ctx *context.Context, chrg *ChargerProfile) error {
	if dbM.SetChargerProfileDrvF != nil {
		return dbM.SetChargerProfileDrvF(ctx, chrg)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveChargerProfileDrv(ctx *context.Context, chr string, rpl string) error {
	if dbM.RemoveChargerProfileDrvF != nil {
		return dbM.RemoveChargerProfileDrvF(ctx, chr, rpl)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetDispatcherProfileDrv(*context.Context, string, string) (*DispatcherProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetDispatcherProfileDrv(*context.Context, *DispatcherProfile) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveDispatcherProfileDrv(*context.Context, string, string) error {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetLoadIDsDrv(ctx *context.Context, loadIDs map[string]int64) error {
	if dbM.SetLoadIDsDrvF != nil {
		return dbM.SetLoadIDsDrvF(ctx, loadIDs)
	}
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

func (dbM *DataDBMock) GetRateProfileDrv(ctx *context.Context, tnt string, id string) (*utils.RateProfile, error) {
	if dbM.GetRateProfileDrvF != nil {
		return dbM.GetRateProfileDrvF(ctx, tnt, id)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetRateProfileDrv(ctx *context.Context, rt *utils.RateProfile) error {
	if dbM.SetRateProfileDrvF != nil {
		return dbM.SetRateProfileDrvF(ctx, rt)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRateProfileDrv(ctx *context.Context, str1 string, str2 string) error {
	if dbM.RemoveRateProfileDrvF != nil {
		return dbM.RemoveRateProfileDrvF(ctx, str1, str2)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetActionProfileDrv(ctx *context.Context, tenant string, ID string) (*ActionProfile, error) {
	if dbM.GetActionProfileDrvF != nil {
		return dbM.GetActionProfileDrvF(ctx, tenant, ID)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetActionProfileDrv(ctx *context.Context, ap *ActionProfile) error {
	if dbM.SetActionProfileDrvF != nil {
		return dbM.SetActionProfileDrvF(ctx, ap)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveActionProfileDrv(ctx *context.Context, tenant string, ID string) error {
	if dbM.RemoveActionProfileDrvF != nil {
		return dbM.RemoveActionProfileDrvF(ctx, tenant, ID)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) GetAccountDrv(ctx *context.Context, str1 string, str2 string) (*utils.Account, error) {
	if dbM.GetAccountDrvF != nil {
		return dbM.GetAccountDrvF(ctx, str1, str2)
	}
	return nil, utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetAccountDrv(ctx *context.Context, profile *utils.Account) error {
	if dbM.SetAccountDrvF != nil {
		return dbM.SetAccountDrvF(ctx, profile)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveAccountDrv(ctx *context.Context, str1 string, str2 string) error {
	if dbM.RemoveAccountDrvF != nil {
		return dbM.RemoveAccountDrvF(ctx, str1, str2)
	}
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) SetVersions(vrs Versions, overwrite bool) (err error) {
	return utils.ErrNotImplemented
}

func (dbM *DataDBMock) RemoveRatingProfileDrv(string) error {
	return utils.ErrNotImplemented
}
