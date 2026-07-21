// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

// SnapshotDB will take the BackupFolderPath (or default backup path if empty) to backup the
// live dump folder taking zip as parameter to zip the backup or not; after which it cleares
// the live dump folder and creates new dump files out of the live internal DB data. Only
// intended for offline internal DB
func (adms *AdminSv1) SnapshotDB(ctx *context.Context, params BackupParams, reply *string) (err error) {
	if err = adms.dm.SnapshotDB(params.BackupFolderPath, params.Zip); err != nil {
		return
	}
	*reply = utils.OK
	return
}
