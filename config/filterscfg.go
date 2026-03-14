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
		if fSCfg.Conns == nil {
			fSCfg.Conns = make(map[string][]*DynamicConns)
		}
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
