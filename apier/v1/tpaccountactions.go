// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new AccountActions profile within a tariff plan
func (self *APIerSv1) SetTPAccountActions(attrs utils.TPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "LoadId", "Tenant", "Account", "ActionPlanId"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPAccountActions([]*utils.TPAccountActions{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPAccountActionsByLoadId struct {
	TPid   string // Tariff plan id
	LoadId string // AccountActions id
}

// Queries specific AccountActions profile on tariff plan
func (self *APIerSv1) GetTPAccountActionsByLoadId(attrs utils.TPAccountActions, reply *[]*utils.TPAccountActions) error {
	mndtryFlds := []string{"TPid", "LoadId"}
	if len(attrs.Account) != 0 { // If account provided as filter, make all related fields mandatory
		mndtryFlds = append(mndtryFlds, "Tenant", "Account")
	}
	if missing := utils.MissingStructFields(&attrs, mndtryFlds); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if aas, err := self.StorDb.GetTPAccountActions(&attrs); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = aas
	}
	return nil
}

type AttrGetTPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // DerivedCharge id
}

// Queries specific DerivedCharge on tariff plan
func (self *APIerSv1) GetTPAccountActions(attrs AttrGetTPAccountActions, reply *utils.TPAccountActions) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "AccountActionsId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	filter := &utils.TPAccountActions{TPid: attrs.TPid}
	if err := filter.SetAccountActionsId(attrs.AccountActionsId); err != nil {
		return err
	}
	if aas, err := self.StorDb.GetTPAccountActions(filter); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *aas[0]
	}
	return nil
}

type AttrGetTPAccountActionIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// Queries AccountActions identities on specific tariff plan.
func (self *APIerSv1) GetTPAccountActionLoadIds(attrs AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAccountActions,
		utils.TPDistinctIds{"loadid"}, nil, &attrs.PaginatorWithSearch); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Queries DerivedCharges identities on specific tariff plan.
func (self *APIerSv1) GetTPAccountActionIds(attrs AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAccountActions,
		utils.TPDistinctIds{"loadid", "tenant", "account"}, nil, &attrs.PaginatorWithSearch); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific AccountActions on Tariff plan
func (self *APIerSv1) RemoveTPAccountActions(attrs AttrGetTPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aa := engine.TpAccountAction{Tpid: attrs.TPid}
	if err := aa.SetAccountActionId(attrs.AccountActionsId); err != nil {
		return err
	}
	if err := self.StorDb.RemTpData(utils.TBLTPAccountActions, aa.Tpid,
		map[string]string{"loadid": aa.Loadid, "tenant": aa.Tenant, "account": aa.Account}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
