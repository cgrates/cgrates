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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

type IndexHealthArgs struct {
	Tenant              string
	IndexCacheLimit     int
	IndexCacheTTL       time.Duration
	IndexCacheStaticTTL bool

	ObjectCacheLimit     int
	ObjectCacheTTL       time.Duration
	ObjectCacheStaticTTL bool

	FilterCacheLimit     int
	FilterCacheTTL       time.Duration
	FilterCacheStaticTTL bool
	APIOpts              map[string]interface{} // Only for dispatcher.
}

type FilterIHReply struct {
	MissingObjects []string            // list of object that are referenced in indexes but are not found in the dataDB
	MissingIndexes map[string][]string // list of missing indexes for each object (the map has the key as the objectID and a list of indexes)
	BrokenIndexes  map[string][]string // list of broken indexes for each object (the map has the key as the index and a list of objects)
	MissingFilters map[string][]string // list of broken references (the map has the key as the filterID and a list of  objectIDs)
}

type ReverseFilterIHReply struct {
	MissingObjects        []string            // list of object that are referenced in indexes but are not found in the dataDB
	MissingReverseIndexes map[string][]string // list of missing indexes for each object (the map has the key as the objectID and a list of indexes)
	BrokenReverseIndexes  map[string][]string // list of broken indexes for each object (the map has the key as the objectID and a list of indexes)
	MissingFilters        map[string][]string // list of broken references (the map has the key as the filterID and a list of  objectIDs)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getFilters returns the filtreIDs and context(if any) for that object
func getFilters(ctx *context.Context, dm *DataManager, indxType, tnt, id string) (filterIDs []string, err error) { // add contexts
	switch indxType {
	case utils.CacheResourceFilterIndexes:
		var rs *ResourceProfile
		if rs, err = dm.GetResourceProfile(ctx, tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = rs.FilterIDs
	case utils.CacheStatFilterIndexes:
		var st *StatQueueProfile
		if st, err = dm.GetStatQueueProfile(ctx, tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = st.FilterIDs
	case utils.CacheThresholdFilterIndexes:
		var th *ThresholdProfile
		if th, err = dm.GetThresholdProfile(ctx, tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = th.FilterIDs
	case utils.CacheRouteFilterIndexes:
		var rt *RouteProfile
		if rt, err = dm.GetRouteProfile(ctx, tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = rt.FilterIDs
	case utils.CacheAttributeFilterIndexes:
		var at *AttributeProfile
		if at, err = dm.GetAttributeProfile(ctx, tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = at.FilterIDs
	case utils.CacheChargerFilterIndexes:
		var ch *ChargerProfile
		if ch, err = dm.GetChargerProfile(ctx, tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = ch.FilterIDs
	case utils.CacheDispatcherFilterIndexes:
		var ds *DispatcherProfile
		if ds, err = dm.GetDispatcherProfile(ctx, tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = ds.FilterIDs

	case utils.CacheRateProfilesFilterIndexes:
		var rp *utils.RateProfile
		if rp, err = dm.GetRateProfile(ctx, tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = rp.FilterIDs
	case utils.CacheActionProfilesFilterIndexes:
		var ap *ActionProfile
		if ap, err = dm.GetActionProfile(ctx, tnt, id, true, false, utils.NonTransactional); err != nil {
			return
		}
		filterIDs = ap.FilterIDs
	case utils.CacheAccountsFilterIndexes:
		var ac *utils.Account
		if ac, err = dm.GetAccount(ctx, tnt, id); err != nil {
			return
		}
		filterIDs = ac.FilterIDs
	default:
		return nil, fmt.Errorf("unsuported index type:<%q>", indxType)
	}
	if filterIDs == nil { // nil means ErrNotFound in cache
		filterIDs = make([]string, 0)
	}
	return
}

// getIHObjFromCache returns all information that is needed from the mentioned object
// uses an extra cache(controled by the API) to optimize data management
func getIHObjFromCache(ctx *context.Context, dm *DataManager, objCache *ltcache.Cache, indxType, tnt, id string) (filtIDs []string, err error) {
	cacheKey := utils.ConcatenatedKey(tnt, id)
	if objVal, ok := objCache.Get(cacheKey); ok {
		if objVal == nil {
			return nil, utils.ErrNotFound
		}
		return objVal.([]string), nil
	}
	if filtIDs, err = getFilters(ctx, dm, indxType, tnt, id); err != nil {
		if err == utils.ErrNotFound {
			objCache.Set(cacheKey, nil, nil)
		}
		return
	}
	objCache.Set(cacheKey, filtIDs, nil)
	return
}

// getIHFltrFromCache returns the Filter
// uses an extra cache(controled by the API) to optimize data management
func getIHFltrFromCache(ctx *context.Context, dm *DataManager, fltrCache *ltcache.Cache, tnt, id string) (fltr *Filter, err error) {
	cacheKey := utils.ConcatenatedKey(tnt, id)
	if fltrVal, ok := fltrCache.Get(cacheKey); ok {
		if fltrVal == nil {
			return nil, utils.ErrNotFound
		}
		return fltrVal.(*Filter), nil
	}
	if fltr, err = dm.GetFilter(ctx, tnt, id,
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
func getIHFltrIdxFromCache(ctx *context.Context, dm *DataManager, fltrIdxCache *ltcache.Cache, idxItmType, tntGrp, idxKey string) (idx utils.StringSet, err error) {
	cacheKey := utils.ConcatenatedKey(tntGrp, idxKey)
	if fltrVal, ok := fltrIdxCache.Get(cacheKey); ok {
		if fltrVal == nil {
			return nil, utils.ErrNotFound
		}
		return fltrVal.(utils.StringSet), nil
	}
	var indexes map[string]utils.StringSet
	if indexes, err = dm.GetIndexes(ctx, idxItmType, tntGrp, idxKey, true, false); err != nil {
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
func getFilterAsIndexSet(ctx *context.Context, dm *DataManager, fltrIdxCache *ltcache.Cache, idxItmType, tntGrp string, fltr *Filter) (indexes map[string]utils.StringSet, err error) {
	indexes = make(map[string]utils.StringSet)
	for _, flt := range fltr.Rules {
		if !FilterIndexTypes.Has(flt.Type) ||
			IsDynamicDPPath(flt.Element) {
			continue
		}
		isDyn := strings.HasPrefix(flt.Element, utils.DynamicDataPrefix)
		for _, fldVal := range flt.Values {
			if IsDynamicDPPath(fldVal) {
				continue
			}
			var idxKey string
			if isDyn {
				if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) { // do not index if both the element and the value is dynamic
					continue
				}
				idxKey = utils.ConcatenatedKey(flt.Type, flt.Element[1:], fldVal)
			} else if strings.HasPrefix(fldVal, utils.DynamicDataPrefix) {
				idxKey = utils.ConcatenatedKey(flt.Type, fldVal[1:], flt.Element)
			} else {
				// do not index not dynamic filters
				continue
			}
			var rcvIndx utils.StringSet
			// only read from cache in case if we do not find the index to not cache the negative response
			if rcvIndx, err = getIHFltrIdxFromCache(ctx, dm, fltrIdxCache, idxItmType, tntGrp, idxKey); err != nil {
				if err != utils.ErrNotFound {
					return
				}
				err = nil
				rcvIndx = make(utils.StringSet) // create an empty index if is not found in DB in case we add them later
			}
			indexes[idxKey] = rcvIndx
		}
	}
	return indexes, nil
}

// updateFilterIHMisingIndx updates the reply with the missing indexes for a specific object( obj->filter->index relation)
func updateFilterIHMisingIndx(ctx *context.Context, dm *DataManager, fltrCache, fltrIdxCache *ltcache.Cache, filterIDs []string, indxType, tnt, tntGrp, itmID string, rply *FilterIHReply) (_ *FilterIHReply, err error) {
	if len(filterIDs) == 0 { // no filter so check the *none:*any:*any index
		idxKey := utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny)
		var rcvIndx utils.StringSet
		if rcvIndx, err = getIHFltrIdxFromCache(ctx, dm, fltrIdxCache, indxType, tntGrp, idxKey); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			key := utils.ConcatenatedKey(tntGrp, idxKey)
			rply.MissingIndexes[key] = append(rply.MissingIndexes[key], itmID)
		} else if !rcvIndx.Has(itmID) {
			key := utils.ConcatenatedKey(tntGrp, idxKey)
			rply.MissingIndexes[key] = append(rply.MissingIndexes[key], itmID)
		}

		return rply, nil
	}
	for _, fltrID := range filterIDs { // parse all the filters
		var fltr *Filter
		if fltr, err = getIHFltrFromCache(ctx, dm, fltrCache, tnt, fltrID); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			fltrID = utils.ConcatenatedKey(tnt, fltrID)
			rply.MissingFilters[fltrID] = append(rply.MissingFilters[fltrID], itmID)
			continue
		}
		var indexes map[string]utils.StringSet
		if indexes, err = getFilterAsIndexSet(ctx, dm, fltrIdxCache, indxType, tntGrp, fltr); err != nil { // build the index from filter
			return
		}
		for key, idx := range indexes { // check if the item is in the indexes
			if !idx.Has(itmID) {
				key = utils.ConcatenatedKey(tntGrp, key)
				rply.MissingIndexes[key] = append(rply.MissingIndexes[key], itmID)
			}
		}
	}
	return rply, nil
}

// GetFltrIdxHealth returns the missing indexes for all objects
func GetFltrIdxHealth(ctx *context.Context, dm *DataManager, fltrCache, fltrIdxCache, objCache *ltcache.Cache, indxType string) (rply *FilterIHReply, err error) {
	// check the objects ( obj->filter->index relation)
	rply = &FilterIHReply{ // prepare the reply
		MissingIndexes: make(map[string][]string),
		BrokenIndexes:  make(map[string][]string),
		MissingFilters: make(map[string][]string),
	}
	objPrfx := utils.CacheIndexesToPrefix[indxType]
	var ids []string
	if ids, err = dm.dataDB.GetKeysForPrefix(ctx, objPrfx); err != nil {
		return
	}
	for _, id := range ids { // get all the objects from DB
		id = strings.TrimPrefix(id, objPrfx)
		tntID := utils.NewTenantID(id)
		var filterIDs []string
		if filterIDs, err = getIHObjFromCache(ctx, dm, objCache, indxType, tntID.Tenant, tntID.ID); err != nil {
			return
		}

		if rply, err = updateFilterIHMisingIndx(ctx, dm, fltrCache, fltrIdxCache, filterIDs, indxType, tntID.Tenant, tntID.Tenant, tntID.ID, rply); err != nil { // update the reply
			return
		}
	}

	// check the indexes( index->filter->obj relation)
	idxPrfx := utils.CacheInstanceToPrefix[indxType]
	var indexKeys []string
	if indexKeys, err = dm.dataDB.GetKeysForPrefix(ctx, idxPrfx); err != nil {
		return
	}
	missingObj := utils.StringSet{}
	for _, dataID := range indexKeys { // get all the indexes
		dataID = strings.TrimPrefix(dataID, idxPrfx)

		splt := utils.SplitConcatenatedKey(dataID) // tntGrp:filterType:fieldName:fieldVal
		lsplt := len(splt)
		if lsplt < 4 {
			err = fmt.Errorf("WRONG_IDX_KEY_FORMAT<%s>", dataID)
			return
		}
		tnt := utils.ConcatenatedKey(splt[:lsplt-3]...) // prefix may contain context/subsystems
		idxKey := utils.ConcatenatedKey(splt[lsplt-3:]...)

		var idx utils.StringSet
		if idx, err = getIHFltrIdxFromCache(ctx, dm, fltrIdxCache, indxType, tnt, idxKey); err != nil {
			return
		}
		for itmID := range idx {
			var filterIDs []string
			if filterIDs, err = getIHObjFromCache(ctx, dm, objCache, indxType, tnt, itmID); err != nil {
				if err != utils.ErrNotFound {
					return
				}
				missingObj.Add(utils.ConcatenatedKey(tnt, itmID))
				err = nil
				continue
			}
			if len(filterIDs) == 0 { // check if the index is *none:*any:*any
				if utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny) != idxKey {
					rply.BrokenIndexes[dataID] = append(rply.BrokenIndexes[dataID], itmID)
				}
				continue
			}
			var hasIndx bool                   // just one filter needs to be the index
			for _, fltrID := range filterIDs { // get the index for each filter from the object
				var fltr *Filter
				if fltr, err = getIHFltrFromCache(ctx, dm, fltrCache, tnt, fltrID); err != nil {
					if err != utils.ErrNotFound {
						return
					}
					err = nil // should be already logged when we parsed all the objects
					continue
				}
				var indexes map[string]utils.StringSet
				if indexes, err = getFilterAsIndexSet(ctx, dm, fltrIdxCache, indxType, tnt, fltr); err != nil {
					return
				}
				idx, has := indexes[idxKey]
				if hasIndx = has && idx.Has(itmID); hasIndx {
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

// getRevFltrIdxHealthFromObj returns the missing reverse indexes for all objects of the given type
func getRevFltrIdxHealthFromObj(ctx *context.Context, dm *DataManager, fltrCache, revFltrIdxCache, objCache *ltcache.Cache, indxType string) (rply *ReverseFilterIHReply, err error) {
	// check the objects ( obj->filter->index relation)
	rply = &ReverseFilterIHReply{ // prepare the reply
		MissingReverseIndexes: make(map[string][]string),
		BrokenReverseIndexes:  make(map[string][]string),
		MissingFilters:        make(map[string][]string),
	}
	objPrfx := utils.CacheIndexesToPrefix[indxType]
	var ids []string
	if ids, err = dm.dataDB.GetKeysForPrefix(ctx, objPrfx); err != nil {
		return
	}
	for _, id := range ids { // get all the objects
		id = strings.TrimPrefix(id, objPrfx)
		tntID := utils.NewTenantID(id)
		var filterIDs []string
		if filterIDs, err = getIHObjFromCache(ctx, dm, objCache, indxType, tntID.Tenant, tntID.ID); err != nil {
			return
		}

		for _, fltrID := range filterIDs {
			if strings.HasPrefix(fltrID, utils.Meta) {
				continue
			}
			if _, err = getIHFltrFromCache(ctx, dm, fltrCache, tntID.Tenant, fltrID); err != nil { // check if the filter exists
				if err != utils.ErrNotFound {
					return
				}
				err = nil
				key := utils.ConcatenatedKey(tntID.Tenant, fltrID)
				rply.MissingFilters[key] = append(rply.MissingFilters[key], tntID.ID)
				continue
			}
			var revIdx utils.StringSet
			if revIdx, err = getIHFltrIdxFromCache(ctx, dm, revFltrIdxCache, utils.CacheReverseFilterIndexes, utils.ConcatenatedKey(tntID.Tenant, fltrID), indxType); err != nil { // check the reverese index
				if err == utils.ErrNotFound {
					rply.MissingReverseIndexes[id] = append(rply.MissingReverseIndexes[id], fltrID)
					err = nil
					continue
				}
				return
			}
			if !revIdx.Has(tntID.ID) {
				rply.MissingReverseIndexes[id] = append(rply.MissingReverseIndexes[id], fltrID)
			}
		}
	}
	return
}

// getRevFltrIdxHealthFromReverse parses the reverse indexes and updates the reply
func getRevFltrIdxHealthFromReverse(ctx *context.Context, dm *DataManager, fltrCache, revFltrIdxCache *ltcache.Cache, objCaches map[string]*ltcache.Cache, rply map[string]*ReverseFilterIHReply) (_ map[string]*ReverseFilterIHReply, err error) {
	var revIndexKeys []string
	if revIndexKeys, err = dm.dataDB.GetKeysForPrefix(ctx, utils.FilterIndexPrfx); err != nil {
		return
	}
	missingObj := utils.StringSet{}
	for _, revIdxKey := range revIndexKeys { // parse all the reverse indexes
		// compose the needed information from key
		revIdxKey = strings.TrimPrefix(revIdxKey, utils.FilterIndexPrfx)
		revIDxSplit := strings.SplitN(revIdxKey, utils.ConcatenatedKeySep, 3)
		tnt, fltrID, indxType := revIDxSplit[0], revIDxSplit[1], revIDxSplit[2]
		revIdxKey = utils.ConcatenatedKey(tnt, fltrID)
		objCache := objCaches[indxType]

		if _, has := rply[indxType]; !has { // make sure that the reply has the type in map
			rply[indxType] = &ReverseFilterIHReply{
				MissingReverseIndexes: make(map[string][]string),
				MissingFilters:        make(map[string][]string),
				BrokenReverseIndexes:  make(map[string][]string),
			}
		}

		var revIdx utils.StringSet
		if revIdx, err = getIHFltrIdxFromCache(ctx, dm, revFltrIdxCache, utils.CacheReverseFilterIndexes, revIdxKey, indxType); err != nil {
			return
		}
		for id := range revIdx {
			var filterIDs []string
			if indxType == utils.CacheRateFilterIndexes {
				spl := strings.SplitN(id, utils.ConcatenatedKeySep, 2)
				rateID := spl[0]
				rprfID := spl[1]
				var rates map[string]*utils.Rate
				if rates, err = getRatesFromCache(ctx, dm, objCache, tnt, rprfID); err != nil {
					if err != utils.ErrNotFound {
						return
					}
					missingObj.Add(utils.ConcatenatedKey(tnt, id))
					rply[indxType].MissingObjects = missingObj.AsSlice()
					err = nil
					continue
				}
				if rate, has := rates[rateID]; !has {
					missingObj.Add(utils.ConcatenatedKey(tnt, id))
					rply[indxType].MissingObjects = missingObj.AsSlice()
					continue
				} else {
					filterIDs = rate.FilterIDs
				}
			} else if filterIDs, err = getIHObjFromCache(ctx, dm, objCache, indxType, tnt, id); err != nil {
				if err == utils.ErrNotFound {
					missingObj.Add(utils.ConcatenatedKey(tnt, id))
					rply[indxType].MissingObjects = missingObj.AsSlice()
					err = nil
					continue
				}
				return
			}
			if !utils.IsSliceMember(filterIDs, fltrID) { // check the filters
				key := utils.ConcatenatedKey(tnt, id)
				rply[indxType].BrokenReverseIndexes[key] = append(rply[indxType].BrokenReverseIndexes[key], fltrID)
			}
		}
	}
	return rply, nil
}

// GetRevFltrIdxHealth will return all the broken indexes
func GetRevFltrIdxHealth(ctx *context.Context, dm *DataManager, fltrCache, revFltrIdxCache *ltcache.Cache, objCaches map[string]*ltcache.Cache) (rply map[string]*ReverseFilterIHReply, err error) {
	rply = make(map[string]*ReverseFilterIHReply)
	for indxType := range utils.CacheIndexesToPrefix { // parse all posible filter indexes
		if indxType == utils.CacheReverseFilterIndexes { // ommit the reverse indexes
			continue
		}
		if rply[indxType], err = getRevFltrIdxHealthFromObj(ctx, dm, fltrCache, revFltrIdxCache, objCaches[indxType], indxType); err != nil {
			return
		}
	}
	if rply[utils.CacheRateFilterIndexes], err = getRevFltrIdxHealthFromRateRates(ctx, dm, fltrCache, revFltrIdxCache, objCaches[utils.CacheRateFilterIndexes]); err != nil {
		return
	}
	rply, err = getRevFltrIdxHealthFromReverse(ctx, dm, fltrCache, revFltrIdxCache, objCaches, rply)
	for k, v := range rply { // should be a safe for (even on rply==nil)
		if len(v.MissingFilters) == 0 && // remove nonpopulated objects
			len(v.MissingObjects) == 0 &&
			len(v.BrokenReverseIndexes) == 0 &&
			len(v.MissingReverseIndexes) == 0 {
			delete(rply, k)
		}
	}
	return
}

// getRatesFromCache returns all rates from rateprofile
// uses an extra cache(controled by the API) to optimize data management
func getRatesFromCache(ctx *context.Context, dm *DataManager, objCache *ltcache.Cache, tnt, rprfID string) (_ map[string]*utils.Rate, err error) {
	cacheKey := utils.ConcatenatedKey(tnt, rprfID)
	if objVal, ok := objCache.Get(cacheKey); ok {
		if objVal == nil {
			return nil, utils.ErrNotFound
		}
		rprf := objVal.(*utils.RateProfile)
		return rprf.Rates, nil
	}

	var rprf *utils.RateProfile
	if rprf, err = dm.GetRateProfile(ctx, tnt, rprfID, true, false, utils.NonTransactional); err != nil {
		if err == utils.ErrNotFound {
			objCache.Set(cacheKey, nil, nil)
		}
		return
	}
	objCache.Set(cacheKey, rprf, nil)
	return rprf.Rates, nil
}

// GetFltrIdxHealth returns the missing indexes for all objects
func GetFltrIdxHealthForRateRates(ctx *context.Context, dm *DataManager, fltrCache, fltrIdxCache, objCache *ltcache.Cache) (rply *FilterIHReply, err error) {
	// check the objects ( obj->filter->index relation)
	rply = &FilterIHReply{
		MissingIndexes: make(map[string][]string),
		BrokenIndexes:  make(map[string][]string),
		MissingFilters: make(map[string][]string),
	}
	var ids []string
	if ids, err = dm.dataDB.GetKeysForPrefix(ctx, utils.RateProfilePrefix); err != nil {
		return
	}
	for _, id := range ids {
		id = strings.TrimPrefix(id, utils.RateProfilePrefix)
		tntID := utils.NewTenantID(id)

		var rates map[string]*utils.Rate
		if rates, err = getRatesFromCache(ctx, dm, objCache, tntID.Tenant, tntID.ID); err != nil {
			return
		}
		for rtID, rate := range rates {
			if rply, err = updateFilterIHMisingIndx(ctx, dm, fltrCache, fltrIdxCache, rate.FilterIDs, utils.CacheRateFilterIndexes, tntID.Tenant, utils.ConcatenatedKey(tntID.Tenant, tntID.ID), rtID, rply); err != nil {
				return
			}
		}
	}

	// check the indexes( index->filter->obj relation)
	var indexKeys []string
	if indexKeys, err = dm.dataDB.GetKeysForPrefix(ctx, utils.RateFilterIndexPrfx); err != nil {
		return
	}
	for _, dataID := range indexKeys {
		dataID = strings.TrimPrefix(dataID, utils.RateFilterIndexPrfx)

		splt := utils.SplitConcatenatedKey(dataID) // tntGrp:filterType:fieldName:fieldVal
		lsplt := len(splt)
		if lsplt < 4 {
			err = fmt.Errorf("WRONG_IDX_KEY_FORMAT<%s>", dataID)
			return
		}
		tnt := splt[0]
		rpID := splt[1]
		tntGrp := utils.ConcatenatedKey(splt[:lsplt-3]...) // prefix may contain context/subsystems
		idxKey := utils.ConcatenatedKey(splt[lsplt-3:]...)

		var idx utils.StringSet
		if idx, err = getIHFltrIdxFromCache(ctx, dm, fltrIdxCache, utils.CacheRateFilterIndexes, tntGrp, idxKey); err != nil {
			return
		}
		for itmID := range idx {
			var rates map[string]*utils.Rate
			if rates, err = getRatesFromCache(ctx, dm, objCache, tnt, rpID); err != nil {
				if err != utils.ErrNotFound {
					return
				}
				rply.MissingObjects = append(rply.MissingObjects, utils.ConcatenatedKey(tntGrp, itmID))
				err = nil
				continue
			}
			var filterIDs []string
			if rate, has := rates[itmID]; !has {
				rply.MissingObjects = append(rply.MissingObjects, utils.ConcatenatedKey(tntGrp, itmID))
				continue
			} else {
				filterIDs = rate.FilterIDs
			}
			if len(filterIDs) == 0 {
				if utils.ConcatenatedKey(utils.MetaNone, utils.MetaAny, utils.MetaAny) != idxKey {
					rply.BrokenIndexes[dataID] = append(rply.BrokenIndexes[dataID], itmID)
				}
				continue
			}
			var hasIndx bool
			for _, fltrID := range filterIDs {
				var fltr *Filter
				if fltr, err = getIHFltrFromCache(ctx, dm, fltrCache, tnt, fltrID); err != nil {
					if err != utils.ErrNotFound {
						return
					}
					err = nil // should be already logged when we parsed all the objects
					continue
				}
				var indexes map[string]utils.StringSet
				if indexes, err = getFilterAsIndexSet(ctx, dm, fltrIdxCache, utils.CacheRateFilterIndexes, tntGrp, fltr); err != nil {
					return
				}
				idx, has := indexes[idxKey]
				if hasIndx = has && idx.Has(itmID); hasIndx {
					break
				}
			}
			if !hasIndx {
				key := utils.ConcatenatedKey(tnt, idxKey)
				rply.BrokenIndexes[key] = append(rply.BrokenIndexes[key], itmID)
			}
		}
	}

	return
}

func getRevFltrIdxHealthFromRateRates(ctx *context.Context, dm *DataManager, fltrCache, revFltrIdxCache, objCache *ltcache.Cache) (rply *ReverseFilterIHReply, err error) {
	// check the objects ( obj->filter->index relation)
	rply = &ReverseFilterIHReply{
		MissingReverseIndexes: make(map[string][]string),
		BrokenReverseIndexes:  make(map[string][]string),
		MissingFilters:        make(map[string][]string),
	}
	var ids []string
	if ids, err = dm.dataDB.GetKeysForPrefix(ctx, utils.RateProfilePrefix); err != nil {
		return
	}
	for _, id := range ids {
		id = strings.TrimPrefix(id, utils.RateProfilePrefix)
		tntID := utils.NewTenantID(id)
		var rates map[string]*utils.Rate
		if rates, err = getRatesFromCache(ctx, dm, objCache, tntID.Tenant, tntID.ID); err != nil {
			return
		}
		for rtID, rate := range rates {
			itmID := utils.ConcatenatedKey(rtID, tntID.ID)
			itmIDWithTnt := utils.ConcatenatedKey(id, rtID)

			for _, fltrID := range rate.FilterIDs {
				if strings.HasPrefix(fltrID, utils.Meta) {
					continue
				}
				if _, err = getIHFltrFromCache(ctx, dm, fltrCache, tntID.Tenant, fltrID); err != nil {
					if err != utils.ErrNotFound {
						return
					}
					err = nil
					key := utils.ConcatenatedKey(tntID.Tenant, fltrID)
					rply.MissingFilters[key] = append(rply.MissingFilters[key], itmID)
					continue
				}
				var revIdx utils.StringSet
				if revIdx, err = getIHFltrIdxFromCache(ctx, dm, revFltrIdxCache, utils.CacheReverseFilterIndexes, utils.ConcatenatedKey(tntID.Tenant, fltrID), utils.CacheRateFilterIndexes); err != nil {
					if err == utils.ErrNotFound {
						rply.MissingReverseIndexes[itmIDWithTnt] = append(rply.MissingReverseIndexes[itmIDWithTnt], fltrID)
						err = nil
						continue
					}
					return
				}
				if !revIdx.Has(itmID) {
					rply.MissingReverseIndexes[itmIDWithTnt] = append(rply.MissingReverseIndexes[itmIDWithTnt], fltrID)
				}
			}
		}
	}
	return
}
