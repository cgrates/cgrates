// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SetTPActionTriggers creates a new ActionTriggers profile within a tariff plan
func (apierSv1 *APIerSv1) SetTPActionTriggers(ctx *context.Context, attrs *utils.TPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPActionTriggers([]*utils.TPActionTriggers{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// AttrGetTPActionTriggers used as argument for GetTPActionTriggers
type AttrGetTPActionTriggers struct {
	TPid string // Tariff plan id
	ID   string // ActionTrigger id
}

// GetTPActionTriggers queries specific ActionTriggers profile on tariff plan
func (apierSv1 *APIerSv1) GetTPActionTriggers(ctx *context.Context, attrs *AttrGetTPActionTriggers, reply *utils.TPActionTriggers) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ats, err := apierSv1.StorDb.GetTPActionTriggers(attrs.TPid, attrs.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *ats[0]
	return nil
}

// AttrGetTPActionTriggerIds used as argument for GetTPActionTriggerIds and RemoveTPActionTriggers
type AttrGetTPActionTriggerIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPActionTriggerIds queries ActionTriggers identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPActionTriggerIds(ctx *context.Context, attrs *AttrGetTPActionTriggerIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPActionTriggers,
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

// RemoveTPActionTriggers removes specific ActionTriggers on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPActionTriggers(ctx *context.Context, attrs *AttrGetTPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	err := apierSv1.StorDb.RemTpData(utils.TBLTPActionTriggers,
		attrs.TPid, map[string]string{utils.TagCfg: attrs.ID})
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
