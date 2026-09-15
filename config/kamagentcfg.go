// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// KamConnCfg represents one connection instance towards Kamailio
type KamConnCfg struct {
	Alias                string
	Address              string
	Reconnects           int
	MaxReconnectInterval time.Duration
}

func (kamCfg *KamConnCfg) loadFromJSONCfg(jsnCfg *KamConnJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Address != nil {
		kamCfg.Address = *jsnCfg.Address
	}
	if jsnCfg.Alias != nil {
		kamCfg.Alias = *jsnCfg.Alias
	}
	if jsnCfg.Reconnects != nil {
		kamCfg.Reconnects = *jsnCfg.Reconnects
	}
	if jsnCfg.Max_reconnect_interval != nil {
		if kamCfg.MaxReconnectInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Max_reconnect_interval); err != nil {
			return
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (kamCfg KamConnCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.AliasCfg:                kamCfg.Alias,
		utils.AddressCfg:              kamCfg.Address,
		utils.ReconnectsCfg:           kamCfg.Reconnects,
		utils.MaxReconnectIntervalCfg: kamCfg.MaxReconnectInterval.String(),
	}
}

// Clone returns a deep copy of KamConnCfg
func (kamCfg KamConnCfg) Clone() *KamConnCfg {
	return &KamConnCfg{
		Alias:                kamCfg.Alias,
		Address:              kamCfg.Address,
		Reconnects:           kamCfg.Reconnects,
		MaxReconnectInterval: kamCfg.MaxReconnectInterval,
	}
}

// KamAgentCfg is the Kamailio config section
type KamAgentCfg struct {
	Enabled           bool
	Conns             map[string][]*DynamicConns
	EvapiConns        []*KamConnCfg
	Timezone          string
	RequestProcessors []*RequestProcessor
}

// loadKamAgentCfg loads the KamAgent section of the configuration
func (ka *KamAgentCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnKamAgentCfg := new(KamAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, KamailioAgentJSON, jsnKamAgentCfg); err != nil {
		return
	}
	return ka.loadFromJSONCfg(jsnKamAgentCfg)
}

func (ka *KamAgentCfg) loadFromJSONCfg(jsnCfg *KamAgentJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		ka.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		for connType, opts := range tagged {
			ka.Conns[connType] = opts
		}
	}
	if jsnCfg.Evapi_conns != nil {
		ka.EvapiConns = make([]*KamConnCfg, len(*jsnCfg.Evapi_conns))
		for idx, jsnConnCfg := range *jsnCfg.Evapi_conns {
			ka.EvapiConns[idx] = getDftKamConnCfg()
			ka.EvapiConns[idx].loadFromJSONCfg(jsnConnCfg)
		}
	}
	if jsnCfg.Timezone != nil {
		ka.Timezone = *jsnCfg.Timezone
	}
	var err error
	if ka.RequestProcessors, err = appendRequestProcessors(ka.RequestProcessors,
		jsnCfg.Request_processors); err != nil {
		return err
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (ka KamAgentCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.EnabledCfg:  ka.Enabled,
		utils.TimezoneCfg: ka.Timezone,
	}
	if ka.EvapiConns != nil {
		evapiConns := make([]map[string]any, len(ka.EvapiConns))
		for i, item := range ka.EvapiConns {
			evapiConns[i] = item.AsMapInterface()
		}
		mp[utils.EvapiConnsCfg] = evapiConns
	}
	mp[utils.ConnsCfg] = stripConns(ka.Conns)
	requestProcessors := make([]map[string]any, len(ka.RequestProcessors))
	for i, item := range ka.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	mp[utils.RequestProcessorsCfg] = requestProcessors
	return mp
}

func (KamAgentCfg) SName() string            { return KamailioAgentJSON }
func (ka KamAgentCfg) CloneSection() Section { return ka.Clone() }

// Clone returns a deep copy of KamAgentCfg
func (ka KamAgentCfg) Clone() (cln *KamAgentCfg) {
	cln = &KamAgentCfg{
		Enabled:  ka.Enabled,
		Timezone: ka.Timezone,
		Conns:    CloneConnsMap(ka.Conns),
	}
	if ka.EvapiConns != nil {
		cln.EvapiConns = make([]*KamConnCfg, len(ka.EvapiConns))
		for i, req := range ka.EvapiConns {
			cln.EvapiConns[i] = req.Clone()
		}
	}
	if ka.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(ka.RequestProcessors))
		for i, req := range ka.RequestProcessors {
			cln.RequestProcessors[i] = req.Clone()
		}
	}
	return
}

// Represents one connection instance towards Kamailio
type KamConnJsonCfg struct {
	Alias                  *string
	Address                *string `json:"address"`
	Reconnects             *int    `json:"reconnects"`
	Max_reconnect_interval *string `json:"maxReconnectInterval"`
}

func diffKamConnJsonCfg(v1, v2 *KamConnCfg) (d *KamConnJsonCfg) {
	d = new(KamConnJsonCfg)
	if v1.Alias != v2.Alias {
		d.Alias = utils.StringPointer(v2.Alias)
	}
	if v1.Address != v2.Address {
		d.Address = utils.StringPointer(v2.Address)
	}
	if v1.Reconnects != v2.Reconnects {
		d.Reconnects = utils.IntPointer(v2.Reconnects)
	}
	if v1.MaxReconnectInterval != v2.MaxReconnectInterval {
		d.Max_reconnect_interval = utils.StringPointer(v2.MaxReconnectInterval.String())
	}
	return
}

// KamAgentJsonCfg kamailio config section
type KamAgentJsonCfg struct {
	Enabled            *bool
	Conns              map[string][]*DynamicConns `json:"conns,omitempty"`
	Evapi_conns        *[]*KamConnJsonCfg         `json:"evapiConns"`
	Timezone           *string
	Request_processors *[]*ReqProcessorJsnCfg `json:"requestProcessors"`
}

func equalsKamConnsCfg(v1, v2 []*KamConnCfg) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].Alias != v2[i].Alias ||
			v1[i].Address != v2[i].Address ||
			v1[i].Reconnects != v2[i].Reconnects ||
			v1[i].MaxReconnectInterval != v2[i].MaxReconnectInterval {
			return false
		}
	}
	return true
}

func diffKamAgentJsonCfg(d *KamAgentJsonCfg, v1, v2 *KamAgentCfg) *KamAgentJsonCfg {
	if d == nil {
		d = new(KamAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	if !equalsKamConnsCfg(v1.EvapiConns, v2.EvapiConns) {
		dft := getDftKamConnCfg()
		conns := make([]*KamConnJsonCfg, len(v2.EvapiConns))
		for i, conn := range v2.EvapiConns {
			conns[i] = diffKamConnJsonCfg(dft, conn)
		}
		d.Evapi_conns = &conns
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	d.Request_processors = diffReqProcessorsJsnCfg(d.Request_processors, v1.RequestProcessors, v2.RequestProcessors)
	return d
}
