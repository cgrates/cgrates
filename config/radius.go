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
	"fmt"
	"slices"

	"maps"

	"github.com/cgrates/cgrates/utils"
)

type RadiusListener struct {
	AuthAddr string
	AcctAddr string
	Network  string // udp or tcp
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
	RequestsCacheKey   RSRParsers
	DMRTemplate        string
	CoATemplate        string
	RequestProcessors  []*RequestProcessor
}

func (ra *RadiusAgentCfg) loadFromJSONCfg(jsnCfg *RadiusAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		ra.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listeners != nil {
		ra.Listeners = make([]RadiusListener, 0, len(*jsnCfg.Listeners))
		for _, listnr := range *jsnCfg.Listeners {
			var ls RadiusListener
			if listnr.Auth_Address != nil {
				ls.AuthAddr = *listnr.Auth_Address
			}
			if listnr.Acct_Address != nil {
				ls.AcctAddr = *listnr.Acct_Address
			}
			if listnr.Network != nil {
				ls.Network = *listnr.Network
			}
			ra.Listeners = append(ra.Listeners, ls)
		}
	}
	if jsnCfg.ClientSecrets != nil {
		if ra.ClientSecrets == nil {
			ra.ClientSecrets = make(map[string]string)
		}
		maps.Copy(ra.ClientSecrets, *jsnCfg.ClientSecrets)
	}
	if jsnCfg.ClientDictionaries != nil {
		if ra.ClientDictionaries == nil {
			ra.ClientDictionaries = make(map[string][]string)
		}
		maps.Copy(ra.ClientDictionaries, *jsnCfg.ClientDictionaries)
	}
	if len(jsnCfg.ClientDaAddresses) != 0 {
		if ra.ClientDaAddresses == nil {
			ra.ClientDaAddresses = make(map[string]DAClientOpts)
		}
		ra.ClientDaAddresses = make(map[string]DAClientOpts, len(jsnCfg.ClientDaAddresses))
		for hostKey, clientOpts := range jsnCfg.ClientDaAddresses {
			cfg := DAClientOpts{}
			cfg.loadFromJSONCfg(clientOpts, hostKey)
			ra.ClientDaAddresses[hostKey] = cfg
		}
	}
	if jsnCfg.SessionSConns != nil {
		ra.SessionSConns = make([]string, len(*jsnCfg.SessionSConns))
		for idx, attrConn := range *jsnCfg.SessionSConns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			ra.SessionSConns[idx] = attrConn
			if attrConn == utils.MetaInternal {
				ra.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			}
		}
	}
	if jsnCfg.StatSConns != nil {
		ra.StatSConns = tagInternalConns(*jsnCfg.StatSConns, utils.MetaStats)
	}
	if jsnCfg.ThresholdSConns != nil {
		ra.ThresholdSConns = tagInternalConns(*jsnCfg.ThresholdSConns, utils.MetaThresholds)
	}
	if jsnCfg.RequestsCacheKey != nil {
		ra.RequestsCacheKey, err = NewRSRParsers(*jsnCfg.RequestsCacheKey, separator)
		if err != nil {
			return fmt.Errorf(
				"failed to initialize RSRParsers based %s value: %w",
				utils.RequestsCacheKeyCfg, err,
			)
		}
	}
	if jsnCfg.DMRTemplate != nil {
		ra.DMRTemplate = *jsnCfg.DMRTemplate
	}
	if jsnCfg.CoATemplate != nil {
		ra.CoATemplate = *jsnCfg.CoATemplate
	}
	if jsnCfg.RequestProcessors != nil {
		for _, reqProcJsn := range *jsnCfg.RequestProcessors {
			rp := new(RequestProcessor)
			var haveID bool
			for _, rpSet := range ra.RequestProcessors {
				if reqProcJsn.ID != nil && rpSet.ID == *reqProcJsn.ID {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
			if err = rp.loadFromJSONCfg(reqProcJsn, separator); err != nil {
				return
			}
			if !haveID {
				ra.RequestProcessors = append(ra.RequestProcessors, rp)
			}
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (lstn *RadiusListener) AsMapInterface(separator string) map[string]any {
	return map[string]any{
		utils.AuthAddrCfg: lstn.AuthAddr,
		utils.AcctAddrCfg: lstn.AcctAddr,
		utils.NetworkCfg:  lstn.Network,
	}

}

// AsMapInterface returns the config as a map[string]any
func (ra *RadiusAgentCfg) AsMapInterface(separator string) map[string]any {
	listeners := make([]map[string]any, len(ra.Listeners))
	for i, item := range ra.Listeners {
		listeners[i] = item.AsMapInterface(separator)
	}
	requestProcessors := make([]map[string]any, len(ra.RequestProcessors))
	for i, item := range ra.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	m := map[string]any{
		utils.EnabledCfg:            ra.Enabled,
		utils.ListenersCfg:          listeners,
		utils.ClientSecretsCfg:      maps.Clone(ra.ClientSecrets),
		utils.ClientDictionariesCfg: maps.Clone(ra.ClientDictionaries),
		utils.RequestsCacheKeyCfg:   ra.RequestsCacheKey.GetRule(separator),
		utils.DMRTemplateCfg:        ra.DMRTemplate,
		utils.CoATemplateCfg:        ra.CoATemplate,
		utils.StatSConnsCfg:         stripInternalConns(ra.StatSConns),
		utils.ThresholdSConnsCfg:    stripInternalConns(ra.ThresholdSConns),
		utils.RequestProcessorsCfg:  requestProcessors,
	}
	if ra.SessionSConns != nil {
		sessionSConns := make([]string, len(ra.SessionSConns))
		for i, item := range ra.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			}
		}
		m[utils.SessionSConnsCfg] = sessionSConns
	}
	if len(ra.ClientDaAddresses) != 0 {
		clientDaAddresses := make(map[string]any)
		for k, v := range ra.ClientDaAddresses {
			clientDaAddresses[k] = v.AsMapInterface()
		}
		m[utils.ClientDaAddressesCfg] = clientDaAddresses
	}
	return m
}

// Clone returns a deep copy of RadiusAgentCfg
func (ra RadiusAgentCfg) Clone() *RadiusAgentCfg {
	clone := &RadiusAgentCfg{
		Enabled:       ra.Enabled,
		Listeners:     slices.Clone(ra.Listeners),
		ClientSecrets: maps.Clone(ra.ClientSecrets),

		// NOTE: shallow clone and value is a slice
		ClientDictionaries: maps.Clone(ra.ClientDictionaries),

		SessionSConns:    slices.Clone(ra.SessionSConns),
		StatSConns:       slices.Clone(ra.StatSConns),
		ThresholdSConns:  slices.Clone(ra.ThresholdSConns),
		RequestsCacheKey: ra.RequestsCacheKey.Clone(),
		DMRTemplate:      ra.DMRTemplate,
		CoATemplate:      ra.CoATemplate,
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

type DAClientOpts struct {
	Transport string                // transport protocol for Dynamic Authorization requests <UDP|TCP>.
	Host      string                // alternative host for DA requests
	Port      int                   // port for Dynamic Authorization requests
	Flags     utils.FlagsWithParams // flags (only *log for now)
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
