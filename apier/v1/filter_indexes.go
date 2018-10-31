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

func (self *ApierV1) RemoveFilterIndexes(arg AttrRemFilterIndexes, reply *string) (err error) {
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
	case utils.MetaAttributes:
		if missing := utils.MissingStructFields(&arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.AttributeProfilePrefix
		key = utils.ConcatenatedKey(arg.Tenant, arg.Context)
	}
	if err = self.DataManager.RemoveFilterIndexes(utils.PrefixToIndexCache[arg.ItemType], key); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (self *ApierV1) GetFilterIndexes(arg AttrGetFilterIndexes, reply *[]string) (err error) {
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
	case utils.MetaAttributes:
		if missing := utils.MissingStructFields(&arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.AttributeProfilePrefix
		key = utils.ConcatenatedKey(arg.Tenant, arg.Context)
	}
	if indexes, err = self.DataManager.GetFilterIndexes(
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
	if arg.Paginator.Limit != nil || arg.Paginator.Offset != nil || arg.Paginator.SearchTerm != "" {
		*reply = arg.Paginator.PaginateStringSlice(indexedSlice)
	} else {
		*reply = indexedSlice
	}
	return nil
}

func (self *ApierV1) ComputeFilterIndexes(args utils.ArgsComputeFilterIndexes, reply *string) error {
	transactionID := utils.GenUUID()
	//ThresholdProfile Indexes
	thdsIndexers, err := self.computeThresholdIndexes(args.Tenant, args.ThresholdIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//StatQueueProfile Indexes
	sqpIndexers, err := self.computeStatIndexes(args.Tenant, args.StatIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//ResourceProfile Indexes
	rsIndexes, err := self.computeResourceIndexes(args.Tenant, args.ResourceIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//SupplierProfile Indexes
	sppIndexes, err := self.computeSupplierIndexes(args.Tenant, args.SupplierIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//AttributeProfile Indexes
	attrIndexes, err := self.computeAttributeIndexes(args.Tenant, args.Context, args.AttributeIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//ChargerProfile  Indexes
	cppIndexes, err := self.computeChargerIndexes(args.Tenant, args.ChargerIDs, transactionID)
	if err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	//Now we move from tmpKey to the right key for each type
	//ThresholdProfile Indexes
	if thdsIndexers != nil {
		if err := thdsIndexers.StoreIndexes(true, transactionID); err != nil {
			if args.ThresholdIDs != nil {
				for _, id := range *args.ThresholdIDs {
					th, err := self.DataManager.GetThresholdProfile(args.Tenant, id, true, false, utils.NonTransactional)
					if err != nil {
						return err
					}
					if err := thdsIndexers.RemoveItemFromIndex(args.Tenant, id, th.FilterIDs); err != nil {
						return err
					}
				}
			}
			return err
		}
	}
	//StatQueueProfile Indexes
	if sqpIndexers != nil {
		if err := sqpIndexers.StoreIndexes(true, transactionID); err != nil {
			for _, id := range *args.StatIDs {
				sqp, err := self.DataManager.GetStatQueueProfile(args.Tenant, id, true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				if err := sqpIndexers.RemoveItemFromIndex(args.Tenant, id, sqp.FilterIDs); err != nil {
					return err
				}
			}
			return err
		}
	}
	//ResourceProfile Indexes
	if rsIndexes != nil {
		if err := rsIndexes.StoreIndexes(true, transactionID); err != nil {
			for _, id := range *args.ResourceIDs {
				rp, err := self.DataManager.GetResourceProfile(args.Tenant, id, true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				if err := rsIndexes.RemoveItemFromIndex(args.Tenant, id, rp.FilterIDs); err != nil {
					return err
				}
			}
			return err
		}
	}
	//SupplierProfile Indexes
	if sppIndexes != nil {
		if err := sppIndexes.StoreIndexes(true, transactionID); err != nil {
			for _, id := range *args.SupplierIDs {
				spp, err := self.DataManager.GetSupplierProfile(args.Tenant, id, true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				if err := sppIndexes.RemoveItemFromIndex(args.Tenant, id, spp.FilterIDs); err != nil {
					return err
				}
			}
			return err
		}
	}
	//AttributeProfile Indexes
	if attrIndexes != nil {
		if err := attrIndexes.StoreIndexes(true, transactionID); err != nil {
			for _, id := range *args.AttributeIDs {
				ap, err := self.DataManager.GetAttributeProfile(args.Tenant, id, true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				if err := attrIndexes.RemoveItemFromIndex(args.Tenant, id, ap.FilterIDs); err != nil {
					return err
				}
			}
			return err
		}
	}
	//ChargerProfile Indexes
	if cppIndexes != nil {
		if err := cppIndexes.StoreIndexes(true, transactionID); err != nil {
			for _, id := range *args.ChargerIDs {
				cpp, err := self.DataManager.GetChargerProfile(args.Tenant, id, true, false, utils.NonTransactional)
				if err != nil {
					return err
				}
				if err := cppIndexes.RemoveItemFromIndex(args.Tenant, id, cpp.FilterIDs); err != nil {
					return err
				}
			}
			return err
		}
	}
	*reply = utils.OK
	return nil
}

func (self *ApierV1) computeThresholdIndexes(tenant string, thIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var thresholdIDs []string
	thdsIndexers := engine.NewFilterIndexer(self.DataManager, utils.ThresholdProfilePrefix, tenant)
	if thIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix)
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
		th, err := self.DataManager.GetThresholdProfile(tenant, id, true, false, utils.NonTransactional)
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
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = self.DataManager.GetFilter(th.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for threshold: %+v",
						fltrID, th)
				}
				return nil, err
			} else {
				thdsIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), th.ID)
			}
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

func (self *ApierV1) computeAttributeIndexes(tenant, context string, attrIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var attributeIDs []string
	attrIndexers := engine.NewFilterIndexer(self.DataManager, utils.AttributeProfilePrefix,
		utils.ConcatenatedKey(tenant, context))
	if attrIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix)
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
		ap, err := self.DataManager.GetAttributeProfile(tenant, id, true, false, utils.NonTransactional)
		if err != nil {
			return nil, err
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
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = self.DataManager.GetFilter(ap.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for attribute: %+v",
						fltrID, ap)
				}
				return nil, err
			} else {
				attrIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), ap.ID)
			}
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

func (self *ApierV1) computeResourceIndexes(tenant string, rsIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var resourceIDs []string
	rpIndexers := engine.NewFilterIndexer(self.DataManager, utils.ResourceProfilesPrefix, tenant)
	if rsIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix)
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
		rp, err := self.DataManager.GetResourceProfile(tenant, id, true, false, utils.NonTransactional)
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
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = self.DataManager.GetFilter(rp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for resource: %+v",
						fltrID, rp)
				}
				return nil, err
			} else {
				rpIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), rp.ID)
			}
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

