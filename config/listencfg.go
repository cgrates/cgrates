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

// Listen config section
type ListenCfg struct {
	RPCJSONListen    string // RPC JSON listening address
	RPCGOBListen     string // RPC GOB listening address
	HTTPListen       string // HTTP listening address
	RPCJSONTLSListen string // RPC JSON TLS listening address
	RPCGOBTLSListen  string // RPC GOB TLS listening address
	HTTPTLSListen    string // HTTP TLS listening address
}

//loadFromJsonCfg loads Database config from JsonCfg
func (lstcfg *ListenCfg) loadFromJsonCfg(jsnListenCfg *ListenJsonCfg) (err error) {
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
