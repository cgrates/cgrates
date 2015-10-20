/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new AccountActions profile within a tariff plan
func (self *ApierV1) SetTPAccountActions(attrs utils.TPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "LoadId", "Tenant", "Account", "Direction", "ActionPlanId", "ActionTriggersId"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aas := engine.APItoModelAccountAction(&attrs)
	if err := self.StorDb.SetTpAccountActions([]engine.TpAccountAction{*aas}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPAccountActionsByLoadId struct {
	TPid   string // Tariff plan id
	LoadId string // AccountActions id
}

// Queries specific AccountActions profile on tariff plan
func (self *ApierV1) GetTPAccountActionsByLoadId(attrs utils.TPAccountActions, reply *[]*utils.TPAccountActions) error {
	mndtryFlds := []string{"TPid", "LoadId"}
	if len(attrs.Account) != 0 { // If account provided as filter, make all related fields mandatory
		mndtryFlds = append(mndtryFlds, "Tenant", "Account", "Direction")
	}
	if missing := utils.MissingStructFields(&attrs, mndtryFlds); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aas := engine.APItoModelAccountAction(&attrs)
	if aa, err := self.StorDb.GetTpAccountActions(aas); err != nil {
		return utils.NewErrServerError(err)
	} else if len(aa) == 0 {
		return utils.ErrNotFound
	} else {

		tpAa, err := engine.TpAccountActions(aa).GetAccountActions()
		if err != nil {
			return err
		}
		var acts []*utils.TPAccountActions
		if len(attrs.Account) != 0 {
			acts = []*utils.TPAccountActions{tpAa[attrs.KeyId()]}
		} else {
			for _, actLst := range tpAa {
				acts = append(acts, actLst)
			}
		}
		*reply = acts
	}
	return nil
}

type AttrGetTPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // DerivedCharge id
}

// Queries specific DerivedCharge on tariff plan
func (self *ApierV1) GetTPAccountActions(attrs AttrGetTPAccountActions, reply *utils.TPAccountActions) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "AccountActionsId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tmpAa := &utils.TPAccountActions{TPid: attrs.TPid}
	if err := tmpAa.SetAccountActionsId(attrs.AccountActionsId); err != nil {
		return err
	}
	tmpAaa := engine.APItoModelAccountAction(tmpAa)
	if aas, err := self.StorDb.GetTpAccountActions(tmpAaa); err != nil {
		return utils.NewErrServerError(err)
	} else if len(aas) == 0 {
		return utils.ErrNotFound
	} else {
		tpAaa, err := engine.TpAccountActions(aas).GetAccountActions()
		if err != nil {
			return err
		}
		aa := tpAaa[tmpAa.KeyId()]
		tpdc := utils.TPAccountActions{
			TPid:             attrs.TPid,
			ActionPlanId:     aa.ActionPlanId,
			ActionTriggersId: aa.ActionTriggersId,
		}
		if err := tpdc.SetAccountActionsId(attrs.AccountActionsId); err != nil {
			return err
		}
		*reply = tpdc
	}
	return nil
}

type AttrGetTPAccountActionIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries AccountActions identities on specific tariff plan.
func (self *ApierV1) GetTPAccountActionLoadIds(attrs AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_ACCOUNT_ACTIONS, utils.TPDistinctIds{"loadid"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Queries DerivedCharges identities on specific tariff plan.
func (self *ApierV1) GetTPAccountActionIds(attrs AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_ACCOUNT_ACTIONS, utils.TPDistinctIds{"loadid", "direction", "tenant", "account"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific AccountActions on Tariff plan
func (self *ApierV1) RemTPAccountActions(attrs AttrGetTPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LoadId", "Tenant", "Account", "Direction"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aa := engine.TpAccountAction{Tpid: attrs.TPid}
	if err := aa.SetAccountActionId(attrs.AccountActionsId); err != nil {
		return err
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_ACCOUNT_ACTIONS, aa.Tpid, map[string]string{"loadid": aa.Loadid, "tenant": aa.Tenant, "account": aa.Account}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
