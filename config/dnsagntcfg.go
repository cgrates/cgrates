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
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type Listener struct {
	Address string
	Network string // udp or tcp
}

// DNSAgentCfg the config section that describes the DNS Agent
type DNSAgentCfg struct {
	Enabled           bool
	Listeners         []Listener
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
	return da.loadFromJSONCfg(jsnDNSCfg)
}

func (da *DNSAgentCfg) loadFromJSONCfg(jsnCfg *DNSAgentJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		da.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listeners != nil {
		da.Listeners = make([]Listener, 0, len(*jsnCfg.Listeners))
		for _, listnr := range *jsnCfg.Listeners {
			var ls Listener
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
		da.SessionSConns = tagInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	da.RequestProcessors, err = appendRequestProcessors(da.RequestProcessors, jsnCfg.Request_processors)
	return
}

func (lstn *Listener) AsMapInterface() map[string]any {
	return map[string]any{
		utils.AddressCfg: lstn.Address,
		utils.NetworkCfg: lstn.Network,
	}
}

// AsMapInterface returns the config as a map[string]any
func (da DNSAgentCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.EnabledCfg:  da.Enabled,
		utils.TimezoneCfg: da.Timezone,
	}

	listeners := make([]map[string]any, len(da.Listeners))
	for i, item := range da.Listeners {
		listeners[i] = item.AsMapInterface()
	}
	mp[utils.ListenersCfg] = listeners

	requestProcessors := make([]map[string]any, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	mp[utils.RequestProcessorsCfg] = requestProcessors

	if da.SessionSConns != nil {
		mp[utils.SessionSConnsCfg] = stripInternalConns(da.SessionSConns)
	}
	return mp
}

func (DNSAgentCfg) SName() string            { return DNSAgentJSON }
func (da DNSAgentCfg) CloneSection() Section { return da.Clone() }

// Clone returns a deep copy of DNSAgentCfg
func (da DNSAgentCfg) Clone() (cln *DNSAgentCfg) {
	cln = &DNSAgentCfg{
		Enabled:   da.Enabled,
		Listeners: da.Listeners,
		Timezone:  da.Timezone,
	}

	if da.Listeners != nil {
		cln.Listeners = make([]Listener, len(da.Listeners))
		copy(cln.Listeners, da.Listeners)
	}
	if da.SessionSConns != nil {
		cln.SessionSConns = slices.Clone(da.SessionSConns)
	}
	if da.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(da.RequestProcessors))
		for i, req := range da.RequestProcessors {
			cln.RequestProcessors[i] = req.Clone()
		}
	}
	return
}

type ListenerJsnCfg struct {
	Address *string
	Network *string
}

// DNSAgentJsonCfg
type DNSAgentJsonCfg struct {
	Enabled            *bool
	Listeners          *[]*ListenerJsnCfg
	Sessions_conns     *[]string
	Timezone           *string
	Request_processors *[]*ReqProcessorJsnCfg
}

func diffDNSAgentJsonCfg(d *DNSAgentJsonCfg, v1, v2 *DNSAgentCfg) *DNSAgentJsonCfg {
	if d == nil {
		d = new(DNSAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}

	minLen := len(v1.Listeners)
	if len(v2.Listeners) < minLen {
		minLen = len(v2.Listeners)
	}

	diffListeners := &[]*ListenerJsnCfg{}

	for i := 0; i < minLen; i++ {
		if v1.Listeners[i].Address != v2.Listeners[i].Address ||
			v1.Listeners[i].Network != v2.Listeners[i].Network {
			*diffListeners = append(*diffListeners, &ListenerJsnCfg{
				Address: utils.StringPointer(v2.Listeners[i].Address),
				Network: utils.StringPointer(v2.Listeners[i].Network),
			})
		}
	}

	if len(v1.Listeners) > minLen {
		for i := minLen; i < len(v1.Listeners); i++ {
			*diffListeners = append(*diffListeners, &ListenerJsnCfg{
				Address: utils.StringPointer(v1.Listeners[i].Address),
				Network: utils.StringPointer(v1.Listeners[i].Network),
			})
		}
	}

	if len(v2.Listeners) > minLen {
		for i := minLen; i < len(v2.Listeners); i++ {
			*diffListeners = append(*diffListeners, &ListenerJsnCfg{
				Address: utils.StringPointer(v2.Listeners[i].Address),
				Network: utils.StringPointer(v2.Listeners[i].Network),
			})
		}
	}

	d.Listeners = diffListeners

	if !slices.Equal(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(stripInternalConns(v2.SessionSConns))
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	d.Request_processors = diffReqProcessorsJsnCfg(d.Request_processors, v1.RequestProcessors, v2.RequestProcessors)
	return d
}
