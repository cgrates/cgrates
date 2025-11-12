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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
	"github.com/cgrates/ltcache"
)

var failedPostCache *ltcache.Cache

// InitFailedPostCache initializes the failed posts cache.
func InitFailedPostCache(ttl time.Duration, static bool) {
	failedPostCache = ltcache.NewCache(-1, ttl, static, false, []func(itmID string, value any){writeFailedPosts})
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
	var content []byte
	err := guardian.Guardian.Guard(context.TODO(), func(_ *context.Context) error {
		var readErr error
		if content, readErr = os.ReadFile(filePath); readErr != nil {
			return readErr
		}
		return os.Remove(filePath)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+filePath)
	if err != nil {
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
