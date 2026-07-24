// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPFilterProfile creates a new FilterProfile within a tariff plan
func (api *APIerSv1) SetTPFilterProfile(attrs *utils.TPFilterProfile, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPFilters([]*utils.TPFilterProfile{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPFilterProfile queries specific FilterProfile on tariff plan
func (api *APIerSv1) GetTPFilterProfile(attr *utils.TPTntID, reply *utils.TPFilterProfile) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	filter, err := api.StorDb.GetTPFilters(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *filter[0]
	return nil
}

type AttrGetTPFilterProfileIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPFilterProfileIds queries FilterProfile identities on specific tariff plan.
func (api *APIerSv1) GetTPFilterProfileIds(attrs *AttrGetTPFilterProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPFilters, utils.TPDistinctIds{"tenant", "id"},
		nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPFilterProfile removes specific FilterProfile on Tariff plan
func (api *APIerSv1) RemoveTPFilterProfile(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.RemTpData(utils.TBLTPFilters, attrs.TPid,
		map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
