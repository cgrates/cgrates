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

type DNSAgentCfg struct {
	Enabled           bool
	Listen            string
	ListenNet         string // udp or tcp
	SessionSConns     []*RemoteHost
	Timezone          string
	RequestProcessors []*RequestProcessor
}

func (da *DNSAgentCfg) loadFromJsonCfg(jsnCfg *DNSAgentJsonCfg, sep string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		da.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_net != nil {
		da.ListenNet = *jsnCfg.Listen_net
	}
	if jsnCfg.Listen != nil {
		da.Listen = *jsnCfg.Listen
	}
	if jsnCfg.Timezone != nil {
		da.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Sessions_conns != nil {
		da.SessionSConns = make([]*RemoteHost, len(*jsnCfg.Sessions_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sessions_conns {
			da.SessionSConns[idx] = NewDfltRemoteHost()
			da.SessionSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(RequestProcessor)
			var haveID bool
			for _, rpSet := range da.RequestProcessors {
				if reqProcJsn.ID != nil && rpSet.ID == *reqProcJsn.ID {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
			if err := rp.loadFromJsonCfg(reqProcJsn, sep); err != nil {
				return nil
			}
			if !haveID {
				da.RequestProcessors = append(da.RequestProcessors, rp)
			}
		}
	}
	return nil
}

// One  request processor configuration
type RequestProcessor struct {
	ID            string
	Tenant        RSRParsers
	Filters       []string
	Flags         utils.FlagsWithParams
	Continue      bool
	Timezone      string
	RequestFields []*FCTemplate
	ReplyFields   []*FCTemplate
}

func (rp *RequestProcessor) loadFromJsonCfg(jsnCfg *ReqProcessorJsnCfg, sep string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.ID != nil {
		rp.ID = *jsnCfg.ID
	}
	if jsnCfg.Filters != nil {
		rp.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			rp.Filters[i] = fltr
		}
	}
	if jsnCfg.Flags != nil {
		if rp.Flags, err = utils.FlagsWithParamsFromSlice(*jsnCfg.Flags); err != nil {
			return
		}
	}
	if jsnCfg.Timezone != nil {
		rp.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Continue != nil {
		rp.Continue = *jsnCfg.Continue
	}
	if jsnCfg.Tenant != nil {
		if rp.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, true, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Request_fields != nil {
		if rp.RequestFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Request_fields, sep); err != nil {
			return
		}
	}
	if jsnCfg.Reply_fields != nil {
		if rp.ReplyFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Reply_fields, sep); err != nil {
			return
		}
	}
	return nil
}
