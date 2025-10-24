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

package ees

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var failedPostCache *ltcache.Cache

func init() {
	failedPostCache = ltcache.NewCache(-1, 5*time.Second, false, false, []func(itmID string, value any){writeFailedPosts}) // configurable  general
}

// SetFailedPostCacheTTL recreates the failed cache
func SetFailedPostCacheTTL(ttl time.Duration, static bool) {
	failedPostCache = ltcache.NewCache(-1, ttl, static, false, []func(itmID string, value any){writeFailedPosts})
}

func writeFailedPosts(_ string, value any) {
	expEv, canConvert := value.(*ExportEvents)
	if !canConvert {
		return
	}
	filePath := expEv.FilePath()
	if err := expEv.WriteToFile(filePath); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to write file <%s> because <%s>",
			utils.CDRs, filePath, err))
	}
}

func AddFailedPost(failedPostsDir, expPath, format string, ev any, opts *config.EventExporterOpts) {
	key := utils.ConcatenatedKey(failedPostsDir, expPath, format)
	// also in case of amqp,amqpv1,s3,sqs and kafka also separe them after queue id
	var amqpQueueID string
	var s3BucketID string
	var sqsQueueID string
	var kafkaTopic string

	if amqpOpts := opts.AMQP; amqpOpts != nil {
		if opts.AMQP.QueueID != nil {
			amqpQueueID = *opts.AMQP.QueueID
		}
	}

	if awsOpts := opts.AWS; awsOpts != nil {
		if opts.AWS.S3BucketID != nil {
			s3BucketID = *opts.AWS.S3BucketID
		}
		if opts.AWS.SQSQueueID != nil {
			sqsQueueID = *opts.AWS.SQSQueueID
		}
	}
	if kfkOpts := opts.Kafka; kfkOpts != nil {
		if opts.Kafka.Topic != nil {
			kafkaTopic = *opts.Kafka.Topic
		}
	}
	if qID := utils.FirstNonEmpty(amqpQueueID, s3BucketID, sqsQueueID,
		kafkaTopic); len(qID) != 0 {
		key = utils.ConcatenatedKey(key, qID)
	}
	var failedPost *ExportEvents
	if x, ok := failedPostCache.Get(key); ok {
		if x != nil {
			failedPost = x.(*ExportEvents)
		}
	}
	if failedPost == nil {
		failedPost = &ExportEvents{
			Path:           expPath,
			Type:           format,
			Opts:           opts,
			failedPostsDir: failedPostsDir,
		}
	}
	failedPost.AddEvent(ev)
	failedPostCache.Set(key, failedPost, nil)
}

// NewExportEventsFromFile returns ExportEvents from the file
// used only on replay failed post
func NewExportEventsFromFile(filePath string) (expEv *ExportEvents, err error) {
	var fileContent []byte
	if err = guardian.Guardian.Guard(func() error {
		if fileContent, err = os.ReadFile(filePath); err != nil {
			return err
		}
		return os.Remove(filePath)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+filePath); err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(fileContent))
	// unmarshall it
	expEv = new(ExportEvents)
	err = dec.Decode(&expEv)
	return
}

// ExportEvents used to save the failed post to file
type ExportEvents struct {
	lk             sync.RWMutex
	Path           string
	Opts           *config.EventExporterOpts
	Type           string
	Events         []any
	failedPostsDir string
}

// FilePath returns the file path it should use for saving the failed events
func (expEv *ExportEvents) FilePath() string {
	return path.Join(expEv.failedPostsDir, utils.EEs+utils.PipeSep+utils.UUIDSha1Prefix()+utils.GOBSuffix)
}

// WriteToFile writes the events to file
func (expEv *ExportEvents) WriteToFile(filePath string) (err error) {
	return guardian.Guardian.Guard(func() error {
		fileOut, err := os.Create(filePath)
		if err != nil {
			return err
		}
		encd := gob.NewEncoder(fileOut)
		err = encd.Encode(expEv)
		fileOut.Close()
		return err
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+filePath)
}

// AddEvent adds one event
func (expEv *ExportEvents) AddEvent(ev any) {
	expEv.lk.Lock()
	expEv.Events = append(expEv.Events, ev)
	expEv.lk.Unlock()
}

// ReplayFailedPosts tryies to post cdrs again
func (expEv *ExportEvents) ReplayFailedPosts(attempts int) (failedEvents *ExportEvents, err error) {
	eeCfg := config.NewEventExporterCfg("ReplayFailedPosts", expEv.Type, expEv.Path, utils.MetaNone,
		attempts, expEv.Opts)
	var ee EventExporter
	if ee, err = NewEventExporter(eeCfg, config.CgrConfig(), nil, nil); err != nil {
		return
	}
	keyFunc := func() string { return utils.EmptyString }
	if expEv.Type == utils.MetaKafkajsonMap || expEv.Type == utils.MetaS3jsonMap {
		keyFunc = utils.UUIDSha1Prefix
	}
	failedEvents = &ExportEvents{
		Path: expEv.Path,
		Opts: expEv.Opts,
		Type: expEv.Type,
	}
	for _, ev := range expEv.Events {
		if err = ExportWithAttempts(ee, ev, keyFunc()); err != nil {
			failedEvents.AddEvent(ev)
		}
	}
	ee.Close()

	switch len(failedEvents.Events) {
	case 0: // none failed to be replayed
		return nil, nil
	case len(expEv.Events): // all failed, return last encountered error
		return failedEvents, err
	default:
		return failedEvents, utils.ErrPartiallyExecuted
	}
}
