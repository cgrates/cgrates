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
	"github.com/cgrates/cgrates/utils"
)

// SetTPThreshold creates a new threshold within a tariff plan
func (apierSv1 *APIerSv1) SetTPThreshold(attr *utils.TPThresholdProfile, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPThresholds([]*utils.TPThresholdProfile{attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPThreshold queries specific Threshold on Tariff plan
func (apierSv1 *APIerSv1) GetTPThreshold(attr *utils.TPTntID, reply *utils.TPThresholdProfile) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	rls, err := apierSv1.StorDb.GetTPThresholds(attr.TPid, attr.Tenant, attr.ID)
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
func (apierSv1 *APIerSv1) GetTPThresholdIDs(attrs *AttrGetTPThresholdIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPThresholds,
		[]string{utils.TenantCfg, utils.IDCfg}, nil, &attrs.PaginatorWithSearch)
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
func (apierSv1 *APIerSv1) RemoveTPThreshold(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPThresholds, attrs.TPid,
		map[string]string{utils.TenantCfg: attrs.Tenant, utils.IDCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil

}
