/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package v1

import (
	"strings"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

type AttrGetFilterIndexes struct {
	Tenant      string
	Context     string
	ItemType    string
	FilterType  string
	FilterField string
	FilterValue string
	utils.Paginator
}

type AttrRemFilterIndexes struct {
	Tenant   string
	Context  string
	ItemType string
}

func (api *APIerSv1) RemoveFilterIndexes(arg AttrRemFilterIndexes, reply *string) (err error) {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ItemType"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	key := arg.Tenant
	switch arg.ItemType {
	case utils.MetaThresholds:
		arg.ItemType = utils.ThresholdProfilePrefix
	case utils.MetaSuppliers:
		arg.ItemType = utils.SupplierProfilePrefix
	case utils.MetaStats:
		arg.ItemType = utils.StatQueueProfilePrefix
	case utils.MetaResources:
		arg.ItemType = utils.ResourceProfilesPrefix
	case utils.MetaChargers:
		arg.ItemType = utils.ChargerProfilePrefix
	case utils.MetaDispatchers:
		if missing := utils.MissingStructFields(&arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.DispatcherProfilePrefix
		key = utils.ConcatenatedKey(arg.Tenant, arg.Context)
	case utils.MetaAttributes:
		if missing := utils.MissingStructFields(&arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.AttributeProfilePrefix
		key = utils.ConcatenatedKey(arg.Tenant, arg.Context)
	}
	if err = api.DataManager.RemoveFilterIndexes(utils.PrefixToIndexCache[arg.ItemType], key); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (api *APIerSv1) GetFilterIndexes(arg AttrGetFilterIndexes, reply *[]string) (err error) {
	var indexes map[string]utils.StringMap
	var indexedSlice []string
	indexesFilter := make(map[string]utils.StringMap)
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ItemType"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	key := arg.Tenant
	switch arg.ItemType {
	case utils.MetaThresholds:
		arg.ItemType = utils.ThresholdProfilePrefix
	case utils.MetaSuppliers:
		arg.ItemType = utils.SupplierProfilePrefix
	case utils.MetaStats:
		arg.ItemType = utils.StatQueueProfilePrefix
	case utils.MetaResources:
		arg.ItemType = utils.ResourceProfilesPrefix
	case utils.MetaChargers:
		arg.ItemType = utils.ChargerProfilePrefix
	case utils.MetaDispatchers:
		if missing := utils.MissingStructFields(&arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.DispatcherProfilePrefix
		key = utils.ConcatenatedKey(arg.Tenant, arg.Context)
	case utils.MetaAttributes:
		if missing := utils.MissingStructFields(&arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.AttributeProfilePrefix
		key = utils.ConcatenatedKey(arg.Tenant, arg.Context)
	case utils.CacheReverseFilterIndexes:
		arg.ItemType = utils.ReverseFilterIndexes
	}
	if indexes, err = api.DataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[arg.ItemType], key, utils.EmptyString, nil); err != nil {
		return err
	}
	if arg.FilterType != "" {
		for val, strmap := range indexes {
			if strings.HasPrefix(val, arg.FilterType) {
				indexesFilter[val] = make(utils.StringMap)
				indexesFilter[val] = strmap
				for _, value := range strmap.Slice() {
					indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
				}
			}
		}
		if len(indexedSlice) == 0 {
			return utils.ErrNotFound
		}
	}
	if arg.FilterField != "" {
		if len(indexedSlice) == 0 {
			indexesFilter = make(map[string]utils.StringMap)
			for val, strmap := range indexes {
				if strings.Index(val, arg.FilterField) != -1 {
					indexesFilter[val] = make(utils.StringMap)
					indexesFilter[val] = strmap
					for _, value := range strmap.Slice() {
						indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(indexedSlice) == 0 {
				return utils.ErrNotFound
			}
		} else {
			var cloneIndexSlice []string
			for val, strmap := range indexesFilter {
				if strings.Index(val, arg.FilterField) != -1 {
					for _, value := range strmap.Slice() {
						cloneIndexSlice = append(cloneIndexSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(cloneIndexSlice) == 0 {
				return utils.ErrNotFound
			}
			indexedSlice = cloneIndexSlice
		}
	}
	if arg.FilterValue != "" {
		if len(indexedSlice) == 0 {
			for val, strmap := range indexes {
				if strings.Index(val, arg.FilterValue) != -1 {
					for _, value := range strmap.Slice() {
						indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(indexedSlice) == 0 {
				return utils.ErrNotFound
			}
		} else {
			var cloneIndexSlice []string
			for val, strmap := range indexesFilter {
				if strings.Index(val, arg.FilterValue) != -1 {
					for _, value := range strmap.Slice() {
						cloneIndexSlice = append(cloneIndexSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(cloneIndexSlice) == 0 {
				return utils.ErrNotFound
			}
			indexedSlice = cloneIndexSlice
		}
	}
	if len(indexedSlice) == 0 {
		for val, strmap := range indexes {
			for _, value := range strmap.Slice() {
				indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
			}
		}
	}
	if arg.Paginator.Limit != nil || arg.Paginator.Offset != nil {
		*reply = arg.Paginator.PaginateStringSlice(indexedSlice)
	} else {
		*reply = indexedSlice
	}
	return nil
}

// ComputeFilterIndexes selects which index filters to recompute
func (api *APIerSv1) ComputeFilterIndexes(args utils.ArgsComputeFilterIndexes, reply *string) (err error) {
	transactionID := utils.GenUUID()
	//ThresholdProfile Indexes
	var thdsIndexers *engine.FilterIndexer
	if args.ThresholdS {
		thdsIndexers, err = engine.ComputeThresholdIndexes(api.DataManager, args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//StatQueueProfile Indexes
	var sqpIndexers *engine.FilterIndexer
	if args.StatS {
		sqpIndexers, err = engine.ComputeStatIndexes(api.DataManager, args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//ResourceProfile Indexes
	var rsIndexes *engine.FilterIndexer
	if args.ResourceS {
		rsIndexes, err = engine.ComputeResourceIndexes(api.DataManager, args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//SupplierProfile Indexes
	var sppIndexes *engine.FilterIndexer
	if args.SupplierS {
		sppIndexes, err = engine.ComputeSupplierIndexes(api.DataManager, args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//AttributeProfile Indexes
	var attrIndexes *engine.FilterIndexer
	if args.AttributeS {
		attrIndexes, err = engine.ComputeAttributeIndexes(api.DataManager, args.Tenant, args.Context, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//ChargerProfile  Indexes
	var cppIndexes *engine.FilterIndexer
	if args.ChargerS {
		cppIndexes, err = engine.ComputeChargerIndexes(api.DataManager, args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//DispatcherProfile Indexes
	var dspIndexes *engine.FilterIndexer
	if args.DispatcherS {
		dspIndexes, err = engine.ComputeDispatcherIndexes(api.DataManager, args.Tenant, args.Context, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}

	//Now we move from tmpKey to the right key for each type
	//ThresholdProfile Indexes
	if thdsIndexers != nil {
		if err = thdsIndexers.StoreIndexes(true, transactionID); err != nil {
			return
		}
	}
	//StatQueueProfile Indexes
	if sqpIndexers != nil {
		if err = sqpIndexers.StoreIndexes(true, transactionID); err != nil {
			return
		}
	}
	//ResourceProfile Indexes
	if rsIndexes != nil {
		if err = rsIndexes.StoreIndexes(true, transactionID); err != nil {
			return
		}
	}
	//SupplierProfile Indexes
	if sppIndexes != nil {
		if err = sppIndexes.StoreIndexes(true, transactionID); err != nil {
			return
		}
	}
	//AttributeProfile Indexes
	if attrIndexes != nil {
		if err = attrIndexes.StoreIndexes(true, transactionID); err != nil {
			return
		}
	}
	//ChargerProfile Indexes
	if cppIndexes != nil {
		if err = cppIndexes.StoreIndexes(true, transactionID); err != nil {
			return
		}
	}
	//DispatcherProfile Indexes
	if dspIndexes != nil {
		if err = dspIndexes.StoreIndexes(true, transactionID); err != nil {
			return
		}
	}
	*reply = utils.OK
	return nil
}

// ComputeFilterIndexIDs computes specific filter indexes
func (api *APIerSv1) ComputeFilterIndexIDs(args utils.ArgsComputeFilterIndexIDs, reply *string) (err error) {
	transactionID := utils.GenUUID()
	//ThresholdProfile Indexes
	thdsIndexers, err := engine.ComputeThresholdIndexes(api.DataManager, args.Tenant, &args.ThresholdIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}

	//StatQueueProfile Indexes
	sqpIndexers, err := engine.ComputeStatIndexes(api.DataManager, args.Tenant, &args.StatIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//ResourceProfile Indexes
	rsIndexes, err := engine.ComputeResourceIndexes(api.DataManager, args.Tenant, &args.ResourceIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//SupplierProfile Indexes
	sppIndexes, err := engine.ComputeSupplierIndexes(api.DataManager, args.Tenant, &args.SupplierIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//AttributeProfile Indexes
	attrIndexes, err := engine.ComputeAttributeIndexes(api.DataManager, args.Tenant, args.Context, &args.AttributeIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//ChargerProfile  Indexes
	cppIndexes, err := engine.ComputeChargerIndexes(api.DataManager, args.Tenant, &args.ChargerIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//DispatcherProfile Indexes
	dspIndexes, err := engine.ComputeDispatcherIndexes(api.DataManager, args.Tenant, args.Context, &args.DispatcherIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}

	//Now we move from tmpKey to the right key for each type
	//ThresholdProfile Indexes
	if thdsIndexers != nil {
		if err = thdsIndexers.StoreIndexes(true, transactionID); err != nil {
			for _, id := range args.ThresholdIDs {
				var th *engine.ThresholdProfile
				if th, err = api.DataManager.GetThresholdProfile(args.Tenant, id, true, false, utils.NonTransactional); err != nil {
					return
				}
				if err = thdsIndexers.RemoveItemFromIndex(args.Tenant, id, th.FilterIDs); err != nil {
					return
				}
			}
			return
		}
	}
	//StatQueueProfile Indexes
	if sqpIndexers != nil {
		if err = sqpIndexers.StoreIndexes(true, transactionID); err != nil {
			for _, id := range args.StatIDs {
				var sqp *engine.StatQueueProfile
				if sqp, err = api.DataManager.GetStatQueueProfile(args.Tenant, id, true, false, utils.NonTransactional); err != nil {
					return
				}
				if err = sqpIndexers.RemoveItemFromIndex(args.Tenant, id, sqp.FilterIDs); err != nil {
					return
				}
			}
			return
		}
	}
	//ResourceProfile Indexes
	if rsIndexes != nil {
		if err = rsIndexes.StoreIndexes(true, transactionID); err != nil {
			for _, id := range args.ResourceIDs {
				var rp *engine.ResourceProfile
				if rp, err = api.DataManager.GetResourceProfile(args.Tenant, id, true, false, utils.NonTransactional); err != nil {
					return
				}
				if err = rsIndexes.RemoveItemFromIndex(args.Tenant, id, rp.FilterIDs); err != nil {
					return
				}
			}
			return
		}
	}
	//SupplierProfile Indexes
	if sppIndexes != nil {
		if err = sppIndexes.StoreIndexes(true, transactionID); err != nil {
			for _, id := range args.SupplierIDs {
				var spp *engine.SupplierProfile
				if spp, err = api.DataManager.GetSupplierProfile(args.Tenant, id, true, false, utils.NonTransactional); err != nil {
					return
				}
				if err = sppIndexes.RemoveItemFromIndex(args.Tenant, id, spp.FilterIDs); err != nil {
					return
				}
			}
			return
		}
	}
	//AttributeProfile Indexes
	if attrIndexes != nil {
		if err = attrIndexes.StoreIndexes(true, transactionID); err != nil {
			for _, id := range args.AttributeIDs {
				var ap *engine.AttributeProfile
				if ap, err = api.DataManager.GetAttributeProfile(args.Tenant, id, true, false, utils.NonTransactional); err != nil {
					return
				}
				if err = attrIndexes.RemoveItemFromIndex(args.Tenant, id, ap.FilterIDs); err != nil {
					return
				}
			}
			return
		}
	}
	//ChargerProfile Indexes
	if cppIndexes != nil {
		if err = cppIndexes.StoreIndexes(true, transactionID); err != nil {
			for _, id := range args.ChargerIDs {
				var cpp *engine.ChargerProfile
				if cpp, err = api.DataManager.GetChargerProfile(args.Tenant, id, true, false, utils.NonTransactional); err != nil {
					return
				}
				if err = cppIndexes.RemoveItemFromIndex(args.Tenant, id, cpp.FilterIDs); err != nil {
					return
				}
			}
			return
		}
	}
	//DispatcherProfile Indexes
	if dspIndexes != nil {
		if err = dspIndexes.StoreIndexes(true, transactionID); err != nil {
			for _, id := range args.DispatcherIDs {
				var dpp *engine.DispatcherProfile
				if dpp, err = api.DataManager.GetDispatcherProfile(args.Tenant, id, true, false, utils.NonTransactional); err != nil {
					return
				}
				if err = dspIndexes.RemoveItemFromIndex(args.Tenant, id, dpp.FilterIDs); err != nil {
					return
				}
			}
			return
		}
	}
	*reply = utils.OK
	return nil
}

func (apierSv1 *APIerSv1) GetAccountActionPlansIndexHealth(args *engine.IndexHealthArgsWith2Ch, reply *engine.AccountActionPlanIHReply) error {
	rp, err := engine.GetAccountActionPlansIndexHealth(apierSv1.DataManager, args.ObjectCacheLimit, args.IndexCacheLimit,
		args.ObjectCacheTTL, args.IndexCacheTTL,
		args.ObjectCacheStaticTTL, args.IndexCacheStaticTTL)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (apierSv1 *APIerSv1) GetReverseDestinationsIndexHealth(args *engine.IndexHealthArgsWith2Ch, reply *engine.ReverseDestinationsIHReply) error {
	rp, err := engine.GetReverseDestinationsIndexHealth(apierSv1.DataManager, args.ObjectCacheLimit, args.IndexCacheLimit,
		args.ObjectCacheTTL, args.IndexCacheTTL,
		args.ObjectCacheStaticTTL, args.IndexCacheStaticTTL)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (apierSv1 *APIerSv1) GetThresholdsIndexesHealth(args *engine.IndexHealthArgsWith3Ch, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(apierSv1.DataManager,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheThresholdFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (apierSv1 *APIerSv1) GetResourcesIndexesHealth(args *engine.IndexHealthArgsWith3Ch, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(apierSv1.DataManager,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheResourceFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (apierSv1 *APIerSv1) GetStatsIndexesHealth(args *engine.IndexHealthArgsWith3Ch, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(apierSv1.DataManager,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheStatFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (apierSv1 *APIerSv1) GetSuppliersIndexesHealth(args *engine.IndexHealthArgsWith3Ch, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(apierSv1.DataManager,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheSupplierFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (apierSv1 *APIerSv1) GetAttributesIndexesHealth(args *engine.IndexHealthArgsWith3Ch, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(apierSv1.DataManager,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheAttributeFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (apierSv1 *APIerSv1) GetChargersIndexesHealth(args *engine.IndexHealthArgsWith3Ch, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(apierSv1.DataManager,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheChargerFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}

func (apierSv1 *APIerSv1) GetDispatchersIndexesHealth(args *engine.IndexHealthArgsWith3Ch, reply *engine.FilterIHReply) error {
	rp, err := engine.GetFltrIdxHealth(apierSv1.DataManager,
		ltcache.NewCache(args.FilterCacheLimit, args.FilterCacheTTL, args.FilterCacheStaticTTL, nil),
		ltcache.NewCache(args.IndexCacheLimit, args.IndexCacheTTL, args.IndexCacheStaticTTL, nil),
		ltcache.NewCache(args.ObjectCacheLimit, args.ObjectCacheTTL, args.ObjectCacheStaticTTL, nil),
		utils.CacheDispatcherFilterIndexes,
	)
	if err != nil {
		return err
	}
	*reply = *rp
	return nil
}
