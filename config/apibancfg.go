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
