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

// Creates a new LcrRules profile within a tariff plan
func (self *ApierV1) SetTPLcrRule(attrs utils.TPLcrRules, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tm := engine.APItoModelLcrRule(&attrs)
	if err := self.StorDb.SetTpLCRs(tm); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPLcrRules struct {
	TPid      string // Tariff plan id
	LcrRuleId string // Lcr id
}

// Queries specific LcrRules profile on tariff plan
func (self *ApierV1) GetTPLcrRule(attr AttrGetTPLcrRules, reply *utils.TPLcrRules) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "LcrRuleId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	lcr := &engine.TpLcrRule{
		Tpid: attr.TPid,
	}
	lcr.SetLcrRuleId(attr.LcrRuleId)
	if lcrs, err := self.StorDb.GetTpLCRs(lcr); err != nil {
		return utils.NewErrServerError(err)
	} else if len(lcrs) == 0 {
		return utils.ErrNotFound
	} else {
		tmMap, err := engine.TpLcrRules(lcrs).GetLcrRules()
		if err != nil {
			return err
		}
		*reply = *tmMap[attr.LcrRuleId]
	}
	return nil
}

type AttrGetTPLcrIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries LcrRules identities on specific tariff plan.
func (self *ApierV1) GetTPLcrRuleIds(attrs AttrGetTPLcrIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_LCRS, utils.TPDistinctIds{"direction", "tenant", "category", "account", "subject"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific LcrRules on Tariff plan
func (self *ApierV1) RemTPLcrRule(attrs AttrGetTPLcrRules, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "LcrRulesId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_LCRS, attrs.TPid, attrs.LcrRuleId); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
