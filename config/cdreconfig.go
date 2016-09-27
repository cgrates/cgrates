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

// One instance of CdrExporter
type CdreConfig struct {
	CdrFormat                  string
	FieldSeparator             rune
	DataUsageMultiplyFactor    float64
	SMSUsageMultiplyFactor     float64
	MMSUsageMultiplyFactor     float64
	GenericUsageMultiplyFactor float64
	CostMultiplyFactor         float64
	CostRoundingDecimals       int
	CostShiftDigits            int
	MaskDestinationID          string
	MaskLength                 int
	ExportDirectory            string
	HeaderFields               []*CfgCdrField
	ContentFields              []*CfgCdrField
	TrailerFields              []*CfgCdrField
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
	if jsnCfg.Sms_usage_multiply_factor != nil {
		self.SMSUsageMultiplyFactor = *jsnCfg.Sms_usage_multiply_factor
	}
	if jsnCfg.Mms_usage_multiply_factor != nil {
		self.MMSUsageMultiplyFactor = *jsnCfg.Mms_usage_multiply_factor
	}
	if jsnCfg.Generic_usage_multiply_factor != nil {
		self.GenericUsageMultiplyFactor = *jsnCfg.Generic_usage_multiply_factor
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
		self.MaskDestinationID = *jsnCfg.Mask_destination_id
	}
	if jsnCfg.Mask_length != nil {
		self.MaskLength = *jsnCfg.Mask_length
	}
	if jsnCfg.Export_directory != nil {
		self.ExportDirectory = *jsnCfg.Export_directory
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

// Clone itself into a new CdreConfig
func (self *CdreConfig) Clone() *CdreConfig {
	clnCdre := new(CdreConfig)
	clnCdre.CdrFormat = self.CdrFormat
	clnCdre.FieldSeparator = self.FieldSeparator
	clnCdre.DataUsageMultiplyFactor = self.DataUsageMultiplyFactor
	clnCdre.SMSUsageMultiplyFactor = self.SMSUsageMultiplyFactor
	clnCdre.MMSUsageMultiplyFactor = self.MMSUsageMultiplyFactor
	clnCdre.GenericUsageMultiplyFactor = self.GenericUsageMultiplyFactor
	clnCdre.CostMultiplyFactor = self.CostMultiplyFactor
	clnCdre.CostRoundingDecimals = self.CostRoundingDecimals
	clnCdre.CostShiftDigits = self.CostShiftDigits
	clnCdre.MaskDestinationID = self.MaskDestinationID
	clnCdre.MaskLength = self.MaskLength
	clnCdre.ExportDirectory = self.ExportDirectory
	clnCdre.HeaderFields = make([]*CfgCdrField, len(self.HeaderFields))
	for idx, fld := range self.HeaderFields {
		clonedVal := *fld
		clnCdre.HeaderFields[idx] = &clonedVal
	}
	clnCdre.ContentFields = make([]*CfgCdrField, len(self.ContentFields))
	for idx, fld := range self.ContentFields {
		clonedVal := *fld
		clnCdre.ContentFields[idx] = &clonedVal
	}
	clnCdre.TrailerFields = make([]*CfgCdrField, len(self.TrailerFields))
	for idx, fld := range self.TrailerFields {
		clonedVal := *fld
		clnCdre.TrailerFields[idx] = &clonedVal
	}
	return clnCdre
}
