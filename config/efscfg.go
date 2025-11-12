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

package config

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type EfsJsonCfg struct {
	Enabled              *bool   `json:"enabled"`
	PosterAttempts       *int    `json:"poster_attempts"`
	FailedPostsDir       *string `json:"failed_posts_dir"`
	FailedPostsTTL       *string `json:"failed_posts_ttl"`
	FailedPostsStaticTTL *bool   `json:"failed_posts_static_ttl"`
}

type EFsCfg struct {
	Enabled              bool
	PosterAttempts       int           // number of attempts before considering post request failed
	FailedPostsDir       string        // directory where failed export requests are stored
	FailedPostsTTL       time.Duration // cache ttl for batching failed posts before writing to disk
	FailedPostsStaticTTL bool          // if false, ttl resets on every cache access
}

func (EFsCfg) SName() string { return EFsJSON }

func (c *EFsCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsonEFsCfg := new(EfsJsonCfg)
	if err = jsnCfg.GetSection(ctx, EFsJSON, jsonEFsCfg); err != nil {
		return
	}
	return c.loadFromJSONCfg(jsonEFsCfg)
}

// loadFromJSONCfg loads EFs config from JsonCfg
func (c *EFsCfg) loadFromJSONCfg(jc *EfsJsonCfg) error {
	if jc == nil {
		return nil
	}
	if jc.Enabled != nil {
		c.Enabled = *jc.Enabled
	}
	if jc.PosterAttempts != nil {
		c.PosterAttempts = *jc.PosterAttempts
	}
	if jc.FailedPostsDir != nil {
		c.FailedPostsDir = *jc.FailedPostsDir
	}
	if jc.FailedPostsTTL != nil {
		var err error
		if c.FailedPostsTTL, err = utils.ParseDurationWithNanosecs(*jc.FailedPostsTTL); err != nil {
			return err
		}
	}
	if jc.FailedPostsStaticTTL != nil {
		c.FailedPostsStaticTTL = *jc.FailedPostsStaticTTL
	}
	return nil
}

// AsMapInterface returns the config of EFsCfg as a map[string]any
func (c EFsCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.EnabledCfg:              c.Enabled,
		utils.FailedPostsDirCfg:       c.FailedPostsDir,
		utils.FailedPostsStaticTTLCfg: c.FailedPostsStaticTTL,
		utils.PosterAttemptsCfg:       c.PosterAttempts,
	}
	if c.FailedPostsTTL != 0 {
		mp[utils.FailedPostsTTLCfg] = c.FailedPostsTTL.String()
	}
	return mp
}

func (c EFsCfg) CloneSection() Section { return c.Clone() }

func (c EFsCfg) Clone() *EFsCfg {
	return &EFsCfg{
		Enabled:              c.Enabled,
		PosterAttempts:       c.PosterAttempts,
		FailedPostsDir:       c.FailedPostsDir,
		FailedPostsTTL:       c.FailedPostsTTL,
		FailedPostsStaticTTL: c.FailedPostsStaticTTL,
	}
}

func diffEFsJsonCfg(d *EfsJsonCfg, v1, v2 *EFsCfg) *EfsJsonCfg {
	if d == nil {
		return new(EfsJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.PosterAttempts != v2.PosterAttempts {
		d.PosterAttempts = utils.IntPointer(v2.PosterAttempts)
	}
	if v1.FailedPostsDir != v2.FailedPostsDir {
		d.FailedPostsDir = utils.StringPointer(v2.FailedPostsDir)
	}
	if v1.FailedPostsTTL != v2.FailedPostsTTL {
		d.FailedPostsTTL = utils.StringPointer(v2.FailedPostsTTL.String())
	}
	if v1.FailedPostsStaticTTL != v2.FailedPostsStaticTTL {
		d.FailedPostsStaticTTL = utils.BoolPointer(v2.FailedPostsStaticTTL)
	}
	return d
}
