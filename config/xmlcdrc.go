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
	Enabled      bool         `xml:"enabled"`         // Enable/Disable the
	CdrsAddress  string       `xml:"cdrs_address"`    // The address where CDRs can be reached
	CdrType      string       `xml:"cdr_type"`        // The type of CDR to process <csv>
	CsvSeparator string       `xml:"field_separator"` // The separator to use when reading csvs
	RunDelay     int64        `xml:"run_delay"`       // Delay between runs
	CdrInDir     string       `xml:"cdr_in_dir"`      // Folder to process CDRs from
	CdrOutDir    string       `xml:"cdr_out_dir"`     // Folder to move processed CDRs to
	CdrSourceId  string       `xml:"cdr_source_id"`   // Source identifier for the processed CDRs
	CdrFields    []*CdrcField `xml:"fields>field"`
}

// Set the defaults
func (cdrcCfg *CgrXmlCdrcCfg) setDefaults() error {
	dfCfg, _ := NewDefaultCGRConfig()
	if len(cdrcCfg.CdrsAddress) == 0 {
		cdrcCfg.CdrsAddress = dfCfg.CdrcCdrs
	}
	if len(cdrcCfg.CdrType) == 0 {
		cdrcCfg.CdrType = dfCfg.CdrcCdrType
	}
	if len(cdrcCfg.CsvSeparator) == 0 {
		cdrcCfg.CsvSeparator = dfCfg.CdrcCsvSep
	}
	if len(cdrcCfg.CdrInDir) == 0 {
		cdrcCfg.CdrInDir = dfCfg.CdrcCdrInDir
	}
	if len(cdrcCfg.CdrOutDir) == 0 {
		cdrcCfg.CdrOutDir = dfCfg.CdrcCdrOutDir
	}
	if len(cdrcCfg.CdrSourceId) == 0 {
		cdrcCfg.CdrSourceId = dfCfg.CdrcSourceId
	}
	if len(cdrcCfg.CdrFields) == 0 {
		for key, cfgRsrFields := range dfCfg.CdrcCdrFields {
			cdrcCfg.CdrFields = append(cdrcCfg.CdrFields, &CdrcField{Id: key, Value: "PLACEHOLDER", rsrFields: cfgRsrFields})
		}
	}
	return nil
}

func (cdrcCfg *CgrXmlCdrcCfg) CdrRSRFields() map[string][]*utils.RSRField {
	rsrFields := make(map[string][]*utils.RSRField)
	for _, fld := range cdrcCfg.CdrFields {
		rsrFields[fld.Id] = fld.rsrFields
	}
	return rsrFields
}

type CdrcField struct {
	XMLName   xml.Name `xml:"field"`
	Id        string   `xml:"id,attr"`
	Value     string   `xml:"value,attr"`
	rsrFields []*utils.RSRField
}

func (cdrcFld *CdrcField) PopulateRSRFields() (err error) {
	if cdrcFld.rsrFields, err = utils.ParseRSRFields(cdrcFld.Value, utils.INFIELD_SEP); err != nil {
		return err
	}
	return nil
}
