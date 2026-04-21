/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// DumpDB will dump all of offline internal DB from memory to a file
func (adms *AdminSv1) DumpDB(ctx *context.Context, ignr *string, reply *string) (err error) {
	if err = adms.dm.DumpDB(); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RewriteDB will rewrite every dump file of offline internal DB
func (adms *AdminSv1) RewriteDB(ctx *context.Context, ignr *string, reply *string) (err error) {
	if err = adms.dm.RewriteDB(); err != nil {
		return
	}
	*reply = utils.OK
	return
}

type BackupParams struct {
	BackupFolderPath string // The path to the folder where the backup will be created
	Zip              bool   // creates a zip compressing the backup
}

// BackupDB will momentarely stop any dumping and rewriting in offline internal DB, until dump folder is backed up in folder path backupFolderPath. Making zip true will create a zip file in the path instead
func (adms *AdminSv1) BackupDB(ctx *context.Context, params BackupParams, reply *string) (err error) {
	if err = adms.dm.BackupDB(params.BackupFolderPath, params.Zip); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RestoreDB is used only for offline internal DB. It attempts to restore the internal DB from
// the latest backup in the specified backupPath. If backupPath is not specified, it will be
// taken from the default's backup path.
// Any data that was dumped from internal DB will be cleared before restoring from backup
func (adms *AdminSv1) RestoreDB(ctx *context.Context, backupFolderPath string, reply *string) (err error) {
	if err = adms.dm.RestoreDB(backupFolderPath); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// RestoreDB will take the BackupFolderPath (or default backup path if empty) to backup the
// live dump folder taking zip as parameter to zip the backup or not, after which it cleares
// the live dump folder and creates new dump files out of the live internal DB data. Only
// intended for offline internal DB
func (adms *AdminSv1) SnapshotDB(ctx *context.Context, params BackupParams, reply *string) (err error) {
	if err = adms.dm.SnapshotDB(params.BackupFolderPath, params.Zip); err != nil {
		return
	}
	*reply = utils.OK
	return
}
