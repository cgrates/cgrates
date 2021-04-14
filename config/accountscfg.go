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

// AccountSCfg is the configuration of ActionS
type AccountSCfg struct {
	Enabled             bool
	AttributeSConns     []string
	RateSConns          []string
	ThresholdSConns     []string
	IndexedSelects      bool
	StringIndexedFields *[]string
	PrefixIndexedFields *[]string
	SuffixIndexedFields *[]string
	NestedFields        bool
	MaxIterations       int
	MaxUsage            *utils.Decimal
}

func (acS *AccountSCfg) loadFromJSONCfg(jsnCfg *AccountSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		acS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Indexed_selects != nil {
		acS.IndexedSelects = *jsnCfg.Indexed_selects
	}
	if jsnCfg.Attributes_conns != nil {
		acS.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, conn := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			acS.AttributeSConns[idx] = conn
			if conn == utils.MetaInternal {
				acS.AttributeSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	if jsnCfg.Rates_conns != nil {
		acS.RateSConns = make([]string, len(*jsnCfg.Rates_conns))
		for idx, conn := range *jsnCfg.Rates_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			acS.RateSConns[idx] = conn
			if conn == utils.MetaInternal {
				acS.RateSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS)
			}
		}
	}
	if jsnCfg.Thresholds_conns != nil {
		acS.ThresholdSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, conn := range *jsnCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			acS.ThresholdSConns[idx] = conn
			if conn == utils.MetaInternal {
				acS.ThresholdSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	if jsnCfg.String_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.String_indexed_fields))
		for i, fID := range *jsnCfg.String_indexed_fields {
			sif[i] = fID
		}
		acS.StringIndexedFields = &sif
	}
	if jsnCfg.Prefix_indexed_fields != nil {
		pif := make([]string, len(*jsnCfg.Prefix_indexed_fields))
		for i, fID := range *jsnCfg.Prefix_indexed_fields {
			pif[i] = fID
		}
		acS.PrefixIndexedFields = &pif
	}
	if jsnCfg.Suffix_indexed_fields != nil {
		sif := make([]string, len(*jsnCfg.Suffix_indexed_fields))
		for i, fID := range *jsnCfg.Suffix_indexed_fields {
			sif[i] = fID
		}
		acS.SuffixIndexedFields = &sif
	}
	if jsnCfg.Nested_fields != nil {
		acS.NestedFields = *jsnCfg.Nested_fields
	}
	if jsnCfg.Max_iterations != nil {
		acS.MaxIterations = *jsnCfg.Max_iterations
	}
	if jsnCfg.Max_usage != nil {
		if acS.MaxUsage, err = utils.NewDecimalFromUsage(*jsnCfg.Max_usage); err != nil {
			return err
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (acS *AccountSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:        acS.Enabled,
		utils.IndexedSelectsCfg: acS.IndexedSelects,
		utils.NestedFieldsCfg:   acS.NestedFields,
		utils.MaxIterations:     acS.MaxIterations,
	}
	if acS.AttributeSConns != nil {
		attributeSConns := make([]string, len(acS.AttributeSConns))
		for i, item := range acS.AttributeSConns {
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.AttributeSConnsCfg] = attributeSConns
	}
	if acS.RateSConns != nil {
		rateSConns := make([]string, len(acS.RateSConns))
		for i, item := range acS.RateSConns {
			rateSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS) {
				rateSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.RateSConnsCfg] = rateSConns
	}
	if acS.ThresholdSConns != nil {
		thresholdSConns := make([]string, len(acS.ThresholdSConns))
		for i, item := range acS.ThresholdSConns {
			thresholdSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				thresholdSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = thresholdSConns
	}
	if acS.StringIndexedFields != nil {
		stringIndexedFields := make([]string, len(*acS.StringIndexedFields))
		for i, item := range *acS.StringIndexedFields {
			stringIndexedFields[i] = item
		}
		initialMP[utils.StringIndexedFieldsCfg] = stringIndexedFields
	}
	if acS.PrefixIndexedFields != nil {
		prefixIndexedFields := make([]string, len(*acS.PrefixIndexedFields))
		for i, item := range *acS.PrefixIndexedFields {
			prefixIndexedFields[i] = item
		}
		initialMP[utils.PrefixIndexedFieldsCfg] = prefixIndexedFields
	}
	if acS.SuffixIndexedFields != nil {
		suffixIndexedFields := make([]string, len(*acS.SuffixIndexedFields))
		for i, item := range *acS.SuffixIndexedFields {
			suffixIndexedFields[i] = item
		}
		initialMP[utils.SuffixIndexedFieldsCfg] = suffixIndexedFields
	}
	if acS.MaxUsage != nil {
		initialMP[utils.MaxUsage] = acS.MaxUsage.String()
	}
	return
}

// Clone returns a deep copy of AccountSCfg
func (acS AccountSCfg) Clone() (cln *AccountSCfg) {
	cln = &AccountSCfg{
		Enabled:        acS.Enabled,
		IndexedSelects: acS.IndexedSelects,
		NestedFields:   acS.NestedFields,
		MaxIterations:  acS.MaxIterations,
		MaxUsage:       acS.MaxUsage,
	}
	if acS.AttributeSConns != nil {
		cln.AttributeSConns = make([]string, len(acS.AttributeSConns))
		for i, con := range acS.AttributeSConns {
			cln.AttributeSConns[i] = con
		}
	}
	if acS.RateSConns != nil {
		cln.RateSConns = make([]string, len(acS.RateSConns))
		for i, con := range acS.RateSConns {
			cln.RateSConns[i] = con
		}
	}
	if acS.ThresholdSConns != nil {
		cln.ThresholdSConns = make([]string, len(acS.ThresholdSConns))
		for i, con := range acS.ThresholdSConns {
			cln.ThresholdSConns[i] = con
		}
	}
	if acS.StringIndexedFields != nil {
		idx := make([]string, len(*acS.StringIndexedFields))
		for i, dx := range *acS.StringIndexedFields {
			idx[i] = dx
		}
		cln.StringIndexedFields = &idx
	}
	if acS.PrefixIndexedFields != nil {
		idx := make([]string, len(*acS.PrefixIndexedFields))
		for i, dx := range *acS.PrefixIndexedFields {
			idx[i] = dx
		}
		cln.PrefixIndexedFields = &idx
	}
	if acS.SuffixIndexedFields != nil {
		idx := make([]string, len(*acS.SuffixIndexedFields))
		for i, dx := range *acS.SuffixIndexedFields {
			idx[i] = dx
		}
		cln.SuffixIndexedFields = &idx
	}
	return
}
