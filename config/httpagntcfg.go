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
	"github.com/cgrates/rpcclient"
)

// HTTPAgentCfgs the config section for HTTP Agent
type HTTPAgentCfgs []*HTTPAgentCfg

// loadHTTPAgentCfg loads the HttpAgent section of the configuration
func (hcfgs *HTTPAgentCfgs) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnHTTPAgntCfg := new([]*HttpAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, HTTPAgentJSON, jsnHTTPAgntCfg); err != nil {
		return
	}
	return hcfgs.loadFromJSONCfg(jsnHTTPAgntCfg, cfg.generalCfg.RSRSep)
}
func (hcfgs *HTTPAgentCfgs) loadFromJSONCfg(jsnHTTPAgntCfg *[]*HttpAgentJsonCfg, separator string) (err error) {
	if jsnHTTPAgntCfg == nil {
		return nil
	}
	for _, jsnCfg := range *jsnHTTPAgntCfg {
		hac := new(HTTPAgentCfg)
		var haveID bool
		if jsnCfg.Id != nil {
			for _, val := range *hcfgs {
				if val.ID == *jsnCfg.Id {
					hac = val
					haveID = true
					break
				}
			}
		}

		if err := hac.loadFromJSONCfg(jsnCfg, separator); err != nil {
			return err
		}
		if !haveID {
			*hcfgs = append(*hcfgs, hac)
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (hcfgs HTTPAgentCfgs) AsMapInterface(separator string) interface{} {
	mp := make([]map[string]interface{}, len(hcfgs))
	for i, item := range hcfgs {
		mp[i] = item.AsMapInterface(separator)
	}
	return mp
}

func (HTTPAgentCfgs) SName() string               { return HTTPAgentJSON }
func (hcfgs HTTPAgentCfgs) CloneSection() Section { return hcfgs.Clone() }

// Clone returns a deep copy of HTTPAgentCfgs
func (hcfgs HTTPAgentCfgs) Clone() *HTTPAgentCfgs {
	cln := make(HTTPAgentCfgs, len(hcfgs))
	for i, h := range hcfgs {
		cln[i] = h.Clone()
	}
	return &cln
}

// HTTPAgentCfg the config for a HTTP Agent
type HTTPAgentCfg struct {
	ID                string // identifier for the agent, so we can update it's processors
	URL               string
	SessionSConns     []string
	RequestPayload    string
	ReplyPayload      string
	RequestProcessors []*RequestProcessor
}

func (ha *HTTPAgentCfg) loadFromJSONCfg(jsnCfg *HttpAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Id != nil {
		ha.ID = *jsnCfg.Id
	}
	if jsnCfg.Url != nil {
		ha.URL = *jsnCfg.Url
	}
	if jsnCfg.Sessions_conns != nil {
		ha.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, connID := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			ha.SessionSConns[idx] = connID
			if connID == utils.MetaInternal ||
				connID == rpcclient.BiRPCInternal {
				ha.SessionSConns[idx] = utils.ConcatenatedKey(connID, utils.MetaSessionS)
			}
		}
	}
	if jsnCfg.Request_payload != nil {
		ha.RequestPayload = *jsnCfg.Request_payload
	}
	if jsnCfg.Reply_payload != nil {
		ha.ReplyPayload = *jsnCfg.Reply_payload
	}
	ha.RequestProcessors, err = appendRequestProcessors(ha.RequestProcessors, jsnCfg.Request_processors, separator)
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (ha *HTTPAgentCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.IDCfg:             ha.ID,
		utils.URLCfg:            ha.URL,
		utils.RequestPayloadCfg: ha.RequestPayload,
		utils.ReplyPayloadCfg:   ha.ReplyPayload,
	}

	if ha.SessionSConns != nil {
		sessionSConns := make([]string, len(ha.SessionSConns))
		for i, item := range ha.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			} else if item == utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS) {
				sessionSConns[i] = rpcclient.BiRPCInternal
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
	}
	requestProcessors := make([]map[string]interface{}, len(ha.RequestProcessors))
	for i, item := range ha.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.RequestProcessorsCfg] = requestProcessors
	return
}

