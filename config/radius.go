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

	"maps"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// RadiusAgentCfg the config section that describes the Radius Agent
type RadiusAgentCfg struct {
	Enabled            bool
	ListenNet          string // udp or tcp
	ListenAuth         string
	ListenAcct         string
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

func (ra *RadiusAgentCfg) loadFromJSONCfg(jsnCfg *RadiusAgentJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		ra.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.ListenNet != nil {
		ra.ListenNet = *jsnCfg.ListenNet
	}
	if jsnCfg.ListenAuth != nil {
		ra.ListenAuth = *jsnCfg.ListenAuth
	}
	if jsnCfg.ListenAcct != nil {
		ra.ListenAcct = *jsnCfg.ListenAcct
	}
	maps.Copy(ra.ClientSecrets, jsnCfg.ClientSecrets)
	maps.Copy(ra.ClientDictionaries, jsnCfg.ClientDictionaries)
	if jsnCfg.SessionSConns != nil {
		ra.SessionSConns = tagInternalConns(*jsnCfg.SessionSConns, utils.MetaSessionS)
	}
	if jsnCfg.StatSConns != nil {
		ra.StatSConns = tagInternalConns(*jsnCfg.StatSConns, utils.MetaStats)
	}
	if jsnCfg.ThresholdSConns != nil {
		ra.ThresholdSConns = tagInternalConns(*jsnCfg.ThresholdSConns, utils.MetaThresholds)
	}
	ra.RequestProcessors, err = appendRequestProcessors(ra.RequestProcessors, jsnCfg.RequestProcessors)
	return
}

// AsMapInterface returns the config as a map[string]any
func (ra RadiusAgentCfg) AsMapInterface() any {
	requestProcessors := make([]map[string]any, len(ra.RequestProcessors))
	for i, item := range ra.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	mp := map[string]any{
		utils.EnabledCfg:            ra.Enabled,
		utils.ListenNetCfg:          ra.ListenNet,
		utils.ListenAuthCfg:         ra.ListenAuth,
		utils.ListenAcctCfg:         ra.ListenAcct,
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
		ListenNet:       ra.ListenNet,
		ListenAuth:      ra.ListenAuth,
		ListenAcct:      ra.ListenAcct,
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

// Radius Agent configuration section
type RadiusAgentJsonCfg struct {
	Enabled            *bool                  `json:"enabled"`
	ListenNet          *string                `json:"listen_net"`
	ListenAuth         *string                `json:"listen_auth"`
	ListenAcct         *string                `json:"listen_acct"`
	ClientSecrets      map[string]string      `json:"client_secrets"`
	ClientDictionaries map[string][]string    `json:"client_dictionaries"`
	SessionSConns      *[]string              `json:"sessions_conns"`
	StatSConns         *[]string              `json:"stats_conns"`
	ThresholdSConns    *[]string              `json:"thresholds_conns"`
	RequestProcessors  *[]*ReqProcessorJsnCfg `json:"request_processors"`
}

func diffRadiusAgentJsonCfg(d *RadiusAgentJsonCfg, v1, v2 *RadiusAgentCfg) *RadiusAgentJsonCfg {
	if d == nil {
		d = new(RadiusAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.ListenNet != v2.ListenNet {
		d.ListenNet = utils.StringPointer(v2.ListenNet)
	}
	if v1.ListenAuth != v2.ListenAuth {
		d.ListenAuth = utils.StringPointer(v2.ListenAuth)
	}
	if v1.ListenAcct != v2.ListenAcct {
		d.ListenAcct = utils.StringPointer(v2.ListenAcct)
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
