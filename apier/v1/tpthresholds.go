// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPThreshold creates a new threshold within a tariff plan
func (api *APIerSv1) SetTPThreshold(attr *utils.TPThresholdProfile, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPThresholds([]*utils.TPThresholdProfile{attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPThreshold queries specific Threshold on Tariff plan
func (api *APIerSv1) GetTPThreshold(attr *utils.TPTntID, reply *utils.TPThresholdProfile) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rls, err := api.StorDb.GetTPThresholds(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rls[0]
	return nil
}

type AttrGetTPThresholdIds struct {
	TPid   string // Tariff plan id
	Tenant string
	utils.PaginatorWithSearch
}

// GetTPThresholdIDs queries Threshold identities on specific tariff plan.
func (api *APIerSv1) GetTPThresholdIDs(attrs *AttrGetTPThresholdIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPThresholds,
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

// RemoveTPThreshold removes specific Threshold on Tariff plan
func (api *APIerSv1) RemoveTPThreshold(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.RemTpData(utils.TBLTPThresholds, attrs.TPid,
		map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil

}
