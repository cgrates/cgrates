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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// DiameterAgentCfg the config section that describes the Diameter Agent
type DiameterAgentCfg struct {
	Enabled                 bool   // enables the diameter agent: <true|false>
	ListenNet               string // sctp or tcp
	Listen                  string // address where to listen for diameter requests <x.y.z.y:1234>
	DictionariesPath        string
	CEApplications          []string
	SessionSConns           []string
	StatSConns              []string
	ThresholdSConns         []string
	OriginHost              string
	OriginRealm             string
	VendorID                int
	ProductName             string
	SyncedConnReqs          bool
	ASRTemplate             string
	RARTemplate             string
	ForcedDisconnect        string
	ConnStatusStatQueueIDs  []string
	ConnStatusThresholdIDs  []string
	ConnHealthCheckInterval time.Duration // peer connection health check interval (0 to disable)
	RequestProcessors       []*RequestProcessor
}

// loadDiameterAgentCfg loads the DiameterAgent section of the configuration
func (da *DiameterAgentCfg) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnDACfg := new(DiameterAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, DiameterAgentJSON, jsnDACfg); err != nil {
		return
	}
	return da.loadFromJSONCfg(jsnDACfg)
}

func (da *DiameterAgentCfg) loadFromJSONCfg(jc *DiameterAgentJsonCfg) (err error) {
	if jc == nil {
		return nil
	}
	if jc.Enabled != nil {
		da.Enabled = *jc.Enabled
	}
	if jc.Listen != nil {
		da.Listen = *jc.Listen
	}
	if jc.ListenNet != nil {
		da.ListenNet = *jc.ListenNet
	}
	if jc.DictionariesPath != nil {
		da.DictionariesPath = *jc.DictionariesPath
	}
	if jc.CEApplications != nil {
		da.CEApplications = make([]string, len(*jc.CEApplications))
		copy(da.CEApplications, *jc.CEApplications)
	}
	if jc.SessionSConns != nil {
		da.SessionSConns = tagInternalConns(*jc.SessionSConns, utils.MetaSessionS)
	}
	if jc.StatSConns != nil {
		da.StatSConns = tagInternalConns(*jc.StatSConns, utils.MetaStats)
	}
	if jc.ThresholdSConns != nil {
		da.ThresholdSConns = tagInternalConns(*jc.ThresholdSConns, utils.MetaThresholds)
	}
	if jc.OriginHost != nil {
		da.OriginHost = *jc.OriginHost
	}
	if jc.OriginRealm != nil {
		da.OriginRealm = *jc.OriginRealm
	}
	if jc.VendorID != nil {
		da.VendorID = *jc.VendorID
	}
	if jc.ProductName != nil {
		da.ProductName = *jc.ProductName
	}
	if jc.SyncedConnRequests != nil {
		da.SyncedConnReqs = *jc.SyncedConnRequests
	}
	if jc.ASRTemplate != nil {
		da.ASRTemplate = *jc.ASRTemplate
	}
	if jc.RARTemplate != nil {
		da.RARTemplate = *jc.RARTemplate
	}
	if jc.ForcedDisconnect != nil {
		da.ForcedDisconnect = *jc.ForcedDisconnect
	}
	if jc.ConnStatusStatQueueIDs != nil {
		da.ConnStatusStatQueueIDs = *jc.ConnStatusStatQueueIDs
	}
	if jc.ConnStatusThresholdIDs != nil {
		da.ConnStatusThresholdIDs = *jc.ConnStatusThresholdIDs
	}
	if jc.ConnHealthCheckInterval != nil {
		da.ConnHealthCheckInterval, err = utils.ParseDurationWithNanosecs(*jc.ConnHealthCheckInterval)
		if err != nil {
			return
		}
	}
	da.RequestProcessors, err = appendRequestProcessors(da.RequestProcessors, jc.RequestProcessors)
	return
}

