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

type ConectoAgentCfg struct {
	Enabled           bool
	HttpUrl           string
	SessionSConns     []*HaPoolConfig
	Timezone          string
	RequestProcessors []*CncProcessorCfg
}

func (ca *ConectoAgentCfg) loadFromJsonCfg(jsnCfg *ConectoAgentJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		ca.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Http_url != nil {
		ca.HttpUrl = *jsnCfg.Http_url
	}
	if jsnCfg.Sessions_conns != nil {
		ca.SessionSConns = make([]*HaPoolConfig, len(*jsnCfg.Sessions_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sessions_conns {
			ca.SessionSConns[idx] = NewDfltHaPoolConfig()
			ca.SessionSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Timezone != nil {
		ca.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(CncProcessorCfg)
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

type CncProcessorCfg struct {
	Id            string
	DryRun        bool
	Filters       []string
	Flags         utils.StringMap
	RequestFields []*CfgCdrField
	ReplyFields   []*CfgCdrField
}

func (cp *CncProcessorCfg) loadFromJsonCfg(jsnCfg *CncProcessorJsnCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Id != nil {
		cp.Id = *jsnCfg.Id
	}
	if jsnCfg.Dry_run != nil {
		cp.DryRun = *jsnCfg.Dry_run
	}
	if jsnCfg.Filters != nil {
		cp.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			cp.Filters[i] = fltr
		}
	}
	if jsnCfg.Flags != nil {
		cp.Flags = utils.StringMapFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Request_fields != nil {
		if cp.RequestFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Request_fields); err != nil {
			return
		}
	}
	if jsnCfg.Reply_fields != nil {
		if cp.ReplyFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Reply_fields); err != nil {
			return
		}
	}
	return nil
}
