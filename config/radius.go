/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"fmt"
	"maps"
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// Radius Agent configuration section
type RadiusAgentJsonCfg struct {
	Enabled            *bool                       `json:"enabled"`
	Listeners          *[]*RadiusListenerJsonCfg   `json:"listeners"`
	ClientSecrets      map[string]string           `json:"client_secrets"`
	ClientDictionaries map[string][]string         `json:"client_dictionaries"`
	ClientDaAddresses  map[string]DAClientOptsJson `json:"client_da_addresses"`
	SessionSConns      *[]string                   `json:"sessions_conns"`
	StatSConns         *[]string                   `json:"stats_conns"`
	ThresholdSConns    *[]string                   `json:"thresholds_conns"`
	RequestsCacheKey   *string                     `json:"requests_cache_key"`
	DMRTemplate        *string                     `json:"dmr_template"`
	CoATemplate        *string                     `json:"coa_template"`
	RequestProcessors  *[]*ReqProcessorJsnCfg      `json:"request_processors"`
}

type RadiusListenerJsonCfg struct {
	Network     *string `json:"network"`
	AuthAddress *string `json:"auth_address"`
	AcctAddress *string `json:"acct_address"`
}

type DAClientOptsJson struct {
	Transport *string  `json:"transport"`
	Host      *string  `json:"host"`
	Port      *int     `json:"port"`
	Flags     []string `json:"flags"`
}

// RadiusAgentCfg the config section that describes the Radius Agent
type RadiusAgentCfg struct {
	Enabled            bool
	Listeners          []RadiusListener
	ClientSecrets      map[string]string
	ClientDictionaries map[string][]string
	ClientDaAddresses  map[string]DAClientOpts
	SessionSConns      []string
	StatSConns         []string
	ThresholdSConns    []string
	RequestsCacheKey   utils.RSRParsers
	DMRTemplate        string
	CoATemplate        string
	RequestProcessors  []*RequestProcessor
}

type RadiusListener struct {
	AuthAddr string
	AcctAddr string
	Network  string // udp or tcp
}

type DAClientOpts struct {
	Transport string                // transport protocol for Dynamic Authorization requests <UDP|TCP>.
	Host      string                // alternative host for DA requests
	Port      int                   // port for Dynamic Authorization requests
	Flags     utils.FlagsWithParams // flags (only *log for now)
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
	if len(jc.ClientDaAddresses) != 0 {
		if ra.ClientDaAddresses == nil {
			ra.ClientDaAddresses = make(map[string]DAClientOpts)
		}
		ra.ClientDaAddresses = make(map[string]DAClientOpts, len(jc.ClientDaAddresses))
		for hostKey, clientOpts := range jc.ClientDaAddresses {
			cfg := DAClientOpts{}
			cfg.loadFromJSONCfg(clientOpts, hostKey)
			ra.ClientDaAddresses[hostKey] = cfg
		}
	}
	if jc.SessionSConns != nil {
		ra.SessionSConns = tagInternalConns(*jc.SessionSConns, utils.MetaSessionS)
	}
	if jc.StatSConns != nil {
		ra.StatSConns = tagInternalConns(*jc.StatSConns, utils.MetaStats)
	}
	if jc.ThresholdSConns != nil {
		ra.ThresholdSConns = tagInternalConns(*jc.ThresholdSConns, utils.MetaThresholds)
	}
	if jc.RequestsCacheKey != nil {
		ra.RequestsCacheKey, err = utils.NewRSRParsers(*jc.RequestsCacheKey, utils.RSRSep)
		if err != nil {
			return fmt.Errorf(
				"failed to initialize RSRParsers based %s value: %w",
				utils.RequestsCacheKeyCfg, err,
			)
		}
	}
	if jc.DMRTemplate != nil {
		ra.DMRTemplate = *jc.DMRTemplate
	}
	if jc.CoATemplate != nil {
		ra.CoATemplate = *jc.CoATemplate
	}
	ra.RequestProcessors, err = appendRequestProcessors(ra.RequestProcessors, jc.RequestProcessors)
	return
}

