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
func (self *ApierV1) SetTPDestination(attrs utils.V1TPDestination, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationId", "Prefixes"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPDestinations([]*utils.TPDestination{attrs.AsTPDestination()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPDestination struct {
	TPid          string // Tariff plan id
	DestinationId string // Destination id
}

// Queries a specific destination
func (self *ApierV1) GetTPDestination(attrs AttrGetTPDestination, reply *utils.V1TPDestination) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if tpDsts, err := self.StorDb.GetTPDestinations(attrs.TPid, attrs.DestinationId); err != nil {
		return utils.APIErrorHandler(err)
	} else if len(tpDsts) == 0 {
		return utils.ErrNotFound
	} else {
		tpDst := tpDsts[0]
		*reply = utils.V1TPDestination{TPid: tpDst.TPid,
			DestinationId: tpDst.Tag, Prefixes: tpDst.Prefixes}
	}
	return nil
}

type AttrGetTPDestinationIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries destination identities on specific tariff plan.
func (self *ApierV1) GetTPDestinationIDs(attrs AttrGetTPDestinationIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_DESTINATIONS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.APIErrorHandler(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

func (self *ApierV1) RemTPDestination(attrs AttrGetTPDestination, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_DESTINATIONS, attrs.TPid, map[string]string{"tag": attrs.DestinationId}); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = "OK"
	}
	return nil
}
