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

	"github.com/cgrates/birpc/context"
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

// loadSIPAgentCfg loads the sip_agent section of the configuration
func (sa *SIPAgentCfg) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnSIPAgentCfg := new(SIPAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, SIPAgentJSON, jsnSIPAgentCfg); err != nil {
		return
	}
	return sa.loadFromJSONCfg(jsnSIPAgentCfg, cfg.generalCfg.RSRSep)
}

func (sa *SIPAgentCfg) loadFromJSONCfg(jsnCfg *SIPAgentJsonCfg, sep string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		sa.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_net != nil {
		sa.ListenNet = *jsnCfg.Listen_net
	}
	if jsnCfg.Listen != nil {
		sa.Listen = *jsnCfg.Listen
	}
	if jsnCfg.Timezone != nil {
		sa.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Sessions_conns != nil {
		sa.SessionSConns = updateBiRPCInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	if jsnCfg.Retransmission_timer != nil {
		if sa.RetransmissionTimer, err = utils.ParseDurationWithNanosecs(*jsnCfg.Retransmission_timer); err != nil {
			return err
		}
	}
	sa.RequestProcessors, err = appendRequestProcessors(sa.RequestProcessors, jsnCfg.Request_processors, sep)
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (sa SIPAgentCfg) AsMapInterface(separator string) interface{} {
	mp := map[string]interface{}{
		utils.EnabledCfg:             sa.Enabled,
		utils.ListenCfg:              sa.Listen,
		utils.ListenNetCfg:           sa.ListenNet,
		utils.TimezoneCfg:            sa.Timezone,
		utils.RetransmissionTimerCfg: sa.RetransmissionTimer.String(),
	}

	requestProcessors := make([]map[string]interface{}, len(sa.RequestProcessors))
	for i, item := range sa.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	mp[utils.RequestProcessorsCfg] = requestProcessors

	if sa.SessionSConns != nil {
		mp[utils.SessionSConnsCfg] = getBiRPCInternalJSONConns(sa.SessionSConns)
	}
	return mp
}

func (SIPAgentCfg) SName() string            { return SIPAgentJSON }
func (sa SIPAgentCfg) CloneSection() Section { return sa.Clone() }

// Clone returns a deep copy of SIPAgentCfg
func (sa SIPAgentCfg) Clone() (cln *SIPAgentCfg) {
	cln = &SIPAgentCfg{
		Enabled:             sa.Enabled,
		Listen:              sa.Listen,
		ListenNet:           sa.ListenNet,
		Timezone:            sa.Timezone,
		RetransmissionTimer: sa.RetransmissionTimer,
	}
	if sa.SessionSConns != nil {
		cln.SessionSConns = utils.CloneStringSlice(sa.SessionSConns)
	}
	if sa.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(sa.RequestProcessors))
		for i, rp := range sa.RequestProcessors {
			cln.RequestProcessors[i] = rp.Clone()
		}
	}
	return
}

// SIPAgentJsonCfg
type SIPAgentJsonCfg struct {
	Enabled              *bool
	Listen               *string
	Listen_net           *string
	Sessions_conns       *[]string
	Timezone             *string
	Retransmission_timer *string
	Request_processors   *[]*ReqProcessorJsnCfg
}

func diffSIPAgentJsonCfg(d *SIPAgentJsonCfg, v1, v2 *SIPAgentCfg, separator string) *SIPAgentJsonCfg {
	if d == nil {
		d = new(SIPAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.Listen != v2.Listen {
		d.Listen = utils.StringPointer(v2.Listen)
	}
	if v1.ListenNet != v2.ListenNet {
		d.Listen_net = utils.StringPointer(v2.ListenNet)
	}
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	if v1.RetransmissionTimer != v2.RetransmissionTimer {
		d.Retransmission_timer = utils.StringPointer(v2.RetransmissionTimer.String())
	}
	d.Request_processors = diffReqProcessorsJsnCfg(d.Request_processors, v1.RequestProcessors, v2.RequestProcessors, separator)
	return d
}
