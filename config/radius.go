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
	"maps"
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type RadiusListener struct {
	AuthAddr string
	AcctAddr string
	Network  string // udp or tcp
}

// AsMapInterface returns the config as a map[string]any.
func (lstn *RadiusListener) AsMapInterface() map[string]any {
	return map[string]any{
		utils.AuthAddrCfg: lstn.AuthAddr,
		utils.AcctAddrCfg: lstn.AcctAddr,
		utils.NetworkCfg:  lstn.Network,
	}
}

// RadiusAgentCfg the config section that describes the Radius Agent
type RadiusAgentCfg struct {
	Enabled            bool
	Listeners          []RadiusListener
	ClientSecrets      map[string]string
	ClientDictionaries map[string][]string
	SessionSConns      []string
	StatSConns         []string
	ThresholdSConns    []string
	RequestProcessors  []*RequestProcessor
}

// loadRadiusAgentCfg loads the RadiusAgent section of the configuration
func (ra *RadiusAgentCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnRACfg := new(RadiusAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, RadiusAgentJSON, jsnRACfg); err != nil {
		return
	}
	return ra.loadFromJSONCfg(jsnRACfg)
}

func (ra *RadiusAgentCfg) loadFromJSONCfg(jc *RadiusAgentJsonCfg) (err error) {
	if jc == nil {
		return nil
	}
	if jc.Enabled != nil {
		ra.Enabled = *jc.Enabled
	}
	if jc.Listeners != nil {
		ra.Listeners = make([]RadiusListener, 0, len(*jc.Listeners))
		for _, jl := range *jc.Listeners {
			var rl RadiusListener
			if jl.AuthAddress != nil {
				rl.AuthAddr = *jl.AuthAddress
			}
			if jl.AcctAddress != nil {
				rl.AcctAddr = *jl.AcctAddress
			}
			if jl.Network != nil {
				rl.Network = *jl.Network
			}
			ra.Listeners = append(ra.Listeners, rl)
		}
	}
	maps.Copy(ra.ClientSecrets, jc.ClientSecrets)
	maps.Copy(ra.ClientDictionaries, jc.ClientDictionaries)
	if jc.SessionSConns != nil {
		ra.SessionSConns = tagInternalConns(*jc.SessionSConns, utils.MetaSessionS)
	}
	if jc.StatSConns != nil {
		ra.StatSConns = tagInternalConns(*jc.StatSConns, utils.MetaStats)
	}
	if jc.ThresholdSConns != nil {
		ra.ThresholdSConns = tagInternalConns(*jc.ThresholdSConns, utils.MetaThresholds)
	}
	ra.RequestProcessors, err = appendRequestProcessors(ra.RequestProcessors, jc.RequestProcessors)
	return
}

// AsMapInterface returns the config as a map[string]any
func (ra RadiusAgentCfg) AsMapInterface() any {
	requestProcessors := make([]map[string]any, len(ra.RequestProcessors))
	for i, item := range ra.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	listeners := make([]map[string]any, len(ra.Listeners))
	for i, item := range ra.Listeners {
		listeners[i] = item.AsMapInterface()
	}
	mp := map[string]any{
		utils.EnabledCfg:            ra.Enabled,
		utils.ListenersCfg:          listeners,
		utils.ClientSecretsCfg:      maps.Clone(ra.ClientSecrets),
		utils.ClientDictionariesCfg: maps.Clone(ra.ClientDictionaries),
		utils.SessionSConnsCfg:      stripInternalConns(ra.SessionSConns),
		utils.StatSConnsCfg:         stripInternalConns(ra.StatSConns),
		utils.ThresholdSConnsCfg:    stripInternalConns(ra.ThresholdSConns),
		utils.RequestProcessorsCfg:  requestProcessors,
	}
	return mp
}

func (RadiusAgentCfg) SName() string            { return RadiusAgentJSON }
func (ra RadiusAgentCfg) CloneSection() Section { return ra.Clone() }

