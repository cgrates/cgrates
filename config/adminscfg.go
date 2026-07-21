// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// AdminSCfg is the configuration of Apier service
type AdminSCfg struct {
	Enabled bool
	Conns   map[string][]*DynamicConns
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
		Conns:   CloneConnsMap(aCfg.Conns),
	}
	return
}

type AdminSJsonCfg struct {
	Enabled *bool
	Conns   map[string][]*DynamicConns `json:"conns,omitempty"`
}

func diffAdminSJsonCfg(d *AdminSJsonCfg, v1, v2 *AdminSCfg) *AdminSJsonCfg {
	if d == nil {
		d = new(AdminSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	return d
}
