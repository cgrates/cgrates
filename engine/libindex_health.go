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
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

type IndexHealthArgsWith2Ch struct {
	IndexCacheLimit     int
	IndexCacheTTL       time.Duration
	IndexCacheStaticTTL bool

	ObjectCacheLimit     int
	ObjectCacheTTL       time.Duration
	ObjectCacheStaticTTL bool
}

type IndexHealthArgsWith3Ch struct {
	IndexCacheLimit     int
	IndexCacheTTL       time.Duration
	IndexCacheStaticTTL bool

	ObjectCacheLimit     int
	ObjectCacheTTL       time.Duration
	ObjectCacheStaticTTL bool

	FilterCacheLimit     int
	FilterCacheTTL       time.Duration
	FilterCacheStaticTTL bool
}

type AccountActionPlanIHReply struct {
	MissingAccountActionPlans map[string][]string // list of missing indexes for each object (the map has the key as the indexKey and a list of objects)
	BrokenReferences          map[string][]string // list of broken references (the map has the key as the objectID and a list of indexes)
}

// add cache in args API
func GetAccountActionPlansIndexHealth(dm *DataManager, objLimit, indexLimit int, objTTL, indexTTL time.Duration, objStaticTTL, indexStaticTTL bool) (rply *AccountActionPlanIHReply, err error) {
	// posible errors
	brokenRef := map[string][]string{}    // the actionPlans match the index but they are missing the account // broken reference
	missingIndex := map[string][]string{} // the indexes are not present but the action plans points to that account // misingAccounts

	// local cache
	indexesCache := ltcache.NewCache(objLimit, objTTL, objStaticTTL, nil)
	objectsCache := ltcache.NewCache(indexLimit, indexTTL, indexStaticTTL, nil)

	getCachedIndex := func(acntID string) (apIDs []string, err error) {
		if x, ok := indexesCache.Get(acntID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
		if apIDs, err = dm.GetAccountActionPlans(acntID, true, false, utils.NonTransactional); err != nil { // read from cache but do not write if not there
			if err == utils.ErrNotFound {
				indexesCache.Set(acntID, nil, nil)
			}
			return
		}
		indexesCache.Set(acntID, apIDs, nil)
		return
	}

	getCachedObject := func(apID string) (obj *ActionPlan, err error) {
		if x, ok := objectsCache.Get(apID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*ActionPlan), nil
		}
		if obj, err = dm.GetActionPlan(apID, true, false, utils.NonTransactional); err != nil { // read from cache but do not write if not there
			if err == utils.ErrNotFound {
				objectsCache.Set(apID, nil, nil)
			}
			return
		}
		objectsCache.Set(apID, obj, nil)
		return
	}

	var acntIDs []string // start with the indexes and check the references
	if acntIDs, err = dm.DataDB().GetKeysForPrefix(utils.AccountActionPlansPrefix); err != nil {
		err = fmt.Errorf("error <%s> querying keys for accountActionPlans", err.Error())
		return
	}

	for _, acntID := range acntIDs {
		acntID = strings.TrimPrefix(acntID, utils.AccountActionPlansPrefix) //
		var apIDs []string
		if apIDs, err = getCachedIndex(acntID); err != nil { // read from cache but do not write if not there
			err = fmt.Errorf("error <%s> querying the accountActionPlan: <%v>", err.Error(), acntID)
			return
		}
		for _, apID := range apIDs {
			var ap *ActionPlan
			if ap, err = getCachedObject(apID); err != nil {
				if err != utils.ErrNotFound {
					err = fmt.Errorf("error <%s> querying the actionPlan: <%v>", err.Error(), apID)
					return
				}
				err = nil
				brokenRef[apID] = nil
				continue

			}
			if !ap.AccountIDs.HasKey(acntID) { // the action plan exists but doesn't point towards the account we have index
				brokenRef[apID] = append(brokenRef[apID], acntID)
			}
		}
	}

	var apIDs []string // we have all the indexes in cache now do a reverse check
	if apIDs, err = dm.DataDB().GetKeysForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
		err = fmt.Errorf("error <%s> querying keys for actionPlans", err.Error())
		return
	}

	for _, apID := range apIDs {
		apID = strings.TrimPrefix(apID, utils.ACTION_PLAN_PREFIX) //
		var ap *ActionPlan
		if ap, err = getCachedObject(apID); err != nil {
			err = fmt.Errorf("error <%s> querying the actionPlan: <%v>", err.Error(), apID)
			return
		}
		for acntID := range ap.AccountIDs {
			var ids []string
			if ids, err = getCachedIndex(acntID); err != nil { // read from cache but do not write if not there
				if err != utils.ErrNotFound {
					err = fmt.Errorf("error <%s> querying the accountActionPlan: <%v>", err.Error(), acntID)
					return
				}
				err = nil
				missingIndex[acntID] = append(missingIndex[acntID], apID)
				continue
			}
			if !utils.IsSliceMember(ids, apID) { // the index doesn't exits for this actionPlan
				missingIndex[acntID] = append(missingIndex[acntID], apID)
			}
		}
	}

	rply = &AccountActionPlanIHReply{
		MissingAccountActionPlans: missingIndex,
		BrokenReferences:          brokenRef,
	}
	return
}

