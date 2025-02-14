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
	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/utils"
)

type TpeSCfg struct {
	Enabled bool
}

type TpeSCfgJson struct {
	Enabled *bool
}

func (TpeSCfg) SName() string { return TPeSJSON }

func (tp *TpeSCfg) Load(ctx *context.Context, db ConfigDB, _ *CGRConfig) (err error) {
	jsn := new(TpeSCfgJson)
	if err = db.GetSection(ctx, tp.SName(), jsn); err != nil {
		return
	}
	if jsn.Enabled != nil {
		tp.Enabled = *jsn.Enabled
	}
	return
}

func (tp TpeSCfg) AsMapInterface() any {
	return map[string]any{
		utils.EnabledCfg: tp.Enabled,
	}
}

func (tp TpeSCfg) CloneSection() Section {
	return tp.Clone()
}

func (tp TpeSCfg) Clone() (tpCln *TpeSCfg) {
	return &TpeSCfg{
		Enabled: tp.Enabled,
	}
}

func diffTpeSCfgJson(d *TpeSCfgJson, v1, v2 *TpeSCfg) *TpeSCfgJson {
	if d == nil {
		d = new(TpeSCfgJson)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	return d
}
