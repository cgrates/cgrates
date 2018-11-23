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

type HttpAgentCfgs []*HttpAgentCfg

func (hcfgs *HttpAgentCfgs) loadFromJsonCfg(jsnHttpAgntCfg *[]*HttpAgentJsonCfg, separator string) (err error) {
	if jsnHttpAgntCfg == nil {
		return nil
	}
	for _, jsnCfg := range *jsnHttpAgntCfg {
		hac := new(HttpAgentCfg)
		var haveID bool
		if jsnCfg.Id != nil {
			for _, val := range *hcfgs {
				if val.ID == *jsnCfg.Id {
					hac = val
					haveID = true
					break
				}
			}
		}

		if err := hac.loadFromJsonCfg(jsnCfg, separator); err != nil {
			return err
		}
		if !haveID {
			*hcfgs = append(*hcfgs, hac)
		}

	}
	return nil
}

type HttpAgentCfg struct {
	ID                string // identifier for the agent, so we can update it's processors
	Url               string
	SessionSConns     []*HaPoolConfig
	RequestPayload    string
	ReplyPayload      string
	RequestProcessors []*HttpAgntProcCfg
}

func (ca *HttpAgentCfg) appendHttpAgntProcCfgs(hps *[]*HttpAgentProcessorJsnCfg, separator string) (err error) {
	if hps == nil {
		return
	}
	for _, reqProcJsn := range *hps {
		rp := new(HttpAgntProcCfg)
		var haveID bool
		if reqProcJsn.Id != nil {
			for _, rpSet := range ca.RequestProcessors {
				if rpSet.Id == *reqProcJsn.Id {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
		}
		if err := rp.loadFromJsonCfg(reqProcJsn, separator); err != nil {
			return err
		}
		if !haveID {
			ca.RequestProcessors = append(ca.RequestProcessors, rp)
		}
	}
	return nil
}

func (ca *HttpAgentCfg) loadFromJsonCfg(jsnCfg *HttpAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Id != nil {
		ca.ID = *jsnCfg.Id
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
	if jsnCfg.Request_payload != nil {
		ca.RequestPayload = *jsnCfg.Request_payload
	}
	if jsnCfg.Reply_payload != nil {
		ca.ReplyPayload = *jsnCfg.Reply_payload
	}
	if err = ca.appendHttpAgntProcCfgs(jsnCfg.Request_processors, separator); err != nil {
		return err
	}
	return nil
}

type HttpAgntProcCfg struct {
	Id                string
	Filters           []string
	Tenant            RSRParsers
	Timezone          string
	Flags             utils.StringMap
	ContinueOnSuccess bool
	RequestFields     []*FCTemplate
	ReplyFields       []*FCTemplate
}

func (ha *HttpAgntProcCfg) loadFromJsonCfg(jsnCfg *HttpAgentProcessorJsnCfg, separator string) (err error) {
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
	if jsnCfg.Tenant != nil {
		if ha.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, true, separator); err != nil {
			return err
		}
	}
	if jsnCfg.Timezone != nil {
		ha.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Flags != nil {
		ha.Flags = utils.StringMapFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Continue_on_success != nil {
		ha.ContinueOnSuccess = *jsnCfg.Continue_on_success
	}
	if jsnCfg.Request_fields != nil {
		if ha.RequestFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Request_fields, separator); err != nil {
			return
		}
	}
	if jsnCfg.Reply_fields != nil {
		if ha.ReplyFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Reply_fields, separator); err != nil {
			return
		}
	}
	return nil
}
