// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
func (ban *APIBanCfg) Clone() (cln *APIBanCfg) {
	if ban == nil {
		return nil
	}
	cln = &APIBanCfg{
		Keys: make([]string, len(ban.Keys)),
	}
	copy(cln.Keys, ban.Keys)
	return
}
