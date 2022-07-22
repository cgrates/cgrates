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
along with this program.  If not, see <http://.gnu.org/licenses/>
*/

package efs

import (
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
		if _, has := args.APIOpts[kafkaTopic]; has {
			kafkaTopic = utils.IfaceAsString(args.APIOpts[utils.KafkaTopic])
		}
		if qID := utils.FirstNonEmpty(amqpQueueID, s3BucketID,
			sqsQueueID, kafkaTopic); len(qID) != 0 {
			key = utils.ConcatenatedKey(key, qID)
		}
	case utils.Kafka:
	}
	var failedPost *FailedExportersLogg
	if x, ok := failedPostCache.Get(key); ok {
		if x != nil {
			failedPost = x.(*FailedExportersLogg)
		}
	}
	if failedPost == nil {
		failedPost = &FailedExportersLogg{
			Path:           args.Path,
			Format:         format,
			Opts:           args.APIOpts,
			Module:         args.Module,
			FailedPostsDir: args.FailedDir,
		}
		failedPostCache.Set(key, failedPost, nil)
	}
	failedPost.AddEvent(args.Event)
	return nil
}

type ArgsReplayFailedPosts struct {
	Tenant               string
	TypeProvider         string
	FailedRequestsInDir  *string  // if defined it will be our source of requests to be replayed
	FailedRequestsOutDir *string  // if defined it will become our destination for files failing to be replayed, *none to be discarded
	Modules              []string // list of modules for which replay the requests, nil for all
}
