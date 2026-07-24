// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/utils"
)

// SetTPActionTriggers creates a new ActionTriggers profile within a tariff plan
func (api *APIerSv1) SetTPActionTriggers(attrs utils.TPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPActionTriggers([]*utils.TPActionTriggers{&attrs}); err != nil {
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
func (api *APIerSv1) GetTPActionTriggers(attrs AttrGetTPActionTriggers, reply *utils.TPActionTriggers) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ats, err := api.StorDb.GetTPActionTriggers(attrs.TPid, attrs.ID)
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
func (api *APIerSv1) GetTPActionTriggerIds(attrs AttrGetTPActionTriggerIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPActionTriggers,
		utils.TPDistinctIds{"tag"}, nil, &attrs.PaginatorWithSearch)
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
func (api *APIerSv1) RemoveTPActionTriggers(attrs AttrGetTPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	err := api.StorDb.RemTpData(utils.TBLTPActionTriggers,
		attrs.TPid, map[string]string{"tag": attrs.ID})
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
