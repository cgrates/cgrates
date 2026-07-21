// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// FilterSCfg the filters config section
type FilterSCfg struct {
	Conns map[string][]*DynamicConns
}

// loadFilterSCfg loads the FilterS section of the configuration
func (fSCfg *FilterSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnFilterSCfg := new(FilterSJsonCfg)
	if err = jsnCfg.GetSection(ctx, FilterSJSON, jsnFilterSCfg); err != nil {
		return
	}
	return fSCfg.loadFromJSONCfg(jsnFilterSCfg)
}

func (fSCfg *FilterSCfg) loadFromJSONCfg(jsnCfg *FilterSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		for connType, opts := range tagged {
			fSCfg.Conns[connType] = opts
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (fSCfg FilterSCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.ConnsCfg: stripConns(fSCfg.Conns),
	}
	return mp
}

func (FilterSCfg) SName() string               { return FilterSJSON }
func (fSCfg FilterSCfg) CloneSection() Section { return fSCfg.Clone() }

// Clone returns a deep copy of FilterSCfg
func (fSCfg FilterSCfg) Clone() (cln *FilterSCfg) {
	cln = &FilterSCfg{
		Conns: CloneConnsMap(fSCfg.Conns),
	}
	return
}

// Filters config
type FilterSJsonCfg struct {
	Conns map[string][]*DynamicConns `json:"conns,omitempty"`
}

func diffFilterSJsonCfg(d *FilterSJsonCfg, v1, v2 *FilterSCfg) *FilterSJsonCfg {
	if d == nil {
		d = new(FilterSJsonCfg)
	}
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	return d
}
