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

// SetTPSags creates a new stat within a tariff plan
func (apierSv1 *APIerSv1) SetTPSar(ctx *context.Context, sar *utils.TPSarsProfile, reply *string) error {
	if missing := utils.MissingStructFields(sar, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sar.Tenant == utils.EmptyString {
		sar.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPSars([]*utils.TPSarsProfile{sar}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPSag queries specific Stat on Tariff plan
func (apierSv1 *APIerSv1) GetTPSar(ctx *context.Context, sar *utils.TPTntID, reply *utils.TPSarsProfile) error {
	if missing := utils.MissingStructFields(sar, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sar.Tenant == utils.EmptyString {
		sar.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	srs, err := apierSv1.StorDb.GetTPSars(sar.TPid, sar.Tenant, sar.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *srs[0]
	return nil
}

type AttrGetTPSarIds struct {
	TPid   string // Tariff plan id
	Tenant string
	utils.PaginatorWithSearch
}

// GetTPSagIDs queries Stat identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPSarIDs(ctx *context.Context, attrs *AttrGetTPSarIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPSars,
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

// RemoveTPSar removes specific Sar on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPSar(ctx *context.Context, sar *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(sar, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sar.Tenant == utils.EmptyString {
		sar.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPSars, sar.TPid,
		map[string]string{utils.TenantCfg: sar.Tenant, utils.IDCfg: sar.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil

}
