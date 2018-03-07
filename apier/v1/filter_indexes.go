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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"strings"
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

type AttrGetFilterReverseIndexes struct {
	Tenant      string
	Context     string
	ItemType    string
	ItemIDs     []string
	FilterType  string
	FilterField string
	FilterValue string
	utils.Paginator
}

func (self *ApierV1) GetFilterReverseIndexes(arg AttrGetFilterReverseIndexes, reply *[]string) (err error) {
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
	case utils.MetaAttributes:
		if missing := utils.MissingStructFields(&arg, []string{"Context"}); len(missing) != 0 { //Params missing
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.AttributeProfilePrefix
		key = utils.ConcatenatedKey(arg.Tenant, arg.Context)
	}
	if arg.ItemIDs != nil {
		indexes = make(map[string]utils.StringMap)
		for _, itemID := range arg.ItemIDs {
			if tmpIndexes, err := self.DataManager.GetFilterReverseIndexes(
				utils.PrefixToRevIndexCache[arg.ItemType], key, map[string]string{itemID: ""}); err != nil {
				return err
			} else {
				for key, val := range tmpIndexes {
					indexes[key] = make(utils.StringMap)
					indexes[key] = val
				}

			}
		}
	} else {
		indexes, err = self.DataManager.GetFilterReverseIndexes(
			utils.PrefixToRevIndexCache[arg.ItemType], key, nil)
		if err != nil {
			return err
		}
	}
	if arg.FilterType != "" {
		for val, strmap := range indexes {
			indexesFilter[val] = make(utils.StringMap)
			for _, value := range strmap.Slice() {
				if strings.HasPrefix(value, arg.FilterType) {
					indexesFilter[val][value] = true
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
				indexesFilter[val] = make(utils.StringMap)
				for _, value := range strmap.Slice() {
					if strings.Index(value, arg.FilterField) != -1 {
						indexesFilter[val][value] = true
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
				for _, value := range strmap.Slice() {
					if strings.Index(value, arg.FilterField) != -1 {
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
				for _, value := range strmap.Slice() {
					if strings.Index(value, arg.FilterValue) != -1 {
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
				for _, value := range strmap.Slice() {
					if strings.Index(value, arg.FilterValue) != -1 {
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
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	//StatQueueProfile Indexes
	sqpIndexers, err := self.computeStatIndexes(args.Tenant, args.StatIDs, transactionID)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	//ResourceProfile Indexes
	rsIndexes, err := self.computeResourceIndexes(args.Tenant, args.ResourceIDs, transactionID)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	//SupplierProfile Indexes
	sppIndexes, err := self.computeSupplierIndexes(args.Tenant, args.SupplierIDs, transactionID)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	//AttributeProfile Indexes
	attrIndexes, err := self.computeAttributeIndexes(args.Tenant, args.AttributeIDs, transactionID)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	//Now we move from tmpKey to the right key for each type
	//ThresholdProfile Indexes
	if thdsIndexers != nil {
		if err := thdsIndexers.StoreIndexes(true, transactionID); err != nil {
			for _, id := range *args.ThresholdIDs {
				if err := thdsIndexers.RemoveItemFromIndex(id); err != nil {
					return err
				}
			}
			return err
		}
	}
	//StatQueueProfile Indexes
	if sqpIndexers != nil {
		if err := sqpIndexers.StoreIndexes(true, transactionID); err != nil {
			for _, id := range *args.StatIDs {
				if err := thdsIndexers.RemoveItemFromIndex(id); err != nil {
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
				if err := thdsIndexers.RemoveItemFromIndex(id); err != nil {
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
				if err := thdsIndexers.RemoveItemFromIndex(id); err != nil {
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
				if err := thdsIndexers.RemoveItemFromIndex(id); err != nil {
					return err
				}
			}
			return err
		}
	}
	*reply = utils.OK
	return nil
}

func (self *ApierV1) computeThresholdIndexes(tenant string, thIDs *[]string, transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
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
		th, err := self.DataManager.GetThresholdProfile(tenant, id, false, utils.NonTransactional)
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
						&engine.FilterRule{
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if strings.HasPrefix(fltrID, utils.Meta) {
				inFltr, err := engine.NewInlineFilter(fltrID)
				if err != nil {
					return nil, err
				}
				fltr, err = inFltr.AsFilter(th.Tenant)
				if err != nil {
					return nil, err
				}
			} else if fltr, err = self.DataManager.GetFilter(th.Tenant, fltrID,
				false, utils.NonTransactional); err != nil {
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

func (self *ApierV1) computeAttributeIndexes(tenant string, attrIDs *[]string, transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
	var attributeIDs []string
	attrIndexers := engine.NewFilterIndexer(self.DataManager, utils.AttributeProfilePrefix, tenant)
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
		ap, err := self.DataManager.GetAttributeProfile(tenant, id, false, utils.NonTransactional)
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
						&engine.FilterRule{
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if strings.HasPrefix(fltrID, utils.Meta) {
				inFltr, err := engine.NewInlineFilter(fltrID)
				if err != nil {
					return nil, err
				}
				fltr, err = inFltr.AsFilter(ap.Tenant)
				if err != nil {
					return nil, err
				}
			} else if fltr, err = self.DataManager.GetFilter(ap.Tenant, fltrID,
				false, utils.NonTransactional); err != nil {
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

func (self *ApierV1) computeResourceIndexes(tenant string, rsIDs *[]string, transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
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
		rp, err := self.DataManager.GetResourceProfile(tenant, id, false, utils.NonTransactional)
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
						&engine.FilterRule{
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if strings.HasPrefix(fltrID, utils.Meta) {
				inFltr, err := engine.NewInlineFilter(fltrID)
				if err != nil {
					return nil, err
				}
				fltr, err = inFltr.AsFilter(rp.Tenant)
				if err != nil {
					return nil, err
				}
			} else if fltr, err = self.DataManager.GetFilter(rp.Tenant, fltrID,
				false, utils.NonTransactional); err != nil {
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

func (self *ApierV1) computeStatIndexes(tenant string, stIDs *[]string, transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
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
		sqp, err := self.DataManager.GetStatQueueProfile(tenant, id, false, utils.NonTransactional)
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
						&engine.FilterRule{
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if strings.HasPrefix(fltrID, utils.Meta) {
				inFltr, err := engine.NewInlineFilter(fltrID)
				if err != nil {
					return nil, err
				}
				fltr, err = inFltr.AsFilter(sqp.Tenant)
				if err != nil {
					return nil, err
				}
			} else if fltr, err = self.DataManager.GetFilter(sqp.Tenant, fltrID,
				false, utils.NonTransactional); err != nil {
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

func (self *ApierV1) computeSupplierIndexes(tenant string, sppIDs *[]string, transactionID string) (filterIndexer *engine.FilterIndexer, err error) {
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
		spp, err := self.DataManager.GetSupplierProfile(tenant, id, false, utils.NonTransactional)
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
						&engine.FilterRule{
							Type:      utils.MetaDefault,
							FieldName: utils.META_ANY,
							Values:    []string{utils.META_ANY},
						},
					},
				}
			} else if strings.HasPrefix(fltrID, utils.Meta) {
				inFltr, err := engine.NewInlineFilter(fltrID)
				if err != nil {
					return nil, err
				}
				fltr, err = inFltr.AsFilter(spp.Tenant)
				if err != nil {
					return nil, err
				}
			} else if fltr, err = self.DataManager.GetFilter(spp.Tenant, fltrID,
				false, utils.NonTransactional); err != nil {
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
