// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Queries all versions from dataDB
func (apierSv1 *APIerSv1) GetDataDBVersions(ctx *context.Context, ign *string, reply *engine.Versions) error {
	if vrs, err := apierSv1.DataManager.DataDB().GetVersions(""); err != nil {
		return utils.NewErrServerError(err)
	} else if len(vrs) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = vrs
	}
	return nil
}

// Queries all versions from stordb
func (apierSv1 *APIerSv1) GetStorDBVersions(ctx *context.Context, ign *string, reply *engine.Versions) error {
	if vrs, err := apierSv1.StorDb.GetVersions(""); err != nil {
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
func (apierSv1 *APIerSv1) SetDataDBVersions(ctx *context.Context, arg *SetVersionsArg, reply *string) error {
	if arg.Versions == nil {
		arg.Versions = engine.CurrentDataDBVersions()
	}
	if err := apierSv1.DataManager.DataDB().SetVersions(arg.Versions, arg.Overwrite); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Queries all versions from stordb
func (apierSv1 *APIerSv1) SetStorDBVersions(ctx *context.Context, arg *SetVersionsArg, reply *string) error {
	if arg.Versions == nil {
		arg.Versions = engine.CurrentDataDBVersions()
	}
	if err := apierSv1.StorDb.SetVersions(arg.Versions, arg.Overwrite); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
