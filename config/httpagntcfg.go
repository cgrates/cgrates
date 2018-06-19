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

type HttpAgentCfg struct {
	Url               string
	SessionSConns     []*HaPoolConfig
	Tenant            utils.RSRFields
	Timezone          string
	RequestPayload    string
	ReplyPayload      string
	RequestProcessors []*HttpAgntProcCfg
}

func (ca *HttpAgentCfg) loadFromJsonCfg(jsnCfg *HttpAgentJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Url != nil {
		ca.Url = *jsnCfg.Url
	}
	if jsnCfg.Sessions_conns != nil {
		ca.SessionSConns = make([]*HaPoolConfig, len(*jsnCfg.Sessions_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sessions_conns {
			ca.SessionSConns[idx] = NewDfltHaPoolConfig()
			ca.SessionSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Tenant != nil {
		if ca.Tenant, err = utils.ParseRSRFields(*jsnCfg.Tenant,
			utils.INFIELD_SEP); err != nil {
			return
		}
	}
	if jsnCfg.Timezone != nil {
		ca.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Request_payload != nil {
		ca.RequestPayload = *jsnCfg.Request_payload
	}
	if jsnCfg.Reply_payload != nil {
		ca.ReplyPayload = *jsnCfg.Reply_payload
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(HttpAgntProcCfg)
			var haveID bool
			for _, rpSet := range ca.RequestProcessors {
				if reqProcJsn.Id != nil && rpSet.Id == *reqProcJsn.Id {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
			if err := rp.loadFromJsonCfg(reqProcJsn); err != nil {
				return nil
			}
			if !haveID {
				ca.RequestProcessors = append(ca.RequestProcessors, rp)
			}
		}
	}
	return nil
}

type HttpAgntProcCfg struct {
	Id                string
	Filters           []string
	Flags             utils.StringMap
	ContinueOnSuccess bool
	RequestFields     []*CfgCdrField
	ReplyFields       []*CfgCdrField
}

func (ha *HttpAgntProcCfg) loadFromJsonCfg(jsnCfg *HttpAgentProcessorJsnCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Id != nil {
		ha.Id = *jsnCfg.Id
	}
	if jsnCfg.Filters != nil {
		ha.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			ha.Filters[i] = fltr
		}
	}
	if jsnCfg.Flags != nil {
		ha.Flags = utils.StringMapFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Continue_on_success != nil {
		ha.ContinueOnSuccess = *jsnCfg.Continue_on_success
	}
	if jsnCfg.Request_fields != nil {
		if ha.RequestFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Request_fields); err != nil {
			return
		}
	}
	if jsnCfg.Reply_fields != nil {
		if ha.ReplyFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Reply_fields); err != nil {
			return
		}
	}
	return nil
}