type ReverseDestinationsIHReply struct {
	MissingReverseDestinations map[string][]string // list of missing indexes for each object (the map has the key as the indexKey and a list of objects)
	BrokenReferences           map[string][]string // list of broken references (the map has the key as the objectID and a list of indexes)
}

// add cache in args API
func GetReverseDestinationsIndexHealth(dm *DataManager, objLimit, indexLimit int, objTTL, indexTTL time.Duration, objStaticTTL, indexStaticTTL bool) (rply *ReverseDestinationsIHReply, err error) {
	// posible errors
	brokenRef := map[string][]string{}    // the actionPlans match the index but they are missing the account // broken reference
	missingIndex := map[string][]string{} // the indexes are not present but the action plans points to that account // misingAccounts

	// local cache
	indexesCache := ltcache.NewCache(objLimit, objTTL, objStaticTTL, nil)
	objectsCache := ltcache.NewCache(indexLimit, indexTTL, indexStaticTTL, nil)

	getCachedIndex := func(prefix string) (dstIDs []string, err error) {
		if x, ok := indexesCache.Get(prefix); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.([]string), nil
		}
		if dstIDs, err = dm.GetReverseDestination(prefix, true, utils.NonTransactional); err != nil { // read from cache but do not write if not there
			if err == utils.ErrNotFound {
				indexesCache.Set(prefix, nil, nil)
			}
			return
		}
		indexesCache.Set(prefix, dstIDs, nil)
		return
	}

	getCachedObject := func(dstID string) (obj *Destination, err error) {
		if x, ok := objectsCache.Get(dstID); ok {
			if x == nil {
				return nil, utils.ErrNotFound
			}
			return x.(*Destination), nil
		}
		if obj, err = dm.GetDestination(dstID, true, utils.NonTransactional); err != nil { // read from cache but do not write if not there
			if err == utils.ErrNotFound {
				objectsCache.Set(dstID, nil, nil)
			}
			return
		}
		objectsCache.Set(dstID, obj, nil)
		return
	}

	var prefixes []string // start with the indexes and check the references
	if prefixes, err = dm.DataDB().GetKeysForPrefix(utils.REVERSE_DESTINATION_PREFIX); err != nil {
		err = fmt.Errorf("error <%s> querying keys for reverseDestinations", err.Error())
		return
	}

	for _, prefix := range prefixes {
		prefix = strings.TrimPrefix(prefix, utils.REVERSE_DESTINATION_PREFIX) //
		var dstIDs []string
		if dstIDs, err = getCachedIndex(prefix); err != nil { // read from cache but do not write if not there
			err = fmt.Errorf("error <%s> querying the reverseDestination: <%v>", err.Error(), prefix)
			return
		}
		for _, dstID := range dstIDs {
			var dst *Destination
			if dst, err = getCachedObject(dstID); err != nil {
				if err != utils.ErrNotFound {
					err = fmt.Errorf("error <%s> querying the destination: <%v>", err.Error(), dstID)
					return
				}
				err = nil
				brokenRef[dstID] = nil
				continue
			}
			if !utils.IsSliceMember(dst.Prefixes, prefix) { // the action plan exists but doesn't point towards the account we have index
				brokenRef[dstID] = append(brokenRef[dstID], prefix)
			}
		}
	}

	var dstIDs []string // we have all the indexes in cache now do a reverse check
	if dstIDs, err = dm.DataDB().GetKeysForPrefix(utils.DESTINATION_PREFIX); err != nil {
		err = fmt.Errorf("error <%s> querying keys for destinations", err.Error())
		return
	}

	for _, dstID := range dstIDs {
		dstID = strings.TrimPrefix(dstID, utils.DESTINATION_PREFIX) //
		var dst *Destination
		if dst, err = getCachedObject(dstID); err != nil {
			err = fmt.Errorf("error <%s> querying the destination: <%v>", err.Error(), dstID)
			return
		}
		for _, prefix := range dst.Prefixes {
			var ids []string
			if ids, err = getCachedIndex(prefix); err != nil { // read from cache but do not write if not there
				if err != utils.ErrNotFound {
					err = fmt.Errorf("error <%s> querying the reverseDestination: <%v>", err.Error(), prefix)
					return
				}
				err = nil
				missingIndex[prefix] = append(missingIndex[prefix], dstID)
				continue
			}
			if !utils.IsSliceMember(ids, dstID) { // the index doesn't exits for this actionPlan
				missingIndex[prefix] = append(missingIndex[prefix], dstID)
			}
		}
	}

	rply = &ReverseDestinationsIHReply{
		MissingReverseDestinations: missingIndex,
		BrokenReferences:           brokenRef,
	}
	return
}

