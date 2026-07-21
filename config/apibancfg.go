// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// APIBanCfg the config for the APIBan Keys
type APIBanCfg struct {
	Enabled bool
	Keys    []string
}

// loadAPIBanCgrCfg loads the Analyzer section of the configuration
func (ban *APIBanCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnAPIBanCfg := new(APIBanJsonCfg)
	if err = jsnCfg.GetSection(ctx, APIBanJSON, jsnAPIBanCfg); err != nil {
		return
	}
	return ban.loadFromJSONCfg(jsnAPIBanCfg)
}

func (ban *APIBanCfg) loadFromJSONCfg(jsnCfg *APIBanJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		ban.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Keys != nil {
		ban.Keys = slices.Clone(*jsnCfg.Keys)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (ban APIBanCfg) AsMapInterface() any {
	return map[string]any{
		utils.EnabledCfg: ban.Enabled,
		utils.KeysCfg:    ban.Keys,
	}
}

func (APIBanCfg) SName() string             { return APIBanJSON }
func (ban APIBanCfg) CloneSection() Section { return ban.Clone() }

// Clone returns a deep copy of APIBanCfg
func (ban APIBanCfg) Clone() (cln *APIBanCfg) {
	return &APIBanCfg{
		Enabled: ban.Enabled,
		Keys:    slices.Clone(ban.Keys),
	}
}

type APIBanJsonCfg struct {
	Enabled *bool
	Keys    *[]string
}

func diffAPIBanJsonCfg(d *APIBanJsonCfg, v1, v2 *APIBanCfg) *APIBanJsonCfg {
	if d == nil {
		d = new(APIBanJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !slices.Equal(v1.Keys, v2.Keys) {
		d.Keys = &v2.Keys
	}
	return d
}
