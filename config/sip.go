// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	Conns               map[string][]*DynamicConns
	Timezone            string
	RetransmissionTimer time.Duration // timeout replies if not reaching back
	RequestProcessors   []*RequestProcessor
}

// loadSIPAgentCfg loads the sip_agent section of the configuration
func (sa *SIPAgentCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnSIPAgentCfg := new(SIPAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, SIPAgentJSON, jsnSIPAgentCfg); err != nil {
		return
	}
	return sa.loadFromJSONCfg(jsnSIPAgentCfg)
}

func (sa *SIPAgentCfg) loadFromJSONCfg(jsnCfg *SIPAgentJsonCfg) (err error) {
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
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		for connType, opts := range tagged {
			sa.Conns[connType] = opts
		}
	}
	if jsnCfg.RetransmissionTimer != nil {
		if sa.RetransmissionTimer, err = utils.ParseDurationWithNanosecs(*jsnCfg.RetransmissionTimer); err != nil {
			return err
		}
	}
	sa.RequestProcessors, err = appendRequestProcessors(sa.RequestProcessors, jsnCfg.RequestProcessors)
	return
}

// AsMapInterface returns the config as a map[string]any
func (sa SIPAgentCfg) AsMapInterface() any {
	requestProcessors := make([]map[string]any, len(sa.RequestProcessors))
	for i, item := range sa.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	mp := map[string]any{
		utils.EnabledCfg:             sa.Enabled,
		utils.ListenCfg:              sa.Listen,
		utils.ListenNetCfg:           sa.ListenNet,
		utils.ConnsCfg:               stripConns(sa.Conns),
		utils.TimezoneCfg:            sa.Timezone,
		utils.RetransmissionTimerCfg: sa.RetransmissionTimer.String(),
		utils.RequestProcessorsCfg:   requestProcessors,
	}
	return mp
}

func (SIPAgentCfg) SName() string            { return SIPAgentJSON }
func (sa SIPAgentCfg) CloneSection() Section { return sa.Clone() }

// Clone returns a deep copy of SIPAgentCfg
func (sa SIPAgentCfg) Clone() *SIPAgentCfg {
	clone := &SIPAgentCfg{
		Enabled:             sa.Enabled,
		Listen:              sa.Listen,
		ListenNet:           sa.ListenNet,
		Conns:               CloneConnsMap(sa.Conns),
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

// SIPAgentJsonCfg
type SIPAgentJsonCfg struct {
	Enabled             *bool                      `json:"enabled"`
	Listen              *string                    `json:"listen"`
	ListenNet           *string                    `json:"listenNet"`
	Conns               map[string][]*DynamicConns `json:"conns,omitempty"`
	Timezone            *string                    `json:"timezone"`
	RetransmissionTimer *string                    `json:"retransmissionTimer"`
	RequestProcessors   *[]*ReqProcessorJsnCfg     `json:"requestProcessors"`
}

func diffSIPAgentJsonCfg(d *SIPAgentJsonCfg, v1, v2 *SIPAgentCfg) *SIPAgentJsonCfg {
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
		d.ListenNet = utils.StringPointer(v2.ListenNet)
	}
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	if v1.RetransmissionTimer != v2.RetransmissionTimer {
		d.RetransmissionTimer = utils.StringPointer(v2.RetransmissionTimer.String())
	}
	d.RequestProcessors = diffReqProcessorsJsnCfg(d.RequestProcessors, v1.RequestProcessors, v2.RequestProcessors)
	return d
}