// Clone returns a deep copy of HTTPAgentCfg
func (ha HTTPAgentCfg) Clone() (cln *HTTPAgentCfg) {
	cln = &HTTPAgentCfg{
		ID:                ha.ID,
		URL:               ha.URL,
		RequestPayload:    ha.RequestPayload,
		ReplyPayload:      ha.ReplyPayload,
		RequestProcessors: make([]*RequestProcessor, len(ha.RequestProcessors)),
	}
	if ha.SessionSConns != nil {
		cln.SessionSConns = make([]string, len(ha.SessionSConns))
		for i, con := range ha.SessionSConns {
			cln.SessionSConns[i] = con
		}
	}
	for i, req := range ha.RequestProcessors {
		cln.RequestProcessors[i] = req.Clone()
	}
	return
}

// Conecto Agent configuration section
type HttpAgentJsonCfg struct {
	Id                 *string
	Url                *string
	Sessions_conns     *[]string
	Request_payload    *string
	Reply_payload      *string
	Request_processors *[]*ReqProcessorJsnCfg
}

func diffHttpAgentJsonCfg(d *HttpAgentJsonCfg, v1, v2 *HTTPAgentCfg, separator string) *HttpAgentJsonCfg {
	if d == nil {
		d = new(HttpAgentJsonCfg)
	}
	if v1.ID != v2.ID {
		d.Id = utils.StringPointer(v2.ID)
	}
	if v1.URL != v2.URL {
		d.Url = utils.StringPointer(v2.URL)
	}
	if v1.RequestPayload != v2.RequestPayload {
		d.Request_payload = utils.StringPointer(v2.RequestPayload)
	}
	if v1.ReplyPayload != v2.ReplyPayload {
		d.Reply_payload = utils.StringPointer(v2.ReplyPayload)
	}
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}

	d.Request_processors = diffReqProcessorsJsnCfg(d.Request_processors, v1.RequestProcessors, v2.RequestProcessors, separator)
	return d
}

func equalsHTTPAgentCfgs(v1, v2 HTTPAgentCfgs) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].ID != v2[i].ID ||
			v1[i].URL != v2[i].URL ||
			!utils.SliceStringEqual(v1[i].SessionSConns, v2[i].SessionSConns) ||
			v1[i].RequestPayload != v2[i].RequestPayload ||
			v1[i].ReplyPayload != v2[i].ReplyPayload ||
			!equalsRequestProcessors(v1[i].RequestProcessors, v2[i].RequestProcessors) {
			return false
		}
	}
	return true
}

func getHttpAgentJsonCfg(d []*HttpAgentJsonCfg, id string) (*HttpAgentJsonCfg, int) {
	for i, v := range d {
		if v.Id != nil && *v.Id == id {
			return v, i
		}
	}
	return nil, -1
}

func getHTTPAgentCfg(d HTTPAgentCfgs, id string) *HTTPAgentCfg {
	for _, v := range d {
		if v.ID == id {
			return v
		}
	}
	return new(HTTPAgentCfg)
}
func diffHttpAgentsJsonCfg(d *[]*HttpAgentJsonCfg, v1, v2 HTTPAgentCfgs, separator string) *[]*HttpAgentJsonCfg {
	if d == nil {
		d = new([]*HttpAgentJsonCfg)
	}
	if !equalsHTTPAgentCfgs(v1, v2) {
		for _, val := range v2 {
			dv, i := getHttpAgentJsonCfg(*d, val.ID)
			dv = diffHttpAgentJsonCfg(dv, getHTTPAgentCfg(v1, val.ID), val, separator)
			if i == -1 {
				*d = append(*d, dv)
			} else {
				(*d)[i] = dv
			}
		}
	}
	return d
}
