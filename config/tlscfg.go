// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "github.com/cgrates/cgrates/utils"

// AttributeSCfg is the configuration of attribute service
type TlsCfg struct {
	ServerCerificate string
	ServerKey        string
	ServerPolicy     int
	ServerName       string
	ClientCerificate string
	ClientKey        string
	CaCertificate    string
}

func (tls *TlsCfg) loadFromJsonCfg(jsnCfg *TlsJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Server_certificate != nil {
		tls.ServerCerificate = *jsnCfg.Server_certificate
	}
	if jsnCfg.Server_key != nil {
		tls.ServerKey = *jsnCfg.Server_key
	}
	if jsnCfg.Client_certificate != nil {
		tls.ClientCerificate = *jsnCfg.Client_certificate
	}
	if jsnCfg.Client_key != nil {
		tls.ClientKey = *jsnCfg.Client_key
	}
	if jsnCfg.Ca_certificate != nil {
		tls.CaCertificate = *jsnCfg.Ca_certificate
	}
	if jsnCfg.Server_name != nil {
		tls.ServerName = *jsnCfg.Server_name
	}
	if jsnCfg.Server_policy != nil {
		tls.ServerPolicy = *jsnCfg.Server_policy
	}
	return
}

func (tls *TlsCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.ServerCerificateCfg: tls.ServerCerificate,
		utils.ServerKeyCfg:        tls.ServerKey,
		utils.ServerPolicyCfg:     tls.ServerPolicy,
		utils.ServerNameCfg:       tls.ServerName,
		utils.ClientCerificateCfg: tls.ClientCerificate,
		utils.ClientKeyCfg:        tls.ClientKey,
		utils.CaCertificateCfg:    tls.CaCertificate,
	}

}
