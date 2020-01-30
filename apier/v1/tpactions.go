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

// Creates a new Actions profile within a tariff plan
func (self *APIerSv1) SetTPActions(attrs utils.TPActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID", "Actions"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPActions([]*utils.TPActions{&attrs}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPActions struct {
	TPid string // Tariff plan id
	ID   string // Actions id
}

// Queries specific Actions profile on tariff plan
func (self *APIerSv1) GetTPActions(attrs AttrGetTPActions, reply *utils.TPActions) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if as, err := self.StorDb.GetTPActions(attrs.TPid, attrs.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *as[0]
	}
	return nil
}

type AttrGetTPActionIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// Queries Actions identities on specific tariff plan.
func (self *APIerSv1) GetTPActionIds(attrs AttrGetTPActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPActions,
		utils.TPDistinctIds{"tag"}, nil, &attrs.PaginatorWithSearch); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Actions on Tariff plan
func (self *APIerSv1) RemoveTPActions(attrs AttrGetTPActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPActions,
		attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
