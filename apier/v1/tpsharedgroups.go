// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPSharedGroups creates a new SharedGroups profile within a tariff plan
func (apierSv1 *APIerSv1) SetTPSharedGroups(ctx *context.Context, attrs *utils.TPSharedGroups, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID, utils.SharedGroups}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPSharedGroups([]*utils.TPSharedGroups{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPSharedGroups struct {
	TPid string // Tariff plan id
	ID   string // SharedGroup id
}

// GetTPSharedGroups queries specific SharedGroup on tariff plan
func (apierSv1 *APIerSv1) GetTPSharedGroups(ctx *context.Context, attrs *AttrGetTPSharedGroups, reply *utils.TPSharedGroups) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sgs, err := apierSv1.StorDb.GetTPSharedGroups(attrs.TPid, attrs.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *sgs[0]
	return nil
}

type AttrGetTPSharedGroupIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPSharedGroupIds queries SharedGroups identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPSharedGroupIds(ctx *context.Context, attrs *AttrGetTPSharedGroupIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPSharedGroups,
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

// RemoveTPSharedGroups removes specific SharedGroups on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPSharedGroups(ctx *context.Context, attrs *AttrGetTPSharedGroups, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPSharedGroups, attrs.TPid, map[string]string{utils.TagCfg: attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
