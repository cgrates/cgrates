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
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

// Creates a new Actions profile within a tariff plan
func (self *ApierV1) SetTPActions(attrs utils.TPActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionsId", "Actions"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	for _, action := range attrs.Actions {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ERR_MANDATORY_IE_MISSING, action.Identifier, missing)
		}
	}
	if err := self.StorDb.SetTPActions(attrs.TPid, map[string][]*utils.TPAction{attrs.ActionsId: attrs.Actions}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPActions struct {
	TPid      string // Tariff plan id
	ActionsId string // Actions id
}

// Queries specific Actions profile on tariff plan
func (self *ApierV1) GetTPActions(attrs AttrGetTPActions, reply *utils.TPActions) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionsId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if acts, err := self.StorDb.GetTpActions(attrs.TPid, attrs.ActionsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(acts) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = utils.TPActions{TPid: attrs.TPid, ActionsId: attrs.ActionsId, Actions: acts[attrs.ActionsId]}
	}
	return nil
}

type AttrGetTPActionIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries Actions identities on specific tariff plan.
func (self *ApierV1) GetTPActionIds(attrs AttrGetTPActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_ACTIONS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Actions on Tariff plan
func (self *ApierV1) RemTPActions(attrs AttrGetTPActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionsId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_ACTIONS, attrs.TPid, attrs.ActionsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
