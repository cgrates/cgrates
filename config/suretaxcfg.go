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
		if st.ClientTracking, err = NewRSRParsers(*jsnCfg.Client_tracking, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Customer_number != nil {
		if st.CustomerNumber, err = NewRSRParsers(*jsnCfg.Customer_number, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Orig_number != nil {
		if st.OrigNumber, err = NewRSRParsers(*jsnCfg.Orig_number, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Term_number != nil {
		if st.TermNumber, err = NewRSRParsers(*jsnCfg.Term_number, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Bill_to_number != nil {
		if st.BillToNumber, err = NewRSRParsers(*jsnCfg.Bill_to_number, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Zipcode != nil {
		if st.Zipcode, err = NewRSRParsers(*jsnCfg.Zipcode, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Plus4 != nil {
		if st.Plus4, err = NewRSRParsers(*jsnCfg.Plus4, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.P2PZipcode != nil {
		if st.P2PZipcode, err = NewRSRParsers(*jsnCfg.P2PZipcode, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.P2PPlus4 != nil {
		if st.P2PPlus4, err = NewRSRParsers(*jsnCfg.P2PPlus4, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Units != nil {
		if st.Units, err = NewRSRParsers(*jsnCfg.Units, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Unit_type != nil {
		if st.UnitType, err = NewRSRParsers(*jsnCfg.Unit_type, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_included != nil {
		if st.TaxIncluded, err = NewRSRParsers(*jsnCfg.Tax_included, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_situs_rule != nil {
		if st.TaxSitusRule, err = NewRSRParsers(*jsnCfg.Tax_situs_rule, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Trans_type_code != nil {
		if st.TransTypeCode, err = NewRSRParsers(*jsnCfg.Trans_type_code, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Sales_type_code != nil {
		if st.SalesTypeCode, err = NewRSRParsers(*jsnCfg.Sales_type_code, utils.InfieldSep); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_exemption_code_list != nil {
		if st.TaxExemptionCodeList, err = NewRSRParsers(*jsnCfg.Tax_exemption_code_list, utils.InfieldSep); err != nil {
			return err
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (st *SureTaxCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	return map[string]interface{}{
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

		utils.ClientTrackingCfg:       st.ClientTracking.GetRule(separator),
		utils.CustomerNumberCfg:       st.CustomerNumber.GetRule(separator),
		utils.OrigNumberCfg:           st.OrigNumber.GetRule(separator),
		utils.TermNumberCfg:           st.TermNumber.GetRule(separator),
		utils.BillToNumberCfg:         st.BillToNumber.GetRule(separator),
		utils.ZipcodeCfg:              st.Zipcode.GetRule(separator),
		utils.Plus4Cfg:                st.Plus4.GetRule(separator),
		utils.P2PZipcodeCfg:           st.P2PZipcode.GetRule(separator),
		utils.P2PPlus4Cfg:             st.P2PPlus4.GetRule(separator),
		utils.UnitsCfg:                st.Units.GetRule(separator),
		utils.UnitTypeCfg:             st.UnitType.GetRule(separator),
		utils.TaxIncludedCfg:          st.TaxIncluded.GetRule(separator),
		utils.TaxSitusRuleCfg:         st.TaxSitusRule.GetRule(separator),
		utils.TransTypeCodeCfg:        st.TransTypeCode.GetRule(separator),
		utils.SalesTypeCodeCfg:        st.SalesTypeCode.GetRule(separator),
		utils.TaxExemptionCodeListCfg: st.TaxExemptionCodeList.GetRule(separator),
	}
}

// Clone returns a deep copy of SureTaxCfg
func (st SureTaxCfg) Clone() (cln *SureTaxCfg) {
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
