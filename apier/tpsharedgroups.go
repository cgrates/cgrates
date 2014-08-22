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

	"github.com/cgrates/cgrates/utils"
)

// Creates a new SharedGroups profile within a tariff plan
func (self *ApierV1) SetTPSharedGroups(attrs utils.TPSharedGroups, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "SharedGroupsId", "SharedGroups"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	/*for _, action := range attrs.SharedGroups {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:SharedGroup:%s:%v", utils.ERR_MANDATORY_IE_MISSING, action.Identifier, missing)
		}
	}*/
	if err := self.StorDb.SetTPSharedGroups(attrs.TPid, map[string][]*utils.TPSharedGroup{attrs.SharedGroupsId: attrs.SharedGroups}); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPSharedGroups struct {
	TPid           string // Tariff plan id
	SharedGroupsId string // SharedGroup id
}

// Queries specific SharedGroup on tariff plan
func (self *ApierV1) GetTPSharedGroups(attrs AttrGetTPSharedGroups, reply *utils.TPSharedGroups) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "SharedGroupsId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if sgs, err := self.StorDb.GetTpSharedGroups(attrs.TPid, attrs.SharedGroupsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(sgs) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = utils.TPSharedGroups{TPid: attrs.TPid, SharedGroupsId: attrs.SharedGroupsId, SharedGroups: sgs[attrs.SharedGroupsId]}
	}
	return nil
}

type AttrGetTPSharedGroupIds struct {
	TPid string // Tariff plan id
}

// Queries SharedGroups identities on specific tariff plan.
func (self *ApierV1) GetTPSharedGroupIds(attrs AttrGetTPSharedGroupIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_SHARED_GROUPS, "id", nil); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific SharedGroups on Tariff plan
func (self *ApierV1) RemTPSharedGroups(attrs AttrGetTPSharedGroups, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "SharedGroupsId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_SHARED_GROUPS, attrs.TPid, attrs.SharedGroupsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
