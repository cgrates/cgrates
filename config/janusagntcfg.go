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
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// JanusConn represents one connection to Janus server
type JanusConn struct {
	Address       string // Address to reach Janus
	Type          string // Connection type
	AdminAddress  string
	AdminPassword string
}

func (jc *JanusConn) loadFromJSONCfg(jsnCfg *JanusConnJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Address != nil {
		jc.Address = *jsnCfg.Address
	}
	if jsnCfg.Type != nil {
		jc.Type = *jsnCfg.Type
	}
	if jsnCfg.AdminAddress != nil {
		jc.AdminAddress = *jsnCfg.AdminAddress
	}
	if jsnCfg.AdminPassword != nil {
		jc.AdminPassword = *jsnCfg.AdminPassword
	}
	return
}

func (jc *JanusConn) AsMapInterface() map[string]any {
	return map[string]any{
		utils.AddressCfg:       jc.Address,
		utils.TypeCfg:          jc.Type,
		utils.AdminAddressCfg:  jc.AdminAddress,
		utils.AdminPasswordCfg: jc.AdminPassword,
	}
}

func (jc *JanusConn) Clone() *JanusConn {
	return &JanusConn{
		Address:       jc.Address,
		Type:          jc.Type,
		AdminAddress:  jc.AdminAddress,
		AdminPassword: jc.AdminPassword,
	}
}

// JanusAgentCfg the config for an Janus Agent
type JanusAgentCfg struct {
	Enabled           bool
	URL               string
	SessionSConns     []string
	JanusConns        []*JanusConn // connections towards Janus
	RequestProcessors []*RequestProcessor
}

func (jaCfg *JanusAgentCfg) loadFromJSONCfg(jsnCfg *JanusAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return
	}

	if jsnCfg.Enabled != nil {
		jaCfg.Enabled = *jsnCfg.Enabled
	}

	if jsnCfg.Url != nil {
		jaCfg.URL = *jsnCfg.Url
	}

	if jsnCfg.Sessions_conns != nil {
		jaCfg.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, connID := range *jsnCfg.Sessions_conns {
			jaCfg.SessionSConns[idx] = connID

			if connID == utils.MetaInternal || connID == rpcclient.BiRPCInternal {
				jaCfg.SessionSConns[idx] = utils.ConcatenatedKey(connID, utils.MetaSessionS)
			}
		}
	}

	if jsnCfg.Janus_conns != nil {
		jaCfg.JanusConns = make([]*JanusConn, len(*jsnCfg.Janus_conns))
		for idx, janConnJsn := range *jsnCfg.Janus_conns {
			jc := new(JanusConn)
			if err = jc.loadFromJSONCfg(janConnJsn); err != nil {
				return
			}
			jaCfg.JanusConns[idx] = jc
		}
	}

	if jsnCfg.RequestProcessors != nil {
		for _, reqProcJsn := range *jsnCfg.RequestProcessors {
			rp := new(RequestProcessor)
			var haveID bool
			for _, rpSet := range jaCfg.RequestProcessors {
				if reqProcJsn.ID != nil && rpSet.ID == *reqProcJsn.ID {
					rp = rpSet
					haveID = true
					break
				}
			}

			if err = rp.loadFromJSONCfg(reqProcJsn, separator); err != nil {
				return
			}
			if !haveID {
				jaCfg.RequestProcessors = append(jaCfg.RequestProcessors, rp)
			}

		}
	}
	return
}

func (jaCfg *JanusAgentCfg) AsMapInterface(separator string) (initialMP map[string]any) {

	initialMP = map[string]any{
		utils.EnabledCfg: jaCfg.Enabled,
		utils.URLCfg:     jaCfg.URL,
	}

	if jaCfg.SessionSConns != nil {
		sessionConns := make([]string, len(jaCfg.SessionSConns))

		for i, item := range jaCfg.SessionSConns {
			sessionConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionConns[i] = utils.MetaInternal
			} else if item == utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS) {
				sessionConns[i] = rpcclient.BiRPCInternal
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionConns
	}

	if jaCfg.JanusConns != nil {
		janConns := make([]map[string]any, len(jaCfg.JanusConns))
		for i, jc := range jaCfg.JanusConns {
			janConns[i] = jc.AsMapInterface()
		}
		initialMP[utils.JanusConnsCfg] = janConns
	}

	requestProcessors := make([]map[string]any, len(jaCfg.RequestProcessors))
	for i, item := range jaCfg.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.RequestProcessorsCfg] = requestProcessors

	return
}

func (jaCfg *JanusAgentCfg) Clone() *JanusAgentCfg {
	cln := &JanusAgentCfg{
		Enabled: jaCfg.Enabled,
		URL:     jaCfg.URL,
	}

	if jaCfg.SessionSConns != nil {
		cln.SessionSConns = make([]string, len(jaCfg.SessionSConns))
		copy(cln.SessionSConns, jaCfg.SessionSConns)
	}

	if jaCfg.JanusConns != nil {
		cln.JanusConns = make([]*JanusConn, len(jaCfg.JanusConns))
		for i, jc := range jaCfg.JanusConns {
			cln.JanusConns[i] = jc.Clone()
		}
	}

	if jaCfg.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(jaCfg.RequestProcessors))
		for i, rp := range jaCfg.RequestProcessors {
			cln.RequestProcessors[i] = rp.Clone()
		}
	}
	return cln
}
