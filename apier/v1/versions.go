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
func (self *APIerSv1) GetDataDBVersions(ign string, reply *engine.Versions) error {
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
func (self *APIerSv1) GetStorDBVersions(ign string, reply *engine.Versions) error {
	if vrs, err := self.StorDb.GetVersions(""); err != nil {
		return utils.NewErrServerError(err)
	} else if len(vrs) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = vrs
	}
	return nil
}

type SetVersionsArg struct {
	Versions  engine.Versions
	Overwrite bool
}

// Queries all versions from dataDB
func (self *APIerSv1) SetDataDBVersions(arg SetVersionsArg, reply *string) error {
	if arg.Versions == nil {
		arg.Versions = engine.CurrentDataDBVersions()
	}
	if err := self.DataManager.DataDB().SetVersions(arg.Versions, arg.Overwrite); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Queries all versions from stordb
func (self *APIerSv1) SetStorDBVersions(arg SetVersionsArg, reply *string) error {
	if arg.Versions == nil {
		arg.Versions = engine.CurrentDataDBVersions()
	}
	if err := self.StorDb.SetVersions(arg.Versions, arg.Overwrite); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
