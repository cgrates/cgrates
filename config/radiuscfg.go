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

// RadiusAgentCfg the config section that describes the Radius Agent
type RadiusAgentCfg struct {
	Enabled            bool
	ListenNet          string // udp or tcp
	ListenAuth         string
	ListenAcct         string
	ClientSecrets      map[string]string
	ClientDictionaries map[string]string
	SessionSConns      []string
	RequestProcessors  []*RequestProcessor
}

func (ra *RadiusAgentCfg) loadFromJSONCfg(jsnCfg *RadiusAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		ra.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_net != nil {
		ra.ListenNet = *jsnCfg.Listen_net
	}
	if jsnCfg.Listen_auth != nil {
		ra.ListenAuth = *jsnCfg.Listen_auth
	}
	if jsnCfg.Listen_acct != nil {
		ra.ListenAcct = *jsnCfg.Listen_acct
	}
	if jsnCfg.Client_secrets != nil {
		for k, v := range jsnCfg.Client_secrets {
			ra.ClientSecrets[k] = v
		}
	}
	if jsnCfg.Client_dictionaries != nil {
		for k, v := range jsnCfg.Client_dictionaries {
			ra.ClientDictionaries[k] = v
		}
	}
	if jsnCfg.Sessions_conns != nil {
		ra.SessionSConns = updateBiRPCInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	ra.RequestProcessors, err = appendRequestProcessors(ra.RequestProcessors, jsnCfg.Request_processors, separator)
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (ra *RadiusAgentCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:    ra.Enabled,
		utils.ListenNetCfg:  ra.ListenNet,
		utils.ListenAuthCfg: ra.ListenAuth,
		utils.ListenAcctCfg: ra.ListenAcct,
	}

	requestProcessors := make([]map[string]interface{}, len(ra.RequestProcessors))
	for i, item := range ra.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.RequestProcessorsCfg] = requestProcessors

	if ra.SessionSConns != nil {
		initialMP[utils.SessionSConnsCfg] = getBiRPCInternalJSONConns(ra.SessionSConns)
	}
	clientSecrets := make(map[string]string)
	for k, v := range ra.ClientSecrets {
		clientSecrets[k] = v
	}
	initialMP[utils.ClientSecretsCfg] = clientSecrets
	clientDictionaries := make(map[string]string)
	for k, v := range ra.ClientDictionaries {
		clientDictionaries[k] = v
	}
	initialMP[utils.ClientDictionariesCfg] = clientDictionaries
	return
}

// Clone returns a deep copy of RadiusAgentCfg
func (ra RadiusAgentCfg) Clone() (cln *RadiusAgentCfg) {
	cln = &RadiusAgentCfg{
		Enabled:            ra.Enabled,
		ListenNet:          ra.ListenNet,
		ListenAuth:         ra.ListenAuth,
		ListenAcct:         ra.ListenAcct,
		ClientSecrets:      make(map[string]string),
		ClientDictionaries: make(map[string]string),
	}
	if ra.SessionSConns != nil {
		cln.SessionSConns = utils.CloneStringSlice(ra.SessionSConns)
	}
	for k, v := range ra.ClientSecrets {
		cln.ClientSecrets[k] = v
	}
	for k, v := range ra.ClientDictionaries {
		cln.ClientDictionaries[k] = v
	}
	if ra.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(ra.RequestProcessors))
		for i, req := range ra.RequestProcessors {
			cln.RequestProcessors[i] = req.Clone()
		}
	}
	return
}

// Radius Agent configuration section
type RadiusAgentJsonCfg struct {
	Enabled             *bool
	Listen_net          *string
	Listen_auth         *string
	Listen_acct         *string
	Client_secrets      map[string]string
	Client_dictionaries map[string]string
	Sessions_conns      *[]string
	Request_processors  *[]*ReqProcessorJsnCfg
}

func diffRadiusAgentJsonCfg(d *RadiusAgentJsonCfg, v1, v2 *RadiusAgentCfg, separator string) *RadiusAgentJsonCfg {
	if d == nil {
		d = new(RadiusAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.ListenNet != v2.ListenNet {
		d.Listen_net = utils.StringPointer(v2.ListenNet)
	}
	if v1.ListenAuth != v2.ListenAuth {
		d.Listen_auth = utils.StringPointer(v2.ListenAuth)
	}
	if v1.ListenAcct != v2.ListenAcct {
		d.Listen_acct = utils.StringPointer(v2.ListenAcct)
	}
	d.Client_secrets = diffMapString(d.Client_secrets, v1.ClientSecrets, v2.ClientSecrets)
	d.Client_dictionaries = diffMapString(d.Client_dictionaries, v1.ClientDictionaries, v2.ClientDictionaries)
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}
	d.Request_processors = diffReqProcessorsJsnCfg(d.Request_processors, v1.RequestProcessors, v2.RequestProcessors, separator)
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
