// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import "github.com/cgrates/cgrates/utils"

// AttributeSCfg is the configuration of attribute service
type AnalyzerSCfg struct {
	Enabled bool
}

func (alS *AnalyzerSCfg) loadFromJsonCfg(jsnCfg *AnalyzerSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		alS.Enabled = *jsnCfg.Enabled
	}
	return nil
}

func (alS *AnalyzerSCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.EnabledCfg: alS.Enabled,
	}
}
