/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import "github.com/cgrates/cgrates/utils"

// APIBanCfg the config for the APIBan Keys
type APIBanCfg struct {
	Keys []string
}

func (ban *APIBanCfg) loadFromJSONCfg(jsnCfg *APIBanJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Keys != nil {
		ban.Keys = make([]string, len(*jsnCfg.Keys))
		copy(ban.Keys, *jsnCfg.Keys)
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (ban *APIBanCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.KeysCfg: ban.Keys,
	}
}

// Clone returns a deep copy of APIBanCfg
func (ban APIBanCfg) Clone() (cln *APIBanCfg) {
	cln = &APIBanCfg{
		Keys: make([]string, len(ban.Keys)),
	}
	copy(cln.Keys, ban.Keys)
	return
}
