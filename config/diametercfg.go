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

type DiameterAgentCfg struct {
	Enabled           bool   // enables the diameter agent: <true|false>
	ListenNet         string // sctp or tcp
	Listen            string // address where to listen for diameter requests <x.y.z.y:1234>
	DictionariesPath  string
	SessionSConns     []*HaPoolConfig // connections towards SMG component
	OriginHost        string
	OriginRealm       string
	VendorId          int
	ProductName       string
	MaxActiveReqs     int // limit the maximum number of requests processed
	ASRTempalte       string
	Templates         map[string][]*FCTemplate
	RequestProcessors []*DARequestProcessor
}

func (da *DiameterAgentCfg) loadFromJsonCfg(jsnCfg *DiameterAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		da.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen != nil {
		da.Listen = *jsnCfg.Listen
	}
	if jsnCfg.Listen_net != nil {
		da.ListenNet = *jsnCfg.Listen_net
	}
	if jsnCfg.Dictionaries_path != nil {
		da.DictionariesPath = *jsnCfg.Dictionaries_path
	}
	if jsnCfg.Sessions_conns != nil {
		da.SessionSConns = make([]*HaPoolConfig, len(*jsnCfg.Sessions_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sessions_conns {
			da.SessionSConns[idx] = NewDfltHaPoolConfig()
			da.SessionSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Origin_host != nil {
		da.OriginHost = *jsnCfg.Origin_host
	}
	if jsnCfg.Origin_realm != nil {
		da.OriginRealm = *jsnCfg.Origin_realm
	}
	if jsnCfg.Vendor_id != nil {
		da.VendorId = *jsnCfg.Vendor_id
	}
	if jsnCfg.Product_name != nil {
		da.ProductName = *jsnCfg.Product_name
	}
	if jsnCfg.Max_active_requests != nil {
		da.MaxActiveReqs = *jsnCfg.Max_active_requests
	}
	if jsnCfg.Asr_template != nil {
		da.ASRTempalte = *jsnCfg.Asr_template
	}
	if jsnCfg.Templates != nil {
		if da.Templates == nil {
			da.Templates = make(map[string][]*FCTemplate)
		}
		for k, jsnTpls := range jsnCfg.Templates {
			if da.Templates[k], err = FCTemplatesFromFCTemplatesJsonCfg(jsnTpls, separator); err != nil {
				return
			}
		}
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(DARequestProcessor)
			var haveID bool
			for _, rpSet := range da.RequestProcessors {
				if reqProcJsn.Id != nil && rpSet.ID == *reqProcJsn.Id {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
			if err := rp.loadFromJsonCfg(reqProcJsn, separator); err != nil {
				return nil
			}
			if !haveID {
				da.RequestProcessors = append(da.RequestProcessors, rp)
			}
		}
	}
	return nil
}

// One Diameter request processor configuration
type DARequestProcessor struct {
	ID                string
	Tenant            RSRParsers
	Filters           []string
	Flags             utils.StringMap
	Timezone          string // timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	ContinueOnSuccess bool
	RequestFields     []*FCTemplate
	ReplyFields       []*FCTemplate
}

func (dap *DARequestProcessor) loadFromJsonCfg(jsnCfg *DARequestProcessorJsnCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Id != nil {
		dap.ID = *jsnCfg.Id
	}
	if jsnCfg.Tenant != nil {
		if dap.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, true, separator); err != nil {
			return
		}
	}
	if jsnCfg.Filters != nil {
		dap.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			dap.Filters[i] = fltr
		}
	}
	if jsnCfg.Flags != nil {
		dap.Flags = utils.StringMapFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Timezone != nil {
		dap.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Continue_on_success != nil {
		dap.ContinueOnSuccess = *jsnCfg.Continue_on_success
	}
	if jsnCfg.Request_fields != nil {
		if dap.RequestFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Request_fields, separator); err != nil {
			return
		}
	}
	if jsnCfg.Reply_fields != nil {
		if dap.ReplyFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Reply_fields, separator); err != nil {
			return
		}
	}
	return nil
}
