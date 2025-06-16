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

// DiameterAgentCfg the config section that describes the Diameter Agent
type DiameterAgentCfg struct {
	Enabled           bool   // enables the diameter agent: <true|false>
	ListenNet         string // sctp or tcp
	Listen            string // address where to listen for diameter requests <x.y.z.y:1234>
	DictionariesPath  string
	SessionSConns     []string
	StatSConns        []string
	ThresholdSConns   []string
	OriginHost        string
	OriginRealm       string
	VendorID          int
	ProductName       string
	SyncedConnReqs    bool
	ASRTemplate       string
	RARTemplate       string
	ForcedDisconnect  string
	RequestProcessors []*RequestProcessor
}

// loadDiameterAgentCfg loads the DiameterAgent section of the configuration
func (da *DiameterAgentCfg) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnDACfg := new(DiameterAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, DiameterAgentJSON, jsnDACfg); err != nil {
		return
	}
	return da.loadFromJSONCfg(jsnDACfg)
}

func (da *DiameterAgentCfg) loadFromJSONCfg(jsnCfg *DiameterAgentJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		da.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen != nil {
		da.Listen = *jsnCfg.Listen
	}
	if jsnCfg.ListenNet != nil {
		da.ListenNet = *jsnCfg.ListenNet
	}
	if jsnCfg.DictionariesPath != nil {
		da.DictionariesPath = *jsnCfg.DictionariesPath
	}
	if jsnCfg.SessionSConns != nil {
		da.SessionSConns = updateBiRPCInternalConns(*jsnCfg.SessionSConns, utils.MetaSessionS)
	}
	if jsnCfg.StatSConns != nil {
		da.StatSConns = updateBiRPCInternalConns(*jsnCfg.StatSConns, utils.MetaStats)
	}
	if jsnCfg.ThresholdSConns != nil {
		da.ThresholdSConns = updateBiRPCInternalConns(*jsnCfg.ThresholdSConns, utils.MetaThresholds)
	}
	if jsnCfg.OriginHost != nil {
		da.OriginHost = *jsnCfg.OriginHost
	}
	if jsnCfg.OriginRealm != nil {
		da.OriginRealm = *jsnCfg.OriginRealm
	}
	if jsnCfg.VendorID != nil {
		da.VendorID = *jsnCfg.VendorID
	}
	if jsnCfg.ProductName != nil {
		da.ProductName = *jsnCfg.ProductName
	}
	if jsnCfg.SyncedConnRequests != nil {
		da.SyncedConnReqs = *jsnCfg.SyncedConnRequests
	}
	if jsnCfg.ASRTemplate != nil {
		da.ASRTemplate = *jsnCfg.ASRTemplate
	}
	if jsnCfg.RARTemplate != nil {
		da.RARTemplate = *jsnCfg.RARTemplate
	}
	if jsnCfg.ForcedDisconnect != nil {
		da.ForcedDisconnect = *jsnCfg.ForcedDisconnect
	}
	da.RequestProcessors, err = appendRequestProcessors(da.RequestProcessors, jsnCfg.RequestProcessors)
	return
}

