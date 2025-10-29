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

package engine

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

func ComputeThresholdIndexes(dm *DataManager, tenant string, thIDs *[]string,
	transactionID string) (filterIndexer *FilterIndexer, err error) {
	var thresholdIDs []string
	var thdsIndexers *FilterIndexer
	if thIDs == nil {
		ids, err := dm.DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			thresholdIDs = append(thresholdIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
		// this will be on ComputeIndexes that contains empty indexes
		thdsIndexers = NewFilterIndexer(dm, utils.ThresholdProfilePrefix, tenant)
	} else {
		// this will be on ComputeIndexesIDs that contains the old indexes from the next getter
		var oldIDx map[string]utils.StringMap
		if oldIDx, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.ThresholdProfilePrefix],
			tenant, utils.EmptyString, nil); err != nil || oldIDx == nil {
			thdsIndexers = NewFilterIndexer(dm, utils.ThresholdProfilePrefix, tenant)
		} else {
			thdsIndexers = NewFilterIndexerWithIndexes(dm, utils.ThresholdProfilePrefix, tenant, oldIDx)
		}
		thresholdIDs = *thIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range thresholdIDs {
		th, err := dm.GetThresholdProfile(tenant, id, true, false, utils.NonTransactional)
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
			var fltr *Filter
			if fltrID == utils.META_NONE {
				fltr = &Filter{
					Tenant: th.Tenant,
					ID:     th.ID,
					Rules: []*FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = GetFilter(dm, th.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for threshold: %+v",
						fltrID, th)
				}
				return nil, err
			}
			thdsIndexers.IndexTPFilter(FilterToTPFilter(fltr), th.ID)
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

func ComputeChargerIndexes(dm *DataManager, tenant string, cppIDs *[]string,
	transactionID string) (filterIndexer *FilterIndexer, err error) {
	var chargerIDs []string
	var cppIndexes *FilterIndexer
	if cppIDs == nil {
		ids, err := dm.DataDB().GetKeysForPrefix(utils.ChargerProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			chargerIDs = append(chargerIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
		// this will be on ComputeIndexes that contains empty indexes
		cppIndexes = NewFilterIndexer(dm, utils.ChargerProfilePrefix, tenant)
	} else {
		// this will be on ComputeIndexesIDs that contains the old indexes from the next getter
		var oldIDx map[string]utils.StringMap
		if oldIDx, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.ChargerProfilePrefix],
			tenant, utils.EmptyString, nil); err != nil || oldIDx == nil {
			cppIndexes = NewFilterIndexer(dm, utils.ChargerProfilePrefix, tenant)
		} else {
			cppIndexes = NewFilterIndexerWithIndexes(dm, utils.ChargerProfilePrefix, tenant, oldIDx)
		}
		chargerIDs = *cppIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range chargerIDs {
		cpp, err := dm.GetChargerProfile(tenant, id, true, false, utils.NonTransactional)
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
			var fltr *Filter
			if fltrID == utils.META_NONE {
				fltr = &Filter{
					Tenant: cpp.Tenant,
					ID:     cpp.ID,
					Rules: []*FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = GetFilter(dm, cpp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for charger: %+v",
						fltrID, cpp)
				}
				return nil, err
			}
			cppIndexes.IndexTPFilter(FilterToTPFilter(fltr), cpp.ID)
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

func ComputeResourceIndexes(dm *DataManager, tenant string, rsIDs *[]string,
	transactionID string) (filterIndexer *FilterIndexer, err error) {
	var resourceIDs []string
	var rpIndexers *FilterIndexer
	if rsIDs == nil {
		ids, err := dm.DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			resourceIDs = append(resourceIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
		// this will be on ComputeIndexes that contains empty indexes
		rpIndexers = NewFilterIndexer(dm, utils.ResourceProfilesPrefix, tenant)
	} else {
		// this will be on ComputeIndexesIDs that contains the old indexes from the next getter
		var oldIDx map[string]utils.StringMap
		if oldIDx, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.ResourceProfilesPrefix],
			tenant, utils.EmptyString, nil); err != nil || oldIDx == nil {
			rpIndexers = NewFilterIndexer(dm, utils.ResourceProfilesPrefix, tenant)
		} else {
			rpIndexers = NewFilterIndexerWithIndexes(dm, utils.ResourceProfilesPrefix, tenant, oldIDx)
		}
		resourceIDs = *rsIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range resourceIDs {
		rp, err := dm.GetResourceProfile(tenant, id, true, false, utils.NonTransactional)
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
			var fltr *Filter
			if fltrID == utils.META_NONE {
				fltr = &Filter{
					Tenant: rp.Tenant,
					ID:     rp.ID,
					Rules: []*FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = GetFilter(dm, rp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for resource: %+v",
						fltrID, rp)
				}
				return nil, err
			}
			rpIndexers.IndexTPFilter(FilterToTPFilter(fltr), rp.ID)
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

func ComputeSupplierIndexes(dm *DataManager, tenant string, sppIDs *[]string,
	transactionID string) (filterIndexer *FilterIndexer, err error) {
	var supplierIDs []string
	var sppIndexers *FilterIndexer
	if sppIDs == nil {
		ids, err := dm.DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			supplierIDs = append(supplierIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
		// this will be on ComputeIndexes that contains empty indexes
		sppIndexers = NewFilterIndexer(dm, utils.SupplierProfilePrefix, tenant)
	} else {
		// this will be on ComputeIndexesIDs that contains the old indexes from the next getter
		var oldIDx map[string]utils.StringMap
		if oldIDx, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.SupplierProfilePrefix],
			tenant, utils.EmptyString, nil); err != nil || oldIDx == nil {
			sppIndexers = NewFilterIndexer(dm, utils.SupplierProfilePrefix, tenant)
		} else {
			sppIndexers = NewFilterIndexerWithIndexes(dm, utils.SupplierProfilePrefix, tenant, oldIDx)
		}
		supplierIDs = *sppIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range supplierIDs {
		spp, err := dm.GetSupplierProfile(tenant, id, true, false, utils.NonTransactional)
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
			var fltr *Filter
			if fltrID == utils.META_NONE {
				fltr = &Filter{
					Tenant: spp.Tenant,
					ID:     spp.ID,
					Rules: []*FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = GetFilter(dm, spp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for suppliers: %+v",
						fltrID, spp)
				}
				return nil, err
			}
			sppIndexers.IndexTPFilter(FilterToTPFilter(fltr), spp.ID)
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

func ComputeStatIndexes(dm *DataManager, tenant string, stIDs *[]string,
	transactionID string) (filterIndexer *FilterIndexer, err error) {
	var statIDs []string
	var sqpIndexers *FilterIndexer
	if stIDs == nil {
		ids, err := dm.DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			statIDs = append(statIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
		// this will be on ComputeIndexes that contains empty indexes
		sqpIndexers = NewFilterIndexer(dm, utils.StatQueueProfilePrefix, tenant)
	} else {
		// this will be on ComputeIndexesIDs that contains the old indexes from the next getter
		var oldIDx map[string]utils.StringMap
		if oldIDx, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.StatQueueProfilePrefix],
			tenant, utils.EmptyString, nil); err != nil || oldIDx == nil {
			sqpIndexers = NewFilterIndexer(dm, utils.StatQueueProfilePrefix, tenant)
		} else {
			sqpIndexers = NewFilterIndexerWithIndexes(dm, utils.StatQueueProfilePrefix, tenant, oldIDx)
		}
		statIDs = *stIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range statIDs {
		sqp, err := dm.GetStatQueueProfile(tenant, id, true, false, utils.NonTransactional)
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
			var fltr *Filter
			if fltrID == utils.META_NONE {
				fltr = &Filter{
					Tenant: sqp.Tenant,
					ID:     sqp.ID,
					Rules: []*FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = GetFilter(dm, sqp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for statqueue: %+v",
						fltrID, sqp)
				}
				return nil, err
			}
			sqpIndexers.IndexTPFilter(FilterToTPFilter(fltr), sqp.ID)
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

func ComputeAttributeIndexes(dm *DataManager, tenant, context string, attrIDs *[]string,
	transactionID string) (filterIndexer *FilterIndexer, err error) {
	var attributeIDs []string
	var attrIndexers *FilterIndexer
	if attrIDs == nil {
		ids, err := dm.DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			attributeIDs = append(attributeIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
		// this will be on ComputeIndexes that contains empty indexes
		attrIndexers = NewFilterIndexer(dm, utils.AttributeProfilePrefix,
			utils.ConcatenatedKey(tenant, context))
	} else {
		// this will be on ComputeIndexesIDs that contains the old indexes from the next getter
		var oldIDx map[string]utils.StringMap
		if oldIDx, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.AttributeProfilePrefix],
			utils.ConcatenatedKey(tenant, context), utils.EmptyString, nil); err != nil || oldIDx == nil {
			attrIndexers = NewFilterIndexer(dm, utils.AttributeProfilePrefix, utils.ConcatenatedKey(tenant, context))
		} else {
			attrIndexers = NewFilterIndexerWithIndexes(dm, utils.AttributeProfilePrefix, utils.ConcatenatedKey(tenant, context), oldIDx)
		}
		attributeIDs = *attrIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range attributeIDs {
		ap, err := dm.GetAttributeProfile(tenant, id, true, false, utils.NonTransactional)
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
			var fltr *Filter
			if fltrID == utils.META_NONE {
				fltr = &Filter{
					Tenant: ap.Tenant,
					ID:     ap.ID,
					Rules: []*FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = GetFilter(dm, ap.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for attribute: %+v",
						fltrID, ap)
				}
				return nil, err
			}
			attrIndexers.IndexTPFilter(FilterToTPFilter(fltr), ap.ID)
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

func ComputeDispatcherIndexes(dm *DataManager, tenant, context string, dspIDs *[]string,
	transactionID string) (filterIndexer *FilterIndexer, err error) {
	var dispatcherIDs []string
	var dspIndexes *FilterIndexer
	if dspIDs == nil {
		ids, err := dm.DataDB().GetKeysForPrefix(utils.DispatcherProfilePrefix)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			dispatcherIDs = append(dispatcherIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
		// this will be on ComputeIndexes that contains empty indexes
		dspIndexes = NewFilterIndexer(dm, utils.DispatcherProfilePrefix,
			utils.ConcatenatedKey(tenant, context))
	} else {
		// this will be on ComputeIndexesIDs that contains the old indexes from the next getter
		var oldIDx map[string]utils.StringMap
		if oldIDx, err = dm.GetFilterIndexes(utils.PrefixToIndexCache[utils.DispatcherProfilePrefix],
			utils.ConcatenatedKey(tenant, context), utils.EmptyString, nil); err != nil || oldIDx == nil {
			dspIndexes = NewFilterIndexer(dm, utils.DispatcherProfilePrefix,
				utils.ConcatenatedKey(tenant, context))
		} else {
			dspIndexes = NewFilterIndexerWithIndexes(dm, utils.DispatcherProfilePrefix,
				utils.ConcatenatedKey(tenant, context), oldIDx)
		}
		dispatcherIDs = *dspIDs
		transactionID = utils.NonTransactional
	}
	for _, id := range dispatcherIDs {
		dsp, err := dm.GetDispatcherProfile(tenant, id, true, false, utils.NonTransactional)
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
			var fltr *Filter
			if fltrID == utils.META_NONE {
				fltr = &Filter{
					Tenant: dsp.Tenant,
					ID:     dsp.ID,
					Rules: []*FilterRule{
						{
							Type:    utils.META_NONE,
							Element: utils.META_ANY,
							Values:  []string{utils.META_ANY},
						},
					},
				}
			} else if fltr, err = GetFilter(dm, dsp.Tenant, fltrID,
				true, false, utils.NonTransactional); err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for dispatcher: %+v",
						fltrID, dsp)
				}
				return nil, err
			}
			dspIndexes.IndexTPFilter(FilterToTPFilter(fltr), dsp.ID)
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
