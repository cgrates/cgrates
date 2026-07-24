// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/utils"
)

// Creates a new resource within a tariff plan
func (self *APIerSv1) SetTPResource(attr *utils.TPResourceProfile, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID", "Limit"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPResources([]*utils.TPResourceProfile{attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// Queries specific Resource on Tariff plan
func (self *APIerSv1) GetTPResource(attr *utils.TPTntID, reply *utils.TPResourceProfile) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rls, err := self.StorDb.GetTPResources(attr.TPid, attr.Tenant, attr.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rls[0]
	}
	return nil
}

type AttrGetTPResourceIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// Queries Resource identities on specific tariff plan.
func (self *APIerSv1) GetTPResourceIDs(attrs *AttrGetTPResourceIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPResources,
		utils.TPDistinctIds{"id"}, nil, &attrs.PaginatorWithSearch); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Resource on Tariff plan
func (self *APIerSv1) RemoveTPResource(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPResources, attrs.TPid, map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil

}
