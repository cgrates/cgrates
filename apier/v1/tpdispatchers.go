// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPDispatcherProfile creates a new DispatcherProfile within a tariff plan
func (apierSv1 *APIerSv1) SetTPDispatcherProfile(ctx *context.Context, attr *utils.TPDispatcherProfile, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPDispatcherProfiles([]*utils.TPDispatcherProfile{attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPDispatcherProfile queries specific DispatcherProfile on Tariff plan
func (apierSv1 *APIerSv1) GetTPDispatcherProfile(ctx *context.Context, attr *utils.TPTntID, reply *utils.TPDispatcherProfile) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	rls, err := apierSv1.StorDb.GetTPDispatcherProfiles(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rls[0]
	return nil
}

type AttrGetTPDispatcherIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPDispatcherProfileIDs queries dispatcher identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPDispatcherProfileIDs(ctx *context.Context, attrs *AttrGetTPDispatcherIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDispatchers, utils.TPDistinctIds{utils.TenantCfg, utils.IDCfg},
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

// RemoveTPDispatcherProfile removes specific DispatcherProfile on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPDispatcherProfile(ctx *context.Context, attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPDispatchers, attrs.TPid,
		map[string]string{utils.TenantCfg: attrs.Tenant, utils.IDCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// SetTPDispatcherHost creates a new DispatcherHost within a tariff plan
func (apierSv1 *APIerSv1) SetTPDispatcherHost(ctx *context.Context, attr *utils.TPDispatcherHost, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPDispatcherHosts([]*utils.TPDispatcherHost{attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPDispatcherHost queries specific DispatcherHosts on Tariff plan
func (apierSv1 *APIerSv1) GetTPDispatcherHost(ctx *context.Context, attr *utils.TPTntID, reply *utils.TPDispatcherHost) error {
	if missing := utils.MissingStructFields(attr, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	if attr.Tenant == utils.EmptyString {
		attr.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	rls, err := apierSv1.StorDb.GetTPDispatcherHosts(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rls[0]
	return nil
}

// GetTPDispatcherHostIDs queries dispatcher host identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPDispatcherHostIDs(ctx *context.Context, attrs *AttrGetTPDispatcherIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDispatcherHosts, utils.TPDistinctIds{utils.TenantCfg, utils.IDCfg},
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

// RemoveTPDispatcherHost removes specific DispatcherHost on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPDispatcherHost(ctx *context.Context, attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPDispatcherHosts, attrs.TPid,
		map[string]string{utils.TenantCfg: attrs.Tenant, utils.IDCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
