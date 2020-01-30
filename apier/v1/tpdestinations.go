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

// Creates a new destination within a tariff plan
func (self *APIerSv1) SetTPDestination(attrs utils.TPDestination, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID", "Prefixes"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPDestinations([]*utils.TPDestination{&attrs}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPDestination struct {
	TPid string // Tariff plan id
	ID   string // Destination id
}

// Queries a specific destination
func (self *APIerSv1) GetTPDestination(attrs AttrGetTPDestination, reply *utils.TPDestination) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if tpDsts, err := self.StorDb.GetTPDestinations(attrs.TPid, attrs.ID); err != nil {
		return utils.APIErrorHandler(err)
	} else if len(tpDsts) == 0 {
		return utils.ErrNotFound
	} else {
		tpDst := tpDsts[0]
		*reply = utils.TPDestination{TPid: tpDst.TPid,
			ID: tpDst.ID, Prefixes: tpDst.Prefixes}
	}
	return nil
}

type AttrGetTPDestinationIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// Queries destination identities on specific tariff plan.
func (self *APIerSv1) GetTPDestinationIDs(attrs AttrGetTPDestinationIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDestinations,
		utils.TPDistinctIds{"tag"}, nil, &attrs.PaginatorWithSearch); err != nil {
		return utils.APIErrorHandler(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Destination on Tariff plan
func (self *APIerSv1) RemoveTPDestination(attrs AttrGetTPDestination, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPDestinations, attrs.TPid, map[string]string{"tag": attrs.ID}); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