// Clone returns a deep copy of RadiusAgentCfg
func (ra RadiusAgentCfg) Clone() *RadiusAgentCfg {
	clone := &RadiusAgentCfg{
		Enabled:         ra.Enabled,
		Listeners:       slices.Clone(ra.Listeners),
		ClientSecrets:   maps.Clone(ra.ClientSecrets),
		SessionSConns:   slices.Clone(ra.SessionSConns),
		StatSConns:      slices.Clone(ra.StatSConns),
		ThresholdSConns: slices.Clone(ra.ThresholdSConns),
	}
	if ra.ClientDictionaries != nil {
		clone.ClientDictionaries = make(map[string][]string, len(ra.ClientDictionaries))
		for key, val := range ra.ClientDictionaries {
			clone.ClientDictionaries[key] = slices.Clone(val)
		}
	}
	if ra.RequestProcessors != nil {
		clone.RequestProcessors = make([]*RequestProcessor, len(ra.RequestProcessors))
		for i, req := range ra.RequestProcessors {
			clone.RequestProcessors[i] = req.Clone()
		}
	}
	return clone
}

type RadiusListenerJsonCfg struct {
	Network     *string `json:"network"`
	AuthAddress *string `json:"auth_address"`
	AcctAddress *string `json:"acct_address"`
}

// Radius Agent configuration section
type RadiusAgentJsonCfg struct {
	Enabled            *bool                     `json:"enabled"`
	Listeners          *[]*RadiusListenerJsonCfg `json:"listeners"`
	ClientSecrets      map[string]string         `json:"client_secrets"`
	ClientDictionaries map[string][]string       `json:"client_dictionaries"`
	SessionSConns      *[]string                 `json:"sessions_conns"`
	StatSConns         *[]string                 `json:"stats_conns"`
	ThresholdSConns    *[]string                 `json:"thresholds_conns"`
	RequestProcessors  *[]*ReqProcessorJsnCfg    `json:"request_processors"`
}

func diffRadiusAgentJsonCfg(d *RadiusAgentJsonCfg, v1, v2 *RadiusAgentCfg) *RadiusAgentJsonCfg {
	if d == nil {
		d = new(RadiusAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !slices.Equal(v1.Listeners, v2.Listeners) {
		listeners := make([]*RadiusListenerJsonCfg, len(v2.Listeners))
		for i, listener := range v2.Listeners {
			listeners[i] = &RadiusListenerJsonCfg{
				AuthAddress: utils.StringPointer(listener.AuthAddr),
				AcctAddress: utils.StringPointer(listener.AcctAddr),
				Network:     utils.StringPointer(listener.Network),
			}
		}
		d.Listeners = &listeners
	}
	d.ClientSecrets = diffMapString(d.ClientSecrets, v1.ClientSecrets, v2.ClientSecrets)
	d.ClientDictionaries = diffMapStringSlice(d.ClientDictionaries, v1.ClientDictionaries, v2.ClientDictionaries)
	if !slices.Equal(v1.SessionSConns, v2.SessionSConns) {
		d.SessionSConns = utils.SliceStringPointer(stripInternalConns(v2.SessionSConns))
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.StatSConns = utils.SliceStringPointer(stripInternalConns(v2.StatSConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.ThresholdSConns = utils.SliceStringPointer(stripInternalConns(v2.ThresholdSConns))
	}
	d.RequestProcessors = diffReqProcessorsJsnCfg(d.RequestProcessors, v1.RequestProcessors, v2.RequestProcessors)
	return d
}

func diffMapString(d, v1, v2 map[string]string) map[string]string {
	if d == nil {
		d = make(map[string]string)
	}
	for k, v := range v2 {
		if val, has := v1[k]; !has || val != v {
			d[k] = v
		}
	}
	return d
}

func diffMapStringSlice(d, v1, v2 map[string][]string) map[string][]string {
	if d == nil {
		d = make(map[string][]string)
	}
	for k, v := range v2 {
		if val, has := v1[k]; !has || !slices.Equal(val, v) {
			d[k] = v
		}
	}
	return d
}
