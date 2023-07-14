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

type SentryPeerCfg struct {
	ClientID     string
	ClientSecret string
	Url          string
}

func (sp *SentryPeerCfg) loadFromJSONCfg(jsnCfg *SentryPeerJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Url != nil {
		sp.Url = *jsnCfg.Url
	}
	if jsnCfg.ClientSecret != nil {
		sp.ClientSecret = *jsnCfg.ClientSecret
	}
	if jsnCfg.ClientID != nil {
		sp.ClientID = *jsnCfg.ClientID
	}
	return
}

func (sp *SentryPeerCfg) AsMapInterface() map[string]any {
	return map[string]any{
		"URL":          sp.Url,
		"ClientSecret": sp.ClientSecret,
		"ClientID":     sp.ClientID,
	}
}

func (sp *SentryPeerCfg) Clone() (cln *SentryPeerCfg) {
	cln = &SentryPeerCfg{
		Url:          sp.Url,
		ClientSecret: sp.ClientSecret,
		ClientID:     sp.ClientID,
	}
	return
}
