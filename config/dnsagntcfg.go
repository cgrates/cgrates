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

type DnsListener struct {
	Address string
	Network string // udp or tcp
}

// DNSAgentCfg the config section that describes the DNS Agent
type DNSAgentCfg struct {
	Enabled           bool
	Listeners         []DnsListener
	SessionSConns     []string
	Timezone          string
	RequestProcessors []*RequestProcessor
}

func (da *DNSAgentCfg) loadFromJSONCfg(jsnCfg *DNSAgentJsonCfg, sep string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		da.Enabled = *jsnCfg.Enabled
	}

	if jsnCfg.Listeners != nil {
		da.Listeners = make([]DnsListener, 0, len(*jsnCfg.Listeners))
		for _, listnr := range *jsnCfg.Listeners {
			var ls DnsListener
			if listnr.Address != nil {
				ls.Address = *listnr.Address
			}
			if listnr.Network != nil {
				ls.Network = *listnr.Network
			}
			da.Listeners = append(da.Listeners, ls)
		}
	}
	if jsnCfg.Timezone != nil {
		da.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Sessions_conns != nil {
		da.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, connID := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			da.SessionSConns[idx] = connID
			if connID == utils.MetaInternal {
				da.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
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
			if err = rp.loadFromJSONCfg(reqProcJsn, sep); err != nil {
				return
			}
			if !haveID {
				da.RequestProcessors = append(da.RequestProcessors, rp)
			}
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (lstn *DnsListener) AsMapInterface(separator string) map[string]any {
	return map[string]any{
		utils.AddressCfg: lstn.Address,
		utils.NetworkCfg: lstn.Network,
	}

}

// AsMapInterface returns the config as a map[string]any
func (da *DNSAgentCfg) AsMapInterface(separator string) (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.EnabledCfg:  da.Enabled,
		utils.TimezoneCfg: da.Timezone,
	}

	listeners := make([]map[string]any, len(da.Listeners))
	for i, item := range da.Listeners {
		listeners[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.ListenersCfg] = listeners

	requestProcessors := make([]map[string]any, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.RequestProcessorsCfg] = requestProcessors

	if da.SessionSConns != nil {
		sessionSConns := make([]string, len(da.SessionSConns))
		for i, item := range da.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
	}
	return
}

// Clone returns a deep copy of DNSAgentCfg
func (da DNSAgentCfg) Clone() (cln *DNSAgentCfg) {
	cln = &DNSAgentCfg{
		Enabled:   da.Enabled,
		Listeners: da.Listeners,
		Timezone:  da.Timezone,
	}

	if da.Listeners != nil {
		cln.Listeners = make([]DnsListener, len(da.Listeners))
		copy(cln.Listeners, da.Listeners)
	}

	if da.SessionSConns != nil {
		cln.SessionSConns = make([]string, len(da.SessionSConns))
		copy(cln.SessionSConns, da.SessionSConns)
	}
	if da.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(da.RequestProcessors))
		for i, req := range da.RequestProcessors {
			cln.RequestProcessors[i] = req.Clone()
		}
	}
	return
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

func (rp *RequestProcessor) loadFromJSONCfg(jsnCfg *ReqProcessorJsnCfg, sep string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.ID != nil {
		rp.ID = *jsnCfg.ID
	}
	if jsnCfg.Filters != nil {
		rp.Filters = make([]string, len(*jsnCfg.Filters))
		copy(rp.Filters, *jsnCfg.Filters)
	}
	if jsnCfg.Flags != nil {
		rp.Flags = utils.FlagsWithParamsFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Timezone != nil {
		rp.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Tenant != nil {
		if rp.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Request_fields != nil {
		if rp.RequestFields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Request_fields, sep); err != nil {
			return
		}
	}
	if jsnCfg.Reply_fields != nil {
		if rp.ReplyFields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Reply_fields, sep); err != nil {
			return
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (rp *RequestProcessor) AsMapInterface(separator string) (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.IDCfg:       rp.ID,
		utils.FiltersCfg:  rp.Filters,
		utils.FlagsCfg:    rp.Flags.SliceFlags(),
		utils.TimezoneCfg: rp.Timezone,
	}
	if rp.Tenant != nil {
		initialMP[utils.TenantCfg] = rp.Tenant.GetRule(separator)
	}
	if rp.RequestFields != nil {
		requestFields := make([]map[string]any, len(rp.RequestFields))
		for i, item := range rp.RequestFields {
			requestFields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.RequestFieldsCfg] = requestFields
	}
	if rp.ReplyFields != nil {
		replyFields := make([]map[string]any, len(rp.ReplyFields))
		for i, item := range rp.ReplyFields {
			replyFields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.ReplyFieldsCfg] = replyFields
	}
	return
}

// Clone returns a deep copy of APIBanCfg
func (rp RequestProcessor) Clone() (cln *RequestProcessor) {
	cln = &RequestProcessor{
		ID:       rp.ID,
		Tenant:   rp.Tenant.Clone(),
		Flags:    rp.Flags.Clone(),
		Timezone: rp.Timezone,
	}
	if rp.Filters != nil {
		cln.Filters = make([]string, len(rp.Filters))
		copy(cln.Filters, rp.Filters)

	}
	if rp.RequestFields != nil {
		cln.RequestFields = make([]*FCTemplate, len(rp.RequestFields))
		for i, rf := range rp.RequestFields {
			cln.RequestFields[i] = rf.Clone()
		}
	}
	if rp.ReplyFields != nil {
		cln.ReplyFields = make([]*FCTemplate, len(rp.ReplyFields))
		for i, rf := range rp.ReplyFields {
			cln.ReplyFields[i] = rf.Clone()
		}
	}
	return
}
