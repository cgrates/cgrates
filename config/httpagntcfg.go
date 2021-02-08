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

// HTTPAgentCfgs the config section for HTTP Agent
type HTTPAgentCfgs []*HTTPAgentCfg

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
func (hcfgs HTTPAgentCfgs) AsMapInterface(separator string) (mp []map[string]interface{}) {
	mp = make([]map[string]interface{}, len(hcfgs))
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
			if connID == utils.MetaInternal {
				ha.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				ha.SessionSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Request_payload != nil {
		ha.RequestPayload = *jsnCfg.Request_payload
	}
	if jsnCfg.Reply_payload != nil {
		ha.ReplyPayload = *jsnCfg.Reply_payload
	}
	if err = ha.appendHTTPAgntProcCfgs(jsnCfg.Request_processors, separator); err != nil {
		return err
	}
	return nil
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
