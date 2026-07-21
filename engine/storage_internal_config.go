// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"encoding/json"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func (iDB *InternalDB) GetSection(_ *context.Context, section string, val any) error {
	result, ok := iDB.db.Get(utils.CacheConfig, section)
	if !ok || result == nil {
		return nil
	}
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, val)
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
