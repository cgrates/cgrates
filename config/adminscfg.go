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
	"github.com/cgrates/cgrates/utils"
)

// AdminSCfg is the configuration of Apier service
type AdminSCfg struct {
	Enabled         bool
	CachesConns     []string // connections towards Cache
	ActionSConns    []string // connections towards Scheduler
	AttributeSConns []string // connections towards AttributeS
	EEsConns        []string // connections towards EEs
}

func (aCfg *AdminSCfg) loadFromJSONCfg(jsnCfg *AdminSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		aCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Caches_conns != nil {
		aCfg.CachesConns = updateInternalConns(*jsnCfg.Caches_conns, utils.MetaCaches)
	}
	if jsnCfg.Actions_conns != nil {
		aCfg.ActionSConns = updateInternalConns(*jsnCfg.Actions_conns, utils.MetaActions)
	}
	if jsnCfg.Attributes_conns != nil {
		aCfg.AttributeSConns = updateInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
	}
	if jsnCfg.Ees_conns != nil {
		aCfg.EEsConns = updateInternalConns(*jsnCfg.Ees_conns, utils.MetaEEs)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (aCfg *AdminSCfg) AsMapInterface() (initialMap map[string]interface{}) {
	initialMap = map[string]interface{}{
		utils.EnabledCfg: aCfg.Enabled,
	}
	if aCfg.CachesConns != nil {
		initialMap[utils.CachesConnsCfg] = getInternalJSONConns(aCfg.CachesConns)
	}
	if aCfg.ActionSConns != nil {
		initialMap[utils.ActionSConnsCfg] = getInternalJSONConns(aCfg.ActionSConns)
	}
	if aCfg.AttributeSConns != nil {
		initialMap[utils.AttributeSConnsCfg] = getInternalJSONConns(aCfg.AttributeSConns)
	}
	if aCfg.EEsConns != nil {
		initialMap[utils.EEsConnsCfg] = getInternalJSONConns(aCfg.EEsConns)
	}
	return
}

// Clone returns a deep copy of ApierCfg
func (aCfg AdminSCfg) Clone() (cln *AdminSCfg) {
	cln = &AdminSCfg{
		Enabled: aCfg.Enabled,
	}
	if aCfg.CachesConns != nil {
		cln.CachesConns = utils.CloneStringSlice(aCfg.CachesConns)
	}
	if aCfg.ActionSConns != nil {
		cln.ActionSConns = utils.CloneStringSlice(aCfg.ActionSConns)
	}
	if aCfg.AttributeSConns != nil {
		cln.AttributeSConns = utils.CloneStringSlice(aCfg.AttributeSConns)
	}
	if aCfg.EEsConns != nil {
		cln.EEsConns = utils.CloneStringSlice(aCfg.EEsConns)
	}
	return
}

type AdminSJsonCfg struct {
	Enabled          *bool
	Caches_conns     *[]string
	Actions_conns    *[]string
	Attributes_conns *[]string
	Ees_conns        *[]string
}

func diffAdminSJsonCfg(d *AdminSJsonCfg, v1, v2 *AdminSCfg) *AdminSJsonCfg {
	if d == nil {
		d = new(AdminSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !utils.SliceStringEqual(v1.CachesConns, v2.CachesConns) {
		d.Caches_conns = utils.SliceStringPointer(getInternalJSONConns(v2.CachesConns))
	}
	if !utils.SliceStringEqual(v1.ActionSConns, v2.ActionSConns) {
		d.Actions_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ActionSConns))
	}
	if !utils.SliceStringEqual(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	if !utils.SliceStringEqual(v1.EEsConns, v2.EEsConns) {
		d.Ees_conns = utils.SliceStringPointer(getInternalJSONConns(v2.EEsConns))
	}
	return d
}
