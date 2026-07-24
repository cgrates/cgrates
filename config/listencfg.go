// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "github.com/cgrates/cgrates/utils"

// Listen config section
type ListenCfg struct {
	RPCJSONListen    string // RPC JSON listening address
	RPCGOBListen     string // RPC GOB listening address
	HTTPListen       string // HTTP listening address
	RPCJSONTLSListen string // RPC JSON TLS listening address
	RPCGOBTLSListen  string // RPC GOB TLS listening address
	HTTPTLSListen    string // HTTP TLS listening address
}

// loadFromJsonCfg loads Database config from JsonCfg
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

func (lstcfg *ListenCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.RPCJSONListenCfg:    lstcfg.RPCJSONListen,
		utils.RPCGOBListenCfg:     lstcfg.RPCGOBListen,
		utils.HTTPListenCfg:       lstcfg.HTTPListen,
		utils.RPCJSONTLSListenCfg: lstcfg.RPCJSONTLSListen,
		utils.RPCGOBTLSListenCfg:  lstcfg.RPCGOBTLSListen,
		utils.HTTPTLSListenCfg:    lstcfg.HTTPTLSListen,
	}
}
