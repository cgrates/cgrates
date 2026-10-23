// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
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

// GetCacheStats returns CacheStats filtered by cacheIDs
func (adms *AdminSv1) GetCacheStats(ctx *context.Context, args *utils.AttrCacheIDsDB,
	rply *map[string]*ltcache.CacheStats) (err error) {
	chS, err := adms.dm.GetCacheStats(args)
	if err != nil {
		return
	}
	*rply = chS
	return nil
}

// GetStats returns CacheStats filtered by cacheIDs
func (adms *AdminSv1) GetStats(ctx *context.Context, args *utils.AttrCacheIDsDB,
	rply *engine.CacheStatsWithMetadata) (err error) {
	chSt, err := adms.dm.GetCacheStats(args)
	if err != nil {
		return
	}
	*rply = engine.CacheStatsWithMetadata{
		CacheStatistics: chSt,
		Metadata: map[string]any{
			utils.NodeID: adms.cfg.GeneralCfg().NodeID,
		},
	}
	return nil
}

// GetItemIDs returns the IDs for cacheID with given prefix
func (adms *AdminSv1) GetItemIDs(ctx *context.Context, args *utils.ArgsGetCacheItemIDsDB,
	rply *[]string) (err error) {
	itmIDs, err := adms.dm.GetItemIDs(args)
	if err != nil {
		return
	}
	*rply = itmIDs
	return nil
}

// HasItem verifies the existence of an Item in db
func (adms *AdminSv1) HasItem(ctx *context.Context, args *utils.ArgsGetCacheItemDB,
	rply *bool) (err error) {
	has, err := adms.dm.HasItem(args)
	if err != nil {
		return
	}
	*rply = has
	return nil
}

// GetItem returns an Item from the db
func (adms *AdminSv1) GetItem(ctx *context.Context, args *utils.ArgsGetCacheItemDB,
	rply *any) (err error) {
	itm, err := adms.dm.GetItem(args)
	if err != nil {
		return
	}
	*rply = itm
	return nil
}

// GetItemWithRemote returns an Item from local or remote db
func (adms *AdminSv1) GetItemWithRemote(ctx *context.Context, args *utils.ArgsGetCacheItemDB,
	rply *any) (err error) {
	itm, err := adms.dm.GetItemWithRemote(ctx, args)
	if err != nil {
		return
	}
	*rply = itm
	return nil
}

// GetItemExpiryTime returns the expiryTime for an item
func (adms *AdminSv1) GetItemExpiryTime(ctx *context.Context, args *utils.ArgsGetCacheItemDB,
	rply *time.Time) (err error) {
	expTime, err := adms.dm.GetItemExpiryTime(args)
	if err != nil {
		return
	}
	*rply = expTime
	return nil
}

// HasGroup checks existence of a group in db
func (adms *AdminSv1) HasGroup(ctx *context.Context, args *utils.ArgsGetGroupDB,
	rply *bool) (err error) {
	has, err := adms.dm.HasGroup(args)
	if err != nil {
		return
	}
	*rply = has
	return nil
}

// GetGroupItemIDs returns a list of itemIDs in a db group
func (adms *AdminSv1) GetGroupItemIDs(ctx *context.Context, args *utils.ArgsGetGroupDB,
	rply *[]string) (err error) {
	itmIDs, err := adms.dm.GetGroupItemIDs(args)
	if err != nil {
		return
	}
	*rply = itmIDs
	return nil
}
