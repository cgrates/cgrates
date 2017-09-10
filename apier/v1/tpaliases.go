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

// Creates a new alias within a tariff plan
func (self *ApierV1) SetTPAlias(attrs utils.TPAliases, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject", "Group"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPAliases([]*utils.TPAliases{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPAlias struct {
	TPid    string // Tariff plan id
	AliasId string
}

// Queries specific Alias on Tariff plan
func (self *ApierV1) GetTPAlias(attr AttrGetTPAlias, reply *utils.TPAliases) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	filter := &utils.TPAliases{TPid: attr.TPid}
	filter.SetId(attr.AliasId)
	if as, err := self.StorDb.GetTPAliases(filter); err != nil {
		return utils.NewErrServerError(err)
	} else if len(as) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = *as[0]
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
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAliases, utils.TPDistinctIds{"direction", "tenant", "category", "account", "subject", "context"}, nil, &attrs.Paginator); err != nil {
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
	if err := self.StorDb.RemTpData(utils.TBLTPAliases, attrs.TPid, map[string]string{"tag": attrs.AliasId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
