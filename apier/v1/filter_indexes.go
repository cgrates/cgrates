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

package v1

import (
	"fmt"
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
	}
	if indexes, err = api.DataManager.GetFilterIndexes(
		utils.PrefixToIndexCache[arg.ItemType], key, "", nil); err != nil {
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
		thdsIndexers, err = api.computeThresholdIndexes(args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//StatQueueProfile Indexes
	var sqpIndexers *engine.FilterIndexer
	if args.StatS {
		sqpIndexers, err = api.computeStatIndexes(args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//ResourceProfile Indexes
	var rsIndexes *engine.FilterIndexer
	if args.ResourceS {
		rsIndexes, err = api.computeResourceIndexes(args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//SupplierProfile Indexes
	var sppIndexes *engine.FilterIndexer
	if args.SupplierS {
		sppIndexes, err = api.computeSupplierIndexes(args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//AttributeProfile Indexes
	var attrIndexes *engine.FilterIndexer
	if args.AttributeS {
		attrIndexes, err = api.computeAttributeIndexes(args.Tenant, args.Context, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//ChargerProfile  Indexes
	var cppIndexes *engine.FilterIndexer
	if args.ChargerS {
		cppIndexes, err = api.computeChargerIndexes(args.Tenant, nil, transactionID)
		if err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	}
	//DispatcherProfile Indexes
	var dspIndexes *engine.FilterIndexer
	if args.DispatcherS {
		dspIndexes, err = api.computeDispatcherIndexes(args.Tenant, args.Context, nil, transactionID)
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
	thdsIndexers, err := api.computeThresholdIndexes(args.Tenant, &args.ThresholdIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//StatQueueProfile Indexes
	sqpIndexers, err := api.computeStatIndexes(args.Tenant, &args.StatIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//ResourceProfile Indexes
	rsIndexes, err := api.computeResourceIndexes(args.Tenant, &args.ResourceIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//SupplierProfile Indexes
	sppIndexes, err := api.computeSupplierIndexes(args.Tenant, &args.SupplierIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//AttributeProfile Indexes
	attrIndexes, err := api.computeAttributeIndexes(args.Tenant, args.Context, &args.AttributeIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//ChargerProfile  Indexes
	cppIndexes, err := api.computeChargerIndexes(args.Tenant, &args.ChargerIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//DispatcherProfile Indexes
	dspIndexes, err := api.computeDispatcherIndexes(args.Tenant, args.Context, &args.DispatcherIDs, transactionID)
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

func (api *APIerSv1) computeThresholdIndexes(tenant string, thIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var thresholdIDs []string
	thdsIndexers := engine.NewFilterIndexer(api.DataManager, utils.ThresholdProfilePrefix, tenant)
	if thIDs == nil {
		ids, err := api.DataManager.DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			thresholdIDs = append(thresholdIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		thresholdIDs = *thIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range thresholdIDs {
		th, err := api.DataManager.GetThresholdProfile(tenant, id, true, false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		fltrIDs := make([]string, len(th.FilterIDs))
		for i, fltrID := range th.FilterIDs {
			fltrIDs[i] = fltrID
		}
		if len(fltrIDs) == 0 {
			fltrIDs = []string{utils.META_NONE}
		}
		for _, fltrID := range fltrIDs {
			var fltr *engine.Filter
			if fltrID == utils.META_NONE {
				fltr = &engine.Filter{
					Tenant: th.Tenant,
					ID:     th.ID,
					Rules: []*engine.FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = engine.GetFilter(api.DataManager, th.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for threshold: %+v",
						fltrID, th)
				}
				return nil, err
			}
			thdsIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), th.ID)
		}
	}
	if transactionID == utils.NonTransactional {
		if err := thdsIndexers.StoreIndexes(true, transactionID); err != nil {
			return nil, err
		}
		return nil, nil
	} else {
		if err := thdsIndexers.StoreIndexes(false, transactionID); err != nil {
			return nil, err
		}
	}
	return thdsIndexers, nil
}

func (api *APIerSv1) computeAttributeIndexes(tenant, context string, attrIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var attributeIDs []string
	attrIndexers := engine.NewFilterIndexer(api.DataManager, utils.AttributeProfilePrefix,
		utils.ConcatenatedKey(tenant, context))
	if attrIDs == nil {
		ids, err := api.DataManager.DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			attributeIDs = append(attributeIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		attributeIDs = *attrIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range attributeIDs {
		ap, err := api.DataManager.GetAttributeProfile(tenant, id, true, false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		if !utils.IsSliceMember(ap.Contexts, context) && context != utils.META_ANY {
			continue
		}
		fltrIDs := make([]string, len(ap.FilterIDs))
		for i, fltrID := range ap.FilterIDs {
			fltrIDs[i] = fltrID
		}
		if len(fltrIDs) == 0 {
			fltrIDs = []string{utils.META_NONE}
		}
		for _, fltrID := range fltrIDs {
			var fltr *engine.Filter
			if fltrID == utils.META_NONE {
				fltr = &engine.Filter{
					Tenant: ap.Tenant,
					ID:     ap.ID,
					Rules: []*engine.FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = engine.GetFilter(api.DataManager, ap.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for attribute: %+v",
						fltrID, ap)
				}
				return nil, err
			}
			attrIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), ap.ID)
		}
	}
	if transactionID == utils.NonTransactional {
		if err := attrIndexers.StoreIndexes(true, transactionID); err != nil {
			return nil, err
		}
		return nil, nil
	} else {
		if err := attrIndexers.StoreIndexes(false, transactionID); err != nil {
			return nil, err
		}
	}
	return attrIndexers, nil
}

func (api *APIerSv1) computeResourceIndexes(tenant string, rsIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var resourceIDs []string
	rpIndexers := engine.NewFilterIndexer(api.DataManager, utils.ResourceProfilesPrefix, tenant)
	if rsIDs == nil {
		ids, err := api.DataManager.DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			resourceIDs = append(resourceIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		resourceIDs = *rsIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range resourceIDs {
		rp, err := api.DataManager.GetResourceProfile(tenant, id, true, false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		fltrIDs := make([]string, len(rp.FilterIDs))
		for i, fltrID := range rp.FilterIDs {
			fltrIDs[i] = fltrID
		}
		if len(fltrIDs) == 0 {
			fltrIDs = []string{utils.META_NONE}
		}
		for _, fltrID := range fltrIDs {
			var fltr *engine.Filter
			if fltrID == utils.META_NONE {
				fltr = &engine.Filter{
					Tenant: rp.Tenant,
					ID:     rp.ID,
					Rules: []*engine.FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = engine.GetFilter(api.DataManager, rp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for resource: %+v",
						fltrID, rp)
				}
				return nil, err
			}
			rpIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), rp.ID)
		}
	}
	if transactionID == utils.NonTransactional {
		if err := rpIndexers.StoreIndexes(true, transactionID); err != nil {
			return nil, err
		}
		return nil, nil
	} else {
		if err := rpIndexers.StoreIndexes(false, transactionID); err != nil {
			return nil, err
		}
	}
	return rpIndexers, nil
}

func (api *APIerSv1) computeStatIndexes(tenant string, stIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var statIDs []string
	sqpIndexers := engine.NewFilterIndexer(api.DataManager, utils.StatQueueProfilePrefix, tenant)
	if stIDs == nil {
		ids, err := api.DataManager.DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			statIDs = append(statIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		statIDs = *stIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range statIDs {
		sqp, err := api.DataManager.GetStatQueueProfile(tenant, id, true, false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		fltrIDs := make([]string, len(sqp.FilterIDs))
		for i, fltrID := range sqp.FilterIDs {
			fltrIDs[i] = fltrID
		}
		if len(fltrIDs) == 0 {
			fltrIDs = []string{utils.META_NONE}
		}
		for _, fltrID := range fltrIDs {
			var fltr *engine.Filter
			if fltrID == utils.META_NONE {
				fltr = &engine.Filter{
					Tenant: sqp.Tenant,
					ID:     sqp.ID,
					Rules: []*engine.FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = engine.GetFilter(api.DataManager, sqp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for statqueue: %+v",
						fltrID, sqp)
				}
				return nil, err
			}
			sqpIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), sqp.ID)
		}
	}
	if transactionID == utils.NonTransactional {
		if err := sqpIndexers.StoreIndexes(true, transactionID); err != nil {
			return nil, err
		}
		return nil, nil
	} else {
		if err := sqpIndexers.StoreIndexes(false, transactionID); err != nil {
			return nil, err
		}
	}
	return sqpIndexers, nil
}

func (api *APIerSv1) computeSupplierIndexes(tenant string, sppIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var supplierIDs []string
	sppIndexers := engine.NewFilterIndexer(api.DataManager, utils.SupplierProfilePrefix, tenant)
	if sppIDs == nil {
		ids, err := api.DataManager.DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			supplierIDs = append(supplierIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		supplierIDs = *sppIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range supplierIDs {
		spp, err := api.DataManager.GetSupplierProfile(tenant, id, true, false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		fltrIDs := make([]string, len(spp.FilterIDs))
		for i, fltrID := range spp.FilterIDs {
			fltrIDs[i] = fltrID
		}
		if len(fltrIDs) == 0 {
			fltrIDs = []string{utils.META_NONE}
		}
		for _, fltrID := range fltrIDs {
			var fltr *engine.Filter
			if fltrID == utils.META_NONE {
				fltr = &engine.Filter{
					Tenant: spp.Tenant,
					ID:     spp.ID,
					Rules: []*engine.FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = engine.GetFilter(api.DataManager, spp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for suppliers: %+v",
						fltrID, spp)
				}
				return nil, err
			}
			sppIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), spp.ID)
		}
	}
	if transactionID == utils.NonTransactional {
		if err := sppIndexers.StoreIndexes(true, transactionID); err != nil {
			return nil, err
		}
		return nil, nil
	} else {
		if err := sppIndexers.StoreIndexes(false, transactionID); err != nil {
			return nil, err
		}
	}
	return sppIndexers, nil
}

func (api *APIerSv1) computeChargerIndexes(tenant string, cppIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var chargerIDs []string
	cppIndexes := engine.NewFilterIndexer(api.DataManager, utils.ChargerProfilePrefix, tenant)
	if cppIDs == nil {
		ids, err := api.DataManager.DataDB().GetKeysForPrefix(utils.ChargerProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			chargerIDs = append(chargerIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		chargerIDs = *cppIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range chargerIDs {
		cpp, err := api.DataManager.GetChargerProfile(tenant, id, true, false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		fltrIDs := make([]string, len(cpp.FilterIDs))
		for i, fltrID := range cpp.FilterIDs {
			fltrIDs[i] = fltrID
		}
		if len(fltrIDs) == 0 {
			fltrIDs = []string{utils.META_NONE}
		}
		for _, fltrID := range fltrIDs {
			var fltr *engine.Filter
			if fltrID == utils.META_NONE {
				fltr = &engine.Filter{
					Tenant: cpp.Tenant,
					ID:     cpp.ID,
					Rules: []*engine.FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = engine.GetFilter(api.DataManager, cpp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for charger: %+v",
						fltrID, cpp)
				}
				return nil, err
			}
			cppIndexes.IndexTPFilter(engine.FilterToTPFilter(fltr), cpp.ID)
		}
	}
	if transactionID == utils.NonTransactional {
		if err := cppIndexes.StoreIndexes(true, transactionID); err != nil {
			return nil, err
		}
		return nil, nil
	} else {
		if err := cppIndexes.StoreIndexes(false, transactionID); err != nil {
			return nil, err
		}
	}
	return cppIndexes, nil
}

func (api *APIerSv1) computeDispatcherIndexes(tenant, context string, dspIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var dispatcherIDs []string
	dspIndexes := engine.NewFilterIndexer(api.DataManager, utils.DispatcherProfilePrefix,
		utils.ConcatenatedKey(tenant, context))
	if dspIDs == nil {
		ids, err := api.DataManager.DataDB().GetKeysForPrefix(utils.DispatcherProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			dispatcherIDs = append(dispatcherIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		dispatcherIDs = *dspIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range dispatcherIDs {
		dsp, err := api.DataManager.GetDispatcherProfile(tenant, id, true, false, utils.NonTransactional)
		if err != nil {
			return nil, err
		}
		if !utils.IsSliceMember(dsp.Subsystems, context) && context != utils.META_ANY {
			continue
		}
		fltrIDs := make([]string, len(dsp.FilterIDs))
		for i, fltrID := range dsp.FilterIDs {
			fltrIDs[i] = fltrID
		}
		if len(fltrIDs) == 0 {
			fltrIDs = []string{utils.META_NONE}
		}
		for _, fltrID := range fltrIDs {
			var fltr *engine.Filter
			if fltrID == utils.META_NONE {
				fltr = &engine.Filter{
					Tenant: dsp.Tenant,
					ID:     dsp.ID,
					Rules: []*engine.FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = engine.GetFilter(api.DataManager, dsp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for dispatcher: %+v",
						fltrID, dsp)
				}
				return nil, err
			}
			dspIndexes.IndexTPFilter(engine.FilterToTPFilter(fltr), dsp.ID)
		}
	}
	if transactionID == utils.NonTransactional {
		if err := dspIndexes.StoreIndexes(true, transactionID); err != nil {
			return nil, err
		}
		return nil, nil
	} else {
		if err := dspIndexes.StoreIndexes(false, transactionID); err != nil {
			return nil, err
		}
	}
	return dspIndexes, nil
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
