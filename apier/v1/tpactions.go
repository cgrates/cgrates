// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPActions creates a new Actions profile within a tariff plan
func (apierSv1 *APIerSv1) SetTPActions(ctx *context.Context, attrs *utils.TPActions, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID, utils.Actions}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPActions([]*utils.TPActions{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPActions struct {
	TPid string // Tariff plan id
	ID   string // Actions id
}

// GetTPActions queries specific Actions profile on tariff plan
func (apierSv1 *APIerSv1) GetTPActions(ctx *context.Context, attrs *AttrGetTPActions, reply *utils.TPActions) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	as, err := apierSv1.StorDb.GetTPActions(attrs.TPid, attrs.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *as[0]
	return nil
}

type AttrGetTPActionIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPActionIds queries Actions identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPActionIds(ctx *context.Context, attrs *AttrGetTPActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPActions,
		utils.TPDistinctIds{utils.TagCfg}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPActions removes specific Actions on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPActions(ctx *context.Context, attrs *AttrGetTPActions, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPActions,
		attrs.TPid, map[string]string{utils.TagCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
