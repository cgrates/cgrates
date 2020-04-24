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

type DNSAgentCfg struct {
	Enabled           bool
	Listen            string
	ListenNet         string // udp or tcp
	SessionSConns     []string
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
		da.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, connID := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				da.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				da.SessionSConns[idx] = connID
			}
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

func (da *DNSAgentCfg) AsMapInterface(separator string) map[string]interface{} {
	requestProcessors := make([]map[string]interface{}, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}

	return map[string]interface{}{
		utils.EnabledCfg:           da.Enabled,
		utils.ListenCfg:            da.Listen,
		utils.ListenNetCfg:         da.ListenNet,
		utils.SessionSConnsCfg:     da.SessionSConns,
		utils.TimezoneCfg:          da.Timezone,
		utils.RequestProcessorsCfg: requestProcessors,
	}

}

// RequestProcessor is the request processor configuration
type RequestProcessor struct {
	ID            string
	Tenant        RSRParsers
	Filters       []string
	Flags         utils.FlagsWithParams
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

func (rp *RequestProcessor) AsMapInterface(separator string) map[string]interface{} {
	replyFields := make([]map[string]interface{}, len(rp.ReplyFields))
	for i, item := range rp.ReplyFields {
		replyFields[i] = item.AsMapInterface(separator)
	}

	requestFields := make([]map[string]interface{}, len(rp.RequestFields))
	for i, item := range rp.RequestFields {
		requestFields[i] = item.AsMapInterface(separator)
	}
	var tenant string
	if rp.Tenant != nil {
		values := make([]string, len(rp.Tenant))
		for i, item := range rp.Tenant {
			values[i] = item.Rules
		}
		tenant = strings.Join(values, separator)
	}

	flags := make(map[string][]string, len(rp.Flags))
	for key, item := range rp.Flags {
		flags[key] = item
	}

	return map[string]interface{}{
		utils.IDCfg:            rp.ID,
		utils.TenantCfg:        tenant,
		utils.FiltersCfg:       rp.Filters,
		utils.FlagsCfg:         flags,
		utils.TimezoneCfgC:     rp.Timezone,
		utils.RequestFieldsCfg: requestFields,
		utils.ReplyFieldsCfg:   replyFields,
	}

}
