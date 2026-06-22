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
	"maps"
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type DiameterListener struct {
	Network string // sctp or tcp
	Address string // address where to listen for diameter requests <x.y.z.y:1234>
}

// DiameterAgentCfg the config section that describes the Diameter Agent
type DiameterAgentCfg struct {
	Enabled                    bool // enables the diameter agent: <true|false>
	Listeners                  []DiameterListener
	DictionariesPath           string
	DictionariesAppendDefaults bool
	CEApplications             []string
	Conns                      map[string][]*DynamicConns
	OriginHost                 string
	OriginRealm                string
	VendorID                   int
	ProductName                string
	SyncedConnReqs             bool
	ASRTemplate                string
	RARTemplate                string
	ForcedDisconnect           string
	ConnStatusStatQueueIDs     []string
	ConnStatusThresholdIDs     []string
	ConnHealthCheckInterval    time.Duration // peer connection health check interval (0 to disable)
	RequestProcessors          []*RequestProcessor
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
	if jc.Listeners != nil {
		da.Listeners = make([]DiameterListener, 0, len(*jc.Listeners))
		for _, listnr := range *jc.Listeners {
			if listnr == nil {
				continue
			}
			var ls DiameterListener
			if listnr.Address != nil {
				ls.Address = *listnr.Address
			}
			if listnr.Network != nil {
				ls.Network = *listnr.Network
			}
			da.Listeners = append(da.Listeners, ls)
		}
	}
	if jc.DictionariesPath != nil {
		da.DictionariesPath = *jc.DictionariesPath
	}
	if jc.DictionariesAppendDefaults != nil {
		da.DictionariesAppendDefaults = *jc.DictionariesAppendDefaults
	}
	if jc.CEApplications != nil {
		da.CEApplications = make([]string, len(*jc.CEApplications))
		copy(da.CEApplications, *jc.CEApplications)
	}
	if jc.Conns != nil {
		tagged := tagConns(jc.Conns)
		maps.Copy(da.Conns, tagged)
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

// AsMapInterface returns the config as a map[string]any
func (lstn *DiameterListener) AsMapInterface() map[string]any {
	return map[string]any{
		utils.NetworkCfg: lstn.Network,
		utils.AddressCfg: lstn.Address,
	}

}

// AsMapInterface returns the config as a map[string]any.
func (da DiameterAgentCfg) AsMapInterface() any {
	requestProcessors := make([]map[string]any, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface()
	}
	listeners := make([]map[string]any, len(da.Listeners))
	for i, item := range da.Listeners {
		listeners[i] = item.AsMapInterface()
	}
	mp := map[string]any{
		utils.EnabledCfg:                    da.Enabled,
		utils.ListenersCfg:                  listeners,
		utils.DictionariesPathCfg:           da.DictionariesPath,
		utils.DictionariesAppendDefaultsCfg: da.DictionariesAppendDefaults,
		utils.ConnsCfg:                      stripConns(da.Conns),
		utils.ConnStatusStatQueueIDsCfg:     da.ConnStatusStatQueueIDs,
		utils.ConnStatusThresholdIDsCfg:     da.ConnStatusThresholdIDs,
		utils.OriginHostCfg:                 da.OriginHost,
		utils.OriginRealmCfg:                da.OriginRealm,
		utils.VendorIDCfg:                   da.VendorID,
		utils.ProductNameCfg:                da.ProductName,
		utils.SyncedConnReqsCfg:             da.SyncedConnReqs,
		utils.ASRTemplateCfg:                da.ASRTemplate,
		utils.RARTemplateCfg:                da.RARTemplate,
		utils.ForcedDisconnectCfg:           da.ForcedDisconnect,
		utils.ConnHealthCheckIntervalCfg:    da.ConnHealthCheckInterval.String(),
		utils.RequestProcessorsCfg:          requestProcessors,
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
		Enabled:                    da.Enabled,
		Listeners:                  slices.Clone(da.Listeners),
		DictionariesPath:           da.DictionariesPath,
		DictionariesAppendDefaults: da.DictionariesAppendDefaults,
		CEApplications:             slices.Clone(da.CEApplications),
		Conns:                      CloneConnsMap(da.Conns),
		OriginHost:                 da.OriginHost,
		OriginRealm:                da.OriginRealm,
		VendorID:                   da.VendorID,
		ProductName:                da.ProductName,
		SyncedConnReqs:             da.SyncedConnReqs,
		ASRTemplate:                da.ASRTemplate,
		RARTemplate:                da.RARTemplate,
		ForcedDisconnect:           da.ForcedDisconnect,
		ConnStatusStatQueueIDs:     slices.Clone(da.ConnStatusStatQueueIDs),
		ConnStatusThresholdIDs:     slices.Clone(da.ConnStatusThresholdIDs),
		ConnHealthCheckInterval:    da.ConnHealthCheckInterval,
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

type DiamListenerJsnCfg struct {
	Address *string `json:"address"`
	Network *string `json:"network"`
}

// DiameterAgent configuration
type DiameterAgentJsonCfg struct {
	Enabled                    *bool                      `json:"enabled"`
	Listeners                  *[]*DiamListenerJsnCfg     `json:"listeners"`
	DictionariesPath           *string                    `json:"dictionariesPath"`
	DictionariesAppendDefaults *bool                      `json:"dictionariesAppendDefaults"`
	CEApplications             *[]string                  `json:"ceApplications"`
	Conns                      map[string][]*DynamicConns `json:"conns,omitempty"`
	OriginHost                 *string                    `json:"originHost"`
	OriginRealm                *string                    `json:"originRealm"`
	VendorID                   *int                       `json:"vendorID"`
	ProductName                *string                    `json:"productName"`
	SyncedConnRequests         *bool                      `json:"syncedConnRequests"`
	ASRTemplate                *string                    `json:"asrTemplate"`
	RARTemplate                *string                    `json:"rarTemplate"`
	ForcedDisconnect           *string                    `json:"forcedDisconnect"`
	ConnStatusStatQueueIDs     *[]string                  `json:"connStatusStatQueueIDs"`
	ConnStatusThresholdIDs     *[]string                  `json:"connStatusThresholdIDs"`
	ConnHealthCheckInterval    *string                    `json:"connHealthCheckInterval"`
	RequestProcessors          *[]*ReqProcessorJsnCfg     `json:"requestProcessors"`
}

func diffDiameterAgentJsonCfg(d *DiameterAgentJsonCfg, v1, v2 *DiameterAgentCfg) *DiameterAgentJsonCfg {
	if d == nil {
		d = new(DiameterAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !slices.Equal(v1.Listeners, v2.Listeners) {
		listeners := make([]*DiamListenerJsnCfg, len(v2.Listeners))
		for i, listener := range v2.Listeners {
			listeners[i] = &DiamListenerJsnCfg{
				Network: utils.StringPointer(listener.Network),
				Address: utils.StringPointer(listener.Address),
			}
		}
		d.Listeners = &listeners
	}
	if v1.DictionariesPath != v2.DictionariesPath {
		d.DictionariesPath = utils.StringPointer(v2.DictionariesPath)
	}
	if v1.DictionariesAppendDefaults != v2.DictionariesAppendDefaults {
		d.DictionariesAppendDefaults = utils.BoolPointer(v2.DictionariesAppendDefaults)
	}
	if !slices.Equal(v1.CEApplications, v2.CEApplications) {
		d.CEApplications = utils.SliceStringPointer(v2.CEApplications)
	}
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
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
