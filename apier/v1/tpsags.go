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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPSags creates a new stataggregator within a tariff plan
func (apierSv1 *APIerSv1) SetTPSag(ctx *context.Context, sag *utils.TPSagsProfile, reply *string) error {
	if missing := utils.MissingStructFields(sag, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sag.Tenant == utils.EmptyString {
		sag.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPSags([]*utils.TPSagsProfile{sag}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPSag queries specific Sag on Tariff plan
func (apierSv1 *APIerSv1) GetTPSag(ctx *context.Context, sag *utils.TPTntID, reply *utils.TPSagsProfile) error {
	if missing := utils.MissingStructFields(sag, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sag.Tenant == utils.EmptyString {
		sag.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	sgs, err := apierSv1.StorDb.GetTPSags(sag.TPid, sag.Tenant, sag.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *sgs[0]
	return nil
}

type AttrGetTPSagIds struct {
	TPid   string // Tariff plan id
	Tenant string
	utils.PaginatorWithSearch
}

// GetTPSagIDs queries Sag identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPSagIDs(ctx *context.Context, attrs *AttrGetTPSagIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPSags,
		utils.TPDistinctIds{utils.TenantCfg, utils.IDCfg}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPSag removes specific Sag on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPSag(ctx *context.Context, sag *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(sag, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sag.Tenant == utils.EmptyString {
		sag.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPSags, sag.TPid,
		map[string]string{utils.TenantCfg: sag.Tenant, utils.IDCfg: sag.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil

}
