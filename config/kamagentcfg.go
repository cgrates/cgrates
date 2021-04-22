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

// NewDfltKamConnConfig returns the first cached default value for a KamConnCfg connection
func NewDfltKamConnConfig() *KamConnCfg {
	if dfltKamConnConfig == nil {
		return new(KamConnCfg) // No defaults, most probably we are building the defaults now
	}
	return dfltKamConnConfig.Clone()
}

// KamConnCfg represents one connection instance towards Kamailio
type KamConnCfg struct {
	Alias      string
	Address    string
	Reconnects int
}

func (kamCfg *KamConnCfg) loadFromJSONCfg(jsnCfg *KamConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		kamCfg.Address = *jsnCfg.Address
	}
	if jsnCfg.Alias != nil {
		kamCfg.Alias = *jsnCfg.Alias
	}
	if jsnCfg.Reconnects != nil {
		kamCfg.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (kamCfg *KamConnCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.AliasCfg:      kamCfg.Alias,
		utils.AddressCfg:    kamCfg.Address,
		utils.ReconnectsCfg: kamCfg.Reconnects,
	}
}

// Clone returns a deep copy of KamConnCfg
func (kamCfg KamConnCfg) Clone() *KamConnCfg {
	return &KamConnCfg{
		Alias:      kamCfg.Alias,
		Address:    kamCfg.Address,
		Reconnects: kamCfg.Reconnects,
	}
}

// KamAgentCfg is the Kamailio config section
type KamAgentCfg struct {
	Enabled       bool
	SessionSConns []string
	CreateCdr     bool
	EvapiConns    []*KamConnCfg
	Timezone      string
}

func (ka *KamAgentCfg) loadFromJSONCfg(jsnCfg *KamAgentJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		ka.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		ka.SessionSConns = updateBiRPCInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	if jsnCfg.Create_cdr != nil {
		ka.CreateCdr = *jsnCfg.Create_cdr
	}
	if jsnCfg.Evapi_conns != nil {
		ka.EvapiConns = make([]*KamConnCfg, len(*jsnCfg.Evapi_conns))
		for idx, jsnConnCfg := range *jsnCfg.Evapi_conns {
			ka.EvapiConns[idx] = NewDfltKamConnConfig()
			ka.EvapiConns[idx].loadFromJSONCfg(jsnConnCfg)
		}
	}
	if jsnCfg.Timezone != nil {
		ka.Timezone = *jsnCfg.Timezone
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (ka *KamAgentCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:   ka.Enabled,
		utils.CreateCdrCfg: ka.CreateCdr,
		utils.TimezoneCfg:  ka.Timezone,
	}
	if ka.EvapiConns != nil {
		evapiConns := make([]map[string]interface{}, len(ka.EvapiConns))
		for i, item := range ka.EvapiConns {
			evapiConns[i] = item.AsMapInterface()
		}
		initialMP[utils.EvapiConnsCfg] = evapiConns
	}
	if ka.SessionSConns != nil {
		initialMP[utils.SessionSConnsCfg] = getBiRPCInternalJSONConns(ka.SessionSConns)
	}
	return
}

// Clone returns a deep copy of KamAgentCfg
func (ka KamAgentCfg) Clone() (cln *KamAgentCfg) {
	cln = &KamAgentCfg{
		Enabled:   ka.Enabled,
		CreateCdr: ka.CreateCdr,
		Timezone:  ka.Timezone,
	}
	if ka.SessionSConns != nil {
		cln.SessionSConns = utils.CloneStringSlice(ka.SessionSConns)
	}
	if ka.EvapiConns != nil {
		cln.EvapiConns = make([]*KamConnCfg, len(ka.EvapiConns))
		for i, req := range ka.EvapiConns {
			cln.EvapiConns[i] = req.Clone()
		}
	}
	return
}

// Represents one connection instance towards Kamailio
type KamConnJsonCfg struct {
	Alias      *string
	Address    *string
	Reconnects *int
}

func diffKamConnJsonCfg(v1, v2 *KamConnCfg) (d *KamConnJsonCfg) {
	d = new(KamConnJsonCfg)
	if v1.Alias != v2.Alias {
		d.Alias = utils.StringPointer(v2.Alias)
	}
	if v1.Address != v2.Address {
		d.Address = utils.StringPointer(v2.Address)
	}
	if v1.Reconnects != v2.Reconnects {
		d.Reconnects = utils.IntPointer(v2.Reconnects)
	}
	return
}

// KamAgentJsonCfg kamailio config section
type KamAgentJsonCfg struct {
	Enabled        *bool
	Sessions_conns *[]string
	Create_cdr     *bool
	Evapi_conns    *[]*KamConnJsonCfg
	Timezone       *string
}

func equalsKamConnsCfg(v1, v2 []*KamConnCfg) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].Alias != v2[i].Alias ||
			v1[i].Address != v2[i].Address ||
			v1[i].Reconnects != v2[i].Reconnects {
			return false
		}
	}
	return true
}

func diffKamAgentJsonCfg(d *KamAgentJsonCfg, v1, v2 *KamAgentCfg) *KamAgentJsonCfg {
	if d == nil {
		d = new(KamAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}
	if v1.CreateCdr != v2.CreateCdr {
		d.Create_cdr = utils.BoolPointer(v2.CreateCdr)
	}
	if !equalsKamConnsCfg(v1.EvapiConns, v2.EvapiConns) {
		dft := NewDfltKamConnConfig()
		conns := make([]*KamConnJsonCfg, len(v2.EvapiConns))
		for i, conn := range v2.EvapiConns {
			conns[i] = diffKamConnJsonCfg(dft, conn)
		}
		d.Evapi_conns = &conns
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	return d
}
