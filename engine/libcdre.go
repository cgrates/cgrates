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
	failedPostCache = ltcache.NewCache(-1, 5*time.Second, false, writeFailedPosts) // configurable  general
}

// SetFailedPostCacheTTL recreates the failed cache
func SetFailedPostCacheTTL(ttl time.Duration) {
	failedPostCache = ltcache.NewCache(-1, ttl, false, writeFailedPosts)
}

func writeFailedPosts(itmID string, value interface{}) {
	expEv, canConvert := value.(*ExportEvents)
	if !canConvert {
		return
	}
	filePath := path.Join(config.CgrConfig().GeneralCfg().FailedPostsDir, expEv.FileName())
	if err := expEv.WriteToFile(filePath); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to write file <%s> because <%s>",
			utils.CDRs, filePath, err))
	}
}

func AddFailedPost(expPath, format, module string, ev interface{}, opts map[string]interface{}) {
	key := utils.ConcatenatedKey(expPath, format, module)
	// also in case of amqp,amqpv1,s3,sqs and kafka also separe them after queue id
	if qID := utils.FirstNonEmpty(utils.IfaceAsString(opts[utils.QueueID]),
		utils.IfaceAsString(opts[utils.KafkaTopic])); len(qID) != 0 {
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
			Path:   expPath,
			Format: format,
			Opts:   opts,
			module: module,
		}
	}
	failedPost.AddEvent(ev)
	failedPostCache.Set(key, failedPost, nil)
}

// NewExportEventsFromFile returns ExportEvents from the file
// used only on replay failed post
func NewExportEventsFromFile(filePath string) (expEv *ExportEvents, err error) {
	var fileContent []byte
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		if fileContent, err = os.ReadFile(filePath); err != nil {
			return 0, err
		}
		return 0, os.Remove(filePath)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+filePath)
	if err != nil {
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
	lk     sync.RWMutex
	Path   string
	Opts   map[string]interface{}
	Format string
	Events []interface{}
	module string
}

// FileName returns the file name it should use for saving the failed events
func (expEv *ExportEvents) FileName() string {
	return expEv.module + utils.PipeSep + utils.UUIDSha1Prefix() + utils.GOBSuffix
}

// SetModule sets the module for this event
func (expEv *ExportEvents) SetModule(mod string) {
	expEv.module = mod
}

// WriteToFile writes the events to file
func (expEv *ExportEvents) WriteToFile(filePath string) (err error) {
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		fileOut, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		encd := gob.NewEncoder(fileOut)
		err = encd.Encode(expEv)
		fileOut.Close()
		return nil, err
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+filePath)
	return
}

// AddEvent adds one event
func (expEv *ExportEvents) AddEvent(ev interface{}) {
	expEv.lk.Lock()
	expEv.Events = append(expEv.Events, ev)
	expEv.lk.Unlock()
}

// ReplayFailedPosts tryies to post cdrs again
func (expEv *ExportEvents) ReplayFailedPosts(attempts int) (failedEvents *ExportEvents, err error) {
	failedEvents = &ExportEvents{
		Path:   expEv.Path,
		Opts:   expEv.Opts,
		Format: expEv.Format,
	}
	var pstr Poster
	keyFunc := func() string { return utils.EmptyString }
	switch expEv.Format {
	case utils.MetaHTTPjsonCDR, utils.MetaHTTPjsonMap, utils.MetaHTTPjson, utils.MetaHTTPPost:
		pstr := NewHTTPPoster(config.CgrConfig().GeneralCfg().ReplyTimeout, expEv.Path,
			utils.PosterTransportContentTypes[expEv.Format],
			config.CgrConfig().GeneralCfg().PosterAttempts)

		for _, ev := range expEv.Events {
			req := ev.(*HTTPPosterRequest)
			err = pstr.PostValues(req.Body, req.Header)
			if err != nil {
				failedEvents.AddEvent(req)
			}
		}
		if len(failedEvents.Events) > 0 {
			err = utils.ErrPartiallyExecuted
		} else {
			failedEvents = nil
		}
		return
	case utils.MetaAMQPjsonCDR, utils.MetaAMQPjsonMap:
		pstr = NewAMQPPoster(expEv.Path, attempts, expEv.Opts)
	case utils.MetaAMQPV1jsonMap:
		pstr = NewAMQPv1Poster(expEv.Path, attempts, expEv.Opts)
	case utils.MetaSQSjsonMap:
		pstr = NewSQSPoster(expEv.Path, attempts, expEv.Opts)
	case utils.MetaKafkajsonMap:
		pstr = NewKafkaPoster(expEv.Path, attempts, expEv.Opts)
		keyFunc = utils.UUIDSha1Prefix
	case utils.MetaS3jsonMap:
		pstr = NewS3Poster(expEv.Path, attempts, expEv.Opts)
		keyFunc = utils.UUIDSha1Prefix
	}
	for _, ev := range expEv.Events {
		if err = pstr.Post(ev.([]byte), keyFunc()); err != nil {
			failedEvents.AddEvent(ev)
		}
	}
	pstr.Close()
	if len(failedEvents.Events) > 0 {
		err = utils.ErrPartiallyExecuted
	} else {
		failedEvents = nil
	}
	return
}
