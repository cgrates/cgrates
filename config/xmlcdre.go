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
)

// The CdrExporter Fixed Width configuration instance
type CgrXmlCdreFwCfg struct {
	Header  *CgrXmlCfgCdrHeader  `xml:"header"`
	Content *CgrXmlCfgCdrContent `xml:"content"`
	Trailer *CgrXmlCfgCdrTrailer `xml:"trailer"`
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
	XMLName   xml.Name `xml:"field"`
	Name      string   `xml:"name,attr"`
	Type      string   `xml:"type,attr"`
	Value     string   `xml:"value,attr"`
	Width     int      `xml:"width,attr"`     // Field width
	Strip     string   `xml:"strip,attr"`     // Strip strategy in case value is bigger than field width <""|left|xleft|right|xright>
	Padding   string   `xml:"padding,attr"`   // Padding strategy in case of value is smaller than width <""left|zeroleft|right>
	Layout    string   `xml:"layout,attr"`    // Eg. time format layout
	Mandatory bool     `xml:"mandatory,attr"` // If field is mandatory, empty value will be considered as error and CDR will not be exported
}
