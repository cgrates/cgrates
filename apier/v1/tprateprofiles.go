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

// SetTPRateProfile creates a new TPRateProfile within a tariff plan
func (apierSv1 *APIerSv1) SetTPRateProfile(attrs *utils.TPRateProfile, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attrs.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPRateProfiles([]*utils.TPRateProfile{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPRateProfile queries specific TPRateProfile on tariff plan
func (apierSv1 *APIerSv1) GetTPRateProfile(attr *utils.TPTntID, reply *utils.TPRateProfile) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attr.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	filter, err := apierSv1.StorDb.GetTPRateProfiles(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *filter[0]
	return nil
}

type AttrGetTPRateProfileIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPRateProfileIds queries TPRateProfiles identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPRateProfileIds(attrs *AttrGetTPRateProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRateProfiles, utils.TPDistinctIds{utils.TenantCfg, utils.IDCfg},
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

// RemoveTPRateProfile removes specific TPRateProfile on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPRateProfile(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := attrs.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPRateProfiles, attrs.TPid,
		map[string]string{utils.TenantCfg: attrs.Tenant, utils.IDCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
