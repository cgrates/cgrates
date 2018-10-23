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
