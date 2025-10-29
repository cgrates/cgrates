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

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// HTTPAgentCfgs the config section for HTTP Agent
type HTTPAgentCfgs []*HTTPAgentCfg

func (hcfgs *HTTPAgentCfgs) loadFromJSONCfg(jsnHTTPAgntCfg *[]*HttpAgentJsonCfg, separator string) (err error) {
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

		if err := hac.loadFromJSONCfg(jsnCfg, separator); err != nil {
			return err
		}
		if !haveID {
			*hcfgs = append(*hcfgs, hac)
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (hcfgs HTTPAgentCfgs) AsMapInterface(separator string) (mp []map[string]any) {
	mp = make([]map[string]any, len(hcfgs))
	for i, item := range hcfgs {
		mp[i] = item.AsMapInterface(separator)
	}
	return
}

// Clone returns a deep copy of HTTPAgentCfgs
func (hcfgs HTTPAgentCfgs) Clone() (cln HTTPAgentCfgs) {
	if hcfgs == nil {
		return
	}
	cln = make(HTTPAgentCfgs, len(hcfgs))
	for i, h := range hcfgs {
		cln[i] = h.Clone()
	}
	return
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

func (ha *HTTPAgentCfg) appendHTTPAgntProcCfgs(hps *[]*ReqProcessorJsnCfg, separator string) (err error) {
	if hps == nil {
		return
	}
	for _, reqProcJsn := range *hps {
		rp := new(RequestProcessor)
		var haveID bool
		if reqProcJsn.ID != nil {
			for _, rpSet := range ha.RequestProcessors {
				if rpSet.ID == *reqProcJsn.ID {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
		}
		if err = rp.loadFromJSONCfg(reqProcJsn, separator); err != nil {
			return
		}
		if !haveID {
			ha.RequestProcessors = append(ha.RequestProcessors, rp)
		}
	}
	return
}

func (ha *HTTPAgentCfg) loadFromJSONCfg(jsnCfg *HttpAgentJsonCfg, separator string) (err error) {
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
		ha.SessionSConns = make([]string, len(*jsnCfg.SessionSConns))
		for idx, connID := range *jsnCfg.SessionSConns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			ha.SessionSConns[idx] = connID
			if connID == utils.MetaInternal ||
				connID == rpcclient.BiRPCInternal {
				ha.SessionSConns[idx] = utils.ConcatenatedKey(connID, utils.MetaSessionS)
			}
		}
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
	if err = ha.appendHTTPAgntProcCfgs(jsnCfg.RequestProcessors, separator); err != nil {
		return err
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (ha *HTTPAgentCfg) AsMapInterface(separator string) map[string]any {
	requestProcessors := make([]map[string]any, len(ha.RequestProcessors))
	for i, item := range ha.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	m := map[string]any{
		utils.IDCfg:                ha.ID,
		utils.URLCfg:               ha.URL,
		utils.RequestPayloadCfg:    ha.RequestPayload,
		utils.ReplyPayloadCfg:      ha.ReplyPayload,
		utils.StatSConnsCfg:        stripInternalConns(ha.StatSConns),
		utils.ThresholdSConnsCfg:   stripInternalConns(ha.ThresholdSConns),
		utils.RequestProcessorsCfg: requestProcessors,
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
		m[utils.SessionSConnsCfg] = sessionSConns
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
