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

// SetTPDestination creates a new destination within a tariff plan
func (apierSv1 *APIerSv1) SetTPDestination(attrs *utils.TPDestination, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID, utils.Prefixes}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.SetTPDestinations([]*utils.TPDestination{attrs}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPDestination struct {
	TPid string // Tariff plan id
	ID   string // Destination id
}

// GetTPDestination queries a specific destination
func (apierSv1 *APIerSv1) GetTPDestination(attrs *AttrGetTPDestination, reply *utils.TPDestination) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tpDsts, err := apierSv1.StorDb.GetTPDestinations(attrs.TPid, attrs.ID)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	if len(tpDsts) == 0 {
		return utils.ErrNotFound
	}
	tpDst := tpDsts[0]
	*reply = utils.TPDestination{TPid: tpDst.TPid,
		ID: tpDst.ID, Prefixes: tpDst.Prefixes}
	return nil
}

type AttrGetTPDestinationIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPDestinationIDs queries destination identities on specific tariff plan.
func (apierSv1 *APIerSv1) GetTPDestinationIDs(attrs *AttrGetTPDestinationIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := apierSv1.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDestinations,
		[]string{utils.TagCfg}, nil, &attrs.PaginatorWithSearch)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	if ids == nil {
		return utils.ErrNotFound
	}
	*reply = ids
	return nil
}

// RemoveTPDestination removes specific Destination on Tariff plan
func (apierSv1 *APIerSv1) RemoveTPDestination(attrs *AttrGetTPDestination, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{utils.TPid, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := apierSv1.StorDb.RemTpData(utils.TBLTPDestinations, attrs.TPid, map[string]string{utils.TagCfg: attrs.ID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
