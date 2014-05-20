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

import (
	"encoding/xml"
	"github.com/cgrates/cgrates/utils"
)

type CgrXmlCdrcCfg struct {
	Enabled     bool        `xml:"enabled"`       // Enable/Disable the
	CdrsAddress string      `xml:"cdrs_address"`  // The address where CDRs can be reached
	CdrsMethod  string      `xml:"cdrs_method"`   // Method to use when posting CDRs
	CdrType     string      `xml:"cdr_type"`      // The type of CDR to process <csv>
	RunDelay    int64       `xml:"run_delay"`     // Delay between runs
	CdrInDir    string      `xml:"cdr_in_dir"`    // Folder to process CDRs from
	CdrOutDir   string      `xml:"cdr_out_dir"`   // Folder to move processed CDRs to
	CdrSourceId string      `xml:"cdr_source_id"` // Source identifier for the processed CDRs
	CdrFields   []CdrcField `xml:"fields>field"`
}

type CdrcField struct {
	XMLName  xml.Name `xml:"field"`
	Id       string   `xml:"id,attr"`
	Filter   string   `xml:"filter,attr"`
	RSRField *utils.RSRField
}

func (cdrcFld *CdrcField) PopulateRSRFIeld() (err error) {
	if cdrcFld.RSRField, err = utils.NewRSRField(cdrcFld.Filter); err != nil {
		return err
	}
	return nil
}
