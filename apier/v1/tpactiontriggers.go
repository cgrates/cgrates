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

// Creates a new ActionTriggers profile within a tariff plan
func (self *ApierV1) SetTPActionTriggers(attrs utils.TPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "ActionTriggersId"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	ats := engine.APItoModelActionTrigger(&attrs)
	if err := self.StorDb.SetTpActionTriggers(ats); err != nil {
		return utils.NewErrServerError(err)
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
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ats, err := self.StorDb.GetTpActionTriggers(attrs.TPid, attrs.ActionTriggersId); err != nil {
		return utils.NewErrServerError(err)
	} else if len(ats) == 0 {
		return utils.ErrNotFound
	} else {
		atsMap, err := engine.TpActionTriggers(ats).GetActionTriggers()
		if err != nil {
			return err
		}
		atRply := &utils.TPActionTriggers{
			TPid:             attrs.TPid,
			ActionTriggersId: attrs.ActionTriggersId,
			ActionTriggers:   atsMap[attrs.ActionTriggersId],
		}
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
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_ACTION_TRIGGERS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific ActionTriggers on Tariff plan
func (self *ApierV1) RemTPActionTriggers(attrs AttrGetTPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTriggersId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_ACTION_TRIGGERS, attrs.TPid, map[string]string{"tag": attrs.ActionTriggersId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
