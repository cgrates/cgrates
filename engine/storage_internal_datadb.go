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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
)

// InternalDB is used as a DataDB and a StorDB
type InternalDB struct {
	tasks               []*Task
	mu                  sync.RWMutex
	stringIndexedFields []string
	prefixIndexedFields []string
	indexedFieldsMutex  sync.RWMutex   // used for reload
	cnter               *utils.Counter // used for OrderID for cdr
	ms                  Marshaler
}

// NewInternalDB constructs an InternalDB
func NewInternalDB(stringIndexedFields, prefixIndexedFields []string, isDataDB bool) (iDB *InternalDB) {
	ms, _ := NewMarshaler(config.CgrConfig().GeneralCfg().DBDataEncoding)
	iDB = &InternalDB{
		stringIndexedFields: stringIndexedFields,
		prefixIndexedFields: prefixIndexedFields,
		cnter:               utils.NewCounter(time.Now().UnixNano(), 0),
		ms:                  ms,
	}
	return
}

// SetStringIndexedFields set the stringIndexedFields, used at StorDB reload (is thread safe)
func (iDB *InternalDB) SetStringIndexedFields(stringIndexedFields []string) {
	iDB.indexedFieldsMutex.Lock()
	iDB.stringIndexedFields = stringIndexedFields
	iDB.indexedFieldsMutex.Unlock()
}

// SetPrefixIndexedFields set the prefixIndexedFields, used at StorDB reload (is thread safe)
func (iDB *InternalDB) SetPrefixIndexedFields(prefixIndexedFields []string) {
	iDB.indexedFieldsMutex.Lock()
	iDB.prefixIndexedFields = prefixIndexedFields
	iDB.indexedFieldsMutex.Unlock()
}

// Close only to implement Storage interface
func (iDB *InternalDB) Close() {}

// Flush clears the cache
func (iDB *InternalDB) Flush(string) error {
	Cache.Clear(nil)
	return nil
}

// SelectDatabase only to implement Storage interface
func (iDB *InternalDB) SelectDatabase(string) (err error) {
	return nil
}

// GetKeysForPrefix returns the keys from cache that have the given prefix
func (iDB *InternalDB) GetKeysForPrefix(prefix string) (ids []string, err error) {
	keyLen := len(utils.DestinationPrefix)
	if len(prefix) < keyLen {
		err = fmt.Errorf("unsupported prefix in GetKeysForPrefix: %s", prefix)
		return
	}
	category := prefix[:keyLen] // prefix length
	queryPrefix := prefix[keyLen:]
	ids = Cache.GetItemIDs(utils.CachePrefixToInstance[category], queryPrefix)
	for i := range ids {
		ids[i] = category + ids[i]
	}
	return
}

