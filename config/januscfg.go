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
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// JanusConn represents one connection to Janus server
type JanusConn struct {
	Address       string // Address to reach Janus
	Type          string // Connection type
	AdminAddress  string
	AdminPassword string
}

// JanusAgentCfg the config for an Janus Agent
type JanusAgentCfg struct {
	Enabled           bool
	URL               string
	SessionSConns     []string
	JanusConns        []*JanusConn // connections towards Janus
	RequestProcessors []*RequestProcessor
}

func (jacfg *JanusAgentCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnJACfg := new(JanusAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, JanusAgentJSON, jsnJACfg); err != nil {
		return
	}
	return jacfg.loadFromJSONCfg(jsnJACfg)
}

func (jc *JanusConn) loadFromJSONCfg(jsnCfg *JanusConnJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Address != nil {
		jc.Address = *jsnCfg.Address
	}
	if jsnCfg.Type != nil {
		jc.Type = *jsnCfg.Type
	}
	if jsnCfg.AdminAddress != nil {
		jc.AdminAddress = *jsnCfg.AdminAddress
	}
	if jsnCfg.AdminPassword != nil {
		jc.AdminPassword = *jsnCfg.AdminPassword
	}
	return
}

func (jc *JanusConn) AsMapInterface() map[string]any {
	mp := map[string]any{
		utils.AddressCfg:       jc.Address,
		utils.TypeCfg:          jc.Type,
		utils.AdminAddressCfg:  jc.AdminAddress,
		utils.AdminPasswordCfg: jc.AdminPassword,
	}
	return mp
}

func (jc *JanusConn) Clone() (cln *JanusConn) {
	cln = &JanusConn{
		Address:       jc.Address,
		Type:          jc.Type,
		AdminAddress:  jc.AdminAddress,
		AdminPassword: jc.AdminPassword,
	}
	return
}

func (jaCfg *JanusAgentCfg) loadFromJSONCfg(jsnCfg *JanusAgentJsonCfg) (err error) {
	if jaCfg == nil {
		return
	}
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		jaCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Url != nil {
		jaCfg.URL = *jsnCfg.Url
	}
	if jsnCfg.Sessions_conns != nil {
		jaCfg.SessionSConns = tagInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	if jsnCfg.Janus_conns != nil {
		jaCfg.JanusConns = make([]*JanusConn, len(*jsnCfg.Janus_conns))
		for idx, janConnJsn := range *jsnCfg.Janus_conns {
			jc := new(JanusConn)
			if err = jc.loadFromJSONCfg(janConnJsn); err != nil {
				return
			}
			jaCfg.JanusConns[idx] = jc
		}
	}
	jaCfg.RequestProcessors, err = appendRequestProcessors(jaCfg.RequestProcessors, jsnCfg.Request_processors)
	return
}

// AsMapInterface returns the config as a map[string]any
func (jaCfg JanusAgentCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.EnabledCfg: jaCfg.Enabled,
		utils.URLCfg:     jaCfg.URL,
	}
	requestProcessors := make([]map[string]any, len(jaCfg.RequestProcessors))
	for i, item := range jaCfg.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	mp[utils.RequestProcessorsCfg] = requestProcessors
	if jaCfg.SessionSConns != nil {
		mp[utils.SessionSConnsCfg] = stripInternalConns(jaCfg.SessionSConns)
	}
	janConns := make([]map[string]any, len(jaCfg.JanusConns))
	for i, jc := range jaCfg.JanusConns {
		janConns[i] = jc.AsMapInterface()
	}
	mp[utils.JanusConnsCfg] = janConns
	return mp
}

func (JanusAgentCfg) SName() string               { return JanusAgentJSON }
func (jacfg JanusAgentCfg) CloneSection() Section { return jacfg.Clone() }

func (jaCfg *JanusAgentCfg) Clone() *JanusAgentCfg {
	cln := &JanusAgentCfg{
		Enabled: jaCfg.Enabled,
		URL:     jaCfg.URL,
	}
	if jaCfg.SessionSConns != nil {
		cln.SessionSConns = slices.Clone(jaCfg.SessionSConns)
	}
	if jaCfg.JanusConns != nil {
		cln.JanusConns = make([]*JanusConn, len(jaCfg.JanusConns))
		for i, jc := range jaCfg.JanusConns {
			cln.JanusConns[i] = jc.Clone()
		}
	}
	if jaCfg.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(jaCfg.RequestProcessors))
		for i, rp := range jaCfg.RequestProcessors {
			cln.RequestProcessors[i] = rp.Clone()
		}
	}
	return cln
}

type JanusAgentJsonCfg struct {
	Enabled            *bool                  `json:"enabled"`
	Url                *string                `json:"url"`
	Sessions_conns     *[]string              `json:"sessions_conns"`
	Janus_conns        *[]*JanusConnJsonCfg   `json:"janus_conns"`
	Request_processors *[]*ReqProcessorJsnCfg `json:"request_processors"`
}

type JanusConnJsonCfg struct {
	Address       *string `json:"address"`
	Type          *string `json:"type"`
	AdminAddress  *string `json:"admin_address"`
	AdminPassword *string `json:"admin_password"`
}

func diffJanusConnJsonCfg(d *JanusConnJsonCfg, v1, v2 *JanusConn) *JanusConnJsonCfg {
	if d == nil {
		d = new(JanusConnJsonCfg)
	}
	if v1.Address != v2.Address {
		d.Address = utils.StringPointer(v2.Address)
	}
	if v1.Type != v2.Type {
		d.Type = utils.StringPointer(v2.Type)
	}
	if v1.AdminAddress != v2.AdminAddress {
		d.AdminAddress = utils.StringPointer(v2.AdminAddress)
	}
	if v1.AdminPassword != v2.AdminPassword {
		d.AdminPassword = utils.StringPointer(v2.AdminPassword)
	}
	return d
}

func getJanusConnJsnCfg(d []*JanusConnJsonCfg, address string) (*JanusConnJsonCfg, int) {
	for i, v := range d {
		if v.Address != nil && *v.Address == address {
			return v, i
		}
	}
	return nil, -1
}

func getJanusConn(d []*JanusConn, address string) *JanusConn {
	for _, v := range d {
		if v.Address == address {
			return v
		}
	}
	return new(JanusConn)
}

func diffJanusConnsJsonCfg(d *[]*JanusConnJsonCfg, v1, v2 []*JanusConn) *[]*JanusConnJsonCfg {
	if d == nil || *d == nil {
		d = &[]*JanusConnJsonCfg{}
	}
	for _, val := range v2 {
		dv, i := getJanusConnJsnCfg(*d, val.Address)
		dv = diffJanusConnJsonCfg(dv, getJanusConn(v1, val.Address), val)
		if i == -1 {
			*d = append(*d, dv)
		} else {
			(*d)[i] = dv
		}
	}
	return d
}

func diffJanusAgentSJsonCfg(d *JanusAgentJsonCfg, v1, v2 *JanusAgentCfg) *JanusAgentJsonCfg {
	if d == nil {
		d = new(JanusAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.URL != v2.URL {
		d.Url = utils.StringPointer(v2.URL)
	}
	if !slices.Equal(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(stripInternalConns(v2.SessionSConns))
	}
	d.Request_processors = diffReqProcessorsJsnCfg(d.Request_processors, v1.RequestProcessors, v2.RequestProcessors)
	d.Janus_conns = diffJanusConnsJsonCfg(d.Janus_conns, v1.JanusConns, v2.JanusConns)
	return d
}
