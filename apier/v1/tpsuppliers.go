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
func (self *APIerSv1) SetTPSupplierProfile(attrs *utils.TPSupplierProfile, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPSuppliers([]*utils.TPSupplierProfile{attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Queries specific SupplierProfile on tariff plan
func (self *APIerSv1) GetTPSupplierProfile(attr *utils.TPTntID, reply *utils.TPSupplierProfile) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if spp, err := self.StorDb.GetTPSuppliers(attr.TPid, attr.Tenant, attr.ID); err != nil {
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
	utils.PaginatorWithSearch
}

// Queries SupplierProfile identities on specific tariff plan.
func (self *APIerSv1) GetTPSupplierProfileIDs(attrs *AttrGetTPSupplierProfileIDs, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPSuppliers,
		utils.TPDistinctIds{"id"}, nil, &attrs.PaginatorWithSearch); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific SupplierProfile on Tariff plan
func (self *APIerSv1) RemoveTPSupplierProfile(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPSuppliers, attrs.TPid,
		map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
