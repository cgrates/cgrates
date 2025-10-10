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
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewAdminSv1(cfg *config.CGRConfig, dm *engine.DataManager, connMgr *engine.ConnManager, fltrS *engine.FilterS,
	storDB engine.StorDB) *AdminSv1 {
	return &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		storDB:  storDB,
		connMgr: connMgr,
		fltrS:   fltrS,
	}
}

type AdminSv1 struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	storDB  engine.StorDB
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	ping
}

// ReplayFailedReplicationsArgs contains args for replaying failed replications.
type ReplayFailedReplicationsArgs struct {
	SourcePath string // path for events to be replayed
	FailedPath string // path for events that failed to replay, *none to discard, defaults to SourcePath if empty
}

// ReplayFailedReplications will repost failed requests found in the SourcePath.
func (a *AdminSv1) ReplayFailedReplications(ctx *context.Context, args ReplayFailedReplicationsArgs, reply *string) error {

	// Set default directories if not provided.
	if args.SourcePath == "" {
		args.SourcePath = a.cfg.DataDbCfg().RplFailedDir
	}
	if args.SourcePath == "" {
		return utils.NewErrServerError(
			errors.New("no source directory specified: both SourcePath and replication_failed_dir configuration are empty"),
		)
	}
	if args.FailedPath == "" {
		args.FailedPath = args.SourcePath
	}

	if err := filepath.WalkDir(args.SourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("<ReplayFailedReplications> failed to access path %s: %v", path, err))
			return nil // skip paths that cause an error
		}
		if d.IsDir() {
			return nil // skip directories
		}

		task, err := engine.NewReplicationTaskFromFile(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to init ExportEvents from %s: %v", path, err)
		}
		// Determine the failover path.
		failoverPath := utils.MetaNone
		if args.FailedPath != utils.MetaNone {
			failoverPath = filepath.Join(args.FailedPath, d.Name())
		}

		if err := task.Execute(ctx, a.connMgr); err != nil && failoverPath != utils.MetaNone {
			// Write the events that failed to be replayed to the failover directory
			if err = task.WriteToFile(ctx, failoverPath); err != nil {
				return fmt.Errorf("failed to write the events that failed to be replayed to %s: %v", path, err)
			}
		}
		return nil
	}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
