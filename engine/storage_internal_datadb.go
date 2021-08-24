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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/utils"
)

// InternalDB is used as a DataDB and a StorDB
type InternalDB struct {
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
func (iDB *InternalDB) GetKeysForPrefix(_ *context.Context, prefix string) (ids []string, err error) {
	keyLen := len(utils.AccountPrefix)
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
	return utils.Internal
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

func (iDB *InternalDB) HasDataDrv(_ *context.Context, category, subject, tenant string) (bool, error) {
	switch category {
	case utils.ResourcesPrefix, utils.ResourceProfilesPrefix, utils.StatQueuePrefix,
		utils.StatQueueProfilePrefix, utils.ThresholdPrefix, utils.ThresholdProfilePrefix,
		utils.FilterPrefix, utils.RouteProfilePrefix, utils.AttributeProfilePrefix,
		utils.ChargerProfilePrefix, utils.DispatcherProfilePrefix, utils.DispatcherHostPrefix:
		return Cache.HasItem(utils.CachePrefixToInstance[category], utils.ConcatenatedKey(tenant, subject)), nil
	}
	return false, errors.New("Unsupported HasData category")
}

func (iDB *InternalDB) GetResourceProfileDrv(_ *context.Context, tenant, id string) (rp *ResourceProfile, err error) {
	if x, ok := Cache.Get(utils.CacheResourceProfiles, utils.ConcatenatedKey(tenant, id)); ok && x != nil {
		return x.(*ResourceProfile), nil
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetResourceProfileDrv(_ *context.Context, rp *ResourceProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheResourceProfiles, rp.TenantID(), rp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveResourceProfileDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheResourceProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetResourceDrv(_ *context.Context, tenant, id string) (r *Resource, err error) {
	if x, ok := Cache.Get(utils.CacheResources, utils.ConcatenatedKey(tenant, id)); ok && x != nil {
		return x.(*Resource), nil
	}
	return nil, utils.ErrNotFound
}

func (iDB *InternalDB) SetResourceDrv(_ *context.Context, r *Resource) (err error) {
	Cache.SetWithoutReplicate(utils.CacheResources, r.TenantID(), r, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveResourceDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheResources, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetLoadHistory(int, bool, string) ([]*utils.LoadInstance, error) {
	return nil, nil
}

func (iDB *InternalDB) AddLoadHistory(*utils.LoadInstance, int, string) error {
	return nil
}

func (iDB *InternalDB) GetStatQueueProfileDrv(_ *context.Context, tenant string, id string) (sq *StatQueueProfile, err error) {
	x, ok := Cache.Get(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*StatQueueProfile), nil

}
func (iDB *InternalDB) SetStatQueueProfileDrv(_ *context.Context, sq *StatQueueProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheStatQueueProfiles, sq.TenantID(), sq, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemStatQueueProfileDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheStatQueueProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetStatQueueDrv(_ *context.Context, tenant, id string) (sq *StatQueue, err error) {
	x, ok := Cache.Get(utils.CacheStatQueues, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*StatQueue), nil
}
func (iDB *InternalDB) SetStatQueueDrv(_ *context.Context, ssq *StoredStatQueue, sq *StatQueue) (err error) {
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
func (iDB *InternalDB) RemStatQueueDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheStatQueues, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetThresholdProfileDrv(_ *context.Context, tenant, id string) (tp *ThresholdProfile, err error) {
	x, ok := Cache.Get(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ThresholdProfile), nil
}

func (iDB *InternalDB) SetThresholdProfileDrv(_ *context.Context, tp *ThresholdProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheThresholdProfiles, tp.TenantID(), tp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemThresholdProfileDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheThresholdProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetThresholdDrv(_ *context.Context, tenant, id string) (th *Threshold, err error) {
	x, ok := Cache.Get(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Threshold), nil
}

func (iDB *InternalDB) SetThresholdDrv(_ *context.Context, th *Threshold) (err error) {
	Cache.SetWithoutReplicate(utils.CacheThresholds, th.TenantID(), th, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveThresholdDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheThresholds, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetFilterDrv(_ *context.Context, tenant, id string) (fltr *Filter, err error) {
	x, ok := Cache.Get(utils.CacheFilters, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*Filter), nil

}

func (iDB *InternalDB) SetFilterDrv(_ *context.Context, fltr *Filter) (err error) {
	if err = fltr.Compile(); err != nil {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheFilters, fltr.TenantID(), fltr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveFilterDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheFilters, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetRouteProfileDrv(_ *context.Context, tenant, id string) (spp *RouteProfile, err error) {
	x, ok := Cache.Get(utils.CacheRouteProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*RouteProfile), nil
}

func (iDB *InternalDB) SetRouteProfileDrv(_ *context.Context, spp *RouteProfile) (err error) {
	if err = spp.Compile(); err != nil {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheRouteProfiles, spp.TenantID(), spp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveRouteProfileDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheRouteProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetAttributeProfileDrv(_ *context.Context, tenant, id string) (attr *AttributeProfile, err error) {
	x, ok := Cache.Get(utils.CacheAttributeProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*AttributeProfile), nil
}

func (iDB *InternalDB) SetAttributeProfileDrv(_ *context.Context, attr *AttributeProfile) (err error) {
	if err = attr.Compile(); err != nil {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheAttributeProfiles, attr.TenantID(), attr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveAttributeProfileDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheAttributeProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetChargerProfileDrv(_ *context.Context, tenant, id string) (ch *ChargerProfile, err error) {
	x, ok := Cache.Get(utils.CacheChargerProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ChargerProfile), nil
}

func (iDB *InternalDB) SetChargerProfileDrv(_ *context.Context, chr *ChargerProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheChargerProfiles, chr.TenantID(), chr, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveChargerProfileDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheChargerProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetDispatcherProfileDrv(_ *context.Context, tenant, id string) (dpp *DispatcherProfile, err error) {
	x, ok := Cache.Get(utils.CacheDispatcherProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*DispatcherProfile), nil
}

func (iDB *InternalDB) SetDispatcherProfileDrv(_ *context.Context, dpp *DispatcherProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheDispatcherProfiles, dpp.TenantID(), dpp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveDispatcherProfileDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheDispatcherProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetItemLoadIDsDrv(_ *context.Context, itemIDPrefix string) (loadIDs map[string]int64, err error) {
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

func (iDB *InternalDB) SetLoadIDsDrv(_ *context.Context, loadIDs map[string]int64) (err error) {
	Cache.SetWithoutReplicate(utils.CacheLoadIDs, utils.LoadIDs, loadIDs, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetDispatcherHostDrv(_ *context.Context, tenant, id string) (dpp *DispatcherHost, err error) {
	x, ok := Cache.Get(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*DispatcherHost), nil
}

func (iDB *InternalDB) SetDispatcherHostDrv(_ *context.Context, dpp *DispatcherHost) (err error) {
	Cache.SetWithoutReplicate(utils.CacheDispatcherHosts, dpp.TenantID(), dpp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveDispatcherHostDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheDispatcherHosts, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetRateProfileDrv(_ *context.Context, tenant, id string) (rpp *utils.RateProfile, err error) {
	x, ok := Cache.Get(utils.CacheRateProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*utils.RateProfile), nil
}

func (iDB *InternalDB) SetRateProfileDrv(_ *context.Context, rpp *utils.RateProfile) (err error) {
	if err = rpp.Compile(); err != nil {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheRateProfiles, rpp.TenantID(), rpp, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveRateProfileDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheRateProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetActionProfileDrv(_ *context.Context, tenant, id string) (ap *ActionProfile, err error) {
	x, ok := Cache.Get(utils.CacheActionProfiles, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*ActionProfile), nil
}

func (iDB *InternalDB) SetActionProfileDrv(_ *context.Context, ap *ActionProfile) (err error) {
	Cache.SetWithoutReplicate(utils.CacheActionProfiles, ap.TenantID(), ap, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveActionProfileDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheActionProfiles, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveLoadIDsDrv() (err error) {
	return utils.ErrNotImplemented
}

func (iDB *InternalDB) GetIndexesDrv(_ *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
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

func (iDB *InternalDB) SetIndexesDrv(_ *context.Context, idxItmType, tntCtx string,
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
			Cache.RemoveWithoutReplicate(idxItmType, dbKey,
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

func (iDB *InternalDB) RemoveIndexesDrv(_ *context.Context, idxItmType, tntCtx, idxKey string) (err error) {
	if idxKey == utils.EmptyString {
		Cache.tCache.RemoveGroup(idxItmType, tntCtx, true, utils.EmptyString)
		return
	}
	Cache.RemoveWithoutReplicate(idxItmType, utils.ConcatenatedKey(tntCtx, idxKey), cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) GetAccountDrv(_ *context.Context, tenant, id string) (ap *utils.Account, err error) {
	x, ok := Cache.Get(utils.CacheAccounts, utils.ConcatenatedKey(tenant, id))
	if !ok || x == nil {
		return nil, utils.ErrNotFound
	}
	return x.(*utils.Account).Clone(), nil
}

func (iDB *InternalDB) SetAccountDrv(_ *context.Context, ap *utils.Account) (err error) {
	Cache.SetWithoutReplicate(utils.CacheAccounts, ap.TenantID(), ap, nil,
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}

func (iDB *InternalDB) RemoveAccountDrv(_ *context.Context, tenant, id string) (err error) {
	Cache.RemoveWithoutReplicate(utils.CacheAccounts, utils.ConcatenatedKey(tenant, id),
		cacheCommit(utils.NonTransactional), utils.NonTransactional)
	return
}
