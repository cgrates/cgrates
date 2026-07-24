// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPSupplierProfile creates a new SupplierProfile within a tariff plan
func (api *APIerSv1) SetTPSupplierProfile(attrs *utils.TPSupplierProfile, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPSuppliers([]*utils.TPSupplierProfile{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPSupplierProfile queries specific SupplierProfile on tariff plan
func (api *APIerSv1) GetTPSupplierProfile(attr *utils.TPTntID, reply *utils.TPSupplierProfile) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	spp, err := api.StorDb.GetTPSuppliers(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *spp[0]
	return nil
}

type AttrGetTPSupplierProfileIDs struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPSupplierProfileIDs queries SupplierProfile identities on specific tariff plan.
func (api *APIerSv1) GetTPSupplierProfileIDs(attrs *AttrGetTPSupplierProfileIDs, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPSuppliers,
		utils.TPDistinctIds{"tenant", "id"}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPSupplierProfile removes specific SupplierProfile on Tariff plan
func (api *APIerSv1) RemoveTPSupplierProfile(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.RemTpData(utils.TBLTPSuppliers, attrs.TPid,
		map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
