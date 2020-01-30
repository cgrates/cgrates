/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

	"github.com/cgrates/cgrates/utils"
)

// Creates a new ActionTimings profile within a tariff plan
func (self *APIerSv1) SetTPActionPlan(attrs utils.TPActionPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID", "ActionPlan"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, at := range attrs.ActionPlan {
		requiredFields := []string{"ActionsId", "TimingId", "Weight"}
		if missing := utils.MissingStructFields(at, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), at.ActionsId, missing)
		}
	}
	if err := self.StorDb.SetTPActionPlans([]*utils.TPActionPlan{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPActionPlan struct {
	TPid string // Tariff plan id
	ID   string // ActionPlans id
}

// Queries specific ActionPlan profile on tariff plan
func (self *APIerSv1) GetTPActionPlan(attrs AttrGetTPActionPlan, reply *utils.TPActionPlan) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if aps, err := self.StorDb.GetTPActionPlans(attrs.TPid, attrs.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *aps[0]
	}
	return nil
}

type AttrGetTPActionPlanIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// Queries ActionPlan identities on specific tariff plan.
func (self *APIerSv1) GetTPActionPlanIds(attrs AttrGetTPActionPlanIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPActionPlans,
		utils.TPDistinctIds{"tag"}, nil, &attrs.PaginatorWithSearch); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific ActionPlan on Tariff plan
func (self *APIerSv1) RemoveTPActionPlan(attrs AttrGetTPActionPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPActionPlans,
		attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
