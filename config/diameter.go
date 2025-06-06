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

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// DiameterAgentCfg the config section that describes the Diameter Agent
type DiameterAgentCfg struct {
	Enabled           bool   // enables the diameter agent: <true|false>
	ListenNet         string // sctp or tcp
	Listen            string // address where to listen for diameter requests <x.y.z.y:1234>
	DictionariesPath  string
	CeApplications    []string
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
	if jsnCfg.ListenNet != nil {
		da.ListenNet = *jsnCfg.ListenNet
	}
	if jsnCfg.DictionariesPath != nil {
		da.DictionariesPath = *jsnCfg.DictionariesPath
	}
	if jsnCfg.CeApplications != nil {
		da.CeApplications = make([]string, len(*jsnCfg.CeApplications))
		copy(da.CeApplications, *jsnCfg.CeApplications)
	}
	if jsnCfg.SessionSConns != nil {
		da.SessionSConns = make([]string, len(*jsnCfg.SessionSConns))
		for idx, attrConn := range *jsnCfg.SessionSConns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			da.SessionSConns[idx] = attrConn
			if attrConn == utils.MetaInternal ||
				attrConn == rpcclient.BiRPCInternal {
				da.SessionSConns[idx] = utils.ConcatenatedKey(attrConn, utils.MetaSessionS)
			}
		}
	}
	if jsnCfg.StatSConns != nil {
		da.StatSConns = tagInternalConns(*jsnCfg.StatSConns, utils.MetaStats)
	}
	if jsnCfg.ThresholdSConns != nil {
		da.ThresholdSConns = tagInternalConns(*jsnCfg.ThresholdSConns, utils.MetaThresholds)
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
	if jsnCfg.RequestProcessors != nil {
		for _, reqProcJsn := range *jsnCfg.RequestProcessors {
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

// AsMapInterface returns the config as a map[string]any
func (da *DiameterAgentCfg) AsMapInterface(separator string) map[string]any {
	m := map[string]any{
		utils.EnabledCfg:          da.Enabled,
		utils.ListenNetCfg:        da.ListenNet,
		utils.ListenCfg:           da.Listen,
		utils.DictionariesPathCfg: da.DictionariesPath,
		utils.OriginHostCfg:       da.OriginHost,
		utils.OriginRealmCfg:      da.OriginRealm,
		utils.VendorIDCfg:         da.VendorID,
		utils.ProductNameCfg:      da.ProductName,
		utils.SyncedConnReqsCfg:   da.SyncedConnReqs,
		utils.ASRTemplateCfg:      da.ASRTemplate,
		utils.RARTemplateCfg:      da.RARTemplate,
		utils.ForcedDisconnectCfg: da.ForcedDisconnect,
		utils.StatSConnsCfg:       stripInternalConns(da.StatSConns),
		utils.ThresholdSConnsCfg:  stripInternalConns(da.ThresholdSConns),
	}

	if da.CeApplications != nil {
		apps := make([]string, len(da.CeApplications))
		copy(apps, da.CeApplications)
		m[utils.CeApplicationsCfg] = apps
	}

	requestProcessors := make([]map[string]any, len(da.RequestProcessors))
	for i, item := range da.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	m[utils.RequestProcessorsCfg] = requestProcessors

	if da.SessionSConns != nil {
		sessionSConns := make([]string, len(da.SessionSConns))
		for i, item := range da.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			} else if item == utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS) {
				sessionSConns[i] = rpcclient.BiRPCInternal
			}
		}
		m[utils.SessionSConnsCfg] = sessionSConns
	}
	return m
}

// Clone returns a deep copy of DiameterAgentCfg
func (da DiameterAgentCfg) Clone() *DiameterAgentCfg {
	clone := &DiameterAgentCfg{
		Enabled:          da.Enabled,
		ListenNet:        da.ListenNet,
		Listen:           da.Listen,
		DictionariesPath: da.DictionariesPath,
		CeApplications:   slices.Clone(da.CeApplications),
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
