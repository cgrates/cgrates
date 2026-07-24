// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
