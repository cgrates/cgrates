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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// SureTaxCfg configuration object
type SureTaxCfg struct {
	URL                  string
	ClientNumber         string
	ValidationKey        string
	BusinessUnit         string
	Timezone             *time.Location // Convert the time of the events to this timezone before sending request out
	IncludeLocalCost     bool
	ReturnFileCode       string
	ResponseGroup        string
	ResponseType         string
	RegulatoryCode       string
	ClientTracking       utils.RSRParsers // Concatenate all of them to get value
	CustomerNumber       utils.RSRParsers
	OrigNumber           utils.RSRParsers
	TermNumber           utils.RSRParsers
	BillToNumber         utils.RSRParsers
	Zipcode              utils.RSRParsers
	Plus4                utils.RSRParsers
	P2PZipcode           utils.RSRParsers
	P2PPlus4             utils.RSRParsers
	Units                utils.RSRParsers
	UnitType             utils.RSRParsers
	TaxIncluded          utils.RSRParsers
	TaxSitusRule         utils.RSRParsers
	TransTypeCode        utils.RSRParsers
	SalesTypeCode        utils.RSRParsers
	TaxExemptionCodeList utils.RSRParsers
}

// loadSureTaxCfg loads the SureTax section of the configuration
func (st *SureTaxCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnSureTaxCfg := new(SureTaxJsonCfg)
	if err = jsnCfg.GetSection(ctx, SureTaxJSON, jsnSureTaxCfg); err != nil {
		return
	}
	return st.loadFromJSONCfg(jsnSureTaxCfg)
}

