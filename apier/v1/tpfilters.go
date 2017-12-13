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

// Creates a new FilterProfile within a tariff plan
func (self *ApierV1) SetTPFilterProfile(attrs utils.TPFilterProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPFilters([]*utils.TPFilterProfile{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPFilterProfile struct {
	TPid string // Tariff plan id
	ID   string // Filter id
}

// Queries specific FilterProfile on tariff plan
func (self *ApierV1) GetTPFilterProfile(attr AttrGetTPFilterProfile, reply *utils.TPFilterProfile) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if filter, err := self.StorDb.GetTPFilters(attr.TPid, attr.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *filter[0]
	}
	return nil
}

type AttrGetTPFilterProfileIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries FilterProfile identities on specific tariff plan.
func (self *ApierV1) GetTPFilterProfileIds(attrs AttrGetTPFilterProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPFilters, utils.TPDistinctIds{"id"}, nil, &attrs.Paginator); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

type AttrRemTPFilterProfile struct {
	TPid   string // Tariff plan id
	Tenant string
	ID     string // Filter id
}

// Removes specific FilterProfile on Tariff plan
func (self *ApierV1) RemTPFilterProfile(attrs AttrRemTPFilterProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPFilters, attrs.TPid, map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
