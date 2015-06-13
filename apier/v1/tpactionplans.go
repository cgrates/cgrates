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
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new ActionTimings profile within a tariff plan
func (self *ApierV1) SetTPActionPlan(attrs utils.TPActionPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionPlanId", "ActionPlan"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, at := range attrs.ActionPlan {
		requiredFields := []string{"ActionsId", "TimingId", "Weight"}
		if missing := utils.MissingStructFields(at, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), at.ActionsId, missing)
		}
	}
	ap := engine.APItoModelActionPlan(&attrs)
	if err := self.StorDb.SetTpActionPlans(ap); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPActionPlan struct {
	TPid string // Tariff plan id
	Id   string // ActionPlans id
}

// Queries specific ActionPlan profile on tariff plan
func (self *ApierV1) GetTPActionPlan(attrs AttrGetTPActionPlan, reply *utils.TPActionPlan) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Id"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ats, err := self.StorDb.GetTpActionPlans(attrs.TPid, attrs.Id); err != nil {
		return utils.NewErrServerError(err)
	} else if len(ats) == 0 {
		return utils.ErrNotFound
	} else { // Got the data we need, convert it
		aps, err := engine.TpActionPlans(ats).GetActionPlans()
		if err != nil {
			return err
		}
		atRply := &utils.TPActionPlan{
			TPid:         attrs.TPid,
			ActionPlanId: attrs.Id,
			ActionPlan:   aps[attrs.Id],
		}
		*reply = *atRply
	}
	return nil
}

type AttrGetTPActionPlanIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries ActionPlan identities on specific tariff plan.
func (self *ApierV1) GetTPActionPlanIds(attrs AttrGetTPActionPlanIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_ACTION_PLANS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific ActionPlan on Tariff plan
func (self *ApierV1) RemTPActionPlan(attrs AttrGetTPActionPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Id"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_ACTION_PLANS, attrs.TPid, attrs.Id); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
