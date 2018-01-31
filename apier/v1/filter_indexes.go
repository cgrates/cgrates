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

func (self *ApierV1) ComputeFilterIndexes(args utils.ArgsComputeFilterIndexes, reply *string) error {
	//ThresholdProfile Indexes
	if err := self.computeThresholdIndexes(args.Tenant, args.ThresholdIDs); err != nil {
		return utils.APIErrorHandler(err)
	}
	//StatQueueProfile Indexes
	if err := self.computeStatIndexes(args.Tenant, args.StatIDs); err != nil {
		return utils.APIErrorHandler(err)
	}
	//ResourceProfile Indexes
	if err := self.computeResourceIndexes(args.Tenant, args.ResourceIDs); err != nil {
		return utils.APIErrorHandler(err)
	}
	//SupplierProfile Indexes
	if err := self.computeSupplierIndexes(args.Tenant, args.SupplierIDs); err != nil {
		return utils.APIErrorHandler(err)
	}
	//AttributeProfile Indexes
	if err := self.computeAttributeIndexes(args.Tenant, args.AttributeIDs); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func (self *ApierV1) computeThresholdIndexes(tenant string, thIDs *[]string) error {
	var thresholdIDs []string
	thdsIndexers := engine.NewFilterIndexer(self.DataManager, utils.ThresholdProfilePrefix, tenant)
	if thIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix)
		if err != nil {
			return err
		}
		for _, id := range ids {
			thresholdIDs = append(thresholdIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		thresholdIDs = *thIDs
	}
	for _, id := range thresholdIDs {
		th, err := self.DataManager.GetThresholdProfile(tenant, id, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		for _, fltrID := range th.FilterIDs {
			fltr, err := self.DataManager.GetFilter(tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for threshold: %+v", fltrID, th)
				}
				return err
			} else {
				thdsIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), th.ID)
			}
		}
	}
	if thIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix,
			tenant, false)); err != nil {
			return err
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix,
			tenant, true)); err != nil {
			return err
		}
	} else {
		indexRemover := engine.NewFilterIndexer(self.DataManager, utils.ThresholdProfilePrefix, tenant)
		for _, id := range thresholdIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return err
			}
		}
	}
	if err := thdsIndexers.StoreIndexes(); err != nil {
		return err
	}
	return nil
}

func (self *ApierV1) computeAttributeIndexes(tenant string, attrIDs *[]string) error {
	var attributeIDs []string
	attrIndexers := engine.NewFilterIndexer(self.DataManager, utils.AttributeProfilePrefix, tenant)
	if attrIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix)
		if err != nil {
			return err
		}
		for _, id := range ids {
			err = attrIndexers.RemoveItemFromIndex(strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				return err
			}
			attributeIDs = append(attributeIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		attributeIDs = *attrIDs
	}
	for _, id := range attributeIDs {
		ap, err := self.DataManager.GetAttributeProfile(tenant, id, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		for _, fltrID := range ap.FilterIDs {
			fltr, err := self.DataManager.GetFilter(tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for stats queue: %+v", fltrID, ap)
				}
				return err
			} else {
				tpFltr := engine.FilterToTPFilter(fltr)
				attrIndexers.IndexTPFilter(tpFltr, ap.ID)

			}
		}
	}
	if attrIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix,
			tenant, false)); err != nil {
			return err
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix,
			tenant, true)); err != nil {
			return err
		}
	} else {
		indexRemover := engine.NewFilterIndexer(self.DataManager, utils.AttributeProfilePrefix, tenant)
		for _, id := range attributeIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return err
			}
		}
	}
	if err := attrIndexers.StoreIndexes(); err != nil {
		return err
	}
	return nil
}

func (self *ApierV1) computeResourceIndexes(tenant string, rsIDs *[]string) error {
	var resourceIDs []string
	rpIndexers := engine.NewFilterIndexer(self.DataManager, utils.ResourceProfilesPrefix, tenant)
	if rsIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix)
		if err != nil {
			return err
		}
		for _, id := range ids {
			err = rpIndexers.RemoveItemFromIndex(strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				return err
			}
			resourceIDs = append(resourceIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		resourceIDs = *rsIDs
	}
	for _, id := range resourceIDs {
		rp, err := self.DataManager.GetResourceProfile(tenant, id, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		for _, fltrID := range rp.FilterIDs {
			fltr, err := self.DataManager.GetFilter(tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for stats queue: %+v", fltrID, rp)
				}
				return err
			} else {
				rpIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), rp.ID)
			}
		}
	}
	if rsIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix,
			tenant, false)); err != nil {
			return err
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix,
			tenant, true)); err != nil {
			return err
		}
	} else {
		indexRemover := engine.NewFilterIndexer(self.DataManager, utils.ResourceProfilesPrefix, tenant)
		for _, id := range resourceIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return err
			}
		}
	}
	if err := rpIndexers.StoreIndexes(); err != nil {
		return err
	}
	return nil
}

func (self *ApierV1) computeStatIndexes(tenant string, stIDs *[]string) error {
	var statIDs []string
	sqpIndexers := engine.NewFilterIndexer(self.DataManager, utils.StatQueueProfilePrefix, tenant)
	if stIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix)
		if err != nil {
			return err
		}
		for _, id := range ids {
			err = sqpIndexers.RemoveItemFromIndex(strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				return err
			}
			statIDs = append(statIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		statIDs = *stIDs
	}
	for _, id := range statIDs {
		sqp, err := self.DataManager.GetStatQueueProfile(tenant, id, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		for _, fltrID := range sqp.FilterIDs {
			fltr, err := self.DataManager.GetFilter(tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for stats queue: %+v", fltrID, sqp)
				}
				return err
			} else {
				sqpIndexers.IndexTPFilter(engine.FilterToTPFilter(fltr), sqp.ID)
			}
		}
	}
	if stIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix,
			tenant, false)); err != nil {
			return err
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix,
			tenant, true)); err != nil {
			return err
		}
	} else {
		indexRemover := engine.NewFilterIndexer(self.DataManager, utils.StatQueueProfilePrefix, tenant)
		for _, id := range statIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return err
			}
		}
	}
	if err := sqpIndexers.StoreIndexes(); err != nil {
		return err
	}
	return nil
}

func (self *ApierV1) computeSupplierIndexes(tenant string, sppIDs *[]string) error {
	var supplierIDs []string
	sppIndexers := engine.NewFilterIndexer(self.DataManager, utils.SupplierProfilePrefix, tenant)
	if sppIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
		if err != nil {
			return err
		}
		for _, id := range ids {
			err = sppIndexers.RemoveItemFromIndex(strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				return err
			}
			supplierIDs = append(supplierIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		supplierIDs = *sppIDs
	}
	for _, id := range supplierIDs {
		spp, err := self.DataManager.GetSupplierProfile(tenant, id, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		for _, fltrID := range spp.FilterIDs {
			fltr, err := self.DataManager.GetFilter(tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for stats queue: %+v", fltrID, spp)
				}
				return err
			} else {
				tpFltr := engine.FilterToTPFilter(fltr)
				sppIndexers.IndexTPFilter(tpFltr, spp.ID)
			}
		}
	}
	if sppIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix,
			tenant, false)); err != nil {
			return err
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix,
			tenant, true)); err != nil {
			return err
		}
	} else {
		indexRemover := engine.NewFilterIndexer(self.DataManager, utils.SupplierProfilePrefix, tenant)
		for _, id := range supplierIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return err
			}
		}
	}
	if err := sppIndexers.StoreIndexes(); err != nil {
		return err
	}
	return nil
}