// AsMapInterface returns the config as a map[string]any.
func (da DiameterAgentCfg) AsMapInterface() any {
	requestProcessors := make([]map[string]any, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	mp := map[string]any{
		utils.EnabledCfg:                 da.Enabled,
		utils.ListenNetCfg:               da.ListenNet,
		utils.ListenCfg:                  da.Listen,
		utils.DictionariesPathCfg:        da.DictionariesPath,
		utils.SessionSConnsCfg:           stripInternalConns(da.SessionSConns),
		utils.StatSConnsCfg:              stripInternalConns(da.StatSConns),
		utils.ThresholdSConnsCfg:         stripInternalConns(da.ThresholdSConns),
		utils.ConnStatusStatQueueIDsCfg:  da.ConnStatusStatQueueIDs,
		utils.ConnStatusThresholdIDsCfg:  da.ConnStatusThresholdIDs,
		utils.OriginHostCfg:              da.OriginHost,
		utils.OriginRealmCfg:             da.OriginRealm,
		utils.VendorIDCfg:                da.VendorID,
		utils.ProductNameCfg:             da.ProductName,
		utils.SyncedConnReqsCfg:          da.SyncedConnReqs,
		utils.ASRTemplateCfg:             da.ASRTemplate,
		utils.RARTemplateCfg:             da.RARTemplate,
		utils.ForcedDisconnectCfg:        da.ForcedDisconnect,
		utils.ConnHealthCheckIntervalCfg: da.ConnHealthCheckInterval.String(),
		utils.RequestProcessorsCfg:       requestProcessors,
	}
	if da.CEApplications != nil {
		apps := make([]string, len(da.CEApplications))
		copy(apps, da.CEApplications)
		mp[utils.CEApplicationsCfg] = apps
	}
	return mp
}

func (DiameterAgentCfg) SName() string            { return DiameterAgentJSON }
func (da DiameterAgentCfg) CloneSection() Section { return da.Clone() }

// Clone returns a deep copy of DiameterAgentCfg
func (da DiameterAgentCfg) Clone() *DiameterAgentCfg {
	clone := &DiameterAgentCfg{
		Enabled:                 da.Enabled,
		ListenNet:               da.ListenNet,
		Listen:                  da.Listen,
		DictionariesPath:        da.DictionariesPath,
		CEApplications:          slices.Clone(da.CEApplications),
		SessionSConns:           slices.Clone(da.SessionSConns),
		StatSConns:              slices.Clone(da.StatSConns),
		ThresholdSConns:         slices.Clone(da.ThresholdSConns),
		OriginHost:              da.OriginHost,
		OriginRealm:             da.OriginRealm,
		VendorID:                da.VendorID,
		ProductName:             da.ProductName,
		SyncedConnReqs:          da.SyncedConnReqs,
		ASRTemplate:             da.ASRTemplate,
		RARTemplate:             da.RARTemplate,
		ForcedDisconnect:        da.ForcedDisconnect,
		ConnStatusStatQueueIDs:  slices.Clone(da.ConnStatusStatQueueIDs),
		ConnStatusThresholdIDs:  slices.Clone(da.ConnStatusThresholdIDs),
		ConnHealthCheckInterval: da.ConnHealthCheckInterval,
	}
	if da.CEApplications != nil {
		clone.CEApplications = make([]string, len(da.CEApplications))
		copy(clone.CEApplications, da.CEApplications)
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
	Enabled                 *bool                  `json:"enabled"`
	Listen                  *string                `json:"listen"`
	ListenNet               *string                `json:"listen_net"`
	DictionariesPath        *string                `json:"dictionaries_path"`
	CEApplications          *[]string              `json:"ce_applications"`
	SessionSConns           *[]string              `json:"sessions_conns"`
	StatSConns              *[]string              `json:"stats_conns"`
	ThresholdSConns         *[]string              `json:"thresholds_conns"`
	OriginHost              *string                `json:"origin_host"`
	OriginRealm             *string                `json:"origin_realm"`
	VendorID                *int                   `json:"vendor_id"`
	ProductName             *string                `json:"product_name"`
	SyncedConnRequests      *bool                  `json:"synced_conn_requests"`
	ASRTemplate             *string                `json:"asr_template"`
	RARTemplate             *string                `json:"rar_template"`
	ForcedDisconnect        *string                `json:"forced_disconnect"`
	ConnStatusStatQueueIDs  *[]string              `json:"conn_status_stat_queue_ids"`
	ConnStatusThresholdIDs  *[]string              `json:"conn_status_threshold_ids"`
	ConnHealthCheckInterval *string                `json:"conn_health_check_interval"`
	RequestProcessors       *[]*ReqProcessorJsnCfg `json:"request_processors"`
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
	if !slices.Equal(v1.CEApplications, v2.CEApplications) {
		d.CEApplications = utils.SliceStringPointer(v2.CEApplications)
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
	if !slices.Equal(v1.ConnStatusStatQueueIDs, v2.ConnStatusStatQueueIDs) {
		d.ConnStatusStatQueueIDs = utils.SliceStringPointer(v2.ConnStatusStatQueueIDs)
	}
	if !slices.Equal(v1.ConnStatusThresholdIDs, v2.ConnStatusThresholdIDs) {
		d.ConnStatusThresholdIDs = utils.SliceStringPointer(v2.ConnStatusThresholdIDs)
	}
	if v1.ConnHealthCheckInterval != v2.ConnHealthCheckInterval {
		d.ConnHealthCheckInterval = utils.StringPointer(v2.ConnHealthCheckInterval.String())
	}
	d.RequestProcessors = diffReqProcessorsJsnCfg(d.RequestProcessors, v1.RequestProcessors, v2.RequestProcessors)
	return d
}