type FilterIHReply struct {
	MissingObjects []string            // list of object that are referenced in indexes but are not found in the dataDB
	MissingIndexes map[string][]string // list of missing indexes for each object (the map has the key as the objectID and a list of indexes)
	BrokenIndexes  map[string][]string // list of broken indexes for each object (the map has the key as the index and a list of objects)
	MissingFilters map[string][]string // list of broken references (the map has the key as the filterID and a list of  objectIDs)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getFiltersAndContexts returns the filtreIDs and context(if any) for that object
func getFiltersAndContexts(dm *DataManager, indxType, tnt, id string) (filterIDs []string, contexts *[]string, err error) { // add contexts
	switch indxType {
	case utils.CacheResourceFilterIndexes:
		var rs *ResourceProfile
		if rs, err = dm.GetResourceProfile(tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = rs.FilterIDs
	case utils.CacheStatFilterIndexes:
		var st *StatQueueProfile
		if st, err = dm.GetStatQueueProfile(tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = st.FilterIDs
	case utils.CacheThresholdFilterIndexes:
		var th *ThresholdProfile
		if th, err = dm.GetThresholdProfile(tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = th.FilterIDs
	case utils.CacheSupplierFilterIndexes:
		var rt *SupplierProfile
		if rt, err = dm.GetSupplierProfile(tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = rt.FilterIDs
	case utils.CacheAttributeFilterIndexes:
		var at *AttributeProfile
		if at, err = dm.GetAttributeProfile(tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = at.FilterIDs
		contexts = &at.Contexts
	case utils.CacheChargerFilterIndexes:
		var ch *ChargerProfile
		if ch, err = dm.GetChargerProfile(tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = ch.FilterIDs
	case utils.CacheDispatcherFilterIndexes:
		var ds *DispatcherProfile
		if ds, err = dm.GetDispatcherProfile(tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = ds.FilterIDs
		contexts = &ds.Subsystems
	default:
		return nil, nil, fmt.Errorf("unsuported index type:<%q>", indxType)
	}
	return
}

// objFIH keeps only the FilterIDs and Contexts from objects
type objFIH struct {
	filterIDs []string
	contexts  *[]string
}

// getIHObjFromCache returns all information that is needed from the mentioned object
// uses an extra cache(controled by the API) to optimize data management
func getIHObjFromCache(dm *DataManager, objCache *ltcache.Cache, indxType, tnt, id string) (obj *objFIH, err error) {
	cacheKey := utils.ConcatenatedKey(tnt, id)
	if objVal, ok := objCache.Get(cacheKey); ok {
		if objVal == nil {
			return nil, utils.ErrNotFound
		}
		return objVal.(*objFIH), nil
	}
	var filtIDs []string
	var contexts *[]string
	if filtIDs, contexts, err = getFiltersAndContexts(dm, indxType, tnt, id); err != nil {
		if err == utils.ErrNotFound {
			objCache.Set(cacheKey, nil, nil)
		}
		return
	}
	obj = &objFIH{
		filterIDs: filtIDs,
		contexts:  contexts,
	}
	objCache.Set(cacheKey, obj, nil)
	return
}

// getIHFltrFromCache returns the Filter
// uses an extra cache(controled by the API) to optimize data management
func getIHFltrFromCache(dm *DataManager, fltrCache *ltcache.Cache, tnt, id string) (fltr *Filter, err error) {
	cacheKey := utils.ConcatenatedKey(tnt, id)
	if fltrVal, ok := fltrCache.Get(cacheKey); ok {
		if fltrVal == nil {
			return nil, utils.ErrNotFound
		}
		return fltrVal.(*Filter), nil
	}
	if fltr, err = GetFilter(dm, tnt, id,
		true, false, utils.NonTransactional); err != nil {
		if err == utils.ErrNotFound {
			fltrCache.Set(cacheKey, nil, nil)
		}
		return
	}
	fltrCache.Set(cacheKey, fltr, nil)
	return
}

// getIHFltrIdxFromCache returns the Filter index
// uses an extra cache(controled by the API) to optimize data management
func getIHFltrIdxFromCache(dm *DataManager, fltrIdxCache *ltcache.Cache, idxItmType, tntCtx, fltrType, fldName, fldValue string) (idx utils.StringMap, err error) {
	idxKey := utils.ConcatenatedKey(fltrType, fldName, fldValue)
	cacheKey := utils.ConcatenatedKey(tntCtx, idxKey)
	if fltrVal, ok := fltrIdxCache.Get(cacheKey); ok {
		if fltrVal == nil {
			return nil, utils.ErrNotFound
		}
		return fltrVal.(utils.StringMap), nil
	}
	var indexes map[string]utils.StringMap
	if indexes, err = dm.GetFilterIndexes(idxItmType, tntCtx, fltrType, map[string]string{fldName: fldValue}); err != nil {
		if err == utils.ErrNotFound {
			fltrIdxCache.Set(cacheKey, nil, nil)
		}
		return
	}
	idx = indexes[idxKey]
	fltrIdxCache.Set(cacheKey, idx, nil)
	return
}

// getFilterAsIndexSet will parse the rules of filter and add them to the index map
func getFilterAsIndexSet(dm *DataManager, fltrIdxCache *ltcache.Cache, idxItmType, tntCtx string, fltr *Filter) (indexes map[string]utils.StringMap, err error) {
	indexes = make(map[string]utils.StringMap)
	for _, flt := range fltr.Rules {
		if flt.Type != utils.MetaString &&
			flt.Type != utils.MetaPrefix {
			continue
		}
		for _, fldVal := range flt.Values {
			var rcvIndx utils.StringMap
			// only read from cache in case if we do not find the index to not cache the negative response
			if rcvIndx, err = getIHFltrIdxFromCache(dm, fltrIdxCache, idxItmType, tntCtx, flt.Type, flt.Element, fldVal); err != nil {
				if err != utils.ErrNotFound {
					return
				}
				err = nil
				rcvIndx = make(utils.StringMap) // create an empty index if is not found in DB in case we add them later
			}
			indexes[utils.ConcatenatedKey(flt.Type, flt.Element, fldVal)] = rcvIndx
		}
	}
	return indexes, nil
}

// updateFilterIHMisingIndx updates the reply with the missing indexes for a specific object( obj->filter->index relation)
func updateFilterIHMisingIndx(dm *DataManager, fltrCache, fltrIdxCache *ltcache.Cache, filterIDs []string,
	indxType, tnt, tntCtx, itmID string, missingFltrs utils.StringSet, rply *FilterIHReply) (_ *FilterIHReply, err error) {
	if len(filterIDs) == 0 { // no filter so check the *none:*any:*any index
		var rcvIndx utils.StringMap
		if rcvIndx, err = getIHFltrIdxFromCache(dm, fltrIdxCache, indxType, tntCtx, utils.META_NONE, utils.META_ANY, utils.META_ANY); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			key := utils.ConcatenatedKey(tntCtx, utils.META_NONE, utils.META_ANY, utils.META_ANY)
			rply.MissingIndexes[key] = append(rply.MissingIndexes[key], itmID)
		} else if !rcvIndx.HasKey(itmID) {
			key := utils.ConcatenatedKey(tntCtx, utils.META_NONE, utils.META_ANY, utils.META_ANY)
			rply.MissingIndexes[key] = append(rply.MissingIndexes[key], itmID)
		}

		return rply, nil
	}
	for _, fltrID := range filterIDs { // parse all the filters
		var fltr *Filter
		if fltr, err = getIHFltrFromCache(dm, fltrCache, tnt, fltrID); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			if tntIdxFltr := utils.ConcatenatedKey(fltrID, itmID); !missingFltrs.Has(tntIdxFltr) { // tntIdxFltr = tnt:idx:id verification to not set the same ID
				missingFltrs.Add(tntIdxFltr)
				rply.MissingFilters[fltrID] = append(rply.MissingFilters[fltrID], itmID)
			}
			continue
		}
		var indexes map[string]utils.StringMap
		if indexes, err = getFilterAsIndexSet(dm, fltrIdxCache, indxType, tntCtx, fltr); err != nil { // build the index from filter
			return
		}
		for key, idx := range indexes { // check if the item is in the indexes
			if !idx.HasKey(itmID) {
				key = utils.ConcatenatedKey(tntCtx, key)
				rply.MissingIndexes[key] = append(rply.MissingIndexes[key], itmID)
			}
		}
	}
	return rply, nil
}

// GetFltrIdxHealth returns the missing indexes for all objects
func GetFltrIdxHealth(dm *DataManager, fltrCache, fltrIdxCache, objCache *ltcache.Cache, indxType string) (rply *FilterIHReply, err error) {
	// check the objects ( obj->filter->index relation)
	rply = &FilterIHReply{ // prepare the reply
		MissingIndexes: make(map[string][]string),
		BrokenIndexes:  make(map[string][]string),
		MissingFilters: make(map[string][]string),
	}
	objPrfx := utils.CacheIndexesToPrefix[indxType]
	var ids []string
	if ids, err = dm.dataDB.GetKeysForPrefix(objPrfx); err != nil {
		return
	}
	missingFltrs := utils.StringSet{} // for checking multiple filters that are missing(to not append the same ID in case)
	for _, id := range ids {          // get all the objects from DB
		id = strings.TrimPrefix(id, objPrfx)
		tntID := utils.NewTenantID(id)
		var obj *objFIH
		if obj, err = getIHObjFromCache(dm, objCache, indxType, tntID.Tenant, tntID.ID); err != nil {
			return
		}

		if obj.contexts == nil { // update the reply
			if rply, err = updateFilterIHMisingIndx(dm, fltrCache, fltrIdxCache, obj.filterIDs, indxType,
				tntID.Tenant, tntID.Tenant, tntID.ID, missingFltrs, rply); err != nil {
				return
			}
		} else {
			for _, ctx := range *obj.contexts {
				if rply, err = updateFilterIHMisingIndx(dm, fltrCache, fltrIdxCache, obj.filterIDs, indxType,
					tntID.Tenant, utils.ConcatenatedKey(tntID.Tenant, ctx), tntID.ID, missingFltrs, rply); err != nil {
					return
				}
			}
		}
	}

	// check the indexes( index->filter->obj relation)
	idxPrfx := utils.CacheInstanceToPrefix[indxType]
	var indexKeys []string
	if indexKeys, err = dm.dataDB.GetKeysForPrefix(idxPrfx); err != nil {
		return
	}
	missingObj := utils.StringSet{}
	for _, dataID := range indexKeys { // get all the indexes
		dataID = strings.TrimPrefix(dataID, idxPrfx)

		splt := utils.SplitConcatenatedKey(dataID) // tntCtx:filterType:fieldName:fieldVal
		lsplt := len(splt)
		if lsplt < 4 {
			err = fmt.Errorf("WRONG_IDX_KEY_FORMAT<%s>", dataID)
			return
		}
		tnt := splt[0]
		var ctx *string
		if lsplt-3 == 2 {
			ctx = &splt[1]
		}

		fieldVal := splt[lsplt-1]
		fieldName := splt[lsplt-2]
		filterType := splt[lsplt-3]

		tntCtx := utils.ConcatenatedKey(splt[:lsplt-3]...) // prefix may contain context/subsystems
		idxKey := utils.ConcatenatedKey(splt[lsplt-3:]...)

		var idx utils.StringMap
		if idx, err = getIHFltrIdxFromCache(dm, fltrIdxCache, indxType, tntCtx, filterType, fieldName, fieldVal); err != nil {
			return
		}
		for itmID := range idx {
			var obj *objFIH
			if obj, err = getIHObjFromCache(dm, objCache, indxType, tnt, itmID); err != nil {
				if err != utils.ErrNotFound {
					return
				}
				missingObj.Add(utils.ConcatenatedKey(tnt, itmID))
				//rply.MissingObjects = append(rply.MissingObjects, utils.ConcatenatedKey(tnt, itmID))
				err = nil
				continue
			}
			if ctx != nil &&
				(obj.contexts == nil || !utils.IsSliceMember(*obj.contexts, *ctx)) { // check the contexts if present
				key := utils.ConcatenatedKey(tntCtx, idxKey)
				rply.MissingIndexes[key] = append(rply.MissingIndexes[key], itmID)
				continue
			}
			if len(obj.filterIDs) == 0 { // check if the index is *none:*any:*any
				if utils.ConcatenatedKey(utils.META_NONE, utils.META_ANY, utils.META_ANY) != idxKey {
					rply.BrokenIndexes[dataID] = append(rply.BrokenIndexes[dataID], itmID)
				}
				continue
			}
			var hasIndx bool                       // just one filter needs to be the index
			for _, fltrID := range obj.filterIDs { // get the index for each filter from the object
				var fltr *Filter
				if fltr, err = getIHFltrFromCache(dm, fltrCache, tnt, fltrID); err != nil {
					if err != utils.ErrNotFound {
						return
					}
					err = nil // should be already logged when we parsed all the objects
					continue
				}
				var indexes map[string]utils.StringMap
				if indexes, err = getFilterAsIndexSet(dm, fltrIdxCache, indxType, tntCtx, fltr); err != nil {
					return
				}
				idx, has := indexes[idxKey]
				if hasIndx = has && idx.HasKey(itmID); hasIndx {
					break
				}
			}
			if !hasIndx {
				key := utils.ConcatenatedKey(tnt, idxKey)
				rply.BrokenIndexes[key] = append(rply.BrokenIndexes[key], itmID)
			}
		}
	}
	rply.MissingObjects = missingObj.AsSlice()
	return
}
