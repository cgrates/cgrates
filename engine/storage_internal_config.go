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

package engine

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func (iDB *InternalDB) GetSection(_ *context.Context, section string, val any) error {
	val, _ = iDB.db.Get(utils.CacheConfig, section)
	return nil
}
func (iDB *InternalDB) SetSection(_ *context.Context, section string, val any) error {
	iDB.db.Set(utils.CacheConfig, section, val, nil,
		true, utils.NonTransactional)
	return nil
}

// Will dump everything inside Configdb to files
func (iDB *InternalDB) DumpConfigDB() (err error) {
	return iDB.db.DumpAll()
}

// Will rewrite every dump file of ConfigDB
func (iDB *InternalDB) RewriteConfigDB() (err error) {
	return iDB.db.RewriteAll()
}

// BackupConfigDB will momentarely stop any dumping and rewriting until all dump folder is backed up in folder path backupFolderPath, making zip true will create a zip file in the path instead
func (iDB *InternalDB) BackupConfigDB(backupFolderPath string, zip bool) (err error) {
	return iDB.db.BackupDumpFolder(backupFolderPath, zip)
}
