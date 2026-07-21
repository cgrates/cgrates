// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// ListenCfg is the listen config section
type ListenCfg struct {
	RPCJSONListen    string // RPC JSON listening address
	RPCGOBListen     string // RPC GOB listening address
	HTTPListen       string // HTTP listening address
	RPCJSONTLSListen string // RPC JSON TLS listening address
	RPCGOBTLSListen  string // RPC GOB TLS listening address
	HTTPTLSListen    string // HTTP TLS listening address
}

// loadListenCfg loads the Listen section of the configuration
func (lstcfg *ListenCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnListenCfg := new(ListenJsonCfg)
	if err = jsnCfg.GetSection(ctx, ListenJSON, jsnListenCfg); err != nil {
		return
	}
	return lstcfg.loadFromJSONCfg(jsnListenCfg)
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

// AsMapInterface returns the config as a map[string]any
func (lstcfg ListenCfg) AsMapInterface() any {
	return map[string]any{
		utils.RPCJSONListenCfg:    lstcfg.RPCJSONListen,
		utils.RPCGOBListenCfg:     lstcfg.RPCGOBListen,
		utils.HTTPListenCfg:       lstcfg.HTTPListen,
		utils.RPCJSONTLSListenCfg: lstcfg.RPCJSONTLSListen,
		utils.RPCGOBTLSListenCfg:  lstcfg.RPCGOBTLSListen,
		utils.HTTPTLSListenCfg:    lstcfg.HTTPTLSListen,
	}
}

func (ListenCfg) SName() string                { return ListenJSON }
func (lstcfg ListenCfg) CloneSection() Section { return lstcfg.Clone() }

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

// Listen config section
type ListenJsonCfg struct {
	Rpc_json     *string `json:"rpcJSON"`
	Rpc_gob      *string `json:"rpcGOB"`
	Http         *string `json:"http"`
	Rpc_json_tls *string `json:"rpcJSONtls"`
	Rpc_gob_tls  *string `json:"rpcGOBtls"`
	Http_tls     *string `json:"httpTLS"`
}

func diffListenJsonCfg(d *ListenJsonCfg, v1, v2 *ListenCfg) *ListenJsonCfg {
	if d == nil {
		d = new(ListenJsonCfg)
	}
	if v1.RPCJSONListen != v2.RPCJSONListen {
		d.Rpc_json = utils.StringPointer(v2.RPCJSONListen)
	}
	if v1.RPCGOBListen != v2.RPCGOBListen {
		d.Rpc_gob = utils.StringPointer(v2.RPCGOBListen)
	}
	if v1.HTTPListen != v2.HTTPListen {
		d.Http = utils.StringPointer(v2.HTTPListen)
	}
	if v1.RPCJSONTLSListen != v2.RPCJSONTLSListen {
		d.Rpc_json_tls = utils.StringPointer(v2.RPCJSONTLSListen)
	}
	if v1.RPCGOBTLSListen != v2.RPCGOBTLSListen {
		d.Rpc_gob_tls = utils.StringPointer(v2.RPCGOBTLSListen)
	}
	if v1.HTTPTLSListen != v2.HTTPTLSListen {
		d.Http_tls = utils.StringPointer(v2.HTTPTLSListen)
	}
	return d
}
