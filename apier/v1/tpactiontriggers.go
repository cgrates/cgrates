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

// Creates a new ActionTriggers profile within a tariff plan
func (self *ApierV1) SetTPActionTriggers(attrs utils.ApiTPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "ActionTriggersId"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if exists, err := self.StorDb.ExistsTPActionTriggers(attrs.TPid, attrs.ActionTriggersId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	aTriggers := make([]*engine.ActionTrigger, len(attrs.ActionTriggers))
	for idx, at := range attrs.ActionTriggers {
		requiredFields := []string{"BalanceType", "Direction", "ThresholdType", "ThresholdValue", "ActionsId", "Weight"}
		if missing := utils.MissingStructFields(&at, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Balance:%s:%v", utils.ERR_MANDATORY_IE_MISSING, at.BalanceType, missing)
		}
		at := &engine.ActionTrigger{
			BalanceId:      at.BalanceType,
			Direction:      at.Direction,
			ThresholdType:  at.ThresholdType,
			ThresholdValue: at.ThresholdValue,
			DestinationId:  at.DestinationId,
			Weight:         at.Weight,
			ActionsId:      at.ActionsId,
		}
		aTriggers[idx] = at
	}

	ats := map[string][]*engine.ActionTrigger{
		attrs.ActionTriggersId: aTriggers}

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
func (self *ApierV1) GetTPActionTriggers(attrs AttrGetTPActionTriggers, reply *utils.ApiTPActionTriggers) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTriggersId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if atsMap, err := self.StorDb.GetTpActionTriggers(attrs.TPid, attrs.ActionTriggersId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if len(atsMap) == 0 {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		atRply := &utils.ApiTPActionTriggers{attrs.TPid, attrs.ActionTriggersId, atsMap[attrs.ActionTriggersId]}
		*reply = *atRply
	}
	return nil
}

type AttrGetTPActionTriggerIds struct {
	TPid string // Tariff plan id
}

// Queries ActionTriggers identities on specific tariff plan.
func (self *ApierV1) GetTPActionTriggerIds(attrs AttrGetTPActionTriggerIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPActionTriggerIds(attrs.TPid); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}
