// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// SetTPAccountActions creates a new AccountActions profile within a tariff plan
func (apierSv1 *APIerSv1) SetTPAccountActions(ctx *context.Context, attrs *utils.TPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(attrs,
		[]string{utils.TPid, utils.LoadId, utils.AccountField, utils.ActionPlanId}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.StorDb.SetTPAccountActions([]*utils.TPAccountActions{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPAccountActionsByLoadId struct {
	TPid   string // Tariff plan id
	LoadId string // AccountActions id
}

// GetTPAccountActionsByLoadId queries specific AccountActions profile on tariff plan
func (apierSv1 *APIerSv1) GetTPAccountActionsByLoadId(ctx *context.Context, attrs *utils.TPAccountActions, reply *[]*utils.TPAccountActions) error {
	mndtryFlds := []string{utils.TPid, utils.LoadId}
	if len(attrs.Account) != 0 { // If account provided as filter, make all related fields mandatory
		mndtryFlds = append(mndtryFlds, utils.AccountField)
	}
	if missing := utils.MissingStructFields(attrs, mndtryFlds); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Tenant == utils.EmptyString {
		attrs.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	aas, err := apierSv1.StorDb.GetTPAccountActions(attrs)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = aas
	return nil
}

type AttrGetTPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // DerivedCharge id
}

// GetTPAccountActions queries specific DerivedCharge on tariff plan
func (apierSv1 *APIerSv1) GetTPAccountActions(ctx *context.Context, attrs *AttrGetTPAccountActions, reply *utils.TPAccountActions) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.AccountActionsId}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	filter := &utils.TPAccountActions{TPid: attrs.TPid}
	if err := filter.SetAccountActionsId(attrs.AccountActionsId); err != nil {
		return err
	}
	aas, err := apierSv1.StorDb.GetTPAccountActions(filter)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *aas[0]
	return nil
}

type AttrGetTPAccountActionIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPAccountActionLoadIds queries AccountActions identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPAccountActionLoadIds(ctx *context.Context, attrs *AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAccountActions,
		utils.TPDistinctIds{utils.Loadid}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// GetTPAccountActionIds queries DerivedCharges identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPAccountActionIds(ctx *context.Context, attrs *AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAccountActions,
		utils.TPDistinctIds{utils.Loadid, utils.TenantCfg, utils.AccountLowerCase}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPAccountActions removes specific AccountActions on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPAccountActions(ctx *context.Context, attrs *AttrGetTPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.AccountActionsId}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aa := engine.AccountActionMdl{Tpid: attrs.TPid}
	if err := aa.SetAccountActionId(attrs.AccountActionsId); err != nil {
		return err
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPAccountActions, aa.Tpid,
		map[string]string{utils.Loadid: aa.Loadid, utils.TenantCfg: aa.Tenant, utils.AccountLowerCase: aa.Account}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
