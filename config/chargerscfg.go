// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// SupplierSCfg is the configuration of supplier service
type ChargerSCfg struct {
	Enabled             bool
	IndexedSelects      bool
	AttributeSConns     []string
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	NestedFields        bool
}

func (cS *ChargerSCfg) loadFromJsonCfg(jsnCfg *ChargerSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		cS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		cS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Attributes_conns != nil {
		cS.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, attrConn := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if attrConn == utils.MetaInternal {
				cS.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			} else {
				cS.AttributeSConns[idx] = attrConn
			}
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
	if jsnCfg.Nested_fields != nil {
		cS.NestedFields = *jsnCfg.Nested_fields
	}
	return
}

func (cS *ChargerSCfg) AsMapInterface() map[string]any {
	attributeSConns := make([]string, len(cS.AttributeSConns))
	for i, item := range cS.AttributeSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
		if item == buf {
			attributeSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaAttributes, utils.EmptyString)
		} else {
			attributeSConns[i] = item
		}
	}
	stringIndexedFields := []string{}
	if cS.StringIndexedFields != nil {
		stringIndexedFields = make([]string, len(*cS.StringIndexedFields))
		for i, item := range *cS.StringIndexedFields {
			stringIndexedFields[i] = item
		}
	}
	prefixIndexedFields := []string{}
	if cS.PrefixIndexedFields != nil {
		prefixIndexedFields = make([]string, len(*cS.PrefixIndexedFields))
		for i, item := range *cS.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
	}
	return map[string]any{
		utils.EnabledCfg:             cS.Enabled,
		utils.IndexedSelectsCfg:      cS.IndexedSelects,
		utils.AttributeSConnsCfg:     attributeSConns,
		utils.StringIndexedFieldsCfg: stringIndexedFields,
		utils.PrefixIndexedFieldsCfg: prefixIndexedFields,
		utils.NestedFieldsCfg:        cS.NestedFields,
	}
}
