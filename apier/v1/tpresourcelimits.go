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

// Creates a new resource limit within a tariff plan
func (self *ApierV1) SetTPResourceLimit(attr utils.TPResourceLimit, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "ID", "Limit"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPResourceLimits([]*utils.TPResourceLimit{&attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPResourceLimit struct {
	TPid string // Tariff plan id
	ID   string
}

// Queries specific ResourceLimit on Tariff plan
func (self *ApierV1) GetTPResourceLimit(attr AttrGetTPResourceLimit, reply *utils.TPResourceLimit) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rls, err := self.StorDb.GetTPResourceLimits(attr.TPid, attr.ID); err != nil {
		return utils.NewErrServerError(err)
	} else if len(rls) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = *rls[0]
	}
	return nil
}

type AttrGetTPResourceLimitIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries ResourceLimit identities on specific tariff plan.
func (self *ApierV1) GetTPResourceLimitIDs(attrs AttrGetTPResourceLimitIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPResourceLimits, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific ResourceLimit on Tariff plan
func (self *ApierV1) RemTPResourceLimit(attrs AttrGetTPResourceLimit, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPResourceLimits, attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil

}
