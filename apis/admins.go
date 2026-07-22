// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	"github.com/cgrates/guardian"
)

func NewAdminSv1(cfg *config.CGRConfig, dm *engine.DataManager, connMgr *engine.ConnManager,
	fltrS *engine.FilterS, locker *guardian.Locker) *AdminSv1 {
	return &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		locker:  locker,
		connMgr: connMgr,
		fltrS:   fltrS,
	}
}

type AdminSv1 struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	locker  *guardian.Locker
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
}

// ReplayFailedReplicationsArgs contains args for replaying failed replications.
type ReplayFailedReplicationsArgs struct {
	SourcePath string // path for events to be replayed
	FailedPath string // path for events that failed to replay, *none to discard, defaults to SourcePath if empty
	DBConnID   string
}

// ReplayFailedReplications will repost failed requests found in the SourcePath.
func (s *AdminSv1) ReplayFailedReplications(ctx *context.Context, args ReplayFailedReplicationsArgs, reply *string) error {
	if args.DBConnID == "" {
		args.DBConnID = utils.MetaDefault
	}
	// Set default directories if not provided.
	if args.SourcePath == "" {
		if dbConn, has := s.cfg.DbCfg().DBConns[args.DBConnID]; has {
			args.SourcePath = dbConn.RplFailedDir
		}

	}
	if args.SourcePath == "" {
		return utils.NewErrServerError(
			errors.New("no source directory specified: both SourcePath and replicationFailedDir configuration are empty"),
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

		task, err := engine.NewReplicationTaskFromFile(ctx, path, s.locker)
		if err != nil {
			return fmt.Errorf("failed to init ExportEvents from %s: %v", path, err)
		}

		// Determine the failover path.
		failoverPath := utils.MetaNone
		if args.FailedPath != utils.MetaNone {
			failoverPath = filepath.Join(args.FailedPath, d.Name())
		}

		if err := task.Execute(ctx, s.connMgr); err != nil && failoverPath != utils.MetaNone {
			// Write the events that failed to be replayed to the failover directory
			if err = task.WriteToFile(ctx, failoverPath, s.locker); err != nil {
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
