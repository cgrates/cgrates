// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPFilterProfile creates a new FilterProfile within a tariff plan
func (apierSv1 *APIerSv1) SetTPFilterProfile(ctx *context.Context, attrs *utils.TPFilterProfile, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPFilters([]*utils.TPFilterProfile{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPFilterProfile queries specific FilterProfile on tariff plan
func (apierSv1 *APIerSv1) GetTPFilterProfile(ctx *context.Context, attr *utils.TPTntID, reply *utils.TPFilterProfile) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	filter, err := apierSv1.StorDb.GetTPFilters(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *filter[0]
	return nil
}

type AttrGetTPFilterProfileIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPFilterProfileIds queries FilterProfile identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPFilterProfileIds(ctx *context.Context, attrs *AttrGetTPFilterProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPFilters, utils.TPDistinctIds{utils.TenantCfg, utils.IDCfg},
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

// RemoveTPFilterProfile removes specific FilterProfile on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPFilterProfile(ctx *context.Context, attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPFilters, attrs.TPid,
		map[string]string{utils.TenantCfg: attrs.Tenant, utils.IDCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
