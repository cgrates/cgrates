// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPRouteProfile creates a new RouteProfile within a tariff plan
func (apierSv1 *APIerSv1) SetTPRouteProfile(ctx *context.Context, attrs *utils.TPRouteProfile, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPRoutes([]*utils.TPRouteProfile{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPRouteProfile queries specific RouteProfile on tariff plan
func (apierSv1 *APIerSv1) GetTPRouteProfile(ctx *context.Context, attr *utils.TPTntID, reply *utils.TPRouteProfile) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	spp, err := apierSv1.StorDb.GetTPRoutes(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *spp[0]
	return nil
}

type AttrGetTPRouteProfileIDs struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPRouteProfileIDs queries RouteProfile identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPRouteProfileIDs(ctx *context.Context, attrs *AttrGetTPRouteProfileIDs, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPRoutes,
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

// RemoveTPRouteProfile removes specific RouteProfile on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPRouteProfile(ctx *context.Context, attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPRoutes, attrs.TPid,
		map[string]string{utils.TenantCfg: attrs.Tenant, utils.IDCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
