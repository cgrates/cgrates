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

import "github.com/cgrates/cgrates/utils"

// ListenCfg is the listen config section
type ListenCfg struct {
	RPCJSONListen    string // RPC JSON listening address
	RPCGOBListen     string // RPC GOB listening address
	HTTPListen       string // HTTP listening address
	RPCJSONTLSListen string // RPC JSON TLS listening address
	RPCGOBTLSListen  string // RPC GOB TLS listening address
	HTTPTLSListen    string // HTTP TLS listening address
}

// loadFromJSONCfg loads Database config from JsonCfg
func (lstcfg *ListenCfg) loadFromJSONCfg(jsnListenCfg *ListenJsonCfg) (err error) {
	if jsnListenCfg == nil {
		return nil
	}
	if jsnListenCfg.Rpc_json != nil {
		lstcfg.RPCJSONListen = *jsnListenCfg.Rpc_json
	}
	if jsnListenCfg.Rpc_gob != nil {
		lstcfg.RPCGOBListen = *jsnListenCfg.Rpc_gob
	}
	if jsnListenCfg.Http != nil {
		lstcfg.HTTPListen = *jsnListenCfg.Http
	}
	if jsnListenCfg.Rpc_json_tls != nil && *jsnListenCfg.Rpc_json_tls != "" {
		lstcfg.RPCJSONTLSListen = *jsnListenCfg.Rpc_json_tls
	}
	if jsnListenCfg.Rpc_gob_tls != nil && *jsnListenCfg.Rpc_gob_tls != "" {
		lstcfg.RPCGOBTLSListen = *jsnListenCfg.Rpc_gob_tls
	}
	if jsnListenCfg.Http_tls != nil && *jsnListenCfg.Http_tls != "" {
		lstcfg.HTTPTLSListen = *jsnListenCfg.Http_tls
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (lstcfg *ListenCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.RPCJSONListenCfg:    lstcfg.RPCJSONListen,
		utils.RPCGOBListenCfg:     lstcfg.RPCGOBListen,
		utils.HTTPListenCfg:       lstcfg.HTTPListen,
		utils.RPCJSONTLSListenCfg: lstcfg.RPCJSONTLSListen,
		utils.RPCGOBTLSListenCfg:  lstcfg.RPCGOBTLSListen,
		utils.HTTPTLSListenCfg:    lstcfg.HTTPTLSListen,
	}
}

// Clone returns a deep copy of ListenCfg
func (lstcfg ListenCfg) Clone() *ListenCfg {
	return &ListenCfg{
		RPCJSONListen:    lstcfg.RPCJSONListen,
		RPCGOBListen:     lstcfg.RPCGOBListen,
		HTTPListen:       lstcfg.HTTPListen,
		RPCJSONTLSListen: lstcfg.RPCJSONTLSListen,
		RPCGOBTLSListen:  lstcfg.RPCGOBTLSListen,
		HTTPTLSListen:    lstcfg.HTTPTLSListen,
	}
}
