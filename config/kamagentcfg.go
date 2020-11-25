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
	dfltVal := *dfltKamConnConfig
	return &dfltVal
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
		ka.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, attrConn := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			ka.SessionSConns[idx] = attrConn
			if attrConn == utils.MetaInternal {
				ka.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			}
		}
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
		sessionSConns := make([]string, len(ka.SessionSConns))
		for i, item := range ka.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
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
		cln.SessionSConns = make([]string, len(ka.SessionSConns))
		for i, con := range ka.SessionSConns {
			cln.SessionSConns[i] = con
		}
	}
	if ka.EvapiConns != nil {
		cln.EvapiConns = make([]*KamConnCfg, len(ka.EvapiConns))
		for i, req := range ka.EvapiConns {
			cln.EvapiConns[i] = req.Clone()
		}
	}
	return
}
