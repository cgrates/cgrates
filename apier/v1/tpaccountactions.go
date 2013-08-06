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
func (self *ApierV1) SetTPAccountActions(attrs utils.ApiTPAccountActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "AccountActionsId", "Tenant", "Account", "Direction", "ActionTimingsId", "ActionTriggersId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if exists, err := self.StorDb.ExistsTPAccountActions(attrs.TPid, attrs.AccountActionsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	aa := map[string]*engine.AccountAction{
		attrs.AccountActionsId: &engine.AccountAction{Tenant: attrs.Tenant, Account: attrs.Account, Direction: attrs.Direction,
			ActionTimingsTag: attrs.ActionTimingsId, ActionTriggersTag: attrs.ActionTriggersId},
	}

	if err := self.StorDb.SetTPAccountActions(attrs.TPid, aa); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // AccountActions id
}

// Queries specific AccountActions profile on tariff plan
func (self *ApierV1) GetTPAccountActions(attrs AttrGetTPAccountActions, reply *utils.ApiTPAccountActions) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "AccountActionsId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if aa, err := self.StorDb.GetTpAccountActions(attrs.TPid, attrs.AccountActionsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(aa) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = utils.ApiTPAccountActions{TPid: attrs.TPid,
			AccountActionsId: attrs.AccountActionsId,
			Tenant:           aa[attrs.AccountActionsId].Tenant,
			Account:          aa[attrs.AccountActionsId].Account,
			Direction:        aa[attrs.AccountActionsId].Direction,
			ActionTimingsId:  aa[attrs.AccountActionsId].ActionTimingsTag,
			ActionTriggersId: aa[attrs.AccountActionsId].ActionTriggersTag}
	}
	return nil
}

type AttrGetTPAccountActionIds struct {
	TPid string // Tariff plan id
}

// Queries AccountActions identities on specific tariff plan.
func (self *ApierV1) GetTPAccountActionIds(attrs AttrGetTPAccountActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPAccountActionIds(attrs.TPid); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}
