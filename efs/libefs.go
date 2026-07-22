// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package efs

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var failedPostCache *ltcache.Cache

// InitFailedPostCache initializes the failed posts cache.
func InitFailedPostCache(ttl time.Duration, static bool) {
	failedPostCache = ltcache.NewCache(-1, ttl, static, false, []func(itmID string, value any){writeFailedPosts}, nil)
}

func writeFailedPosts(_ string, value any) {
	expEv, canConvert := value.(*FailedExportersLog)
	if !canConvert {
		return
	}
	filePath := expEv.FilePath()
	expEv.lk.RLock()
	defer expEv.lk.RUnlock()
	if err := WriteToFile(filePath, expEv); err != nil {
		utils.Logger.Warning(fmt.Sprintf("Unable to write failed post to file <%s> because <%s>",
			filePath, err))
		return
	}
}

// FilePath returns the file path it should use for saving the failed events
func (expEv *FailedExportersLog) FilePath() string {
	return path.Join(expEv.FailedPostsDir, expEv.Module+utils.PipeSep+utils.UUIDSha1Prefix()+utils.GOBSuffix)
}

type FailoverPoster interface {
	ReplayFailedPosts(*context.Context, int, string) error
}

// WriteToFile writes the events to file
func WriteToFile(filePath string, expEv FailoverPoster) (err error) {
	fileOut, err := os.Create(filePath)
	if err != nil {
		return err
	}
	encd := gob.NewEncoder(fileOut)
	gob.Register(new(utils.CGREvent))
	err = encd.Encode(expEv)
	fileOut.Close()
	return
}

// NewFailoverPosterFromFile returns ExportEvents from the file
// used only on replay failed post
func NewFailoverPosterFromFile(filePath, provider string, efs *EfS) (FailoverPoster, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err := os.Remove(filePath); err != nil {
		return nil, err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(content))
	var expEv FailedExportersLog
	if err := dec.Decode(&expEv); err != nil {
		return nil, err
	}

	switch provider {
	case utils.EEs:
		opts, err := AsOptsEESConfig(expEv.Opts)
		if err != nil {
			return nil, err
		}
		return &FailedExportersEEs{
			cfg:            efs.cfg,
			module:         expEv.Module,
			failedPostsDir: expEv.FailedPostsDir,
			Path:           expEv.Path,
			Opts:           opts,
			Events:         expEv.Events,
			Format:         expEv.Format,

			connMngr: efs.connMgr,
		}, nil
	case utils.Kafka:
		expEv.cfg = efs.cfg
		expEv.connMngr = efs.connMgr
		return &expEv, nil
	default:
		return nil, errors.New("invalid provider")
	}
}
