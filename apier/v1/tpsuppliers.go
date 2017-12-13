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

// Creates a new SupplierProfile within a tariff plan
func (self *ApierV1) SetTPSupplierProfile(attrs utils.TPSupplierProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPSuppliers([]*utils.TPSupplierProfile{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPSupplierProfile struct {
	TPid string // Tariff plan id
	ID   string // Filter id
}

// Queries specific SupplierProfile on tariff plan
func (self *ApierV1) GetTPSupplierProfile(attr AttrGetTPSupplierProfile, reply *utils.TPSupplierProfile) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if spp, err := self.StorDb.GetTPSuppliers(attr.TPid, attr.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *spp[0]
	}
	return nil
}

type AttrGetTPSupplierProfileIDs struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries SupplierProfile identities on specific tariff plan.
func (self *ApierV1) GetTPSupplierProfileIDs(attrs AttrGetTPSupplierProfileIDs, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPSuppliers, utils.TPDistinctIds{"id"}, nil, &attrs.Paginator); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

type AttrRemTPSupplierProfile struct {
	TPid   string // Tariff plan id
	Tenant string
	ID     string // LCR id
}

// Removes specific SupplierProfile on Tariff plan
func (self *ApierV1) RemTPSupplierProfile(attrs AttrRemTPSupplierProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPSuppliers, attrs.TPid, map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
