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
func (self *ApierV1) SetTPUser(attrs utils.TPUsers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "UserName", "Tenant"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPUsers([]*utils.TPUsers{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPUser struct {
	TPid     string // Tariff plan id
	Tenant   string
	UserName string
}

// Queries specific User on Tariff plan
func (self *ApierV1) GetTPUser(attr AttrGetTPUser, reply *utils.TPUsers) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "UserName", "Tenant"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	filter := &utils.TPUsers{TPid: attr.TPid, UserName: attr.UserName, Tenant: attr.Tenant}
	if tms, err := self.StorDb.GetTPUsers(filter); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *tms[0]
	}
	return nil
}

type AttrGetTPUserIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries alias identities on specific tariff plan.
func (self *ApierV1) GetTPUserIds(attrs AttrGetTPUserIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPUsers, utils.TPDistinctIds{"tenant", "user_name"}, nil, &attrs.Paginator); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific User on Tariff plan
func (self *ApierV1) RemTPUser(attrs AttrGetTPUser, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Tenant", "UserName"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPUsers, attrs.TPid, map[string]string{"tenant": attrs.Tenant, "user_name": attrs.UserName}); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = utils.OK
	}
	return nil
}
