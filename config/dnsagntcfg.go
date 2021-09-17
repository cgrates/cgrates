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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// DNSAgentCfg the config section that describes the DNS Agent
type DNSAgentCfg struct {
	Enabled           bool
	Listen            string
	ListenNet         string // udp or tcp
	SessionSConns     []string
	Timezone          string
	RequestProcessors []*RequestProcessor
}

// loadDNSAgentCfg loads the DNSAgent section of the configuration
func (da *DNSAgentCfg) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnDNSCfg := new(DNSAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, DNSAgentJSON, jsnDNSCfg); err != nil {
		return
	}
	return da.loadFromJSONCfg(jsnDNSCfg, cfg.generalCfg.RSRSep)
}

func (da *DNSAgentCfg) loadFromJSONCfg(jsnCfg *DNSAgentJsonCfg, sep string) (err error) {
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
		da.SessionSConns = updateBiRPCInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	da.RequestProcessors, err = appendRequestProcessors(da.RequestProcessors, jsnCfg.Request_processors, sep)
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (da DNSAgentCfg) AsMapInterface(separator string) interface{} {
	mp := map[string]interface{}{
		utils.EnabledCfg:   da.Enabled,
		utils.ListenCfg:    da.Listen,
		utils.ListenNetCfg: da.ListenNet,
		utils.TimezoneCfg:  da.Timezone,
	}

	requestProcessors := make([]map[string]interface{}, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	mp[utils.RequestProcessorsCfg] = requestProcessors

	if da.SessionSConns != nil {
		mp[utils.SessionSConnsCfg] = getBiRPCInternalJSONConns(da.SessionSConns)
	}
	return mp
}

func (DNSAgentCfg) SName() string            { return DNSAgentJSON }
func (da DNSAgentCfg) CloneSection() Section { return da.Clone() }

// Clone returns a deep copy of DNSAgentCfg
func (da DNSAgentCfg) Clone() (cln *DNSAgentCfg) {
	cln = &DNSAgentCfg{
		Enabled:   da.Enabled,
		Listen:    da.Listen,
		ListenNet: da.ListenNet,
		Timezone:  da.Timezone,
	}
	if da.SessionSConns != nil {
		cln.SessionSConns = utils.CloneStringSlice(da.SessionSConns)
	}
	if da.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(da.RequestProcessors))
		for i, req := range da.RequestProcessors {
			cln.RequestProcessors[i] = req.Clone()
		}
	}
	return
}

// DNSAgentJsonCfg
type DNSAgentJsonCfg struct {
	Enabled            *bool
	Listen             *string
	Listen_net         *string
	Sessions_conns     *[]string
	Timezone           *string
	Request_processors *[]*ReqProcessorJsnCfg
}

func diffDNSAgentJsonCfg(d *DNSAgentJsonCfg, v1, v2 *DNSAgentCfg, separator string) *DNSAgentJsonCfg {
	if d == nil {
		d = new(DNSAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.Listen != v2.Listen {
		d.Listen = utils.StringPointer(v2.Listen)
	}
	if v1.ListenNet != v2.ListenNet {
		d.Listen_net = utils.StringPointer(v2.ListenNet)
	}
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	d.Request_processors = diffReqProcessorsJsnCfg(d.Request_processors, v1.RequestProcessors, v2.RequestProcessors, separator)
	return d
}
