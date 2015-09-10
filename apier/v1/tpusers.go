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
func (self *ApierV1) SetTPUser(attrs utils.TPUsers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject", "Group"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tm := engine.APItoModelUsers(&attrs)
	if err := self.StorDb.SetTpUsers(tm); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPUser struct {
	TPid   string // Tariff plan id
	UserId string
}

// Queries specific User on Tariff plan
func (self *ApierV1) GetTPUser(attr AttrGetTPUser, reply *utils.TPUsers) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	usr := &engine.TpUser{
		Tpid: attr.TPid,
	}
	usr.SetId(attr.UserId)
	if tms, err := self.StorDb.GetTpUsers(usr); err != nil {
		return utils.NewErrServerError(err)
	} else if len(tms) == 0 {
		return utils.ErrNotFound
	} else {
		tmMap, err := engine.TpUsers(tms).GetUsers()
		if err != nil {
			return err
		}
		*reply = *tmMap[usr.GetId()]
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
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_USERS, utils.TPDistinctIds{"tenant", "user_name"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific User on Tariff plan
func (self *ApierV1) RemTPUser(attrs AttrGetTPUser, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "UserId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_USERS, attrs.TPid); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