// AsMapInterface returns the config as a map[string]any
func (da DiameterAgentCfg) AsMapInterface() any {
	requestProcessors := make([]map[string]any, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	mp := map[string]any{
		utils.EnabledCfg:           da.Enabled,
		utils.ListenNetCfg:         da.ListenNet,
		utils.ListenCfg:            da.Listen,
		utils.DictionariesPathCfg:  da.DictionariesPath,
		utils.SessionSConnsCfg:     getBiRPCInternalJSONConns(da.SessionSConns),
		utils.StatSConnsCfg:        getBiRPCInternalJSONConns(da.StatSConns),
		utils.ThresholdSConnsCfg:   getBiRPCInternalJSONConns(da.ThresholdSConns),
		utils.OriginHostCfg:        da.OriginHost,
		utils.OriginRealmCfg:       da.OriginRealm,
		utils.VendorIDCfg:          da.VendorID,
		utils.ProductNameCfg:       da.ProductName,
		utils.SyncedConnReqsCfg:    da.SyncedConnReqs,
		utils.ASRTemplateCfg:       da.ASRTemplate,
		utils.RARTemplateCfg:       da.RARTemplate,
		utils.ForcedDisconnectCfg:  da.ForcedDisconnect,
		utils.RequestProcessorsCfg: requestProcessors,
	}
	return mp
}

func (DiameterAgentCfg) SName() string            { return DiameterAgentJSON }
func (da DiameterAgentCfg) CloneSection() Section { return da.Clone() }

// Clone returns a deep copy of DiameterAgentCfg
func (da DiameterAgentCfg) Clone() *DiameterAgentCfg {
	clone := &DiameterAgentCfg{
		Enabled:          da.Enabled,
		ListenNet:        da.ListenNet,
		Listen:           da.Listen,
		DictionariesPath: da.DictionariesPath,
		SessionSConns:    slices.Clone(da.SessionSConns),
		StatSConns:       slices.Clone(da.StatSConns),
		ThresholdSConns:  slices.Clone(da.ThresholdSConns),
		OriginHost:       da.OriginHost,
		OriginRealm:      da.OriginRealm,
		VendorID:         da.VendorID,
		ProductName:      da.ProductName,
		SyncedConnReqs:   da.SyncedConnReqs,
		ASRTemplate:      da.ASRTemplate,
		RARTemplate:      da.RARTemplate,
		ForcedDisconnect: da.ForcedDisconnect,
	}
	if da.RequestProcessors != nil {
		clone.RequestProcessors = make([]*RequestProcessor, len(da.RequestProcessors))
		for i, req := range da.RequestProcessors {
			clone.RequestProcessors[i] = req.Clone()
		}
	}
	return clone
}

// DiameterAgent configuration
type DiameterAgentJsonCfg struct {
	Enabled            *bool                  `json:"enabled"`
	Listen             *string                `json:"listen"`
	ListenNet          *string                `json:"listen_net"`
	DictionariesPath   *string                `json:"dictionaries_path"`
	SessionSConns      *[]string              `json:"sessions_conns"`
	StatSConns         *[]string              `json:"stats_conns"`
	ThresholdSConns    *[]string              `json:"thresholds_conns"`
	OriginHost         *string                `json:"origin_host"`
	OriginRealm        *string                `json:"origin_realm"`
	VendorID           *int                   `json:"vendor_id"`
	ProductName        *string                `json:"product_name"`
	SyncedConnRequests *bool                  `json:"synced_conn_requests"`
	ASRTemplate        *string                `json:"asr_template"`
	RARTemplate        *string                `json:"rar_template"`
	ForcedDisconnect   *string                `json:"forced_disconnect"`
	RequestProcessors  *[]*ReqProcessorJsnCfg `json:"request_processors"`
}

func diffDiameterAgentJsonCfg(d *DiameterAgentJsonCfg, v1, v2 *DiameterAgentCfg) *DiameterAgentJsonCfg {
	if d == nil {
		d = new(DiameterAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.ListenNet != v2.ListenNet {
		d.ListenNet = utils.StringPointer(v2.ListenNet)
	}
	if v1.Listen != v2.Listen {
		d.Listen = utils.StringPointer(v2.Listen)
	}
	if v1.DictionariesPath != v2.DictionariesPath {
		d.DictionariesPath = utils.StringPointer(v2.DictionariesPath)
	}
	if !slices.Equal(v1.SessionSConns, v2.SessionSConns) {
		d.SessionSConns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.StatSConns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.StatSConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.ThresholdSConns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.ThresholdSConns))
	}
	if v1.OriginHost != v2.OriginHost {
		d.OriginHost = utils.StringPointer(v2.OriginHost)
	}
	if v1.OriginRealm != v2.OriginRealm {
		d.OriginRealm = utils.StringPointer(v2.OriginRealm)
	}
	if v1.VendorID != v2.VendorID {
		d.VendorID = utils.IntPointer(v2.VendorID)
	}
	if v1.ProductName != v2.ProductName {
		d.ProductName = utils.StringPointer(v2.ProductName)
	}
	if v1.SyncedConnReqs != v2.SyncedConnReqs {
		d.SyncedConnRequests = utils.BoolPointer(v2.SyncedConnReqs)
	}
	if v1.ASRTemplate != v2.ASRTemplate {
		d.ASRTemplate = utils.StringPointer(v2.ASRTemplate)
	}
	if v1.RARTemplate != v2.RARTemplate {
		d.RARTemplate = utils.StringPointer(v2.RARTemplate)
	}
	if v1.ForcedDisconnect != v2.ForcedDisconnect {
		d.ForcedDisconnect = utils.StringPointer(v2.ForcedDisconnect)
	}
	d.RequestProcessors = diffReqProcessorsJsnCfg(d.RequestProcessors, v1.RequestProcessors, v2.RequestProcessors)
	return d
}
