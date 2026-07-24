// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPCharger creates a new ChargerProfile within a tariff plan
func (api *APIerSv1) SetTPCharger(attr *utils.TPChargerProfile, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPChargers([]*utils.TPChargerProfile{attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPCharger queries specific ChargerProfile on Tariff plan
func (api *APIerSv1) GetTPCharger(attr *utils.TPTntID, reply *utils.TPChargerProfile) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rls, err := api.StorDb.GetTPChargers(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rls[0]
	return nil
}

type AttrGetTPChargerIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPChargerIDs queries Charger identities on specific tariff plan.
func (api *APIerSv1) GetTPChargerIDs(attrs *AttrGetTPChargerIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPChargers, utils.TPDistinctIds{"tenant", "id"},
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

// RemoveTPCharger removes specific ChargerProfile on Tariff plan
func (api *APIerSv1) RemoveTPCharger(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.RemTpData(utils.TBLTPChargers, attrs.TPid,
		map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
