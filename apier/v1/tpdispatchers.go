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

//SetTPDispatcher creates a new DispatcherProfile within a tariff plan
func (self *ApierV1) SetTPDispatcher(attr *utils.TPDispatcherProfile, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.SetTPDispatchers([]*utils.TPDispatcherProfile{attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//GetTPCharger queries specific DispatcherProfile on Tariff plan
func (self *ApierV1) GetTPDispatcher(attr *utils.TPTntID, reply *utils.TPDispatcherProfile) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rls, err := self.StorDb.GetTPDispatchers(attr.TPid, attr.Tenant, attr.ID); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rls[0]
	}
	return nil
}

type AttrGetTPDispatcherIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

//GetTPDispatcherIDs queries dispatcher identities on specific tariff plan.
func (self *ApierV1) GetTPDispatcherIDs(attrs *AttrGetTPDispatcherIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDispatchers, utils.TPDistinctIds{"id"},
		nil, &attrs.Paginator); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = ids
	}
	return nil
}

//RemTPCharger removes specific DispatcherProfile on Tariff plan
func (self *ApierV1) RemTPDispatcher(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBLTPDispatchers, attrs.TPid,
		map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = utils.OK
	}
	return nil
}
