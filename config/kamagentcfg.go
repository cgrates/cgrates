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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// Represents one connection instance towards Kamailio
type KamConnCfg struct {
	Alias      string
	Address    string
	Reconnects int
}

func (self *KamConnCfg) loadFromJsonCfg(jsnCfg *KamConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		self.Address = *jsnCfg.Address
	}
	if jsnCfg.Alias != nil {
		self.Alias = *jsnCfg.Alias
	}
	if jsnCfg.Reconnects != nil {
		self.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

func (kamCfg *KamConnCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.AliasCfg:      kamCfg.Alias,
		utils.AddressCfg:    kamCfg.Address,
		utils.ReconnectsCfg: kamCfg.Reconnects,
	}
}

// SM-Kamailio config section
type KamAgentCfg struct {
	Enabled       bool
	SessionSConns []string
	CreateCdr     bool
	EvapiConns    []*KamConnCfg
	Timezone      string
}

func (ka *KamAgentCfg) loadFromJsonCfg(jsnCfg *KamAgentJsonCfg) error {
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
			if attrConn == utils.MetaInternal {
				ka.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				ka.SessionSConns[idx] = attrConn
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
			ka.EvapiConns[idx].loadFromJsonCfg(jsnConnCfg)
		}
	}
	return nil
}

func (ka *KamAgentCfg) AsMapInterface() map[string]interface{} {
	evapiConns := make([]map[string]interface{}, len(ka.EvapiConns))
	for i, item := range ka.EvapiConns {
		evapiConns[i] = item.AsMapInterface()
	}

	sessionSConns := make([]string, len(ka.SessionSConns))
	for i, item := range ka.SessionSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
		if item == buf {
			sessionSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaSessionS, utils.EmptyString)
		} else {
			sessionSConns[i] = item
		}
	}

	return map[string]interface{}{
		utils.EnabledCfg:       ka.Enabled,
		utils.SessionSConnsCfg: sessionSConns,
		utils.CreateCdrCfg:     ka.CreateCdr,
		utils.EvapiConnsCfg:    evapiConns,
		utils.TimezoneCfg:      ka.Timezone,
	}

}
