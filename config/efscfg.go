// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type EfsJsonCfg struct {
	Enabled              *bool   `json:"enabled"`
	PosterAttempts       *int    `json:"posterAttempts"`
	FailedPostsDir       *string `json:"failedPostsDir"`
	FailedPostsTTL       *string `json:"failedPostsTTL"`
	FailedPostsStaticTTL *bool   `json:"failedPostsStaticTTL"`
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