// Loads/re-loads data from json config object
func (st *SureTaxCfg) loadFromJSONCfg(jsnCfg *SureTaxJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Url != nil {
		st.URL = *jsnCfg.Url
	}
	if jsnCfg.Client_number != nil {
		st.ClientNumber = *jsnCfg.Client_number
	}
	if jsnCfg.Validation_key != nil {
		st.ValidationKey = *jsnCfg.Validation_key
	}
	if jsnCfg.Business_unit != nil {
		st.BusinessUnit = *jsnCfg.Business_unit
	}
	if jsnCfg.Timezone != nil {
		if st.Timezone, err = time.LoadLocation(*jsnCfg.Timezone); err != nil {
			return err
		}
	}
	if jsnCfg.Include_local_cost != nil {
		st.IncludeLocalCost = *jsnCfg.Include_local_cost
	}
	if jsnCfg.Return_file_code != nil {
		st.ReturnFileCode = *jsnCfg.Return_file_code
	}
	if jsnCfg.Response_group != nil {
		st.ResponseGroup = *jsnCfg.Response_group
	}
	if jsnCfg.Response_type != nil {
		st.ResponseType = *jsnCfg.Response_type
	}
	if jsnCfg.Regulatory_code != nil {
		st.RegulatoryCode = *jsnCfg.Regulatory_code
	}
	if jsnCfg.Client_tracking != nil {
		if st.ClientTracking, err = utils.NewRSRParsers(*jsnCfg.Client_tracking, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Customer_number != nil {
		if st.CustomerNumber, err = utils.NewRSRParsers(*jsnCfg.Customer_number, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Orig_number != nil {
		if st.OrigNumber, err = utils.NewRSRParsers(*jsnCfg.Orig_number, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Term_number != nil {
		if st.TermNumber, err = utils.NewRSRParsers(*jsnCfg.Term_number, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Bill_to_number != nil {
		if st.BillToNumber, err = utils.NewRSRParsers(*jsnCfg.Bill_to_number, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Zipcode != nil {
		if st.Zipcode, err = utils.NewRSRParsers(*jsnCfg.Zipcode, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Plus4 != nil {
		if st.Plus4, err = utils.NewRSRParsers(*jsnCfg.Plus4, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.P2PZipcode != nil {
		if st.P2PZipcode, err = utils.NewRSRParsers(*jsnCfg.P2PZipcode, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.P2PPlus4 != nil {
		if st.P2PPlus4, err = utils.NewRSRParsers(*jsnCfg.P2PPlus4, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Units != nil {
		if st.Units, err = utils.NewRSRParsers(*jsnCfg.Units, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Unit_type != nil {
		if st.UnitType, err = utils.NewRSRParsers(*jsnCfg.Unit_type, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_included != nil {
		if st.TaxIncluded, err = utils.NewRSRParsers(*jsnCfg.Tax_included, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_situs_rule != nil {
		if st.TaxSitusRule, err = utils.NewRSRParsers(*jsnCfg.Tax_situs_rule, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Trans_type_code != nil {
		if st.TransTypeCode, err = utils.NewRSRParsers(*jsnCfg.Trans_type_code, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Sales_type_code != nil {
		if st.SalesTypeCode, err = utils.NewRSRParsers(*jsnCfg.Sales_type_code, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_exemption_code_list != nil {
		if st.TaxExemptionCodeList, err = utils.NewRSRParsers(*jsnCfg.Tax_exemption_code_list, utils.InfieldSep); err != nil {
			return err
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (st SureTaxCfg) AsMapInterface() any {
	return map[string]any{
		utils.URLCfg:              st.URL,
		utils.ClientNumberCfg:     st.ClientNumber,
		utils.ValidationKeyCfg:    st.ValidationKey,
		utils.BusinessUnitCfg:     st.BusinessUnit,
		utils.TimezoneCfg:         st.Timezone.String(),
		utils.IncludeLocalCostCfg: st.IncludeLocalCost,
		utils.ReturnFileCodeCfg:   st.ReturnFileCode,
		utils.ResponseGroupCfg:    st.ResponseGroup,
		utils.ResponseTypeCfg:     st.ResponseType,
		utils.RegulatoryCodeCfg:   st.RegulatoryCode,

		utils.ClientTrackingCfg:       st.ClientTracking.GetRule(),
		utils.CustomerNumberCfg:       st.CustomerNumber.GetRule(),
		utils.OrigNumberCfg:           st.OrigNumber.GetRule(),
		utils.TermNumberCfg:           st.TermNumber.GetRule(),
		utils.BillToNumberCfg:         st.BillToNumber.GetRule(),
		utils.ZipcodeCfg:              st.Zipcode.GetRule(),
		utils.Plus4Cfg:                st.Plus4.GetRule(),
		utils.P2PZipcodeCfg:           st.P2PZipcode.GetRule(),
		utils.P2PPlus4Cfg:             st.P2PPlus4.GetRule(),
		utils.UnitsCfg:                st.Units.GetRule(),
		utils.UnitTypeCfg:             st.UnitType.GetRule(),
		utils.TaxIncludedCfg:          st.TaxIncluded.GetRule(),
		utils.TaxSitusRuleCfg:         st.TaxSitusRule.GetRule(),
		utils.TransTypeCodeCfg:        st.TransTypeCode.GetRule(),
		utils.SalesTypeCodeCfg:        st.SalesTypeCode.GetRule(),
		utils.TaxExemptionCodeListCfg: st.TaxExemptionCodeList.GetRule(),
	}
}

func (SureTaxCfg) SName() string            { return SureTaxJSON }
func (st SureTaxCfg) CloneSection() Section { return st.Clone() }

// Clone returns a deep copy of SureTaxCfg
func (st SureTaxCfg) Clone() *SureTaxCfg {
	loc := *time.UTC
	if st.Timezone != nil {
		loc = *st.Timezone
	}
	return &SureTaxCfg{
		URL:              st.URL,
		ClientNumber:     st.ClientNumber,
		ValidationKey:    st.ValidationKey,
		BusinessUnit:     st.BusinessUnit,
		Timezone:         &loc,
		IncludeLocalCost: st.IncludeLocalCost,
		ReturnFileCode:   st.ReturnFileCode,
		ResponseGroup:    st.ResponseGroup,
		ResponseType:     st.ResponseType,
		RegulatoryCode:   st.RegulatoryCode,

		ClientTracking:       st.ClientTracking.Clone(),
		CustomerNumber:       st.CustomerNumber.Clone(),
		OrigNumber:           st.OrigNumber.Clone(),
		TermNumber:           st.TermNumber.Clone(),
		BillToNumber:         st.BillToNumber.Clone(),
		Zipcode:              st.Zipcode.Clone(),
		Plus4:                st.Plus4.Clone(),
		P2PZipcode:           st.P2PZipcode.Clone(),
		P2PPlus4:             st.P2PPlus4.Clone(),
		Units:                st.Units.Clone(),
		UnitType:             st.UnitType.Clone(),
		TaxIncluded:          st.TaxIncluded.Clone(),
		TaxSitusRule:         st.TaxSitusRule.Clone(),
		TransTypeCode:        st.TransTypeCode.Clone(),
		SalesTypeCode:        st.SalesTypeCode.Clone(),
		TaxExemptionCodeList: st.TaxExemptionCodeList.Clone(),
	}
}

// SureTax config section
type SureTaxJsonCfg struct {
	Url                     *string
	Client_number           *string
	Validation_key          *string
	Business_unit           *string
	Timezone                *string
	Include_local_cost      *bool
	Return_file_code        *string
	Response_group          *string
	Response_type           *string
	Regulatory_code         *string
	Client_tracking         *string
	Customer_number         *string
	Orig_number             *string
	Term_number             *string
	Bill_to_number          *string
	Zipcode                 *string
	Plus4                   *string
	P2PZipcode              *string
	P2PPlus4                *string
	Units                   *string
	Unit_type               *string
	Tax_included            *string
	Tax_situs_rule          *string
	Trans_type_code         *string
	Sales_type_code         *string
	Tax_exemption_code_list *string
}

func diffSureTaxJsonCfg(d *SureTaxJsonCfg, v1, v2 *SureTaxCfg) *SureTaxJsonCfg {
	if d == nil {
		d = new(SureTaxJsonCfg)
	}
	if v1.URL != v2.URL {
		d.Url = utils.StringPointer(v2.URL)
	}
	if v1.ClientNumber != v2.ClientNumber {
		d.Client_number = utils.StringPointer(v2.ClientNumber)
	}
	if v1.ValidationKey != v2.ValidationKey {
		d.Validation_key = utils.StringPointer(v2.ValidationKey)
	}
	if v1.BusinessUnit != v2.BusinessUnit {
		d.Business_unit = utils.StringPointer(v2.BusinessUnit)
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone.String())
	}
	if v1.IncludeLocalCost != v2.IncludeLocalCost {
		d.Include_local_cost = utils.BoolPointer(v2.IncludeLocalCost)
	}
	if v1.ReturnFileCode != v2.ReturnFileCode {
		d.Return_file_code = utils.StringPointer(v2.ReturnFileCode)
	}
	if v1.ResponseGroup != v2.ResponseGroup {
		d.Response_group = utils.StringPointer(v2.ResponseGroup)
	}
	if v1.ResponseType != v2.ResponseType {
		d.Response_type = utils.StringPointer(v2.ResponseType)
	}
	if v1.RegulatoryCode != v2.RegulatoryCode {
		d.Regulatory_code = utils.StringPointer(v2.RegulatoryCode)
	}
	rs1 := v1.ClientTracking.GetRule()
	rs2 := v2.ClientTracking.GetRule()
	if rs1 != rs2 {
		d.Client_tracking = utils.StringPointer(rs2)
	}
	rs1 = v1.CustomerNumber.GetRule()
	rs2 = v2.CustomerNumber.GetRule()
	if rs1 != rs2 {
		d.Customer_number = utils.StringPointer(rs2)
	}
	rs1 = v1.OrigNumber.GetRule()
	rs2 = v2.OrigNumber.GetRule()
	if rs1 != rs2 {
		d.Orig_number = utils.StringPointer(rs2)
	}
	rs1 = v1.TermNumber.GetRule()
	rs2 = v2.TermNumber.GetRule()
	if rs1 != rs2 {
		d.Term_number = utils.StringPointer(rs2)
	}
	rs1 = v1.BillToNumber.GetRule()
	rs2 = v2.BillToNumber.GetRule()
	if rs1 != rs2 {
		d.Bill_to_number = utils.StringPointer(rs2)
	}
	rs1 = v1.Zipcode.GetRule()
	rs2 = v2.Zipcode.GetRule()
	if rs1 != rs2 {
		d.Zipcode = utils.StringPointer(rs2)
	}
	rs1 = v1.Plus4.GetRule()
	rs2 = v2.Plus4.GetRule()
	if rs1 != rs2 {
		d.Plus4 = utils.StringPointer(rs2)
	}
	rs1 = v1.P2PZipcode.GetRule()
	rs2 = v2.P2PZipcode.GetRule()
	if rs1 != rs2 {
		d.P2PZipcode = utils.StringPointer(rs2)
	}
	rs1 = v1.P2PPlus4.GetRule()
	rs2 = v2.P2PPlus4.GetRule()
	if rs1 != rs2 {
		d.P2PPlus4 = utils.StringPointer(rs2)
	}
	rs1 = v1.Units.GetRule()
	rs2 = v2.Units.GetRule()
	if rs1 != rs2 {
		d.Units = utils.StringPointer(rs2)
	}
	rs1 = v1.UnitType.GetRule()
	rs2 = v2.UnitType.GetRule()
	if rs1 != rs2 {
		d.Unit_type = utils.StringPointer(rs2)
	}
	rs1 = v1.TaxIncluded.GetRule()
	rs2 = v2.TaxIncluded.GetRule()
	if rs1 != rs2 {
		d.Tax_included = utils.StringPointer(rs2)
	}
	rs1 = v1.TaxSitusRule.GetRule()
	rs2 = v2.TaxSitusRule.GetRule()
	if rs1 != rs2 {
		d.Tax_situs_rule = utils.StringPointer(rs2)
	}
	rs1 = v1.TransTypeCode.GetRule()
	rs2 = v2.TransTypeCode.GetRule()
	if rs1 != rs2 {
		d.Trans_type_code = utils.StringPointer(rs2)
	}
	rs1 = v1.SalesTypeCode.GetRule()
	rs2 = v2.SalesTypeCode.GetRule()
	if rs1 != rs2 {
		d.Sales_type_code = utils.StringPointer(rs2)
	}
	rs1 = v1.TaxExemptionCodeList.GetRule()
	rs2 = v2.TaxExemptionCodeList.GetRule()
	if rs1 != rs2 {
		d.Tax_exemption_code_list = utils.StringPointer(rs2)
	}
	return d
}
