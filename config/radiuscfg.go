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

import "github.com/cgrates/cgrates/utils"

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

func (self *RadiusAgentCfg) loadFromJsonCfg(jsnCfg *RadiusAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_net != nil {
		self.ListenNet = *jsnCfg.Listen_net
	}
	if jsnCfg.Listen_auth != nil {
		self.ListenAuth = *jsnCfg.Listen_auth
	}
	if jsnCfg.Listen_acct != nil {
		self.ListenAcct = *jsnCfg.Listen_acct
	}
	if jsnCfg.Client_secrets != nil {
		if self.ClientSecrets == nil {
			self.ClientSecrets = make(map[string]string)
		}
		for k, v := range *jsnCfg.Client_secrets {
			self.ClientSecrets[k] = v
		}
	}
	if jsnCfg.Client_dictionaries != nil {
		if self.ClientDictionaries == nil {
			self.ClientDictionaries = make(map[string]string)
		}
		for k, v := range *jsnCfg.Client_dictionaries {
			self.ClientDictionaries[k] = v
		}
	}
	if jsnCfg.Sessions_conns != nil {
		self.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, attrConn := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if attrConn == utils.MetaInternal {
				self.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				self.SessionSConns[idx] = attrConn
			}
		}
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(RequestProcessor)
			var haveID bool
			for _, rpSet := range self.RequestProcessors {
				if reqProcJsn.ID != nil && rpSet.ID == *reqProcJsn.ID {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
			if err := rp.loadFromJsonCfg(reqProcJsn, separator); err != nil {
				return nil
			}
			if !haveID {
				self.RequestProcessors = append(self.RequestProcessors, rp)
			}
		}
	}
	return nil
}

func (ra *RadiusAgentCfg) AsMapInterface() map[string]interface{} {

	requestProcessors := make([]map[string]interface{}, len(ra.RequestProcessors))
	for i, item := range ra.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}

	return map[string]interface{}{
		utils.EnabledCfg:            ra.Enabled,
		utils.ListenNetCfg:          ra.ListenNet,
		utils.ListenAuthCfg:         ra.ListenAuth,
		utils.ListenAcctCfg:         ra.ListenAcct,
		utils.ClientSecretsCfg:      ra.ClientSecrets,
		utils.ClientDictionariesCfg: ra.ClientDictionaries,
		utils.SessionSConnsCfg:      ra.SessionSConns,
		utils.RequestProcessorsCfg:  requestProcessors,
	}

}
