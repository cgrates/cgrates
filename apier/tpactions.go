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
	"time"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cgrates/engine"
)

// Creates a new Actions profile within a tariff plan
func (self *Apier) SetTPActions(attrs utils.TPActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionsId", "Actions"}); len(missing) != 0 {
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	for _, action := range attrs.Actions {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units", "ExpiryTime")
		}
		if missing := utils.MissingStructFields(&action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ERR_MANDATORY_IE_MISSING, action.Identifier, missing)
		}
	}
	if exists, err := self.StorDb.ExistsTPActions(attrs.TPid, attrs.ActionsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if exists {
		return errors.New(utils.ERR_DUPLICATE)
	}
	acts := make([]*engine.Action, len(attrs.Actions))
	for idx, act := range attrs.Actions {
		acts[idx] = &engine.Action{
			ActionType:		act.Identifier,
			BalanceId: 	act.BalanceType,
			Direction:       act.Direction,
			Units:          act.Units,
			ExpirationDate: time.Unix(act.ExpiryTime,0),
			DestinationTag: act.DestinationId,
			RateType:      act.RateType,
			RateValue:     act.Rate,
			MinutesWeight: act.MinutesWeight,
			Weight:        act.Weight,
		}
	}
	if err := self.StorDb.SetTPActions(attrs.TPid, map[string][]*engine.Action{attrs.ActionsId: acts}); err != nil {
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
func (self *Apier) GetTPActions(attrs AttrGetTPActions, reply *utils.TPActions) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionsId"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if acts, err := self.StorDb.GetTPActions(attrs.TPid, attrs.ActionsId); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if acts == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = *acts
	}
	return nil
}

type AttrGetTPActionIds struct {
	TPid string // Tariff plan id
}

// Queries Actions identities on specific tariff plan.
func (self *Apier) GetTPActionIds(attrs AttrGetTPActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return fmt.Errorf("%s:%v", utils.ERR_MANDATORY_IE_MISSING, missing)
	}
	if ids, err := self.StorDb.GetTPActionIds(attrs.TPid); err != nil {
		return fmt.Errorf("%s:%s", utils.ERR_SERVER_ERROR, err.Error())
	} else if ids == nil {
		return errors.New(utils.ERR_NOT_FOUND)
	} else {
		*reply = ids
	}
	return nil
}
