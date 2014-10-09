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

package v1

import (
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

// Creates a new ActionTriggers profile within a tariff plan
func (self *ApierV1) SetTPActionTriggers(attrs utils.TPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "ActionTriggersId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	ats := map[string][]*utils.TPActionTrigger{
		attrs.ActionTriggersId: attrs.ActionTriggers}

	if err := self.StorDb.SetTPActionTriggers(attrs.TPid, ats); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	}
	*reply = "OK"
	return nil
}

type AttrGetTPActionTriggers struct {
	TPid             string // Tariff plan id
	ActionTriggersId string // ActionTrigger id
}

// Queries specific ActionTriggers profile on tariff plan
func (self *ApierV1) GetTPActionTriggers(attrs AttrGetTPActionTriggers, reply *utils.TPActionTriggers) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTriggersId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if atsMap, err := self.StorDb.GetTpActionTriggers(attrs.TPid, attrs.ActionTriggersId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(atsMap) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		atRply := &utils.TPActionTriggers{attrs.TPid, attrs.ActionTriggersId, atsMap[attrs.ActionTriggersId]}
		*reply = *atRply
	}
	return nil
}

type AttrGetTPActionTriggerIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries ActionTriggers identities on specific tariff plan.
func (self *ApierV1) GetTPActionTriggerIds(attrs AttrGetTPActionTriggerIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPTableIds(attrs.TPid, utils.TBL_TP_ACTION_TRIGGERS, utils.TPDistinctIds{"id"}, nil, &attrs.Paginator); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific ActionTriggers on Tariff plan
func (self *ApierV1) RemTPActionTriggers(attrs AttrGetTPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTriggersId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if err := self.StorDb.RemTPData(utils.TBL_TP_ACTION_TRIGGERS, attrs.TPid, attrs.ActionTriggersId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else {
		*reply = "OK"
	}
	return nil
}