func (self *ApierV1) computeStatIndexes(tenant string, stIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var statIDs []string
	sqpIndexers := engine.NewFilterIndexer(self.DataManager, utils.StatQueueProfilePrefix, tenant)
	if stIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix)
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
		sqp, err := self.DataManager.GetStatQueueProfile(tenant, id, true, false, utils.NonTransactional)
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
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = self.DataManager.GetFilter(sqp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for statqueue: %+v",
						fltrID, sqp)
				}
				return nil, err
			} else {
				sqpIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), sqp.ID)
			}
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

func (self *ApierV1) computeSupplierIndexes(tenant string, sppIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var supplierIDs []string
	sppIndexers := engine.NewFilterIndexer(self.DataManager, utils.SupplierProfilePrefix, tenant)
	if sppIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
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
		spp, err := self.DataManager.GetSupplierProfile(tenant, id, true, false, utils.NonTransactional)
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
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = self.DataManager.GetFilter(spp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for suppliers: %+v",
						fltrID, spp)
				}
				return nil, err
			} else {
				sppIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), spp.ID)
			}
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

func (self *ApierV1) computeChargerIndexes(tenant string, cppIDs *[]string,
	transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var chargerIDs []string
	cppIndexes := engine.NewFilterIndexer(self.DataManager, utils.ChargerProfilePrefix, tenant)
	if cppIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.ChargerProfilePrefix)
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
		cpp, err := self.DataManager.GetChargerProfile(tenant, id, true, false, utils.NonTransactional)
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
							Type:      utils.META_NONE,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = self.DataManager.GetFilter(cpp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for charger: %+v",
						fltrID, cpp)
				}
				return nil, err
			} else {
				cppIndexes.IndexTPFilter(engine.FilterToTPFilter(fltr), cpp.ID)
			}
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
