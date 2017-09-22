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
	"github.com/cgrates/cgrates/utils"
)

// Creates a new LcrRules profile within a tariff plan
func (self *ApierV1) SetTPLcrRule(attrs utils.TPLcrRules, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPLCRs([]*utils.TPLcrRules{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPLcrRules struct {
	TPid      string // Tariff plan id
	LcrRuleId string // Lcr id
}

type AttrRemTPLcrRules struct {
	TPid      string // Tariff plan id
	Direction string
	Tenant    string
	Category  string
	Account   string
	Subject   string
}

// Queries specific LcrRules profile on tariff plan
func (self *ApierV1) GetTPLcrRule(attr AttrGetTPLcrRules, reply *utils.TPLcrRules) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "LcrRuleId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	filter := &utils.TPLcrRules{TPid: attr.TPid}
	filter.SetId(attr.LcrRuleId)
	if lcrs, err := self.StorDb.GetTPLCRs(filter); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *lcrs[0]
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
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPLcrs, utils.TPDistinctIds{"direction", "tenant", "category", "account", "subject"}, nil, &attrs.Paginator); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific LcrRules on Tariff plan
func (self *ApierV1) RemTPLcrRule(attrs AttrRemTPLcrRules, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPLcrs, attrs.TPid, map[string]string{"direction": attrs.Direction, "tenant": attrs.Tenant, "category": attrs.Category, "account": attrs.Account, "subject": attrs.Subject}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
