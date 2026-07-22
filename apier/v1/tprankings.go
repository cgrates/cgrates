// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPRankings creates a new ranking within a tariff plan
func (apierSv1 *APIerSv1) SetTPRanking(ctx *context.Context, rng *utils.TPRankingProfile, reply *string) error {
	if missing := utils.MissingStructFields(rng, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rng.Tenant == utils.EmptyString {
		rng.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPRankings([]*utils.TPRankingProfile{rng}); err != nil {
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
func (apierSv1 *APIerSv1) RemoveTPRanking(ctx *context.Context, rng *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(rng, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rng.Tenant == utils.EmptyString {
		rng.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPRankings, rng.TPid,
		map[string]string{utils.TenantCfg: rng.Tenant, utils.IDCfg: rng.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil

}
