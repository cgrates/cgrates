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

type RadiusAgentCfg struct {
	Enabled            bool
	ListenNet          string // udp or tcp
	ListenAuth         string
	ListenAcct         string
	ClientSecrets      map[string]string
	ClientDictionaries map[string]string
	SessionSConns      []*HaPoolConfig
	RequestProcessors  []*RARequestProcessor
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
		self.SessionSConns = make([]*HaPoolConfig, len(*jsnCfg.Sessions_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sessions_conns {
			self.SessionSConns[idx] = NewDfltHaPoolConfig()
			self.SessionSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(RARequestProcessor)
			var haveID bool
			for _, rpSet := range self.RequestProcessors {
				if reqProcJsn.Id != nil && rpSet.Id == *reqProcJsn.Id {
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

// One Diameter request processor configuration
type RARequestProcessor struct {
	Id                string
	Tenant            RSRParsers
	Filters           []string
	Timezone          string
	Flags             utils.StringMap
	ContinueOnSuccess bool
	RequestFields     []*FCTemplate
	ReplyFields       []*FCTemplate
}

func (self *RARequestProcessor) loadFromJsonCfg(jsnCfg *RAReqProcessorJsnCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Id != nil {
		self.Id = *jsnCfg.Id
	}
	if jsnCfg.Filters != nil {
		self.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			self.Filters[i] = fltr
		}
	}
	if jsnCfg.Flags != nil {
		self.Flags = utils.StringMapFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Continue_on_success != nil {
		self.ContinueOnSuccess = *jsnCfg.Continue_on_success
	}
	if jsnCfg.Tenant != nil {
		if self.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, true, separator); err != nil {
			return err
		}
	}
	if jsnCfg.Timezone != nil {
		self.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Request_fields != nil {
		if self.RequestFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Request_fields, separator); err != nil {
			return
		}
	}
	if jsnCfg.Reply_fields != nil {
		if self.ReplyFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Reply_fields, separator); err != nil {
			return
		}
	}
	return nil
}
