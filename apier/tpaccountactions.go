/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package apier

import (
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new AccountActions profile within a tariff plan
func (self *ApierV1) SetTPAccountActions(attrs utils.TPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "LoadId", "Tenant", "Account", "Direction", "ActionPlanId", "ActionTriggersId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.SetTPAccountActions(attrs.TPid, map[string]*utils.TPAccountActions{attrs.KeyId(): &attrs}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
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
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if aa, err := self.StorDb.GetTpAccountActions(&attrs); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(aa) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		var acts []*utils.TPAccountActions
		if len(attrs.Account) != 0 {
			acts = []*utils.TPAccountActions{aa[attrs.KeyId()]}
		} else {
			for _, actLst := range aa {
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
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	tmpAa := &utils.TPAccountActions{TPid: attrs.TPid}
	if err := tmpAa.SetAccountActionsId(attrs.AccountActionsId); err != nil {
		return err
	}
	if aas, err := self.StorDb.GetTpAccountActions(tmpAa); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(aas) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		aa := aas[tmpAa.KeyId()]
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
	TPid         string // Tariff plan id
	Page         int
	ItemsPerPage int
	SearchTerm   string
}

// Queries AccountActions identities on specific tariff plan.
func (self *ApierV1) GetTPAccountActionLoadIds(attrs AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_ACCOUNT_ACTIONS, utils.TPDistinctIds{"loadid"}, nil, &utils.TPPagination{Page: attrs.Page, ItemsPerPage: attrs.ItemsPerPage, SearchTerm: attrs.SearchTerm}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Queries DerivedCharges identities on specific tariff plan.
func (self *ApierV1) GetTPAccountActionIds(attrs AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_ACCOUNT_ACTIONS, utils.TPDistinctIds{"loadid", "direction", "tenant", "account"}, nil, &utils.TPPagination{Page: attrs.Page, ItemsPerPage: attrs.ItemsPerPage, SearchTerm: attrs.SearchTerm}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific AccountActions on Tariff plan
func (self *ApierV1) RemTPAccountActions(attrs AttrGetTPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LoadId", "Tenant", "Account", "Direction"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	aa := engine.TpAccountAction{Tpid: attrs.TPid}
	if err := aa.SetAccountActionId(attrs.AccountActionsId); err != nil {
		return err
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_ACCOUNT_ACTIONS, aa.Tpid, aa.Loadid, aa.Direction, aa.Tenant, aa.Account); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
