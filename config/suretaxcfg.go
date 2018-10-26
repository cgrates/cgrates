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
	ClientTracking       utils.RSRFields // Concatenate all of them to get value
	CustomerNumber       utils.RSRFields
	OrigNumber           utils.RSRFields
	TermNumber           utils.RSRFields
	BillToNumber         utils.RSRFields
	Zipcode              utils.RSRFields
	Plus4                utils.RSRFields
	P2PZipcode           utils.RSRFields
	P2PPlus4             utils.RSRFields
	Units                utils.RSRFields
	UnitType             utils.RSRFields
	TaxIncluded          utils.RSRFields
	TaxSitusRule         utils.RSRFields
	TransTypeCode        utils.RSRFields
	SalesTypeCode        utils.RSRFields
	TaxExemptionCodeList utils.RSRFields
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
		if self.ClientTracking, err = utils.ParseRSRFields(*jsnCfg.Client_tracking, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Customer_number != nil {
		if self.CustomerNumber, err = utils.ParseRSRFields(*jsnCfg.Customer_number, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Orig_number != nil {
		if self.OrigNumber, err = utils.ParseRSRFields(*jsnCfg.Orig_number, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Term_number != nil {
		if self.TermNumber, err = utils.ParseRSRFields(*jsnCfg.Term_number, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Bill_to_number != nil {
		if self.BillToNumber, err = utils.ParseRSRFields(*jsnCfg.Bill_to_number, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Zipcode != nil {
		if self.Zipcode, err = utils.ParseRSRFields(*jsnCfg.Zipcode, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Plus4 != nil {
		if self.Plus4, err = utils.ParseRSRFields(*jsnCfg.Plus4, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.P2PZipcode != nil {
		if self.P2PZipcode, err = utils.ParseRSRFields(*jsnCfg.P2PZipcode, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.P2PPlus4 != nil {
		if self.P2PPlus4, err = utils.ParseRSRFields(*jsnCfg.P2PPlus4, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Units != nil {
		if self.Units, err = utils.ParseRSRFields(*jsnCfg.Units, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Unit_type != nil {
		if self.UnitType, err = utils.ParseRSRFields(*jsnCfg.Unit_type, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_included != nil {
		if self.TaxIncluded, err = utils.ParseRSRFields(*jsnCfg.Tax_included, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_situs_rule != nil {
		if self.TaxSitusRule, err = utils.ParseRSRFields(*jsnCfg.Tax_situs_rule, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Trans_type_code != nil {
		if self.TransTypeCode, err = utils.ParseRSRFields(*jsnCfg.Trans_type_code, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Sales_type_code != nil {
		if self.SalesTypeCode, err = utils.ParseRSRFields(*jsnCfg.Sales_type_code, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Tax_exemption_code_list != nil {
		if self.TaxExemptionCodeList, err = utils.ParseRSRFields(*jsnCfg.Tax_exemption_code_list, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	return nil
}