func (iDB *InternalDB) RemoveKeysForPrefix(prefix string) (err error) {
	var keys []string
	if keys, err = iDB.GetKeysForPrefix(prefix); err != nil {
		return
	}
	for _, key := range keys {
		Cache.RemoveWithoutReplicate(utils.CacheReverseDestinations, key,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) GetVersions(itm string) (vrs Versions, err error) {
	x, ok := Cache.Get(utils.CacheVersions, utils.VersionName)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	provVrs := x.(Versions)
	if itm != "" {
		if _, has := provVrs[itm]; !has {
			return nil, utils.ErrNotFound
		}
		return Versions{itm: provVrs[itm]}, nil
	}
	return provVrs, nil
}

func (iDB *InternalDB) SetVersions(vrs Versions, overwrite bool) (err error) {
	if overwrite {
		if err = iDB.RemoveVersions(nil); err != nil {
			return
		}
	}
	x, ok := Cache.Get(utils.CacheVersions, utils.VersionName)
	if !ok || x == nil {
		Cache.SetWithoutReplicate(utils.CacheVersions, utils.VersionName, vrs, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return
	}
	provVrs := x.(Versions)
	for key, val := range vrs {
		provVrs[key] = val
	}
	Cache.SetWithoutReplicate(utils.CacheVersions, utils.VersionName, provVrs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveVersions(vrs Versions) (err error) {
	if len(vrs) != 0 {
		var internalVersions Versions
		x, ok := Cache.Get(utils.CacheVersions, utils.VersionName)
		if !ok || x == nil {
			return utils.ErrNotFound
		}
		internalVersions = x.(Versions)
		for key := range vrs {
			delete(internalVersions, key)
		}
		Cache.SetWithoutReplicate(utils.CacheVersions, utils.VersionName, internalVersions, nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return
	}
	Cache.RemoveWithoutReplicate(utils.CacheVersions, utils.VersionName,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

// GetStorageType returns the storage type
func (iDB *InternalDB) GetStorageType() string {
	return utils.INTERNAL
}

// IsDBEmpty returns true if the cache is empty
func (iDB *InternalDB) IsDBEmpty() (isEmpty bool, err error) {
	for cacheInstance := range utils.CacheInstanceToPrefix {
		if len(Cache.GetItemIDs(cacheInstance, utils.EmptyString)) != 0 {
			return
		}
	}
	isEmpty = true
	return
}

func (iDB *InternalDB) HasDataDrv(category, subject, tenant string) (bool, error) {
	switch category {
	case utils.DestinationPrefix, utils.RatingPlanPrefix, utils.RatingProfilePrefix,
		utils.ActionPrefix, utils.ActionPlanPrefix, utils.AccountPrefix:
		return Cache.HasItem(utils.CachePrefixToInstance[category], subject), nil
	case utils.ResourcesPrefix, utils.ResourceProfilesPrefix, utils.StatQueuePrefix,
		utils.StatQueueProfilePrefix, utils.ThresholdPrefix, utils.ThresholdProfilePrefix,
		utils.FilterPrefix, utils.RouteProfilePrefix, utils.AttributeProfilePrefix,
		utils.ChargerProfilePrefix, utils.DispatcherProfilePrefix, utils.DispatcherHostPrefix:
		return Cache.HasItem(utils.CachePrefixToInstance[category], utils.ConcatenatedKey(tenant, subject)), nil
	}
	return false, errors.New("Unsupported HasData category")
}

func (iDB *InternalDB) GetRatingPlanDrv(id string) (rp *RatingPlan, err error) {
	x, ok := Cache.Get(utils.CacheRatingPlans, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*RatingPlan), nil
}

func (iDB *InternalDB) SetRatingPlanDrv(rp *RatingPlan) (err error) {
	Cache.SetWithoutReplicate(utils.CacheRatingPlans, rp.Id, rp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveRatingPlanDrv(id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheRatingPlans, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetRatingProfileDrv(id string) (rp *RatingProfile, err error) {
	x, ok := Cache.Get(utils.CacheRatingProfiles, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*RatingProfile), nil
}

func (iDB *InternalDB) SetRatingProfileDrv(rp *RatingProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheRatingProfiles, rp.Id, rp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveRatingProfileDrv(id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheRatingProfiles, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetDestinationDrv(key, _ string) (dest *Destination, err error) {
	if x, ok := Cache.Get(utils.CacheDestinations, key); ok && x != nil {
		return x.(*Destination), nil
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetDestinationDrv(dest *Destination, transactionID string) (err error) {
	Cache.SetWithoutReplicate(utils.CacheDestinations, dest.Id, dest, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveDestinationDrv(destID string, transactionID string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheDestinations, destID,
		cacheCommit(transactionID), transactionID)
	return
}

func (iDB *InternalDB) RemoveReverseDestinationDrv(dstID, prfx, transactionID string) (err error) {
	var revDst []string
	if Cache.HasItem(utils.CacheReverseDestinations, prfx) {
		x, ok := Cache.Get(utils.CacheReverseDestinations, prfx)
		if !ok || x == nil {
			return utils.ErrNotFound
		}
		revDst = x.([]string)
	}
	mpRevDst := utils.NewStringSet(revDst)
	mpRevDst.Remove(dstID)
	if mpRevDst.Size() != 0 {
		Cache.SetWithoutReplicate(utils.CacheReverseDestinations, prfx, mpRevDst.AsSlice(), nil,
			cacheCommit(transactionID), transactionID)
	} else {
		Cache.RemoveWithoutReplicate(utils.CacheReverseDestinations, prfx,
			cacheCommit(transactionID), transactionID)
	}
	return
}

func (iDB *InternalDB) SetReverseDestinationDrv(destID string, prefixes []string, transactionID string) (err error) {
	for _, p := range prefixes {
		var revDst []string
		if Cache.HasItem(utils.CacheReverseDestinations, p) {
			if x, ok := Cache.Get(utils.CacheReverseDestinations, p); ok && x != nil {
				revDst = x.([]string)
			}
		}
		mpRevDst := utils.NewStringSet(revDst)
		mpRevDst.Add(destID)
		// for ReverseDestination we will use Groups
		Cache.SetWithoutReplicate(utils.CacheReverseDestinations, p, mpRevDst.AsSlice(), nil,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) GetReverseDestinationDrv(prefix, transactionID string) (ids []string, err error) {
	if x, ok := Cache.Get(utils.CacheReverseDestinations, prefix); ok && x != nil {
		if ids = x.([]string); len(ids) != 0 {
			return
		}
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) GetActionsDrv(id string) (acts Actions, err error) {
	if x, ok := Cache.Get(utils.CacheActions, id); ok && x != nil {
		return x.(Actions), err
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetActionsDrv(id string, acts Actions) (err error) {
	Cache.SetWithoutReplicate(utils.CacheActions, id, acts, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveActionsDrv(id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheActions, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetSharedGroupDrv(id string) (sh *SharedGroup, err error) {
	if x, ok := Cache.Get(utils.CacheSharedGroups, id); ok && x != nil {
		return x.(*SharedGroup).Clone(), err
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetSharedGroupDrv(sh *SharedGroup) (err error) {
	Cache.SetWithoutReplicate(utils.CacheSharedGroups, sh.Id, sh, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveSharedGroupDrv(id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheSharedGroups, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetActionTriggersDrv(id string) (at ActionTriggers, err error) {
	if x, ok := Cache.Get(utils.CacheActionTriggers, id); ok && x != nil {
		return x.(ActionTriggers).Clone(), err
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetActionTriggersDrv(id string, at ActionTriggers) (err error) {
	Cache.SetWithoutReplicate(utils.CacheActionTriggers, id, at, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveActionTriggersDrv(id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheActionTriggers, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetActionPlanDrv(key string, skipCache bool,
	transactionID string) (ats *ActionPlan, err error) {
	if x, ok := Cache.Get(utils.CacheActionPlans, key); ok && x != nil {
		return x.(*ActionPlan), nil
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetActionPlanDrv(key string, ats *ActionPlan,
	overwrite bool, transactionID string) (err error) {
	cCommit := cacheCommit(transactionID)
	if len(ats.ActionTimings) == 0 {
		Cache.RemoveWithoutReplicate(utils.CacheActionPlans, key,
			cCommit, transactionID)
		return
	}
	if !overwrite {
		// get existing action plan to merge the account ids
		if existingAts, _ := iDB.GetActionPlanDrv(key, true,
			transactionID); existingAts != nil {
			if ats.AccountIDs == nil && len(existingAts.AccountIDs) > 0 {
				ats.AccountIDs = make(utils.StringMap)
			}
			for accID := range existingAts.AccountIDs {
				ats.AccountIDs[accID] = true
			}
		}
	}
	Cache.SetWithoutReplicate(utils.CacheActionPlans, key, ats, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveActionPlanDrv(key string, transactionID string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheActionPlans, key, cacheCommit(transactionID), transactionID)
	return
}

func (iDB *InternalDB) GetAllActionPlansDrv() (ats map[string]*ActionPlan, err error) {
	var keys []string
	if keys, err = iDB.GetKeysForPrefix(utils.ActionPlanPrefix); err != nil {
		return
	}

	ats = make(map[string]*ActionPlan, len(keys))
	for _, key := range keys {
		var ap *ActionPlan
		if ap, err = iDB.GetActionPlanDrv(key[len(utils.ActionPlanPrefix):], false, utils.NonTransactional); err != nil {
			ats = nil
			return
		}
		ats[key[len(utils.ActionPlanPrefix):]] = ap
	}
	return
}

func (iDB *InternalDB) GetAccountActionPlansDrv(acntID string,
	skipCache bool, transactionID string) (apIDs []string, err error) {
	if x, ok := Cache.Get(utils.CacheAccountActionPlans, acntID); ok && x != nil {
		return x.([]string), nil
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetAccountActionPlansDrv(acntID string, apIDs []string, overwrite bool) (err error) {
	if !overwrite {
		var oldaPlIDs []string
		if oldaPlIDs, err = iDB.GetAccountActionPlansDrv(acntID,
			true, utils.NonTransactional); err != nil && err != utils.ErrNotFound {
			return
		}
		err = nil
		for _, oldAPid := range oldaPlIDs {
			if !utils.IsSliceMember(apIDs, oldAPid) {
				apIDs = append(apIDs, oldAPid)
			}
		}
	}
	Cache.SetWithoutReplicate(utils.CacheAccountActionPlans, acntID, apIDs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemAccountActionPlansDrv(acntID string, apIDs []string) (err error) {
	if len(apIDs) == 0 {
		Cache.RemoveWithoutReplicate(utils.CacheAccountActionPlans, acntID,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return
	}
	var oldaPlIDs []string
	if oldaPlIDs, err = iDB.GetAccountActionPlansDrv(acntID,
		true, utils.NonTransactional); err != nil {
		return
	}
	for i := 0; i < len(oldaPlIDs); {
		if utils.IsSliceMember(apIDs, oldaPlIDs[i]) {
			oldaPlIDs = append(oldaPlIDs[:i], oldaPlIDs[i+1:]...)
			continue
		}
		i++
	}
	if len(oldaPlIDs) == 0 {
		Cache.RemoveWithoutReplicate(utils.CacheAccountActionPlans, acntID,
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
		return
	}
	Cache.SetWithoutReplicate(utils.CacheAccountActionPlans, acntID, oldaPlIDs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) PushTask(t *Task) (err error) {
	iDB.mu.Lock()
	iDB.tasks = append(iDB.tasks, t)
	iDB.mu.Unlock()
	return
}

func (iDB *InternalDB) PopTask() (t *Task, err error) {
	iDB.mu.Lock()
	if len(iDB.tasks) > 0 {
		t = iDB.tasks[0]
		iDB.tasks[0] = nil
		iDB.tasks = iDB.tasks[1:]
	} else {
		err = utils.ErrNotFound
	}
	iDB.mu.Unlock()
	return
}

func (iDB *InternalDB) GetAccountDrv(id string) (acc *Account, err error) {
	if x, ok := Cache.Get(utils.CacheAccounts, id); ok && x != nil {
		return x.(*Account).Clone(), nil
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetAccountDrv(acc *Account) (err error) {
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(acc.BalanceMap) == 0 {
		if ac, err := iDB.GetAccountDrv(acc.ID); err == nil && !ac.allBalancesExpired() {
			ac.ActionTriggers = acc.ActionTriggers
			ac.UnitCounters = acc.UnitCounters
			ac.AllowNegative = acc.AllowNegative
			ac.Disabled = acc.Disabled
			acc = ac
		}
	}
	acc.UpdateTime = time.Now()
	Cache.SetWithoutReplicate(utils.CacheAccounts, acc.ID, acc, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveAccountDrv(id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheAccounts, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetResourceProfileDrv(tenant, id string) (rp *ResourceProfile, err error) {
	if x, ok := Cache.Get(utils.CacheResourceProfiles, utils.ConcatenatedKey(tenant, id)); ok && x != nil {
		return x.(*ResourceProfile), nil
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetResourceProfileDrv(rp *ResourceProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheResourceProfiles, rp.TenantID(), rp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveResourceProfileDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheResourceProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetResourceDrv(tenant, id string) (r *Resource, err error) {
	if x, ok := Cache.Get(utils.CacheResources, utils.ConcatenatedKey(tenant, id)); ok && x != nil {
		return x.(*Resource), nil
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetResourceDrv(r *Resource) (err error) {
	Cache.SetWithoutReplicate(utils.CacheResources, r.TenantID(), r, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveResourceDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheResources, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetTimingDrv(id string) (tmg *utils.TPTiming, err error) {
	x, ok := Cache.Get(utils.CacheTimings, id)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*utils.TPTiming), nil
}

func (iDB *InternalDB) SetTimingDrv(timing *utils.TPTiming) (err error) {
	Cache.SetWithoutReplicate(utils.CacheTimings, timing.ID, timing, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveTimingDrv(id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheTimings, id,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error) {
	return nil, nil
}

func (iDB *InternalDB) AddLoadHistory(*utils.LoadInstance, int, string) error {
	return nil
}

func (iDB *InternalDB) GetStatQueueProfileDrv(tenant string, id string) (sq *StatQueueProfile, err error) {
	x, ok := Cache.Get(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*StatQueueProfile), nil

}
func (iDB *InternalDB) SetStatQueueProfileDrv(sq *StatQueueProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheStatQueueProfiles, sq.TenantID(), sq, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemStatQueueProfileDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetStatQueueDrv(tenant, id string) (sq *StatQueue, err error) {
	x, ok := Cache.Get(utils.CacheStatQueues, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*StatQueue), nil
}
func (iDB *InternalDB) SetStatQueueDrv(ssq *StoredStatQueue, sq *StatQueue) (err error) {
	if sq == nil {
		sq, err = ssq.AsStatQueue(iDB.ms)
		if err != nil {
			return
		}
	}
	Cache.SetWithoutReplicate(utils.CacheStatQueues, utils.ConcatenatedKey(sq.Tenant, sq.ID), sq, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
func (iDB *InternalDB) RemStatQueueDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheStatQueues, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetThresholdProfileDrv(tenant, id string) (tp *ThresholdProfile, err error) {
	x, ok := Cache.Get(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ThresholdProfile), nil
}

func (iDB *InternalDB) SetThresholdProfileDrv(tp *ThresholdProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheThresholdProfiles, tp.TenantID(), tp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemThresholdProfileDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetThresholdDrv(tenant, id string) (th *Threshold, err error) {
	x, ok := Cache.Get(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Threshold), nil
}

func (iDB *InternalDB) SetThresholdDrv(th *Threshold) (err error) {
	Cache.SetWithoutReplicate(utils.CacheThresholds, th.TenantID(), th, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveThresholdDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetFilterDrv(tenant, id string) (fltr *Filter, err error) {
	x, ok := Cache.Get(utils.CacheFilters, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Filter), nil

}

func (iDB *InternalDB) SetFilterDrv(fltr *Filter) (err error) {
	if err = fltr.Compile(); err != nil {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheFilters, fltr.TenantID(), fltr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveFilterDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheFilters, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetRouteProfileDrv(tenant, id string) (spp *RouteProfile, err error) {
	x, ok := Cache.Get(utils.CacheRouteProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*RouteProfile), nil
}

func (iDB *InternalDB) SetRouteProfileDrv(spp *RouteProfile) (err error) {
	if err = spp.Compile(); err != nil {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheRouteProfiles, spp.TenantID(), spp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveRouteProfileDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheRouteProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetAttributeProfileDrv(tenant, id string) (attr *AttributeProfile, err error) {
	x, ok := Cache.Get(utils.CacheAttributeProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*AttributeProfile), nil
}

func (iDB *InternalDB) SetAttributeProfileDrv(attr *AttributeProfile) (err error) {
	if err = attr.Compile(); err != nil {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheAttributeProfiles, attr.TenantID(), attr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveAttributeProfileDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheAttributeProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetChargerProfileDrv(tenant, id string) (ch *ChargerProfile, err error) {
	x, ok := Cache.Get(utils.CacheChargerProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ChargerProfile), nil
}

func (iDB *InternalDB) SetChargerProfileDrv(chr *ChargerProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheChargerProfiles, chr.TenantID(), chr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveChargerProfileDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheChargerProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetDispatcherProfileDrv(tenant, id string) (dpp *DispatcherProfile, err error) {
	x, ok := Cache.Get(utils.CacheDispatcherProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*DispatcherProfile), nil
}

func (iDB *InternalDB) SetDispatcherProfileDrv(dpp *DispatcherProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheDispatcherProfiles, dpp.TenantID(), dpp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveDispatcherProfileDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheDispatcherProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetItemLoadIDsDrv(itemIDPrefix string) (loadIDs map[string]int64, err error) {
	x, ok := Cache.Get(utils.CacheLoadIDs, utils.LoadIDs)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	loadIDs = x.(map[string]int64)
	if itemIDPrefix != utils.EmptyString {
		return map[string]int64{itemIDPrefix: loadIDs[itemIDPrefix]}, nil
	}
	return
}

func (iDB *InternalDB) SetLoadIDsDrv(loadIDs map[string]int64) (err error) {
	Cache.SetWithoutReplicate(utils.CacheLoadIDs, utils.LoadIDs, loadIDs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetDispatcherHostDrv(tenant, id string) (dpp *DispatcherHost, err error) {
	x, ok := Cache.Get(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*DispatcherHost), nil
}

func (iDB *InternalDB) SetDispatcherHostDrv(dpp *DispatcherHost) (err error) {
	Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, dpp.TenantID(), dpp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveDispatcherHostDrv(tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveLoadIDsDrv() (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetIndexesDrv(idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
	if idxKey == utils.EmptyString { // return all
		indexes = make(map[string]utils.StringSet)
		for _, dbKey := range Cache.tCache.GetGroupItemIDs(idxItmType, tntCtx) {
			x, ok := Cache.Get(idxItmType, dbKey)
			if !ok || x == nil {
				continue
			}
			dbKey = strings.TrimPrefix(dbKey, tntCtx+utils.ConcatenatedKeySep)
			indexes[dbKey] = x.(utils.StringSet).Clone()
		}
		if len(indexes) == 0 {
			return nil, utils.ErrNotFound
		}
		return
	}
	dbKey := utils.ConcatenatedKey(tntCtx, idxKey)
	x, ok := Cache.Get(idxItmType, dbKey)
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return map[string]utils.StringSet{
		idxKey: x.(utils.StringSet).Clone(),
	}, nil
}

func (iDB *InternalDB) SetIndexesDrv(idxItmType, tntCtx string,
	indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
	if commit && transactionID != utils.EmptyString {
		for _, dbKey := range Cache.tCache.GetGroupItemIDs(idxItmType, tntCtx) {
			if !strings.HasPrefix(dbKey, "tmp_") || !strings.HasSuffix(dbKey, transactionID) {
				continue
			}
			x, ok := Cache.Get(idxItmType, dbKey)
			if !ok || x == nil {
				continue
			}
			Cache.RemoveWithoutReplicate(idxItmType, dbKey,
				cacheCommit(utils.NonTransactional), utils.NonTransactional)
			key := strings.TrimSuffix(strings.TrimPrefix(dbKey, "tmp_"), utils.ConcatenatedKeySep+transactionID)
			Cache.SetWithoutReplicate(idxItmType, key, x, []string{tntCtx},
				cacheCommit(utils.NonTransactional), utils.NonTransactional)
		}
		return
	}
	for idxKey, indx := range indexes {
		dbKey := utils.ConcatenatedKey(tntCtx, idxKey)
		if transactionID != utils.EmptyString {
			dbKey = "tmp_" + utils.ConcatenatedKey(dbKey, transactionID)
		}
		if len(indx) == 0 {
			Cache.SetWithoutReplicate(idxItmType, dbKey, nil, []string{tntCtx},
				cacheCommit(utils.NonTransactional), utils.NonTransactional)
			continue
		}
		//to be the same as HMSET
		if x, ok := Cache.Get(idxItmType, dbKey); ok && x != nil {
			indx = utils.JoinStringSet(indx, x.(utils.StringSet))
		}
		Cache.SetWithoutReplicate(idxItmType, dbKey, indx, []string{tntCtx},
			cacheCommit(utils.NonTransactional), utils.NonTransactional)
	}
	return
}

func (iDB *InternalDB) RemoveIndexesDrv(idxItmType, tntCtx, idxKey string) (err error) {
	if idxKey == utils.EmptyString {
		Cache.tCache.RemoveGroup(idxItmType, tntCtx, true, utils.EmptyString)
		return
	}
	Cache.RemoveWithoutReplicate(idxItmType, utils.ConcatenatedKey(tntCtx, idxKey), cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
