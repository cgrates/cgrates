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

import "github.com/cgrates/cgrates/utils"

// SupplierSCfg is the configuration of supplier service
type SupplierSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	AttributeSConns     []string
	ResourceSConns      []string
	StatSConns          []string
	DefaultRatio        int
}

func (spl *SupplierSCfg) loadFromJsonCfg(jsnCfg *SupplierSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		spl.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		spl.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		spl.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		spl.PrefixIndexedFields = &pif
	}
	if jsnCfg.Attributes_conns != nil {
		spl.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, conn := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				spl.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			} else {
				spl.AttributeSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Resources_conns != nil {
		spl.ResourceSConns = make([]string, len(*jsnCfg.Resources_conns))
		for idx, conn := range *jsnCfg.Resources_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				spl.ResourceSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
			} else {
				spl.ResourceSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Stats_conns != nil {
		spl.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, conn := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if conn == utils.MetaInternal {
				spl.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			} else {
				spl.StatSConns[idx] = conn
			}
		}
	}
	if jsnCfg.Default_ratio != nil {
		spl.DefaultRatio = *jsnCfg.Default_ratio
	}
	return nil
}
