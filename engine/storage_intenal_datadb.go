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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

type IntenalStorage struct {
	tasks [][]byte
	ms    Marshaler
	cache *ltcache.TransCache
}

func NewInternalStorage() *IntenalStorage {
	return &IntenalStorage{cache: ltcache.NewTransCache(config.CgrConfig().CacheCfg().AsTransCacheConfig()),
		ms: NewCodecMsgpackMarshaler()}
}

func NewInternalStorageJson() (internalStorage *IntenalStorage) {
	internalStorage = NewInternalStorage()
	internalStorage.ms = new(JSONBufMarshaler)
	return
}

func (is *IntenalStorage) Close() {}

func (is *IntenalStorage) Flush(_ string) error {
	is.cache = ltcache.NewTransCache(config.CgrConfig().CacheCfg().AsTransCacheConfig())
	return nil
}

func (is *IntenalStorage) Marshaler() Marshaler {
	return is.ms
}

func (is *IntenalStorage) SelectDatabase(dbName string) (err error) {
	return nil
}

func (is *IntenalStorage) GetKeysForPrefix(string) ([]string, error) {
	// keysForPrefix := make([]string, 0)
	// for key := range ms.dict {
	// 	if strings.HasPrefix(key, prefix) {
	// 		keysForPrefix = append(keysForPrefix, key)
	// 	}
	// }
	// is.cache.GetItemIDs(chID, prfx)
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) RebuildReverseForPrefix(string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveReverseForPrefix(string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetVersions(itm string) (vrs Versions, err error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetVersions(vrs Versions, overwrite bool) (err error) {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveVersions(vrs Versions) (err error) {
	return utils.ErrNotImplemented
}

func (is *IntenalStorage) GetStorageType() string {
	return utils.INTERNAL
}
func (is *IntenalStorage) IsDBEmpty() (resp bool, err error) {
	return false, utils.ErrNotImplemented
}

func (is *IntenalStorage) HasDataDrv(string, string, string) (bool, error) {
	return false, utils.ErrNotImplemented
}
func (is *IntenalStorage) GetRatingPlanDrv(string) (*RatingPlan, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetRatingPlanDrv(*RatingPlan) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveRatingPlanDrv(key string) (err error) {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetRatingProfileDrv(string) (*RatingProfile, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetRatingProfileDrv(*RatingProfile) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveRatingProfileDrv(string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetDestination(string, bool, string) (*Destination, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetDestination(*Destination, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveDestination(string, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) SetReverseDestination(*Destination, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetReverseDestination(string, bool, string) ([]string, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) UpdateReverseDestination(*Destination, *Destination, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetActionsDrv(string) (Actions, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetActionsDrv(string, Actions) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveActionsDrv(string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetSharedGroupDrv(string) (*SharedGroup, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetSharedGroupDrv(*SharedGroup) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveSharedGroupDrv(id, transactionID string) (err error) {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetActionTriggersDrv(string) (ActionTriggers, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetActionTriggersDrv(string, ActionTriggers) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveActionTriggersDrv(string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetActionPlan(string, bool, string) (*ActionPlan, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetActionPlan(string, *ActionPlan, bool, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveActionPlan(key string, transactionID string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetAllActionPlans() (map[string]*ActionPlan, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) GetAccountActionPlans(acntID string, skipCache bool,
	transactionID string) (apIDs []string, err error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetAccountActionPlans(acntID string, apIDs []string, overwrite bool) (err error) {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemAccountActionPlans(acntID string, apIDs []string) (err error) {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) PushTask(*Task) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) PopTask() (*Task, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) GetAccount(string) (*Account, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetAccount(*Account) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveAccount(string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetResourceProfileDrv(string, string) (*ResourceProfile, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetResourceProfileDrv(*ResourceProfile) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveResourceProfileDrv(string, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetResourceDrv(string, string) (*Resource, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetResourceDrv(*Resource) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveResourceDrv(string, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetTimingDrv(string) (*utils.TPTiming, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetTimingDrv(*utils.TPTiming) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveTimingDrv(string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) AddLoadHistory(*utils.LoadInstance, int, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetFilterIndexesDrv(cacheID, itemIDPrefix, filterType string,
	fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetFilterIndexesDrv(cacheID, itemIDPrefix string,
	indexes map[string]utils.StringMap, commit bool, transactionID string) (err error) {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveFilterIndexesDrv(cacheID, itemIDPrefix string) (err error) {
	return utils.ErrNotImplemented
}

func (is *IntenalStorage) MatchFilterIndexDrv(cacheID, itemIDPrefix,
	filterType, fieldName, fieldVal string) (itemIDs utils.StringMap, err error) {
	return nil, utils.ErrNotImplemented
}

func (is *IntenalStorage) GetStatQueueProfileDrv(tenant string, id string) (*StatQueueProfile, error) {
	x, ok := is.cache.Get(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*StatQueueProfile), nil

}
func (is *IntenalStorage) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	is.cache.Set(utils.CacheStatQueueProfiles, sq.TenantID(), sq, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (is *IntenalStorage) RemStatQueueProfileDrv(tenant, id string) (err error) {
	is.cache.Remove(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (is *IntenalStorage) GetStoredStatQueueDrv(tenant, id string) (sq *StoredStatQueue, err error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetStoredStatQueueDrv(sq *StoredStatQueue) (err error) {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemStoredStatQueueDrv(tenant, id string) (err error) {
	return
}

func (is *IntenalStorage) GetThresholdProfileDrv(tenant, id string) (tp *ThresholdProfile, err error) {
	x, ok := is.cache.Get(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ThresholdProfile), nil
}

func (is *IntenalStorage) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	is.cache.Set(utils.CacheThresholdProfiles, tp.TenantID(), tp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (is *IntenalStorage) RemThresholdProfileDrv(tenant, id string) (err error) {
	is.cache.Remove(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (is *IntenalStorage) GetThresholdDrv(tenant, id string) (*Threshold, error) {
	x, ok := is.cache.Get(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Threshold), nil
}

func (is *IntenalStorage) SetThresholdDrv(th *Threshold) (err error) {
	is.cache.Set(utils.CacheThresholds, th.TenantID(), th, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (is *IntenalStorage) RemoveThresholdDrv(tenant, id string) (err error) {
	is.cache.Remove(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (is *IntenalStorage) GetFilterDrv(tenant, id string) (*Filter, error) {
	x, ok := is.cache.Get(utils.CacheFilters, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Filter), nil
}

func (is *IntenalStorage) SetFilterDrv(fltr *Filter) (err error) {
	is.cache.Set(utils.CacheFilters, fltr.TenantID(), fltr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (is *IntenalStorage) RemoveFilterDrv(tenant, id string) (err error) {
	is.cache.Remove(utils.CacheFilters, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (is *IntenalStorage) GetSupplierProfileDrv(string, string) (*SupplierProfile, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetSupplierProfileDrv(*SupplierProfile) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveSupplierProfileDrv(string, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetAttributeProfileDrv(string, string) (*AttributeProfile, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetAttributeProfileDrv(*AttributeProfile) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveAttributeProfileDrv(string, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetChargerProfileDrv(string, string) (*ChargerProfile, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetChargerProfileDrv(*ChargerProfile) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveChargerProfileDrv(string, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetDispatcherProfileDrv(string, string) (*DispatcherProfile, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetDispatcherProfileDrv(*DispatcherProfile) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveDispatcherProfileDrv(string, string) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetLoadIDsDrv(loadIDs map[string]int64) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) GetDispatcherHostDrv(string, string) (*DispatcherHost, error) {
	return nil, utils.ErrNotImplemented
}
func (is *IntenalStorage) SetDispatcherHostDrv(*DispatcherHost) error {
	return utils.ErrNotImplemented
}
func (is *IntenalStorage) RemoveDispatcherHostDrv(string, string) error {
	return utils.ErrNotImplemented
}
