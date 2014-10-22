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
	"fmt"
	"github.com/cgrates/cgrates/utils"
	"io"
)

// Decodes a reader enforcing specific format of the configuration file
func ParseCgrXmlConfig(reader io.Reader) (*CgrXmlCfgDocument, error) {
	xmlConfig := new(CgrXmlCfgDocument)
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(xmlConfig); err != nil {
		return nil, err
	}
	if err := xmlConfig.cacheAll(); err != nil {
		return nil, err
	}
	return xmlConfig, nil
}

// XML CDR field, used for both cdrc and cdre
type XmlCfgCdrField struct {
	XMLName    xml.Name `xml:"field"`
	Tag        *string  `xml:"tag,attr"`
	Type       *string  `xml:"type,attr"`
	CdrFieldId *string  `xml:"cdr_field,attr"`
	Value      *string  `xml:"value,attr"`
	Width      *int     `xml:"width,attr"`     // Field width
	Strip      *string  `xml:"strip,attr"`     // Strip strategy in case value is bigger than field width <""|left|xleft|right|xright>
	Padding    *string  `xml:"padding,attr"`   // Padding strategy in case of value is smaller than width <""left|zeroleft|right>
	Layout     *string  `xml:"layout,attr"`    // Eg. time format layout
	Filter     *string  `xml:"filter,attr"`    // Eg. combimed filters
	Mandatory  *bool    `xml:"mandatory,attr"` // If field is mandatory, empty value will be considered as error and CDR will not be exported
}

// One CDRC Configuration instance
type CgrXmlCdrcCfg struct {
	Enabled                 *bool             `xml:"enabled"`                    // Enable/Disable the
	CdrsAddress             *string           `xml:"cdrs_address"`               // The address where CDRs can be reached
	CdrFormat               *string           `xml:"cdr_format"`                 // The type of CDR to process <csv>
	FieldSeparator          *string           `xml:"field_separator"`            // The separator to use when reading csvs
	DataUsageMultiplyFactor *int64            `xml:"data_usage_multiply_factor"` // Conversion factor for data usage
	RunDelay                *int64            `xml:"run_delay"`                  // Delay between runs
	CdrInDir                *string           `xml:"cdr_in_dir"`                 // Folder to process CDRs from
	CdrOutDir               *string           `xml:"cdr_out_dir"`                // Folder to move processed CDRs to
	CdrSourceId             *string           `xml:"cdr_source_id"`              // Source identifier for the processed CDRs
	CdrFields               []*XmlCfgCdrField `xml:"fields>field"`
}

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

// CDR header
type CgrXmlCfgCdrHeader struct {
	XMLName xml.Name          `xml:"header"`
	Fields  []*XmlCfgCdrField `xml:"fields>field"`
}

// CDR content
type CgrXmlCfgCdrContent struct {
	XMLName xml.Name          `xml:"content"`
	Fields  []*XmlCfgCdrField `xml:"fields>field"`
}

// CDR trailer
type CgrXmlCfgCdrTrailer struct {
	XMLName xml.Name          `xml:"trailer"`
	Fields  []*XmlCfgCdrField `xml:"fields>field"`
}

// Define a format for configuration file, one doc contains more configuration instances, identified by section, type and id
type CgrXmlCfgDocument struct {
	XMLName        xml.Name               `xml:"document"`
	Type           string                 `xml:"type,attr"`
	Configurations []*CgrXmlConfiguration `xml:"configuration"`
	cdrcs          map[string]*CgrXmlCdrcCfg
	cdres          map[string]*CgrXmlCdreCfg // Cahe cdrexporter instances, key will be the ID
}

// Storage for raw configuration
type CgrXmlConfiguration struct {
	XMLName   xml.Name `xml:"configuration"`
	Section   string   `xml:"section,attr"`
	Id        string   `xml:"id,attr"`
	RawConfig []byte   `xml:",innerxml"` // Used to store the configuration struct, as raw so we can store different types
}

func (cfgInst *CgrXmlConfiguration) rawConfigElement() []byte {
	rawConfig := append([]byte("<element>"), cfgInst.RawConfig...) // Encapsulate the rawConfig in one element so we can Unmarshall into one struct
	rawConfig = append(rawConfig, []byte("</element>")...)
	return rawConfig
}

func (xmlCfg *CgrXmlCfgDocument) cacheAll() error {
	for _, cacheFunc := range []func() error{xmlCfg.cacheCdrcCfgs, xmlCfg.cacheCdreCfgs} {
		if err := cacheFunc(); err != nil {
			return err
		}
	}
	return nil
}

// Avoid building from raw config string always, so build cache here
func (xmlCfg *CgrXmlCfgDocument) cacheCdrcCfgs() error {
	xmlCfg.cdrcs = make(map[string]*CgrXmlCdrcCfg)
	for _, cfgInst := range xmlCfg.Configurations {
		if cfgInst.Section != utils.CDRC {
			continue // Another type of config instance, not interesting to process
		}
		cdrcCfg := new(CgrXmlCdrcCfg)
		if err := xml.Unmarshal(cfgInst.rawConfigElement(), cdrcCfg); err != nil {
			return err
		} else if cdrcCfg == nil {
			return fmt.Errorf("Could not unmarshal config instance: %s", cfgInst.Id)
		}
		xmlCfg.cdrcs[cfgInst.Id] = cdrcCfg
	}
	return nil
}

// Avoid building from raw config string always, so build cache here
func (xmlCfg *CgrXmlCfgDocument) cacheCdreCfgs() error {
	xmlCfg.cdres = make(map[string]*CgrXmlCdreCfg)
	for _, cfgInst := range xmlCfg.Configurations {
		if cfgInst.Section != utils.CDRE {
			continue
		}
		cdreCfg := new(CgrXmlCdreCfg)
		if err := xml.Unmarshal(cfgInst.rawConfigElement(), cdreCfg); err != nil {
			return err
		} else if cdreCfg == nil {
			return fmt.Errorf("Could not unmarshal CgrXmlCdreCfg: %s", cfgInst.Id)
		}
		xmlCfg.cdres[cfgInst.Id] = cdreCfg
	}
	return nil
}

// Return instances or filtered instance of cdrefw configuration
func (xmlCfg *CgrXmlCfgDocument) GetCdreCfgs(instName string) map[string]*CgrXmlCdreCfg {
	if len(instName) != 0 {
		if cfg, hasIt := xmlCfg.cdres[instName]; !hasIt {
			return nil
		} else {
			return map[string]*CgrXmlCdreCfg{instName: cfg}
		}
	}
	return xmlCfg.cdres
}

// Return instances or filtered instance of cdrc configuration
func (xmlCfg *CgrXmlCfgDocument) GetCdrcCfgs(instName string) map[string]*CgrXmlCdrcCfg {
	if len(instName) != 0 {
		if cfg, hasIt := xmlCfg.cdrcs[instName]; !hasIt {
			return nil
		} else {
			return map[string]*CgrXmlCdrcCfg{instName: cfg} // Filtered
		}
	}
	return xmlCfg.cdrcs // Unfiltered
}
