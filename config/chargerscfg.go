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
type ChargerSCfg struct {
	Enabled             bool
	AttributeSConns     []*HaPoolConfig
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
}

func (cS *ChargerSCfg) loadFromJsonCfg(jsnCfg *ChargerSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		cS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Attributes_conns != nil {
		cS.AttributeSConns = make([]*HaPoolConfig, len(*jsnCfg.Attributes_conns))
		for idx, jsnHaCfg := range *jsnCfg.Attributes_conns {
			cS.AttributeSConns[idx] = NewDfltHaPoolConfig()
			cS.AttributeSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		cS.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		cS.PrefixIndexedFields = &pif
	}
	return
}
