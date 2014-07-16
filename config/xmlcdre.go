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

// The CdrExporter configuration instance
type CgrXmlCdreCfg struct {
	CdrFormat               *string              `xml:"cdr_format"`
	FieldSeparator          *string              `xml:"field_separator"`
	DataUsageMultiplyFactor *float64             `xml:"data_usage_multiply_factor"`
	CostMultiplyFactor      *float64             `xml:"cost_multiply_factor"`
	CostRoundingDecimals    *int                 `xml:"cost_rounding_decimals"`
	CostShiftDigits         *int                 `xml:"cost_shift_digits"`
	MaskDestId              *string              `xml:"mask_destination_id"`
	MaskLength              *int                 `xml:"mask_length"`
	ExportDir               *string              `xml:"export_dir"`
	Header                  *CgrXmlCfgCdrHeader  `xml:"export_template>header"`
	Content                 *CgrXmlCfgCdrContent `xml:"export_template>content"`
	Trailer                 *CgrXmlCfgCdrTrailer `xml:"export_template>trailer"`
}

func (xmlCdreCfg *CgrXmlCdreCfg) AsCdreConfig() *CdreConfig {
	cdreCfg, _ := NewDefaultCdreConfig()
	if xmlCdreCfg.CdrFormat != nil {
		cdreCfg.CdrFormat = *xmlCdreCfg.CdrFormat
	}
	if xmlCdreCfg.FieldSeparator != nil && len(*xmlCdreCfg.FieldSeparator) == 1 {
		sepStr := *xmlCdreCfg.FieldSeparator
		cdreCfg.FieldSeparator = rune(sepStr[0])
	}
	if xmlCdreCfg.DataUsageMultiplyFactor != nil {
		cdreCfg.DataUsageMultiplyFactor = *xmlCdreCfg.DataUsageMultiplyFactor
	}
	if xmlCdreCfg.CostMultiplyFactor != nil {
		cdreCfg.CostMultiplyFactor = *xmlCdreCfg.CostMultiplyFactor
	}
	if xmlCdreCfg.CostRoundingDecimals != nil {
		cdreCfg.CostRoundingDecimals = *xmlCdreCfg.CostRoundingDecimals
	}
	if xmlCdreCfg.CostShiftDigits != nil {
		cdreCfg.CostShiftDigits = *xmlCdreCfg.CostShiftDigits
	}
	if xmlCdreCfg.MaskDestId != nil {
		cdreCfg.MaskDestId = *xmlCdreCfg.MaskDestId
	}
	if xmlCdreCfg.MaskLength != nil {
		cdreCfg.MaskLength = *xmlCdreCfg.MaskLength
	}
	if xmlCdreCfg.ExportDir != nil {
		cdreCfg.ExportDir = *xmlCdreCfg.ExportDir
	}
	if xmlCdreCfg.Header != nil {
		cdreCfg.HeaderFields = make([]*CdreCdrField, len(xmlCdreCfg.Header.Fields))
		for idx, xmlFld := range xmlCdreCfg.Header.Fields {
			cdreCfg.HeaderFields[idx] = xmlFld.AsCdreCdrField()
		}
	}
	if xmlCdreCfg.Content != nil {
		cdreCfg.ContentFields = make([]*CdreCdrField, len(xmlCdreCfg.Content.Fields))
		for idx, xmlFld := range xmlCdreCfg.Content.Fields {
			cdreCfg.ContentFields[idx] = xmlFld.AsCdreCdrField()
		}
	}
	if xmlCdreCfg.Trailer != nil {
		cdreCfg.TrailerFields = make([]*CdreCdrField, len(xmlCdreCfg.Trailer.Fields))
		for idx, xmlFld := range xmlCdreCfg.Trailer.Fields {
			cdreCfg.TrailerFields[idx] = xmlFld.AsCdreCdrField()
		}
	}
	return cdreCfg
}

// CDR header
type CgrXmlCfgCdrHeader struct {
	XMLName xml.Name             `xml:"header"`
	Fields  []*CgrXmlCfgCdrField `xml:"fields>field"`
}

// CDR content
type CgrXmlCfgCdrContent struct {
	XMLName xml.Name             `xml:"content"`
	Fields  []*CgrXmlCfgCdrField `xml:"fields>field"`
}

// CDR trailer
type CgrXmlCfgCdrTrailer struct {
	XMLName xml.Name             `xml:"trailer"`
	Fields  []*CgrXmlCfgCdrField `xml:"fields>field"`
}

// CDR field
type CgrXmlCfgCdrField struct {
	XMLName          xml.Name        `xml:"field"`
	Name             string          `xml:"name,attr"`
	Type             string          `xml:"type,attr"`
	Value            string          `xml:"value,attr"`
	Width            int             `xml:"width,attr"`     // Field width
	Strip            string          `xml:"strip,attr"`     // Strip strategy in case value is bigger than field width <""|left|xleft|right|xright>
	Padding          string          `xml:"padding,attr"`   // Padding strategy in case of value is smaller than width <""left|zeroleft|right>
	Layout           string          `xml:"layout,attr"`    // Eg. time format layout
	Filter           string          `xml:"filter,attr"`    // Eg. combimed filters
	Mandatory        bool            `xml:"mandatory,attr"` // If field is mandatory, empty value will be considered as error and CDR will not be exported
	valueAsRsrField  *utils.RSRField // Cached if the need arrises
	filterAsRsrField *utils.RSRField
}

func (cdrFld *CgrXmlCfgCdrField) populateRSRField() (err error) {
	cdrFld.valueAsRsrField, err = utils.NewRSRField(cdrFld.Value)
	return err
}

func (cdrFld *CgrXmlCfgCdrField) populateFltrRSRField() (err error) {
	cdrFld.filterAsRsrField, err = utils.NewRSRField(cdrFld.Filter)
	return err
}

func (cdrFld *CgrXmlCfgCdrField) ValueAsRSRField() *utils.RSRField {
	return cdrFld.valueAsRsrField
}

func (cdrFld *CgrXmlCfgCdrField) AsCdreCdrField() *CdreCdrField {
	return &CdreCdrField{
		Name:            cdrFld.Name,
		Type:            cdrFld.Type,
		Value:           cdrFld.Value,
		Width:           cdrFld.Width,
		Strip:           cdrFld.Strip,
		Padding:         cdrFld.Padding,
		Layout:          cdrFld.Layout,
		Filter:          cdrFld.filterAsRsrField,
		Mandatory:       cdrFld.Mandatory,
		valueAsRsrField: cdrFld.valueAsRsrField,
	}
}
