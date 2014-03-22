/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"errors"
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
	return xmlConfig, nil
}

// Define a format for configuration file, one doc contains more configuration instances, identified by section, type and id
type CgrXmlCfgDocument struct {
	XMLName        xml.Name               `xml:"document"`
	Type           string                 `xml:"type,attr"`
	Configurations []*CgrXmlConfiguration `xml:"configuration"`
}

// Storage for raw configuration
type CgrXmlConfiguration struct {
	XMLName   xml.Name `xml:"configuration"`
	Section   string   `xml:"section,attr"`
	Type      string   `xml:"type,attr"`
	Id        string   `xml:"id,attr"`
	RawConfig []byte   `xml:",innerxml"`
}

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
	XMLName xml.Name `xml:"field"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	Width   string   `xml:"width,attr"`
}

func (xmlCfg *CgrXmlCfgDocument) GetCdreFWCfg(instName string) (*CgrXmlCdreFwCfg, error) {
	cdrefwCfg := new(CgrXmlCdreFwCfg)
	for _, cfgInst := range xmlCfg.Configurations {
		if cfgInst.Section != "cdre" || cfgInst.Type != utils.CDR_FIXED_WIDTH || cfgInst.Id != instName {
			continue
		}
		rawConfig := append([]byte("<element>"), cfgInst.RawConfig...)
		rawConfig = append(rawConfig, []byte("</element>")...)
		if err := xml.Unmarshal(rawConfig, cdrefwCfg); err != nil { // Encapsulate the rawConfig in one element so we can Unmarshall
			return nil, err
		} else if cdrefwCfg == nil {
			return nil, errors.New("Could not unmarshal CgrXmlCdreFwCfg")
		}
		return cdrefwCfg, nil
	}
	return nil, nil
}
