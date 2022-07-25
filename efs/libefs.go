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

package efs

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var failedPostCache *ltcache.Cache

func init() {
	failedPostCache = ltcache.NewCache(-1, 5*time.Second, true, writeFailedPosts)
}

// SetFailedPostCacheTTL recreates the failed cache
func SetFailedPostCacheTTL(ttl time.Duration) {
	failedPostCache = ltcache.NewCache(-1, ttl, true, writeFailedPosts)
}

func writeFailedPosts(_ string, value interface{}) {
	expEv, canConvert := value.(*FailedExportersLogg)
	if !canConvert {
		return
	}
	filePath := expEv.FilePath()
	expEv.lk.RLock()
	if err := WriteToFile(filePath, expEv); err != nil {
		utils.Logger.Warning(fmt.Sprintf("Unable to write failed post to file <%s> because <%s>",
			filePath, err))
		expEv.lk.RUnlock()
		return
	}
	expEv.lk.RUnlock()
}

// FilePath returns the file path it should use for saving the failed events
func (expEv *FailedExportersLogg) FilePath() string {
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
func NewFailoverPosterFromFile(filePath, providerType string, efs *EfS) (failPoster FailoverPoster, err error) {
	var fileContent []byte
	err = guardian.Guardian.Guard(context.TODO(), func(_ *context.Context) error {
		if fileContent, err = os.ReadFile(filePath); err != nil {
			return err
		}
		return os.Remove(filePath)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+filePath)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(fileContent))
	// unmarshall it
	expEv := new(FailedExportersLogg)
	err = dec.Decode(&expEv)
	switch providerType {
	case utils.EEs:
		opts, err := AsOptsEESConfig(expEv.Opts)
		if err != nil {
			return nil, err
		}
		failPoster = &FailedExportersEEs{
			module:         expEv.Module,
			failedPostsDir: expEv.FailedPostsDir,
			Path:           expEv.Path,
			Opts:           opts,
			Events:         expEv.Events,
			Format:         expEv.Format,

			connMngr: efs.connMgr,
		}
	case utils.Kafka:
		expEv.cfg = efs.cfg
		expEv.connMngr = efs.connMgr
		failPoster = expEv
	}
	return
}
