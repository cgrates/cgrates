/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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

// One instance of CdrExporter
type CdreConfig struct {
	CdrFormat               string
	FieldSeparator          rune
	DataUsageMultiplyFactor float64
	CostMultiplyFactor      float64
	CostRoundingDecimals    int
	CostShiftDigits         int
	MaskDestId              string
	MaskLength              int
	ExportDir               string
	HeaderFields            []*CfgCdrField
	ContentFields           []*CfgCdrField
	TrailerFields           []*CfgCdrField
}

func (self *CdreConfig) loadFromJsonCfg(jsnCfg *CdreJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Cdr_format != nil {
		self.CdrFormat = *jsnCfg.Cdr_format
	}
	if jsnCfg.Field_separator != nil && len(*jsnCfg.Field_separator) > 0 { // Make sure we got at least one character so we don't get panic here
		sepStr := *jsnCfg.Field_separator
		self.FieldSeparator = rune(sepStr[0])
	}
	if jsnCfg.Data_usage_multiply_factor != nil {
		self.DataUsageMultiplyFactor = *jsnCfg.Data_usage_multiply_factor
	}
	if jsnCfg.Cost_multiply_factor != nil {
		self.CostMultiplyFactor = *jsnCfg.Cost_multiply_factor
	}
	if jsnCfg.Cost_rounding_decimals != nil {
		self.CostRoundingDecimals = *jsnCfg.Cost_rounding_decimals
	}
	if jsnCfg.Cost_shift_digits != nil {
		self.CostShiftDigits = *jsnCfg.Cost_shift_digits
	}
	if jsnCfg.Mask_destination_id != nil {
		self.MaskDestId = *jsnCfg.Mask_destination_id
	}
	if jsnCfg.Mask_length != nil {
		self.MaskLength = *jsnCfg.Mask_length
	}
	if jsnCfg.Export_dir != nil {
		self.ExportDir = *jsnCfg.Export_dir
	}
	if jsnCfg.Header_fields != nil {
		if self.HeaderFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Header_fields); err != nil {
			return err
		}
	}
	if jsnCfg.Content_fields != nil {
		if self.ContentFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Content_fields); err != nil {
			return err
		}
	}
	if jsnCfg.Trailer_fields != nil {
		if self.TrailerFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Trailer_fields); err != nil {
			return err
		}
	}
	return nil
}
