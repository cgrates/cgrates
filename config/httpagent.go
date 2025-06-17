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

// HTTPAgentCfgs the config section for HTTP Agent
type HTTPAgentCfgs []*HTTPAgentCfg

// loadHTTPAgentCfg loads the HttpAgent section of the configuration
func (hcfgs *HTTPAgentCfgs) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnHTTPAgntCfg := new([]*HttpAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, HTTPAgentJSON, jsnHTTPAgntCfg); err != nil {
		return
	}
	return hcfgs.loadFromJSONCfg(jsnHTTPAgntCfg)
}
func (hcfgs *HTTPAgentCfgs) loadFromJSONCfg(jsnHTTPAgntCfg *[]*HttpAgentJsonCfg) (err error) {
	if jsnHTTPAgntCfg == nil {
		return nil
	}
	for _, jsnCfg := range *jsnHTTPAgntCfg {
		hac := new(HTTPAgentCfg)
		var haveID bool
		if jsnCfg.ID != nil {
			for _, val := range *hcfgs {
				if val.ID == *jsnCfg.ID {
					hac = val
					haveID = true
					break
				}
			}
		}

		if err := hac.loadFromJSONCfg(jsnCfg); err != nil {
			return err
		}
		if !haveID {
			*hcfgs = append(*hcfgs, hac)
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (hcfgs HTTPAgentCfgs) AsMapInterface() any {
	mp := make([]map[string]any, len(hcfgs))
	for i, item := range hcfgs {
		mp[i] = item.AsMapInterface()
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
	StatSConns        []string
	ThresholdSConns   []string
	RequestPayload    string
	ReplyPayload      string
	RequestProcessors []*RequestProcessor
}

func (ha *HTTPAgentCfg) loadFromJSONCfg(jsnCfg *HttpAgentJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.ID != nil {
		ha.ID = *jsnCfg.ID
	}
	if jsnCfg.URL != nil {
		ha.URL = *jsnCfg.URL
	}
	if jsnCfg.SessionSConns != nil {
		ha.SessionSConns = tagInternalConns(*jsnCfg.SessionSConns, utils.MetaSessionS)
	}
	if jsnCfg.StatSConns != nil {
		ha.StatSConns = tagInternalConns(*jsnCfg.StatSConns, utils.MetaStats)
	}
	if jsnCfg.ThresholdSConns != nil {
		ha.ThresholdSConns = tagInternalConns(*jsnCfg.ThresholdSConns, utils.MetaThresholds)
	}

	if jsnCfg.RequestPayload != nil {
		ha.RequestPayload = *jsnCfg.RequestPayload
	}
	if jsnCfg.ReplyPayload != nil {
		ha.ReplyPayload = *jsnCfg.ReplyPayload
	}
	ha.RequestProcessors, err = appendRequestProcessors(ha.RequestProcessors, jsnCfg.RequestProcessors)
	return
}

// AsMapInterface returns the config as a map[string]any
func (ha *HTTPAgentCfg) AsMapInterface() map[string]any {
	requestProcessors := make([]map[string]any, len(ha.RequestProcessors))
	for i, item := range ha.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	m := map[string]any{
		utils.IDCfg:                ha.ID,
		utils.URLCfg:               ha.URL,
		utils.SessionSConnsCfg:     stripInternalConns(ha.SessionSConns),
		utils.StatSConnsCfg:        stripInternalConns(ha.StatSConns),
		utils.ThresholdSConnsCfg:   stripInternalConns(ha.ThresholdSConns),
		utils.RequestPayloadCfg:    ha.RequestPayload,
		utils.ReplyPayloadCfg:      ha.ReplyPayload,
		utils.RequestProcessorsCfg: requestProcessors,
	}
	return m
}

// Clone returns a deep copy of HTTPAgentCfg
func (ha HTTPAgentCfg) Clone() *HTTPAgentCfg {
	clone := &HTTPAgentCfg{
		ID:                ha.ID,
		URL:               ha.URL,
		SessionSConns:     slices.Clone(ha.SessionSConns),
		StatSConns:        slices.Clone(ha.StatSConns),
		ThresholdSConns:   slices.Clone(ha.ThresholdSConns),
		RequestPayload:    ha.RequestPayload,
		ReplyPayload:      ha.ReplyPayload,
		RequestProcessors: make([]*RequestProcessor, len(ha.RequestProcessors)),
	}
	for i, req := range ha.RequestProcessors {
		clone.RequestProcessors[i] = req.Clone()
	}
	return clone
}

// Conecto Agent configuration section
type HttpAgentJsonCfg struct {
	ID                *string                `json:"id"`
	URL               *string                `json:"url"`
	SessionSConns     *[]string              `json:"sessions_conns"`
	StatSConns        *[]string              `json:"stats_conns"`
	ThresholdSConns   *[]string              `json:"thresholds_conns"`
	RequestPayload    *string                `json:"request_payload"`
	ReplyPayload      *string                `json:"reply_payload"`
	RequestProcessors *[]*ReqProcessorJsnCfg `json:"request_processors"`
}

func diffHttpAgentJsonCfg(d *HttpAgentJsonCfg, v1, v2 *HTTPAgentCfg) *HttpAgentJsonCfg {
	if d == nil {
		d = new(HttpAgentJsonCfg)
	}
	if v1.ID != v2.ID {
		d.ID = utils.StringPointer(v2.ID)
	}
	if v1.URL != v2.URL {
		d.URL = utils.StringPointer(v2.URL)
	}
	if v1.RequestPayload != v2.RequestPayload {
		d.RequestPayload = utils.StringPointer(v2.RequestPayload)
	}
	if v1.ReplyPayload != v2.ReplyPayload {
		d.ReplyPayload = utils.StringPointer(v2.ReplyPayload)
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
	d.RequestProcessors = diffReqProcessorsJsnCfg(d.RequestProcessors, v1.RequestProcessors, v2.RequestProcessors)
	return d
}

func equalsHTTPAgentCfgs(v1, v2 HTTPAgentCfgs) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].ID != v2[i].ID ||
			v1[i].URL != v2[i].URL ||
			!slices.Equal(v1[i].SessionSConns, v2[i].SessionSConns) ||
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
		if v.ID != nil && *v.ID == id {
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
func diffHttpAgentsJsonCfg(d *[]*HttpAgentJsonCfg, v1, v2 HTTPAgentCfgs) *[]*HttpAgentJsonCfg {
	if d == nil {
		d = new([]*HttpAgentJsonCfg)
	}
	if !equalsHTTPAgentCfgs(v1, v2) {
		for _, val := range v2 {
			dv, i := getHttpAgentJsonCfg(*d, val.ID)
			dv = diffHttpAgentJsonCfg(dv, getHTTPAgentCfg(v1, val.ID), val)
			if i == -1 {
				*d = append(*d, dv)
			} else {
				(*d)[i] = dv
			}
		}
	}
	return d
}
