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
along with this program.  If not, see <http://.gnu.org/licenses/>
*/

package efs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type EfS struct {
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
	eesMux  sync.RWMutex
}

// NewEfs is the constructor for the Efs
func NewEfs(cfg *config.CGRConfig, connMgr *engine.ConnManager) *EfS {
	return &EfS{
		cfg:     cfg,
		connMgr: connMgr,
	}
}

// V1ProcessEvent will write into gob formnat file the Events that were failed to be exported.
func (efs *EfS) V1ProcessEvent(ctx *context.Context, args *utils.ArgsFailedPosts, reply *string) error {
	var format string
	if _, has := args.APIOpts[utils.Format]; has {
		format = utils.IfaceAsString(args.APIOpts[utils.Format])
	}
	key := utils.ConcatenatedKey(args.FailedDir, args.Path, format, args.Module)
	switch args.Module {
	case utils.EEs:
		// also in case of amqp,amqpv1,s3,sqs and kafka also separe them after queue id
		var amqpQueueID string
		var s3BucketID string
		var sqsQueueID string
		var kafkaTopic string
		if _, has := args.APIOpts[utils.AMQPQueueID]; has {
			amqpQueueID = utils.IfaceAsString(args.APIOpts[utils.AMQPQueueID])
		}
		if _, has := args.APIOpts[utils.S3Bucket]; has {
			s3BucketID = utils.IfaceAsString(args.APIOpts[utils.S3Bucket])
		}
		if _, has := args.APIOpts[utils.SQSQueueID]; has {
			sqsQueueID = utils.IfaceAsString(args.APIOpts[utils.SQSQueueID])
		}
		if _, has := args.APIOpts[utils.KafkaTopic]; has {
			kafkaTopic = utils.IfaceAsString(args.APIOpts[utils.KafkaTopic])
		}
		if qID := utils.FirstNonEmpty(amqpQueueID, s3BucketID,
			sqsQueueID, kafkaTopic); len(qID) != 0 {
			key = utils.ConcatenatedKey(key, qID)
		}
	case utils.Kafka:
	}
	var failedPost *FailedExportersLog
	if x, ok := failedPostCache.Get(key); ok {
		if x != nil {
			failedPost = x.(*FailedExportersLog)
		}
	}
	if failedPost == nil {
		failedPost = &FailedExportersLog{
			Path:           args.Path,
			Format:         format,
			Opts:           args.APIOpts,
			Module:         args.Module,
			FailedPostsDir: utils.FirstNonEmpty(args.FailedDir, efs.cfg.EFsCfg().FailedPostsDir),
		}
		failedPostCache.Set(key, failedPost, nil)
	}
	failedPost.AddEvent(args.Event)
	*reply = utils.OK
	return nil
}

// ReplayEventsParams contains parameters for replaying failed posts.
type ReplayEventsParams struct {
	Tenant     string
	Provider   string   // source of failed posts
	SourcePath string   // path for events to be replayed
	FailedPath string   // path for events that failed to replay, *none to discard, defaults to SourceDir if empty
	Modules    []string // list of modules to replay requests for, nil for all
}

// V1ReplayEvents will read the Events from gob files that were failed to be exported and try to re-export them again.
func (efS *EfS) V1ReplayEvents(ctx *context.Context, args ReplayEventsParams, reply *string) error {

	// Set default tenant and directories if not provided.
	if args.Tenant == "" {
		args.Tenant = efS.cfg.GeneralCfg().DefaultTenant
	}
	if args.SourcePath == "" {
		args.SourcePath = efS.cfg.EFsCfg().FailedPostsDir
	}
	if args.FailedPath == "" {
		args.FailedPath = args.SourcePath
	}

	if err := filepath.WalkDir(args.SourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed to access path %s: %v", utils.EFs, path, err))
			return nil // skip paths that cause an error
		}
		if d.IsDir() {
			return nil // skip directories
		}

		// Skip files not belonging to the specified modules.
		if len(args.Modules) != 0 && !slices.ContainsFunc(args.Modules, func(mod string) bool {
			return strings.HasPrefix(d.Name(), mod)
		}) {
			utils.Logger.Info(fmt.Sprintf("<%s> skipping file %s: not found within specified modules", utils.EFs, d.Name()))
			return nil
		}

		expEv, err := NewFailoverPosterFromFile(path, args.Provider, efS)
		if err != nil {
			return fmt.Errorf("failed to init failover poster from %s: %v", path, err)
		}

		// Determine the failover path.
		failoverPath := utils.MetaNone
		if args.FailedPath != utils.MetaNone {
			failoverPath = filepath.Join(args.FailedPath, d.Name())
		}

		err = expEv.ReplayFailedPosts(ctx, efS.cfg.EFsCfg().PosterAttempts, args.Tenant)
		if err != nil && failoverPath != utils.MetaNone {
			// Write the events that failed to be replayed to the failover directory.
			if err = WriteToFile(failoverPath, expEv); err != nil {
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
