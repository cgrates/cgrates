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

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

// SureTax configuration object
type SureTaxCfg struct {
	Url                  string
	ClientNumber         string
	ValidationKey        string
	BusinessUnit         string
	Timezone             *time.Location // Convert the time of the events to this timezone before sending request out
	IncludeLocalCost     bool
	ReturnFileCode       string
	ResponseGroup        string
	ResponseType         string
	RegulatoryCode       string
	ClientTracking       RSRParsers // Concatenate all of them to get value
	CustomerNumber       RSRParsers
	OrigNumber           RSRParsers
	TermNumber           RSRParsers
	BillToNumber         RSRParsers
	Zipcode              RSRParsers
	Plus4                RSRParsers
	P2PZipcode           RSRParsers
	P2PPlus4             RSRParsers
	Units                RSRParsers
	UnitType             RSRParsers
	TaxIncluded          RSRParsers
	TaxSitusRule         RSRParsers
	TransTypeCode        RSRParsers
	SalesTypeCode        RSRParsers
	TaxExemptionCodeList RSRParsers
}

// Loads/re-loads data from json config object
func (self *SureTaxCfg) loadFromJsonCfg(jsnCfg *SureTaxJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Url != nil {
		self.Url = *jsnCfg.Url
	}
	if jsnCfg.Client_number != nil {
		self.ClientNumber = *jsnCfg.Client_number
	}
	if jsnCfg.Validation_key != nil {
		self.ValidationKey = *jsnCfg.Validation_key
	}
	if jsnCfg.Business_unit != nil {
		self.BusinessUnit = *jsnCfg.Business_unit
	}
	if jsnCfg.Timezone != nil {
		if self.Timezone, err = time.LoadLocation(*jsnCfg.Timezone); err != nil {
			return err
		}
	}
	if jsnCfg.Include_local_cost != nil {
		self.IncludeLocalCost = *jsnCfg.Include_local_cost
	}
	if jsnCfg.Return_file_code != nil {
		self.ReturnFileCode = *jsnCfg.Return_file_code
	}
	if jsnCfg.Response_group != nil {
		self.ResponseGroup = *jsnCfg.Response_group
	}
	if jsnCfg.Response_type != nil {
		self.ResponseType = *jsnCfg.Response_type
	}
	if jsnCfg.Regulatory_code != nil {
		self.RegulatoryCode = *jsnCfg.Regulatory_code
	}
	if jsnCfg.Client_tracking != nil {
		if self.ClientTracking, err = NewRSRParsers(*jsnCfg.Client_tracking, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Customer_number != nil {
		if self.CustomerNumber, err = NewRSRParsers(*jsnCfg.Customer_number, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Orig_number != nil {
		if self.OrigNumber, err = NewRSRParsers(*jsnCfg.Orig_number, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Term_number != nil {
		if self.TermNumber, err = NewRSRParsers(*jsnCfg.Term_number, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Bill_to_number != nil {
		if self.BillToNumber, err = NewRSRParsers(*jsnCfg.Bill_to_number, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Zipcode != nil {
		if self.Zipcode, err = NewRSRParsers(*jsnCfg.Zipcode, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Plus4 != nil {
		if self.Plus4, err = NewRSRParsers(*jsnCfg.Plus4, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.P2PZipcode != nil {
		if self.P2PZipcode, err = NewRSRParsers(*jsnCfg.P2PZipcode, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.P2PPlus4 != nil {
		if self.P2PPlus4, err = NewRSRParsers(*jsnCfg.P2PPlus4, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Units != nil {
		if self.Units, err = NewRSRParsers(*jsnCfg.Units, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Unit_type != nil {
		if self.UnitType, err = NewRSRParsers(*jsnCfg.Unit_type, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_included != nil {
		if self.TaxIncluded, err = NewRSRParsers(*jsnCfg.Tax_included, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_situs_rule != nil {
		if self.TaxSitusRule, err = NewRSRParsers(*jsnCfg.Tax_situs_rule, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Trans_type_code != nil {
		if self.TransTypeCode, err = NewRSRParsers(*jsnCfg.Trans_type_code, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Sales_type_code != nil {
		if self.SalesTypeCode, err = NewRSRParsers(*jsnCfg.Sales_type_code, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_exemption_code_list != nil {
		if self.TaxExemptionCodeList, err = NewRSRParsers(*jsnCfg.Tax_exemption_code_list, true, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	return nil
}

func (st *SureTaxCfg) AsMapInterface() map[string]interface{} {

	clientTracking := make([]string, len(st.ClientTracking))
	for i, item := range st.ClientTracking {
		clientTracking[i] = item.Rules
	}

	customerNumber := make([]string, len(st.CustomerNumber))
	for i, item := range st.CustomerNumber {
		customerNumber[i] = item.Rules
	}

	origNumber := make([]string, len(st.OrigNumber))
	for i, item := range st.OrigNumber {
		origNumber[i] = item.Rules
	}

	termNumber := make([]string, len(st.TermNumber))
	for i, item := range st.TermNumber {
		termNumber[i] = item.Rules
	}

	billToNumber := make([]string, len(st.BillToNumber))
	for i, item := range st.BillToNumber {
		billToNumber[i] = item.Rules
	}

	zipcode := make([]string, len(st.Zipcode))
	for i, item := range st.Zipcode {
		zipcode[i] = item.Rules
	}

	plus4 := make([]string, len(st.Plus4))
	for i, item := range st.Plus4 {
		plus4[i] = item.Rules
	}

	p2PZipcode := make([]string, len(st.P2PZipcode))
	for i, item := range st.P2PZipcode {
		p2PZipcode[i] = item.Rules
	}

	p2PPlus4 := make([]string, len(st.P2PPlus4))
	for i, item := range st.P2PPlus4 {
		p2PPlus4[i] = item.Rules
	}

	units := make([]string, len(st.Units))
	for i, item := range st.Units {
		units[i] = item.Rules
	}

	unitType := make([]string, len(st.UnitType))
	for i, item := range st.UnitType {
		unitType[i] = item.Rules
	}

	taxIncluded := make([]string, len(st.TaxIncluded))
	for i, item := range st.TaxIncluded {
		taxIncluded[i] = item.Rules
	}

	taxSitusRule := make([]string, len(st.TaxSitusRule))
	for i, item := range st.TaxSitusRule {
		taxSitusRule[i] = item.Rules
	}

	transTypeCode := make([]string, len(st.TransTypeCode))
	for i, item := range st.TransTypeCode {
		transTypeCode[i] = item.Rules
	}

	salesTypeCode := make([]string, len(st.SalesTypeCode))
	for i, item := range st.SalesTypeCode {
		salesTypeCode[i] = item.Rules
	}

	taxExemptionCodeList := make([]string, len(st.TaxExemptionCodeList))
	for i, item := range st.TaxExemptionCodeList {
		taxExemptionCodeList[i] = item.Rules
	}

	return map[string]interface{}{
		utils.UrlCfg:                  st.Url,
		utils.ClientNumberCfg:         st.ClientNumber,
		utils.ValidationKeyCfg:        st.ValidationKey,
		utils.BusinessUnitCfg:         st.BusinessUnit,
		utils.TimezoneCfg:             st.Timezone,
		utils.IncludeLocalCostCfg:     st.IncludeLocalCost,
		utils.ReturnFileCodeCfg:       st.ReturnFileCode,
		utils.ResponseGroupCfg:        st.ResponseGroup,
		utils.ResponseTypeCfg:         st.ResponseType,
		utils.RegulatoryCodeCfg:       st.RegulatoryCode,
		utils.ClientTrackingCfg:       st.ClientTracking,
		utils.CustomerNumberCfg:       st.CustomerNumber,
		utils.OrigNumberCfg:           st.OrigNumber,
		utils.TermNumberCfg:           st.TermNumber,
		utils.BillToNumberCfg:         st.BillToNumber,
		utils.ZipcodeCfg:              st.Zipcode,
		utils.Plus4Cfg:                st.Plus4,
		utils.P2PZipcodeCfg:           st.P2PZipcode,
		utils.P2PPlus4Cfg:             st.P2PPlus4,
		utils.UnitsCfg:                st.Units,
		utils.UnitTypeCfg:             st.UnitType,
		utils.TaxIncludedCfg:          st.TaxIncluded,
		utils.TaxSitusRuleCfg:         st.TaxSitusRule,
		utils.TransTypeCodeCfg:        st.TransTypeCode,
		utils.SalesTypeCodeCfg:        st.SalesTypeCode,
		utils.TaxExemptionCodeListCfg: st.TaxExemptionCodeList,
	}

}
