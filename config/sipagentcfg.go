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
	"time"

	"github.com/cgrates/cgrates/utils"
)

// SIPAgentCfg the config for the SIPAgent
type SIPAgentCfg struct {
	Enabled             bool
	Listen              string
	ListenNet           string // udp or tcp
	SessionSConns       []string
	Timezone            string
	RetransmissionTimer time.Duration // timeout replies if not reaching back
	RequestProcessors   []*RequestProcessor
}

func (da *SIPAgentCfg) loadFromJSONCfg(jsnCfg *SIPAgentJsonCfg, sep string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		da.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_net != nil {
		da.ListenNet = *jsnCfg.Listen_net
	}
	if jsnCfg.Listen != nil {
		da.Listen = *jsnCfg.Listen
	}
	if jsnCfg.Timezone != nil {
		da.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Sessions_conns != nil {
		da.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, connID := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				da.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				da.SessionSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Retransmission_timer != nil {
		if da.RetransmissionTimer, err = utils.ParseDurationWithNanosecs(*jsnCfg.Retransmission_timer); err != nil {
			return err
		}
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(RequestProcessor)
			var haveID bool
			for _, rpSet := range da.RequestProcessors {
				if reqProcJsn.ID != nil && rpSet.ID == *reqProcJsn.ID {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
			if err = rp.loadFromJSONCfg(reqProcJsn, sep); err != nil {
				return
			}
			if !haveID {
				da.RequestProcessors = append(da.RequestProcessors, rp)
			}
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (da *SIPAgentCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:             da.Enabled,
		utils.ListenCfg:              da.Listen,
		utils.ListenNetCfg:           da.ListenNet,
		utils.TimezoneCfg:            da.Timezone,
		utils.RetransmissionTimerCfg: da.RetransmissionTimer,
	}

	requestProcessors := make([]map[string]interface{}, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.RequestProcessorsCfg] = requestProcessors

	if da.SessionSConns != nil {
		sessionSConns := make([]string, len(da.SessionSConns))
		for i, item := range da.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
	}
	return
}
