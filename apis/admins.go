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

package apis

import (
	"os"
	"path"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewAdminSv1(cfg *config.CGRConfig, dm *engine.DataManager, connMgr *engine.ConnManager, fltrS *engine.FilterS) *AdminSv1 {
	return &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: connMgr,
		fltrS:   fltrS,
	}
}

type AdminSv1 struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	ping
}

type ArgsReplyFailedPosts struct {
	TypeProvider         string
	FailedRequestsInDir  *string  // if defined it will be our source of requests to be replayed
	FailedRequestsOutDir *string  // if defined it will become our destination for files failing to be replayed, *none to be discarded
	Modules              []string // list of modules for which replay the requests, nil for all
}

// ReplayFailedPosts will repost failed posts found in the FailedRequestsInDir
func (admS *AdminSv1) ReplayFailedPosts(ctx *context.Context, args *ArgsReplyFailedPosts, reply *string) error {
	failedPostsDir := admS.cfg.LoggerCfg().Opts.FailedPostsDir
	if args.FailedRequestsInDir != nil && *args.FailedRequestsInDir != utils.EmptyString {
		failedPostsDir = *args.FailedRequestsInDir
	}
	failedOutDir := failedPostsDir
	if args.FailedRequestsOutDir != nil && *args.FailedRequestsOutDir != utils.EmptyString {
		failedOutDir = *args.FailedRequestsOutDir
	}
	// check all the files in the FailedPostsInDirectory
	filesInDir, err := os.ReadDir(failedPostsDir)
	if err != nil {
		return err
	}
	if len(filesInDir) == 0 {
		return utils.ErrNotFound
	}
	// check every file and check if any of them match the modules
	for _, file := range filesInDir {
		if len(args.Modules) != 0 {
			var allowedModule bool
			for _, module := range args.Modules {
				if strings.HasPrefix(file.Name(), module) {
					allowedModule = true
					break
				}
			}
			if !allowedModule {
				continue
			}
		}
		filePath := path.Join(failedPostsDir, file.Name())
		var expEv utils.FailoverPoster
		if expEv, err = ees.NewFailoverPosterFromFile(filePath, args.TypeProvider); err != nil {
			return err
		}
		// check if the failed out dir path is the same as the same in dir in order to export again in case of failure
		failoverPath := utils.MetaNone
		if failedOutDir != utils.MetaNone {
			failoverPath = path.Join(failedOutDir, file.Name())
		}

		err = expEv.ReplayFailedPosts(admS.cfg.EFsCfg().PosterAttempts)
		if err != nil && failedOutDir != utils.MetaNone { // Got error from HTTPPoster could be that content was not written, we need to write it ourselves
			if err = utils.WriteToFile(failoverPath, expEv); err != nil {
				return utils.NewErrServerError(err)
			}
		}

	}
	*reply = utils.OK
	return nil
}
