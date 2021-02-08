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

// RadiusAgentCfg the config section that describes the Radius Agent
type RadiusAgentCfg struct {
	Enabled            bool
	ListenNet          string // udp or tcp
	ListenAuth         string
	ListenAcct         string
	ClientSecrets      map[string]string
	ClientDictionaries map[string]string
	SessionSConns      []string
	RequestProcessors  []*RequestProcessor
}

func (ra *RadiusAgentCfg) loadFromJSONCfg(jsnCfg *RadiusAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		ra.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_net != nil {
		ra.ListenNet = *jsnCfg.Listen_net
	}
	if jsnCfg.Listen_auth != nil {
		ra.ListenAuth = *jsnCfg.Listen_auth
	}
	if jsnCfg.Listen_acct != nil {
		ra.ListenAcct = *jsnCfg.Listen_acct
	}
	if jsnCfg.Client_secrets != nil {
		if ra.ClientSecrets == nil {
			ra.ClientSecrets = make(map[string]string)
		}
		for k, v := range *jsnCfg.Client_secrets {
			ra.ClientSecrets[k] = v
		}
	}
	if jsnCfg.Client_dictionaries != nil {
		if ra.ClientDictionaries == nil {
			ra.ClientDictionaries = make(map[string]string)
		}
		for k, v := range *jsnCfg.Client_dictionaries {
			ra.ClientDictionaries[k] = v
		}
	}
	if jsnCfg.Sessions_conns != nil {
		ra.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, attrConn := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			ra.SessionSConns[idx] = attrConn
			if attrConn == utils.MetaInternal {
				ra.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			}
		}
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(RequestProcessor)
			var haveID bool
			for _, rpSet := range ra.RequestProcessors {
				if reqProcJsn.ID != nil && rpSet.ID == *reqProcJsn.ID {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
			if err = rp.loadFromJSONCfg(reqProcJsn, separator); err != nil {
				return
			}
			if !haveID {
				ra.RequestProcessors = append(ra.RequestProcessors, rp)
			}
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (ra *RadiusAgentCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:    ra.Enabled,
		utils.ListenNetCfg:  ra.ListenNet,
		utils.ListenAuthCfg: ra.ListenAuth,
		utils.ListenAcctCfg: ra.ListenAcct,
	}

	requestProcessors := make([]map[string]interface{}, len(ra.RequestProcessors))
	for i, item := range ra.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.RequestProcessorsCfg] = requestProcessors

	if ra.SessionSConns != nil {
		sessionSConns := make([]string, len(ra.SessionSConns))
		for i, item := range ra.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
	}
	clientSecrets := make(map[string]string)
	for k, v := range ra.ClientSecrets {
		clientSecrets[k] = v
	}
	initialMP[utils.ClientSecretsCfg] = clientSecrets
	clientDictionaries := make(map[string]string)
	for k, v := range ra.ClientDictionaries {
		clientDictionaries[k] = v
	}
	initialMP[utils.ClientDictionariesCfg] = clientDictionaries
	return
}

// Clone returns a deep copy of RadiusAgentCfg
func (ra RadiusAgentCfg) Clone() (cln *RadiusAgentCfg) {
	cln = &RadiusAgentCfg{
		Enabled:            ra.Enabled,
		ListenNet:          ra.ListenNet,
		ListenAuth:         ra.ListenAuth,
		ListenAcct:         ra.ListenAcct,
		ClientSecrets:      make(map[string]string),
		ClientDictionaries: make(map[string]string),
	}
	if ra.SessionSConns != nil {
		cln.SessionSConns = make([]string, len(ra.SessionSConns))
		for i, con := range ra.SessionSConns {
			cln.SessionSConns[i] = con
		}
	}
	for k, v := range ra.ClientSecrets {
		cln.ClientSecrets[k] = v
	}
	for k, v := range ra.ClientDictionaries {
		cln.ClientDictionaries[k] = v
	}
	if ra.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(ra.RequestProcessors))
		for i, req := range ra.RequestProcessors {
			cln.RequestProcessors[i] = req.Clone()
		}
	}
	return
}
