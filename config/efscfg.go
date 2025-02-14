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

package config

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type EFsCfg struct {
	Enabled        bool
	PosterAttempts int           // Time to wait before writing the failed posts in a single file
	FailedPostsDir string        // Directory path where we store failed http requests
	FailedPostsTTL time.Duration // Directory path where we store failed http requests
}

func (EFsCfg) SName() string { return EFsJSON }

func (efsCfg *EFsCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsonEFsCfg := new(EfsJsonCfg)
	if err = jsnCfg.GetSection(ctx, EFsJSON, jsonEFsCfg); err != nil {
		return
	}
	return efsCfg.loadFromJSONCfg(jsonEFsCfg)
}

// loadFromJSONCfg loads EFs config from JsonCfg
func (efsCfg *EFsCfg) loadFromJSONCfg(jsonEFsCfg *EfsJsonCfg) (err error) {
	if jsonEFsCfg == nil {
		return
	}
	if jsonEFsCfg.Enabled != nil {
		efsCfg.Enabled = *jsonEFsCfg.Enabled
	}
	if jsonEFsCfg.Poster_attempts != nil {
		efsCfg.PosterAttempts = *jsonEFsCfg.Poster_attempts
	}
	if jsonEFsCfg.Failed_posts_dir != nil {
		efsCfg.FailedPostsDir = *jsonEFsCfg.Failed_posts_dir
	}
	if jsonEFsCfg.Failed_posts_ttl != nil {
		if efsCfg.FailedPostsTTL, err = utils.ParseDurationWithNanosecs(*jsonEFsCfg.Failed_posts_ttl); err != nil {
			return
		}
	}
	return
}

// AsMapInterface returns the config of EFsCfg as a map[string]any
func (efsCfg EFsCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.EnabledCfg:        efsCfg.Enabled,
		utils.FailedPostsDirCfg: efsCfg.FailedPostsDir,
		utils.PosterAttemptsCfg: efsCfg.PosterAttempts,
	}
	if efsCfg.FailedPostsTTL != 0 {
		mp[utils.FailedPostsTTLCfg] = efsCfg.FailedPostsTTL.String()
	}
	return mp
}

func (efsCfg EFsCfg) CloneSection() Section { return efsCfg.Clone() }

func (efsCfg EFsCfg) Clone() *EFsCfg {
	return &EFsCfg{
		Enabled:        efsCfg.Enabled,
		PosterAttempts: efsCfg.PosterAttempts,
		FailedPostsDir: efsCfg.FailedPostsDir,
		FailedPostsTTL: efsCfg.FailedPostsTTL,
	}
}

type EfsJsonCfg struct {
	Enabled          *bool
	Poster_attempts  *int
	Failed_posts_dir *string
	Failed_posts_ttl *string
}

func diffEFsJsonCfg(d *EfsJsonCfg, v1, v2 *EFsCfg) *EfsJsonCfg {
	if d == nil {
		return new(EfsJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.PosterAttempts != v2.PosterAttempts {
		d.Poster_attempts = utils.IntPointer(v2.PosterAttempts)
	}
	if v1.FailedPostsDir != v2.FailedPostsDir {
		d.Failed_posts_dir = utils.StringPointer(v2.FailedPostsDir)
	}
	if v1.FailedPostsTTL != v2.FailedPostsTTL {
		d.Failed_posts_ttl = utils.StringPointer(v2.FailedPostsTTL.String())
	}
	return d
}
