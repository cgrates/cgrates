// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/utils"
)

// Creates a new SharedGroups profile within a tariff plan
func (self *APIerSv1) SetTPSharedGroups(attrs utils.TPSharedGroups, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID", "SharedGroups"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPSharedGroups([]*utils.TPSharedGroups{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPSharedGroups struct {
	TPid string // Tariff plan id
	ID   string // SharedGroup id
}

// Queries specific SharedGroup on tariff plan
func (self *APIerSv1) GetTPSharedGroups(attrs AttrGetTPSharedGroups, reply *utils.TPSharedGroups) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sgs, err := self.StorDb.GetTPSharedGroups(attrs.TPid, attrs.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *sgs[0]
	}
	return nil
}

type AttrGetTPSharedGroupIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// Queries SharedGroups identities on specific tariff plan.
func (self *APIerSv1) GetTPSharedGroupIds(attrs AttrGetTPSharedGroupIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPSharedGroups,
		utils.TPDistinctIds{"tag"}, nil, &attrs.PaginatorWithSearch); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific SharedGroups on Tariff plan
func (self *APIerSv1) RemoveTPSharedGroups(attrs AttrGetTPSharedGroups, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPSharedGroups, attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
