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

// SupplierSCfg is the configuration of supplier service
type AliasSCfg struct {
	Enabled       bool
	IndexedFields []string
}

func (alS *AliasSCfg) loadFromJsonCfg(jsnCfg *AliasSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		alS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_fields != nil {
		alS.IndexedFields = make([]string, len(*jsnCfg.Indexed_fields))
		for i, fID := range *jsnCfg.Indexed_fields {
			alS.IndexedFields[i] = fID
		}
	}
	return
}
