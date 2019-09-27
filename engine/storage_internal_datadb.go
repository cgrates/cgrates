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
	"bytes"
	"compress/zlib"
	"io/ioutil"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

type InternalDB struct {
	tasks [][]byte
	ms    Marshaler
	db    *ltcache.TransCache
}

func NewInternalDB() *InternalDB {
	return &InternalDB{db: ltcache.NewTransCache(config.CgrConfig().CacheCfg().AsTransCacheConfig()),
		ms: NewCodecMsgpackMarshaler()}
}

func NewInternalDBJson() (InternalDB *InternalDB) {
	InternalDB = NewInternalDB()
	InternalDB.ms = new(JSONBufMarshaler)
	return
}

func (iDB *InternalDB) Close() {}

func (iDB *InternalDB) Flush(_ string) error {
	iDB.db = ltcache.NewTransCache(config.CgrConfig().CacheCfg().AsTransCacheConfig())
	return nil
}

func (iDB *InternalDB) Marshaler() Marshaler {
	return iDB.ms
}

func (iDB *InternalDB) SelectDatabase(dbName string) (err error) {
	return nil
}

func (iDB *InternalDB) GetKeysForPrefix(string) ([]string, error) {
	// keysForPrefix := make([]string, 0)
	// for key := range ms.dict {
	// 	if strings.HasPrefix(key, prefix) {
	// 		keysForPrefix = append(keysForPrefix, key)
	// 	}
	// }
	// iDB.cache.GetItemIDs(chID, prfx)
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) RebuildReverseForPrefix(string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) RemoveReverseForPrefix(string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetVersions(itm string) (vrs Versions, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) SetVersions(vrs Versions, overwrite bool) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) RemoveVersions(vrs Versions) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetStorageType() string {
	return utils.INTERNAL
}

func (iDB *InternalDB) IsDBEmpty() (resp bool, err error) {
	return false, utils.ErrNotImplemented
}

func (iDB *InternalDB) HasDataDrv(string, string, string) (bool, error) {
	return false, utils.ErrNotImplemented
}

func (iDB *InternalDB) GetRatingPlanDrv(id string) (rp *RatingPlan, err error) {
	x, ok := iDB.db.Get(utils.CacheRatingPlans, id)
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	b := bytes.NewBuffer(x.([]byte))
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil, err
	}
	out, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	r.Close()
	err = iDB.ms.Unmarshal(out, &rp)
	if err != nil {
		return nil, err
	}
	return
}

