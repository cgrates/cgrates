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

// Queries all versions from dataDB
func (self *ApierV1) GetDataDBVersions(arg string, reply *engine.Versions) error {
	if vrs, err := self.DataManager.DataDB().GetVersions(""); err != nil {
		return utils.NewErrServerError(err)
	} else if len(vrs) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = vrs
	}
	return nil
}

// Queries all versions from stordb
func (self *ApierV1) GetStorDBVersions(arg string, reply *engine.Versions) error {
	if vrs, err := self.StorDb.GetVersions(""); err != nil {
		return utils.NewErrServerError(err)
	} else if len(vrs) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = vrs
	}
	return nil
}
