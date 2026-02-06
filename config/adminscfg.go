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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// AdminSCfg is the configuration of Apier service
type AdminSCfg struct {
	Enabled bool
	Conns   map[string][]*DynamicStringSliceOpt
}

// loadApierCfg loads the Apier section of the configuration
func (aCfg *AdminSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnApierCfg := new(AdminSJsonCfg)
	if err = jsnCfg.GetSection(ctx, AdminSJSON, jsnApierCfg); err != nil {
		return
	}
	return aCfg.loadFromJSONCfg(jsnApierCfg)
}

func (aCfg *AdminSCfg) loadFromJSONCfg(jsnCfg *AdminSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		aCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		for connType, opts := range tagged {
			aCfg.Conns[connType] = opts
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (aCfg AdminSCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.EnabledCfg: aCfg.Enabled,
		utils.ConnsCfg:   stripConns(aCfg.Conns),
	}
	return mp
}

func (AdminSCfg) SName() string              { return AdminSJSON }
func (aCfg AdminSCfg) CloneSection() Section { return aCfg.Clone() }

// Clone returns a deep copy of ApierCfg
func (aCfg AdminSCfg) Clone() (cln *AdminSCfg) {
	cln = &AdminSCfg{
		Enabled: aCfg.Enabled,
		Conns:   CloneConnsOpt(aCfg.Conns),
	}
	return
}

type AdminSJsonCfg struct {
	Enabled *bool
	Conns   map[string][]*DynamicStringSliceOpt `json:"conns,omitempty"`
}

func diffAdminSJsonCfg(d *AdminSJsonCfg, v1, v2 *AdminSCfg) *AdminSJsonCfg {
	if d == nil {
		d = new(AdminSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !ConnsEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	return d
}
