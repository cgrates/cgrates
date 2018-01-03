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

/*
type ArgsComputeFilterIndexes struct {
	Tenant       string
	AttributeIDs *[]string
	ResourceIDs  *[]string
	StatIDs      *[]string
	SupplierIDs  *[]string
	ThresholdIDs *[]string
}
*/

func (self *ApierV1) ComputeFilterIndexes(args utils.ArgsComputeFilterIndexes, reply *string) error {
	//ThresholdProfile Indexes
	var thresholdIDs []string
	thdsIndexers := engine.NewReqFilterIndexer(self.DataManager, utils.ThresholdProfilePrefix, args.Tenant)
	if args.ThresholdIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.ThresholdProfilePrefix)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, id := range ids {
			thresholdIDs = append(thresholdIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		thresholdIDs = *args.ThresholdIDs
	}

	for _, id := range thresholdIDs {
		th, err := self.DataManager.GetThresholdProfile(args.Tenant, id, false, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, fltrID := range th.FilterIDs {
			fltr, err := self.DataManager.GetFilter(args.Tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for threshold: %+v", fltrID, th)
				}
				return utils.APIErrorHandler(err)
			} else {
				tpFltr := engine.FilterToTPFilter(fltr)
				thdsIndexers.IndexTPFilter(tpFltr, th.ID)
			}
		}
	}
	if args.ThresholdIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix,
			args.Tenant, false)); err != nil {
			if err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix,
			args.Tenant, true), ""); err != nil {
			if err != utils.ErrNotFound {
				return utils.APIErrorHandler(err)
			}
		}
	} else {
		indexRemover := engine.NewReqFilterIndexer(self.DataManager, utils.ThresholdProfilePrefix, args.Tenant)
		for _, id := range thresholdIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return utils.APIErrorHandler(err)
			}
		}
	}
	if err := thdsIndexers.StoreIndexes(); err != nil {
		return utils.APIErrorHandler(err)
	}
	//StatQueueProfile Indexes

	var statIDs []string
	sqpIndexers := engine.NewReqFilterIndexer(self.DataManager, utils.StatQueueProfilePrefix, args.Tenant)
	if args.StatIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.StatQueueProfilePrefix)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, id := range ids {
			err = sqpIndexers.RemoveItemFromIndex(strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				return utils.APIErrorHandler(err)
			}
			statIDs = append(statIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		statIDs = *args.StatIDs
	}
	for _, id := range statIDs {
		sqp, err := self.DataManager.GetStatQueueProfile(args.Tenant, id, false, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, fltrID := range sqp.FilterIDs {
			fltr, err := self.DataManager.GetFilter(args.Tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for stats queue: %+v", fltrID, sqp)
				}
				return utils.APIErrorHandler(err)
			} else {
				tpFltr := engine.FilterToTPFilter(fltr)
				sqpIndexers.IndexTPFilter(tpFltr, sqp.ID)
			}
		}
	}
	if args.StatIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix,
			args.Tenant, false)); err != nil {
			return utils.APIErrorHandler(err)
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix,
			args.Tenant, true), ""); err != nil {
			return utils.APIErrorHandler(err)
		}

	} else {
		indexRemover := engine.NewReqFilterIndexer(self.DataManager, utils.StatQueueProfilePrefix, args.Tenant)
		for _, id := range statIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return utils.APIErrorHandler(err)
			}
		}
	}
	if err := sqpIndexers.StoreIndexes(); err != nil {
		return err
	}
	//ResourceProfile Indexes
	var resourceIDs []string
	rpIndexers := engine.NewReqFilterIndexer(self.DataManager, utils.ResourceProfilesPrefix, args.Tenant)
	if args.ResourceIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.ResourceProfilesPrefix)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, id := range ids {
			err = rpIndexers.RemoveItemFromIndex(strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				return utils.APIErrorHandler(err)
			}
			resourceIDs = append(resourceIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		resourceIDs = *args.ResourceIDs
	}
	for _, id := range resourceIDs {
		rp, err := self.DataManager.GetResourceProfile(args.Tenant, id, false, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, fltrID := range rp.FilterIDs {
			fltr, err := self.DataManager.GetFilter(args.Tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for stats queue: %+v", fltrID, rp)
				}
				return utils.APIErrorHandler(err)
			} else {
				tpFltr := engine.FilterToTPFilter(fltr)
				rpIndexers.IndexTPFilter(tpFltr, rp.ID)

			}
		}
	}
	if args.ResourceIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix,
			args.Tenant, false)); err != nil {
			return utils.APIErrorHandler(err)
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix,
			args.Tenant, true), ""); err != nil {
			return utils.APIErrorHandler(err)
		}
	} else {
		indexRemover := engine.NewReqFilterIndexer(self.DataManager, utils.ResourceProfilesPrefix, args.Tenant)
		for _, id := range resourceIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return utils.APIErrorHandler(err)
			}
		}
	}
	if err := rpIndexers.StoreIndexes(); err != nil {
		return err
	}
	//SupplierProfile Indexes
	var supplierIDs []string
	sppIndexers := engine.NewReqFilterIndexer(self.DataManager, utils.SupplierProfilePrefix, args.Tenant)
	if args.SupplierIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, id := range ids {
			err = sppIndexers.RemoveItemFromIndex(strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				return utils.APIErrorHandler(err)
			}
			supplierIDs = append(supplierIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		supplierIDs = *args.SupplierIDs
	}
	for _, id := range supplierIDs {
		spp, err := self.DataManager.GetSupplierProfile(args.Tenant, id, false, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, fltrID := range spp.FilterIDs {
			fltr, err := self.DataManager.GetFilter(args.Tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for stats queue: %+v", fltrID, spp)
				}
				return utils.APIErrorHandler(err)
			} else {
				tpFltr := engine.FilterToTPFilter(fltr)
				sppIndexers.IndexTPFilter(tpFltr, spp.ID)
			}
		}
	}
	if args.SupplierIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix,
			args.Tenant, false)); err != nil {
			return utils.APIErrorHandler(err)
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix,
			args.Tenant, true), ""); err != nil {
			return utils.APIErrorHandler(err)
		}

	} else {
		indexRemover := engine.NewReqFilterIndexer(self.DataManager, utils.SupplierProfilePrefix, args.Tenant)
		for _, id := range resourceIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return utils.APIErrorHandler(err)
			}
		}
	}
	if err := sppIndexers.StoreIndexes(); err != nil {
		return err
	}
	//AttributeProfile Indexes
	var attributeIDs []string
	attrIndexers := engine.NewReqFilterIndexer(self.DataManager, utils.AttributeProfilePrefix, args.Tenant)
	if args.AttributeIDs == nil {
		ids, err := self.DataManager.DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, id := range ids {
			err = attrIndexers.RemoveItemFromIndex(strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
			if err != nil && err.Error() != utils.ErrNotFound.Error() {
				return utils.APIErrorHandler(err)
			}
			attributeIDs = append(attributeIDs, strings.Split(id, utils.CONCATENATED_KEY_SEP)[1])
		}
	} else {
		attributeIDs = *args.AttributeIDs
	}
	for _, id := range attributeIDs {
		ap, err := self.DataManager.GetAttributeProfile(args.Tenant, id, false, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		for _, fltrID := range ap.FilterIDs {
			fltr, err := self.DataManager.GetFilter(args.Tenant, fltrID, false, utils.NonTransactional)
			if err != nil {
				if err == utils.ErrNotFound {
					err = fmt.Errorf("broken reference to filter: %+v for stats queue: %+v", fltrID, ap)
				}
				return utils.APIErrorHandler(err)
			} else {
				tpFltr := engine.FilterToTPFilter(fltr)
				attrIndexers.IndexTPFilter(tpFltr, ap.ID)

			}
		}
	}
	if args.AttributeIDs == nil {
		if err := self.DataManager.RemoveFilterIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix,
			args.Tenant, false)); err != nil {
			return utils.APIErrorHandler(err)
		}
		if err := self.DataManager.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix,
			args.Tenant, true), ""); err != nil {
			return utils.APIErrorHandler(err)
		}

	} else {
		indexRemover := engine.NewReqFilterIndexer(self.DataManager, utils.AttributeProfilePrefix, args.Tenant)
		for _, id := range attributeIDs {
			if err := indexRemover.RemoveItemFromIndex(id); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return utils.APIErrorHandler(err)
			}
		}
	}
	if err := attrIndexers.StoreIndexes(); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}
