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

// Creates a new ChargerProfile within a tariff plan
func (self *ApierV1) SetTPCharger(attr utils.TPChargerProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPChargers([]*utils.TPChargerProfile{&attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type AttrGetTPCharger struct {
	TPid string // Tariff plan id
	ID   string
}

// Queries specific ChargerProfile on Tariff plan
func (self *ApierV1) GetTPCharger(attr AttrGetTPCharger, reply *utils.TPChargerProfile) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rls, err := self.StorDb.GetTPChargers(attr.TPid, attr.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rls[0]
	}
	return nil
}

type AttrGetTPChargerIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries Resource identities on specific tariff plan.
func (self *ApierV1) GetTPChargerIDs(attrs AttrGetTPChargerIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPChargers, utils.TPDistinctIds{"id"}, nil, &attrs.Paginator); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Resource on Tariff plan
func (self *ApierV1) RemTPCharger(attrs AttrGetTPCharger, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPChargers, attrs.TPid, map[string]string{"id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
