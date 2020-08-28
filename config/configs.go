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

// ConfigSCfg config for listening over http
type ConfigSCfg struct {
	Enabled bool
	Listen  string
}

//loadFromJsonCfg loads Database config from JsonCfg
func (cScfg *ConfigSCfg) loadFromJsonCfg(jsnCfg *ConfigSCfgJson) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		cScfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen != nil {
		cScfg.Listen = *jsnCfg.Listen
	}
	return
}
