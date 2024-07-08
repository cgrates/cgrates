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

// SetTPRankings creates a new ranking within a tariff plan
func (apierSv1 *APIerSv1) SetTPRanking(ctx *context.Context, sag *utils.TPRankingProfile, reply *string) error {
	if missing := utils.MissingStructFields(sag, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sag.Tenant == utils.EmptyString {
		sag.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPRankings([]*utils.TPRankingProfile{sag}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPRanking queries specific Ranking on Tariff plan
func (apierSv1 *APIerSv1) GetTPRanking(ctx *context.Context, ranking *utils.TPTntID, reply *utils.TPRankingProfile) error {
	if missing := utils.MissingStructFields(ranking, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ranking.Tenant == utils.EmptyString {
		ranking.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	sgs, err := apierSv1.StorDb.GetTPRankings(ranking.TPid, ranking.Tenant, ranking.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *sgs[0]
	return nil
}

type AttrGetTPRankingIds struct {
	TPid   string // Tariff plan id
	Tenant string
	utils.PaginatorWithSearch
}

// GetTPRankingIDs queries Ranking identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPRankingIDs(ctx *context.Context, attrs *AttrGetTPRankingIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRankings,
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

// RemoveTPRanking removes specific Ranking on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPRanking(ctx *context.Context, sag *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(sag, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sag.Tenant == utils.EmptyString {
		sag.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPRankings, sag.TPid,
		map[string]string{utils.TenantCfg: sag.Tenant, utils.IDCfg: sag.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil

}
