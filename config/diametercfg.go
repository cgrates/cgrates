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
	"strings"

	"github.com/cgrates/cgrates/utils"
)

type DiameterAgentCfg struct {
	Enabled           bool   // enables the diameter agent: <true|false>
	ListenNet         string // sctp or tcp
	Listen            string // address where to listen for diameter requests <x.y.z.y:1234>
	DictionariesPath  string
	SessionSConns     []string
	OriginHost        string
	OriginRealm       string
	VendorId          int
	ProductName       string
	ConcurrentReqs    int // limit the maximum number of requests processed
	SyncedConnReqs    bool
	ASRTemplate       string
	RARTemplate       string
	ForcedDisconnect  string
	RequestProcessors []*RequestProcessor
}

func (da *DiameterAgentCfg) loadFromJsonCfg(jsnCfg *DiameterAgentJsonCfg, separator string) (err error) {
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
		da.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, attrConn := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if attrConn == utils.MetaInternal {
				da.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				da.SessionSConns[idx] = attrConn
			}
		}
	}
	if jsnCfg.Origin_host != nil {
		da.OriginHost = *jsnCfg.Origin_host
	}
	if jsnCfg.Origin_realm != nil {
		da.OriginRealm = *jsnCfg.Origin_realm
	}
	if jsnCfg.Vendor_id != nil {
		da.VendorId = *jsnCfg.Vendor_id
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

func (ds *DiameterAgentCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:            ds.Enabled,
		utils.ListenNetCfg:          ds.ListenNet,
		utils.ListenCfg:             ds.Listen,
		utils.DictionariesPathCfg:   ds.DictionariesPath,
		utils.OriginHostCfg:         ds.OriginHost,
		utils.OriginRealmCfg:        ds.OriginRealm,
		utils.VendorIDCfg:           ds.VendorId,
		utils.ProductNameCfg:        ds.ProductName,
		utils.ConcurrentRequestsCfg: ds.ConcurrentReqs,
		utils.SyncedConnReqsCfg:     ds.SyncedConnReqs,
		utils.ASRTemplateCfg:        ds.ASRTemplate,
		utils.RARTemplateCfg:        ds.RARTemplate,
		utils.ForcedDisconnectCfg:   ds.ForcedDisconnect,
	}

	requestProcessors := make([]map[string]interface{}, len(ds.RequestProcessors))
	for i, item := range ds.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.RequestProcessorsCfg] = requestProcessors

	if ds.SessionSConns != nil {
		sessionSConns := make([]string, len(ds.SessionSConns))
		for i, item := range ds.SessionSConns {
			buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			if item == buf {
				sessionSConns[i] = strings.TrimSuffix(item, utils.CONCATENATED_KEY_SEP+utils.MetaSessionS)
			} else {
				sessionSConns[i] = item
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
	}
	return
}
