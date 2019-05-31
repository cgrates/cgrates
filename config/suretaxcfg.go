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