// AsMapInterface returns the config as a map[string]any
func (ra RadiusAgentCfg) AsMapInterface() any {
	listeners := make([]map[string]any, len(ra.Listeners))
	for i, item := range ra.Listeners {
		listeners[i] = item.AsMapInterface()
	}
	clientDaAddresses := make(map[string]any, len(ra.ClientDaAddresses))
	for k, v := range ra.ClientDaAddresses {
		clientDaAddresses[k] = v.AsMapInterface()
	}
	requestProcessors := make([]map[string]any, len(ra.RequestProcessors))
	for i, item := range ra.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	mp := map[string]any{
		utils.EnabledCfg:            ra.Enabled,
		utils.ListenersCfg:          listeners,
		utils.ClientSecretsCfg:      maps.Clone(ra.ClientSecrets),
		utils.ClientDictionariesCfg: maps.Clone(ra.ClientDictionaries),
		utils.ClientDaAddressesCfg:  clientDaAddresses,
		utils.RequestsCacheKeyCfg:   ra.RequestsCacheKey.GetRule(),
		utils.DMRTemplateCfg:        ra.DMRTemplate,
		utils.CoATemplateCfg:        ra.CoATemplate,
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
		Enabled:          ra.Enabled,
		Listeners:        slices.Clone(ra.Listeners),
		ClientSecrets:    maps.Clone(ra.ClientSecrets),
		SessionSConns:    slices.Clone(ra.SessionSConns),
		StatSConns:       slices.Clone(ra.StatSConns),
		ThresholdSConns:  slices.Clone(ra.ThresholdSConns),
		DMRTemplate:      ra.DMRTemplate,
		CoATemplate:      ra.CoATemplate,
		RequestsCacheKey: ra.RequestsCacheKey,
	}
	if ra.ClientDictionaries != nil {
		clone.ClientDictionaries = make(map[string][]string, len(ra.ClientDictionaries))
		for key, val := range ra.ClientDictionaries {
			clone.ClientDictionaries[key] = slices.Clone(val)
		}
	}
	if len(ra.ClientDaAddresses) != 0 {
		clone.ClientDaAddresses = make(map[string]DAClientOpts, len(ra.ClientDaAddresses))
		for k, v := range ra.ClientDaAddresses {
			clone.ClientDaAddresses[k] = *v.Clone()
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

// AsMapInterface returns the config as a map[string]any.
func (l *RadiusListener) AsMapInterface() map[string]any {
	return map[string]any{
		utils.AuthAddrCfg: l.AuthAddr,
		utils.AcctAddrCfg: l.AcctAddr,
		utils.NetworkCfg:  l.Network,
	}
}

func (cda *DAClientOpts) loadFromJSONCfg(jsnCfg DAClientOptsJson, defaultHost string) error {
	cda.Transport = utils.UDP
	if jsnCfg.Transport != nil {
		cda.Transport = *jsnCfg.Transport
	}
	cda.Host = defaultHost
	if jsnCfg.Host != nil {
		cda.Host = *jsnCfg.Host
	}
	cda.Port = 3799
	if jsnCfg.Port != nil {
		cda.Port = *jsnCfg.Port
	}
	if jsnCfg.Flags != nil {
		cda.Flags = utils.FlagsWithParamsFromSlice(jsnCfg.Flags)
	}
	return nil
}

func (cda *DAClientOpts) Clone() *DAClientOpts {
	cln := DAClientOpts{
		Transport: cda.Transport,
		Host:      cda.Host,
		Port:      cda.Port,
	}
	if cda.Flags != nil {
		cln.Flags = cda.Flags.Clone()
	}
	return &cln
}

func (cda *DAClientOpts) AsMapInterface() map[string]any {
	mp := map[string]any{
		utils.TransportCfg: cda.Transport,
		utils.HostCfg:      cda.Host,
		utils.PortCfg:      cda.Port,
	}
	if len(cda.Flags) != 0 {
		mp[utils.FlagsCfg] = cda.Flags.SliceFlags()
	}
	return mp
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
	if !maps.EqualFunc(v1.ClientDaAddresses, v2.ClientDaAddresses,
		func(v1, v2 DAClientOpts) bool {
			return v1.Transport == v2.Transport &&
				v1.Host == v2.Host &&
				v1.Port == v2.Port &&
				v1.Flags.Equal(v2.Flags)
		},
	) {
		clientDaAddresses := make(map[string]DAClientOptsJson, len(v2.ClientDaAddresses))
		for k, opts := range v2.ClientDaAddresses {
			jsonOpts := DAClientOptsJson{
				Transport: utils.StringPointer(opts.Transport),
				Host:      utils.StringPointer(opts.Host),
				Port:      utils.IntPointer(opts.Port),
			}
			if len(opts.Flags) != 0 {
				jsonOpts.Flags = opts.Flags.SliceFlags()
			}
			clientDaAddresses[k] = jsonOpts
		}
		d.ClientDaAddresses = clientDaAddresses
	}
	if !slices.Equal(v1.SessionSConns, v2.SessionSConns) {
		d.SessionSConns = utils.SliceStringPointer(stripInternalConns(v2.SessionSConns))
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.StatSConns = utils.SliceStringPointer(stripInternalConns(v2.StatSConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.ThresholdSConns = utils.SliceStringPointer(stripInternalConns(v2.ThresholdSConns))
	}
	cacheKey1 := v1.RequestsCacheKey.GetRule()
	cacheKey2 := v2.RequestsCacheKey.GetRule()
	if cacheKey1 != cacheKey2 {
		d.RequestsCacheKey = utils.StringPointer(cacheKey2)
	}
	if v1.DMRTemplate != v2.DMRTemplate {
		d.DMRTemplate = utils.StringPointer(v2.DMRTemplate)
	}
	if v1.CoATemplate != v2.CoATemplate {
		d.CoATemplate = utils.StringPointer(v2.CoATemplate)
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
