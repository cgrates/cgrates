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

// DiameterAgentCfg the config section that describes the Diameter Agent
type DiameterAgentCfg struct {
	Enabled           bool   // enables the diameter agent: <true|false>
	ListenNet         string // sctp or tcp
	Listen            string // address where to listen for diameter requests <x.y.z.y:1234>
	DictionariesPath  string
	SessionSConns     []string
	OriginHost        string
	OriginRealm       string
	VendorID          int
	ProductName       string
	ConcurrentReqs    int // limit the maximum number of requests processed
	SyncedConnReqs    bool
	ASRTemplate       string
	RARTemplate       string
	ForcedDisconnect  string
	RequestProcessors []*RequestProcessor
}

func (da *DiameterAgentCfg) loadFromJSONCfg(jsnCfg *DiameterAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		da.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen != nil {
		da.Listen = *jsnCfg.Listen
	}
	if jsnCfg.Listen_net != nil {
		da.ListenNet = *jsnCfg.Listen_net
	}
	if jsnCfg.Dictionaries_path != nil {
		da.DictionariesPath = *jsnCfg.Dictionaries_path
	}
	if jsnCfg.Sessions_conns != nil {
		da.SessionSConns = updateBiRPCInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	if jsnCfg.Origin_host != nil {
		da.OriginHost = *jsnCfg.Origin_host
	}
	if jsnCfg.Origin_realm != nil {
		da.OriginRealm = *jsnCfg.Origin_realm
	}
	if jsnCfg.Vendor_id != nil {
		da.VendorID = *jsnCfg.Vendor_id
	}
	if jsnCfg.Product_name != nil {
		da.ProductName = *jsnCfg.Product_name
	}
	if jsnCfg.Concurrent_requests != nil {
		da.ConcurrentReqs = *jsnCfg.Concurrent_requests
	}
	if jsnCfg.Synced_conn_requests != nil {
		da.SyncedConnReqs = *jsnCfg.Synced_conn_requests
	}
	if jsnCfg.Asr_template != nil {
		da.ASRTemplate = *jsnCfg.Asr_template
	}
	if jsnCfg.Rar_template != nil {
		da.RARTemplate = *jsnCfg.Rar_template
	}
	if jsnCfg.Forced_disconnect != nil {
		da.ForcedDisconnect = *jsnCfg.Forced_disconnect
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
			if err = rp.loadFromJSONCfg(reqProcJsn, separator); err != nil {
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
func (da *DiameterAgentCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:            da.Enabled,
		utils.ListenNetCfg:          da.ListenNet,
		utils.ListenCfg:             da.Listen,
		utils.DictionariesPathCfg:   da.DictionariesPath,
		utils.OriginHostCfg:         da.OriginHost,
		utils.OriginRealmCfg:        da.OriginRealm,
		utils.VendorIDCfg:           da.VendorID,
		utils.ProductNameCfg:        da.ProductName,
		utils.ConcurrentRequestsCfg: da.ConcurrentReqs,
		utils.SyncedConnReqsCfg:     da.SyncedConnReqs,
		utils.ASRTemplateCfg:        da.ASRTemplate,
		utils.RARTemplateCfg:        da.RARTemplate,
		utils.ForcedDisconnectCfg:   da.ForcedDisconnect,
	}

	requestProcessors := make([]map[string]interface{}, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.RequestProcessorsCfg] = requestProcessors

	if da.SessionSConns != nil {
		initialMP[utils.SessionSConnsCfg] = getBiRPCInternalJSONConns(da.SessionSConns)
	}
	return
}

// Clone returns a deep copy of DiameterAgentCfg
func (da DiameterAgentCfg) Clone() (cln *DiameterAgentCfg) {
	cln = &DiameterAgentCfg{
		Enabled:          da.Enabled,
		ListenNet:        da.ListenNet,
		Listen:           da.Listen,
		DictionariesPath: da.DictionariesPath,
		OriginHost:       da.OriginHost,
		OriginRealm:      da.OriginRealm,
		VendorID:         da.VendorID,
		ProductName:      da.ProductName,
		ConcurrentReqs:   da.ConcurrentReqs,
		SyncedConnReqs:   da.SyncedConnReqs,
		ASRTemplate:      da.ASRTemplate,
		RARTemplate:      da.RARTemplate,
		ForcedDisconnect: da.ForcedDisconnect,
	}
	if da.SessionSConns != nil {
		cln.SessionSConns = utils.CloneStringSlice(da.SessionSConns)
	}
	if da.RequestProcessors != nil {
		cln.RequestProcessors = make([]*RequestProcessor, len(da.RequestProcessors))
		for i, req := range da.RequestProcessors {
			cln.RequestProcessors[i] = req.Clone()
		}
	}
	return
}

// DiameterAgent configuration
type DiameterAgentJsonCfg struct {
	Enabled              *bool
	Listen               *string
	Listen_net           *string
	Dictionaries_path    *string
	Sessions_conns       *[]string
	Origin_host          *string
	Origin_realm         *string
	Vendor_id            *int
	Product_name         *string
	Concurrent_requests  *int
	Synced_conn_requests *bool
	Asr_template         *string
	Rar_template         *string
	Forced_disconnect    *string
	Request_processors   *[]*ReqProcessorJsnCfg
}

func diffDiameterAgentJsonCfg(d *DiameterAgentJsonCfg, v1, v2 *DiameterAgentCfg, separator string) *DiameterAgentJsonCfg {
	if d == nil {
		d = new(DiameterAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.ListenNet != v2.ListenNet {
		d.Listen_net = utils.StringPointer(v2.ListenNet)
	}
	if v1.Listen != v2.Listen {
		d.Listen = utils.StringPointer(v2.Listen)
	}
	if v1.DictionariesPath != v2.DictionariesPath {
		d.Dictionaries_path = utils.StringPointer(v2.DictionariesPath)
	}
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}
	if v1.OriginHost != v2.OriginHost {
		d.Origin_host = utils.StringPointer(v2.OriginHost)
	}
	if v1.OriginRealm != v2.OriginRealm {
		d.Origin_realm = utils.StringPointer(v2.OriginRealm)
	}
	if v1.VendorID != v2.VendorID {
		d.Vendor_id = utils.IntPointer(v2.VendorID)
	}
	if v1.ProductName != v2.ProductName {
		d.Product_name = utils.StringPointer(v2.ProductName)
	}
	if v1.ConcurrentReqs != v2.ConcurrentReqs {
		d.Concurrent_requests = utils.IntPointer(v2.ConcurrentReqs)
	}
	if v1.SyncedConnReqs != v2.SyncedConnReqs {
		d.Synced_conn_requests = utils.BoolPointer(v2.SyncedConnReqs)
	}
	if v1.ASRTemplate != v2.ASRTemplate {
		d.Asr_template = utils.StringPointer(v2.ASRTemplate)
	}
	if v1.RARTemplate != v2.RARTemplate {
		d.Rar_template = utils.StringPointer(v2.RARTemplate)
	}
	if v1.ForcedDisconnect != v2.ForcedDisconnect {
		d.Forced_disconnect = utils.StringPointer(v2.ForcedDisconnect)
	}
	d.Request_processors = diffReqProcessorsJsnCfg(d.Request_processors, v1.RequestProcessors, v2.RequestProcessors, separator)
	return d
}
