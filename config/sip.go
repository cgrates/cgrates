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
	"time"

	"github.com/cgrates/cgrates/utils"
)

// SIPAgentCfg the config for the SIPAgent
type SIPAgentCfg struct {
	Enabled             bool
	Listen              string
	ListenNet           string // udp or tcp
	SessionSConns       []string
	StatSConns          []string
	ThresholdSConns     []string
	Timezone            string
	RetransmissionTimer time.Duration // timeout replies if not reaching back
	RequestProcessors   []*RequestProcessor
}

func (sa *SIPAgentCfg) loadFromJSONCfg(jsnCfg *SIPAgentJsonCfg, sep string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		sa.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.ListenNet != nil {
		sa.ListenNet = *jsnCfg.ListenNet
	}
	if jsnCfg.Listen != nil {
		sa.Listen = *jsnCfg.Listen
	}
	if jsnCfg.Timezone != nil {
		sa.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.SessionSConns != nil {
		sa.SessionSConns = make([]string, len(*jsnCfg.SessionSConns))
		for idx, connID := range *jsnCfg.SessionSConns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			sa.SessionSConns[idx] = connID
			if connID == utils.MetaInternal {
				sa.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			}
		}
	}
	if jsnCfg.StatSConns != nil {
		sa.StatSConns = tagInternalConns(*jsnCfg.StatSConns, utils.MetaStats)
	}
	if jsnCfg.ThresholdSConns != nil {
		sa.ThresholdSConns = tagInternalConns(*jsnCfg.ThresholdSConns, utils.MetaThresholds)
	}
	if jsnCfg.RetransmissionTimer != nil {
		if sa.RetransmissionTimer, err = utils.ParseDurationWithNanosecs(*jsnCfg.RetransmissionTimer); err != nil {
			return err
		}
	}
	if jsnCfg.RequestProcessors != nil {
		for _, reqProcJsn := range *jsnCfg.RequestProcessors {
			rp := new(RequestProcessor)
			var haveID bool
			for _, rpSet := range sa.RequestProcessors {
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
				sa.RequestProcessors = append(sa.RequestProcessors, rp)
			}
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (sa *SIPAgentCfg) AsMapInterface(separator string) map[string]any {
	requestProcessors := make([]map[string]any, len(sa.RequestProcessors))
	for i, item := range sa.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	m := map[string]any{
		utils.EnabledCfg:             sa.Enabled,
		utils.ListenCfg:              sa.Listen,
		utils.ListenNetCfg:           sa.ListenNet,
		utils.StatSConnsCfg:          stripInternalConns(sa.StatSConns),
		utils.ThresholdSConnsCfg:     stripInternalConns(sa.ThresholdSConns),
		utils.TimezoneCfg:            sa.Timezone,
		utils.RetransmissionTimerCfg: sa.RetransmissionTimer,
		utils.RequestProcessorsCfg:   requestProcessors,
	}
	if sa.SessionSConns != nil {
		sessionSConns := make([]string, len(sa.SessionSConns))
		for i, item := range sa.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			}
		}
		m[utils.SessionSConnsCfg] = sessionSConns
	}
	return m
}

// Clone returns a deep copy of SIPAgentCfg
func (sa SIPAgentCfg) Clone() *SIPAgentCfg {
	clone := &SIPAgentCfg{
		Enabled:             sa.Enabled,
		Listen:              sa.Listen,
		ListenNet:           sa.ListenNet,
		SessionSConns:       slices.Clone(sa.SessionSConns),
		StatSConns:          slices.Clone(sa.StatSConns),
		ThresholdSConns:     slices.Clone(sa.ThresholdSConns),
		Timezone:            sa.Timezone,
		RetransmissionTimer: sa.RetransmissionTimer,
	}
	if sa.RequestProcessors != nil {
		clone.RequestProcessors = make([]*RequestProcessor, len(sa.RequestProcessors))
		for i, rp := range sa.RequestProcessors {
			clone.RequestProcessors[i] = rp.Clone()
		}
	}
	return clone
}
