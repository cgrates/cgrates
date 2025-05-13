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
package apis

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// DumpStorDB will dump all of stordb from memory to a file
func (adms *AdminSv1) DumpStorDB(ctx *context.Context, ignr *string, reply *string) (err error) {
	if err = adms.storDB.DumpStorDB(); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// Will rewrite every dump file of StorDB
func (adms *AdminSv1) RewriteStorDB(ctx *context.Context, ignr *string, reply *string) (err error) {
	if err = adms.storDB.RewriteStorDB(); err != nil {
		return
	}
	*reply = utils.OK
	return
}

// BackupStorDB will momentarely stop any dumping and rewriting in storDB, until dump folder is backed up in folder path backupFolderPath. Making zip true will create a zip file in the path instead
func (adms *AdminSv1) BackupStorDB(ctx *context.Context, params DumpBackupParams, reply *string) (err error) {
	if err = adms.storDB.BackupStorDB(params.BackupFolderPath, params.Zip); err != nil {
		return
	}
	*reply = utils.OK
	return
}
