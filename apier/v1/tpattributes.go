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

// Creates a new AttributeProfile within a tariff plan
func (self *ApierV1) SetTPAttributeProfile(attrs utils.TPAttributeProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPAttributes([]*utils.TPAttributeProfile{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPAttributeProfile struct {
	TPid string // Tariff plan id
	ID   string
}

// Queries specific AttributeProfile on Tariff plan
func (self *ApierV1) GetTPAttributeProfile(attr AttrGetTPAttributeProfile, reply *utils.TPAttributeProfile) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if als, err := self.StorDb.GetTPAttributes(attr.TPid, attr.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *als[0]
	}
	return nil
}

type AttrGetTPAttributeProfileIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries attribute identities on specific tariff plan.
func (self *ApierV1) GetTPAttributeProfileIds(attrs AttrGetTPAttributeProfileIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPAttributes, utils.TPDistinctIds{"id"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

type AttrRemTPAttributeProfile struct {
	TPid   string // Tariff plan id
	Tenant string
	ID     string // Attribute id
}

// Removes specific AttributeProfile on Tariff plan
func (self *ApierV1) RemTPAttributeProfile(attrs AttrRemTPAttributeProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPAttributes, attrs.TPid, map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
