/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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

// Creates a new alias within a tariff plan
func (self *ApierV1) SetTPAlias(attrs utils.AttrSetTPAlias, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject", "Group"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tm := engine.APItoModelAliases(&attrs)
	if err := self.StorDb.SetTpAliases([]engine.TpAlias{*tm}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPAlias struct {
	TPid      string // Tariff plan id
	Direction string
	Tenant    string
	Category  string
	Account   string
	Subject   string
	Group     string
}

// Queries specific Alias on Tariff plan
func (self *ApierV1) GetTPAlias(attr AttrGetTPAlias, reply *engine.Alias) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	al := &engine.TpAlias{
		Tpid:      attr.TPid,
		Direction: attr.Direction,
		Tenant:    attr.Tenant,
		Category:  attr.Category,
		Account:   attr.Account,
		Subject:   attr.Subject,
		Group:     attr.Group,
	}
	if tms, err := self.StorDb.GetTpAliases(al); err != nil {
		return utils.NewErrServerError(err)
	} else if len(tms) == 0 {
		return utils.ErrNotFound
	} else {
		tmMap, err := engine.TpAliases(tms).GetAliases()
		if err != nil {
			return err
		}
		*reply = *tmMap[al.GetId()]
	}
	return nil
}

type AttrGetTPAliasIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries alias identities on specific tariff plan.
func (self *ApierV1) GetTPAliasIds(attrs AttrGetTPAliasIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_ALIASES, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Alias on Tariff plan
func (self *ApierV1) RemTPAlias(attrs AttrGetTPAlias, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "AliasId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_ALIASES, attrs.TPid); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
