// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPAttributeProfile creates a new AttributeProfile within a tariff plan
func (apierSv1 *APIerSv1) SetTPAttributeProfile(ctx *context.Context, attrs *utils.TPAttributeProfile, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPAttributes([]*utils.TPAttributeProfile{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPAttributeProfile queries specific AttributeProfile on Tariff plan
func (apierSv1 *APIerSv1) GetTPAttributeProfile(ctx *context.Context, attr *utils.TPTntID, reply *utils.TPAttributeProfile) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	als, err := apierSv1.StorDb.GetTPAttributes(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *als[0]
	return nil
}

type AttrGetTPAttributeProfileIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPAttributeProfileIds queries attribute identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPAttributeProfileIds(ctx *context.Context, attrs *AttrGetTPAttributeProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAttributes,
		utils.TPDistinctIds{utils.TenantCfg, utils.IDCfg}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if ids == nil {
		return utils.ErrNotFound
	}
	*reply = ids
	return nil
}

// RemoveTPAttributeProfile removes specific AttributeProfile on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPAttributeProfile(ctx *context.Context, attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPAttributes, attrs.TPid,
		map[string]string{utils.TenantCfg: attrs.Tenant, utils.IDCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