func (iDB *InternalDB) SetRatingPlanDrv(rp *RatingPlan) (err error) {
	result, err := iDB.ms.Marshal(rp)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(result)
	w.Close()
	iDB.db.Set(utils.CacheRatingPlans, rp.Id, b, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveRatingPlanDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheRatingPlans, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetRatingProfileDrv(string) (*RatingProfile, error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) SetRatingProfileDrv(*RatingProfile) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) RemoveRatingProfileDrv(string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetDestination(string, bool, string) (*Destination, error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) SetDestination(*Destination, string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) RemoveDestination(string, string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) SetReverseDestination(*Destination, string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetReverseDestination(string, bool, string) ([]string, error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) UpdateReverseDestination(*Destination, *Destination, string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetActionsDrv(id string) (Actions, error) {
	x, ok := iDB.db.Get(utils.CacheActions, id)
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(Actions), nil
}

func (iDB *InternalDB) SetActionsDrv(id string, acts Actions) (err error) {
	iDB.db.Set(utils.CacheActions, id, acts, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveActionsDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheActions, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetSharedGroupDrv(id string) (*SharedGroup, error) {
	x, ok := iDB.db.Get(utils.CacheSharedGroups, id)
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*SharedGroup), nil
}

func (iDB *InternalDB) SetSharedGroupDrv(sh *SharedGroup) (err error) {
	iDB.db.Set(utils.CacheSharedGroups, sh.Id, sh, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveSharedGroupDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheSharedGroups, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetActionTriggersDrv(id string) (ActionTriggers, error) {
	x, ok := iDB.db.Get(utils.CacheActionTriggers, id)
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(ActionTriggers), nil
}

func (iDB *InternalDB) SetActionTriggersDrv(id string, at ActionTriggers) (err error) {
	iDB.db.Set(utils.CacheActionTriggers, id, at, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveActionTriggersDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheActionTriggers, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetActionPlan(string, bool, string) (*ActionPlan, error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) SetActionPlan(string, *ActionPlan, bool, string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) RemoveActionPlan(key string, transactionID string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetAllActionPlans() (map[string]*ActionPlan, error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) GetAccountActionPlans(acntID string, skipCache bool,
	transactionID string) (apIDs []string, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) SetAccountActionPlans(acntID string, apIDs []string, overwrite bool) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) RemAccountActionPlans(acntID string, apIDs []string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) PushTask(*Task) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) PopTask() (*Task, error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) GetAccount(string) (*Account, error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) SetAccount(*Account) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) RemoveAccount(string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetResourceProfileDrv(tenant, id string) (*ResourceProfile, error) {
	x, ok := iDB.db.Get(utils.CacheResourceProfiles, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ResourceProfile), nil
}

func (iDB *InternalDB) SetResourceProfileDrv(rp *ResourceProfile) (err error) {
	iDB.db.Set(utils.CacheResourceProfiles, rp.TenantID(), rp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveResourceProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheResourceProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetResourceDrv(tenant, id string) (*Resource, error) {
	x, ok := iDB.db.Get(utils.CacheResources, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Resource), nil
}

func (iDB *InternalDB) SetResourceDrv(r *Resource) (err error) {
	iDB.db.Set(utils.CacheResources, r.TenantID(), r, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveResourceDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheResources, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetTimingDrv(id string) (*utils.TPTiming, error) {
	x, ok := iDB.db.Get(utils.CacheTimings, id)
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*utils.TPTiming), nil
}

func (iDB *InternalDB) SetTimingDrv(timing *utils.TPTiming) (err error) {
	iDB.db.Set(utils.CacheTimings, timing.ID, timing, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveTimingDrv(id string) (err error) {
	iDB.db.Remove(utils.CacheTimings, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) AddLoadHistory(*utils.LoadInstance, int, string) error {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetFilterIndexesDrv(cacheID, itemIDPrefix, filterType string,
	fldNameVal map[string]string) (indexes map[string]utils.StringMap, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) SetFilterIndexesDrv(cacheID, itemIDPrefix string,
	indexes map[string]utils.StringMap, commit bool, transactionID string) (err error) {
	return utils.ErrNotImplemented
}
func (iDB *InternalDB) RemoveFilterIndexesDrv(cacheID, itemIDPrefix string) (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) MatchFilterIndexDrv(cacheID, itemIDPrefix,
	filterType, fieldName, fieldVal string) (itemIDs utils.StringMap, err error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) GetStatQueueProfileDrv(tenant string, id string) (*StatQueueProfile, error) {
	x, ok := iDB.db.Get(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*StatQueueProfile), nil

}
func (iDB *InternalDB) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	iDB.db.Set(utils.CacheStatQueueProfiles, sq.TenantID(), sq, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemStatQueueProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetStoredStatQueueDrv(tenant, id string) (sq *StoredStatQueue, err error) {
	x, ok := iDB.db.Get(utils.CacheStatQueues, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*StoredStatQueue), nil
}
func (iDB *InternalDB) SetStoredStatQueueDrv(sq *StoredStatQueue) (err error) {
	iDB.db.Set(utils.CacheStatQueues, utils.ConcatenatedKey(sq.Tenant, sq.ID), sq, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemStoredStatQueueDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheStatQueues, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetThresholdProfileDrv(tenant, id string) (tp *ThresholdProfile, err error) {
	x, ok := iDB.db.Get(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ThresholdProfile), nil
}

func (iDB *InternalDB) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	iDB.db.Set(utils.CacheThresholdProfiles, tp.TenantID(), tp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemThresholdProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetThresholdDrv(tenant, id string) (*Threshold, error) {
	x, ok := iDB.db.Get(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Threshold), nil
}

func (iDB *InternalDB) SetThresholdDrv(th *Threshold) (err error) {
	iDB.db.Set(utils.CacheThresholds, th.TenantID(), th, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveThresholdDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetFilterDrv(tenant, id string) (*Filter, error) {
	x, ok := iDB.db.Get(utils.CacheFilters, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Filter), nil
}

func (iDB *InternalDB) SetFilterDrv(fltr *Filter) (err error) {
	iDB.db.Set(utils.CacheFilters, fltr.TenantID(), fltr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveFilterDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheFilters, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetSupplierProfileDrv(tenant, id string) (*SupplierProfile, error) {
	x, ok := iDB.db.Get(utils.CacheSupplierProfiles, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*SupplierProfile), nil
}
func (iDB *InternalDB) SetSupplierProfileDrv(spp *SupplierProfile) (err error) {
	iDB.db.Set(utils.CacheSupplierProfiles, spp.TenantID(), spp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveSupplierProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheSupplierProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetAttributeProfileDrv(tenant, id string) (*AttributeProfile, error) {
	x, ok := iDB.db.Get(utils.CacheAttributeProfiles, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*AttributeProfile), nil
}
func (iDB *InternalDB) SetAttributeProfileDrv(attr *AttributeProfile) (err error) {
	iDB.db.Set(utils.CacheAttributeProfiles, attr.TenantID(), attr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheAttributeProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetChargerProfileDrv(tenant, id string) (*ChargerProfile, error) {
	x, ok := iDB.db.Get(utils.CacheChargerProfiles, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ChargerProfile), nil
}
func (iDB *InternalDB) SetChargerProfileDrv(chr *ChargerProfile) (err error) {
	iDB.db.Set(utils.CacheChargerProfiles, chr.TenantID(), chr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveChargerProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheChargerProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetDispatcherProfileDrv(tenant, id string) (*DispatcherProfile, error) {
	x, ok := iDB.db.Get(utils.CacheDispatcherProfiles, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*DispatcherProfile), nil
}
func (iDB *InternalDB) SetDispatcherProfileDrv(dpp *DispatcherProfile) (err error) {
	iDB.db.Set(utils.CacheDispatcherProfiles, dpp.TenantID(), dpp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveDispatcherProfileDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheDispatcherProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	x, ok := iDB.db.Get(utils.CacheLoadIDs, utils.LoadIDs)
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	loadIDs = x.(map[string]int64)
	if itemIDPrefix != utils.EmptyString {
		return map[string]int64{itemIDPrefix: loadIDs[itemIDPrefix]}, nil
	}
	return

}
func (iDB *InternalDB) SetLoadIDsDrv(loadIDs map[string]int64) (err error) {
	iDB.db.Set(utils.CacheLoadIDs, utils.LoadIDs, loadIDs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) GetDispatcherHostDrv(tenant, id string) (*DispatcherHost, error) {
	x, ok := iDB.db.Get(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, id))
	if ok && x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*DispatcherHost), nil
}
func (iDB *InternalDB) SetDispatcherHostDrv(dpp *DispatcherHost) (err error) {
	iDB.db.Set(utils.CacheDispatcherHosts, dpp.TenantID(), dpp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemoveDispatcherHostDrv(tenant, id string) (err error) {
	iDB.db.Remove(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
