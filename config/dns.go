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
	StatSConns        []string
	ThresholdSConns   []string
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
	if jsnCfg.SessionSConns != nil {
		da.SessionSConns = tagInternalConns(*jsnCfg.SessionSConns, utils.MetaSessionS)
	}
	if jsnCfg.StatSConns != nil {
		da.StatSConns = tagInternalConns(*jsnCfg.StatSConns, utils.MetaStats)
	}
	if jsnCfg.ThresholdSConns != nil {
		da.ThresholdSConns = tagInternalConns(*jsnCfg.ThresholdSConns, utils.MetaThresholds)
	}
	da.RequestProcessors, err = appendRequestProcessors(da.RequestProcessors, jsnCfg.RequestProcessors)
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
	listeners := make([]map[string]any, len(da.Listeners))
	for i, item := range da.Listeners {
		listeners[i] = item.AsMapInterface()
	}
	requestProcessors := make([]map[string]any, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	mp := map[string]any{
		utils.EnabledCfg:           da.Enabled,
		utils.ListenersCfg:         listeners,
		utils.SessionSConnsCfg:     stripInternalConns(da.SessionSConns),
		utils.StatSConnsCfg:        stripInternalConns(da.StatSConns),
		utils.ThresholdSConnsCfg:   stripInternalConns(da.ThresholdSConns),
		utils.TimezoneCfg:          da.Timezone,
		utils.RequestProcessorsCfg: requestProcessors,
	}
	return mp
}

func (DNSAgentCfg) SName() string            { return DNSAgentJSON }
func (da DNSAgentCfg) CloneSection() Section { return da.Clone() }

// Clone returns a deep copy of DNSAgentCfg
func (da DNSAgentCfg) Clone() *DNSAgentCfg {
	clone := &DNSAgentCfg{
		Enabled:         da.Enabled,
		Listeners:       slices.Clone(da.Listeners),
		SessionSConns:   slices.Clone(da.SessionSConns),
		StatSConns:      slices.Clone(da.StatSConns),
		ThresholdSConns: slices.Clone(da.ThresholdSConns),
		Timezone:        da.Timezone,
	}
	if da.RequestProcessors != nil {
		clone.RequestProcessors = make([]*RequestProcessor, len(da.RequestProcessors))
		for i, req := range da.RequestProcessors {
			clone.RequestProcessors[i] = req.Clone()
		}
	}
	return clone
}

type ListenerJsnCfg struct {
	Address *string
	Network *string
}

// DNSAgentJsonCfg
type DNSAgentJsonCfg struct {
	Enabled           *bool                  `json:"enabled"`
	Listeners         *[]*ListenerJsnCfg     `json:"listeners"`
	SessionSConns     *[]string              `json:"sessions_conns"`
	StatSConns        *[]string              `json:"stats_conns"`
	ThresholdSConns   *[]string              `json:"thresholds_conns"`
	Timezone          *string                `json:"timezone"`
	RequestProcessors *[]*ReqProcessorJsnCfg `json:"request_processors"`
}

func diffDNSAgentJsonCfg(d *DNSAgentJsonCfg, v1, v2 *DNSAgentCfg) *DNSAgentJsonCfg {
	if d == nil {
		d = new(DNSAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}

	minLen := min(len(v2.Listeners), len(v1.Listeners))

	diffListeners := &[]*ListenerJsnCfg{}

	for i := range minLen {
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
		d.SessionSConns = utils.SliceStringPointer(stripInternalConns(v2.SessionSConns))
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.StatSConns = utils.SliceStringPointer(stripInternalConns(v2.StatSConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.ThresholdSConns = utils.SliceStringPointer(stripInternalConns(v2.ThresholdSConns))
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	d.RequestProcessors = diffReqProcessorsJsnCfg(d.RequestProcessors, v1.RequestProcessors, v2.RequestProcessors)
	return d
}
