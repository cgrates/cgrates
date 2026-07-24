// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPAttributeProfile creates a new AttributeProfile within a tariff plan
func (api *APIerSv1) SetTPAttributeProfile(attrs *utils.TPAttributeProfile, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPAttributes([]*utils.TPAttributeProfile{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPAttributeProfile queries specific AttributeProfile on Tariff plan
func (api *APIerSv1) GetTPAttributeProfile(attr *utils.TPTntID, reply *utils.TPAttributeProfile) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	als, err := api.StorDb.GetTPAttributes(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *als[0]
	return nil
}

type AttrGetTPAttributeProfileIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPAttributeProfileIds queries attribute identities on specific tariff plan.
func (api *APIerSv1) GetTPAttributeProfileIds(attrs *AttrGetTPAttributeProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAttributes,
		utils.TPDistinctIds{"tenant", "id"}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if ids == nil {
		return utils.ErrNotFound
	}
	*reply = ids
	return nil
}

// RemoveTPAttributeProfile removes specific AttributeProfile on Tariff plan
func (api *APIerSv1) RemoveTPAttributeProfile(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.RemTpData(utils.TBLTPAttributes, attrs.TPid,
		map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
