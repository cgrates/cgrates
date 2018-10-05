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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

//SetFilter add a new Filter
func (self *ApierV1) SetFilter(attrs *engine.Filter, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.DataManager.SetFilter(attrs); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//GetFilter returns a Filter
func (self *ApierV1) GetFilter(arg utils.TenantID, reply *engine.Filter) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if fltr, err := self.DataManager.GetFilter(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *fltr
	}
	return nil
}

// GetFilterIDs returns list of Filter IDs registered for a tenant
func (apierV1 *ApierV1) GetFilterIDs(tenant string, fltrIDs *[]string) error {
	prfx := utils.FilterPrefix + tenant + ":"
	keys, err := apierV1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*fltrIDs = retIDs
	return nil
}

//RemoveFilter  remove a specific filter
func (self *ApierV1) RemoveFilter(arg utils.TenantID, reply *string) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.DataManager.RemoveFilter(arg.Tenant, arg.ID, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = utils.OK
	return nil
}
