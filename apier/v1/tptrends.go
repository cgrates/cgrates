/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPTrends creates a new trend within a tariff plan
func (apierSv1 *APIerSv1) SetTPTrend(ctx *context.Context, trend *utils.TPTrendsProfile, reply *string) error {
	if missing := utils.MissingStructFields(trend, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if trend.Tenant == utils.EmptyString {
		trend.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPTrends([]*utils.TPTrendsProfile{trend}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPTrend queries specific Trend on Tariff plan
func (apierSv1 *APIerSv1) GetTPTrend(ctx *context.Context, trend *utils.TPTntID, reply *utils.TPTrendsProfile) error {
	if missing := utils.MissingStructFields(trend, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if trend.Tenant == utils.EmptyString {
		trend.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	trs, err := apierSv1.StorDb.GetTPTrends(trend.TPid, trend.Tenant, trend.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *trs[0]
	return nil
}

type AttrGetTPTrendIds struct {
	TPid   string // Tariff plan id
	Tenant string
	utils.PaginatorWithSearch
}

// GetTPTrendIDs queries Trend indetities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPTrendIDs(ctx *context.Context, attrs *AttrGetTPTrendIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPTrends,
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

// RemoveTPTrend removes specific Trend on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPTrend(ctx *context.Context, trend *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(trend, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if trend.Tenant == utils.EmptyString {
		trend.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPTrends, trend.TPid,
		map[string]string{utils.TenantCfg: trend.Tenant, utils.IDCfg: trend.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil

}
